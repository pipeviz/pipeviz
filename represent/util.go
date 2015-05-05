package represent

import "github.com/mndrix/ps"

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
func mapPropPairs(mid int, pairs ...p) (m ps.Map) {
	for k, kv := range pairs {
		pairs[k] = p{K: kv.K, V: Property{MsgSrc: mid, Value: kv.V}}
	}
	return mapPairs(pairs...)
}

type flatVTuple struct {
	Id       int
	V        flatVertex
	InEdges  []flatEdge
	OutEdges []flatEdge
}

type flatVertex struct {
	VType VType
	Props map[string]Property
}

type flatEdge struct {
	Id     int
	Source int
	Target int
	EType  EType
	// TODO do these *actually* need to be persistent structures?
	Props map[string]Property
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
func (vt VertexTuple) flat() (flat flatVTuple) {
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
