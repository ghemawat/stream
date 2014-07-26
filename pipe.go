/*
Package pipe provides filters that can be chained together in a manner
similar to Unix pipelines.

Each filter takes as input a sequence of strings (read from a channel)
and produces as output a sequence of strings (written to a channel).

Filters can be chained together (e.g., via the Run function), the
output of one filter is fed as input to the next filter.  The empty
input is passed to the first filter. The following sequence will
print two lines to standard output:

	err := pipe.Run(
		pipe.Echo("hello", "world"),
		pipe.Reverse(),
		pipe.WriteLines(os.Stdout),
	)

An application can implement its own filters easily. For example,
Repeat(n) returns a filter that repeats every input n times.

	func Repeat(n int) FilterFunc {
		return func(arg pipe.Arg) error {
			for s := range arg.In {
				for i := 0; i < n; i++ {
					arg.Out <- s
				}
			}
			return nil
		}
	}

	pipe.Run(
		pipe.Echo("hello"),
		Repeat(10),
	)

Note that Repeat returns a FilterFunc, which is a common implementation
of the Filter interface.
*/
package pipe

import (
	"fmt"
	"sync"
)

// filterErrors records errors accumulated during the execution of a filter.
type filterErrors struct {
	mu     sync.Mutex
	errors []error
}

func (e *filterErrors) record(err error) {
	if err != nil {
		e.mu.Lock()
		e.errors = append(e.errors, err)
		e.mu.Unlock()
	}
}

func (e *filterErrors) getError() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	switch len(e.errors) {
	case 0:
		return nil
	case 1:
		return e.errors[0]
	default:
		return fmt.Errorf("pipe errors: %s", e.errors)
	}
}

// Arg contains the data passed to Filter.Run. Arg.In is a channel that
// produces the input to the filter, and Arg.Out is a channel that
// receives the output from the filter.
type Arg struct {
	In    <-chan string
	Out   chan<- string
	dummy bool // To allow later expansion
}

// The Filter interface represents a process that takes as input a
// sequence of strings from a channel and produces a sequence on
// another channel.
type Filter interface {
	// Run reads a sequence of items from Arg.In and produces a
	// sequence of items on Arg.Out.  Run returns nil on success,
	// an error otherwise.  Run must *not* close the Arg.Out
	// channel.
	Run(Arg) error
}

// FilterFunc is an adapter type that allows the use of ordinary
// functions as Filters.  If f is a function with the appropriate
// signature, FilterFunc(f) is a Filter that calls f.
type FilterFunc func(Arg) error

func (f FilterFunc) Run(arg Arg) error { return f(arg) }

// Sequence returns a filter that is the concatenation of all filter arguments.
// The output of a filter is fed as input to the next filter.
func Sequence(filters ...Filter) Filter {
	return FilterFunc(func(arg Arg) error {
		e := &filterErrors{}
		in := arg.In
		for _, f := range filters {
			c := make(chan string, 10000)
			go runFilter(f, Arg{In: in, Out: c}, e)
			in = c
		}
		for s := range in {
			arg.Out <- s
		}
		return e.getError()
	})
}

// Run executes the sequence of filters and discards all output.
// It returns either nil, an error if any filter reported an error.
func Run(filters ...Filter) error {
	return ForEach(Sequence(filters...), func(s string) {})
}

// ForEach calls fn(s) for every item s in the output of filter and
// returns either nil, or any error reported by the execution of the filter.
func ForEach(filter Filter, fn func(s string)) error {
	in := make(chan string, 0)
	close(in)
	out := make(chan string, 10000)
	e := &filterErrors{}
	go runFilter(filter, Arg{In: in, Out: out}, e)
	for s := range out {
		fn(s)
	}
	return e.getError()
}

// Output returns a slice that contains all items that are
// the output of filters.
func Output(filters ...Filter) ([]string, error) {
	var result []string
	err := ForEach(Sequence(filters...), func(s string) {
		result = append(result, s)
	})
	if err != nil {
		result = nil // Discard results on error
	}
	return result, err
}

func runFilter(f Filter, arg Arg, e *filterErrors) {
	e.record(f.Run(arg))
	close(arg.Out)
	for _ = range arg.In { // Discard all unhandled input
	}
}

// Echo emits items.
func Echo(items ...string) Filter {
	return FilterFunc(func(arg Arg) error {
		for _, s := range items {
			arg.Out <- s
		}
		return nil
	})
}

// Numbers emits the integers x..y
func Numbers(x, y int) Filter {
	return FilterFunc(func(arg Arg) error {
		for i := x; i <= y; i++ {
			arg.Out <- fmt.Sprintf("%d", i)
		}
		return nil
	})
}

// Map calls fn(x) for every item x and yields the outputs of the fn calls.
func Map(fn func(string) string) Filter {
	return FilterFunc(func(arg Arg) error {
		for s := range arg.In {
			arg.Out <- fn(s)
		}
		return nil
	})
}

// If emits every input x for which fn(x) is true.
func If(fn func(string) bool) Filter {
	return FilterFunc(func(arg Arg) error {
		for s := range arg.In {
			if fn(s) {
				arg.Out <- s
			}
		}
		return nil
	})
}

// Uniq squashes adjacent identical items in arg.In into a single output.
func Uniq() Filter {
	return FilterFunc(func(arg Arg) error {
		first := true
		last := ""
		for s := range arg.In {
			if first || last != s {
				arg.Out <- s
			}
			last = s
			first = false
		}
		return nil
	})
}

// UniqWithCount squashes adjacent identical items in arg.In into a single
// output prefixed with the count of identical items.
func UniqWithCount() Filter {
	return FilterFunc(func(arg Arg) error {
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
		return nil
	})
}

// Reverse yields items in the reverse of the order it received them.
func Reverse() Filter {
	return FilterFunc(func(arg Arg) error {
		var data []string
		for s := range arg.In {
			data = append(data, s)
		}
		for i := len(data) - 1; i >= 0; i-- {
			arg.Out <- data[i]
		}
		return nil
	})
}

// First yields the first n items that it receives.
func First(n int) Filter {
	return FilterFunc(func(arg Arg) error {
		emitted := 0
		for s := range arg.In {
			if emitted < n {
				arg.Out <- s
				emitted++
			}
		}
		return nil
	})
}

// DropFirst yields all items except for the first n items that it receives.
func DropFirst(n int) Filter {
	return FilterFunc(func(arg Arg) error {
		emitted := 0
		for s := range arg.In {
			if emitted >= n {
				arg.Out <- s
			}
			emitted++
		}
		return nil
	})
}

// Last yields the last n items that it receives.
func Last(n int) Filter {
	return FilterFunc(func(arg Arg) error {
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
		return nil
	})
}

// DropLast yields all items except for the last n items that it receives.
func DropLast(n int) Filter {
	return FilterFunc(func(arg Arg) error {
		var buf []string
		for s := range arg.In {
			buf = append(buf, s)
			if len(buf) > n {
				arg.Out <- buf[0]
				buf = buf[1:]
			}
		}
		return nil
	})
}

// NumberLines prefixes its item with its index in the input sequence
// (starting at 1).
func NumberLines() Filter {
	return FilterFunc(func(arg Arg) error {
		line := 1
		for s := range arg.In {
			arg.Out <- fmt.Sprintf("%5d %s", line, s)
			line++
		}
		return nil
	})
}

// Slice emits s[startOffset:endOffset] for each input item s.  Note
// that Slice follows Go conventions, and unlike the "cut" utility,
// offsets are numbered starting at zero, and the end offset is not
// included in the output.
func Slice(startOffset, endOffset int) Filter {
	return FilterFunc(func(arg Arg) error {
		for s := range arg.In {
			if len(s) > endOffset {
				s = s[:endOffset]
			}
			if len(s) < startOffset {
				s = ""
			} else {
				s = s[startOffset:]
			}
			arg.Out <- s
		}
		return nil
	})
}

// Columns splits each item into columns and yields the concatenation
// of the columns numbers passed as arguments.  Columns are numbered
// starting at 1.  If a column number is bigger than the number of columns
// in an item, it is skipped.
func Columns(columns ...int) Filter {
	return FilterFunc(func(arg Arg) error {
		for _, c := range columns {
			if c <= 0 {
				return fmt.Errorf("pipe.Columns: invalid column number %d", c)
			}
		}
		for s := range arg.In {
			result := ""
			for _, col := range columns {
				if _, c := column(s, col); c != "" {
					if result != "" {
						result = result + " "
					}
					result = result + c
				}
			}
			arg.Out <- result
		}
		return nil
	})
}
