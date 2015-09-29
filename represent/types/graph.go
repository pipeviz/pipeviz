package types

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
	Merge(Message) CoreGraph

	// Enumerates the outgoing edges from the ego vertex, limiting the result set
	// to those that pass the provided filter (if any).
	OutWith(egoId int, ef EFilter) EdgeVector

	// Enumerates the incoming edges from the ego vertex, limiting the result set
	// to those that pass the provided filter (if any).
	InWith(egoId int, ef EFilter) EdgeVector

	// Enumerates the successors (targets of outgoing edges) from the ego vertex,
	// limiting the result set to those that pass the provided edge and vertex
	// filters (if any).
	SuccessorsWith(egoId int, vef VEFilter) VertexTupleVector

	// Enumerates the predecessors (sources of incoming edges) from the ego vertex,
	// limiting the result set to those that pass the provided edge and vertex
	// filters (if any).
	PredecessorsWith(egoId int, vef VEFilter) VertexTupleVector

	// Enumerates the vertices that pass the provided vertex filter (if any).
	VerticesWith(vf VFilter) VertexTupleVector

	// Gets the vertex tuple associated with a given id.
	Get(id int) (VertexTuple, error)

	// Returns the message id for the current version of the graph. The graph's
	// contents are guaranteed to represent the state resulting from a correct
	// in-order interpretation of the messages up to the id, inclusive.
	MsgId() uint64
}
