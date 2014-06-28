// TODO:
//
// Fork and merge filter sequences.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"unicode"
)

type Arg struct {
	in  <-chan string
	out chan<- string
	// TODO: add an error channel here?
}

// Filter reads a sequence of strings from a channel and produces a
// sequence on another channel.  Many implementations of Filter are
// provided.
type Filter func(Arg)

// Each() returns a channel that contains all output emitted by a
// sequence of filters. The sequence of filters is fed an empty stream
// as the input.
func Each(filters ...Filter) <-chan string {
	in := make(chan string, 0)
	close(in)
	out := make(chan string, 10000)
	go runAndClose(Sequence(filters...), Arg{in, out})
	return out
}

// Sequence returns a filter that is the concatenation of all filter arguments.
func Sequence(filters ...Filter) Filter {
	return func(arg Arg) {
		in := arg.in
		for _, f := range filters {
			c := make(chan string, 10000)
			go runAndClose(f, Arg{in, c})
			in = c
		}
		copydata(Arg{in, arg.out})
	}
}

// Print prints all output emitted by a sequence of filters. The
// sequence of filters is fed an empty stream as the input.
func Print(filters ...Filter) {
	for s := range Each(filters...) {
		fmt.Println(s)
	}
}

func runAndClose(f Filter, arg Arg) {
	f(arg)
	close(arg.out)
}

// copydata copies all items read from in to out.
func copydata(arg Arg) {
	for s := range arg.in {
		arg.out <- s
	}
}

// Echo copies its input and then emits item.
func Echo(items ...string) Filter {
	return func(arg Arg) {
		copydata(arg)
		for _, s := range items {
			arg.out <- s
		}
	}
}

// Numbers copies its input and then emits the integers x..y
func Numbers(x, y int) Filter {
	return func(arg Arg) {
		copydata(arg)
		for i := x; i <= y; i++ {
			arg.out <- fmt.Sprintf("%d", i)
		}
	}
}

type FindType int

const (
	FILES    FindType = 1
	DIRS     FindType = 2
	SYMLINKS FindType = 4
	ALL      FindType = FILES | DIRS | SYMLINKS
)

func Find(ft FindType, dir string) Filter {
	return func(arg Arg) {
		copydata(arg)
		filepath.Walk(dir, func(f string, s os.FileInfo, e error) error {
			if ft&FILES != 0 && s.Mode().IsRegular() ||
				ft&DIRS != 0 && s.Mode().IsDir() ||
				ft&SYMLINKS != 0 && s.Mode()&os.ModeSymlink != 0 {
				arg.out <- f
			}
			return nil
		})
	}
}

func Cat(filenames ...string) Filter {
	return func(arg Arg) {
		copydata(arg)
		for _, f := range filenames {
			file, err := os.Open(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				arg.out <- scanner.Text()
			}
			file.Close()
		}
	}
}

func System(cmd string, args ...string) Filter {
	return func(arg Arg) {
		copydata(arg)
		out, err := exec.Command(cmd, args...).Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		scanner := bufio.NewScanner(bytes.NewBuffer(out))
		for scanner.Scan() {
			arg.out <- scanner.Text()
		}
	}
}

// If emits every input x for which fn(x) is true.
func If(fn func(string) bool) Filter {
	return func(arg Arg) {
		for s := range arg.in {
			if fn(s) {
				arg.out <- s
			}
		}
	}
}

// Grep emits every input x that matches the regular expression r.
func Grep(r string) Filter {
	re := regexp.MustCompile(r)
	return If(re.MatchString)
}

// GrepNot emits every input x that does not match the regular expression r.
func GrepNot(r string) Filter {
	re := regexp.MustCompile(r)
	return If(func(s string) bool { return !re.MatchString(s) })
}

// Uniq squashes adjacent identical items in in into a single output.
func Uniq(arg Arg) {
	first := true
	last := ""
	for s := range arg.in {
		if first || last != s {
			arg.out <- s
		}
		last = s
		first = false
	}
}

// UniqWithCount squashes adjacent identical items in in into a single
// output prefixed with the count of identical items.
func UniqWithCount(arg Arg) {
	current := ""
	count := 0
	for s := range arg.in {
		if s != current {
			if count > 0 {
				arg.out <- fmt.Sprintf("%d %s", count, current)
			}
			count = 0
			current = s
		}
		count++
	}
	if count > 0 {
		arg.out <- fmt.Sprintf("%d %s", count, current)
	}
}

// Parallel calls fn(x, out) for every item x in a pool of n goroutines.
func Parallel(n int, fn func(string, chan<- string)) Filter {
	// TODO: Maintain input order?
	// (a) Input goroutine generates <index, str> pairs
	// (b) n appliers read pairs and produce <index, fn(str)> pairs
	// (c) Output goroutine reads and emits in order
	return func(arg Arg) {
		wg := &sync.WaitGroup{}
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				for s := range arg.in {
					fn(s, arg.out)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func ReplaceMatch(r, replacement string) Filter {
	re := regexp.MustCompile(r)
	return func(arg Arg) {
		for s := range arg.in {
			arg.out <- re.ReplaceAllString(s, replacement)
		}
	}
}

func DeleteMatch(r string) Filter {
	return ReplaceMatch(r, "")
}

// Function that compares one sort key.
type SortPart func(a, b string) int

// column(s, n) returns 0,x where x is the nth column (1-based) in s,
// or -1,"" if s does not have n columns.  A zero column number is
// treated specially: 0,s is returned.
func column(s string, n int) (int, string) {
	if n == 0 {
		return 0, s
	}
	currentColumn := 0
	wstart := -1
	for i, c := range s {
		sp := unicode.IsSpace(c)
		switch {
		case !sp && wstart < 0: // Start of word
			currentColumn++
			wstart = i
		case sp && wstart >= 0 && currentColumn == n: // End of nth col
			return 0, s[wstart:i]
		case sp && wstart >= 0: // End of another column
			wstart = -1
		}
	}
	if wstart >= 0 && currentColumn == n { // nth column ends string
		return 0, s[wstart:]
	}

	// col not found. Treat as a value smaller than all strings
	return -1, ""
}

// Text returns a partial sort predicate that compares the nth column
// lexicographically.
func Text(n int) SortPart {
	return func(a, b string) int {
		a1, a2 := column(a, n)
		b1, b2 := column(b, n)
		switch {
		case a1 < b1:
			return -1
		case a1 > b1:
			return +1
		case a2 < b2:
			return -1
		case a2 > b2:
			return +1
		}
		return 0
	}
}

// Num returns a partial sort predicate that compares the nth column
// numerically.
func Num(n int) SortPart {
	return func(a, b string) int {
		a1, a2 := column(a, n)
		b1, b2 := column(b, n)
		switch {
		case a1 < b1:
			return -1
		case a1 > b1:
			return +1
		}

		// Convert columns from strings to numbers.
		a3, a4 := strconv.ParseInt(a2, 0, 64)
		b3, b4 := strconv.ParseInt(b2, 0, 64)

		if a4 != b4 {
			// Errors sort after numbers.
			if a4 != nil { // a had a parse error, b did not
				return +1
			} else { // b had a parse error, a did not
				return -1
			}
		}

		switch {
		case a3 < b3:
			return -1
		case a3 > b3:
			return +1
		}
		return 0
	}
}

func Rev(p SortPart) SortPart {
	return func(a, b string) int {
		return p(b, a)
	}
}

// columns allows sorting by a sequence of SortParts.
type columns struct {
	Data  []string
	Parts []SortPart
}

func (cs columns) Len() int      { return len(cs.Data) }
func (cs columns) Swap(i, j int) { cs.Data[i], cs.Data[j] = cs.Data[j], cs.Data[i] }
func (cs columns) Less(i, j int) bool {
	a := cs.Data[i]
	b := cs.Data[j]

	for _, p := range cs.Parts {
		c := p(a, b)
		if c != 0 {
			return c < 0
		}
	}
	return a < b
}

func Sort(parts ...SortPart) Filter {
	return func(arg Arg) {
		cs := columns{Parts: parts}
		for s := range arg.in {
			cs.Data = append(cs.Data, s)
		}
		sort.Sort(cs)
		for _, s := range cs.Data {
			arg.out <- s
		}
	}
}

func Reverse(arg Arg) {
	var data []string
	for s := range arg.in {
		data = append(data, s)
	}
	for i := len(data) - 1; i >= 0; i-- {
		arg.out <- data[i]
	}
}

func First(n int) Filter {
	return func(arg Arg) {
		emitted := 0
		for s := range arg.in {
			if emitted < n {
				arg.out <- s
				emitted++
			}
		}
	}
}

func DropFirst(n int) Filter {
	return func(arg Arg) {
		emitted := 0
		for s := range arg.in {
			if emitted >= n {
				arg.out <- s
			}
			emitted++
		}
	}
}

func Last(n int) Filter {
	return func(arg Arg) {
		var buf []string
		for s := range arg.in {
			buf = append(buf, s)
			if len(buf) > n {
				buf = buf[1:]
			}
		}
		for _, s := range buf {
			arg.out <- s
		}
	}
}

func DropLast(n int) Filter {
	return func(arg Arg) {
		var buf []string
		for s := range arg.in {
			buf = append(buf, s)
			if len(buf) > n {
				arg.out <- buf[0]
				buf = buf[1:]
			}
		}
	}
}

func NumberLines(arg Arg) {
	line := 1
	for s := range arg.in {
		arg.out <- fmt.Sprintf("%5d %s", line, s)
		line++
	}
}

func Cut(start, end int) Filter {
	return func(arg Arg) {
		for s := range arg.in {
			if len(s) > end {
				s = s[:end+1]
			}
			if len(s) < start {
				s = ""
			} else {
				s = s[start:]
			}
			arg.out <- s
		}
	}
}

// Prints the columns in order.  Column 0 is interpreted as full string.
func Select(columns ...int) Filter {
	return func(arg Arg) {
		for s := range arg.in {
			result := ""
			for _, col := range columns {
				if e, c := column(s, col); e == 0 && c != "" {
					if result != "" {
						result = result + " "
					}
					result = result + c
				}
			}
			arg.out <- result
		}
	}
}

func main() {
	dbl := func(arg Arg) {
		for s := range arg.in {
			arg.out <- s
			arg.out <- s
		}
	}

	hash := func(f string, out chan<- string) {
		file, err := os.Open(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		hasher := sha1.New()
		_, err = io.Copy(hasher, file)
		file.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		out <- fmt.Sprintf("%x %s", hasher.Sum(nil), f)
	}

	d := Echo(
		"8 1",
		"8 3 x",
		"8 3 w",
		"8 2",
		"4 5",
		"9 3",
		"12 13",
		"12 5",
	)
	Print(d, Sort(Text(1), Text(2)), Echo("----"))
	Print(d, Sort(Num(1), Num(2)), Echo("----"))
	Print(d, Sort(Text(1), Num(2)), Echo("----"))
	Print(d, Sort(Rev(Num(1)), Num(2)), Echo("----"))
	Print(d, Sort(), Echo("----"))
	Print(d, Sort(Text(2)), Echo("----"))

	Print(Numbers(1, 10),
		Grep("3"),
		dbl)

	Print(Sequence())
	Print(Sequence(Echo("1 of 1")))
	Print(Sequence(Echo("1 of 2"), Echo("2 of 2")))

	Print(Numbers(1, 100),
		Grep("3"),
		GrepNot("7"),
		dbl,
		Uniq,
		ReplaceMatch("^(.)$", "x$1"),
		Sort(),
		ReplaceMatch("^(.)", "$1 "),
		dbl,
		DeleteMatch(" .$"),
		UniqWithCount,
		Sort(Num(1)),
		Reverse,
		Echo("==="))

	Print(Find(FILES, "/home/sanjay/tmp"),
		Grep("/tmp/x"),
		GrepNot("/sub2/"),
		Parallel(4, hash),
		ReplaceMatch(" /home/sanjay/", " HOME/"))

	Print(Echo("a"), Echo("b"), Echo("c"))

	Print(Cat("/home/sanjay/.bashrc"),
		First(10),
		NumberLines,
		Cut(1, 50),
		ReplaceMatch("^", "LINE:"),
		Last(3))

	Print(Numbers(1, 10), DropFirst(8))
	Print(Numbers(1, 10), DropLast(7))

	Print(Echo("=== all ==="), Find(ALL, "/home/sanjay/tmp/x"))
	Print(Echo("=== dirs ==="), Find(FILES, "/home/sanjay/tmp/x"))
	Print(Echo("=== files ==="), Find(DIRS, "/home/sanjay/tmp/x"))

	Print(
		System("find", "/home/sanjay/tmp/y", "-ls"),
		Sort(Num(7), Text(11)),
	)

	// Reconcile example
	Print(Echo("------------"))
	Print(
		Find(FILES, "/home/sanjay/tmp/y"),
		GrepNot(`/home/sanjay/(\.Trash|Library)/`),
		Parallel(4, hash),
		Sort(Text(2)),
	)

	// Reconcile example (alternate)
	Print(Echo("------------"))
	Print(
		System("find", "/home/sanjay/tmp/y", "-type", "f", "-print"),
		Parallel(4, hash),
		Sort(Text(2)),
	)

}
