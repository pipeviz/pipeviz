package types

import (
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/interpret"
)

type SplitData struct {
	Vertex    StdVertex
	EdgeSpecs EdgeSpecs
}

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
	OutWith(egoId int, ef EFilter) (es []StdEdge)

	// Enumerates the incoming edges from the ego vertex, limiting the result set
	// to those that pass the provided filter (if any).
	InWith(egoId int, ef EFilter) (es []StdEdge)

	// Enumerates the successors (targets of outgoing edges) from the ego vertex,
	// limiting the result set to those that pass the provided edge and vertex
	// filters (if any).
	SuccessorsWith(egoId int, vef VEFilter) (vts []VertexTuple)

	// Enumerates the predecessors (sources of incoming edges) from the ego vertex,
	// limiting the result set to those that pass the provided edge and vertex
	// filters (if any).
	PredecessorsWith(egoId int, vef VEFilter) (vts []VertexTuple)

	// Enumerates the vertices that pass the provided vertex filter (if any).
	VerticesWith(vf VFilter) (vs []VertexTuple)

	// Gets the vertex tuple associated with a given id.
	Get(id int) (VertexTuple, error)

	// Returns the message id for the current version of the graph. The graph's
	// contents are guaranteed to represent the state resulting from a correct
	// in-order interpretation of the messages up to the id, inclusive.
	MsgId() uint64
}

type UnifyInstructionForm interface {
	Vertex() StdVertex
	EdgeSpecs() []EdgeSpec
}

type blah interface {
	//Unify(CoreGraph, SplitData) int
	//Resolve
}

// TODO for now, no structure to this. change to queryish form later
type EdgeSpec interface {
	// Given a graph and a vtTuple root, searches for an existing edge
	// that this EdgeSpec would supercede
	//FindExisting(*CoreGraph, vtTuple) (StandardEdge, bool)

	// Resolves the spec into a real edge, merging as appropriate with
	// any existing edge (as returned from FindExisting)
	//Resolve(*CoreGraph, vtTuple, int) (StandardEdge, bool)
}

type EdgeSpecs []EdgeSpec

type flatVTuple struct {
	Id       int        `json:"id"`
	V        flatVertex `json:"vertex"`
	InEdges  []flatEdge `json:"inEdges"`
	OutEdges []flatEdge `json:"outEdges"`
}

type flatVertex struct {
	VType VType               `json:"type"`
	Props map[string]Property `json:"properties"`
}

type flatEdge struct {
	Id     int                 `json:"id"`
	Source int                 `json:"source"`
	Target int                 `json:"target"`
	EType  EType               `json:"etype"`
	Props  map[string]Property `json:"properties"`
}

func vtoflat(v StdVertex) (flat flatVertex) {
	flat.VType = v.Typ()

	flat.Props = make(map[string]Property)
	v.Props().ForEach(func(k string, v ps.Any) {
		flat.Props[k] = v.(Property) // TODO err...ever not property?
	})
	return
}

func etoflat(e StdEdge) (flat flatEdge) {
	flat = flatEdge{
		Id:     e.ID,
		Source: e.Source,
		Target: e.Target,
		EType:  e.EType,
		Props:  make(map[string]Property),
	}

	e.Props.ForEach(func(k2 string, v2 ps.Any) {
		flat.Props[k2] = v2.(Property)
	})

	return
}

// flatten the persistent structures in the vtTuple down into conventional ones (typically for easy printing).
// TODO if this is gonna be exported, it's gotta be cleaned up
func (vt VertexTuple) Flat() (flat flatVTuple) {
	flat.Id = vt.ID
	flat.V = vtoflat(vt.Vertex)

	vt.InEdges.ForEach(func(k string, v ps.Any) {
		e := v.(StdEdge)
		fe := etoflat(e)
		flat.InEdges = append(flat.InEdges, fe)
	})

	vt.OutEdges.ForEach(func(k string, v ps.Any) {
		e := v.(StdEdge)
		fe := etoflat(e)
		flat.OutEdges = append(flat.OutEdges, fe)
	})

	return flat
}
