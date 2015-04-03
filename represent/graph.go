package represent

import (
	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
)

// the main graph construct
type CoreGraph struct {
	list    map[int]vtTuple
	vserial int
}

type VtxI interface {
	// Merges another vertex into this vertex. Error is indicated if the
	// dynamic types do not match.
	Merge(VtxI) (VtxI, error)
	// Returns a string representing the object type. Used for namespacing keys, etc.
	// While this is (currently) implemented as a method, its result must be invariant.
	// TODO use string-const generator, other tricks to enforce invariance, compact space use
	Typ() string
	// Returns a persistent map with the vertex's properties.
	// TODO generate more type-restricted versions of the map?
	Props() ps.Map
}

type Vertex struct {
	props []Property
}

type vtTuple struct {
	v Vertex
	e []StandardEdge
}

func (v Vertex) Merge(iv Vertex) Vertex {
	return iv
}

type Property struct {
	MsgSrc int
	Key    string
	Value  interface{}
}

type rootedEdgeSpecTuple struct {
	vid int
	es  EdgeSpecs
}

type edgeSpecSet []rootedEdgeSpecTuple

func (ess edgeSpecSet) EdgeCount() (i int) {
	for _, tuple := range ess {
		i = i + len(tuple.es)
	}
	return
}

// the method to merge a message into the graph
func (g *CoreGraph) Merge(msg interpret.Message) {
	var ess edgeSpecSet

	// TODO see if some lookup + locality of reference perf benefit can be gained keeping all vertices loaded up locally

	// Process incoming elements from the message
	msg.Each(func(d interface{}) {
		// Split each input element into vertex and edge specs
		// TODO errs
		vtx, edges, _ := Split(d, msg.Id)
		// Ensure the vertex is present
		vid := g.ensureVertex(vtx)

		// Collect edge specs for later processing
		ess = append(ess, rootedEdgeSpecTuple{
			vid: vid,
			es:  edges,
		})
	})

	// All vertices processed. now, process edges in passes, ensuring that each
	// pass diminishes the number of remaining edges. If it doesn't, the remaining
	// edges need to be attached to null-vertices of the appropriate type.
	//
	// This is a little wasteful, but it's the simplest way to let any possible
	// dependencies between edges work themselves out. It has provably incorrect
	// cases, however, and will need to be replaced.
	var ec, lec int
	for ec = ess.EdgeCount(); ec != lec; ec = ess.EdgeCount() {
		lec = ec
		for _, tuple := range ess {
			for k, spec := range tuple.es {
				edge, success := Resolve(g, spec)
				if success {
					g.ensureEdge(edge)
					tuple.es = append(tuple.es[:k], tuple.es[k+1:]...)
				}
			}
		}
	}
}

// Ensures the vertex is present. Merges according to type-specific logic if
// it is present, otherwise adds the vertex.
//
// Either way, return value is the vid for the vertex.
func (g *CoreGraph) ensureVertex(vtx Vertex) (vid int) {
	vid = g.Find(vtx)

	if vid != 0 {
		vt := g.list[vid]
		g.list[vid] = vtTuple{e: vt.e, v: vt.v.Merge(vtx)}
	} else {
		g.vserial++
		g.list[g.vserial] = vtTuple{v: vtx}
		vid = g.vserial
	}

	return vid
}

// TODO add edge to the structure. blah blah handwave blah blah
func (g *CoreGraph) ensureEdge(e StandardEdge) {

}

// the func we eventually aim to fulfill, replacing Merge for integrating messages
//func (g CoreGraph) Cons(interpret.Message) CoreGraph, Delta, error {}

// Searches for an instance of the vertex within the graph. If found,
// returns the vertex id, otherwise returns 0.
//
// TODO this is really a querying method. needs to be replaced by that whole subsystem
func (g *CoreGraph) Find(vtx Vertex) int {
	// FIXME so very hilariously O(n)

	var chk interpret.Identifier
	for _, idf := range interpret.Identifiers {
		if idf.CanIdentify(vtx) {
			chk = idf
		}
	}

	// we hit this case iff there's an object type our identifiers can't work
	// with. which should, eventually, be structurally impossible by this point
	if chk == nil {
		return 0
	}

	for id, v := range g.list {
		if chk.Matches(v, vtx) {
			return id
		}
	}

	return 0
}

func (g *CoreGraph) Vertices(f func(Vertex, int) bool) {
	for id, vt := range g.list {
		if f(vt.v, id) {
			return
		}
	}
}

func (g *CoreGraph) Get(id int) (Vertex, error) {
	return Vertex{}, nil
}
