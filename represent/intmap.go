package represent

import (
	"strconv"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/types/system"
)

// intMapV is a persistent map from uint64 keys to VertexTuple values.
type intMapV struct {
	ps.Map
}

// newIntMapV creates a new intMapV for use.
func newIntMapV() intMapV {
	return intMapV{ps.NewMap()}
}

// Set takes a uint64 key and a VertexTuple, sets the VertexTuple as the value for
// that key and returns a new intMap pointer (ish...since this is a fake wrapper atm)
func (m intMapV) Set(key uint64, value system.VertexTuple) intMapV {
	return intMapV{m.Map.Set(strconv.FormatUint(key, 10), value)}
}

// Get takes a uint64 key and returns the VertexTuple at that key, or an empty
// VertexTuple and false.
func (m intMapV) Get(key uint64) (system.VertexTuple, bool) {
	vt, exists := m.Map.Lookup(strconv.FormatUint(key, 10))

	if exists {
		return vt.(system.VertexTuple), true
	} else {
		return system.VertexTuple{}, false
	}
}

// Delete takes a uint64 key and deletes the key from the map. It returns a
// pointer (ish) to a new intMap.
func (m intMapV) Delete(key uint64) intMapV {
	return intMapV{m.Map.Delete(strconv.FormatUint(key, 10))}
}

// intMapE is a persistent map from uint64 keys to system.StdEdge values.
// TODO can't use this until the edge, vt types move back into this package
type intMapE struct {
	ps.Map
}

// newIntMap creates a new intMapE for use.
func newIntMapE() intMapE {
	return intMapE{ps.NewMap()}
}

// Set takes a uint64 key and a StdEdge, sets the StdEdge as the value for
// that key and returns a new intMapE pointer (ish...since this is a fake wrapper atm)
func (m intMapE) Set(key uint64, value system.StdEdge) intMapE {
	return intMapE{m.Map.Set(strconv.FormatUint(key, 10), value)}
}

// Get takes a uint64 key and returns the StdEdge at that key, or an empty
// StdEdge and false.
func (m intMapE) Get(key uint64) (system.StdEdge, bool) {
	vt, exists := m.Map.Lookup(strconv.FormatUint(key, 10))

	if exists {
		return vt.(system.StdEdge), true
	} else {
		return system.StdEdge{}, false
	}
}

// Delete takes a uint64 key and deletes the key from the map. It returns a
// pointer (ish) to a new intMapE.
func (m intMapE) Delete(key uint64) intMapE {
	return intMapE{m.Map.Delete(strconv.FormatUint(key, 10))}
}
