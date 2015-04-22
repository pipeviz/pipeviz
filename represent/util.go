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

// flatten the persistent structures in the vtTuple down into conventional ones (typically for easy printing).
func (vt vtTuple) flat() (flat flatVTuple) {
	flat.Id = vt.id
	flat.V.VType = vt.v.Typ()

	flat.V.Props = make(map[string]Property)
	vt.v.Props().ForEach(func(k string, v ps.Any) {
		flat.V.Props[k] = v.(Property) // TODO err...ever not property?
	})

	vt.ie.ForEach(func(k string, v ps.Any) {
		e := v.(StandardEdge)
		fe := flatEdge{
			Id:     e.id,
			Source: e.Source,
			Target: e.Target,
			EType:  e.EType,
			Props:  make(map[string]Property),
		}

		e.Props.ForEach(func(k2 string, v2 ps.Any) {
			fe.Props[k2] = v2.(Property)
		})

		flat.InEdges = append(flat.InEdges, fe)
	})

	vt.oe.ForEach(func(k string, v ps.Any) {
		e := v.(StandardEdge)
		fe := flatEdge{
			Id:     e.id,
			Source: e.Source,
			Target: e.Target,
			EType:  e.EType,
			Props:  make(map[string]Property),
		}

		e.Props.ForEach(func(k2 string, v2 ps.Any) {
			fe.Props[k2] = v2.(Property)
		})

		flat.OutEdges = append(flat.OutEdges, fe)
	})

	return flat
}
