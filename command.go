package pipe

import (
	"bytes"
	"os/exec"
)

// CommandOutput executes "cmd args..." and produces one item per line in
// the output of the command.
func CommandOutput(cmd string, args ...string) Filter {
	// TODO: Also add xargs, unix command filter
	return func(arg Arg) {
		passThrough(arg)
		out, err := exec.Command(cmd, args...).Output()
		if err != nil {
			reportError(err)
		} else {
			splitIntoLines(bytes.NewBuffer(out), arg)
		}
	}
}
