package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/kardianos/osext"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/clients/githelp"
)

type instrumentCmd struct {
	ncom, ncheck, refs, history bool
	target                      string
}

func instrumentCommand() *cobra.Command {
	ic := &instrumentCmd{}
	cmd := &cobra.Command{
		Use:   "instrument <repository>",
		Short: "Instruments a git repository to talk with a pipeviz server.",
		Long:  "Instruments a git repository by setting up hook scripts to emit data about the state of the local repository to a pipeviz server. If no path to a repository is provided, it will search for one from the working directory.",
		Run:   ic.run,
	}

	cmd.Flags().BoolVar(&ic.ncom, "no-post-commit", false, "Do not set up a post-commit hook.")
	cmd.Flags().BoolVar(&ic.ncheck, "no-post-checkout", false, "Do not set up a post-checkout hook.")
	cmd.Flags().BoolVar(&ic.history, "send-history", false, "Send all known history from the repository to the server. Encompasses --refs.")
	cmd.Flags().BoolVar(&ic.refs, "send-refs", false, "Sends the position of all local refs to the server.")
	cmd.Flags().StringVarP(&ic.target, "target", "t", "", "Address of the target pipeviz daemon. Required if an address is not already set in git config.")

	return cmd
}

const (
	postCommit = `#!/bin/sh
{{ binpath }} hook-post-commit
`
	postCheckout = `#!/bin/sh
{{ binpath }} hook-post-checkout $*
`
)

func (ic *instrumentCmd) run(cmd *cobra.Command, args []string) {
	var err error
	repo := getRepoOrExit(args...)

	// If we can't get an ident, error out
	_, err = githelp.GetRepoIdent(repo)
	if err != nil {
		log.Fatalf("Failed to retrieve a stable identifier for this repository; hooks will not work correctly. Aborting.")
	}

	if ic.target == "" {
		ic.target, err = githelp.GetTargetAddr(repo)
		if err != nil {
			log.Fatalf("No pipeviz server target provided, and one is not already registered in git's config.")
		}
	} else {
		cfg, err := repo.Config()
		if err != nil {
			log.Fatalln("Error attempting to retrieve git config", err)
		}

		err = cfg.SetString("pipeviz.target", ic.target)
		if err != nil {
			log.Fatalln("Error while writing pipeviz target to config", err)
		}
		fmt.Println("Set target pipeviz server to", ic.target)
	}

	// Write the post-commit hook, unless user said no
	if !ic.ncom {
		f, err := os.Create(repo.Path() + "/hooks/post-commit")
		if err != nil {
			log.Fatalln("Error while attempting to open post-commit hook for writing:", err)
		}

		tmpl, err := template.New("post-commit").Funcs(template.FuncMap{
			"binpath": osext.Executable,
		}).Parse(postCommit)

		if err != nil {
			log.Fatalln("Error while parsing script template:", err)
		}

		err = tmpl.Execute(f, nil)
		if err != nil {
			log.Fatalln("Error while writing to post-commit hook file:", err)
		}
		f.Chmod(0755)
		fmt.Println("Wrote post-commit hook.")
	}

	// Write the post-checkout hook, unless user said no
	if !ic.ncheck {
		f, err := os.Create(repo.Path() + "/hooks/post-checkout")
		if err != nil {
			log.Fatalln("Error while attempting to open post-commit hook for writing:", err)
		}

		tmpl, err := template.New("post-checkout").Funcs(template.FuncMap{
			"binpath": osext.Executable,
		}).Parse(postCheckout)

		if err != nil {
			log.Fatalln("Error while parsing script template:", err)
		}

		err = tmpl.Execute(f, nil)
		if err != nil {
			log.Fatalln("Error while writing to post-checkout hook file:", err)
		}
		f.Chmod(0755)
		fmt.Println("Wrote post-checkout hook.")
	}

	if ic.refs || ic.history {
		syncHistory(repo, ic.history)
	}

	//repo.Free()
}
