package pipe

import (
	"fmt"
	"os/exec"
)

// CommandMode controls the input/output handling for Filters returned
// by Command.
type CommandMode int

// OUTPUT and INPUT_OUTPUT control the input/output behavior of a command.
const (
	OUTPUT       CommandMode = 1
	INPUT_OUTPUT             = 2
)

// Command executes "command args...".
//
// If mode is pipe.INPUT_OUTPUT, the filter's input items are fed as
// standard input to the command, one line per input item.  Otherwise,
// the input items are copied verbatim to the filter output before the
// command is executed.
//
// If mode is pipe.OUTPUT or pipe.INPUT_OUTPUT, the standard output of
// the command is split into lines and the lines form the output of
// the filter (with trailing newlines removed).
func Command(mode CommandMode, command string, args ...string) Filter {
	if mode != OUTPUT && mode != INPUT_OUTPUT {
		return errorFilter(fmt.Errorf("pipe.Command: invalid mode %d", mode))
	}
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

		if false && mode == OUTPUT {
			passThrough(arg)
		} else {
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
		}

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
