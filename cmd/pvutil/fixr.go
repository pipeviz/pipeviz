package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
)

func fixrCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fixr [-p|--prompt] ([-t|--target] <server>) <dir>...",
		Short: "Reads JSON message fixtures from a directory and sends them to a pipeviz daemon.",
		Long:  `Given one or more directories containing a set of JSON pipeviz message fixtures, the fixr command reads those fixtures in lexical order, sending them to the target server in order.`,
		Run:   runFixr,
	}

	var prompt bool
	var target string
	cmd.Flags().BoolVarP(&prompt, "prompt", "p", false, "Interactively prompts the user to send or skip each message.")
	cmd.Flags().StringVarP(&target, "target", "t", "http://localhost:2309", "Address of the target pipeviz daemon. Default to http://localhost:2309")

	return cmd
}

// Holds dir and fileinfo together
type fileInfo struct {
	d string
	f os.FileInfo
}

// Main entry point for running fixr subcommand.
func runFixr(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatalf("Must provide at least one directory argument to fixr.")
	}
	files := make([]fileInfo, 0)
	for _, dir := range args {
		fl, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Fatalf("Failed to read directory '%v' with error %v\n", dir, err)
		}

		for _, f := range fl {
			if match, _ := regexp.MatchString("\\.json$", f.Name()); match && !f.IsDir() {
				files = append(files, fileInfo{d: dir, f: f})
			}
		}
	}

	if cmd.Flags().Lookup("prompt").Changed {
		runFixrInteractive(cmd, files)
	} else {
		runFixrBatch(cmd, files)
	}
}

// Performs the fixr run with user interaction, confirming for each message.
func runFixrInteractive(cmd *cobra.Command, files []fileInfo) {
	client := http.Client{Timeout: 5 * time.Second}
FileLoop:
	for k, f := range files {
		src, err := ioutil.ReadFile(f.d + "/" + f.f.Name())
		if err != nil {
			erro.Printf("Failed to read file '%v/%v' with error %v\n", f.d, f.f.Name(), err)
			continue
		}
		fmt.Printf("\nMessage #%v (%v/%v) contents:\n%s", k+1, f.d, f.f.Name(), src)

		for {
			fmt.Printf("Message #%v: (S)end, s(k)ip, or (q)uit? ", k+1)
			reader := bufio.NewReader(os.Stdin)

			raw, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("Bad input, terminating fixr\n")
			}

			input := strings.Split(strings.Trim(raw, " \n"), " ")
			if len(input) < 1 {
				log.Fatalf("TODO huh how would this happen\n")
			}
			switch input[0] {
			case "k", "skip":
				continue FileLoop
			case "", "s", "send":
				resp, err := client.Post(cmd.Flags().Lookup("target").Value.String(), "application/json", bytes.NewReader(src))
				fmt.Printf("Sending %v/%v...", f.d, f.f.Name())
				if err != nil {
					log.Fatalf(err.Error())
				}

				bod, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					erro.Println(err)
				}
				resp.Body.Close()

				fmt.Printf("%v, msgid %v\n", resp.StatusCode, string(bod))
				continue FileLoop
			case "q", "quit":
				fmt.Printf("Quitting...\n")
				os.Exit(1)
			}
		}
	}
}

// Performs the fixr run in a batch, sending all fixtures found immediately and without confirmation.
func runFixrBatch(cmd *cobra.Command, files []fileInfo) {
	client := http.Client{Timeout: 5 * time.Second}
	for _, f := range files {
		src, err := ioutil.ReadFile(f.d + "/" + f.f.Name())
		if err != nil {
			erro.Printf("Failed to read file '%v/%v' with error %v\n", f.d, f.f.Name(), err)
			continue
		}

		resp, err := client.Post(cmd.Flags().Lookup("target").Value.String(), "application/json", bytes.NewReader(src))
		fmt.Printf("Sending %v/%v...", f.d, f.f.Name())
		if err != nil {
			log.Fatalln(err)
		}

		bod, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			erro.Println(err)
		}
		resp.Body.Close()

		fmt.Printf("%v, msgid %v\n", resp.StatusCode, string(bod))
	}
}
