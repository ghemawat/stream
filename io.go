package pipe

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Cat emits each line from each named file in order.
//
// One difference from the "cat" binary is that any input items are
// copied verbatim to the output before any data from the named files
// is emitted. So the following two pipelines are equivalent:
//	Cat("a", "b")
//	Sequence(Cat("a"), Cat(b"))
func Cat(filenames ...string) Filter {
	return func(arg Arg) {
		passThrough(arg)
		for _, f := range filenames {
			file, err := os.Open(f)
			if err != nil {
				arg.ReportError(err)
				continue
			}
			splitIntoLines(file, arg)
			file.Close()
		}
	}
}

// WriteLines prints each input item s followed by a newline to
// writer; and in addition it emits s.  Therefore WriteLines()
// can be used like the "tee" command, which can often be useful
// for debugging.
func WriteLines(writer io.Writer) Filter {
	return func(arg Arg) {
		reported := false
		for s := range arg.In {
			_, err := fmt.Fprintln(writer, s)
			if !reported && err != nil {
				arg.ReportError(err)
				reported = true
			}
			arg.Out <- s
		}
	}
}

// ReadLines emits each line found in reader.  Any input items are
// copied verbatim to the output before reader is processed.
func ReadLines(reader io.Reader) Filter {
	return func(arg Arg) {
		passThrough(arg)
		splitIntoLines(reader, arg)
	}
}

func splitIntoLines(rd io.Reader, arg Arg) {
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		arg.Out <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		arg.ReportError(err)
	}
}
