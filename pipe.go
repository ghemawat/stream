// Package pipe provides filters that can be chained together in a manner
// similar to Unix pipelines. Each filter is a function that takes as
// input a sequence of strings (read from a channel) and produces as
// output a sequence of strings (written to a channel).
package pipe

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"unicode"
)

// Arg contains the data passed to a Filter. The important parts are
// a channel that produces the input to the filter, and a channel
// that receives the output from the filter.  It may be extended
// in the future to contain more fields.
type Arg struct {
	In  <-chan string // In yields the sequence of items that are the input to a Filter.
	Out chan<- string // Out consumes the sequence of items that are the output of a Filter.
	// TODO: add an error channel here?
}

// Filter reads a sequence of strings from a channel and produces a
// sequence on another channel.
type Filter func(Arg)

// ForEach() returns a channel that contains all output emitted by a
// sequence of filters. The empty stream is fed as input to the first filter.
// The output of each filter is fed as input to the next filter. The
// output of the last filter is returned.
func ForEach(filters ...Filter) <-chan string {
	in := make(chan string, 0)
	close(in)
	out := make(chan string, 10000)
	go runAndClose(Sequence(filters...), Arg{in, out})
	return out
}

// Sequence returns a filter that is the concatenation of all filter arguments.
// The output of a filter is fed as input to the next filter.
func Sequence(filters ...Filter) Filter {
	return func(arg Arg) {
		in := arg.In
		for _, f := range filters {
			c := make(chan string, 10000)
			go runAndClose(f, Arg{in, c})
			in = c
		}
		passThrough(Arg{in, arg.Out})
	}
}

// Print() prints all items emitted by a sequence of filters, one per
// line. The empty stream is fed as input to the first filter.  The
// output of each filter is fed as input to the next filter. The
// output of the last filter is printed.
func Print(filters ...Filter) {
	for s := range ForEach(filters...) {
		fmt.Println(s)
	}
}

func runAndClose(f Filter, arg Arg) {
	f(arg)
	close(arg.Out)
}

// passThrough copies all items read from in to out.
func passThrough(arg Arg) {
	for s := range arg.In {
		arg.Out <- s
	}
}

// Echo copies its input and then emits items.
func Echo(items ...string) Filter {
	return func(arg Arg) {
		passThrough(arg)
		for _, s := range items {
			arg.Out <- s
		}
	}
}

// Numbers copies its input and then emits the integers x..y
func Numbers(x, y int) Filter {
	return func(arg Arg) {
		passThrough(arg)
		for i := x; i <= y; i++ {
			arg.Out <- fmt.Sprintf("%d", i)
		}
	}
}

// FindMatch is a mask that controls the types of nodes emitted by Find.
type FindMatch int

const (
	FILES    FindMatch = 1    // Match regular files
	DIRS               = 2    // Match directories
	SYMLINKS           = 4    // Match symbolic links
	ALL                = 0xff // Match everything
)

// Find copies all input and then produces a sequence of items, one
// per file/directory/symlink found at or under dir that matches mask.
func Find(mask FindMatch, dir string) Filter {
	return func(arg Arg) {
		passThrough(arg)
		filepath.Walk(dir, func(f string, s os.FileInfo, e error) error {
			if mask&ALL == ALL ||
				mask&FILES != 0 && s.Mode().IsRegular() ||
				mask&DIRS != 0 && s.Mode().IsDir() ||
				mask&SYMLINKS != 0 && s.Mode()&os.ModeSymlink != 0 {
				arg.Out <- f
			}
			return nil
		})
	}
}

// Cat copies all input and then emits each line from each named file in order.
func Cat(filenames ...string) Filter {
	return func(arg Arg) {
		passThrough(arg)
		for _, f := range filenames {
			file, err := os.Open(f)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				arg.Out <- scanner.Text()
			}
			file.Close()
		}
	}
}

// System executes "cmd args..." and produces one item per line in
// the output of the command.
func System(cmd string, args ...string) Filter {
	// TODO: Also add xargs, unix command filter
	return func(arg Arg) {
		passThrough(arg)
		out, err := exec.Command(cmd, args...).Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		scanner := bufio.NewScanner(bytes.NewBuffer(out))
		for scanner.Scan() {
			arg.Out <- scanner.Text()
		}
	}
}

// If emits every input x for which fn(x) is true.
func If(fn func(string) bool) Filter {
	return func(arg Arg) {
		for s := range arg.In {
			if fn(s) {
				arg.Out <- s
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

// Uniq squashes adjacent identical items in arg.In into a single output.
func Uniq() Filter {
	return func(arg Arg) {
		first := true
		last := ""
		for s := range arg.In {
			if first || last != s {
				arg.Out <- s
			}
			last = s
			first = false
		}
	}
}

// UniqWithCount squashes adjacent identical items in arg.In into a single
// output prefixed with the count of identical items.
func UniqWithCount() Filter {
	return func(arg Arg) {
		current := ""
		count := 0
		for s := range arg.In {
			if s != current {
				if count > 0 {
					arg.Out <- fmt.Sprintf("%d %s", count, current)
				}
				count = 0
				current = s
			}
			count++
		}
		if count > 0 {
			arg.Out <- fmt.Sprintf("%d %s", count, current)
		}
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
				for s := range arg.In {
					fn(s, arg.Out)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

type parItem struct {
	index int
	value string
}

// MapConcurrent calls fn(x) for every item x in a pool of n
// goroutines and yields the outputs of the fn calls. The output order
// matches the input order.
func MapConcurrent(n int, fn func(string) string) Filter {
	return func(arg Arg) {
		// Attach a sequence number to each item.
		source := make(chan parItem, 10000)
		go func() {
			i := 0
			for s := range arg.In {
				source <- parItem{i, s}
				i++
			}
			close(source)
		}()

		// We keep track of outputs in a map indexed by the
		// sequence number of the item.  These items are
		// yielded in order.
		var mu sync.Mutex
		buffered := make(map[int]parItem)
		next := 0

		// Process the items in n go routines.
		wg := &sync.WaitGroup{}
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				for item := range source {
					s := fn(item.value)

					// Record item and yield in order
					mu.Lock()
					buffered[item.index] = parItem{item.index, s}
					for {
						x, ok := buffered[next]
						if !ok {
							break
						}
						arg.Out <- x.value
						delete(buffered, next)
						next++
					}
					mu.Unlock()
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

// Substitute replaces all occurrences of the regular expression r in
// an input item with replacement.  The replacement string can contain
// $1, $2, etc. which represent submatches of r.
func Substitute(r, replacement string) Filter {
	re := regexp.MustCompile(r)
	return func(arg Arg) {
		for s := range arg.In {
			arg.Out <- re.ReplaceAllString(s, replacement)
		}
	}
}

// SortComparer is a function type that compares a and b and returns -1 if
// a occurs before b, +1 if a occurs after b key, 0 otherwise.  See Sort.
type SortComparer func(a, b string) int

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

// Textual returns a SortComparer that compares the nth column
// lexicographically.  A string that does not contain an nth column
// sorts before all strings that contain an nth column.  If n == 0,
// the entire string is treated as one column.
func Textual(n int) SortComparer {
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

// Numeric returns a SortComparer that compares the nth column numerically.
// A string that does not contain an nth column sorts before all strings
// that contain an nth column. If the nth column is not a number, it
// sorts after all strings that contain an nth column that is a number.
// If n == 0, the entire string is treated as one column.
func Numeric(n int) SortComparer {
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

// Descending returns a SortComparer that orders elements opposite of p.
func Descending(p SortComparer) SortComparer {
	return func(a, b string) int {
		return p(b, a)
	}
}

// columns is an interface for  sorting by a sequence of SortComparers.
type columns struct {
	Data []string
	Cmp  []SortComparer
}

func (c columns) Len() int      { return len(c.Data) }
func (c columns) Swap(i, j int) { c.Data[i], c.Data[j] = c.Data[j], c.Data[i] }
func (c columns) Less(i, j int) bool {
	a := c.Data[i]
	b := c.Data[j]
	for _, p := range c.Cmp {
		r := p(a, b)
		if r != 0 {
			return r < 0
		}
	}
	return a < b
}

// Sort sorts its inputs by the specified sequence of comparers.
func Sort(comparers ...SortComparer) Filter {
	return func(arg Arg) {
		cs := columns{Cmp: comparers}
		for s := range arg.In {
			cs.Data = append(cs.Data, s)
		}
		sort.Sort(cs)
		for _, s := range cs.Data {
			arg.Out <- s
		}
	}
}

// Reverse yields items in the reverse of the order it received them.
func Reverse() Filter {
	return func(arg Arg) {
		var data []string
		for s := range arg.In {
			data = append(data, s)
		}
		for i := len(data) - 1; i >= 0; i-- {
			arg.Out <- data[i]
		}
	}
}

// First yields the first n items that it receives.
func First(n int) Filter {
	return func(arg Arg) {
		emitted := 0
		for s := range arg.In {
			if emitted < n {
				arg.Out <- s
				emitted++
			}
		}
	}
}

// DropFirst yields all items except for the first n items that it receives.
func DropFirst(n int) Filter {
	return func(arg Arg) {
		emitted := 0
		for s := range arg.In {
			if emitted >= n {
				arg.Out <- s
			}
			emitted++
		}
	}
}

// Last yields the last n items that it receives.
func Last(n int) Filter {
	return func(arg Arg) {
		var buf []string
		for s := range arg.In {
			buf = append(buf, s)
			if len(buf) > n {
				buf = buf[1:]
			}
		}
		for _, s := range buf {
			arg.Out <- s
		}
	}
}

// DropFirst yields all items except for the last n items that it receives.
func DropLast(n int) Filter {
	return func(arg Arg) {
		var buf []string
		for s := range arg.In {
			buf = append(buf, s)
			if len(buf) > n {
				arg.Out <- buf[0]
				buf = buf[1:]
			}
		}
	}
}

// NumberLines prefixes its item with its index in the input sequence
// (starting at 1).
func NumberLines() Filter {
	return func(arg Arg) {
		line := 1
		for s := range arg.In {
			arg.Out <- fmt.Sprintf("%5d %s", line, s)
			line++
		}
	}
}

// Cut emits just the bytes indexed [start..end] of each input item.
func Cut(start, end int) Filter {
	return func(arg Arg) {
		for s := range arg.In {
			if len(s) > end {
				s = s[:end+1]
			}
			if len(s) < start {
				s = ""
			} else {
				s = s[start:]
			}
			arg.Out <- s
		}
	}
}

// Select splits each item into columns and yields the concatenation
// of the columns numbers passed as arguments to Select.  Columns are
// numbered starting at 1. A column number of 0 is interpreted as the
// full string.
func Select(columns ...int) Filter {
	return func(arg Arg) {
		for s := range arg.In {
			result := ""
			for _, col := range columns {
				if e, c := column(s, col); e == 0 && c != "" {
					if result != "" {
						result = result + " "
					}
					result = result + c
				}
			}
			arg.Out <- result
		}
	}
}
