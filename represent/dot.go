package represent

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/represent/types"
)

// Generates a .dot-format representation of the given CoreGraph, suitable for
// rendering into output by graphviz (or other utilities).
func GenerateDot(g CoreGraph) []byte {
	buf := new(bytes.Buffer)

	// begin the graph
	buf.WriteString("digraph{\n")
	buf.WriteString("fontsize=16")

	// first, write all vertices
	for _, v := range g.VerticesWith(Qbv()) {
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
			prop := val.(types.Property)
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
	for _, v := range g.VerticesWith(Qbv()) {
		v.OutEdges.ForEach(func(k string, val ps.Any) {
			edge := val.(StandardEdge)
			buf.WriteString(fmt.Sprintf(
				"\t\"v%d\" -> \"v%d\" [\n\tlabel=\"id: %d\netype: %s",
				edge.Source, edge.Target, edge.ID, edge.EType))

			edge.Props.ForEach(func(k2 string, val2 ps.Any) {
				prop := val2.(types.Property)
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
