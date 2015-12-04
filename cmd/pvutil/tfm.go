package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/pipeviz/pipeviz/message/mtf"
	"github.com/pipeviz/pipeviz/schema"
	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
)

type tfm struct {
	list, keepInvalid, quiet bool
	transforms               string
	errWriter                *log.Logger
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
	cmd.Flags().BoolVarP(&t.keepInvalid, "keep-invalid", "k", false, "Keep non-validating results. Normally non-valid results will be discarded (no output if stdin, no write if files); this bypasses that behavior. Note that malformed JSON will *never* be used, regardless of this flag.")
	cmd.Flags().BoolVarP(&t.quiet, "quiet", "q", false, "Suppress error and informational output.")

	return cmd
}

// Run executes the tfm command.
func (t *tfm) Run(cmd *cobra.Command, args []string) {
	if t.quiet {
		t.errWriter = log.New(ioutil.Discard, "", 0)
	} else {
		t.errWriter = log.New(os.Stderr, "", 0)
	}

	if t.list {
		fmt.Printf("%s\n", strings.Join(mtf.List(), "\n"))
		os.Exit(0)
	}

	if t.transforms == "" {
		t.errWriter.Printf("Must specify at least one transform to apply\n")
		os.Exit(1)
	}

	tf, _ := mtf.Get(strings.Split(t.transforms, ",")...)
	if len(tf) == 0 {
		t.errWriter.Printf("None of the specified transforms could be found. See `pvutil tfm -l` for a list.")
		os.Exit(1)
	}

	// Figure out if we have something from stdin by checking stdin's fd type.
	// Not sure how cross-platform this is, but works on darwin and linux, at least.
	stat, _ := os.Stdin.Stat()
	hasStdin := (stat.Mode() & os.ModeCharDevice) == 0

	if hasStdin {
		if len(args) != 0 {
			t.errWriter.Printf("Cannot operate on both stdin and files\n")
			os.Exit(1)
		}
		t.runStdin(cmd, tf)
	} else {
		if len(args) == 0 {
			t.errWriter.Printf("Must pass either a set of target files, or some data on stdin\n")
			os.Exit(1)
		}

		t.runFiles(cmd, tf, args)
	}
}

func (t *tfm) runStdin(cmd *cobra.Command, tl mtf.TransformList) {

}

func (t *tfm) runFiles(cmd *cobra.Command, tl mtf.TransformList, names []string) {
	var atLeastOne bool
	var wg sync.WaitGroup

	fn := func(name string) {
		defer wg.Done()
		f, err := os.OpenFile(name, os.O_RDWR, 0666)
		defer f.Close()
		if err != nil {
			t.errWriter.Printf("Error while opening file %q: %s\n", name, err)
			return
		}

		bb, changed, err := tl.Transform(f)
		if err != nil {
			t.errWriter.Printf(err.Error())
			return
		}
		if !changed {
			t.errWriter.Printf("No changes resulted from applying transforms to %q\n", name)
			return
		}

		result, err := schema.Master().Validate(gojsonschema.NewStringLoader(string(bb)))
		if err != nil {
			t.errWriter.Printf("Error while validating final transformed result: %s\n", err)
			return
		}

		if !result.Valid() {
			// Buffer so that there's no interleaving of printed output
			w := bytes.NewBuffer(make([]byte, 0))

			fmt.Fprintf(w, "Errors encountered while validating %q:\n", name)
			for _, desc := range result.Errors() {
				fmt.Fprintf(w, "%s\n", desc)
			}

			t.errWriter.Print(w)
			if !t.keepInvalid {
				return
			}
		}

		final := new(bytes.Buffer)
		// Ensure JSON is pretty-printed with proper indentation
		json.Indent(final, bb, "", "    ")

		err = f.Truncate(0)
		if err != nil {
			t.errWriter.Printf("Error while truncating file before writing %q: %s\n", name, err)
			return
		}

		_, err = io.Copy(f, final)
		if err != nil {
			t.errWriter.Printf("Error while writing data to disk %q: %s\n", name, err)
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
