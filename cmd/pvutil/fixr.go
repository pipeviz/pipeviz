package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"
)

func fixrCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fixr [-p|--prompt] ([-t|--target] <server>) <dir>",
		Short: "Reads JSON message fixtures from a directory and sends them to a pipeviz daemon.",
		Long:  `Given a directory containing a set of JSON pipeviz message fixtures, the fixr command reads those fixtures in lexical order, sending them to the target server in order.`,
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

func runFixrInteractive(cmd *cobra.Command, files []fileInfo) {

}

func runFixrBatch(cmd *cobra.Command, files []fileInfo) {
	client := http.Client{Timeout: 5 * time.Second}
	for _, f := range files {
		src, err := ioutil.ReadFile(f.d + "/" + f.f.Name())
		if err != nil {
			log.Printf("Failed to read file '%v/%v' with error %v\n", f.d, f.f.Name(), err)
			continue
		}

		resp, err := client.Post(cmd.Flags().Lookup("target").Value.String(), "application/json", bytes.NewReader(src))
		log.Printf("Sent %v/%v...", f.d, f.f.Name())
		if err != nil {
			log.Print(err)
		}

		bod, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print(err)
		}
		resp.Body.Close()

		log.Printf("%v, msgid %v", resp.StatusCode, string(bod))
	}
}
