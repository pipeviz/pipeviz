package types

// EFilter describes an Edge Filter, used by the traversal/query
// system to govern traversals ad limit results.
type EFilter interface {
	EType() EType
	EProps() []PropPair
}

// VFilter describes an Vertex Filter, used by the traversal/query
// system to govern traversals ad limit results.
type VFilter interface {
	VType() VType
	VProps() []PropPair
}

// VEFilter is both a VFilter and an EFilter.
type VEFilter interface {
	EFilter
	VFilter
}
