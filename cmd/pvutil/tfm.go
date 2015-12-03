package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type tfm struct {
	list, keepInvalid bool
	transforms        string
}

// A MessageTransformer takes a message and transforms it into a new message.
type MessageTransformer interface {
	Transform([]byte) (result []byte, changed bool, err error)
}

func tfmCommand() *cobra.Command {
	t := &tfm{}
	cmd := &cobra.Command{
		Use:   "tfm <files>...",
		Short: "Transforms older pipeviz messages into newer forms.",
		Long:  "tfm takes messages as input and applies the indicated transformations to them, typically with the goal of converting them into a more proper, valid form. It can take messages on stdin (in which case it writes to stdout), or by specifying file name(s) (in which case it writes directly back to the file).",
		Run:   t.Run,
	}

	cmd.Flags().BoolVarP(&t.list, "list", "l", false, "List available transforms")
	cmd.Flags().StringVarP(&t.transforms, "transforms", "t", "", "A comma-separated list of transforms to apply. At least one transform must be applied.")
	cmd.Flags().BoolVarP(&t.keepInvalid, "keep-invalid", "k", false, "Keep non-validating results. Normally invalid results will be discarded (no output if stdin, no writeback if files); this bypasses that behavior.")

	return cmd
}

// Run executes the tfm command.
func (t *tfm) Run(cmd *cobra.Command, args []string) {
	if t.list {
		// TODO list available transforms
		os.Exit(0)
	}

	if t.transforms == "" {
		fmt.Println("Must specify at least one transform to apply")
		os.Exit(1)
	}

	tf := getTransfomers(t.transforms)
	if len(tf) == 0 {
		// Output from getTransformers is sufficient, can just exit directly here
		os.Exit(1)
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
		t.runStdin(cmd, tf)
	} else {
		if len(args) == 0 {
			fmt.Println("Must pass either a set of target files, or some data on stdin")
			os.Exit(1)
		}
		files := getFiles(args)
		if len(files) == 0 {
			// Output from getFiles is sufficient, can just exit directly here
			os.Exit(1)
		}
		t.runFiles(cmd, tf, files)
	}
}

func (t *tfm) runStdin(cmd *cobra.Command, tf []MessageTransformer) {

}

func (t *tfm) runFiles(cmd *cobra.Command, tf []MessageTransformer, files []*os.File) {

}

func getTransfomers(list string) []MessageTransformer {
	//fmt.Printf("No transform exists with the name %q\n", name)
	return make([]MessageTransformer, 0)
}

func getFiles(args []string) []*os.File {
	return nil
}
