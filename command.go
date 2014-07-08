package pipe

import (
	"os/exec"
)

// CommandOutput executes "cmd args..." and produces one item per line in
// the output of the command.
//
// Any input items are copied verbatim to the output before cmd is executed.
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
