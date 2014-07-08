package pipe

import (
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

// WriteLines emits each input item s and in addition prints s to writer
// followed by a newline.
func WriteLines(writer io.Writer) Filter {
	return func(arg Arg) {
		for s := range arg.In {
			io.WriteString(writer, s)
			io.WriteString(writer, "\n")
			arg.Out <- s
		}
	}
}

// ReadLines emits each line found in reader.
// Any input items are copied verbatim to the output before reader is processed.
func ReadLines(reader io.Reader) Filter {
	return func(arg Arg) {
		passThrough(arg)
		splitIntoLines(reader, arg)
	}
}
