package represent

import (
	"fmt"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/represent/types"
)

func GenericMerge(old, nu ps.Map) ps.Map {
	nu.ForEach(func(key string, val ps.Any) {
		prop := val.(Property)
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
		m = m.Set("nick", Property{MsgSrc: mid, Value: e.Nick})
	}

	return m
}

func assignAddress(mid uint64, a interpret.Address, m ps.Map, excl bool) ps.Map {
	if a.Hostname != "" {
		if excl {
			m = m.Delete("ipv4")
			m = m.Delete("ipv6")
		}
		m = m.Set("hostname", Property{MsgSrc: mid, Value: a.Hostname})
	}
	if a.Ipv4 != "" {
		if excl {
			m = m.Delete("hostname")
			m = m.Delete("ipv6")
		}
		m = m.Set("ipv4", Property{MsgSrc: mid, Value: a.Ipv4})
	}
	if a.Ipv6 != "" {
		if excl {
			m = m.Delete("hostname")
			m = m.Delete("ipv4")
		}
		m = m.Set("ipv6", Property{MsgSrc: mid, Value: a.Ipv6})
	}

	return m
}

type vertexYumPkg struct {
	props ps.Map
}

func (vtx vertexYumPkg) Props() ps.Map {
	return vtx.props
}

func (vtx vertexYumPkg) Typ() types.VType {
	return "pkg-yum"
}

func (vtx vertexYumPkg) Merge(ivtx types.Vtx) (types.Vtx, error) {
	if _, ok := ivtx.(vertexYumPkg); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}
