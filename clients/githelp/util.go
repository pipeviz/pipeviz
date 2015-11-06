/*
Package githelp contains shared helpers for git-related operations done by clients.

It is dependent on libgit2, via git2go. That means CGo.
*/
package githelp

import (
	"errors"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/gopkg.in/libgit2/git2go.v22"
)

// The (version of?) the RFC2822 format git uses to format its date output,
// for use as a layout to pass to Go's time.Format and time.Parse
const GitDateFormat = "Mon Jan 2 2006 15:04:05 -0700"

// GetTargetAddr gets the address for the target pipeviz instance registered
// in a git repository's config file.
func GetTargetAddr(repo *git.Repository) (string, error) {
	cfg, err := repo.Config()
	if err != nil {
		return "", err
	}

	return cfg.LookupString("pipeviz.target")
}

// GetRepoIdent returns a string identifier that uniquely identifies
// this repository.
func GetRepoIdent(repo *git.Repository) (r string, err error) {
	// FIXME can't pretend this works for much longer
	var remote *git.Remote

	// try origin & upstream, return that if it's there
	if remote, err = repo.LookupRemote("origin"); err == nil {
		r = remote.Url()
		return
	} else if remote, err = repo.LookupRemote("upstream"); err == nil {
		r = remote.Url()
		return
	}

	var remotes []string
	if remotes, err = repo.ListRemotes(); err != nil {
		return
	}

	// otherwise, just grab the first in the list
	for _, rname := range remotes {
		remote, err = repo.LookupRemote(rname)
		if err != nil {
			return "", err
		}
		r = remote.Url()
		return
	}

	return "", errors.New("No remotes found")
}
