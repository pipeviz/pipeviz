package represent

import (
	"bytes"

	log "github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/types/system"
)

// OutWith inspects the indicated vertex's set of out-edges, returning a slice of
// those that match on type and properties. ETypeNone and nil can be passed
// for the last two parameters respectively, in which case the filters will
// be bypassed.
func (g *coreGraph) OutWith(egoID uint64, ef system.EFilter) (es system.EdgeVector) {
	return g.arcWith(egoID, ef, false)
}

// InWith inspects the indicated vertex's set of in-edges, returning a slice of
// those that match on type and properties. ETypeNone and nil can be passed
// for the last two parameters respectively, in which case the filters will
// be bypassed.
func (g *coreGraph) InWith(egoID uint64, ef system.EFilter) (es system.EdgeVector) {
	return g.arcWith(egoID, ef, true)
}

func (g *coreGraph) arcWith(egoID uint64, ef system.EFilter, in bool) (es system.EdgeVector) {
	etype, props := ef.EType(), ef.EProps()
	vt, err := g.Get(egoID)
	if err != nil {
		// vertex doesn't exist
		return
	}

	var fef func(k string, v ps.Any)
	// TODO specialize the func for zero-cases
	fef = func(k string, v ps.Any) {
		edge := v.(system.StdEdge)
		if etype != system.ETypeNone && etype != edge.EType {
			// etype doesn't match
			return
		}

		if filterCmp(props, edge.Props) {
			es = append(es, edge)
		}
	}

	if in {
		vt.InEdges.ForEach(fef)
	} else {
		vt.OutEdges.ForEach(fef)
	}

	return
}

// SuccessorsWith returns a slice of vtTuples that are successors of the given vid,
// constraining the list to those that are connected by edges that pass the EdgeFilter,
// and the successor vertices pass the VertexFilter.
func (g *coreGraph) SuccessorsWith(egoID uint64, vef system.VEFilter) (vts system.VertexTupleVector) {
	return g.adjacentWith(egoID, vef, false)
}

// PredecessorsWith returns a slice of vtTuples that are predecessors of the given vid,
// constraining the list to those that are connected by edges that pass the EdgeFilter,
// and the predecessor vertices pass the VertexFilter.
func (g *coreGraph) PredecessorsWith(egoID uint64, vef system.VEFilter) (vts system.VertexTupleVector) {
	return g.adjacentWith(egoID, vef, true)
}

func (g *coreGraph) adjacentWith(egoID uint64, vef system.VEFilter, in bool) (vts system.VertexTupleVector) {
	etype, eprops := vef.EType(), vef.EProps()
	vtype, vprops := vef.VType(), vef.VProps()
	vt, err := g.Get(egoID)
	if err != nil {
		// vertex doesn't exist
		return
	}

	// Allow some quick parallelism by continuing to search edges while loading vertices
	// TODO performance test this to see if it's at all worth it
	vidchan := make(chan uint64, 10)

	var feef func(k string, v ps.Any)
	// TODO specialize the func for zero-cases
	feef = func(k string, v ps.Any) {
		edge := v.(system.StdEdge)
		if etype != system.ETypeNone && etype != edge.EType {
			// etype doesn't match
			return
		}

		if filterCmp(eprops, edge.Props) {
			if in {
				vidchan <- edge.Source
			} else {
				vidchan <- edge.Target
			}
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
	visited := make(map[uint64]struct{})
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
		if vtype != system.VTypeNone && vtype != adjvt.Vertex.Typ() {
			continue
		}

		if !filterCmp(vprops, adjvt.Vertex.Props()) {
			continue VertexInspector
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
func (g *coreGraph) VerticesWith(vf system.VFilter) (vs system.VertexTupleVector) {
	vtype, props := vf.VType(), vf.VProps()
	g.vtuples.ForEach(func(_ string, val ps.Any) {
		vt := val.(system.VertexTuple)
		if vtype != system.VTypeNone && vt.Vertex.Typ() != vtype {
			return
		}

		if filterCmp(props, vt.Vertex.Props()) {
			vs = append(vs, vt)
		}
	})

	return vs
}

// filterCmp compares all the values from the PropPair with values in the ps.Map,
// and returns true iff they all match. Non-existence is considered a match failure.
func filterCmp(pp []system.PropPair, m ps.Map) bool {
	for _, p := range pp {
		prop, exists := m.Lookup(p.K)
		if !exists {
			return false
		}

		dprop := prop.(system.Property)
		switch tv := dprop.Value.(type) {
		default:
			if tv != p.V {
				return false
			}
		case []byte:
			cmptv, ok := p.V.([]byte)
			if !ok || !bytes.Equal(tv, cmptv) {
				return false
			}
		}
	}

	return true
}
