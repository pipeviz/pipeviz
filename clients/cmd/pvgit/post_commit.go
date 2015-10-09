package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/clients/githelp"
	"github.com/tag1consulting/pipeviz/types/semantic"
	"gopkg.in/libgit2/git2go.v22"
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
	//tgt := cmd.Flags().Lookup("target").Value.String()
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Error getting cwd:", err)
	}

	repostr, err := git.Discover(cwd, false, []string{"/"})
	if err != nil {
		log.Fatalln(err)
	}

	repo, err := git.OpenRepository(repostr)
	if err != nil {
		log.Fatalf("Error opening repo at %s: %s", cwd+"/.git", err)
	}

	head, err := repo.Head()
	if err != nil {
		log.Fatalf("Could not get repo HEAD")
	}

	commit, err := repo.LookupCommit(head.Target())
	if err != nil {
		log.Fatalf("Could not find commit pointed at by HEAD")
	}

	authsig := commit.Author()
	cmt := semantic.Commit{
		Sha1Str: hex.EncodeToString(commit.Id()[:]),
		Author:  authsig.Name + "<" + authsig.Email + ">",
		Date:    authsig.When.Format(githelp.GitDateFormat),
		Subject: commit.Summary(),
	}

	ident, err := githelp.GetRepoIdent(repo)
	if err != nil {
		log.Fatalln("Failed to retrieve identifier for repository")
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
