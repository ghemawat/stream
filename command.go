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
	return func(arg Arg) error {
		cmd := exec.Command(command, args...)
		input, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		output, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}
		go func() {
			for s := range arg.In {
				fmt.Fprintln(input, s)
			}
			input.Close()
		}()
		if err := splitIntoLines(output, arg); err != nil {
			cmd.Wait()
			return err
		}
		return cmd.Wait()
	}
}

// Xargs executes "command args... items..." where items are the input
// to the Xargs filter.  The handling of items may be split across
// multiple executions of command (typically to meet command line
// length restrictions).  The standard output of the execution(s) is
// split into lines and the lines form the output of the filter (with
// trailing newlines removed).
func Xargs(command string, args ...string) Filter {
	return func(arg Arg) error {
		// Compute argument length limit per execution
		const limitBytes = 4096 - 100 // Posix limit with some slop
		baseBytes := len(command)
		var items []string
		for _, a := range args {
			items = append(items, a)
			baseBytes += 1 + len(a)
		}

		// Helper that executes command with accumulated arguments.
		run := func() error {
			cmd := exec.Command(command, items...)
			output, err := cmd.StdoutPipe()
			if err != nil {
				return err
			}
			if err := cmd.Start(); err != nil {
				return err
			}
			if err := splitIntoLines(output, arg); err != nil {
				cmd.Wait()
				return err
			}
			return cmd.Wait()
		}

		// Buffer items until we hit length limit
		usedBytes := baseBytes
		for s := range arg.In {
			if len(items) > len(args) && usedBytes+1+len(s) >= limitBytes {
				err := run()
				if err != nil {
					return err
				}
				items = items[0:len(args)]
				usedBytes = baseBytes
			}
			items = append(items, s)
			usedBytes += 1 + len(s)
		}
		if len(items) > len(args) {
			return run()
		}
		return nil
	}
}
