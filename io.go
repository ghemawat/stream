package pipe

import (
	"io"
	"os"
)

// Cat copies all input and then emits each line from each named file in order.
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

// Tee copies all input to both writer and the output channel.
func Tee(writer io.Writer) Filter {
	return func(arg Arg) {
		for s := range arg.In {
			io.WriteString(writer, s)
			io.WriteString(writer, "\n")
			arg.Out <- s
		}
	}
}

// Lines copies all input and then emits each line found in reader.
func Lines(reader io.Reader) Filter {
	return func(arg Arg) {
		passThrough(arg)
		splitIntoLines(reader, arg)
	}
}
