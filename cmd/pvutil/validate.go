package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	gjs "github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
	"github.com/pipeviz/pipeviz/schema"
)

const (
	FileReadFail = 1 << iota
	ValidationFail
	ValidationError
)

func validateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <dir>...",
		Short: "Reads JSON message fixtures from a directory and validates them against the master schema.",
		Long:  `Given one or more directories containing a set of JSON pipeviz message fixtures, validates each of those messages against the master schema and reports any failures.`,
		Run:   runValidate,
	}

	return cmd
}

func runValidate(cmd *cobra.Command, args []string) {
	var errors int
	for _, dir := range args {
		fl, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Printf("Failed to read directory '%v' with error %v\n", dir, err)
		}

		for _, f := range fl {
			if match, _ := regexp.MatchString("\\.json$", f.Name()); match && !f.IsDir() {
				src, err := ioutil.ReadFile(dir + "/" + f.Name())
				if err != nil {
					errors |= FileReadFail
					fmt.Printf("Failed to read fixture file %v/%v\n", dir, f.Name())
					continue
				}

				result, err := schema.Master().Validate(gjs.NewStringLoader(string(src)))
				if err != nil {
					errors |= ValidationError
					fmt.Printf("Validation process terminated with errors for %v/%v. Error: \n%v\n", dir, f.Name(), err.Error())
					continue
				}

				if !result.Valid() {
					errors |= ValidationFail
					fmt.Printf("Errors in %v/%v:\n", dir, f.Name())
					for _, desc := range result.Errors() {
						fmt.Printf("\t%s\n", desc)
					}
				} else {
					fmt.Printf("%v/%v successfully validated\n", dir, f.Name())
				}
			}
		}
	}
	os.Exit(errors)
}
