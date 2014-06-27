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

type Filter func(<-chan string, chan<- string)

func (f Filter) To(g Filter) Filter {
	c := make(chan string, 10000)
	return func(in <-chan string, out chan<- string) {
		go g(c, out)
		f(in, c)
	}
}

func drain(in <-chan string) {
	for _ = range in {
	}
}

func Run(f Filter) {
	input := make(chan string, 0)
	close(input)
	output := make(chan string, 10000)
	go drain(output)
	f(input, output)
}

func Print(f Filter) {
	input := make(chan string, 0)
	close(input)
	output := make(chan string, 10000)
	go f(input, output)
	for s := range output {
		fmt.Println(s)
	}
}

func Echo(items ...string) Filter {
	return func(in <-chan string, out chan<- string) {
		go drain(in)
		for _, s := range items {
			out <- s
		}
		close(out)
	}
}

func Seq(x, y int) Filter {
	return func(in <-chan string, out chan<- string) {
		go drain(in)
		for i := x; i <= y; i++ {
			out <- fmt.Sprintf("%d", i)
		}
		close(out)
	}
}

func If(fn func(string) bool) Filter {
	return func(in <-chan string, out chan<- string) {
		for s := range in {
			if fn(s) {
				out <- s
			}
		}
		close(out)
	}
}

func Grep(r string) Filter {
	re := regexp.MustCompile(r)
	return If(re.MatchString)
}

func GrepNot(r string) Filter {
	re := regexp.MustCompile(r)
	return If(func(s string) bool { return !re.MatchString(s) })
}

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
	close(out)
}

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
	close(out)
}

func Apply(fn func(string) (string, bool)) Filter {
	return func(in <-chan string, out chan<- string) {
		for s := range in {
			if r, ok := fn(s); ok {
				out <- r
			}
		}
		close(out)
	}
}

func ApplyParallel(n int, fn func(string) (string, bool)) Filter {
	// TODO: Maintain input order?
	return func(in <-chan string, out chan<- string) {
		wg := &sync.WaitGroup{}
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				for s := range in {
					if r, ok := fn(s); ok {
						out <- r
					}
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(out)
	}
}

func ReplaceMatch(r, replacement string) Filter {
	re := regexp.MustCompile(r)
	return Apply(func(s string) (string, bool) {
		return re.ReplaceAllString(s, replacement), true
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
	close(out)
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
		close(out)
	}
}

func naturalSplit(s string) []string {
	r := regexp.MustCompile(`\s+`).Split(s, -1)
	for len(r) > 0 && r[0] == "" {
		r = r[1:]
	}
	for len(r) > 0 && r[len(r)-1] == "" {
		r = r[:len(r)-1]
	}
	return r
}

func toNum(cols []string, column int) (int, int64) {
	if len(cols) <= column {
		// Missing data is sorted before all numbers
		return -1, 0
	}
	x, err := strconv.ParseInt(cols[column], 0, 64)
	if err != nil {
		// Bad data is sorted after all numbers
		return +1, 0
	}
	return 0, x
}

func SortNumeric(column int) Filter {
	return SortBy(func(a, b string) bool {
		ca := naturalSplit(a)
		cb := naturalSplit(b)
		a1, a2 := toNum(ca, column)
		b1, b2 := toNum(cb, column)
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
	close(out)
}

func Find(dir string, matcher func(string, os.FileInfo, error) bool) Filter {
	return func(in <-chan string, out chan<- string) {
		go drain(in)
		filepath.Walk(dir, func(f string, s os.FileInfo, e error) error {
			if matcher == nil || matcher(f, s, e) {
				out <- f
			}
			return nil
		})
		close(out)
	}
}

func FindFiles(dir string) Filter {
	return Find(dir, func(f string, s os.FileInfo, e error) bool {
		return s.Mode().IsRegular()
	})
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
	close(out)
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
		close(out)
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
		close(out)
	}
}

func NumberLines(in <-chan string, out chan<- string) {
	line := 1
	for s := range in {
		out <- fmt.Sprintf("%5d %s", line, s)
		line++
	}
	close(out)
}

// TODO:
//  DropFirst(n)
//  DropLast(n)

func main() {
	dbl := func(in <-chan string, out chan<- string) {
		for s := range in {
			out <- s
			out <- s
		}
		close(out)
	}

	hash := func(f string) (string, bool) {
		file, err := os.Open(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return "", false
		}
		hasher := sha1.New()
		_, err = io.Copy(hasher, file)
		file.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return "", false
		}
		return fmt.Sprintf("%x %s", hasher.Sum(nil), f), true
	}

	Print(Seq(1, 100).
		To(Grep("3")).
		To(GrepNot("7")).
		To(dbl).
		To(Uniq).
		To(ReplaceMatch("^(.)$", "x$1")).
		To(Sort).
		To(ReplaceMatch("^(.)", "$1 ")).
		To(dbl).
		To(DeleteMatch(" .$")).
		To(UniqWithCount).
		To(SortNumeric(1)).
		To(Reverse))

	Print(FindFiles("/home/sanjay/tmp").
		To(ApplyParallel(4, hash)).
		To(Grep("/tmp/x")).
		To(GrepNot("/sub2/")).
		To(ReplaceMatch(" /home/sanjay/", " ")))

	Print(Echo("a", "b", "c"))

	Print(Echo("/home/sanjay/.bashrc").
		To(FileLines).
		To(First(10)).
		To(NumberLines).
		To(Last(3)))

	// Reconcile part 1
	// Print(FindFiles(dir).To(ApplyParallel(4, hash)))
}
