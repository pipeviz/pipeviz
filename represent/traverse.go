package represent

import "github.com/mndrix/ps"

type EFilter interface {
	EType() EType
	EProps() []PropQ
}

type VFilter interface {
	VType() VType
	VProps() []PropQ
}

type VEFilter interface {
	EFilter
	VFilter
}

type edgeFilter struct {
	etype EType
	props []PropQ
}

type vertexFilter struct {
	vtype VType
	props []PropQ
}

type bothFilter struct {
	VFilter
	EFilter
}

// all temporary functions to just make query building a little easier for now
func (vf vertexFilter) VType() VType {
	return vf.vtype
}

func (vf vertexFilter) VProps() []PropQ {
	return vf.props
}

func (vf vertexFilter) EType() EType {
	return ETypeNone
}

func (vf vertexFilter) EProps() []PropQ {
	return nil
}

func (vf vertexFilter) and(ef EFilter) VEFilter {
	return bothFilter{vf, ef}
}

func (ef edgeFilter) VType() VType {
	return VTypeNone
}

func (ef edgeFilter) VProps() []PropQ {
	return nil
}

func (ef edgeFilter) EType() EType {
	return ef.etype
}

func (ef edgeFilter) EProps() []PropQ {
	return ef.props
}

func (ef edgeFilter) and(vf VFilter) VEFilter {
	return bothFilter{vf, ef}
}

// first string is vtype, then pairs after that are props
func qbv(v ...interface{}) vertexFilter {
	switch len(v) {
	case 0:
		return vertexFilter{VTypeNone, nil}
	case 1, 2:
		return vertexFilter{v[0].(VType), nil}
	default:
		vf := vertexFilter{v[0].(VType), nil}
		v = v[1:]

		var k string
		var v2 interface{}
		for len(v) > 1 {
			k, v2, v = v[0].(string), v[1], v[2:]
			vf.props = append(vf.props, PropQ{k, v2})
		}

		return vf
	}
}

// first string is etype, then pairs after that are props
func qbe(v ...interface{}) edgeFilter {
	switch len(v) {
	case 0:
		return edgeFilter{ETypeNone, nil}
	case 1, 2:
		return edgeFilter{v[0].(EType), nil}
	default:
		ef := edgeFilter{v[0].(EType), nil}
		v = v[1:]

		var k string
		var v2 interface{}
		for len(v) > 1 {
			k, v2, v = v[0].(string), v[1], v[2:]
			ef.props = append(ef.props, PropQ{k, v2})
		}

		return ef
	}
}

// Inspects the indicated vertex's set of out-edges, returning a slice of
// those that match on type and properties. ETypeNone and nil can be passed
// for the last two parameters respectively, in which case the filters will
// be bypassed.
func (g *CoreGraph) OutWith(egoId int, ef EFilter) (es []StandardEdge) {
	return g.arcWith(egoId, ef, false)
}

// Inspects the indicated vertex's set of in-edges, returning a slice of
// those that match on type and properties. ETypeNone and nil can be passed
// for the last two parameters respectively, in which case the filters will
// be bypassed.
func (g *CoreGraph) InWith(egoId int, ef EFilter) (es []StandardEdge) {
	return g.arcWith(egoId, ef, true)
}

func (g *CoreGraph) arcWith(egoId int, ef EFilter, in bool) (es []StandardEdge) {
	etype, props := ef.EType(), ef.EProps()
	vt, err := g.Get(egoId)
	if err != nil {
		// vertex doesn't exist
		return
	}

	var fef func(k string, v ps.Any)
	// TODO specialize the func for zero-cases
	fef = func(k string, v ps.Any) {
		edge := v.(StandardEdge)
		if etype != ETypeNone && etype != edge.EType {
			// etype doesn't match
			return
		}

		for _, p := range props {
			prop, exists := edge.Props.Lookup(p.K)
			if !exists || prop != p.V {
				return
			}
		}

		es = append(es, edge)
	}

	if in {
		vt.ie.ForEach(fef)
	} else {
		vt.oe.ForEach(fef)
	}

	return
}

// Return a slice of vtTuples that are successors of the given vid, constraining the list
// to those that are connected by edges that pass the EdgeFilter, and the successor
// vertices pass the VertexFilter.
func (g *CoreGraph) SuccessorsWith(egoId int, vef VEFilter) (vts []vtTuple) {
	return g.adjacentWith(egoId, vef, false)
}

// Return a slice of vtTuples that are predecessors of the given vid, constraining the list
// to those that are connected by edges that pass the EdgeFilter, and the predecessor
// vertices pass the VertexFilter.
func (g *CoreGraph) PredecessorsWith(egoId int, vef VEFilter) (vts []vtTuple) {
	return g.adjacentWith(egoId, vef, true)
}

func (g *CoreGraph) adjacentWith(egoId int, vef VEFilter, in bool) (vts []vtTuple) {
	etype, eprops := vef.EType(), vef.EProps()
	vtype, vprops := vef.VType(), vef.VProps()
	vt, err := g.Get(egoId)
	if err != nil {
		// vertex doesn't exist
		return
	}

	// Allow some quick parallelism by continuing to search edges while loading vertices
	// TODO performance test this to see if it's at all worth it
	vidchan := make(chan int, 10)

	var feef func(k string, v ps.Any)
	// TODO specialize the func for zero-cases
	feef = func(k string, v ps.Any) {
		edge := v.(StandardEdge)
		if etype != ETypeNone && etype != edge.EType {
			// etype doesn't match
			return
		}

		for _, p := range eprops {
			prop, exists := edge.Props.Lookup(p.K)
			if !exists || prop != p.V {
				return
			}
		}

		if in {
			vidchan <- edge.Source
		} else {
			vidchan <- edge.Target
		}
	}

	go func() {
		if in {
			vt.ie.ForEach(feef)
		} else {
			vt.oe.ForEach(feef)
		}
		close(vidchan)
	}()

	for vid := range vidchan {
		adjvt, err := g.Get(vid)
		if err != nil {
			// TODO panic, really?
			panic("got nonexistent vid - should be impossible")
		}

		// FIXME can't rely on Typ() method here, need to store it
		if vtype != VTypeNone && vtype != adjvt.v.Typ() {
			continue
		}

		for _, p := range vprops {
			prop, exists := adjvt.v.Props().Lookup(p.K)
			if !exists || prop != p.V {
				continue
			}
		}

		vts = append(vts, adjvt)
	}

	return
}

// Returns a slice of vertices matching the filter conditions.
//
// The first parameter, VType, filters on vertex type; passing VTypeNone
// will bypass the filter.
//
// The second parameter allows filtering on a k/v property pair.
func (g *CoreGraph) VerticesWith(vf VFilter) (vs []vtTuple) {
	vtype, props := vf.VType(), vf.VProps()
	g.vtuples.ForEach(func(_ string, val ps.Any) {
		vt := val.(vtTuple)
		if vt.v.Typ() != vtype {
			return
		}

		for _, p := range props {
			prop, exists := vt.v.Props().Lookup(p.K)
			if !exists || prop != p.V {
				return
			}
		}

		vs = append(vs, vt)
	})

	return vs
}
