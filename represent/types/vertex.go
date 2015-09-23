package types

import (
	"errors"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
)

// A value indicating a vertex's type. For now, done as a string.
type VType string

// A value indicating an edge's type. For now, done as a string.
type EType string

type Property struct {
	MsgSrc uint64      `json:"msgsrc"`
	Value  interface{} `json:"value"`
}

type PropPair struct {
	K string
	V interface{}
}

type VertexTuple struct {
	ID       int
	Vertex   StdVertex
	InEdges  ps.Map
	OutEdges ps.Map
}

// NewVertex creates a new Vertex object, assigning each PropPair to Props
// using the provided msgid.
func NewVertex(typ VType, msgid uint64, p ...PropPair) (v StdVertex) {
	v.Type, v.Properties = typ, fillMapIgnoreZero(msgid, p...)
	return v
}

func fillMapIgnoreZero(msgid uint64, p ...PropPair) ps.Map {
	m := ps.NewMap()
	var zero bool
	var err error

	for _, pair := range p {
		if zero, err = isZero(pair.V); !zero && err == nil {
			m = m.Set(pair.K, Property{MsgSrc: msgid, Value: pair.V})
		}
	}

	return m
}

type StdVertex struct {
	Type       VType  `json:"type"`
	Properties ps.Map `json:"properties"`
}

type emptyChecker interface {
	IsEmpty() bool
}

func isZero(v interface{}) (bool, error) {
	switch v.(type) {
	case bool:
		return v == false, nil
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
		return v == 0, nil
	case float32, float64:
		return v == 0.0, nil
	case string:
		return v == "", nil
	case [20]byte:
		return v == [20]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, nil
	default:
		if ec, ok := v.(emptyChecker); ok {
			return ec.IsEmpty(), nil
		}
	}
	return false, errors.New("No static zero value defined for provided type")
}

// Returns a string representing the object type. Used for namespacing keys, etc.
// While this is (currently) implemented as a method, its result must be invariant.
// TODO use string-const generator, other tricks to enforce invariance, compact space use
func (v StdVertex) Typ() VType {
	return v.Type
}

// Returns a persistent map with the vertex's properties.
// TODO generate more type-restricted versions of the map?
func (v StdVertex) Props() ps.Map {
	return v.Properties
}

// Merge merges another vertex into this vertex. Error is indicated if the
// dynamic types do not match.
func (v StdVertex) Merge(v2 StdVertex) (StdVertex, error) {
	var old, nu ps.Map = v.Props(), v2.Props()

	nu.ForEach(func(key string, val ps.Any) {
		old.Set(key, val)
	})

	v.Properties = old

	return v, nil
}
