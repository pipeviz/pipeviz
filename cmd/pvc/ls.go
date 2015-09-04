package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	gjs "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/schema"
)

// TODO we just use this as a way to namespace common names
type lsCmd struct{}

func lsCommand() *cobra.Command {
	lsc := lsCmd{}
	cmd := &cobra.Command{
		Use:   "ls [-n|--no-detect]",
		Short: "Generates a pipeviz message describing a logic state.",
		Long:  "Generates a valid message describing a logic state from user input, then sends the message to a target pipeviz server.",
		Run:   lsc.runGenLS,
	}

	var nodetect bool
	cmd.Flags().BoolVarP(&nodetect, "no-detect", "n", false, "Skip automated detection of suggested values.")

	return cmd
}

// runGenLS is the main entry point for running the logic state-generating
// ls subcommand.
func (lsc lsCmd) runGenLS(cmd *cobra.Command, args []string) {
	ls := &interpret.LogicState{}

	if !cmd.Flags().Lookup("no-detect").Changed {
		*ls = lsc.detectDefaults()
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

	fmt.Fprintln(w, "Generating a logic state message...")
	reader := bufio.NewReader(os.Stdin)
MenuLoop:
	for {
		fmt.Fprintf(w, "\n")
		lsc.printCurrentState(w, *ls)

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
				msg, err := toJSONBytes(*ls)
				if err != nil {
					log.Fatalf("\nFailed to marshal JSON of logic state object, no message sent\n")
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
				byts, err := toJSONBytes(*ls)
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
				} else if 0 < num && num < 10 {
					switch num {
					case 1:
						lsc.collectPath(w, reader, ls)
					case 2:
						lsc.collectHostFQDN(w, reader, ls)
					case 3:
						lsc.collectHostNick(w, reader, ls)
					case 4:
						lsc.collectCommit(w, reader, ls)
					case 5:
						lsc.collectVersion(w, reader, ls)
					case 6:
						lsc.collectSemver(w, reader, ls)
					case 7:
						lsc.collectLgroup(w, reader, ls)
					case 8:
						lsc.collectNick(w, reader, ls)
					case 9:
						lsc.collectType(w, reader, ls)
					}
					continue MenuLoop
				} else {
					continue
				}
			}
		}
	}
}

func (lsc lsCmd) collectPath(w io.Writer, r io.Reader, ls *interpret.LogicState) {
	fmt.Fprintf(w, "\n\nEditing Path\nCurrent Value: %q\n", ls.Path)
	fmt.Fprint(w, "\nNew value: ")

	scn := bufio.NewScanner(r)
	for {
		scn.Scan()
		path := scn.Text()

		f, err := os.Open(filepath.Clean(path))
		if err == nil {
			ls.Path = f.Name()

			stat, _ := f.Stat()
			if stat.IsDir() {
				ls.Type = "code"
			} else if strings.HasSuffix(f.Name(), ".so") {
				ls.Type = "library"
			} else {
				ls.Type = "binary"
			}

			break
		} else {
			fmt.Fprintf(w, "\n%s does not exist.\n\nNew value: ", path)
		}
	}
}

func (lsc lsCmd) collectHostFQDN(w io.Writer, r io.Reader, ls *interpret.LogicState) {
	fmt.Fprintf(w, "\n\nEditing Host FQDN\nCurrent Value: %q\n", ls.Environment.Address.Hostname)
	fmt.Fprint(w, "\nNew value: ")

	scn := bufio.NewScanner(r)
	scn.Scan()
	ls.Environment.Address.Hostname = scn.Text()
}

func (lsc lsCmd) collectHostNick(w io.Writer, r io.Reader, ls *interpret.LogicState) {
	fmt.Fprintf(w, "\n\nEditing Host Nick\nCurrent Value: %q\n", ls.Environment.Nick)
	fmt.Fprint(w, "\nNew value: ")

	scn := bufio.NewScanner(r)
	scn.Scan()
	ls.Environment.Nick = scn.Text()
}

func (lsc lsCmd) collectCommit(w io.Writer, r io.Reader, ls *interpret.LogicState) {
	fmt.Fprintf(w, "\n\nEditing Git commit\nCurrent Value: %q\n", ls.ID.CommitStr)
	fmt.Fprint(w, "\nNew value: ")

	for {
		scn := bufio.NewScanner(r)
		scn.Scan()

		commit := scn.Text()
		byts, err := hex.DecodeString(commit)
		if err != nil || len(byts) != 20 {
			fmt.Fprintf(w, "\n%s is not a valid Git commit.\n\nNew value: ", commit)
		} else {
			ls.ID.CommitStr = commit
			break
		}
	}
}

func (lsc lsCmd) collectVersion(w io.Writer, r io.Reader, ls *interpret.LogicState) {
	fmt.Fprintf(w, "\n\nEditing Version\nCurrent Value: %q\n", ls.ID.Version)
	fmt.Fprint(w, "\nNew value: ")

	scn := bufio.NewScanner(r)
	scn.Scan()
	ls.ID.Version = scn.Text()
}

func (lsc lsCmd) collectSemver(w io.Writer, r io.Reader, ls *interpret.LogicState) {
	fmt.Fprintf(w, "\n\nEditing Semver\nCurrent Value: %q\n", ls.ID.Semver)
	fmt.Fprint(w, "\nNew value: ")

	scn := bufio.NewScanner(r)
	scn.Scan()
	ls.ID.Semver = scn.Text()
}

func (lsc lsCmd) collectLgroup(w io.Writer, r io.Reader, ls *interpret.LogicState) {
	fmt.Fprintf(w, "\n\nEditing Logical group\nCurrent Value: %q\n", ls.Lgroup)
	fmt.Fprint(w, "\nNew value: ")

	scn := bufio.NewScanner(r)
	scn.Scan()
	ls.Lgroup = scn.Text()
}

func (lsc lsCmd) collectNick(w io.Writer, r io.Reader, ls *interpret.LogicState) {
	fmt.Fprintf(w, "\n\nEditing Logical group\nCurrent Value: %q\n", ls.Nick)
	fmt.Fprint(w, "\nNew value: ")

	scn := bufio.NewScanner(r)
	scn.Scan()
	ls.Nick = scn.Text()
}

func (lsc lsCmd) collectType(w io.Writer, r io.Reader, ls *interpret.LogicState) {
	fmt.Fprintf(w, "\n\nEditing Logical group\nCurrent Value: %q\n", ls.Type)
	fmt.Fprint(w, "\nNew value: ")

	scn := bufio.NewScanner(r)
	scn.Scan()
	ls.Type = scn.Text()
}

func (lsc lsCmd) detectDefaults() (ls interpret.LogicState) {
	var err error
	ls.Environment.Address.Hostname, err = os.Hostname()
	if err != nil {
		ls.Environment.Address.Hostname = ""
	}

	return ls
}

// printMenu prints to stdout a menu showing the current data in the
// message to be generated.
func (lsc lsCmd) printCurrentState(w io.Writer, e interpret.LogicState) {
	fmt.Fprintln(w, "Logic state data:")
	var n int

	n++
	fmt.Fprintf(w, "  %v. Path: %q\n", n, e.Path)

	n++
	fmt.Fprintf(w, "  %v. Host FQDN: %q\n", n, e.Environment.Address.Hostname)
	n++
	fmt.Fprintf(w, "  %v. Host Nick: %q\n", n, e.Environment.Nick)

	n++
	fmt.Fprintf(w, "  %v. Git commit: %q\n", n, e.ID.CommitStr)
	n++
	fmt.Fprintf(w, "  %v. Version: %q\n", n, e.ID.Version)
	n++
	fmt.Fprintf(w, "  %v. Semver:  %q\n", n, e.ID.Semver)

	n++
	fmt.Fprintf(w, "  %v. Logical group:  %q\n", n, e.Lgroup)

	n++
	fmt.Fprintf(w, "  %v. Nick: %q\n", n, e.Nick)
	n++
	fmt.Fprintf(w, "  %v. Type: %q\n", n, e.Type)

	validateAndPrint(w, e)
}
