package main

import (
	"log"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/clients/githelp"
	"github.com/tag1consulting/pipeviz/ingest"
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

	m := new(ingest.Message)
	m.Add(commitToSemanticForm(commit, ident))

	recordHead(m, repo)
	sendMapToPipeviz(m, repo)
}
