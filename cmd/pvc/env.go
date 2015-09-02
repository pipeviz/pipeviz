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
		e = detectEnvDefaults()
	}

	// Write directly to stdout, at least for now
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

	e.Os = runtime.GOOS

	return e
}

// printMenu prints to stdout a menu showing the current data in the
// message to be generated.
func (ec envCmd) printCurrentState(w io.Writer, e interpret.Environment) {
	fmt.Fprintln(w, "Environment data:")
	var n int

	n++
	//if e.Address.Hostname == "" {
	//fmt.Fprintf(w, "  %v. *FQDN: [empty]\n", n)
	//} else {
	fmt.Fprintf(w, "  %v. FQDN: %q\n", n, e.Address.Hostname)
	//}

	n++
	fmt.Fprintf(w, "  %v. Ipv4: %q\n", n, e.Address.Ipv4)
	n++
	fmt.Fprintf(w, "  %v. Ipv6: %q\n", n, e.Address.Ipv6)

	n++
	fmt.Fprintf(w, "  %v. OS: %q\n", n, e.Os)

	n++
	fmt.Fprintf(w, "  %v. Nickname: %q\n", n, e.Nick)

	n++
	fmt.Fprintf(w, "  %v. Provider: %q\n", n, e.Nick)
}
