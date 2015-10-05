package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/spf13/cobra"
	gjs "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
	"github.com/tag1consulting/pipeviz/ingest"
	"github.com/tag1consulting/pipeviz/represent"
	"github.com/tag1consulting/pipeviz/represent/q"
	"github.com/tag1consulting/pipeviz/schema"
	"github.com/tag1consulting/pipeviz/types/system"
)

func dotDumperCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dotdump [-o|--output <path>] <dir>...",
		Short: "Dumps a graphviz-generated png of some/all of the JSON fixtures.",
		Long: `This tool generates a dot representation of a pipeviz core graph,
		having constructed that graph by merging in all JSON fixtures found in all the specified directories, in lexicographic order.`,
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
	raw, err := schema.Master()
	if err != nil {
		panic(fmt.Sprint("Failed to open master schema file, test must abort. message:", err.Error()))
	}

	schemaMaster, err := gjs.NewSchema(gjs.NewStringLoader(string(raw)))
	if err != nil {
		panic("bad schema...?")
	}

	if len(args) < 1 {
		log.Fatalf("Must provide at least one directory argument to dotdumper.")
	}

	var k uint64 = 0
	for _, dir := range args {
		fl, err := ioutil.ReadDir(dir)
		if err != nil {
			erro.Printf("Failed to read directory '%v' with error %v\n", dir, err)
		}

		for _, f := range fl {
			if match, _ := regexp.MatchString("\\.json$", f.Name()); match && !f.IsDir() {
				src, err := ioutil.ReadFile(dir + "/" + f.Name())
				if err != nil {
					erro.Printf("Failed to read fixture file %v/%v\n", dir, f.Name())
					continue
				}

				result, err := schemaMaster.Validate(gjs.NewStringLoader(string(src)))
				if err != nil {
					erro.Printf("Validation process terminated with errors for %v/%v. Error: \n%v\n", dir, f.Name(), err.Error())
					continue
				}

				if !result.Valid() {
					for _, desc := range result.Errors() {
						erro.Printf("\t%s\n", desc)
					}
				} else {
					k++
					m := ingest.Message{}
					json.Unmarshal(src, &m)

					g = g.Merge(k, m.UnificationForm())
					fmt.Printf("Merged message %v/%v into graph\n", dir, f.Name())
				}
			}
		}
	}

	pathflag := cmd.Flags().Lookup("output")
	if pathflag.Changed {
		ioutil.WriteFile(pathflag.Value.String(), GenerateDot(g), 0644)
	} else {
		fmt.Println(string(GenerateDot(g)))
	}
}

// Generates a .dot-format representation of the given CoreGraph, suitable for
// rendering into output by graphviz (or other utilities).
func GenerateDot(g system.CoreGraph) []byte {
	buf := new(bytes.Buffer)

	// begin the graph
	buf.WriteString("digraph{\n")
	buf.WriteString("fontsize=16")

	// first, write all vertices
	for _, v := range g.VerticesWith(q.Qbv()) {
		lbltype := "label"
		var props string
		switch v.Vertex.Typ() {
		case "environment":
			props = "\tshape=house,style=filled,fillcolor=orange,fontsize=20\n"
		case "logic-state":
			props = "\tshape=box3d,style=filled,fillcolor=purple,fontcolor=white,fontsize=18,\n"
		case "process":
			props = "\tshape=oval,style=filled,fillcolor=green,\n"
		case "dataset", "parent-dataset":
			props = "\tshape=folder,style=filled,fillcolor=brown,fontcolor=white,\n"
		case "comm":
			props = "\tshape=doubleoctagon,style=filled,fillcolor=cyan,\n"
		case "git-commit":
			props = "\tshape=box,style=filled,fillcolor=grey\n"
		case "git-tag", "git-branch":
			props = "\tshape=cds,margin=\"0.22,0.22\",\n"
		case "test-result":
			props = "\tshape=note\n"
		}

		buf.WriteString(fmt.Sprintf(
			"\t\"v%d\" [%s%s=\"id: %d\nvtype: %s",
			v.ID, props, lbltype, v.ID, v.Vertex.Typ()))

		v.Vertex.Props().ForEach(func(k string, val ps.Any) {
			prop := val.(system.Property)
			var format string
			switch pv := prop.Value.(type) {
			case []byte, [20]byte:
				format = "%x"
			case int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
				format = "%d"
			case string:
				prop.Value = strings.Trim(strconv.QuoteToASCII(pv), `"`)
				format = "%s"
			default:
				format = "%s"
			}
			buf.WriteString(fmt.Sprintf(
				"\n%s: "+format+" (%d)",
				k, prop.Value, prop.MsgSrc))
		})

		buf.WriteString("\"\n")

		buf.WriteString("];\n")
	}

	// pass through a second time to write all edges
	for _, v := range g.VerticesWith(q.Qbv()) {
		v.OutEdges.ForEach(func(k string, val ps.Any) {
			edge := val.(system.StdEdge)
			buf.WriteString(fmt.Sprintf(
				"\t\"v%d\" -> \"v%d\" [\n\tlabel=\"id: %d\netype: %s",
				edge.Source, edge.Target, edge.ID, edge.EType))

			edge.Props.ForEach(func(k2 string, val2 ps.Any) {
				prop := val2.(system.Property)
				var format string
				switch pv := prop.Value.(type) {
				case []byte:
					format = "%x"
				case int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
					format = "%d"
				case string:
					prop.Value = strings.Trim(strconv.QuoteToASCII(pv), `"`)
					format = "%s"
				default:
					format = "%s"
				}
				buf.WriteString(fmt.Sprintf(
					"\n%s: "+format+" (%d)",
					k2, prop.Value, prop.MsgSrc))
			})

			buf.WriteString("\"\n")

			switch edge.EType {
			case "envlink":
				buf.WriteString("\tstyle=dashed\n")
			}

			buf.WriteString("];\n")
		})
	}

	// close out the graph
	buf.WriteString("}\n")

	return buf.Bytes()
}
