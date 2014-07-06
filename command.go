package pipe

import (
	"os"
	"os/exec"
)

// CommandOutput executes "cmd args..." and produces one item per line in
// the output of the command.
func CommandOutput(cmd string, args ...string) Filter {
	// TODO: Also add xargs, unix command filter
	return func(arg Arg) {
		passThrough(arg)
		cmd := exec.Command(cmd, args...)
		output, err := cmd.StdoutPipe()
		if err == nil {
			err = cmd.Start()
		}
		if err == nil {
			splitIntoLines(output, arg)
			err = cmd.Wait()
		}
		if err != nil {
			arg.ReportError(err)
		}
	}
}

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
