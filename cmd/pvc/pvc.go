package main

import "github.com/spf13/cobra"

func main() {
	root := &cobra.Command{Use: "pvc"}
	root.AddCommand(envCommand())

	var target string
	root.Flags().StringVarP(&target, "target", "t", "http://localhost:2309", "Address of the target pipeviz daemon. Defaults to http://localhost:2309")

	root.Execute()
}
