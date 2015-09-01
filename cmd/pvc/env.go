package main

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/interpret"
)

// TODO we just use this as a way to namespace common names
type envCmd struct{}

func envCommand() *cobra.Command {
	ec := envCmd{}
	cmd := &cobra.Command{
		Use:   "env [-n|--no-detect]",
		Short: "Generates a pipeviz message describing an environment.",
		Long:  "Generates a valid pipeviz message describing an environment. Sends the message to a target server, if one is provided; otherwise prints the message and exits.",
		Run:   ec.runGenEnv,
	}

	var nodetect bool
	cmd.Flags().BoolVarP(&nodetect, "no-detect", "n", false, "Skip automated detection of suggested values.")

	return cmd
}

// runGenEnv is the main entry point for running the environment-generating
// env subcommand.
func (ec envCmd) runGenEnv(cmd *cobra.Command, args []string) {
	var e interpret.Environment

	if !cmd.Flags().Lookup("no-detect").Changed {
		e = interpret.Environment{}
	}

	// For now, write directly to stdout
	w := os.Stdout

	fmt.Println("Generating an environment message...")
	ec.printCurrentState(w, e)
	//reader := bufio.NewReader(os.Stdin)
}

// Inspects the currently running system to fill in some default values.
func detectEnvDefaults() (e interpret.Environment) {
	var err error
	e.Address.Hostname, err = os.Hostname()
	if err != nil {
		e.Address.Hostname = ""
	}

	// Relying on the running program, not the actual system, but...
	e.Os = runtime.GOOS

	return e
}

// printMenu prints to stdout a menu showing the current data in the
// message to be generated.
func (ec envCmd) printCurrentState(w io.Writer, e interpret.Environment) {
	fmt.Println()
	if e.Address.Hostname == "" {
		fmt.Print("  1. FQDN: [empty] ")
	}
}
