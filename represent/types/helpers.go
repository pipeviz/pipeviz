package types

type SplitData struct {
	Vertex    Vertex
	EdgeSpecs EdgeSpecs
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
