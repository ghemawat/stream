// TODO:
//
// Fork and merge filter sequences.
package main

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"
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
		for _, f := range filters {
			c := make(chan string, 10000)
			go runAndClose(f, arg)
			arg.in = c
		}
		copydata(arg)
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

// Apply calls fn(x, out) in order for every item x.
func Apply(fn func(string, chan<- string)) Filter {
	return func(arg Arg) {
		for s := range arg.in {
			fn(s, arg.out)
		}
	}
}

// ApplyParallel calls fn(x, out) for every item x in a pool of n goroutines.
func ApplyParallel(n int, fn func(string, chan<- string)) Filter {
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
	return Apply(func(s string, out chan<- string) {
		out <- re.ReplaceAllString(s, replacement)
	})
}

func DeleteMatch(r string) Filter {
	return ReplaceMatch(r, "")
}

func Sort(arg Arg) {
	var data []string
	for s := range arg.in {
		data = append(data, s)
	}
	sort.Strings(data)
	for _, s := range data {
		arg.out <- s
	}
}

type fnStringSlice struct {
	sort.StringSlice
	fn func(string, string) bool
}

func (fs fnStringSlice) Less(i, j int) bool {
	return fs.fn(fs.StringSlice[i], fs.StringSlice[j])
}

func SortBy(fn func(string, string) bool) Filter {
	return func(arg Arg) {
		var data fnStringSlice
		data.fn = fn
		for s := range arg.in {
			data.StringSlice = append(data.StringSlice, s)
		}
		sort.Sort(data)
		for _, s := range data.StringSlice {
			arg.out <- s
		}
	}
}

func SortNumeric(column int) Filter {
	toNum := func(s string) (int, int64) {
		r := regexp.MustCompile(`\s+`).Split(s, -1)
		for len(r) > 0 && r[0] == "" {
			r = r[1:]
		}
		for len(r) > 0 && r[len(r)-1] == "" {
			r = r[:len(r)-1]
		}
		if len(r) <= column {
			// Missing data is sorted before all numbers
			return -1, 0
		}
		x, err := strconv.ParseInt(r[column], 0, 64)
		if err != nil {
			// Bad data is sorted after all numbers
			return +1, 0
		}
		return 0, x
	}
	return SortBy(func(a, b string) bool {
		a1, a2 := toNum(a)
		b1, b2 := toNum(b)
		if a1 != b1 {
			return a1 < b1
		}
		return a2 < b2
	})
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
	return Apply(func(s string, out chan<- string) {
		if len(s) > end {
			s = s[:end+1]
		}
		if len(s) < start {
			s = ""
		} else {
			s = s[start:]
		}
		out <- s
	})
}

func main() {
	dbl := Apply(func(s string, out chan<- string) { out <- s; out <- s })
	_ = func(arg Arg) {
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

	Print(Sequence())
	Print(Sequence(Echo("1 of 1")))
	Print(Sequence(Echo("1 of 2"), Echo("2 of 2")))

	Print(Numbers(1, 100),
		Grep("3"),
		GrepNot("7"),
		dbl,
		Uniq,
		ReplaceMatch("^(.)$", "x$1"),
		Sort,
		ReplaceMatch("^(.)", "$1 "),
		dbl,
		DeleteMatch(" .$"),
		UniqWithCount,
		SortNumeric(1),
		Reverse,
		Echo("==="))

	Print(Find(FILES, "/home/sanjay/tmp"),
		Grep("/tmp/x"),
		GrepNot("/sub2/"),
		ApplyParallel(4, hash),
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

	// Reconcile part 1
	Print(Echo("------------"))
	Print(
		Find(FILES, "/home/sanjay/tmp/y"),
		ApplyParallel(4, hash),
	)
}
