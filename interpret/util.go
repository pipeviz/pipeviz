package interpret

import (
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/maputil"
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

func findMatchingEnvId(g types.CoreGraph, edge types.StdEdge, vtv types.VertexTupleVector) int {
	for _, candidate := range vtv {
		for _, edge2 := range g.OutWith(candidate.ID, helpers.Qbe(types.EType("envlink"))) {
			if edge2.Target == edge.Target {
				return candidate.ID
			}
		}
	}

	return 0
}

func assignAddress(mid uint64, a Address, m ps.Map, excl bool) ps.Map {
	if a.Hostname != "" {
		if excl {
			m = m.Delete("ipv4")
			m = m.Delete("ipv6")
		}
		m = m.Set("hostname", types.Property{MsgSrc: mid, Value: a.Hostname})
	}
	if a.Ipv4 != "" {
		if excl {
			m = m.Delete("hostname")
			m = m.Delete("ipv6")
		}
		m = m.Set("ipv4", types.Property{MsgSrc: mid, Value: a.Ipv4})
	}
	if a.Ipv6 != "" {
		if excl {
			m = m.Delete("hostname")
			m = m.Delete("ipv4")
		}
		m = m.Set("ipv6", types.Property{MsgSrc: mid, Value: a.Ipv6})
	}

	return m
}

func findEnvironment(g types.CoreGraph, props ps.Map) (envid int, success bool) {
	rv := g.VerticesWith(helpers.Qbv(types.VType("environment")))
	for _, vt := range rv {
		if maputil.AnyMatch(props, vt.Vertex.Props(), "hostname", "ipv4", "ipv6", "nick") {
			return vt.ID, true
		}
	}

	return
}

func findDataset(g types.CoreGraph, envid int, name []string) (id int, success bool) {
	// first time through use the parent type
	vtype := types.VType("parent-dataset")
	id = envid

	var n string
	for len(name) > 0 {
		n, name = name[0], name[1:]
		rv := g.PredecessorsWith(id, helpers.Qbv(vtype, "name", n))
		vtype = "dataset"

		if len(rv) != 1 {
			return 0, false
		}

		id = rv[0].ID
	}

	return id, true
}
