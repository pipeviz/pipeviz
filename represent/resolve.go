package represent

import (
	"fmt"

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
func Resolve(g CoreGraph, mid int, src vtTuple, d EdgeSpec) (StandardEdge, bool) {
	switch es := d.(type) {
	case interpret.EnvLink:
		return resolveEnvLink(g, mid, src, es)
	case interpret.DataLink:
		return resolveDataLink(g, mid, src, es)
	case SpecCommit:
		return resolveSpecCommit(g, mid, src, es)
	case SpecLocalLogic:
		return resolveSpecLocalLogic(g, mid, src, es)
	case SpecNetListener:
		return resolveNetListener(g, mid, src, es)
	case SpecUnixDomainListener:
		return resolveUnixDomainListener(g, mid, src, es)
	case SpecDatasetHierarchy:
		return resolveSpecDatasetHierarchy(g, mid, src, es)
	case SpecParentDataset:
		return resolveSpecParentDataset(g, mid, src, es)
	case interpret.DataProvenance:
		return resolveDataProvenance(g, mid, src, es)
	case interpret.DataAlpha:
		return resolveDataAlpha(g, mid, src, es)
	default:
		panic(fmt.Sprintf("OMG WUT NO SUPPORT FER %T", d)) // FIXME panic lulz
	}

	return StandardEdge{}, false
}

func resolveEnvLink(g CoreGraph, mid int, src vtTuple, es interpret.EnvLink) (e StandardEdge, success bool) {
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

	rv := g.VerticesWith(qbv(VType("environment")))
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

func resolveDataLink(g CoreGraph, mid int, src vtTuple, es interpret.DataLink) (e StandardEdge, success bool) {
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

		re := g.OutWith(src.id, qbe(EType("datalink"), "name", es.Name))
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
		rv = g.VerticesWith(qbv(VType("environment")))
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
		rv = g.PredecessorsWith(envid, qbv(VType("comm"), "port", es.ConnNet.Port, "proto", es.ConnNet.Proto).and(qbe(EType("envlink"))))

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
		rv = g.PredecessorsWith(envid, qbv(VType("comm"), "path", es.ConnUnix).and(qbe(EType("envlink"))))
		if len(rv) != 1 {
			return
		}
		sock = rv[0]
	}

	// With sock in hand, now find its proc
	rv = g.SuccessorsWith(sock.id, qbv(VType("process")))
	if len(rv) != 1 {
		// TODO could/will we ever allow >1?
		return
	}

	rv = g.SuccessorsWith(rv[0].id, qbv(VType("dataset")))
	if len(rv) != 1 {
		return
	}
	dataset := rv[0]

	// if the spec indicates a subset, find it
	if es.Subset != "" {
		rv = g.SuccessorsWith(rv[0].id, qbv(VType("dataset"), "name", es.Subset))
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

func resolveSpecCommit(g CoreGraph, mid int, src vtTuple, es SpecCommit) (e StandardEdge, success bool) {
	e = StandardEdge{
		Source: src.id,
		Props:  ps.NewMap(),
		EType:  "version",
	}

	re := g.OutWith(src.id, qbe(EType("version"), "sha1", es.Sha1))
	// TODO could there ever be >1?
	if len(re) == 1 {
		success = true
		e.Target = re[0].Target
		e.id = re[0].id
	} else {
		rv := g.VerticesWith(qbv(VType("commit"), "sha1", es.Sha1))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].id
		}
	}

	return
}

func resolveSpecLocalLogic(g CoreGraph, mid int, src vtTuple, es SpecLocalLogic) (e StandardEdge, success bool) {
	e = StandardEdge{
		Source: src.id,
		Props:  ps.NewMap(),
		EType:  "logic-link",
	}

	// search for existing link
	re := g.OutWith(src.id, qbe(EType("logic-link"), "path", es.Path))
	if len(re) == 1 {
		// TODO don't set the path prop again, it's the unique id...meh, same question here w/uniqueness as above
		success = true
		e = re[0]
		return
	}

	// no existing link found, search for proc directly
	envid, _, _ := findEnv(g, src)
	rv := g.PredecessorsWith(envid, qbv(VType("logic-state"), "path", es.Path))
	if len(rv) == 1 {
		success = true
		e.Target = rv[0].id
	}

	return
}

func resolveNetListener(g CoreGraph, mid int, src vtTuple, es SpecNetListener) (e StandardEdge, success bool) {
	// check for existing edge; this one is quite straightforward
	re := g.OutWith(src.id, qbe(EType("listening"), "type", "port", "port", es.Port, "proto", es.Proto))
	if len(re) == 1 {
		return re[0], success
	}

	e = StandardEdge{
		Source: src.id,
		Props:  ps.NewMap(),
		EType:  "listening",
	}

	e.Props = e.Props.Set("port", Property{MsgSrc: mid, Value: es.Port})
	e.Props = e.Props.Set("proto", Property{MsgSrc: mid, Value: es.Proto})

	envid, _, hasenv := findEnv(g, src)
	if hasenv {
		rv := g.PredecessorsWith(envid, qbv(VType("comm"), "type", "port", "port", es.Port))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].id
		}
	}

	return
}

func resolveUnixDomainListener(g CoreGraph, mid int, src vtTuple, es SpecUnixDomainListener) (e StandardEdge, success bool) {
	// check for existing edge; this one is quite straightforward
	re := g.OutWith(src.id, qbe(EType("listening"), "type", "unix", "path", es.Path))
	if len(re) == 1 {
		return re[0], success
	}

	e = StandardEdge{
		Source: src.id,
		Props:  ps.NewMap(),
		EType:  "listening",
	}

	e.Props = e.Props.Set("path", Property{MsgSrc: mid, Value: es.Path})

	envid, _, hasenv := findEnv(g, src)
	if hasenv {
		rv := g.PredecessorsWith(envid, qbv(VType("comm"), "type", "unix", "path", es.Path))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].id
		}
	}

	return
}

func resolveSpecDatasetHierarchy(g CoreGraph, mid int, src vtTuple, es SpecDatasetHierarchy) (e StandardEdge, success bool) {
	e = StandardEdge{
		Source: src.id,
		Props:  ps.NewMap(),
		EType:  "dataset-hierarchy",
	}
	e.Props = e.Props.Set("parent", Property{MsgSrc: mid, Value: es.NamePath[0]})

	// check for existing link - there can be only be one
	re := g.OutWith(src.id, qbe(EType("dataset-hierarchy")))
	if len(re) == 1 {
		success = true
		e = re[0]
		// TODO semantics should preclude this from being able to change, but doing it dirty means force-setting it anyway for now
		e.Props = e.Props.Set("parent", Property{MsgSrc: mid, Value: es.NamePath[0]})
		return
	}

	// no existing link found; search for proc directly
	envid, _, _ := findEnv(g, src)
	rv := g.PredecessorsWith(envid, qbv(VType("parent-dataset"), "name", es.NamePath[0]))
	if len(rv) != 0 { // >1 shouldn't be possible
		success = true
		e.Target = rv[0].id
	}

	return
}

func resolveSpecParentDataset(g CoreGraph, mid int, src vtTuple, es SpecParentDataset) (e StandardEdge, success bool) {
	e = StandardEdge{
		Source: src.id,
		Props:  ps.NewMap(),
		EType:  "dataset-gateway",
	}
	e.Props = e.Props.Set("name", Property{MsgSrc: mid, Value: es.Name})

	// check for existing link - there can be only be one
	re := g.OutWith(src.id, qbe(EType("dataset-gateway")))
	if len(re) == 1 {
		success = true
		e = re[0]
		// TODO semantics should preclude this from being able to change, but doing it dirty means force-setting it anyway for now
	} else {

		// no existing link found; search for proc directly
		envid, _, _ := findEnv(g, src)
		rv := g.PredecessorsWith(envid, qbv(VType("parent-dataset"), "name", es.Name))
		if len(rv) != 0 { // >1 shouldn't be possible
			success = true
			e.Target = rv[0].id
		}
	}

	return
}

func resolveDataProvenance(g CoreGraph, mid int, src vtTuple, es interpret.DataProvenance) (e StandardEdge, success bool) {
	// FIXME this presents another weird case where "success" is not binary. We *could*
	// find an already-existing data-provenance edge, but then have some net-addr params
	// change which cause it to fail to resolve to an environment. If we call that successful,
	// then it won't try to resolve again later...though, hm, just call it unsuccessful and
	// then try again one more time. Maybe it is fine. THINK IT THROUGH.

	e = StandardEdge{
		Source: src.id,
		Props:  ps.NewMap(),
		EType:  "data-provenance",
	}
	e.Props = assignAddress(mid, es.Address, e.Props, false)

	re := g.OutWith(src.id, qbe(EType("data-provenance")))
	if len(re) == 1 {
		// TODO wasteful, blargh
		reresolve := mapValEqAnd(e.Props, re[0].Props, "hostname", "ipv4", "ipv6")

		e = re[0]
		if es.SnapTime != "" {
			e.Props = e.Props.Set("snap-time", Property{MsgSrc: mid, Value: es.SnapTime})
		}

		if reresolve {
			e.Props = assignAddress(mid, es.Address, e.Props, true)
		} else {
			return e, true
		}
	}

	envid, found := FindEnvironment(g, e.Props)
	if !found {
		// TODO returning this already-modified edge necessitates that the core system
		// disregard 'failed' edges. which should be fine, that should be a guarantee
		return e, false
	}

	e.Target, success = FindDataset(g, envid, es.Dataset)
	return
}

func resolveDataAlpha(g CoreGraph, mid int, src vtTuple, es interpret.DataAlpha) (e StandardEdge, success bool) {
	// TODO this makes a loop...are we cool with that?
	success = true // impossible to fail here
	e = StandardEdge{
		Source: src.id,
		Target: src.id,
		Props:  ps.NewMap(),
		EType:  "data-provenance",
	}

	re := g.OutWith(src.id, qbe(EType("data-provenance")))
	if len(re) == 1 {
		e = re[0]
	}

	return
}

// Searches the given vertex's out-edges to find its environment's vertex id.
//
// Also conveniently initializes a StandardEdge to the standard zero-state for an envlink.
func findEnv(g CoreGraph, vt vtTuple) (vid int, edge StandardEdge, success bool) {
	edge = StandardEdge{
		Source: vt.id,
		Props:  ps.NewMap(),
		EType:  "envlink",
	}

	if vt.id != 0 {
		re := g.OutWith(vt.id, qbe(EType("envlink")))
		if len(re) == 1 {
			vid, edge, success = re[0].Target, re[0], true
		}
	}

	return
}
