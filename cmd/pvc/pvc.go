package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
	"github.com/tag1consulting/pipeviz/interpret"
)

type menuLevel interface {
	Info() ([]byte, error)
	Prompt() ([]byte, error)
	Accept(string) error
	Next(*cliRunner) *cliRunner
}

// cliRunner coordinates control over and interaction with a level
// of interaction in the UI
type cliRunner struct {
	parent *cliRunner
	obj    menuLevel
	w      io.Writer
}

func main() {
	root := &cobra.Command{Use: "pvc"}
	root.AddCommand(envCommand())
	root.AddCommand(lsCommand())

	var target string
	root.PersistentFlags().StringVarP(&target, "target", "t", "http://localhost:2309", "Address of the target pipeviz daemon.")

	root.Execute()
}

// wrapForJSON converts data into a map that will serialize
// appropriate pipeviz message JSON.
func wrapForJSON(v interface{}) map[string]interface{} {
	m := make(map[string]interface{})

	switch obj := v.(type) {
	case interpret.Environment:
		m["environments"] = []interpret.Environment{obj}
	case interpret.LogicState:
		m["logic-states"] = []interpret.LogicState{obj}
	}

	return m
}

func toJSONBytes(v interface{}) ([]byte, error) {
	// Convert the data to a map that will write out the correct JSON
	m := wrapForJSON(v)

	msg, err := json.Marshal(m)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("\nError while marshaling data to JSON for validation: %s\n", err.Error()))
	}

	return msg, nil
}

func validateAndPrint(w io.Writer, v interface{}) {
	msg, err := toJSONBytes(v)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	// Validate the current state of the message
	result, err := schemaMaster.Validate(gojsonschema.NewStringLoader(string(msg)))
	if err != nil {
		fmt.Fprintf(w, "\nError while attempting to validate data: %s\n", err.Error())
		return
	}
	if !result.Valid() {
		fmt.Fprintln(w, "\nAs it stands now, the data will fail validation if sent to a pipeviz server. Errors:")
		for _, desc := range result.Errors() {
			fmt.Fprintf(w, "\t%s\n", desc)
		}
	}
}

func runCreate(cmd *cobra.Command, args []string) {
	// Create the root runner
	//cr := &cliRunner{
	//w: os.Stdout,
	//}
}

type mainMenu struct {
}

func (m *mainMenu) Info() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("Which type of state would you like to describe to pipeviz: ")

	return b.Bytes(), nil
}
