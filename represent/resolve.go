package represent

import (
	log "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
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

	log.WithFields(log.Fields{
		"system":   "edge-resolution",
		"edgespec": d,
	}).Error("No resolution method for provided EdgeSpec") // TODO is it ok to not halt here?

	return types.StdEdge{}, false
}
