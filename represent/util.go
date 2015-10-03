package represent

import (
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/types/system"
)

// Just reuse types.PropPair
type p system.PropPair

// Transform a slice of kv pairs into a map.
func mapPairs(pairs ...p) (m ps.Map) {
	m = ps.NewMap()
	for _, kv := range pairs {
		m = m.Set(kv.K, kv.V)
	}

	return
}

// Transform a slice of kv pairs into a map, embedding the value into a Property{}
// with the provided mid as Property.MsgSrc.
func mapPropPairs(mid uint64, pairs ...p) (m ps.Map) {
	for k, kv := range pairs {
		pairs[k] = p{K: kv.K, V: system.Property{MsgSrc: mid, Value: kv.V}}
	}
	return mapPairs(pairs...)
}
