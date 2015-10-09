package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/clients/githelp"
	"github.com/tag1consulting/pipeviz/types/semantic"
)

func postCommitHookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook-post-commit",
		Short: "Processes a git post-commit hook event.",
		Run:   runPostCommit,
	}

	return cmd
}

func runPostCommit(cmd *cobra.Command, args []string) {
	repo := getRepoOrExit()

	head, err := repo.Head()
	if err != nil {
		log.Fatalf("Could not get repo HEAD")
	}

	commit, err := repo.LookupCommit(head.Target())
	if err != nil {
		log.Fatalf("Could not find commit pointed at by HEAD")
	}

	ident, err := githelp.GetRepoIdent(repo)
	if err != nil {
		log.Fatalln("Failed to retrieve identifier for repository")
	}

	cmt := commitToSemanticForm(commit, ident)

	// TODO if on a branch, include update to branch pointer
	// TODO if we're not operating on a bare repository (no idea how that could happen), then report the working copy change

	addr, err := githelp.GetTargetAddr(repo)
	// Find the target pipeviz instance from the git config
	if err != nil {
		log.Fatalln("Could not find target address in config:", err)
	}

	msg, err := json.Marshal(map[string]interface{}{
		"commits": []semantic.Commit{cmt},
	})
	if err != nil {
		log.Fatalln("JSON encoding failed with err:", err)
	}

	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(addr, "application/json", bytes.NewReader(msg))
	if err != nil {
		log.Fatalln("Error on sending to server:", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Fatalln("Message rejected by pipeviz server")
	}
}
