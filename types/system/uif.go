package system

type UnifyInstructionForm interface {
	Vertex() StdVertex
	Unify(CoreGraph, UnifyInstructionForm) int
	EdgeSpecs() []EdgeSpec
	ScopingSpecs() []EdgeSpec
}

// TODO for now, no structure to this. change to queryish form later
type EdgeSpec interface {
	// Resolves the spec into a real edge, merging as appropriate with
	// any existing edge (as returned from FindExisting)
	Resolve(CoreGraph, uint64, VertexTuple) (StdEdge, bool)
}

type EdgeSpecs []EdgeSpec

type SplitData struct {
	Vertex    StdVertex
	EdgeSpecs EdgeSpecs
}
