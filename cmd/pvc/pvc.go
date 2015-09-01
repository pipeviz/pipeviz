package main

import "github.com/spf13/cobra"

func main() {
	root := &cobra.Command{Use: "pvutil"}
	root.Execute()
}
