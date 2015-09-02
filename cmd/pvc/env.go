package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

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
	e := &interpret.Environment{}

	if !cmd.Flags().Lookup("no-detect").Changed {
		*e = detectEnvDefaults()
	}

	// Write directly to stdout, at least for now
	w := os.Stdout

	client := http.Client{Timeout: 5 * time.Second}

	fmt.Fprintln(w, "Generating an environment message...")
	reader := bufio.NewReader(os.Stdin)
MenuLoop:
	for {
		fmt.Fprintf(w, "\n")
		ec.printCurrentState(w, *e)
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "Select a value to edit by number, (s)end, or (q)uit: ")

		var input string
		for {
			l, err := fmt.Fscanln(reader, &input)
			if l > 1 || err != nil {
				continue
			}

			switch input {
			case "q", "quit":
				fmt.Fprintf(w, "\nQuitting; message was not sent\n")
				os.Exit(1)
			case "s", "send":
				msg, err := json.Marshal(e)
				if err != nil {
					log.Fatalf("\nFailed to marshal JSON of environment object, no message sent\n")
				}

				resp, err := client.Post(cmd.Flags().Lookup("target").Value.String(), "application/json", bytes.NewReader(msg))
				if err != nil {
					log.Fatalf(err.Error())
				}

				bod, err := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					log.Fatalf(err.Error())
				}

				if resp.StatusCode >= 200 && resp.StatusCode <= 300 {
					fmt.Printf("%v, msgid %v\n", resp.StatusCode, string(bod))
				} else {
					fmt.Printf("Message was rejected with code %v and message %v\n", resp.StatusCode, string(bod))
				}

			default:
				num, interr := strconv.Atoi(input)
				if interr != nil {
					continue
				} else if 0 < num && num < 7 {
					switch num {
					case 1:
						collectEnv(w, reader, e)
					case 2:
					case 3:
					case 4:
					case 5:
					case 6:
					}
					continue MenuLoop
				} else {
					continue
				}
			}
		}
	}
}

func collectEnv(w io.Writer, r io.Reader, e *interpret.Environment) {
	fmt.Fprintf(w, "\n\nEditing FQDN\nCurrent Value: %q\n", e.Address.Hostname)
	fmt.Fprint(w, "New value: ")

	for {
		var input string
		_, err := fmt.Fscanln(r, &input)
		if err == nil {
			e.Address.Hostname = input
			break
		}

		fmt.Fprintf(w, "\nInvalid input. New value: ")
	}
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
