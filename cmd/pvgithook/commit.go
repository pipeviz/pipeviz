package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/types/semantic"
)

func postCommitHookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post-commit",
		Short: "Processes a git post-commit hook event.",
		Run:   runPostCommit,
	}

	return cmd
}

func runPostCommit(cmd *cobra.Command, args []string) {
	//tgt := cmd.Flags().Lookup("target").Value.String()
	gitcmd := exec.Command("git", "log", "-1", "--pretty=format:'%H%n%an <%ae>%n%ad%n%P%n%s'", "HEAD")
	ret, err := gitcmd.Output()
	if err != nil {
		fmt.Printf("Error returned from git command: %s\n", err)
		os.Exit(1)
	}

	scn := bufio.NewScanner(bytes.NewReader(ret))
	commit := semantic.Commit{}

	if !scn.Scan() {
		fmt.Printf("Output terminated early with err: %s\n", scn.Err())
	}
	commit.Sha1Str = scn.Text()

	if !scn.Scan() {
		fmt.Printf("Output terminated early with err: %s\n", scn.Err())
	}
	commit.Author = scn.Text()

	if !scn.Scan() {
		fmt.Printf("Output terminated early with err: %s\n", scn.Err())
	}
	commit.Date = scn.Text()

	if !scn.Scan() {
		fmt.Printf("Output terminated early with err: %s\n", scn.Err())
	}
	commit.ParentsStr = strings.Split(scn.Text(), " ")

	if !scn.Scan() {
		fmt.Printf("Output terminated early with err: %s\n", scn.Err())
	}
	commit.Subject = scn.Text()

	if scn.Scan() {
		fmt.Printf("More output than expected, err: %s\n", scn.Err())
	}
}
