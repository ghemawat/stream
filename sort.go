package pipe

import (
	"sort"
	"strconv"
	"unicode"
)

// SortComparer is a function type that compares a and b and returns -1 if
// a occurs before b, +1 if a occurs after b, 0 otherwise.  See Sort.
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
			}
			// b had a parse error, a did not
			return -1
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
	return FilterFunc(func(arg Arg) error {
		cs := columns{Cmp: comparers}
		for s := range arg.In {
			cs.Data = append(cs.Data, s)
		}
		sort.Sort(cs)
		for _, s := range cs.Data {
			arg.Out <- s
		}
		return nil
	})
}
