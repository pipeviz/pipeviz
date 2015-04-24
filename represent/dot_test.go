package represent

import (
	"io/ioutil"
	"testing"

	"github.com/mndrix/ps"
)

// not actually a test func, just an easy place to dump drawing the dot file for now
func TestDrawDot(t *testing.T) {
	var g CoreGraph = &coreGraph{vtuples: ps.NewMap(), vserial: 0}

	// build up the graph with all our fixture data
	g = g.Merge(msgs[0])
	// fourth message has the other envs
	g = g.Merge(msgs[3])
	// third msg is commits, all standalone
	g = g.Merge(msgs[2])
	// sixth is datasets
	g = g.Merge(msgs[5])
	// logic state msgs
	g = g.Merge(msgs[1])
	g = g.Merge(msgs[6])
	g = g.Merge(msgs[4])
	// processes msg
	g = g.Merge(msgs[7])

	ioutil.WriteFile("graph.dot", generateDot(g), 0644)
}
