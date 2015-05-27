package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/tag1consulting/pipeviz/fixtures"
	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/represent"
)

func dotDumperCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dotdump [-o|--output <path>] ([-a|--all] | <fixture number>...)",
		Short: "Dumps a graphviz-generated png of some/all of the JSON fixtures.",
		Long: `This tool generates a dot representation of a pipeviz core graph,
		having constructed that graph by merging in a series of the fixture messages
		in the order designated by the parameters.`,
		Run: runDotDumper,
	}

	var all bool
	var path string
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Includes all input message fixtures, in order. Arguments are ignored.")
	cmd.Flags().StringVarP(&path, "output", "o", "", "Specifies a file to write dot output to. Otherwise, prints to stdout.")

	return cmd
}

func runDotDumper(cmd *cobra.Command, args []string) {
	g := represent.NewGraph()

	names := fixtures.AssetNames()
	if cmd.Flags().Lookup("all").Changed {
		for _, name := range names {
			d, _ := fixtures.Asset(name)

			parts := strings.Split(name, ".")
			i, err := strconv.Atoi(parts[0])
			if err != nil {
				// will only error if we stop using numbered names
				panic(fmt.Sprintf("Non-numeric fixture filename encountered: %q", name))
			}
			m := interpret.Message{Id: i}

			json.Unmarshal(d, &m)
			g = g.Merge(m)
		}
	} else {
		for k, arg := range args {
			var d []byte
			for _, name := range names {
				parts := strings.Split(name, ".")
				if parts[0] == arg {
					d, _ = fixtures.Asset(name)
					break
				}
			}

			if d == nil {
				panic(fmt.Sprintf("No fixture message corresponding to %q", arg))
			}

			// This assigns message ids based on the order, not the filename. While
			// more correct, it may be a bit confusing for unaware users
			m := interpret.Message{Id: k + 1}
			json.Unmarshal(d, &m)

			g = g.Merge(m)
		}
	}

	pathflag := cmd.Flags().Lookup("output")
	if pathflag.Changed {
		ioutil.WriteFile(pathflag.Value.String(), represent.GenerateDot(g), 0644)
	} else {
		fmt.Println(string(represent.GenerateDot(g)))
	}
}
