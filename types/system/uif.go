package system

// Unifier is a type capable of representing itself as one or more UnifyInstructionForms.
type Unifier interface {
	UnificationForm() []UnifyInstructionForm
}

// UnifyInstructionForm describes a set of methods that express all the data necessary
// to fully unify a discrete datum (and its relationships) within the system.
type UnifyInstructionForm interface {
	Vertex() ProtoVertex
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
