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
	case SpecCommit:
		return resolveSpecCommit(g, mid, src, es)
	case SpecLocalLogic:
		return resolveSpecLocalLogic(g, mid, src, es)
	case interpret.DataAlpha:
		return resolveDataAlpha(g, mid, src, es)
	}

	return StandardEdge{}, false
}

func resolveEnvLink(g *CoreGraph, mid int, src vtTuple, es interpret.EnvLink) (e StandardEdge, success bool) {
	_, e, success = findEnv(g, src)

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
		return
	}

	rv := g.VerticesWith(qbv("environment"))
	for _, vt := range rv {
		// TODO this'll be cross-package eventually - reorg needed
		if matchEnvLink(e.Props, vt.v.Props()) {
			success = true
			e.Target = vt.id
			break
		}
	}

	return
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

		re := g.OutWith(src.id, qbe("datalink", "name", es.Name))
		if len(re) == 1 {
			success = true
			e = re[0]
		}
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
		return
	}

	var sock vtTuple
	var rv []vtTuple // just for reuse
	// If net, must scan; if local, a bit easier.
	if !isLocal {
		// First, find the environment vertex
		rv = g.VerticesWith(qbv("environment"))
		var envid int
		for _, vt := range rv {
			// TODO matchAddress() func will need to be reorged to cross-package eventually - export!
			if matchAddress(e.Props, vt.v.Props()) {
				envid = vt.id
				break
			}
		}

		// No matching env found, bail out
		if envid == 0 {
			return
		}

		// Now, walk the environment's edges to find the vertex representing the port
		//ef := edgeFilter{EType: "envlink"}
		//vf := vertexFilter{VType: "comm", Props: []PropQ{
		//{"port", es.ConnNet.Port},
		//{"proto", es.ConnNet.Proto},
		//}}
		rv = g.PredecessorsWith(envid, qbv("comm", "port", es.ConnNet.Port, "proto", es.ConnNet.Proto).and(qbe("envlink")))

		if len(rv) != 1 {
			return
		}
		sock = rv[0]
	} else {
		envid, _, exists := findEnv(g, src)

		if !exists {
			// this is would be a pretty weird case
			return
		}

		// Walk the graph to find the vertex representing the unix socket
		//ef := edgeFilter{EType: "envlink"}
		//vf := vertexFilter{VType: "comm", Props: []PropQ{{"path", es.ConnUnix.Path}}}
		rv = g.PredecessorsWith(envid, qbv("comm", "path", es.ConnUnix).and(qbe("envlink")))
		if len(rv) != 1 {
			return
		}
		sock = rv[0]
	}

	// With sock in hand, now find its proc
	rv = g.SuccessorsWith(sock.id, qbv("process"))
	if len(rv) != 1 {
		// TODO could/will we ever allow >1?
		return
	}

	rv = g.SuccessorsWith(rv[0].id, qbv("dataset"))
	if len(rv) != 1 {
		return
	}
	dataset := rv[0]

	// if the spec indicates a subset, find it
	if es.Subset != "" {
		rv = g.SuccessorsWith(rv[0].id, qbv("dataset", "name", es.Subset))
		if len(rv) != 1 {
			return
		}
		dataset = rv[0]
	}

	// FIXME only recording the final target id is totally broken; see https://github.com/sdboyer/pipeviz/issues/37

	// Aaaand we found our target.
	success = true
	e.Target = dataset.id
	return
}

func resolveSpecCommit(g *CoreGraph, mid int, src vtTuple, es SpecCommit) (e StandardEdge, success bool) {
	e = StandardEdge{
		Source: src.id,
		Props:  ps.NewMap(),
		EType:  "version",
	}

	re := g.OutWith(src.id, qbe("version", "sha1", es.Sha1))
	// TODO could there ever be >1?
	if len(re) == 1 {
		success = true
		e.Target = re[0].Target
		e.id = re[0].id
	}

	return
}

func resolveSpecLocalLogic(g *CoreGraph, mid int, src vtTuple, es SpecLocalLogic) (e StandardEdge, success bool) {
	e = StandardEdge{
		Source: src.id,
		Props:  ps.NewMap(),
		EType:  "logic-link",
	}

	// search for existing link
	re := g.OutWith(src.id, qbe("logic-link", "path", es.Path))
	if len(re) == 1 {
		// TODO don't set the path prop again, it's the unique id...meh, same question here w/uniqueness as above
		success = true
		e = re[0]
		return
	}

	// no existing link found, search for proc directly
	envid, _, _ := findEnv(g, src)
	rv := g.PredecessorsWith(envid, qbv("logic-state", "path", es.Path))
	if len(rv) == 1 {
		success = true
		e.Target = rv[0].id
	}

	return
}

func resolveDataAlpha(g *CoreGraph, mid int, src vtTuple, es interpret.DataAlpha) (e StandardEdge, success bool) {
	// TODO this makes a loop...are we cool with that?
	success = true // impossible to fail here
	e = StandardEdge{
		Source: src.id,
		Target: src.id,
		Props:  ps.NewMap(),
		EType:  "data-provenance",
	}

	re := g.OutWith(src.id, qbe("data-provenance"))
	if len(re) == 1 {
		e = re[0]
	}

	return
}

// Searches the given vertex's out-edges to find its environment's vertex id.
//
// Also conveniently initializes a StandardEdge to the standard zero-state for an envlink.
func findEnv(g *CoreGraph, vt vtTuple) (vid int, edge StandardEdge, success bool) {
	edge = StandardEdge{
		Source: vt.id,
		Props:  ps.NewMap(),
		EType:  "envlink",
	}

	re := g.OutWith(vt.id, qbe("envlink"))
	if len(re) == 1 {
		vid, edge, success = re[0].Target, re[0], true
	}

	return
}
