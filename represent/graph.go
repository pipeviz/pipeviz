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
	Typ() string
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
		vtx, edges, _ := Split(d, msg.Id)
		// Ensure the vertex is present
		tuple := g.ensureVertex(vtx)

		// Collect edge specs for later processing
		ess = append(ess, &veProcessingInfo{
			vt: tuple,
			es: edges,
		})
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
				edge, success := Resolve(g, info.vt, spec)
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

// the func we eventually aim to fulfill, replacing Merge for integrating messages
//func (g CoreGraph) Cons(interpret.Message) CoreGraph, Delta, error {}

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

func (g *CoreGraph) Vertices(f func(Vertex, int) bool) {
	for id, vt := range g.list {
		if f(vt.v, id) {
			return
		}
	}
}

func (g *CoreGraph) Get(id int) (Vertex, error) {
	if id > g.vserial {
		return nil, errors.New(fmt.Sprintf("Graph has only ", g.vserial, "elements, no vertex yet exists with id", id))
	}

	vtx, exists := g.vtuples.Lookup(strconv.Itoa(id))
	if exists {
		return vtx.(vtTuple).v, nil
	} else {
		return nil, errors.New(fmt.Sprintf("No vertex exists with id", id, "at the present revision of the graph"))
	}
}

// Inspects the indicated vertex's set of out-edges, returning a slice of
// those that match on type and properties. ETypeNone and nil can be passed
// for the last two parameters respectively, in which case the filters will
// be bypassed.
func (g *CoreGraph) OutWith(egoId int, etype EType, props []PropQ) (es []StandardEdge) {
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

	vt.oe.ForEach(fef)
	return
}
