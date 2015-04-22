package represent

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
)

var i2a = strconv.Itoa

/*
CoreGraph is the interface provided by pipeviz' main graph object.

It is a persistent/immutable datastructure: only the Merge() method is able
to change the graph, but that method returns a pointer to a new graph rather
than updating in-place. (These copies typically share structure for efficiency)

All other methods are read-only, and generally provide composable parts that
work together to facilitate the creation of larger traversals/queries.
*/
type CoreGraph interface {
	// Merge a message into the graph, returning a pointer to the new graph
	// that contains the resulting updates.
	Merge(msg interpret.Message) CoreGraph

	// Enumerates the outgoing edges from the ego vertex, limiting the result set
	// to those that pass the provided filter (if any).
	OutWith(egoId int, ef EFilter) (es []StandardEdge)

	// Enumerates the incoming edges from the ego vertex, limiting the result set
	// to those that pass the provided filter (if any).
	InWith(egoId int, ef EFilter) (es []StandardEdge)

	// Enumerates the successors (targets of outgoing edges) from the ego vertex,
	// limiting the result set to those that pass the provided edge and vertex
	// filters (if any).
	SuccessorsWith(egoId int, vef VEFilter) (vts []vtTuple)

	// Enumerates the predecessors (sources of incoming edges) from the ego vertex,
	// limiting the result set to those that pass the provided edge and vertex
	// filters (if any).
	PredecessorsWith(egoId int, vef VEFilter) (vts []vtTuple)

	// Enumerates the vertices that pass the provided vertex filter (if any).
	VerticesWith(vf VFilter) (vs []vtTuple)

	// Gets the vertex tuple associated with a given id.
	Get(id int) (vtTuple, error)
}

// the main graph construct
type coreGraph struct {
	// TODO experiment with replacing with a hash array-mapped trie
	vtuples ps.Map
	vserial int
}

type (
	// A value indicating a vertex's type. For now, done as a string.
	VType string
	// A value indicating an edge's type. For now, done as a string.
	EType string
)

const (
	VTypeNone VType = ""
	ETypeNone EType = ""
)

// Used in queries to specify property k/v pairs.
type PropQ struct {
	K string
	V interface{}
}

type Vertex interface {
	// Merges another vertex into this vertex. Error is indicated if the
	// dynamic types do not match.
	Merge(Vertex) (Vertex, error)
	// Returns a string representing the object type. Used for namespacing keys, etc.
	// While this is (currently) implemented as a method, its result must be invariant.
	// TODO use string-const generator, other tricks to enforce invariance, compact space use
	Typ() VType
	// Returns a persistent map with the vertex's properties.
	// TODO generate more type-restricted versions of the map?
	Props() ps.Map
	// Reports the keys of the properties used as the distinguishing identifiers for this vertex.
	//Ids() []string
}

type vtTuple struct {
	id int
	v  Vertex
	ie ps.Map
	oe ps.Map
}

type Property struct {
	MsgSrc int
	Value  interface{}
}

type veProcessingInfo struct {
	vt vtTuple
	es EdgeSpecs
}

type edgeSpecSet []*veProcessingInfo

func (ess edgeSpecSet) EdgeCount() (i int) {
	for _, tuple := range ess {
		i = i + len(tuple.es)
	}
	return
}

// the method to merge a message into the graph
func (g *coreGraph) Merge(msg interpret.Message) CoreGraph {
	var ess edgeSpecSet

	// TODO see if some lookup + locality of reference perf benefit can be gained keeping all vertices loaded up locally

	// Process incoming elements from the message
	msg.Each(func(d interface{}) {
		// Split each input element into vertex and edge specs
		// TODO errs
		sds, _ := Split(d, msg.Id)

		// Ensure vertices are present
		var tuples []vtTuple
		for _, sd := range sds {
			tuples = append(tuples, g.ensureVertex(msg.Id, sd))
		}

		// Collect edge specs for later processing
		for k, tuple := range tuples {
			ess = append(ess, &veProcessingInfo{
				vt: tuple,
				es: sds[k].EdgeSpecs,
			})
		}
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
		for _, info := range ess {
			for k, spec := range info.es {
				edge, success := Resolve(g, msg.Id, info.vt, spec)
				if success {
					edge.Source = info.vt.id
					if edge.id == 0 {
						// new edge, allocate a new id for it
						g.vserial++
						edge.id = g.vserial
					}

					// FIXME make sure assignment makes it back into ess slice
					info.vt.oe = info.vt.oe.Set(i2a(edge.id), edge)
					g.vtuples = g.vtuples.Set(i2a(info.vt.id), info.vt)

					// TODO setting multiple times is silly and wasteful
					any, _ := g.vtuples.Lookup(i2a(edge.Target))
					tvt := any.(vtTuple)
					tvt.ie = tvt.ie.Set(i2a(edge.id), edge)

					info.es = append(info.es[:k], info.es[k+1:]...)
				}
			}
		}
	}

	// TODO attempt to de-orphan items here?
	return g
}

// Ensures the vertex is present. Merges according to type-specific logic if
// it is present, otherwise adds the vertex.
//
// Either way, return value is the vid for the vertex.
func (g *coreGraph) ensureVertex(msgid int, sd SplitData) (final vtTuple) {
	vid := Identify(g, sd)

	if vid == 0 {
		final = vtTuple{v: sd.Vertex, ie: ps.NewMap(), oe: ps.NewMap()}
		g.vserial += 1
		final.id = g.vserial
		g.vtuples = g.vtuples.Set(i2a(g.vserial), final)
	} else {
		ivt, _ := g.vtuples.Lookup(i2a(vid))
		vt := ivt.(vtTuple)

		// TODO err
		nu, _ := vt.v.Merge(sd.Vertex)
		final = vtTuple{id: vid, ie: vt.ie, oe: vt.oe, v: nu}
		g.vtuples = g.vtuples.Set(i2a(vid), final)
	}

	return
}

// Gets the vtTuple for a given vertex id.
func (g *coreGraph) Get(id int) (vtTuple, error) {
	if id > g.vserial {
		return vtTuple{}, errors.New(fmt.Sprintf("Graph has only %d elements, no vertex yet exists with id %d", g.vserial, id))
	}

	vtx, exists := g.vtuples.Lookup(i2a(id))
	if exists {
		return vtx.(vtTuple), nil
	} else {
		return vtTuple{}, errors.New(fmt.Sprintf("No vertex exists with id %d at the present revision of the graph", id))
	}
}
