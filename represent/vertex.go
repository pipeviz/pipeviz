package represent

import (
	"fmt"

	"github.com/mndrix/ps"
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

// All of the vertex types behave identically for now, but this will change as things mature

type environmentVertex struct {
	props ps.Map
}

func (vtx environmentVertex) Props() ps.Map {
	return vtx.props
}

func (vtx environmentVertex) Typ() VType {
	return "environment"
}

func (vtx environmentVertex) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(environmentVertex); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type logicStateVertex struct {
	props ps.Map
}

func (vtx logicStateVertex) Props() ps.Map {
	return vtx.props
}

func (vtx logicStateVertex) Typ() VType {
	return "logicState"
}

func (vtx logicStateVertex) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(logicStateVertex); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type processVertex struct {
	props ps.Map
}

func (vtx processVertex) Props() ps.Map {
	return vtx.props
}

func (vtx processVertex) Typ() VType {
	return "process"
}

func (vtx processVertex) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(processVertex); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type commVertex struct {
	props ps.Map
}

func (vtx commVertex) Props() ps.Map {
	return vtx.props
}

func (vtx commVertex) Typ() VType {
	return "comm"
}

func (vtx commVertex) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(commVertex); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type datasetVertex struct {
	props ps.Map
}

func (vtx datasetVertex) Props() ps.Map {
	return vtx.props
}

func (vtx datasetVertex) Typ() VType {
	return "dataset"
}

func (vtx datasetVertex) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(datasetVertex); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type commitVertex struct {
	props ps.Map
}

func (vtx commitVertex) Props() ps.Map {
	return vtx.props
}

func (vtx commitVertex) Typ() VType {
	return "commit"
}

func (vtx commitVertex) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(commitVertex); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}

type commitMetaVertex struct {
	props ps.Map
}

func (vtx commitMetaVertex) Props() ps.Map {
	return vtx.props
}

func (vtx commitMetaVertex) Typ() VType {
	return "commitMeta"
}

func (vtx commitMetaVertex) Merge(ivtx Vertex) (Vertex, error) {
	if _, ok := ivtx.(commitMetaVertex); !ok {
		// NOTE remember, formatting with types means reflection
		return nil, fmt.Errorf("Attempted to merge vertex type %T into vertex type %T", ivtx, vtx)
	}

	vtx.props = GenericMerge(vtx.props, ivtx.Props())
	return vtx, nil
}
