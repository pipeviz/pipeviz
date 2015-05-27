package represent

import "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"

type StandardEdge struct {
	id     int
	Source int
	Target int
	EType  EType
	// TODO do these *actually* need to be persistent structures?
	Props ps.Map
}

// Edge partials contain the shorthand structural information common to all edges,
// but do not carry any extensible/property information. They are stored within the
// vtTuple to allow basic traversals and checks without having to fully load the
// separate edge object.
type edgePartial struct {
	Id   int    // edge id
	Ovid int    // the "other" vertex id. from this scope, unknown if head or tail
	Typ  uint32 // edge type
}
