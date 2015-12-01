package system

import "github.com/mndrix/ps"

type flatVTuple struct {
	ID       uint64     `json:"id"`
	V        flatVertex `json:"vertex"`
	InEdges  []flatEdge `json:"inEdges"`
	OutEdges []flatEdge `json:"outEdges"`
}

type flatVertex struct {
	VType VType               `json:"type"`
	Props map[string]Property `json:"properties"`
}

type flatEdge struct {
	ID     uint64              `json:"id"`
	Source uint64              `json:"source"`
	Target uint64              `json:"target"`
	EType  EType               `json:"etype"`
	Props  map[string]Property `json:"properties"`
}

func vtoflat(v StdVertex) (flat flatVertex) {
	flat.VType = v.Typ()

	flat.Props = make(map[string]Property)
	v.Props().ForEach(func(k string, v ps.Any) {
		flat.Props[k] = v.(Property) // TODO err...ever not property?
	})
	return
}

func etoflat(e StdEdge) (flat flatEdge) {
	flat = flatEdge{
		ID:     e.ID,
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
	flat.ID = vt.ID
	flat.V = vtoflat(vt.Vertex)

	vt.InEdges.ForEach(func(k string, v ps.Any) {
		e := v.(StdEdge)
		fe := etoflat(e)
		flat.InEdges = append(flat.InEdges, fe)
	})

	vt.OutEdges.ForEach(func(k string, v ps.Any) {
		e := v.(StdEdge)
		fe := etoflat(e)
		flat.OutEdges = append(flat.OutEdges, fe)
	})

	return flat
}
