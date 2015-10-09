package main

import (
	"encoding/hex"
	"log"
	"os"

	"github.com/tag1consulting/pipeviz/clients/githelp"
	"github.com/tag1consulting/pipeviz/types/semantic"

	"gopkg.in/libgit2/git2go.v22"
)

func getRepoOrExit(paths ...string) *git.Repository {
	var path string
	var err error

	if len(paths) > 0 {
		path = paths[0]
	} else if path, err = os.Getwd(); err != nil {
		log.Fatalln("Error getting cwd:", err)
	}

	repostr, err := git.Discover(path, false, []string{"/"})
	if err != nil {
		log.Fatalln(err)
	}

	repo, err := git.OpenRepository(repostr)
	if err != nil {
		log.Fatalf("Error opening repo at %s: %s", repostr+"/.git", err)
	}

	return repo
}

func commitToSemanticForm(in *git.Commit, ident string) (out semantic.Commit) {
	authsig := in.Author()
	out = semantic.Commit{
		Sha1Str:    hex.EncodeToString(in.Id()[:]),
		Author:     authsig.Name + "<" + authsig.Email + ">",
		Date:       authsig.When.Format(githelp.GitDateFormat),
		Subject:    in.Summary(),
		Repository: ident,
	}

	// Parent list is base 0, though 0 is really "first parent"
	for i := uint(0); i < in.ParentCount(); i++ {
		out.ParentsStr = append(out.ParentsStr, hex.EncodeToString(in.ParentId(i)[:]))
	}

	return
}
