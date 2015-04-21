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
			tuples = append(tuples, g.ensureVertex(msg.Id, sd))
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
func (g *CoreGraph) ensureVertex(msgid int, sd SplitData) (final vtTuple) {
	var vid int
	newvt := vtTuple{v: sd.Vertex, ie: ps.NewMap(), oe: ps.NewMap()}

	// FIXME so very hilariously O(n)
	matches := g.VerticesWith(qbv(sd.Vertex.Typ()))
	if len(matches) == 0 {
		g.vserial += 1
		newvt.id = g.vserial
		g.vtuples = g.vtuples.Set(strconv.Itoa(g.vserial), final)
		return newvt
	}
	// More than one match found. Iteratively narrow matches.

	// First, see if we can resolve the primary identifying edge
	for _, es := range sd.EdgeSpecs {
		// FIXME totally not ok to hardcode the edge spec types
		switch d := es.(type) {
		default:
			continue
		case interpret.EnvLink:
			edge, success := Resolve(g, msgid, newvt, d)
			if success {
				continue
			}
			break

		case SpecLocalLogic:
			edge, success := Resolve(g, msgid, newvt, d)
			if !success {
				continue
			}
			break
		}
	}

	var chk Identifier
	for _, idf := range Identifiers {
		if idf.CanIdentify(sd.Vertex) {
			chk = idf
		}
	}

	if chk == nil {
		// TODO obviously this is just to canary; change to error when stabilized
		panic("missing identify checker")
	}

	// destructive zero-allocations filtering
	filtered := matches[:0]
	for _, vt := range matches {
		// TODO calling off to dynamic code for comparison is not acceptable long term
		// TODO identifiers are broken right now, they look for edge props on the vtx
		if chk.Matches(sd.Vertex, vt.v) {
			filtered = append(filtered, vt)
		}
	}

	// Now just the ones left for which we need edges to differentiate
	filtered2 := filtered[:0]
	for _, vt := range filtered {
		//for es := range vt.
	}

	if tuples == nil {
	} else {
		ivt, _ := g.vtuples.Lookup(strconv.Itoa(vid))
		vt := ivt.(vtTuple)

		// TODO err
		nu, _ := vt.v.Merge(sd.Vertex)
		final = vtTuple{id: vid, ie: vt.ie, oe: vt.oe, v: nu}
		g.vtuples = g.vtuples.Set(strconv.Itoa(vid), final)
		//g.list[vid] = vtTuple{id: vid, e: vt.e, v: nu}
	}

	return
}

// Gets the vtTuple for a given vertex id.
func (g *CoreGraph) Get(id int) (vtTuple, error) {
	if id > g.vserial {
		return vtTuple{}, errors.New(fmt.Sprintf("Graph has only %d elements, no vertex yet exists with id %d", g.vserial, id))
	}

	vtx, exists := g.vtuples.Lookup(strconv.Itoa(id))
	if exists {
		return vtx.(vtTuple), nil
	} else {
		return vtTuple{}, errors.New(fmt.Sprintf("No vertex exists with id %d at the present revision of the graph", id))
	}
}
