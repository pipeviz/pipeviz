package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/clients/githelp"
	"gopkg.in/libgit2/git2go.v22"
)

type instrumentCmd struct {
	ncom, ncheck, history, refs bool
	target                      string
}

func instrumentCommand() *cobra.Command {
	ic := instrumentCmd{}
	cmd := &cobra.Command{
		Use:   "instrument [--no-post-commit] [--no-post-checkout] ([--history] | [--refs]) [-t|--target=<addr>] <repository>",
		Short: "Instruments a git repository to talk with a pipeviz server.",
		Long:  "Instruments a git repository by setting up hook scripts to emit data about the state of the local repository to a pipeviz server. If no path to a repository is provided, it will search for one from the working directory.",
		Run:   ic.run,
	}

	cmd.Flags().BoolVar(&ic.ncom, "no-post-commit", false, "Do not configure a post-commit hook.")
	cmd.Flags().BoolVar(&ic.ncheck, "no-post-checkout", false, "Do not configure a post-checkout hook.")
	cmd.Flags().BoolVar(&ic.history, "history", false, "Send all known history from the repository to the server. Encompasses --refs.")
	cmd.Flags().BoolVar(&ic.refs, "refs", false, "Sends the position of all local refs to the server.")
	cmd.Flags().StringVarP(&ic.target, "target", "t", "", "Address of the target pipeviz daemon. Required if an address is not already set in git config.")

	return cmd
}

const (
	postCommit = `
#!/bin/sh
{{ GoBin }} hook-post-commit
`
	postCheckout = `
#!/bin/sh
{{ GoBin }} hook-post-checkout
`
)

func (ic instrumentCmd) run(cmd *cobra.Command, args []string) {
	var path string
	var err error

	if len(args) > 0 {
		path = args[0]
	} else if path, err = os.Getwd(); err != nil {
		fmt.Println("Error getting cwd:", err)
		os.Exit(1)
	}

	repostr, err := git.Discover(path, false, []string{"/"})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	repo, err := git.OpenRepository(repostr)
	if err != nil {
		fmt.Printf("Error opening repo at %s: %s", repostr+"/.git", err)
		os.Exit(1)
	}

	if ic.target == "" {
		ic.target, err = githelp.GetTargetAddr(repo)
		if err != nil {
			fmt.Printf("No pipeviz server target provided, and one is not already registered in git's config.")
			os.Exit(1)
		}
	} else {
		cfg, err := repo.Config()
		if err != nil {
			fmt.Println("Error attempting to retrieve git config", err)
			os.Exit(1)
		}

		err = cfg.SetString("pipeviz.target", ic.target)
		if err != nil {
			fmt.Println("Error while writing pipeviz target to config", err)
			os.Exit(1)
		}
		fmt.Println("Set target pipeviz server to ", ic.target)
	}

	// Write the post-commit hook, unless user said no
	if !ic.ncom {
		f, err := os.Create(repo.Path() + "/hooks/post-commit")
		if err != nil {
			fmt.Println("Error while attempting to open post-commit hook for writing:", err)
			os.Exit(1)
		}
		_, err = f.WriteString(strings.Replace(postCommit, "{{ GoBin }}", os.Args[0], -1))
		if err != nil {
			fmt.Println("Error while writing to post-commit hook file:", err)
			os.Exit(1)
		}
		f.Chmod(0755)
		fmt.Println("Wrote post-commit hook.")
	}

	// Write the post-checkout hook, unless user said no
	if !ic.ncheck {
		f, err := os.Create(repo.Path() + "/hooks/post-checkout")
		if err != nil {
			fmt.Println("Error while attempting to open post-commit hook for writing:", err)
			os.Exit(1)
		}
		_, err = f.WriteString(strings.Replace(postCheckout, "{{ GoBin }}", os.Args[0], -1))
		if err != nil {
			fmt.Println("Error while writing to post-checkout hook file:", err)
			os.Exit(1)
		}
		f.Chmod(0755)
		fmt.Println("Wrote post-checkout hook.")
	}

	// TODO history and refs
}
