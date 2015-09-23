package represent

import (
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/represent/types"
)

func GenericMerge(old, nu ps.Map) ps.Map {
	nu.ForEach(func(key string, val ps.Any) {
		prop := val.(types.Property)
		switch val := prop.Value.(type) {
		case int: // TODO handle all builtin numeric types
			// TODO not clear atm where responsibility sits for ensuring no zero values, but
			// it should probably not be here (at vtx creation instead). drawback of that is
			// that it could allow sloppy plugin code to gunk up the system
			if val != 0 {
				old = old.Set(key, prop)
			}
		case string:
			if val != "" {
				old = old.Set(key, prop)
			}
		case []byte:
			if len(val) != 0 {
				old = old.Set(key, prop)
			}
		}

	})

	return old
}

func assignEnvLink(mid uint64, e interpret.EnvLink, m ps.Map, excl bool) ps.Map {
	m = assignAddress(mid, e.Address, m, excl)
	// nick is logically separate from network identity, so excl has no effect
	if e.Nick != "" {
		m = m.Set("nick", types.Property{MsgSrc: mid, Value: e.Nick})
	}

	return m
}
