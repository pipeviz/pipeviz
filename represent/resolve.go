package represent

import (
	log "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/represent/helpers"
	"github.com/tag1consulting/pipeviz/represent/types"
)

type resolver interface {
	Resolve(g types.CoreGraph, mid uint64, src types.VertexTuple) (e types.StdEdge, success bool)
}

// Attempts to resolve an EdgeSpec into a real edge. This process has two steps:
//
// 1. Finding the target node.
// 2. Seeing if this edge already exists between source and target.
//
// It is the responsibility of the edge spec's type handler to determine what "if an edge
// already exists" means, as well as whether to overwrite/merge or duplicate the edge in such a case.
func Resolve(g types.CoreGraph, mid uint64, src types.VertexTuple, d types.EdgeSpec) (types.StdEdge, bool) {
	if r, ok := d.(resolver); ok {
		return r.Resolve(g, mid, src)
	}

	switch es := d.(type) {
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

func resolveDataProvenance(g types.CoreGraph, mid uint64, src types.VertexTuple, es interpret.DataProvenance) (e types.StdEdge, success bool) {
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

func resolveDataAlpha(g types.CoreGraph, mid uint64, src types.VertexTuple, es interpret.DataAlpha) (e types.StdEdge, success bool) {
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
