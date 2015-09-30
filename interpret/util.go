package interpret

import (
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/represent/helpers"
	"github.com/tag1consulting/pipeviz/represent/types"
)

// pp is a convenience function to create a types.PropPair. The compiler
// always inlines it, so there's no cost.
func pp(k string, v interface{}) types.PropPair {
	return types.PropPair{K: k, V: v}
}

// uif is a standard struct that expresses a types.UnifyInstructionForm
type uif struct {
	v  types.StdVertex
	u  func(types.CoreGraph, types.UnifyInstructionForm) int
	e  []types.EdgeSpec
	se []types.EdgeSpec
}

func (u uif) Vertex() types.StdVertex {
	return u.v
}

func (u uif) Unify(g types.CoreGraph, u2 types.UnifyInstructionForm) int {
	// TODO u2 should be redundant, should always be same as u
	return u.u(g, u2)
}

func (u uif) EdgeSpecs() []types.EdgeSpec {
	return u.e
}

func (u uif) ScopingSpecs() []types.EdgeSpec {
	return u.se
}

// Searches the given vertex's out-edges to find its environment's vertex id.
//
// Also conveniently initializes a StandardEdge to the standard zero-state for an envlink.
func findEnv(g types.CoreGraph, vt types.VertexTuple) (vid int, edge types.StdEdge, success bool) {
	edge = types.StdEdge{
		Source: vt.ID,
		Props:  ps.NewMap(),
		EType:  "envlink",
	}

	if vt.ID != 0 {
		re := g.OutWith(vt.ID, helpers.Qbe(types.EType("envlink")))
		if len(re) == 1 {
			vid, edge, success = re[0].Target, re[0], true
		}
	}

	return
}

func emptyVT(v types.StdVertex) types.VertexTuple {
	return types.VertexTuple{
		Vertex:   v,
		InEdges:  ps.NewMap(),
		OutEdges: ps.NewMap(),
	}
}
