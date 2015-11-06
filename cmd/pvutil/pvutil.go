package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
)

// shared stderr logger, for those that need it
var erro = log.New(os.Stderr, "", 0)

var (
	version = "dev"
	vflag   bool
)

func main() {
	root := &cobra.Command{
		Use: "pvutil",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if vflag {
				fmt.Println("pvutil", cmd.Name(), "version", version)
				os.Exit(0)
			}
		},
	}

	root.AddCommand(dotDumperCommand())
	root.AddCommand(fixrCommand())
	root.AddCommand(validateCommand())

	root.PersistentFlags().BoolVarP(&vflag, "version", "v", false, "Print version")
	root.Execute()
}
