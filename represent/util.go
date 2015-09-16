package represent

import (
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/represent/types"
)

// Just reuse PropQ
type p PropQ

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
		pairs[k] = p{K: kv.K, V: Property{MsgSrc: mid, Value: kv.V}}
	}
	return mapPairs(pairs...)
}

type flatVTuple struct {
	Id       int        `json:"id"`
	V        flatVertex `json:"vertex"`
	InEdges  []flatEdge `json:"inEdges"`
	OutEdges []flatEdge `json:"outEdges"`
}

type flatVertex struct {
	VType types.VType         `json:"type"`
	Props map[string]Property `json:"properties"`
}

type flatEdge struct {
	Id     int                 `json:"id"`
	Source int                 `json:"source"`
	Target int                 `json:"target"`
	EType  types.EType         `json:"etype"`
	Props  map[string]Property `json:"properties"`
}

func vtoflat(v types.Vtx) (flat flatVertex) {
	flat.VType = v.Typ()

	flat.Props = make(map[string]Property)
	v.Props().ForEach(func(k string, v ps.Any) {
		flat.Props[k] = v.(Property) // TODO err...ever not property?
	})
	return
}

func etoflat(e StandardEdge) (flat flatEdge) {
	flat = flatEdge{
		Id:     e.id,
		Source: e.Source,
		Target: e.Target,
		EType:  e.EType,
		Props:  make(map[string]Property),
	}

	e.Props.ForEach(func(k2 string, v2 ps.Any) {
		flat.Props[k2] = v2.(Property)
	})

	return
}

// flatten the persistent structures in the vtTuple down into conventional ones (typically for easy printing).
// TODO if this is gonna be exported, it's gotta be cleaned up
func (vt VertexTuple) Flat() (flat flatVTuple) {
	flat.Id = vt.id
	flat.V = vtoflat(vt.v)

	vt.ie.ForEach(func(k string, v ps.Any) {
		e := v.(StandardEdge)
		fe := etoflat(e)
		flat.InEdges = append(flat.InEdges, fe)
	})

	vt.oe.ForEach(func(k string, v ps.Any) {
		e := v.(StandardEdge)
		fe := etoflat(e)
		flat.OutEdges = append(flat.OutEdges, fe)
	})

	return flat
}
