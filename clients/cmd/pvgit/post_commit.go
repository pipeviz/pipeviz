package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/clients/githelp"
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

	repostr, err := git.Discover(cwd, false, []string{"/"})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	repo, err := git.OpenRepository(repostr)
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
		Sha1Str: hex.EncodeToString(commit.Id()[:]),
		Author:  authsig.Name + "<" + authsig.Email + ">",
		Date:    authsig.When.Format(gitRFC2822),
		Subject: commit.Summary(),
	}

	ident, err := githelp.GetRepoIdent(repo)
	if err != nil {
		fmt.Println("Failed to retrieve identifier for repository")
		os.Exit(1)
	}
	cmt.Repository = ident

	// Parent list is base 0, though 0 is really "first parent"
	for i := uint(0); i < commit.ParentCount(); i++ {
		cmt.ParentsStr = append(cmt.ParentsStr, hex.EncodeToString(commit.ParentId(i)[:]))
	}

	// TODO if on a branch, include update to branch pointer
	// TODO if we're not operating on a bare repository (no idea how that could happen), then report the working copy change

	addr, err := githelp.GetTargetAddr(repo)
	// Find the target pipeviz instance from the git config
	if err != nil {
		fmt.Println("Could not find target address in config:", err)
		os.Exit(1)
	}

	msg, err := json.Marshal(map[string]interface{}{
		"commits": []semantic.Commit{cmt},
	})
	if err != nil {
		fmt.Println("JSON encoding failed with err:", err)
	}

	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Post(addr, "application/json", bytes.NewReader(msg))
	if err != nil {
		fmt.Println("Error on sending to server:", err)
		os.Exit(2)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Println("Message rejected by pipeviz server")
		os.Exit(2)
	}
}
