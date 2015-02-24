package interpret

import (
	"github.com/sdboyer/gogl"
)

type EdgeList map[int64]interface{}

type VertexContainer struct {
	Vertex gogl.Vertex
	Id     int64
	Edges  EdgeList
}

// this implementation is not threadsafe, but that's fine-ish for now
// because our design guarantees no parallel writes on the graph. parallel
// reads will be dealt with later as part of making the graph persistent
type mGraph struct {
	list       []VertexContainer
	size       int
	vtxCounter int64
}

/* mGraph shared methods */

// Traverses the graph's vertices in random order, passing each vertex to the
// provided closure.
func (g *mGraph) Vertices(f gogl.VertexStep) {
	for v := range g.list {
		if f(v) {
			return
		}
	}
}

// Indicates whether or not the given vertex is present in the graph.
func (g *mGraph) HasVertex(vertex gogl.Vertex) (exists bool) {
	//_, exists = g.list[vertex]
	return
}

// Searches for an instance of the vertex within the graph. If found,
// returns the vertex id and true; otherwise returns 0 and false.
//
// For now this is terrible and hardcodes all the business logic (type-switchy)
// inside the graph. Need to massively refactor later.
func (g *mGraph) Find(vertex gogl.Vertex) (int64, bool) {
	// FIXME big switch, then dispatch off to unexported funcs...?
	for _, vc := range g.list {
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
			g.list = append(g.list, VertexContainer{Vertex: vertex, Id: g.vtxCounter, Edges: make(EdgeList)})
		}
	}

	return
}

// Returns the outdegree of the provided vertex. If the vertex is not present in the
// graph, the second return value will be false.
func (g *mGraph) OutDegreeOf(vertex gogl.Vertex) (degree int, exists bool) {
	if exists = g.hasVertex(vertex); exists {
		degree = len(g.list[vertex])
	}
	return
}

// Returns the indegree of the provided vertex. If the vertex is not present in the
// graph, the second return value will be false.
//
// Note that getting indegree is inefficient for directed adjacency lists; it requires
// a full scan of the graph's edge set.
func (g *mGraph) InDegreeOf(vertex gogl.Vertex) (degree int, exists bool) {
	return inDegreeOf(g, vertex)
}

// Returns the degree of the provided vertex, counting both in and out-edges.
func (g *mGraph) DegreeOf(vertex gogl.Vertex) (degree int, exists bool) {
	indegree, exists := inDegreeOf(g, vertex)
	outdegree, exists := g.OutDegreeOf(vertex)
	return indegree + outdegree, exists
}

// Enumerates the set of all edges incident to the provided vertex.
func (g *mGraph) IncidentTo(v gogl.Vertex, f gogl.EdgeStep) {
	eachEdgeIncidentToDirected(g, v, f)
}

// Enumerates the vertices adjacent to the provided vertex.
func (g *mGraph) AdjacentTo(start gogl.Vertex, f gogl.VertexStep) {
	g.IncidentTo(start, func(e Edge) bool {
		u, v := e.Both()
		if u == start {
			return f(v)
		} else {
			return f(u)
		}
	})
}

// Enumerates the set of out-edges for the provided vertex.
func (g *mGraph) ArcsFrom(v gogl.Vertex, f gogl.ArcStep) {
	if !g.hasVertex(v) {
		return
	}

	for adjacent, data := range g.list[v] {
		if f(NewDataArc(v, adjacent, data)) {
			return
		}
	}
}

func (g *mGraph) SuccessorsOf(v gogl.Vertex, f gogl.VertexStep) {
	eachVertexInAdjacencyList(g.list, v, f)
}

// Enumerates the set of in-edges for the provided vertex.
func (g *mGraph) ArcsTo(v gogl.Vertex, f gogl.ArcStep) {
	if !g.hasVertex(v) {
		return
	}

	for candidate, adjacent := range g.list {
		for target, data := range adjacent {
			if target == v {
				if f(NewDataArc(candidate, target, data)) {
					return
				}
			}
		}
	}
}

func (g *mGraph) PredecessorsOf(v gogl.Vertex, f gogl.VertexStep) {
	eachPredecessorOf(g.list, v, f)
}

// Traverses the set of edges in the graph, passing each edge to the
// provided closure.
func (g *mGraph) Edges(f gogl.EdgeStep) {
	for source, adjacent := range g.list {
		for target, data := range adjacent {
			if f(NewDataEdge(source, target, data)) {
				return
			}
		}
	}
}

// Traverses the set of arcs in the graph, passing each arc to the
// provided closure.
func (g *mGraph) Arcs(f gogl.ArcStep) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for source, adjacent := range g.list {
		for target, data := range adjacent {
			if f(NewDataArc(source, target, data)) {
				return
			}
		}
	}
}

// Indicates whether or not the given edge is present in the graph. It matches
// based solely on the presence of an edge, disregarding edge property.
func (g *mGraph) HasEdge(edge Edge) bool {
	u, v := edge.Both()
	_, exists := g.list[u][v]
	if !exists {
		_, exists = g.list[v][u]
	}
	return exists
}

// Indicates whether or not the given arc is present in the graph.
func (g *mGraph) HasArc(arc gogl.Arc) bool {
	_, exists := g.list[arc.Source()][arc.Target()]
	return exists
}

// Indicates whether or not the given property edge is present in the graph.
// It will only match if the provided DataEdge has the same property as
// the edge contained in the graph.
func (g *mGraph) HasDataEdge(edge gogl.DataEdge) bool {
	u, v := edge.Both()
	if data, exists := g.list[u][v]; exists {
		return data == edge.Data()
	} else if data, exists = g.list[v][u]; exists {
		return data == edge.Data()
	}
	return false
}

// Indicates whether or not the given data arc is present in the graph.
// It will only match if the provided DataEdge has the same data as
// the edge contained in the graph.
func (g *mGraph) HasDataArc(arc gogl.DataArc) bool {
	if data, exists := g.list[arc.Source()][arc.Target()]; exists {
		return data == arc.Data()
	}
	return false
}

// Returns the density of the graph. Density is the ratio of edge count to the
// number of edges there would be in complete graph (maximum edge count).
func (g *mGraph) Density() float64 {
	order := g.Order()
	return float64(g.Size()) / float64(order*(order-1))
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

// Adds arcs to the graph.
func (g *mGraph) AddArcs(arcs ...gogl.DataArc) {
	if len(arcs) == 0 {
		return
	}

	g.addArcs(arcs...)
}

// Adds a new arc to the graph.
func (g *mGraph) addArcs(arcs ...gogl.DataArc) {
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
