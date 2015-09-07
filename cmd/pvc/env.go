package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	gjs "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/schema"
)

var schemaMaster *gjs.Schema

// TODO we just use this as a way to namespace common names
type envCmd struct{}

func envCommand() *cobra.Command {
	ec := envCmd{}
	cmd := &cobra.Command{
		Use:   "env [-n|--no-detect]",
		Short: "Generates a pipeviz message describing an environment.",
		Long:  "Generates a valid message describing an environment from user input, then sends the message to a target pipeviz server.",
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

	// Prep schema to validate the messages as we go
	raw, err := schema.Master()
	if err != nil {
		fmt.Fprintln(w, "WARNING: Failed to open master schema file; pvc cannot validate outgoing messages.")
	}

	schemaMaster, err = gjs.NewSchema(gjs.NewStringLoader(string(raw)))
	if err != nil {
		panic("bad schema...?")
	}

	client := http.Client{Timeout: 5 * time.Second}

	fmt.Fprintln(w, "Generating an environment message...")
	reader := bufio.NewReader(os.Stdin)
MenuLoop:
	for {
		fmt.Fprintf(w, "\n")
		ec.printCurrentState(w, *e)

		var input string
		for {
			fmt.Fprintf(w, "\nSelect a value to edit by number, (p)rint current JSON message, (s)end, or (q)uit: ")
			l, err := fmt.Fscanln(reader, &input)
			if l > 1 || err != nil {
				continue
			}

			switch input {
			case "q", "quit":
				fmt.Fprintf(w, "\nQuitting; message was not sent\n")
				os.Exit(1)
			case "s", "send":
				msg, err := toJSONBytes(*e)
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
					fmt.Printf("Message accepted (HTTP code %v), msgid %v\n", resp.StatusCode, string(bod))
				} else {
					fmt.Printf("Message was rejected with HTTP code %v and message %v\n", resp.StatusCode, string(bod))
				}
				break MenuLoop

			case "p", "print":
				byts, err := toJSONBytes(*e)
				if err != nil {
					fmt.Fprintf(w, "Error while generating JSON for printing: %q", err.Error())
					continue MenuLoop
				}

				var prettied bytes.Buffer
				err = json.Indent(&prettied, byts, "", "    ")
				if err != nil {
					fmt.Fprintf(w, "Error while generating JSON for printing: %q", err.Error())
					continue MenuLoop
				}

				fmt.Fprintf(w, "\nMessage that will be sent to %s:\n", cmd.Flags().Lookup("target").Value)
				prettied.WriteTo(w)
				w.WriteString("\n")

			default:
				num, interr := strconv.Atoi(input)
				if interr != nil {
					continue
				} else if 0 < num && num < 7 {
					switch num {
					case 1:
						collectFQDN(w, reader, e)
					case 2:
						collectIpv4(w, reader, e)
					case 3:
						collectIpv6(w, reader, e)
					case 4:
						collectOS(w, reader, e)
					case 5:
						collectNick(w, reader, e)
					case 6:
						collectProvider(w, reader, e)
					}
					continue MenuLoop
				} else {
					continue
				}
			}
		}
	}
}

func collectFQDN(w io.Writer, r io.Reader, e *interpret.Environment) {
	fmt.Fprintf(w, "\n\nEditing FQDN\nCurrent Value: %q\n", e.Address.Hostname)
	fmt.Fprint(w, "New value: ")

	scn := bufio.NewScanner(r)
	scn.Scan()
	e.Address.Hostname = scn.Text()
}

func collectIpv4(w io.Writer, r io.Reader, e *interpret.Environment) {
	fmt.Fprintf(w, "\n\nEditing IPv4\nCurrent Value: %q\n", e.Address.Ipv4)
	fmt.Fprint(w, "New value: ")

	for {
		var input string
		_, err := fmt.Fscanln(r, &input)
		if err == nil {
			addr := net.ParseIP(input)
			if addr == nil {
				// failed to parse IP, invalid input
				fmt.Fprintf(w, "\nNot a valid IP address.\nNew value: ")
			} else if addr.To4() == nil {
				// not a valid IPv4
				fmt.Fprintf(w, "\nNot a valid IPv4 address.\nNew value: ")
			} else {
				e.Address.Ipv4 = addr.String()
				break
			}
		} else {
			fmt.Fprintf(w, "\nInvalid input.\nNew value: ")
		}
	}
}

func collectIpv6(w io.Writer, r io.Reader, e *interpret.Environment) {
	fmt.Fprintf(w, "\n\nEditing IPv6\nCurrent Value: %q\n", e.Address.Ipv6)
	fmt.Fprint(w, "New value: ")

	for {
		var input string
		_, err := fmt.Fscanln(r, &input)
		if err == nil {
			addr := net.ParseIP(input)
			if addr == nil {
				// failed to parse IP, invalid input
				fmt.Fprintf(w, "\nNot a valid IP address.\nNew value: ")
			} else if addr.To16() == nil {
				// not a valid IPv6
				fmt.Fprintf(w, "\nNot a valid IPv6 address.\nNew value: ")
			} else {
				e.Address.Ipv6 = addr.To16().String()
				break
			}
		} else {
			fmt.Fprintf(w, "\nInvalid input.\nNew value: ")
		}
	}
}

func collectOS(w io.Writer, r io.Reader, e *interpret.Environment) {
	fmt.Fprintf(w, "\n\nEditing OS\nCurrent Value: %q\n", e.OS)
	fmt.Fprint(w, "New value: ")

	scn := bufio.NewScanner(r)
	scn.Scan()
	e.OS = scn.Text()
}

func collectNick(w io.Writer, r io.Reader, e *interpret.Environment) {
	fmt.Fprintf(w, "\n\nEditing Nick\nCurrent Value: %q\n", e.Nick)
	fmt.Fprint(w, "New value: ")

	scn := bufio.NewScanner(r)
	scn.Scan()
	e.Nick = scn.Text()
}

func collectProvider(w io.Writer, r io.Reader, e *interpret.Environment) {
	fmt.Fprintf(w, "\n\nEditing Provider\nCurrent Value: %q\n", e.Provider)
	fmt.Fprint(w, "New value: ")

	scn := bufio.NewScanner(r)
	scn.Scan()
	e.Provider = scn.Text()
}

// Inspects the currently running system to fill in some default values.
func detectEnvDefaults() (e interpret.Environment) {
	var err error
	e.Address.Hostname, err = os.Hostname()
	if err != nil {
		e.Address.Hostname = ""
	}

	e.OS = runtime.GOOS

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
	fmt.Fprintf(w, "  %v. OS: %q\n", n, e.OS)

	n++
	fmt.Fprintf(w, "  %v. Nick: %q\n", n, e.Nick)

	n++
	fmt.Fprintf(w, "  %v. Provider: %q\n", n, e.Provider)

	validateAndPrint(w, e)
}
