package represent

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/interpret"
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
	msgid, vserial int
	vtuples        ps.Map
	orphans        edgeSpecSet // FIXME breaks immut
}

func NewGraph() CoreGraph {
	return &coreGraph{vtuples: ps.NewMap(), vserial: 0}
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
	vt    vtTuple
	es    EdgeSpecs
	msgid int
}

type edgeSpecSet []*veProcessingInfo

func (ess edgeSpecSet) EdgeCount() (i int) {
	for _, tuple := range ess {
		i = i + len(tuple.es)
	}
	return
}

// Clones the graph object, and the map pointers it contains.
func (g *coreGraph) clone() *coreGraph {
	var cp coreGraph
	cp = *g
	return &cp
}

// the method to merge a message into the graph
func (og *coreGraph) Merge(msg interpret.Message) CoreGraph {
	var ess edgeSpecSet

	g := og.clone()
	g.msgid = msg.Id

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
				vt:    tuple,
				es:    sds[k].EdgeSpecs,
				msgid: msg.Id,
			})
		}
	})

	// Reinclude the held-over set of orphans for edge (re-)resolutions
	var ess2 edgeSpecSet
	// TODO lots of things very wrong with this approach, but works for first pass
	for _, orphan := range g.orphans {
		// vertex ident failed; try again now that new vertices are present
		if orphan.vt.id == 0 {
			orphan.vt = g.ensureVertex(orphan.msgid, SplitData{orphan.vt.v, orphan.es})
		} else {
			// ensure we have latest version of vt
			vt, err := g.Get(orphan.vt.id)
			if err != nil {
				// but if that vid has gone away, forget about it completely
				continue
			}
			orphan.vt = vt
		}

		ess2 = append(ess2, orphan)
	}

	// Put orphan stuff first so that it's guaranteed to be overwritten on conflict
	ess = append(ess2, ess...)

	// All vertices processed. now, process edges in passes, ensuring that each
	// pass diminishes the number of remaining edges. If it doesn't, the remaining
	// edges need to be attached to null-vertices of the appropriate type.
	//
	// This is a little wasteful, but it's the simplest way to let any possible
	// dependencies between edges work themselves out. It has provably incorrect
	// cases, however, and will need to be replaced.
	var ec, lec, pass int
	for ec = ess.EdgeCount(); ec != 0 && ec != lec; ec = ess.EdgeCount() {
		pass += 1
		lec = ec
		for infokey, info := range ess {
			specs := info.es
			info.es = info.es[:0]
			for _, spec := range specs {
				edge, success := Resolve(g, msg.Id, info.vt, spec)
				if success {

					edge.Source = info.vt.id
					if edge.id == 0 {
						// new edge, allocate a new id for it
						g.vserial++
						edge.id = g.vserial
					}

					info.vt.oe = info.vt.oe.Set(i2a(edge.id), edge)
					g.vtuples = g.vtuples.Set(i2a(info.vt.id), info.vt)

					any, _ := g.vtuples.Lookup(i2a(edge.Target))
					tvt := any.(vtTuple)
					tvt.ie = tvt.ie.Set(i2a(edge.id), edge)
					g.vtuples = g.vtuples.Set(i2a(tvt.id), tvt)
				} else {
					// FIXME mem leaks if done this way...?
					info.es = append(info.es, spec)
				}
			}
			// set the processing info back into its original position in the slice
			ess[infokey] = info
		}
	}

	g.orphans = g.orphans[:0]
	for _, info := range ess {
		if len(info.es) == 0 {
			continue
		}

		g.orphans = append(g.orphans, info)
	}

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
		// TODO remove this - temporarily cheat here by promoting EnvLink resolution, since so much relies on it
		for _, spec := range sd.EdgeSpecs {
			switch spec.(type) {
			case interpret.EnvLink, SpecDatasetHierarchy:
				edge, success := Resolve(g, msgid, final, spec)
				if success { // could fail if corresponding env not yet declared
					g.vserial += 1
					edge.id = g.vserial
					final.oe = final.oe.Set(i2a(edge.id), edge)

					// set edge in reverse direction, too
					any, _ := g.vtuples.Lookup(i2a(edge.Target))
					tvt := any.(vtTuple)
					tvt.ie = tvt.ie.Set(i2a(edge.id), edge)
					g.vtuples = g.vtuples.Set(i2a(tvt.id), tvt)
					g.vtuples = g.vtuples.Set(i2a(final.id), final)
				}
			}
		}
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
