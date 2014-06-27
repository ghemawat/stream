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

// Filter reads a sequence of strings from a channel and produces a
// sequence on another channel.  Many implementations of Filter are
// provided.
type Filter func(<-chan string, chan<- string)

// Each() returns a channel that contains all output emitted by a
// sequence of filters. The sequence of filters is fed an empty stream
// as the input.
func Each(filters ...Filter) <-chan string {
	c := make(chan string, 0)
	close(c) // No data sent to first filter
	for _, f := range filters {
		next := make(chan string, 10000)
		go func(x Filter, in <-chan string, out chan<- string) {
			x(in, out)
			close(out)
		}(f, c, next)
		c = next
	}
	return c // will contain output of last filter
}

// Print prints all output emitted by a sequence of filters. The
// sequence of filters is fed an empty stream as the input.
func Print(filters ...Filter) {
	for s := range Each(filters...) {
		fmt.Println(s)
	}
}

// copydata copies all items read from in to out.
func copydata(in <-chan string, out chan<- string) {
	for s := range in {
		out <- s
	}
}

// Echo copies its input and then emits item.
func Echo(item string) Filter {
	return func(in <-chan string, out chan<- string) {
		copydata(in, out)
		out <- item
	}
}

// Seq copies its input and then emits the integers x..y
func Seq(x, y int) Filter {
	return func(in <-chan string, out chan<- string) {
		copydata(in, out)
		for i := x; i <= y; i++ {
			out <- fmt.Sprintf("%d", i)
		}
	}
}

// If emits every input x for which fn(x) is true.
func If(fn func(string) bool) Filter {
	return func(in <-chan string, out chan<- string) {
		for s := range in {
			if fn(s) {
				out <- s
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
func Uniq(in <-chan string, out chan<- string) {
	first := true
	last := ""
	for s := range in {
		if first || last != s {
			out <- s
		}
		last = s
		first = false
	}
}

// UniqWithCount squashes adjacent identical items in in into a single
// output prefixed with the count of identical items.
func UniqWithCount(in <-chan string, out chan<- string) {
	current := ""
	count := 0
	for s := range in {
		if s != current {
			if count > 0 {
				out <- fmt.Sprintf("%d %s", count, current)
			}
			count = 0
			current = s
		}
		count++
	}
	if count > 0 {
		out <- fmt.Sprintf("%d %s", count, current)
	}
}

// Apply calls fn(x, out) in order for every item x.
func Apply(fn func(string, chan<- string)) Filter {
	return func(in <-chan string, out chan<- string) {
		for s := range in {
			fn(s, out)
		}
	}
}

// ApplyParallel calls fn(x, out) for every item x in a pool of n goroutines.
func ApplyParallel(n int, fn func(string, chan<- string)) Filter {
	// TODO: Maintain input order?
	// (a) Input goroutine generates <index, str> pairs
	// (b) n appliers read pairs and produce <index, fn(str)> pairs
	// (c) Output goroutine reads and emits in order
	return func(in <-chan string, out chan<- string) {
		wg := &sync.WaitGroup{}
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				for s := range in {
					fn(s, out)
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

func Sort(in <-chan string, out chan<- string) {
	var data []string
	for s := range in {
		data = append(data, s)
	}
	sort.Strings(data)
	for _, s := range data {
		out <- s
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
	return func(in <-chan string, out chan<- string) {
		var data fnStringSlice
		data.fn = fn
		for s := range in {
			data.StringSlice = append(data.StringSlice, s)
		}
		sort.Sort(data)
		for _, s := range data.StringSlice {
			out <- s
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

func Reverse(in <-chan string, out chan<- string) {
	var data []string
	for s := range in {
		data = append(data, s)
	}
	for i := len(data) - 1; i >= 0; i-- {
		out <- data[i]
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
	return func(in <-chan string, out chan<- string) {
		copydata(in, out)
		filepath.Walk(dir, func(f string, s os.FileInfo, e error) error {
			if ft&FILES != 0 && s.Mode().IsRegular() ||
				ft&DIRS != 0 && s.Mode().IsDir() ||
				ft&SYMLINKS != 0 && s.Mode()&os.ModeSymlink != 0 {
				out <- f
			}
			return nil
		})
	}
}

func FileLines(in <-chan string, out chan<- string) {
	for f := range in {
		file, err := os.Open(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			out <- scanner.Text()
		}
		file.Close()
	}
}

func First(n int) Filter {
	return func(in <-chan string, out chan<- string) {
		emitted := 0
		for s := range in {
			if emitted < n {
				out <- s
				emitted++
			}
		}
	}
}

func DropFirst(n int) Filter {
	return func(in <-chan string, out chan<- string) {
		emitted := 0
		for s := range in {
			if emitted >= n {
				out <- s
			}
			emitted++
		}
	}
}

func Last(n int) Filter {
	return func(in <-chan string, out chan<- string) {
		var buf []string
		for s := range in {
			buf = append(buf, s)
			if len(buf) > n {
				buf = buf[1:]
			}
		}
		for _, s := range buf {
			out <- s
		}
	}
}

func DropLast(n int) Filter {
	return func(in <-chan string, out chan<- string) {
		var buf []string
		for s := range in {
			buf = append(buf, s)
			if len(buf) > n {
				out <- buf[0]
				buf = buf[1:]
			}
		}
	}
}

func NumberLines(in <-chan string, out chan<- string) {
	line := 1
	for s := range in {
		out <- fmt.Sprintf("%5d %s", line, s)
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
	dbl := func(in <-chan string, out chan<- string) {
		for s := range in {
			out <- s
			out <- s
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

	Print()
	Print(Echo("a"))
	Print(Echo("a"), Echo("b"))

	Print(Seq(1, 100),
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

	Print(Echo("/home/sanjay/.bashrc"),
		FileLines,
		First(10),
		NumberLines,
		Cut(3, 50),
		Last(3))

	Print(Seq(1, 10), DropFirst(8))
	Print(Seq(1, 10), DropLast(7))

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
