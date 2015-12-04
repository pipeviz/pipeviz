package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

	tf, missing := mtf.Get(strings.Split(t.transforms, ",")...)
	if len(tf) == 0 {
		t.errWriter.Printf("None of the requested transforms could be found. See `pvutil tfm -l` for a list.")
		os.Exit(1)
	}
	if len(missing) != 0 {
		t.errWriter.Printf("The following requested transforms could not be found: %s\n", strings.Join(missing, ","))
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
		t.runStdin(tf)
	} else {
		if len(args) == 0 {
			t.errWriter.Printf("Must pass either a set of target files, or some data on stdin\n")
			os.Exit(1)
		}

		t.runFiles(tf, args)
	}
}

func (t *tfm) runStdin(tl mtf.TransformList) {
	contents, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		t.errWriter.Printf("Error while reading data from stdin: %s\n", err)
	}

	bb, _, err := t.transformAndValidate(contents, tl, "")
	if err != nil {
		os.Exit(1)
	}

	final := new(bytes.Buffer)
	// Try to ensure JSON is pretty-printed with proper indentation
	err = json.Indent(final, bb, "", "    ")
	if err != nil {
		final = bytes.NewBuffer(bb)
	}

	final.WriteTo(os.Stdout)
}

func (t *tfm) runFiles(tl mtf.TransformList, names []string) {
	var updated []string
	var wg sync.WaitGroup

	fn := func(name string) {
		f, err := os.OpenFile(name, os.O_RDWR, 0666)
		defer wg.Done()
		defer f.Close()

		if err != nil {
			t.errWriter.Printf("Error while opening file %q: %s\n", name, err)
			return
		}

		contents, err := ioutil.ReadAll(f)
		if err != nil {
			t.errWriter.Printf("Error while reading from file %q: %s\n", name, err)
		}
		bb, changed, err := t.transformAndValidate(contents, tl, name)
		if err != nil || !changed {
			return
		}

		final := new(bytes.Buffer)
		// Try to ensure JSON is pretty-printed with proper indentation
		err = json.Indent(final, bb, "", "    ")
		if err != nil {
			final = bytes.NewBuffer(bb)
		}

		// Handling errors for these cases is just overkill silliness
		_ = f.Truncate(0)
		_, _ = f.Seek(0, 0)

		_, err = final.WriteTo(f)
		if err != nil {
			t.errWriter.Printf("Error while writing data to disk %q: %s\n", name, err)
			return
		}

		updated = append(updated, name)
	}

	// Run all transforms in their own goroutines
	for _, name := range names {
		wg.Add(1)
		go fn(name)
	}

	wg.Wait()

	if len(updated) == 0 {
		t.errWriter.Println("\nNo message files were updated.")
		os.Exit(1)
	} else {
		t.errWriter.Println("\nThe following files were updated:")
		for _, name := range updated {
			t.errWriter.Printf("\t%s\n", name)
		}
	}
}

func (t *tfm) transformAndValidate(msg []byte, tl mtf.TransformList, name string) (final []byte, changed bool, err error) {
	final, changed, err = tl.Transform(msg)
	if err != nil {
		t.errWriter.Printf(err.Error())
		return
	}
	if !changed {
		// if name == "" {
		// 	t.errWriter.Printf("No changes resulted from applying transforms\n")
		// } else {
		// 	t.errWriter.Printf("No changes resulted from applying transforms to %q\n", name)
		// }
		return
	}

	result, err := schema.Master().Validate(gojsonschema.NewStringLoader(string(final)))
	if err != nil {
		t.errWriter.Printf("Error while validating final transformed result: %s\n", err)
		return
	}

	if !result.Valid() {
		// Buffer so that there's no interleaving of printed output
		w := bytes.NewBuffer(make([]byte, 0))

		if name == "" {
			fmt.Fprintf(w, "Errors encountered during validation:\n")
		} else {
			fmt.Fprintf(w, "Errors encountered while validating %q:\n", name)
		}

		for _, desc := range result.Errors() {
			fmt.Fprintf(w, "\t%s\n", desc)
		}

		t.errWriter.Print(w)
		if !t.keepInvalid {
			return final, changed, errors.New("message failed validation, skipping")
		}
	}

	return
}
