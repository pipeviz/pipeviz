package represent

import (
	"bytes"
	"fmt"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
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
		switch v.v.(type) {
		case vertexEnvironment:
			props = "\tshape=house,style=filled,fillcolor=orange,fontsize=20\n"
		case vertexLogicState:
			props = "\tshape=box3d,style=filled,fillcolor=purple,fontcolor=white,fontsize=18,\n"
		case vertexProcess:
			props = "\tshape=oval,style=filled,fillcolor=green,\n"
		case vertexDataset, vertexParentDataset:
			props = "\tshape=folder,style=filled,fillcolor=brown,fontcolor=white,\n"
		case vertexComm:
			props = "\tshape=doubleoctagon,style=filled,fillcolor=cyan,\n"
		case vertexCommit:
			props = "\tshape=box,style=filled,fillcolor=grey\n"
		case vertexVcsLabel:
			props = "\tshape=cds,margin=\"0.22,0.22\",\n"
		case vertexTestResult:
			props = "\tshape=note\n"
		}

		buf.WriteString(fmt.Sprintf(
			"\t\"v%d\" [%s%s=\"id: %d\nvtype: %s",
			v.id, props, lbltype, v.id, v.v.Typ()))

		v.v.Props().ForEach(func(k string, val ps.Any) {
			prop := val.(Property)
			var format string
			switch prop.Value.(type) {
			case []byte:
				format = "%x"
			case int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
				format = "%d"
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
		v.oe.ForEach(func(k string, val ps.Any) {
			edge := val.(StandardEdge)
			buf.WriteString(fmt.Sprintf(
				"\t\"v%d\" -> \"v%d\" [\n\tlabel=\"id: %d\netype: %s",
				edge.Source, edge.Target, edge.id, edge.EType))

			edge.Props.ForEach(func(k2 string, val2 ps.Any) {
				prop := val2.(Property)
				var format string
				switch prop.Value.(type) {
				case []byte:
					format = "%x"
				case int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
					format = "%d"
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
