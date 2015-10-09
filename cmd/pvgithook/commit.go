package main

import "github.com/spf13/cobra"

func postCommitHookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post-commit",
		Short: "Processes a git post-commit event.",
		Run:   runPostCommit,
	}

	var all bool
	var path string
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Includes all input message fixtures, in order. Arguments are ignored.")
	cmd.Flags().StringVarP(&path, "output", "o", "", "Specifies a file to write dot output to. Otherwise, prints to stdout.")

	return cmd
}

func runPostCommit(cmd *cobra.Command, args []string) {
}
