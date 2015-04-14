package represent

import (
	"fmt"

	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
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

func assignEnvLink(mid int, e interpret.EnvLink, m ps.Map, excl bool) ps.Map {
	m = assignAddress(mid, e.Address, m, excl)
	// nick is logically separate from network identity, so excl has no effect
	if e.Nick != "" {
		m = m.Set("nick", Property{MsgSrc: mid, Value: e.Nick})
	}

	return m
}

func assignAddress(mid int, a interpret.Address, m ps.Map, excl bool) ps.Map {
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

// All of the vertex types behave identically for now, but this will change as things mature

type vertexEnvironment struct {
	props ps.Map
}

func (vtx vertexEnvironment) Props() ps.Map {
	return vtx.props
}

func (vtx vertexEnvironment) Typ() VType {
	return "environment"
}

func (vtx vertexEnvironment) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(vertexEnvironment); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type vertexLogicState struct {
	props ps.Map
}

func (vtx vertexLogicState) Props() ps.Map {
	return vtx.props
}

func (vtx vertexLogicState) Typ() VType {
	return "logic-state"
}

func (vtx vertexLogicState) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(vertexLogicState); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type vertexProcess struct {
	props ps.Map
}

func (vtx vertexProcess) Props() ps.Map {
	return vtx.props
}

func (vtx vertexProcess) Typ() VType {
	return "process"
}

func (vtx vertexProcess) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(vertexProcess); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type vertexComm struct {
	props ps.Map
}

func (vtx vertexComm) Props() ps.Map {
	return vtx.props
}

func (vtx vertexComm) Typ() VType {
	return "comm"
}

func (vtx vertexComm) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(vertexComm); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type vertexDataset struct {
	props ps.Map
}

func (vtx vertexDataset) Props() ps.Map {
	return vtx.props
}

func (vtx vertexDataset) Typ() VType {
	return "dataset"
}

func (vtx vertexDataset) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(vertexDataset); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type vertexCommit struct {
	props ps.Map
}

func (vtx vertexCommit) Props() ps.Map {
	return vtx.props
}

func (vtx vertexCommit) Typ() VType {
	return "commit"
}

func (vtx vertexCommit) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(vertexCommit); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type vertexVcsLabel struct {
	props ps.Map
}

func (vtx vertexVcsLabel) Props() ps.Map {
	return vtx.props
}

func (vtx vertexVcsLabel) Typ() VType {
	return "vcs-label"
}

func (vtx vertexVcsLabel) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(vertexVcsLabel); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type vertexTestResult struct {
	props ps.Map
}

func (vtx vertexTestResult) Props() ps.Map {
	return vtx.props
}

func (vtx vertexTestResult) Typ() VType {
	return "test-result"
}

func (vtx vertexTestResult) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(vertexTestResult); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type vertexParentDataset struct {
	props ps.Map
}

func (vtx vertexParentDataset) Props() ps.Map {
	return vtx.props
}

func (vtx vertexParentDataset) Typ() VType {
	return "parent-dataset"
}

func (vtx vertexParentDataset) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(vertexParentDataset); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}
