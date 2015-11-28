package semantic

import (
	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/maputil"
	"github.com/pipeviz/pipeviz/represent/q"
	"github.com/pipeviz/pipeviz/types/system"
)

// pp is a convenience function to create a types.PropPair. The compiler
// always inlines it, so there's no cost.
func pp(k string, v interface{}) system.PropPair {
	return system.PropPair{K: k, V: v}
}

// uif is a standard struct that expresses a types.UnifyInstructionForm
type uif struct {
	v  system.ProtoVertex
	e  []system.EdgeSpec
	se []system.EdgeSpec
}

func (u uif) Vertex() system.ProtoVertex {
	return u.v
}

func (u uif) EdgeSpecs() []system.EdgeSpec {
	return u.e
}

func (u uif) ScopingSpecs() []system.EdgeSpec {
	return u.se
}

// pv is a shared/common type to implement system.ProtoVertex
type pv struct {
	typ   system.VType
	props system.RawProps
}

func (p pv) Type() system.VType {
	return p.typ
}

func (p pv) Properties() system.RawProps {
	return p.props
}

// Searches the given vertex's out-edges to find its environment's vertex id.
//
// Also conveniently initializes a StandardEdge to the standard zero-state for an envlink.
func findEnv(g system.CoreGraph, vt system.VertexTuple) (vid uint64, edge system.StdEdge, success bool) {
	edge = system.StdEdge{
		Source: vt.ID,
		Props:  ps.NewMap(),
		EType:  "envlink",
	}

	if vt.ID != 0 {
		re := g.OutWith(vt.ID, q.Qbe(system.EType("envlink")))
		if len(re) == 1 {
			vid, edge, success = re[0].Target, re[0], true
		}
	}

	edge.Incomplete = !success
	return
}

// TODO it would be better to not have to have this here, at all.
func emptyVT(v system.ProtoVertex) system.VertexTuple {
	var props []system.PropPair
	for k, v := range v.Properties() {
		props = append(props, pp(k, v))
	}

	return system.VertexTuple{
		Vertex:   system.NewVertex(v.Type(), 0, props...),
		InEdges:  ps.NewMap(),
		OutEdges: ps.NewMap(),
	}
}

func findMatchingEnvId(g system.CoreGraph, edge system.StdEdge, vtv system.VertexTupleVector) uint64 {
	for _, candidate := range vtv {
		for _, edge2 := range g.OutWith(candidate.ID, q.Qbe(system.EType("envlink"))) {
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
		m = m.Set("hostname", system.Property{MsgSrc: mid, Value: a.Hostname})
	}
	if a.Ipv4 != "" {
		if excl {
			m = m.Delete("hostname")
			m = m.Delete("ipv6")
		}
		m = m.Set("ipv4", system.Property{MsgSrc: mid, Value: a.Ipv4})
	}
	if a.Ipv6 != "" {
		if excl {
			m = m.Delete("hostname")
			m = m.Delete("ipv4")
		}
		m = m.Set("ipv6", system.Property{MsgSrc: mid, Value: a.Ipv6})
	}

	return m
}

func findEnvironment(g system.CoreGraph, props ps.Map) (envid uint64, success bool) {
	rv := g.VerticesWith(q.Qbv(system.VType("environment")))
	for _, vt := range rv {
		if maputil.AnyMatch(props, vt.Vertex.Props(), "hostname", "ipv4", "ipv6", "nick") {
			return vt.ID, true
		}
	}

	return
}

func findDataset(g system.CoreGraph, envid uint64, name []string) (id uint64, success bool) {
	// first time through use the parent type
	vtype := system.VType("parent-dataset")
	id = envid

	var n string
	for len(name) > 0 {
		n, name = name[0], name[1:]
		rv := g.PredecessorsWith(id, q.Qbv(vtype, "name", n))
		vtype = "dataset"

		if len(rv) != 1 {
			return 0, false
		}

		id = rv[0].ID
	}

	return id, true
}
