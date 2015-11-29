package represent

import (
	"fmt"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/maputil"
	"github.com/pipeviz/pipeviz/types/semantic"
	"github.com/pipeviz/pipeviz/types/system"
)

// partialRecord is a pair of vertex id and edge id. They identify an edge that needs
// to be resolved.
type partialRecord struct {
	vid, eid uint64
}

// Merge takes a set of input data in UnifyInstructionForms, merges it into
// the graph and returns a pointer to the new version of the graph.
func (og *coreGraph) Merge(msgid uint64, uifs []system.UnifyInstructionForm) system.CoreGraph {
	// TODO use a buffering pool to minimize allocs
	var ess edgeSpecSet

	logEntry := logrus.WithFields(logrus.Fields{
		"system": "engine",
		"msgid":  msgid,
	})

	logEntry.Infof("Merging message %d into graph", msgid)

	g := og.clone()
	g.msgid = msgid
	var newPartials []partialRecord

	// Ensure vertices, then record into intermediate, orphan-enabling container
	for _, uif := range uifs {
		vt, np, err := toTuple(g, msgid, uif)
		if err != nil {
			// Error was already logged in toTuple, so here we can just skip the bugger
			continue
		}
		if len(np) != 0 {
			newPartials = append(newPartials, np...)
		}

		ess = append(ess, &veProcessingInfo{
			vt:  vt,
			uif: uif, // TODO REMOVE
			// copy out the edges for later bookkeeping
			e:     append(uif.ScopingSpecs(), uif.EdgeSpecs()...),
			msgid: msgid,
		})
	}

	logEntry.Infof("Adding %d orphan edge spec sets from previous merges", len(g.orphans))
	// Reinclude the held-over set of orphans for edge (re-)resolutions
	var ess2 edgeSpecSet
	// TODO lots of things very wrong with this approach, but works for first pass
	for _, orphan := range g.orphans {
		// vertex ident failed; try again now that new vertices are present
		if orphan.vt.ID == 0 {
			// this is temporary anyway, but there's no way an err could come back on an existing orphan
			orphan.vt, _, _ = toTuple(g, orphan.msgid, orphan.uif)
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
	var specs []system.EdgeSpec
	for ec = ess.EdgeCount(); ec != 0 && ec != lec; ec = ess.EdgeCount() {
		pass++
		lec = ec
		l2 := logEntry.WithFields(logrus.Fields{
			"pass":       pass,
			"edge-count": ec,
		})

		l2.Debug("Beginning edge resolution pass")
		for infokey, info := range ess {
			l3 := logEntry.WithFields(logrus.Fields{
				"vid":   info.vt.ID,
				"vtype": info.vt.Vertex.Type,
			})
			// Ensure our local copy of the tuple is up to date
			info.vt, _ = g.Get(info.vt.ID)

			// Zero-alloc filtering technique
			specs, info.e = info.e, info.e[:0]
			for _, spec := range specs {
				l3.Debugf("Resolving EdgeSpec of type %T", spec)
				edge, success, err := semantic.Resolve(spec, g, msgid, info.vt)

				// non-nil err indicates a type that has not registered its resolver correctly
				if err != nil {
					l3.WithFields(logrus.Fields{
						"spec-type": fmt.Sprintf("%T", spec),
						"err":       err,
					}).Errorf("Resolve returned err, indicating resolve func was not correctly registered for spec type. Discarding spec")
					continue
				}

				if success {
					l4 := l3.WithFields(logrus.Fields{
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
					g.vtuples = g.vtuples.Set(info.vt.ID, info.vt)

					tvt, _ := g.vtuples.Get(edge.Target)

					tvt.InEdges = tvt.InEdges.Set(i2a(edge.ID), edge)
					g.vtuples = g.vtuples.Set(tvt.ID, tvt)
				} else {
					l3.Debug("Unsuccessful edge resolution; reattempt on next pass")
					// FIXME mem leaks if done this way...?
					info.e = append(info.e, spec)
				}
			}
			// set the processing info back into its original position in the slice
			ess[infokey] = info
		}
	}
	logEntry.WithField("passes", pass).Info("Edge resolution complete")

	g.orphans = g.orphans[:0]
	for _, info := range ess {
		if len(info.e) == 0 {
			continue
		}

		g.orphans = append(g.orphans, info)
	}
	logEntry.Infof("Adding %d orphan edge spec sets from previous merges", len(g.orphans))

	return g
}

// toTuple performs a single pass over a UIF: it attempts unify the vertex and resolve its edge specs.
//
// Returns a fully-formed VertexTuple, or an error iff basic vertex unification failed.
func toTuple(g *coreGraph, msgid uint64, sd system.UnifyInstructionForm) (system.VertexTuple, []partialRecord, error) {
	logEntry := logrus.WithFields(logrus.Fields{
		"system": "engine",
		"msgid":  msgid,
		"vtype":  sd.Vertex().Type(),
	})

	// TODO this is where resolving of scoping edges should have to be done

	logEntry.Debug("Performing vertex unification")
	vid, err := semantic.Unify(g, sd)
	if err != nil {
		logEntry.WithFields(logrus.Fields{
			"vtx-type": sd.Vertex().Type(),
			"err":      err,
		}).Errorf("Unify returned err, indicating unify func was not correctly registered for vtx type. Discarding vtx")
		// At this point, all we can do is return an err
		return system.VertexTuple{}, nil, err
	}

	// FIXME
	// There is a 2x2 matrix resulting in four possible (main) paths:
	//
	// - Unification succeeded and the UIF has no scoping specs
	// - Unification failed and the UIF has no scoping specs
	// - Unification succeeded and the UIF has some scoping specs
	// - Unification failed and the UIF has some scoping specs
	//
	// Things are generally clearer in the first two cases (no scoping specs)
	// because there's no icky circular dependency feeling. In the third case,
	// there's a particularly big potential gap due to the fact that the vtx Unify
	// *could* come back with one vertex, but then actually resolving the
	// scoping specs could be resolved to point to some different scope. What on
	// earth would that even mean? No idea. This is case in point why we need
	// unification as a separate process for edges.

	if vid == 0 {
		v, p := toNewTuple(g, msgid, sd, logEntry)
		return v, p, nil
	} else {
		v, p := toExistingTuple(g, vid, msgid, sd, logEntry)
		return v, p, nil
	}
}

func toNewTuple(g *coreGraph, msgid uint64, uif system.UnifyInstructionForm, log *logrus.Entry) (system.VertexTuple, []partialRecord) {
	log.Debug("No match on unification, creating a new vertex")

	var partials []partialRecord
	final := system.VertexTuple{
		Vertex: system.StdVertex{
			Type:       uif.Vertex().Type(),
			Properties: maputil.RawMapToPropPMap(msgid, false, uif.Vertex().Properties()),
		},
		Incomplete: false,
		InEdges:    ps.NewMap(),
		OutEdges:   ps.NewMap(),
	}

	g.vserial++
	final.ID = g.vserial
	g.vtuples = g.vtuples.Set(g.vserial, final)

	// Resolve scoping edge specs early, here
	for _, spec := range uif.ScopingSpecs() {
		log.Debugf("Doing early resolve on EdgeSpec of type %T", spec)
		edge, success, err := semantic.Resolve(spec, g, msgid, final)

		if err != nil {
			log.WithFields(logrus.Fields{
				"spec-type": fmt.Sprintf("%T", spec),
				"err":       err,
			}).Errorf("Resolve returned err, indicating resolve func was not correctly registered for spec type. Discarding spec")
			// If a scoping spec is registered incorrectly, this vertex can never be complete.
			// This really is a *huge* integrity problem, but there's little we can sanely do to keep
			// a marker around, so we just skip this edge entirely with the expectation that the user
			// will have to make a fix to the semantic system for anything to make sense anyway.
			final.Incomplete = true
			continue
		}

		// Currently, Success -> ¬Incomplete. While the onus should be on the
		// implementor to maintain this implication, we check here and
		// register an error if there is a mismatch.
		if (success && edge.Incomplete) || (!success && !edge.Incomplete) {
			log.WithFields(logrus.Fields{
				"etype":      edge.EType,
				"success":    success,
				"incomplete": edge.Incomplete,
			}).Error("Disagreement between Resolve 'success' return and edge's Incomplete flag; trusting return")
		}

		g.vserial++
		edge.ID = g.vserial
		final.OutEdges = final.OutEdges.Set(i2a(edge.ID), edge)

		if success {
			edge.Incomplete = false // normalize based on Resolve's return
			log.WithField("target-vid", edge.Target).Debug("Early resolve succeeded")

			// set edge in reverse direction, too
			tvt, _ := g.vtuples.Get(edge.Target)
			tvt.InEdges = tvt.InEdges.Set(i2a(edge.ID), edge)

			g.vtuples = g.vtuples.Set(tvt.ID, tvt)
		} else {
			edge.Incomplete = true  // normalize based on Resolve's return
			final.Incomplete = true // if a scoping edge is incomplete, so is the vertex
			partials = append(partials, partialRecord{vid: final.ID, eid: edge.Target})
			log.Debug("Early resolve failed")
		}

		// FIXME this can be removed...right?
		//g.vtuples = g.vtuples.Set(final.ID, final)
	}

	for _, spec := range uif.EdgeSpecs() {
		edge, success, err := semantic.Resolve(spec, g, msgid, final)
		if err != nil {
			log.WithFields(logrus.Fields{
				"spec-type": fmt.Sprintf("%T", spec),
				"err":       err,
			}).Errorf("Resolve returned err, indicating resolve func was not correctly registered for spec type. Discarding spec")
			// Not nearly as big a deal to miss one of these as to miss a scoping spec, but still log the error.
			continue
		}

		if (success && edge.Incomplete) || (!success && !edge.Incomplete) {
			log.WithFields(logrus.Fields{
				"etype":      edge.EType,
				"success":    success,
				"incomplete": edge.Incomplete,
			}).Error("Disagreement between Resolve 'success' return and edge's Incomplete flag; trusting return")
		}

		g.vserial++
		edge.ID = g.vserial
		final.OutEdges = final.OutEdges.Set(i2a(edge.ID), edge)

		if success {
			edge.Incomplete = false // normalize based on Resolve's return
			log.WithField("target-vid", edge.Target).Debug("Initial resolve successful")

			// set edge in reverse direction, too
			tvt, _ := g.vtuples.Get(edge.Target)
			tvt.InEdges = tvt.InEdges.Set(i2a(edge.ID), edge)

			g.vtuples = g.vtuples.Set(tvt.ID, tvt)
		} else {
			edge.Incomplete = true // normalize based on Resolve's return
			partials = append(partials, partialRecord{vid: final.ID, eid: edge.Target})
			log.Debug("Early resolve failed")
		}

		// FIXME this can be removed...right?
		//g.vtuples = g.vtuples.Set(final.ID, final)
	}

	g.vtuples = g.vtuples.Set(final.ID, final)
	return final, partials
}

func toExistingTuple(g *coreGraph, vid, msgid uint64, uif system.UnifyInstructionForm, log *logrus.Entry) (system.VertexTuple, []partialRecord) {
	var err error

	log.WithField("vid", vid).Debug("Unification resulted in match")
	vt, _ := g.vtuples.Get(vid)

	var partials []partialRecord
	// TODO ugh, pointless stupid duplicating/copy
	vt.Vertex, err = vt.Vertex.Merge(system.StdVertex{
		Type:       uif.Vertex().Type(),
		Properties: maputil.RawMapToPropPMap(msgid, false, uif.Vertex().Properties()),
	})

	if err != nil {
		log.WithFields(logrus.Fields{
			"vid": vid,
			"err": err,
		}).Warn("Merge of vertex properties returned an error; vertex will continue update into graph anyway")
	}

	for _, spec := range uif.ScopingSpecs() {
		log.Debugf("Re-resolving on EdgeSpec of type %T", spec)
		edge, success, err := semantic.Resolve(spec, g, msgid, vt)

		if err != nil {
			log.WithFields(logrus.Fields{
				"spec-type": fmt.Sprintf("%T", spec),
				"err":       err,
			}).Errorf("Resolve returned err, indicating resolve func was not correctly registered for spec type. Discarding spec")
			// FIXME marking the vertex incomplete, but with no recourse to completeness...SO wrong.
			// TBH this should maybe panic, this state is absolutely nonsensical.
			vt.Incomplete = true
			continue
		}

		// Currently, Success -> ¬Incomplete. While the onus should be on the
		// implementor to maintain this implication, we check here and
		// register an error if there is a mismatch.
		if (success && edge.Incomplete) || (!success && !edge.Incomplete) {
			log.WithFields(logrus.Fields{
				"etype":      edge.EType,
				"success":    success,
				"incomplete": edge.Incomplete,
			}).Error("Disagreement between Resolve 'success' return and edge's Incomplete flag; trusting return")
		}

		// The edge may or may not be new.
		// TODO reorganize once edge unification is split out
		if edge.ID == 0 {
			g.vserial++
			edge.ID = g.vserial
		}

		vt.OutEdges = vt.OutEdges.Set(i2a(edge.ID), edge)

		if success {
			edge.Incomplete = false // normalize based on Resolve's return
			log.WithField("target-vid", edge.Target).Debug("Early resolve succeeded")

			// set edge in reverse direction, too
			tvt, _ := g.vtuples.Get(edge.Target)
			tvt.InEdges = tvt.InEdges.Set(i2a(edge.ID), edge)

			g.vtuples = g.vtuples.Set(tvt.ID, tvt)
		} else {
			edge.Incomplete = true // normalize based on Resolve's return
			vt.Incomplete = true   // if a scoping edge is incomplete, so is the vertex
			partials = append(partials, partialRecord{vid: vt.ID, eid: edge.Target})
			log.Debug("Early resolve failed")
		}

		// FIXME this can be removed...right?
		//g.vtuples = g.vtuples.Set(vt.ID, vt)
	}

	for _, spec := range uif.EdgeSpecs() {
		edge, success, err := semantic.Resolve(spec, g, msgid, vt)
		if err != nil {
			log.WithFields(logrus.Fields{
				"spec-type": fmt.Sprintf("%T", spec),
				"err":       err,
			}).Errorf("Resolve returned err, indicating resolve func was not correctly registered for spec type. Discarding spec")
			// Not nearly as big a deal to miss one of these as to miss a scoping spec, but still log the error.
			continue
		}

		// A weird possibility exists here: unify with an existing, complete edge, but become incomplete.
		if (success && edge.Incomplete) || (!success && !edge.Incomplete) {
			log.WithFields(logrus.Fields{
				"etype":      edge.EType,
				"success":    success,
				"incomplete": edge.Incomplete,
			}).Error("Disagreement between Resolve 'success' return and edge's Incomplete flag; trusting return")
		}

		// The edge may or may not be new.
		// TODO reorganize once edge unification is split out
		if edge.ID == 0 {
			g.vserial++
			edge.ID = g.vserial
		}

		vt.OutEdges = vt.OutEdges.Set(i2a(edge.ID), edge)

		if success {
			edge.Incomplete = false // normalize based on Resolve's return
			log.WithField("target-vid", edge.Target).Debug("Initial resolve successful")

			// set edge in reverse direction, too
			tvt, _ := g.vtuples.Get(edge.Target)
			tvt.InEdges = tvt.InEdges.Set(i2a(edge.ID), edge)

			g.vtuples = g.vtuples.Set(tvt.ID, tvt)
		} else {
			edge.Incomplete = true // normalize based on Resolve's return
			partials = append(partials, partialRecord{vid: vt.ID, eid: edge.Target})
			log.Debug("Early resolve failed")
		}

		// FIXME this can be removed...right?
		//g.vtuples = g.vtuples.Set(vt.ID, vt)
	}

	g.vtuples = g.vtuples.Set(vid, vt)
	return vt, partials
}
