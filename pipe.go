// TODO:
//
// Fork and merge filter sequences.
// Trim (drops leading/trailing spaces)
//
// SplitRecordsAt(regexp)
//	Joins lines and then splits at regexp.
//	E.g., for blank line paras:
//		SplitRecordsAt("\n\n")
//	E.g., Go top level function boundaries:
//		SplitRecordsBefore(`^func\s`)
//	E.g., End at a closing brace/paren at start of line.
//		SplitRecordsAfter(`^[})]\s*\n`)
// Tentative operation:
//   buf := ""
//   k := 1024  // How much to accumulate before searching
//   for each input item {
//	add to buf
//	if len(buf) >= k {
//	  k = k*2  // Grow k to prevent quadratic regexp scanning cost
//	  while buf matches regexp {
//		split at/around regexp
//		yield prefix
//		buf = suffix
//		k = len(prefix) * 2  // Shrink back down
//	  }
//	}
//   }
//   yield buf  // last record
//
// Column code treats quoted strings as one column.
// Provide mechanisms to produce quoted columns.  Maybe:
//	Sequence(filesrc, AddColumn(hasher, size))
// Will produce:
//	'file name with spaces' 0x234134414 1232213
//
// Another quoter: regexp plus list of numbers.
//	Extract(`^(\d+)\s+(.*)$`, 2, 1)
//
// Find, ls, etc. produce quoted strings.
//
// Others:
//	Split(re)
//	SplitKeepEmpty(re)
//
// Also produce a way to remove quotes:
//	Sequence(filesrc, AddColumn(size), Sort(Num(2)), Select(2, 1), Unquote)
//
// Does Print() automatically unquote?  No.
//
// Maybe represent data as a sequence of Value objects.  Preserve columns
// internally.  Will it be too much mechanism and semantic surprises?
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

// Arg contains the data passed to a Filter. It mainly consists of
// a channel that produces the input to the filter, and a channel
// that receives the output from the filter.  It may be extended
// in the future to contain more fields.
type Arg struct {
	in  <-chan string
	out chan<- string
	// TODO: add an error channel here?
}

// Filter reads a sequence of strings from a channel and produces a
// sequence on another channel.
type Filter func(Arg)

// ForEach() returns a channel that contains all output emitted by a
// sequence of filters. The sequence of filters is fed an empty stream
// as the input.
func ForEach(filters ...Filter) <-chan string {
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
		passThrough(Arg{in, arg.out})
	}
}

// Print prints all output emitted by a sequence of filters. The
// sequence of filters is fed an empty stream as the input.
func Print(filters ...Filter) {
	for s := range ForEach(filters...) {
		fmt.Println(s)
	}
}

func runAndClose(f Filter, arg Arg) {
	f(arg)
	close(arg.out)
}

// passThrough copies all items read from in to out.
func passThrough(arg Arg) {
	for s := range arg.in {
		arg.out <- s
	}
}

// Echo copies its input and then emits item.
func Echo(items ...string) Filter {
	return func(arg Arg) {
		passThrough(arg)
		for _, s := range items {
			arg.out <- s
		}
	}
}

// Numbers copies its input and then emits the integers x..y
func Numbers(x, y int) Filter {
	return func(arg Arg) {
		passThrough(arg)
		for i := x; i <= y; i++ {
			arg.out <- fmt.Sprintf("%d", i)
		}
	}
}

// FindType is a mask that controls the types of nodes emitted by Find.
type FindType int

const (
	FILES    FindType = 1
	DIRS     FindType = 2
	SYMLINKS FindType = 4
	ALL      FindType = FILES | DIRS | SYMLINKS
)

// Find copies all input and then produces a sequence of items, one
// per file/directory/symlink found at or under dir that matches t.
func Find(t FindType, dir string) Filter {
	return func(arg Arg) {
		passThrough(arg)
		filepath.Walk(dir, func(f string, s os.FileInfo, e error) error {
			if t&FILES != 0 && s.Mode().IsRegular() ||
				t&DIRS != 0 && s.Mode().IsDir() ||
				t&SYMLINKS != 0 && s.Mode()&os.ModeSymlink != 0 {
				arg.out <- f
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
				arg.out <- scanner.Text()
			}
			file.Close()
		}
	}
}

// System executes "cmd args..." and produces one item per line in
// the output of the command.
func System(cmd string, args ...string) Filter {
	return func(arg Arg) {
		passThrough(arg)
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

// ReplaceMatch all occurrences of the regular expression r in an input item
// with replacement.  The replacement string can contain $1, $2, etc. which
// represent submatches of r.
func ReplaceMatch(r, replacement string) Filter {
	re := regexp.MustCompile(r)
	return func(arg Arg) {
		for s := range arg.in {
			arg.out <- re.ReplaceAllString(s, replacement)
		}
	}
}

// DeleteMatch deletes all occurrences of r in an input item.
func DeleteMatch(r string) Filter {
	return ReplaceMatch(r, "")
}

// Comparer is a function type that compares a and b and returns -1 if
// a occurs before b, +1 if a occurs after b key, 0 otherwise.
type Comparer func(a, b string) int

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

// Text returns a Comparer that compares the nth column lexicographically.
func Text(n int) Comparer {
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

// Num returns a Comparer that compares the nth column numerically.
func Num(n int) Comparer {
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

// Rev returns a Comparer that orders elements opposite of p.
func Rev(p Comparer) Comparer {
	return func(a, b string) int {
		return p(b, a)
	}
}

// columns is an interface for  sorting by a sequence of Comparers.
type columns struct {
	Data []string
	Cmp  []Comparer
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
func Sort(comparers ...Comparer) Filter {
	return func(arg Arg) {
		cs := columns{Cmp: comparers}
		for s := range arg.in {
			cs.Data = append(cs.Data, s)
		}
		sort.Sort(cs)
		for _, s := range cs.Data {
			arg.out <- s
		}
	}
}

// Reverse yields items in the reverse of the order it received them.
func Reverse(arg Arg) {
	var data []string
	for s := range arg.in {
		data = append(data, s)
	}
	for i := len(data) - 1; i >= 0; i-- {
		arg.out <- data[i]
	}
}

// First yields the first n items that it receives.
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

// DropFirst yields all items except for the first n items that it receives.
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

// Last yields the last n items that it receives.
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

// DropFirst yields all items except for the last n items that it receives.
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

// NumberLines prefixes its item with its index in the input sequence
// (starting at 1).
func NumberLines(arg Arg) {
	line := 1
	for s := range arg.in {
		arg.out <- fmt.Sprintf("%5d %s", line, s)
		line++
	}
}

// Cut emits just the bytes indexed [start..end] of each input item.
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

// Select splits each item into columns and yields the concatenation
// of the columns numbers passed as arguments to Select.  Columns are
// numbered starting at 1. A column number of 0 is interpreted as the
// full string.
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
