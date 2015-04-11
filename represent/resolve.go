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
		return resolveDataLink(g, mid, src, es)
		//case SpecCommit:
		//return resolveSpecCommit(g, src, es)
		//case SpecLocalLogic:
		//return resolveSpecLocalLogic(g, src, es)
	}

	return StandardEdge{}, false
}

func resolveEnvLink(g *CoreGraph, mid int, src vtTuple, es interpret.EnvLink) (e StandardEdge, success bool) {
	e = StandardEdge{
		Source: src.id,
		Props:  ps.NewMap(),
		EType:  "envlink",
	}

	// First, check if this vertex already *has* an outbound envlink; semantics dictate there can be only one.
	src.oe.ForEach(func(_ string, val ps.Any) {
		edge := val.(StandardEdge)
		if edge.EType == "envlink" {
			success = true
			// FIXME need a way to cut out early
			e = edge
		}
	})

	// Whether we find a match or not, have to merge in the EnvLink
	if es.Address.Hostname != "" {
		e.Props = e.Props.Set("hostname", Property{MsgSrc: mid, Value: es.Address.Hostname})
	}
	if es.Address.Ipv4 != "" {
		e.Props = e.Props.Set("ipv4", Property{MsgSrc: mid, Value: es.Address.Ipv4})
	}
	if es.Address.Ipv6 != "" {
		e.Props = e.Props.Set("ipv6", Property{MsgSrc: mid, Value: es.Address.Ipv6})
	}
	if es.Nick != "" {
		e.Props = e.Props.Set("nick", Property{MsgSrc: mid, Value: es.Nick})
	}

	// If we already found the matching edge, bail out now
	if success {
		return e, true
	}

	rv := g.VerticesWith(qbv("environment"))
	var envid int
	for _, vt := range rv {
		// TODO this'll be cross-package eventually - reorg needed
		if matchEnvLink(e.Props, vt.v.Props()) {
			envid = vt.id
			break
		}
	}

	// No matching env found, bail out
	if envid == 0 {
		return e, false
	}

	return e, success
}

func resolveDataLink(g *CoreGraph, mid int, src vtTuple, es interpret.DataLink) (e StandardEdge, success bool) {
	e = StandardEdge{
		Source: src.id,
		Props:  ps.NewMap(),
		EType:  "datalink",
	}

	// DataLinks have a 'name' field that is expected to be unique for the source, if present
	if es.Name != "" {
		// TODO 'name' is a traditional unique key; a change in it inherently denotes a new edge. how to handle this?
		// FIXME this approach just always updates the mid, which is weird?
		e.Props = e.Props.Set("name", Property{MsgSrc: mid, Value: es.Name})

		src.oe.ForEach(func(_ string, val ps.Any) {
			edge := val.(StandardEdge)
			if name, exists := edge.Props.Lookup("name"); exists && edge.EType == "datalink" && name == es.Name {
				// FIXME need a way to cut out early
				success = true
				e = edge
			}
		})
	}

	if es.Type != "" {
		e.Props = e.Props.Set("type", Property{MsgSrc: mid, Value: es.Type})
	}
	if es.Subset != "" {
		e.Props = e.Props.Set("subset", Property{MsgSrc: mid, Value: es.Subset})
	}
	if es.Interaction != "" {
		e.Props = e.Props.Set("interaction", Property{MsgSrc: mid, Value: es.Interaction})
	}

	// Special bits: if we have ConnUnix data, eliminate ConnNet data, and vice-versa.
	var isLocal bool
	if es.ConnUnix.Path != "" {
		isLocal = true
		e.Props = e.Props.Set("path", Property{MsgSrc: mid, Value: es.ConnUnix.Path})
		e.Props = e.Props.Delete("hostname")
		e.Props = e.Props.Delete("ipv4")
		e.Props = e.Props.Delete("ipv6")
		e.Props = e.Props.Delete("port")
		e.Props = e.Props.Delete("proto")
	} else {
		e.Props = e.Props.Set("port", Property{MsgSrc: mid, Value: es.ConnNet.Port})
		e.Props = e.Props.Set("proto", Property{MsgSrc: mid, Value: es.ConnNet.Proto})

		// can only be one of hostname, ipv4 or ipv6
		if es.ConnNet.Hostname != "" {
			e.Props = e.Props.Set("hostname", Property{MsgSrc: mid, Value: es.ConnNet.Hostname})
		} else if es.ConnNet.Ipv4 != "" {
			e.Props = e.Props.Set("ipv4", Property{MsgSrc: mid, Value: es.ConnNet.Ipv4})
		} else {
			e.Props = e.Props.Set("ipv6", Property{MsgSrc: mid, Value: es.ConnNet.Ipv6})
		}
	}

	if success {
		return e, true
	}

	var sock vtTuple
	var rv []vtTuple // just for reuse
	// If net, must scan; if local, a bit easier.
	if !isLocal {
		// First, find the environment vertex
		rv = g.VerticesWith(qbv("environment"))
		var envid int
		for _, vt := range rv {
			// TODO this'll be cross-package eventually - reorg needed
			if matchAddress(e.Props, vt.v.Props()) {
				envid = vt.id
				break
			}
		}

		// No matching env found, bail out
		if envid == 0 {
			return e, false
		}

		// Now, walk the environment's edges to find the vertex representing the port
		ef := EdgeFilter{EType: "envlink"}
		vf := VertexFilter{VType: "comm", Props: []PropQ{
			{"port", es.ConnNet.Port},
			{"proto", es.ConnNet.Proto},
		}}
		rv = g.PredecessorsWith(envid, BothFilter{vf, ef})

		if len(rv) != 1 {
			return e, false
		}
		sock = rv[0]
	} else {
		envid, exists := findEnv(g, src)
		if !exists {
			// this is would be a pretty weird case
			return e, false
		}

		// Walk the graph to find the vertex representing the unix socket
		ef := EdgeFilter{EType: "envlink"}
		vf := VertexFilter{VType: "comm", Props: []PropQ{{"path", es.ConnUnix.Path}}}
		rv = g.PredecessorsWith(envid, BothFilter{vf, ef})
		if len(rv) != 1 {
			return e, false
		}
		sock = rv[0]
	}

	// With sock in hand, now find its proc
	rv = g.SuccessorsWith(sock.id, qbv("process").bf())
	if len(rv) != 1 {
		// TODO could/will we ever allow >1?
		return e, false
	}

	rv = g.SuccessorsWith(rv[0].id, qbv("dataset").bf())
	if len(rv) != 1 {
		return e, false
	}
	dataset := rv[0]

	// if the spec indicates a subset, find it
	if es.Subset != "" {
		rv = g.SuccessorsWith(rv[0].id, qbv("dataset", "name", es.Subset).bf())
		if len(rv) != 1 {
			return e, false
		}
		dataset = rv[0]
	}

	// FIXME only recording the final target id is totally broken; see https://github.com/sdboyer/pipeviz/issues/37

	// Aaaand we found our target.
	e.Target = dataset.id
	return e, true
}

//func resolveSpecCommit(g *CoreGraph, src vtTuple, e SpecCommit) (StandardEdge, bool) {
//g.Vertices(func(vtx Vertex, id int) bool {})
//}

//func resolveSpecLocalLogic(g *CoreGraph, src vtTuple, e SpecLocalLogic) (StandardEdge, bool) {
//g.Vertices(func(vtx Vertex, id int) bool {})
//}

// Searches the given vertex's out-edges to find its environment's vertex id.
func findEnv(g *CoreGraph, vt vtTuple) (id int, success bool) {
	vt.oe.ForEach(func(_ string, val ps.Any) {
		edge := val.(StandardEdge)
		if edge.EType == "envlink" {
			success = true
			// FIXME need a way to cut out early
			id = edge.Target
		}
	})

	return
}
