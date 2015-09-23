package types

import "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"

type SplitData struct {
	Vertex    Vertex
	EdgeSpecs EdgeSpecs
}

// TODO for now, no structure to this. change to queryish form later
type EdgeSpec interface {
	// Given a graph and a vtTuple root, searches for an existing edge
	// that this EdgeSpec would supercede
	//FindExisting(*CoreGraph, vtTuple) (StandardEdge, bool)

	// Resolves the spec into a real edge, merging as appropriate with
	// any existing edge (as returned from FindExisting)
	//Resolve(*CoreGraph, vtTuple, int) (StandardEdge, bool)
}

type EdgeSpecs []EdgeSpec

type flatVTuple struct {
	Id       int        `json:"id"`
	V        flatVertex `json:"vertex"`
	InEdges  []flatEdge `json:"inEdges"`
	OutEdges []flatEdge `json:"outEdges"`
}

type flatVertex struct {
	VType VType               `json:"type"`
	Props map[string]Property `json:"properties"`
}

type flatEdge struct {
	Id     int                 `json:"id"`
	Source int                 `json:"source"`
	Target int                 `json:"target"`
	EType  EType               `json:"etype"`
	Props  map[string]Property `json:"properties"`
}

func vtoflat(v Vertex) (flat flatVertex) {
	flat.VType = v.Typ()

	flat.Props = make(map[string]Property)
	v.Props().ForEach(func(k string, v ps.Any) {
		flat.Props[k] = v.(Property) // TODO err...ever not property?
	})
	return
}

func etoflat(e StandardEdge) (flat flatEdge) {
	flat = flatEdge{
		Id:     e.ID,
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
	flat.Id = vt.ID
	flat.V = vtoflat(vt.Vertex)

	vt.InEdges.ForEach(func(k string, v ps.Any) {
		e := v.(StandardEdge)
		fe := etoflat(e)
		flat.InEdges = append(flat.InEdges, fe)
	})

	vt.OutEdges.ForEach(func(k string, v ps.Any) {
		e := v.(StandardEdge)
		fe := etoflat(e)
		flat.OutEdges = append(flat.OutEdges, fe)
	})

	return flat
}
