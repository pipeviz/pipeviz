package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/pipeviz/pipeviz/schema"
	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
)

type tfm struct {
	list, keepInvalid, quiet bool
	transforms               string
	errWriter                io.Writer
}

// A MessageTransformer takes a message and transforms it into a new message.
type MessageTransformer interface {
	fmt.Stringer
	Transform(io.Reader, io.Writer) (changed bool, err error)
}

type MessageTransformers []MessageTransformer

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
	cmd.Flags().BoolVarP(&t.keepInvalid, "keep-invalid", "k", false, "Keep non-validating results. Normally non-valid results will be discarded (no output if stdin, no write if files); this bypasses that behavior. Note that malformed JSON will *never* be used, regardless of this flag.")
	cmd.Flags().BoolVarP(&t.quiet, "quiet", "q", false, "Suppress error and informational output.")

	return cmd
}

// Run executes the tfm command.
func (t *tfm) Run(cmd *cobra.Command, args []string) {
	if t.quiet {
		t.errWriter = ioutil.Discard
	} else {
		t.errWriter = os.Stderr
	}

	if t.list {
		// TODO list available transforms
		os.Exit(0)
	}

	if t.transforms == "" {
		fmt.Fprintf(t.errWriter, "Must specify at least one transform to apply\n")
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
			fmt.Fprintf(t.errWriter, "Cannot operate on both stdin and files\n")
			os.Exit(1)
		}
		t.runStdin(cmd, tf)
	} else {
		if len(args) == 0 {
			fmt.Fprintf(t.errWriter, "Must pass either a set of target files, or some data on stdin\n")
			os.Exit(1)
		}

		t.runFiles(cmd, tf, args)
	}
}

func (t *tfm) runStdin(cmd *cobra.Command, tfs MessageTransformers) {

}

func (t *tfm) runFiles(cmd *cobra.Command, tfs MessageTransformers, names []string) {
	var atLeastOne bool
	var wg sync.WaitGroup

	fn := func(name string) {
		defer wg.Done()
		f, err := os.OpenFile(name, os.O_RDWR, 0666)
		defer f.Close()
		if err != nil {
			fmt.Fprintf(t.errWriter, "Error while opening file %q: %s\n", name, err)
			return
		}

		var changed bool
		var w, w2 *bytes.Buffer
		io.Copy(w, f)
		for _, tf := range tfs {
			didchange, terr := tf.Transform(w, w2)
			if didchange {
				changed = true
			}
			if terr != nil {
				fmt.Fprintf(t.errWriter, "Skipping transform of %q due to error while applying transform %q: %s\n", name, tf, terr)
				return
			}

			w, w2 = w2, w
			w2.Reset()
		}
		var finished []byte
		io.Copy(bytes.NewBuffer(finished), w)

		if !changed {
			fmt.Fprintf(t.errWriter, "No changes resulted from applying transforms to %q\n", name)
			return
		}

		result, err := schema.Master().Validate(gojsonschema.NewStringLoader(string(finished)))
		if !result.Valid() {
			// Buffer so that there's no interleaving of printed output
			w := bytes.NewBuffer(make([]byte, 0))

			fmt.Fprintf(w, "Errors encountered while validating %q:\n", name)
			for _, desc := range result.Errors() {
				fmt.Fprintf(w, "%s\n", desc)
			}

			fmt.Fprint(t.errWriter, w)
			if !t.keepInvalid {
				return
			}
		}

		// Ensure JSON is pretty-printed with proper indentation
		json.Indent(w2, finished, "", "    ")

		err = f.Truncate(0)
		if err != nil {
			fmt.Fprintf(t.errWriter, "Error while truncating file before writing %q: %s\n", name, err)
			return
		}

		_, err = io.Copy(f, w2)
		if err != nil {
			fmt.Fprintf(t.errWriter, "Error while writing data to disk %q: %s\n", name, err)
			return
		}

		atLeastOne = true
	}

	// Run all transforms in their own goroutines
	for _, name := range names {
		wg.Add(1)
		go fn(name)
	}

	wg.Wait()
}

func getTransfomers(list string) []MessageTransformer {
	//fmt.Fprintf(t.errWriter, "No transform exists with the name %q\n", name)
	return make([]MessageTransformer, 0)
}
