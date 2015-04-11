package represent

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
)

// the main graph construct
type CoreGraph struct {
	list map[int]vtTuple
	// TODO experiment with replacing with a hash array-mapped trie
	vtuples ps.Map
	vserial int
}

type (
	// A value indicating a vertex's type. For now, done as a string.
	VType string
	// A value indicating an edge's type. For now, done as a string.
	EType string
)

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

const (
	VTypeNone VType = ""
	ETypeNone EType = ""
)

// Used in queries to specify property k/v pairs.
type PropQ struct {
	K string
	V interface{}
}

type Vertex interface {
	// Merges another vertex into this vertex. Error is indicated if the
	// dynamic types do not match.
	Merge(Vertex) (Vertex, error)
	// Returns a string representing the object type. Used for namespacing keys, etc.
	// While this is (currently) implemented as a method, its result must be invariant.
	// TODO use string-const generator, other tricks to enforce invariance, compact space use
	Typ() VType
	// Returns a persistent map with the vertex's properties.
	// TODO generate more type-restricted versions of the map?
	Props() ps.Map
}

type vtTuple struct {
	id int
	v  Vertex
	ie ps.Map
	oe ps.Map
}

type Property struct {
	MsgSrc int
	Key    string
	Value  interface{}
}

type veProcessingInfo struct {
	vt vtTuple
	es EdgeSpecs
}

type edgeSpecSet []*veProcessingInfo

func (ess edgeSpecSet) EdgeCount() (i int) {
	for _, tuple := range ess {
		i = i + len(tuple.es)
	}
	return
}

// the method to merge a message into the graph
func (g *CoreGraph) Merge(msg interpret.Message) {
	var ess edgeSpecSet

	// TODO see if some lookup + locality of reference perf benefit can be gained keeping all vertices loaded up locally

	// Process incoming elements from the message
	msg.Each(func(d interface{}) {
		// Split each input element into vertex and edge specs
		// TODO errs
		sds, _ := Split(d, msg.Id)
		// Ensure vertices are present

		var tuples []vtTuple
		for _, sd := range sds {
			// TODO this can't work now, dammit, need edge context. UGH
			tuples = append(tuples, g.ensureVertex(sd.Vertex))
		}

		// Collect edge specs for later processing
		for k, tuple := range tuples {
			ess = append(ess, &veProcessingInfo{
				vt: tuple,
				es: sds[k].EdgeSpecs,
			})
		}
	})

	// All vertices processed. now, process edges in passes, ensuring that each
	// pass diminishes the number of remaining edges. If it doesn't, the remaining
	// edges need to be attached to null-vertices of the appropriate type.
	//
	// This is a little wasteful, but it's the simplest way to let any possible
	// dependencies between edges work themselves out. It has provably incorrect
	// cases, however, and will need to be replaced.
	var ec, lec int
	for ec = ess.EdgeCount(); ec != lec; ec = ess.EdgeCount() {
		lec = ec
		for _, info := range ess {
			for k, spec := range info.es {
				edge, success := Resolve(g, msg.Id, info.vt, spec)
				if success {
					edge.Source = info.vt.id
					if edge.id == 0 {
						// new edge, allocate a new id for it
						g.vserial++
						edge.id = g.vserial
					}

					// FIXME make sure assignment makes it back into ess slice
					info.vt.oe = info.vt.oe.Set(strconv.Itoa(edge.id), edge)
					g.vtuples = g.vtuples.Set(strconv.Itoa(info.vt.id), info.vt)

					// TODO setting multiple times is silly and wasteful
					any, _ := g.vtuples.Lookup(strconv.Itoa(edge.Target))
					tvt := any.(vtTuple)
					tvt.ie = tvt.ie.Set(strconv.Itoa(edge.id), edge)

					info.es = append(info.es[:k], info.es[k+1:]...)
				}
			}
		}
	}

	// TODO attempt to de-orphan items here?
}

// Ensures the vertex is present. Merges according to type-specific logic if
// it is present, otherwise adds the vertex.
//
// Either way, return value is the vid for the vertex.
func (g *CoreGraph) ensureVertex(vtx Vertex) (final vtTuple) {
	vid := g.Find(vtx)

	if vid != 0 {
		ivt, _ := g.vtuples.Lookup(strconv.Itoa(vid))
		vt := ivt.(vtTuple)

		// TODO err
		nu, _ := vt.v.Merge(vtx)
		final = vtTuple{id: vid, ie: vt.ie, oe: vt.oe, v: nu}
		g.vtuples = g.vtuples.Set(strconv.Itoa(vid), final)
		//g.list[vid] = vtTuple{id: vid, e: vt.e, v: nu}
	} else {
		g.vserial++
		vid = g.vserial
		final = vtTuple{id: vid, v: vtx, ie: ps.NewMap(), oe: ps.NewMap()}
		g.vtuples = g.vtuples.Set(strconv.Itoa(vid), final)
		//g.list[g.vserial] = vtTuple{v: vtx}
	}

	return
}

// Searches for an instance of the vertex within the graph. If found,
// returns the vertex id, otherwise returns 0.
//
// TODO this is really a querying method. needs to be replaced by that whole subsystem
func (g *CoreGraph) Find(vtx Vertex) int {
	// FIXME so very hilariously O(n)

	var chk Identifier
	for _, idf := range Identifiers {
		if idf.CanIdentify(vtx) {
			chk = idf
		}
	}

	// we hit this case iff there's an object type our identifiers can't work
	// with. which should, eventually, be structurally impossible by this point
	if chk == nil {
		return 0
	}

	for id, v := range g.list {
		if chk.Matches(v.v, vtx) {
			return id
		}
	}

	return 0
}

// Gets the vtTuple for a given vertex id.
func (g *CoreGraph) Get(id int) (vtTuple, error) {
	if id > g.vserial {
		return vtTuple{}, errors.New(fmt.Sprintf("Graph has only ", g.vserial, "elements, no vertex yet exists with id", id))
	}

	vtx, exists := g.vtuples.Lookup(strconv.Itoa(id))
	if exists {
		return vtx.(vtTuple), nil
	} else {
		return vtTuple{}, errors.New(fmt.Sprintf("No vertex exists with id", id, "at the present revision of the graph"))
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
