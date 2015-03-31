package interpret

import (
	"github.com/sdboyer/gogl"
)

type EdgeList map[int]interface{}

type VertexContainer struct {
	Vertex gogl.Vertex
	Edges  EdgeList
}

// this implementation is not threadsafe, but that's fine-ish for now
// because our design guarantees no parallel writes on the graph. parallel
// reads will be dealt with later as part of making the graph persistent
type mGraph struct {
	list       map[int]VertexContainer
	size       int
	vtxCounter int
}

/* mGraph shared methods */

// Returns the vertex associated with the given id, if one can be found.
//
// If no vertex is associated with the given id, returns nil, false.
func (g *mGraph) GetVertex(id int) (vertex gogl.Vertex, exists bool) {
	var vc VertexContainer
	vc, exists = g.list[id]

	if exists {
		return vc.Vertex, true
	}

	return nil, exists
}

// Traverses the graph's vertices in random order, passing each vertex to the
// provided closure.
func (g *mGraph) Vertices(f gogl.VertexStep) {
	for _, v := range g.list {
		if f(v.Vertex) {
			return
		}
	}
}

// Indicates whether or not the given vertex is present in the graph.
func (g *mGraph) HasVertex(vertex gogl.Vertex) (exists bool) {
	_, exists = g.Find(vertex)
	return
}

// Searches for an instance of the vertex within the graph. If found,
// returns the vertex id and true; otherwise returns 0 and false.
func (g *mGraph) Find(vertex gogl.Vertex) (int, bool) {
	// FIXME so very hilariously O(n)

	var chk Identifier
	for _, idf := range Identifiers {
		if idf.CanIdentify(vertex) {
			chk = idf
		}
	}

	// we hit this case iff there's an object type our identifiers can't work
	// with. which should, eventually, be structurally impossible by this point
	if chk == nil {
		return 0, false
	}

	for id, vc := range g.list {
		if chk.Matches(vc.Vertex, vertex) {
			return id, true
		}
	}

	return 0, false
}

// Returns the order (number of vertices) in the graph.
func (g *mGraph) Order() int {
	return len(g.list)
}

// Returns the size (number of edges) in the graph.
func (g *mGraph) Size() int {
	return g.size
}

// Adds the provided vertices to the graph. If a provided vertex is
// already present in the graph, it is a no-op (for that vertex only).
func (g *mGraph) EnsureVertex(vertices ...gogl.Vertex) {
	if len(vertices) == 0 {
		return
	}

	for _, vertex := range vertices {
		if !g.HasVertex(vertex) {
			g.vtxCounter++
			g.list[g.vtxCounter] = VertexContainer{Vertex: vertex, Edges: make(EdgeList)}
		}
	}

	return
}

// Enumerates the vertices adjacent to the provided vertex.
func (g *mGraph) AdjacentTo(start gogl.Vertex, f gogl.VertexStep) {
	g.IncidentTo(start, func(e gogl.Edge) bool {
		u, v := e.Both()
		if u == start {
			return f(v)
		} else {
			return f(u)
		}
	})
}

// Removes a vertex from the graph. Also removes any edges of which that
// vertex is a member.
func (g *mGraph) RemoveVertex(vertices ...gogl.Vertex) {
	if len(vertices) == 0 {
		return
	}

	for _, vertex := range vertices {
		if g.hasVertex(vertex) {
			g.size -= len(g.list[vertex])
			delete(g.list, vertex)

			for _, adjacent := range g.list {
				if _, has := adjacent[vertex]; has {
					delete(adjacent, vertex)
					g.size--
				}
			}
		}
	}
	return
}

// Adds a new arc to the graph.
func (g *mGraph) AddArcs(arcs ...MessageArc) {
	for _, arc := range arcs {
		u, v := arc.Both()
		g.ensureVertex(u, v)

		if _, exists := g.list[u][v]; !exists {
			g.list[u][v] = arc.Data()
			g.size++
		}
	}
}

// Removes arcs from the graph. This does NOT remove vertex members of the
// removed arcs.
func (g *mGraph) RemoveArcs(arcs ...gogl.DataArc) {
	if len(arcs) == 0 {
		return
	}

	for _, arc := range arcs {
		s, t := arc.Both()
		if _, exists := g.list[s][t]; exists {
			delete(g.list[s], t)
			g.size--
		}
	}
}

func (g *mGraph) Transpose() gogl.Digraph {
	g2 := &mGraph{}
	g2.list = make(map[Vertex]map[Vertex]interface{})

	// Guess at average indegree by looking at ratio of edges to vertices, use that to initially size the adjacency maps
	startcap := int(g.Size() / g.Order())

	for source, adjacent := range g.list {
		if !g2.hasVertex(source) {
			g2.list[source] = make(map[Vertex]interface{}, startcap+1)
		}

		for target, data := range adjacent {
			if !g2.hasVertex(target) {
				g2.list[target] = make(map[Vertex]interface{}, startcap+1)
			}
			g2.list[target][source] = data
		}
	}

	return g2
}
