package main

import (
	"log"
	"os"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
)

// shared stderr logger, for those that need it
var erro = log.New(os.Stderr, "", 0)

func main() {
	root := &cobra.Command{Use: "pvutil"}
	root.AddCommand(dotDumperCommand())
	root.AddCommand(fixrCommand())
	root.AddCommand(validateCommand())
	root.Execute()
}
