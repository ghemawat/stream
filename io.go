package pipe

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Cat emits each line from each named file in order. If no arguments
// are specified, Cat copies its input to its output.
func Cat(filenames ...string) Filter {
	return func(arg Arg) error {
		if len(filenames) == 0 {
			for s := range arg.In {
				arg.Out <- s
			}
			return nil
		}
		for _, f := range filenames {
			file, err := os.Open(f)
			if err == nil {
				err = splitIntoLines(file, arg)
				file.Close()
			}
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// WriteLines prints each input item s followed by a newline to
// writer; and in addition it emits s.  Therefore WriteLines()
// can be used like the "tee" command, which can often be useful
// for debugging.
func WriteLines(writer io.Writer) Filter {
	return func(arg Arg) error {
		for s := range arg.In {
			if _, err := fmt.Fprintln(writer, s); err != nil {
				return err
			}
			arg.Out <- s
		}
		return nil
	}
}

// ReadLines emits each line found in reader.
func ReadLines(reader io.Reader) Filter {
	return func(arg Arg) error {
		return splitIntoLines(reader, arg)
	}
}

func splitIntoLines(rd io.Reader, arg Arg) error {
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		arg.Out <- scanner.Text()
	}
	return scanner.Err()
}

// Progress copies all items to its output and reports a progress
// message to writer every interval items.
func Progress(writer io.Writer, interval int) Filter {
	return func(arg Arg) error {
		if interval <= 0 {
			return fmt.Errorf("pipe.Progress: invalid interval %d", interval)
		}
		seen := 0
		for s := range arg.In {
			seen++
			if seen%interval == 0 {
				fmt.Fprintf(writer, "... %d\n", seen)
			}
			arg.Out <- s
		}
		return nil
	}
}
