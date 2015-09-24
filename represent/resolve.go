package represent

import (
	log "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/represent/helpers"
	"github.com/tag1consulting/pipeviz/represent/types"
)

// Attempts to resolve an EdgeSpec into a real edge. This process has two steps:
//
// 1. Finding the target node.
// 2. Seeing if this edge already exists between source and target.
//
// It is the responsibility of the edge spec's type handler to determine what "if an edge
// already exists" means, as well as whether to overwrite/merge or duplicate the edge in such a case.
func Resolve(g CoreGraph, mid uint64, src types.VertexTuple, d types.EdgeSpec) (types.StdEdge, bool) {
	switch es := d.(type) {
	case interpret.EnvLink:
		return resolveEnvLink(g, mid, src, es)
	case interpret.DataLink:
		return resolveDataLink(g, mid, src, es)
	case SpecCommit:
		return resolveSpecCommit(g, mid, src, es)
	case SpecGitCommitParent:
		return resolveSpecGitCommitParent(g, mid, src, es)
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
		log.WithFields(log.Fields{
			"system":   "edge-resolution",
			"edgespec": d,
		}).Error("No resolution method for provided EdgeSpec") // TODO is it ok to not halt here?
	}

	return types.StdEdge{}, false
}

func resolveEnvLink(g CoreGraph, mid uint64, src types.VertexTuple, es interpret.EnvLink) (e types.StdEdge, success bool) {
	_, e, success = findEnv(g, src)

	// Whether we find a match or not, have to merge in the EnvLink
	if es.Address.Hostname != "" {
		e.Props = e.Props.Set("hostname", types.Property{MsgSrc: mid, Value: es.Address.Hostname})
	}
	if es.Address.Ipv4 != "" {
		e.Props = e.Props.Set("ipv4", types.Property{MsgSrc: mid, Value: es.Address.Ipv4})
	}
	if es.Address.Ipv6 != "" {
		e.Props = e.Props.Set("ipv6", types.Property{MsgSrc: mid, Value: es.Address.Ipv6})
	}
	if es.Nick != "" {
		e.Props = e.Props.Set("nick", types.Property{MsgSrc: mid, Value: es.Nick})
	}

	// If we already found the matching edge, bail out now
	if success {
		return
	}

	rv := g.VerticesWith(helpers.Qbv(types.VType("environment")))
	for _, vt := range rv {
		// TODO this'll be cross-package eventually - reorg needed
		if matchEnvLink(e.Props, vt.Vertex.Props()) {
			success = true
			e.Target = vt.ID
			break
		}
	}

	return
}

func resolveDataLink(g CoreGraph, mid uint64, src types.VertexTuple, es interpret.DataLink) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "datalink",
	}

	// DataLinks have a 'name' field that is expected to be unique for the source, if present
	if es.Name != "" {
		// TODO 'name' is a traditional unique key; a change in it inherently denotes a new edge. how to handle this?
		// FIXME this approach just always updates the mid, which is weird?
		e.Props = e.Props.Set("name", types.Property{MsgSrc: mid, Value: es.Name})

		re := g.OutWith(src.ID, helpers.Qbe(types.EType("datalink"), "name", es.Name))
		if len(re) == 1 {
			success = true
			e = re[0]
		}
	}

	if es.Type != "" {
		e.Props = e.Props.Set("type", types.Property{MsgSrc: mid, Value: es.Type})
	}
	if es.Subset != "" {
		e.Props = e.Props.Set("subset", types.Property{MsgSrc: mid, Value: es.Subset})
	}
	if es.Interaction != "" {
		e.Props = e.Props.Set("interaction", types.Property{MsgSrc: mid, Value: es.Interaction})
	}

	// Special bits: if we have ConnUnix data, eliminate ConnNet data, and vice-versa.
	var isLocal bool
	if es.ConnUnix.Path != "" {
		isLocal = true
		e.Props = e.Props.Set("path", types.Property{MsgSrc: mid, Value: es.ConnUnix.Path})
		e.Props = e.Props.Delete("hostname")
		e.Props = e.Props.Delete("ipv4")
		e.Props = e.Props.Delete("ipv6")
		e.Props = e.Props.Delete("port")
		e.Props = e.Props.Delete("proto")
	} else {
		e.Props = e.Props.Set("port", types.Property{MsgSrc: mid, Value: es.ConnNet.Port})
		e.Props = e.Props.Set("proto", types.Property{MsgSrc: mid, Value: es.ConnNet.Proto})

		// can only be one of hostname, ipv4 or ipv6
		if es.ConnNet.Hostname != "" {
			e.Props = e.Props.Set("hostname", types.Property{MsgSrc: mid, Value: es.ConnNet.Hostname})
		} else if es.ConnNet.Ipv4 != "" {
			e.Props = e.Props.Set("ipv4", types.Property{MsgSrc: mid, Value: es.ConnNet.Ipv4})
		} else {
			e.Props = e.Props.Set("ipv6", types.Property{MsgSrc: mid, Value: es.ConnNet.Ipv6})
		}
	}

	if success {
		return
	}

	var sock types.VertexTuple
	var rv []types.VertexTuple // just for reuse
	// If net, must scan; if local, a bit easier.
	if !isLocal {
		// First, find the environment vertex
		rv = g.VerticesWith(helpers.Qbv(types.VType("environment")))
		var envid int
		for _, vt := range rv {
			// TODO matchAddress() func will need to be reorged to cross-package eventually - export!
			if matchAddress(e.Props, vt.Vertex.Props()) {
				envid = vt.ID
				break
			}
		}

		// No matching env found, bail out
		if envid == 0 {
			return
		}

		// Now, walk the environment's edges to find the vertex representing the port
		rv = g.PredecessorsWith(envid, helpers.Qbv(types.VType("comm"), "type", "port", "port", es.ConnNet.Port).And(helpers.Qbe(types.EType("envlink"))))

		if len(rv) != 1 {
			return
		}
		sock = rv[0]

		// With sock in hand, now find its proc
		rv = g.PredecessorsWith(sock.ID, helpers.Qbe(types.EType("listening"), "proto", es.ConnNet.Proto).And(helpers.Qbv(types.VType("process"))))
		if len(rv) != 1 {
			// TODO could/will we ever allow >1?
			return
		}
	} else {
		envid, _, exists := findEnv(g, src)

		if !exists {
			// this is would be a pretty weird case
			return
		}

		// Walk the graph to find the vertex representing the unix socket
		rv = g.PredecessorsWith(envid, helpers.Qbv(types.VType("comm"), "path", es.ConnUnix.Path).And(helpers.Qbe(types.EType("envlink"))))
		if len(rv) != 1 {
			return
		}
		sock = rv[0]

		// With sock in hand, now find its proc
		rv = g.PredecessorsWith(sock.ID, helpers.Qbv(types.VType("process")).And(helpers.Qbe(types.EType("listening"))))
		if len(rv) != 1 {
			// TODO could/will we ever allow >1?
			return
		}
	}

	rv = g.SuccessorsWith(rv[0].ID, helpers.Qbv(types.VType("parent-dataset")))
	// FIXME this absolutely could be more than 1
	if len(rv) != 1 {
		return
	}
	dataset := rv[0]

	// if the spec indicates a subset, find it
	if es.Subset != "" {
		rv = g.PredecessorsWith(rv[0].ID, helpers.Qbv(types.VType("dataset"), "name", es.Subset).And(helpers.Qbe(types.EType("dataset-hierarchy"))))
		if len(rv) != 1 {
			return
		}
		dataset = rv[0]
	}

	// FIXME only recording the final target id is totally broken; see https://github.com/tag1consulting/pipeviz/issues/37

	// Aaaand we found our target.
	success = true
	e.Target = dataset.ID
	return
}

func resolveSpecCommit(g CoreGraph, mid uint64, src types.VertexTuple, es SpecCommit) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "version",
	}
	e.Props = e.Props.Set("sha1", types.Property{MsgSrc: mid, Value: es.Sha1})

	re := g.OutWith(src.ID, helpers.Qbe(types.EType("version")))
	if len(re) > 0 {
		sha1, _ := re[0].Props.Lookup("sha1")
		e.ID = re[0].ID // FIXME setting the id to non-0 AND failing is currently unhandled
		if sha1.(types.Property).Value == es.Sha1 {
			success = true
			e.Target = re[0].Target
		} else {
			rv := g.VerticesWith(helpers.Qbv(types.VType("commit"), "sha1", es.Sha1))
			if len(rv) == 1 {
				success = true
				e.Target = rv[0].ID
			}
		}
	} else {
		rv := g.VerticesWith(helpers.Qbv(types.VType("commit"), "sha1", es.Sha1))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}

func resolveSpecGitCommitParent(g CoreGraph, mid uint64, src types.VertexTuple, es SpecGitCommitParent) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "parent-commit",
	}

	re := g.OutWith(src.ID, helpers.Qbe(types.EType("parent-commit"), "pnum", es.ParentNum))
	if len(re) > 0 {
		success = true
		e.Target = re[0].Target
		e.ID = re[0].ID
	} else {
		rv := g.VerticesWith(helpers.Qbv(types.VType("commit"), "sha1", es.Sha1))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].ID
			e.Props = e.Props.Set("pnum", types.Property{MsgSrc: mid, Value: es.ParentNum})
			e.Props = e.Props.Set("sha1", types.Property{MsgSrc: mid, Value: es.Sha1})
		}
	}

	return
}

func resolveSpecLocalLogic(g CoreGraph, mid uint64, src types.VertexTuple, es SpecLocalLogic) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "logic-link",
	}

	// search for existing link
	re := g.OutWith(src.ID, helpers.Qbe(types.EType("logic-link"), "path", es.Path))
	if len(re) == 1 {
		// TODO don't set the path prop again, it's the unique id...meh, same question here w/uniqueness as above
		success = true
		e = re[0]
		return
	}

	// no existing link found, search for proc directly
	envid, _, _ := findEnv(g, src)
	rv := g.PredecessorsWith(envid, helpers.Qbv(types.VType("logic-state"), "path", es.Path))
	if len(rv) == 1 {
		success = true
		e.Target = rv[0].ID
	}

	return
}

func resolveNetListener(g CoreGraph, mid uint64, src types.VertexTuple, es SpecNetListener) (e types.StdEdge, success bool) {
	// check for existing edge; this one is quite straightforward
	re := g.OutWith(src.ID, helpers.Qbe(types.EType("listening"), "type", "port", "port", es.Port, "proto", es.Proto))
	if len(re) == 1 {
		return re[0], true
	}

	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "listening",
	}

	e.Props = e.Props.Set("port", types.Property{MsgSrc: mid, Value: es.Port})
	e.Props = e.Props.Set("proto", types.Property{MsgSrc: mid, Value: es.Proto})

	envid, _, hasenv := findEnv(g, src)
	if hasenv {
		rv := g.PredecessorsWith(envid, helpers.Qbv(types.VType("comm"), "type", "port", "port", es.Port))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}

func resolveUnixDomainListener(g CoreGraph, mid uint64, src types.VertexTuple, es SpecUnixDomainListener) (e types.StdEdge, success bool) {
	// check for existing edge; this one is quite straightforward
	re := g.OutWith(src.ID, helpers.Qbe(types.EType("listening"), "type", "unix", "path", es.Path))
	if len(re) == 1 {
		return re[0], true
	}

	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "listening",
	}

	e.Props = e.Props.Set("path", types.Property{MsgSrc: mid, Value: es.Path})

	envid, _, hasenv := findEnv(g, src)
	if hasenv {
		rv := g.PredecessorsWith(envid, helpers.Qbv(types.VType("comm"), "type", "unix", "path", es.Path))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}

func resolveSpecDatasetHierarchy(g CoreGraph, mid uint64, src types.VertexTuple, es SpecDatasetHierarchy) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "dataset-hierarchy",
	}
	e.Props = e.Props.Set("parent", types.Property{MsgSrc: mid, Value: es.NamePath[0]})

	// check for existing link - there can be only be one
	re := g.OutWith(src.ID, helpers.Qbe(types.EType("dataset-hierarchy")))
	if len(re) == 1 {
		success = true
		e = re[0]
		// TODO semantics should preclude this from being able to change, but doing it dirty means force-setting it anyway for now
		e.Props = e.Props.Set("parent", types.Property{MsgSrc: mid, Value: es.NamePath[0]})
		return
	}

	// no existing link found; search for proc directly
	envid, _, _ := findEnv(g, src)
	rv := g.PredecessorsWith(envid, helpers.Qbv(types.VType("parent-dataset"), "name", es.NamePath[0]))
	if len(rv) != 0 { // >1 shouldn't be possible
		success = true
		e.Target = rv[0].ID
	}

	return
}

func resolveSpecParentDataset(g CoreGraph, mid uint64, src types.VertexTuple, es SpecParentDataset) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "dataset-gateway",
	}
	e.Props = e.Props.Set("name", types.Property{MsgSrc: mid, Value: es.Name})

	// check for existing link - there can be only be one
	re := g.OutWith(src.ID, helpers.Qbe(types.EType("dataset-gateway")))
	if len(re) == 1 {
		success = true
		e = re[0]
		// TODO semantics should preclude this from being able to change, but doing it dirty means force-setting it anyway for now
	} else {

		// no existing link found; search for proc directly
		envid, _, _ := findEnv(g, src)
		rv := g.PredecessorsWith(envid, helpers.Qbv(types.VType("parent-dataset"), "name", es.Name))
		if len(rv) != 0 { // >1 shouldn't be possible
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}

func resolveDataProvenance(g CoreGraph, mid uint64, src types.VertexTuple, es interpret.DataProvenance) (e types.StdEdge, success bool) {
	// FIXME this presents another weird case where "success" is not binary. We *could*
	// find an already-existing data-provenance edge, but then have some net-addr params
	// change which cause it to fail to resolve to an environment. If we call that successful,
	// then it won't try to resolve again later...though, hm, just call it unsuccessful and
	// then try again one more time. Maybe it is fine. THINK IT THROUGH.

	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "data-provenance",
	}
	e.Props = assignAddress(mid, es.Address, e.Props, false)

	re := g.OutWith(src.ID, helpers.Qbe(types.EType("data-provenance")))
	if len(re) == 1 {
		// TODO wasteful, blargh
		reresolve := mapValEqAnd(e.Props, re[0].Props, "hostname", "ipv4", "ipv6")

		e = re[0]
		if es.SnapTime != "" {
			e.Props = e.Props.Set("snap-time", types.Property{MsgSrc: mid, Value: es.SnapTime})
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

func resolveDataAlpha(g CoreGraph, mid uint64, src types.VertexTuple, es interpret.DataAlpha) (e types.StdEdge, success bool) {
	// TODO this makes a loop...are we cool with that?
	success = true // impossible to fail here
	e = types.StdEdge{
		Source: src.ID,
		Target: src.ID,
		Props:  ps.NewMap(),
		EType:  "data-provenance",
	}

	re := g.OutWith(src.ID, helpers.Qbe(types.EType("data-provenance")))
	if len(re) == 1 {
		e = re[0]
	}

	return
}

// Searches the given vertex's out-edges to find its environment's vertex id.
//
// Also conveniently initializes a StandardEdge to the standard zero-state for an envlink.
func findEnv(g CoreGraph, vt types.VertexTuple) (vid int, edge types.StdEdge, success bool) {
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
