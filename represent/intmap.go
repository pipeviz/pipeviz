package represent

import (
	"strconv"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/types/system"
)

// intMap is a persistent map from uint64 keys to VertexTuple values.
type intMap struct {
	ps.Map
}

// newIntMap creates a new intMap for use.
func newIntMap() intMap {
	return intMap{ps.NewMap()}
}

// Set takes a uint64 key and a VertexTuple, sets the VertexTuple as the value for
// that key and returns a new intMap pointer (ish...since this is a fake wrapper atm)
func (m intMap) Set(key uint64, value system.VertexTuple) intMap {
	return intMap{m.Map.Set(strconv.FormatUint(key, 10), value)}
}

// Get takes a uint64 key and returns the VertexTuple at that key, or an empty
// VertexTuple and false.
func (m intMap) Get(key uint64) (system.VertexTuple, bool) {
	vt, exists := m.Map.Lookup(strconv.FormatUint(key, 10))

	if exists {
		return vt.(system.VertexTuple), true
	} else {
		return system.VertexTuple{}, false
	}
}

// Delete takes a uint64 key and deletes the key from the map. It returns a
// pointer (ish) to a new intMap.
func (m intMap) Delete(key uint64, value ps.Any) intMap {
	return intMap{m.Map.Delete(strconv.FormatUint(key, 10))}
}
