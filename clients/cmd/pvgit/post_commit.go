package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/types/semantic"
	"gopkg.in/libgit2/git2go.v22"
)

const gitRFC2822 = "Mon Jan 2 2006 15:04:05 -0700"

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
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting cwd:", err)
		os.Exit(1)
	}

	repo, err := git.OpenRepository(cwd + "/.git")
	if err != nil {
		fmt.Printf("Error opening repo at %s: %s", cwd+"/.git", err)
		os.Exit(1)
	}

	head, err := repo.Head()
	if err != nil {
		fmt.Printf("Could not get repo HEAD")
		os.Exit(1)
	}

	commit, err := repo.LookupCommit(head.Target())
	if err != nil {
		fmt.Printf("Could not find commit pointed at by HEAD")
		os.Exit(1)
	}

	authsig := commit.Author()
	cmt := semantic.Commit{
		Sha1:    semantic.Sha1(*commit.Id()),
		Author:  authsig.Name + "<" + authsig.Email + ">",
		Date:    authsig.When.Format(gitRFC2822),
		Subject: commit.Summary(),
	}

	// Parent list is base 0, though 0 is really "first parent"
	for i := uint(0); i < commit.ParentCount(); i++ {
		cmt.ParentsStr = append(cmt.ParentsStr, hex.EncodeToString(commit.ParentId(i)[:]))
	}
}
