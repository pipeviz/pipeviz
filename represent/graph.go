package represent

import (
	"github.com/sdboyer/gogl"
	"github.com/sdboyer/pipeviz/interpret"
)

// the main graph construct
type CoreGraph struct {
	vlist   map[int]Vertex
	vserial int
}

type Vertex struct {
	props []Property
}

type Property struct {
	MsgSrc int
	Key    string
	Value  string
}

// some kind of structure representing the delta introduced by a message
type Delta struct{}

// the method to merge a message into the graph
func (g *CoreGraph) Merge(interpret.Message) Delta {

}

// the func we eventually aim to fulfill, replacing Merge for integrating messages
//func (g CoreGraph) Cons(interpret.Message) CoreGraph, Delta, error {}

// Searches for an instance of the vertex within the graph. If found,
// returns the vertex id and true; otherwise returns 0 and false.
func (g *CoreGraph) Find(vertex gogl.Vertex) (int, bool) {
	// FIXME so very hilariously O(n)

	var chk Identifier
	for _, idf := range Identifiers {
		if idf.CanIdentify(vertex) {
			chk = idf
		}
	}

	// we hit this case iff there's an object type our identifiers can't work
	// with. which should, eventually, be structurally impossible by this point
	if chk == nil {
		return 0, false
	}

	for id, vc := range g.list {
		if chk.Matches(vc.Vertex, vertex) {
			return id, true
		}
	}

	return 0, false
}
