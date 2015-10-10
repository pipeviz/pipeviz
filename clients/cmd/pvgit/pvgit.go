package main

import "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"

func main() {
	root := &cobra.Command{Use: "pvgit"}
	root.AddCommand(postCommitHookCommand())
	root.AddCommand(syncCommand())
	root.AddCommand(instrumentCommand())

	//var target string
	//root.PersistentFlags().StringVarP(&target, "target", "t", "http://localhost:2309", "Address of the target pipeviz daemon.")

	root.Execute()
}
