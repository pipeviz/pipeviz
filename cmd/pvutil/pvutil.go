package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pipeviz/pipeviz/version"
	"github.com/spf13/cobra"
)

// shared stderr logger, for those that need it
var erro = log.New(os.Stderr, "", 0)

func main() {
	root := &cobra.Command{
		Use: "pvutil",
	}

	root.AddCommand(dotDumperCommand())
	root.AddCommand(fixrCommand())
	root.AddCommand(validateCommand())
	root.AddCommand(tfmCommand())

	var vflag bool
	root.PersistentFlags().BoolVarP(&vflag, "version", "v", false, "Print version")
	root.ParseFlags(os.Args)
	if vflag {
		fmt.Println("pvutil", root.Name(), "version", version.Version())
		os.Exit(0)
	}

	root.Execute()
}
