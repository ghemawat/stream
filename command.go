package pipe

import (
	"bytes"
	"fmt"
	"os"
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
			fmt.Fprintln(os.Stderr, err)
			return
		}
		splitIntoLines(bytes.NewBuffer(out), arg)
	}
}
