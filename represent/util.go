package represent

import (
	"github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/types/system"
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
