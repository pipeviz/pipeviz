package represent

import (
	"errors"
	"fmt"
	"strconv"

	log "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/represent/types"
)

var i2a = strconv.Itoa

// the main graph construct
type coreGraph struct {
	// TODO experiment with replacing with a hash array-mapped trie
	msgid   uint64
	vserial int
	vtuples ps.Map
	orphans edgeSpecSet // FIXME breaks immut
}

func NewGraph() types.CoreGraph {
	log.WithFields(log.Fields{
		"system": "engine",
	}).Debug("New coreGraph created")

	return &coreGraph{vtuples: ps.NewMap(), vserial: 0}
}

type veProcessingInfo struct {
	vt    types.VertexTuple
	es    types.EdgeSpecs
	msgid uint64
}

type edgeSpecSet []*veProcessingInfo

func (ess edgeSpecSet) EdgeCount() (i int) {
	for _, tuple := range ess {
		i = i + len(tuple.es)
	}
	return
}

// Clones the graph object, and the map pointers it contains.
func (g *coreGraph) clone() *coreGraph {
	var cp coreGraph
	cp = *g
	return &cp
}

func (g *coreGraph) MsgId() uint64 {
	return g.msgid
}

// the method to merge a message into the graph
func (og *coreGraph) Merge(msg interpret.Message) types.CoreGraph {
	// TODO use a buffering pool to minimize allocs
	var ess edgeSpecSet

	logEntry := log.WithFields(log.Fields{
		"system": "engine",
		"msgid":  msg.Id,
	})

	logEntry.Infof("Merging message %d into graph", msg.Id)

	g := og.clone()
	g.msgid = msg.Id

	// Process incoming elements from the message
	msg.Each(func(d interface{}) {
		// Split each input element into vertex and edge specs
		sds, err := Split(d, msg.Id)

		if err != nil {
			logEntry.WithField("err", err).Warnf("Error while splitting input element of type %T; discarding", d)
			return
		}

		// Ensure vertices are present
		var tuples []types.VertexTuple
		for _, sd := range sds {
			tuples = append(tuples, g.ensureVertex(msg.Id, sd))
		}

		// Collect edge specs for later processing
		for k, tuple := range tuples {
			ess = append(ess, &veProcessingInfo{
				vt:    tuple,
				es:    sds[k].EdgeSpecs,
				msgid: msg.Id,
			})
		}
	})
	logEntry.Infof("Splitting all message elements produced %d edge spec sets", len(ess))

	logEntry.Infof("Adding %d orphan edge spec sets from previous merges", len(g.orphans))
	// Reinclude the held-over set of orphans for edge (re-)resolutions
	var ess2 edgeSpecSet
	// TODO lots of things very wrong with this approach, but works for first pass
	for _, orphan := range g.orphans {
		// vertex ident failed; try again now that new vertices are present
		if orphan.vt.ID == 0 {
			orphan.vt = g.ensureVertex(orphan.msgid, types.SplitData{orphan.vt.Vertex, orphan.es})
		} else {
			// ensure we have latest version of vt
			vt, err := g.Get(orphan.vt.ID)
			if err != nil {
				// but if that vid has gone away, forget about it completely
				logEntry.Infof("Orphan vid %d went away, discarding from orphan list", orphan.vt.ID)
				continue
			}
			orphan.vt = vt
		}

		ess2 = append(ess2, orphan)
	}

	// Put orphan stuff first so that it's guaranteed to be overwritten on conflict
	ess = append(ess2, ess...)

	// All vertices processed. Now, process edges in passes, ensuring that each
	// pass diminishes the number of remaining edges. If it doesn't, the remaining
	// edges need to be attached to null-vertices of the appropriate type.
	//
	// This is a little wasteful, but it's the simplest way to let any possible
	// dependencies between edges work themselves out. It has provably incorrect
	// cases, however, and will need to be replaced.
	var ec, lec, pass int
	for ec = ess.EdgeCount(); ec != 0 && ec != lec; ec = ess.EdgeCount() {
		pass += 1
		lec = ec
		l2 := logEntry.WithFields(log.Fields{
			"pass":       pass,
			"edge-count": ec,
		})

		l2.Debug("Beginning edge resolution pass")
		for infokey, info := range ess {
			l3 := logEntry.WithFields(log.Fields{
				"vid":   info.vt.ID,
				"vtype": info.vt.Vertex.Typ(),
			})
			specs := info.es
			info.es = info.es[:0]
			for _, spec := range specs {
				l3.Debugf("Resolving EdgeSpec of type %T", spec)
				edge, success := Resolve(g, msg.Id, info.vt, spec)
				if success {
					l4 := l3.WithFields(log.Fields{
						"target-vid": edge.Target,
						"etype":      edge.EType,
					})

					l4.Debug("Edge resolved successfully")

					edge.Source = info.vt.ID
					if edge.ID == 0 {
						// new edge, allocate a new id for it
						g.vserial++
						edge.ID = g.vserial
						l4.WithField("edge-id", edge.ID).Debug("New edge created")
					} else {
						l4.WithField("edge-id", edge.ID).Debug("Edge will merge over existing edge")
					}

					info.vt.OutEdges = info.vt.OutEdges.Set(i2a(edge.ID), edge)
					g.vtuples = g.vtuples.Set(i2a(info.vt.ID), info.vt)

					any, _ := g.vtuples.Lookup(i2a(edge.Target))

					tvt := any.(types.VertexTuple)
					tvt.InEdges = tvt.InEdges.Set(i2a(edge.ID), edge)
					g.vtuples = g.vtuples.Set(i2a(tvt.ID), tvt)
				} else {
					l3.Debug("Unsuccessful edge resolution; reattempt on next pass")
					// FIXME mem leaks if done this way...?
					info.es = append(info.es, spec)
				}
			}
			// set the processing info back into its original position in the slice
			ess[infokey] = info
		}
	}
	logEntry.WithField("passes", pass).Info("Edge resolution complete")

	g.orphans = g.orphans[:0]
	for _, info := range ess {
		if len(info.es) == 0 {
			continue
		}

		g.orphans = append(g.orphans, info)
	}
	logEntry.Infof("Adding %d orphan edge spec sets from previous merges", len(g.orphans))

	return g
}

// Ensures the vertex is present. Merges according to type-specific logic if
// it is present, otherwise adds the vertex.
//
// Either way, return value is the vid for the vertex.
func (g *coreGraph) ensureVertex(msgid uint64, sd types.SplitData) (final types.VertexTuple) {
	logEntry := log.WithFields(log.Fields{
		"system": "engine",
		"msgid":  msgid,
		"vtype":  sd.Vertex.Typ(),
	})

	logEntry.Debug("Performing vertex unification")
	vid := Identify(g, sd)

	if vid == 0 {
		logEntry.Debug("No match on unification, creating new vertex")
		final = types.VertexTuple{Vertex: sd.Vertex, InEdges: ps.NewMap(), OutEdges: ps.NewMap()}
		g.vserial += 1
		final.ID = g.vserial
		g.vtuples = g.vtuples.Set(i2a(g.vserial), final)
		// TODO remove this - temporarily cheat here by promoting EnvLink resolution, since so much relies on it
		for _, spec := range sd.EdgeSpecs {
			switch spec.(type) {
			case interpret.EnvLink, SpecDatasetHierarchy:
				logEntry.Debugf("Doing early resolve on EdgeSpec of type %T", spec)
				edge, success := Resolve(g, msgid, final, spec)
				if success { // could fail if corresponding env not yet declared
					logEntry.WithField("target-vid", edge.Target).Debug("Early resolve succeeded")
					g.vserial += 1
					edge.ID = g.vserial
					final.OutEdges = final.OutEdges.Set(i2a(edge.ID), edge)

					// set edge in reverse direction, too
					any, _ := g.vtuples.Lookup(i2a(edge.Target))
					tvt := any.(types.VertexTuple)
					tvt.InEdges = tvt.InEdges.Set(i2a(edge.ID), edge)
					g.vtuples = g.vtuples.Set(i2a(tvt.ID), tvt)
					g.vtuples = g.vtuples.Set(i2a(final.ID), final)
				} else {
					logEntry.Debug("Early resolve failed")
				}
			}
		}
	} else {
		logEntry.WithField("vid", vid).Debug("Unification resulted in match")
		ivt, _ := g.vtuples.Lookup(i2a(vid))
		vt := ivt.(types.VertexTuple)

		nu, err := vt.Vertex.Merge(sd.Vertex)
		if err != nil {
			logEntry.WithFields(log.Fields{
				"vid": vid,
				"err": err,
			}).Warn("Merge of vertex properties returned an error; vertex will continue update into graph anyway")
		}

		final = types.VertexTuple{ID: vid, InEdges: vt.InEdges, OutEdges: vt.OutEdges, Vertex: nu}
		g.vtuples = g.vtuples.Set(i2a(vid), final)
	}

	return
}

// Gets the vtTuple for a given vertex id.
func (g *coreGraph) Get(id int) (types.VertexTuple, error) {
	if id > g.vserial {
		return types.VertexTuple{}, errors.New(fmt.Sprintf("Graph has only %d elements, no vertex yet exists with id %d", g.vserial, id))
	}

	vtx, exists := g.vtuples.Lookup(i2a(id))
	if exists {
		return vtx.(types.VertexTuple), nil
	} else {
		return types.VertexTuple{}, errors.New(fmt.Sprintf("No vertex exists with id %d at the present revision of the graph", id))
	}
}
