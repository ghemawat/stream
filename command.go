package pipe

import (
	"fmt"
	"os/exec"
)

// Command executes "command args...".
//
// The filter's input items are fed as standard input to the command,
// one line per input item. The standard output of the command is
// split into lines and the lines form the output of the filter (with
// trailing newlines removed).
func Command(command string, args ...string) Filter {
	return func(arg Arg) {
		cmd := exec.Command(command, args...)
		output, err := cmd.StdoutPipe()
		if err != nil {
			arg.ReportError(err)
			for _ = range arg.In {
				// Discard
			}
			return
		}

		// Send incoming items to command's standard input
		input, err := cmd.StdinPipe()
		if err != nil {
			arg.ReportError(err)
			for _ = range arg.In {
				// Discard
			}
			return
		}
		go func() {
			for s := range arg.In {
				fmt.Fprintln(input, s)
			}
			input.Close()
		}()

		// Start the command and process it's standard output.
		err = cmd.Start()
		if err == nil {
			splitIntoLines(output, arg)
			err = cmd.Wait()
		}
		if err != nil {
			arg.ReportError(err)
		}
	}
}

// Example:
//	pipe.Command(pipe.OUTPUT, "find", ".")
//	pipe.Command(pipe.INPUT_OUTPUT, "wc", "-l")
