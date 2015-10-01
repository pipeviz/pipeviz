package represent

import (
	"bytes"

	log "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/represent/types"
)

// Inspects the indicated vertex's set of out-edges, returning a slice of
// those that match on type and properties. ETypeNone and nil can be passed
// for the last two parameters respectively, in which case the filters will
// be bypassed.
func (g *coreGraph) OutWith(egoId int, ef types.EFilter) (es types.EdgeVector) {
	return g.arcWith(egoId, ef, false)
}

// Inspects the indicated vertex's set of in-edges, returning a slice of
// those that match on type and properties. ETypeNone and nil can be passed
// for the last two parameters respectively, in which case the filters will
// be bypassed.
func (g *coreGraph) InWith(egoId int, ef types.EFilter) (es types.EdgeVector) {
	return g.arcWith(egoId, ef, true)
}

func (g *coreGraph) arcWith(egoId int, ef types.EFilter, in bool) (es types.EdgeVector) {
	etype, props := ef.EType(), ef.EProps()
	vt, err := g.Get(egoId)
	if err != nil {
		// vertex doesn't exist
		return
	}

	var fef func(k string, v ps.Any)
	// TODO specialize the func for zero-cases
	fef = func(k string, v ps.Any) {
		edge := v.(types.StdEdge)
		if etype != types.ETypeNone && etype != edge.EType {
			// etype doesn't match
			return
		}

		for _, p := range props {
			eprop, exists := edge.Props.Lookup(p.K)
			if !exists {
				return
			}

			deprop := eprop.(types.Property)
			switch tv := deprop.Value.(type) {
			default:
				if tv != p.V {
					return
				}
			case []byte:
				cmptv, ok := p.V.([]byte)
				if !ok || !bytes.Equal(tv, cmptv) {
					return
				}
			}
		}

		es = append(es, edge)
	}

	if in {
		vt.InEdges.ForEach(fef)
	} else {
		vt.OutEdges.ForEach(fef)
	}

	return
}

// Return a slice of vtTuples that are successors of the given vid, constraining the list
// to those that are connected by edges that pass the EdgeFilter, and the successor
// vertices pass the VertexFilter.
func (g *coreGraph) SuccessorsWith(egoId int, vef types.VEFilter) (vts types.VertexTupleVector) {
	return g.adjacentWith(egoId, vef, false)
}

// Return a slice of vtTuples that are predecessors of the given vid, constraining the list
// to those that are connected by edges that pass the EdgeFilter, and the predecessor
// vertices pass the VertexFilter.
func (g *coreGraph) PredecessorsWith(egoId int, vef types.VEFilter) (vts types.VertexTupleVector) {
	return g.adjacentWith(egoId, vef, true)
}

func (g *coreGraph) adjacentWith(egoId int, vef types.VEFilter, in bool) (vts types.VertexTupleVector) {
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
		edge := v.(types.StdEdge)
		if etype != types.ETypeNone && etype != edge.EType {
			// etype doesn't match
			return
		}

		for _, p := range eprops {
			eprop, exists := edge.Props.Lookup(p.K)
			if !exists {
				return
			}

			deprop := eprop.(types.Property)
			switch tv := deprop.Value.(type) {
			default:
				if tv != p.V {
					return
				}
			case []byte:
				cmptv, ok := p.V.([]byte)
				if !ok || !bytes.Equal(tv, cmptv) {
					return
				}
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
			vt.InEdges.ForEach(feef)
		} else {
			vt.OutEdges.ForEach(feef)
		}
		close(vidchan)
	}()

	// Keep track of the vertices we've collected, for deduping
	visited := make(map[int]struct{})
VertexInspector:
	for vid := range vidchan {
		if _, seenit := visited[vid]; seenit {
			// already visited this vertex; go back to waiting on the chan
			continue
		}
		// mark this vertex as black
		visited[vid] = struct{}{}

		adjvt, err := g.Get(vid)
		if err != nil {
			log.WithFields(log.Fields{
				"system": "engine",
				"err":    err,
			}).Error("Attempted to get nonexistent vid during traversal - should be impossible")
			continue
		}

		// FIXME can't rely on Typ() method here, need to store it
		if vtype != types.VTypeNone && vtype != adjvt.Vertex.Typ() {
			continue
		}

		for _, p := range vprops {
			vprop, exists := adjvt.Vertex.Props().Lookup(p.K)
			if !exists {
				continue VertexInspector
			}

			dvprop := vprop.(types.Property)
			switch tv := dvprop.Value.(type) {
			default:
				if tv != p.V {
					continue VertexInspector
				}
			case []byte:
				cmptv, ok := p.V.([]byte)
				if !ok || !bytes.Equal(tv, cmptv) {
					continue VertexInspector
				}
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
func (g *coreGraph) VerticesWith(vf types.VFilter) (vs types.VertexTupleVector) {
	vtype, props := vf.VType(), vf.VProps()
	g.vtuples.ForEach(func(_ string, val ps.Any) {
		vt := val.(types.VertexTuple)
		if vtype != types.VTypeNone && vt.Vertex.Typ() != vtype {
			return
		}

		for _, p := range props {
			vprop, exists := vt.Vertex.Props().Lookup(p.K)
			if !exists {
				return
			}

			dvprop := vprop.(types.Property)
			switch tv := dvprop.Value.(type) {
			default:
				if tv != p.V {
					return
				}
			case []byte:
				cmptv, ok := p.V.([]byte)
				if !ok || !bytes.Equal(tv, cmptv) {
					return
				}
			}
		}

		vs = append(vs, vt)
	})

	return vs
}
