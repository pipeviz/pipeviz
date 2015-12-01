package system

import "github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"

type StdEdge struct {
	ID, Source, Target uint64
	Incomplete         bool
	Props              ps.Map
	EType              EType
}

// EdgeVector contains an ordered list of StdEdges. It is the return value of
// some types of traversal queries.
type EdgeVector []StdEdge

// IDs returns a slice containing the numerical identifiers for all the edges in the vector.
// FIXME pointer receiver for this? not sure if this being slice-based makes it weird
func (ev EdgeVector) IDs() []uint64 {
	if len(ev) == 0 {
		return nil
	}

	ret := make([]uint64, len(ev))
	for k, e := range ev {
		ret[k] = e.ID
	}

	return ret
}

// IDs returns a slice containing the numerical identifiers for the target vertex of all edges in the vector.
func (ev EdgeVector) TargetIDs() []uint64 {
	if len(ev) == 0 {
		return nil
	}

	ret := make([]uint64, len(ev))
	for k, e := range ev {
		ret[k] = e.Target
	}

	return ret
}

// IDs returns a slice containing the numerical identifiers for the source vertex of all edges in the vector.
func (ev EdgeVector) SourceIDs() []uint64 {
	if len(ev) == 0 {
		return nil
	}

	ret := make([]uint64, len(ev))
	for k, e := range ev {
		ret[k] = e.Source
	}

	return ret
}

// TargetSuccessorsWith performs a successor walk step from each target vertex
// in the edge vector.
func (ev EdgeVector) TargetSuccessorsWith(g CoreGraph, vef VEFilter) VertexTupleVector {
	if len(ev) == 0 {
		return nil
	}

	ret := make(VertexTupleVector, 0)

	for _, e := range ev {
		ret = append(ret, g.SuccessorsWith(e.Target, vef)...)
	}

	return ret
}

// SourcePredecessorsWith performs a predecessor walk step from each source vertex
// in the edge vector.
func (ev EdgeVector) SourcePredecessorsWith(g CoreGraph, vef VEFilter) VertexTupleVector {
	if len(ev) == 0 {
		return nil
	}

	ret := make(VertexTupleVector, 0)

	for _, e := range ev {
		ret = append(ret, g.PredecessorsWith(e.Source, vef)...)
	}

	return ret
}

// TargetOutWith finds edges outbound from each target vertex in the edge vector.
func (ev EdgeVector) TargetOutWith(g CoreGraph, ef EFilter) EdgeVector {
	if len(ev) == 0 {
		return nil
	}

	ret := make(EdgeVector, 0)

	for _, e := range ev {
		ret = append(ret, g.OutWith(e.Target, ef)...)
	}

	return ret
}

// SourceInWith finds edges outbound from each target vertex in the edge vector.
func (ev EdgeVector) SourceInWith(g CoreGraph, ef EFilter) EdgeVector {
	if len(ev) == 0 {
		return nil
	}

	ret := make(EdgeVector, 0)

	for _, e := range ev {
		ret = append(ret, g.InWith(e.Source, ef)...)
	}

	return ret
}
