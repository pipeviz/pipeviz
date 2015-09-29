package types

import "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"

type (
	// A value indicating a vertex's type. For now, done as a string.
	VType string
	// A value indicating an edge's type. For now, done as a string.
	EType string
)

const (
	// Constant representing the zero value for VTypes. Reserved; no actual vertex can be this.
	VTypeNone VType = ""
	// Constant representing the zero value for ETypes. Reserved; no actual edge can be this.
	ETypeNone EType = ""
)

// VertexTuple is the base storage object used by the graph engine.
type VertexTuple struct {
	ID       int
	Vertex   StdVertex
	InEdges  ps.Map
	OutEdges ps.Map
}
