package represent

import "github.com/sdboyer/pipeviz/interpret"

// Attempts to resolve an EdgeSpec into a real edge. This process has two steps:
//
// 1. Finding the target node.
// 2. Seeing if this edge already exists between source and target.
//
// It is the responsibility of the edge spec's type handler to determine what "if an edge
// already exists" means, as well as whether to overwrite/merge or duplicate the edge in such a case.
func Resolve(g *CoreGraph, d EdgeSpec) (StandardEdge, bool) {
	switch es := d.(type) {
	case interpret.EnvLink:
		return resolveEnvLink(g, es)
	case interpret.DataLink:
		return resolveDataLink(g, es)
	case SpecCommit:
		return resolveSpecCommit(g, es)
	case SpecLocalLogic:
		return resolveSpecLocalLogic(g, es)
	}

	return StandardEdge{}, false
}

func resolveEnvLink(g *CoreGraph, e interpret.EnvLink) (StandardEdge, bool) {
	g.Vertices(func(vtx Vertex, id int) bool {
		// TODO add something to vertex properties to make it easier to check prop membership
	})
}

func resolveDataLink(g *CoreGraph, e interpret.DataLink) (StandardEdge, bool) {
	g.Vertices(func(vtx Vertex, id int) bool {})
}

func resolveSpecCommit(g *CoreGraph, e SpecCommit) (StandardEdge, bool) {
	g.Vertices(func(vtx Vertex, id int) bool {})
}

func resolveSpecLocalLogic(g *CoreGraph, e SpecLocalLogic) (StandardEdge, bool) {
	g.Vertices(func(vtx Vertex, id int) bool {})
}
