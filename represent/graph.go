package represent

import "github.com/sdboyer/pipeviz/interpret"

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
	Value  interface{}
}

// some kind of structure representing the delta introduced by a message
type Delta struct{}

// the method to merge a message into the graph
func (g *CoreGraph) Merge(msg interpret.Message) Delta {

	msg.Each(func(d interface{}) {
		vertex, edges, err := Split(d, msg.Id)
		vid := g.Find(vertex)

		if vid != 0 {

		}
	})

	return Delta{}
}

// the func we eventually aim to fulfill, replacing Merge for integrating messages
//func (g CoreGraph) Cons(interpret.Message) CoreGraph, Delta, error {}

// Searches for an instance of the vertex within the graph. If found,
// returns the vertex id, otherwise returns 0.
func (g *CoreGraph) Find(vertex Vertex) int {
	// FIXME so very hilariously O(n)

	var chk interpret.Identifier
	for _, idf := range interpret.Identifiers {
		if idf.CanIdentify(vertex) {
			chk = idf
		}
	}

	// we hit this case iff there's an object type our identifiers can't work
	// with. which should, eventually, be structurally impossible by this point
	if chk == nil {
		return 0
	}

	for id, v := range g.vlist {
		if chk.Matches(v, vertex) {
			return id
		}
	}

	return 0
}
