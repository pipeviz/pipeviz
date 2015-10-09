package main

import "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"

func main() {
	root := &cobra.Command{Use: "pvgithook"}
	root.Execute()
}
