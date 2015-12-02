package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type tfm struct {
	list, keepInvalid bool
}

func tfmCommand() *cobra.Command {
	t := &tfm{}
	cmd := &cobra.Command{
		Use:   "tfm <files or directories>...",
		Short: "Transforms older pipeviz messages into newer forms.",
		Long:  "tfm takes messages as input and applies the indicated transformations to them, typically with the goal of converting them into a more proper, valid form. It can take messages on stdin (in which case it writes to stdout), or by specifying file or directory name(s) (in which case it writes directly back to the file).",
		Run:   t.Run,
	}

	cmd.Flags().BoolVarP(&t.list, "list", "l", false, "List available transforms")
	cmd.Flags().BoolVarP(&t.keepInvalid, "keep-invalid", "k", false, "Keep non-validating results. Normally invalid results will be discarded (no output if stdin, no writeback if files); this bypasses that behavior.")

	return cmd
}

// Run executes the tfm command.
func (t *tfm) Run(cmd *cobra.Command, args []string) {
	if t.list {
		// TODO list available transforms
		os.Exit(0)
	}

	// Figure out if we have something from stdin by checking stdin's fd type.
	// Not sure how cross-platform this is, but works on darwin and linux, at least.
	stat, _ := os.Stdin.Stat()
	hasStdin := (stat.Mode() & os.ModeCharDevice) == 0

	if hasStdin {
		if len(args) != 0 {
			fmt.Println("Cannot operate on both stdin and files")
			os.Exit(1)
		}
		t.runStdin(cmd)
	} else {
		if len(args) == 0 {
			fmt.Println("Must pass either a set of target files and directories, or some data on stdin")
			os.Exit(1)
		}
		t.runFiles(cmd, args)
	}
}

func (t *tfm) runStdin(cmd *cobra.Command) {
}

func (t *tfm) runFiles(cmd *cobra.Command, args []string) {
}
