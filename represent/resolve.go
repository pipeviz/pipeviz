package represent

import (
	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
)

// Attempts to resolve an EdgeSpec into a real edge. This process has two steps:
//
// 1. Finding the target node.
// 2. Seeing if this edge already exists between source and target.
//
// It is the responsibility of the edge spec's type handler to determine what "if an edge
// already exists" means, as well as whether to overwrite/merge or duplicate the edge in such a case.
func Resolve(g *CoreGraph, mid int, src vtTuple, d EdgeSpec) (StandardEdge, bool) {
	switch es := d.(type) {
	case interpret.EnvLink:
		return resolveEnvLink(g, mid, src, es)
	case interpret.DataLink:
		return resolveDataLink(g, src, es)
	case SpecCommit:
		return resolveSpecCommit(g, src, es)
	case SpecLocalLogic:
		return resolveSpecLocalLogic(g, src, es)
	}

	return StandardEdge{}, false
}

func resolveEnvLink(g *CoreGraph, mid int, src vtTuple, es interpret.EnvLink) (e StandardEdge, success bool) {
	e = StandardEdge{
		Source: src.id,
		Spec:   ps.NewMap(),
		Props:  ps.NewMap(),
	}

	// Whether we find a match or not, have to record the EnvLink
	if es.Address.Hostname != "" {
		e.Spec = e.Spec.Set("hostname", Property{MsgSrc: mid, Value: es.Address.Hostname})
	}
	if es.Address.Ipv4 != "" {
		e.Spec = e.Spec.Set("ipv4", Property{MsgSrc: mid, Value: es.Address.Ipv4})
	}
	if es.Address.Ipv6 != "" {
		e.Spec = e.Spec.Set("ipv6", Property{MsgSrc: mid, Value: es.Address.Ipv6})
	}
	if es.Nick != "" {
		e.Spec = e.Spec.Set("nick", Property{MsgSrc: mid, Value: es.Nick})
	}

	g.Vertices(func(vtx Vertex, id int) bool {
		if v, ok := vtx.(environmentVertex); !ok {
			return false
		}

		// TODO for now we're just gonna return out the first matching edge
		props := vtx.Props()

		var val interface{}
		var exists bool

		if val, exists := props.Lookup("hostname"); exists && val == es.Address.Hostname {
			success = true
			e.Target = id
			return true
		}
		if val, exists := props.Lookup("ipv4"); exists && val == es.Address.Ipv4 {
			success = true
			e.Target = id
			return true
		}
		if val, exists := props.Lookup("ipv6"); exists && val == es.Address.Ipv6 {
			success = true
			e.Target = id
			return true
		}
		if val, exists := props.Lookup("nick"); exists && val == es.Nick {
			success = true
			e.Target = id
			return true
		}

		return false
	})

	return e, success
}

func resolveDataLink(g *CoreGraph, src vtTuple, e interpret.DataLink) (StandardEdge, bool) {
	g.Vertices(func(vtx Vertex, id int) bool {})
}

func resolveSpecCommit(g *CoreGraph, src vtTuple, e SpecCommit) (StandardEdge, bool) {
	g.Vertices(func(vtx Vertex, id int) bool {})
}

func resolveSpecLocalLogic(g *CoreGraph, src vtTuple, e SpecLocalLogic) (StandardEdge, bool) {
	g.Vertices(func(vtx Vertex, id int) bool {})
}
