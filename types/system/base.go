package system

import "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"

type (
	// A value indicating a vertex's type. For now, done as a string.
	VType string
	// A value indicating an edge's type. For now, done as a string.
	EType string
)

const (
	// Constant representing the zero value for VTypes. Reserved; no actual vertex can be this.
	VTypeNone VType = ""
	// Constant representing the zero value for ETypes. Reserved; no actual edge can be this.
	ETypeNone EType = ""
)

type ProtoVertex interface {
	Type() VType
	Properties() RawProps
	// The CoreGraph, and the defined scoping edge specs (if any)
	//Unify(g CoreGraph, se []EdgeSpec) uint64
}

// VertexTuple is the base storage object used by the graph engine.
type VertexTuple struct {
	ID       uint64
	Vertex   StdVertex
	InEdges  ps.Map
	OutEdges ps.Map
}

// VertexTupleVector contains an ordered list of VertexTuples. It is the return
// value of some types of traversal queries.
//
// VertexTupleVectors are ephemeral. The behavior of its traversal methods are
// undefined if provided a graph different from the one that created it.
type VertexTupleVector []VertexTuple

// IDs returns a slice containing the numerical identifiers for all the tuples in the vector.
// FIXME pointer receiver for this? not sure if this being slice-based makes it weird
func (vtv VertexTupleVector) IDs() []uint64 {
	if len(vtv) == 0 {
		return nil
	}

	ret := make([]uint64, len(vtv))
	for k, vt := range vtv {
		ret[k] = vt.ID
	}

	return ret
}

// SuccessorsWith performs a successor walk step from multiple input vertices.
//
// The returned VertexTupleVector contains only the destination vertices. It is
// not possible to determine which intermediate vertex (in this VertexTupleVector)
// was traversed to arrive at a result vertex.
func (vtv VertexTupleVector) SuccessorsWith(g CoreGraph, vef VEFilter) VertexTupleVector {
	if len(vtv) == 0 {
		return nil
	}

	ret := make(VertexTupleVector, 0)

	for _, vt := range vtv {
		ret = append(ret, g.SuccessorsWith(vt.ID, vef)...)
	}

	return ret
}

// PredecessorsWith performs a predecessor walk step from multiple input vertices.
//
// The returned VertexTupleVector contains only the destination vertices. It is
// not possible to determine which intermediate vertex (in this VertexTupleVector)
// was traversed to arrive at a result vertex.
func (vtv VertexTupleVector) PredecessorsWith(g CoreGraph, vef VEFilter) VertexTupleVector {
	if len(vtv) == 0 {
		return nil
	}

	ret := make(VertexTupleVector, 0)

	for _, vt := range vtv {
		ret = append(ret, g.PredecessorsWith(vt.ID, vef)...)
	}

	return ret
}

// OutWith finds edges outbound from multiple input vertices.
func (vtv VertexTupleVector) OutWith(g CoreGraph, ef EFilter) EdgeVector {
	if len(vtv) == 0 {
		return nil
	}

	ret := make(EdgeVector, 0)

	for _, vt := range vtv {
		ret = append(ret, g.OutWith(vt.ID, ef)...)
	}

	return ret
}

// InWith finds edges inbound to multiple input vertices.
func (vtv VertexTupleVector) InWith(g CoreGraph, ef EFilter) EdgeVector {
	if len(vtv) == 0 {
		return nil
	}

	ret := make(EdgeVector, 0)

	for _, vt := range vtv {
		ret = append(ret, g.InWith(vt.ID, ef)...)
	}

	return ret
}
