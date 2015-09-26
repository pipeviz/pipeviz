package interpret

import "github.com/tag1consulting/pipeviz/represent/types"

// roCoreGraph is a dummy local interface that covers pipeviz' main graph object.
type roCoreGraph interface {
	// Enumerates the outgoing edges from the ego vertex, limiting the result set
	// to those that pass the provided filter (if any).
	OutWith(egoId int, ef types.EFilter) (es []types.StdEdge)

	// Enumerates the incoming edges from the ego vertex, limiting the result set
	// to those that pass the provided filter (if any).
	InWith(egoId int, ef types.EFilter) (es []types.StdEdge)

	// Enumerates the successors (targets of outgoing edges) from the ego vertex,
	// limiting the result set to those that pass the provided edge and vertex
	// filters (if any).
	SuccessorsWith(egoId int, vef types.VEFilter) (vts []types.VertexTuple)

	// Enumerates the predecessors (sources of incoming edges) from the ego vertex,
	// limiting the result set to those that pass the provided edge and vertex
	// filters (if any).
	PredecessorsWith(egoId int, vef types.VEFilter) (vts []types.VertexTuple)

	// Enumerates the vertices that pass the provided vertex filter (if any).
	VerticesWith(vf types.VFilter) (vs []types.VertexTuple)

	// Gets the vertex tuple associated with a given id.
	Get(id int) (types.VertexTuple, error)

	// Returns the message id for the current version of the graph. The graph's
	// contents are guaranteed to represent the state resulting from a correct
	// in-order interpretation of the messages up to the id, inclusive.
	MsgId() uint64
}
