package represent

import (
	log "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/represent/types"
)

func Identify(g CoreGraph, sd SplitData) int {
	ids, definitive := identifyDefault(g, sd)
	// default one didn't do it, use specialists to narrow further

	switch sd.Vertex.Typ() {
	case "logic-state", "comm", "parent-dataset", "dataset":
		// TODO vertexDataset needs special disambiguation; it has a second structural edge (poor idea anyway)
		if len(ids) == 1 && definitive {
			return ids[0]
		}
		return 0
	case "test-result":
		return identifyByGitHashSpec(g, sd, ids)
	}

	if ids == nil {
		return 0
	} else if len(ids) == 1 {
		return ids[0]
	}

	log.WithFields(log.Fields{
		"system": "identification",
		"vtype":  sd.Vertex.Typ(),
	}).Panic("Vertex identification failed") // TODO much better info for this
	panic("Vertex identification failed")
}

// Peforms a generalized search for vertex identification, with a particular head-nod to
// those many vertices that need an EnvLink to fully resolve their identity.
//
// Returns a slice of candidate ids, and a bool indicating whether, if there is only one
// result, that result should be considered a definitive match.
//
// FIXME the responsibility murkiness is making this a horrible snarl, fix this shit ASAP
func identifyDefault(g CoreGraph, sd SplitData) (ret []int, definitive bool) {
	matches := g.VerticesWith(Qbv(sd.Vertex.Typ()))
	if len(matches) == 0 {
		// no vertices of this type, safe to bail early
		return nil, false
	}

	// do simple pass with identifiers to check possible matches
	var chk Identifier
	for _, idf := range Identifiers {
		if idf.CanIdentify(sd.Vertex) {
			chk = idf
		}
	}

	if chk == nil {
		// TODO obviously this is just to canary; change to error when stabilized
		log.WithFields(log.Fields{
			"system": "identification",
			"vtype":  sd.Vertex.Typ(),
		}).Panic("Missing identify checker for vertex type") // TODO much better info for this
	}

	filtered := matches[:0] // destructive zero-alloc filtering
	for _, candidate := range matches {
		if chk.Matches(sd.Vertex, candidate.v) {
			filtered = append(filtered, candidate)
		}
	}

	var envlink interpret.EnvLink
	var hasEl bool
	// see if we have an envlink in the edgespecs - if so, filter with it
	for _, es := range sd.EdgeSpecs {
		if el, ok := es.(interpret.EnvLink); ok {
			hasEl = true
			envlink = el
		}
	}

	// filter again to avoid multiple vertices overwriting each other
	// TODO this is a temporary measure until we move identity edge resolution up into vtx identification process
	filtered2 := filtered[:0]
	if hasEl {
		newvt := VertexTuple{v: sd.Vertex, ie: ps.NewMap(), oe: ps.NewMap()}
		edge, success := Resolve(g, 0, newvt, envlink)

		if !success {
			// FIXME failure to resolve envlink doesn't necessarily mean no match
			return nil, false
		}

		for _, candidate := range filtered {
			for _, edge2 := range g.OutWith(candidate.id, Qbe(types.EType("envlink"))) {
				filtered2 = append(filtered2, candidate)
				if edge2.Target == edge.Target {
					return []int{candidate.id}, true
				}
			}
		}
	} else {
		// no el, suggesting our candidate list is good
		filtered2 = filtered
	}

	for _, vt := range filtered2 {
		ret = append(ret, vt.id)
	}

	return ret, false
}

// Narrow a match list by looking for alignment on a git commit sha1
func identifyByGitHashSpec(g CoreGraph, sd SplitData, matches []int) int {
	for _, es := range sd.EdgeSpecs {
		// first find the commit spec
		if spec, ok := es.(SpecCommit); ok {
			// then search otherwise-matching vertices for a corresponding sha1 edge
			for _, matchvid := range matches {
				if len(g.OutWith(matchvid, Qbe(types.EType("version"), "sha1", spec.Sha1))) == 1 {
					return matchvid
				}
			}

			break // there *should* be one and only one
		}
	}

	// no match, it's a newbie
	return 0
}

// older stuff below

var Identifiers []Identifier

func init() {
	Identifiers = []Identifier{
		IdentifierDataset{},
		IdentifierCommit{},
		IdentifierGitTag{},
		IdentifierGitBranch{},
		IdentifierTestResult{},
		IdentifierParentDataset{},
		IdentifierYumPkg{},
		IdentifierGeneric{},
	}
}

// Identifiers represent the logic for identifying specific types of objects
// that may be contained within the graph, and finding matches between these
// types of objects
type Identifier interface {
	CanIdentify(data types.Vtx) bool
	Matches(a types.Vtx, b types.Vtx) bool
}

// New generic identifier - temporary!
type IdentifierGeneric struct{}

func (i IdentifierGeneric) CanIdentify(data types.Vtx) bool {
	switch data.Typ() {
	case "environment", "logic-state", "process", "comm":
		return true
	default:
		return false
	}
}

func (i IdentifierGeneric) Matches(a types.Vtx, b types.Vtx) bool {
	if a.Typ() != b.Typ() {
		return false
	}

	switch a.Typ() {
	case "environment":
		return matchAddress(a.Props(), b.Props())
	case "logic-state":
		return mapValEq(a.Props(), b.Props(), "path")
	case "process":
		return mapValEq(a.Props(), b.Props(), "pid")
	case "comm":
		_, haspath := a.Props().Lookup("path")
		if haspath {
			return mapValEqAnd(a.Props(), b.Props(), "type", "path")
		} else {
			return mapValEqAnd(a.Props(), b.Props(), "type", "port")
		}
	default:
		return false
	}
}

// Helper func to match addresses
func matchAddress(a, b ps.Map) bool {
	// For now, match if *any* non-empty of hostname, ipv4, or ipv6 match
	// TODO this needs moar thinksies
	if mapValEq(a, b, "hostname") {
		return true
	}
	if mapValEq(a, b, "ipv4") {
		return true
	}
	if mapValEq(a, b, "ipv6") {
		return true
	}

	return false
}

// Helper func to match env links
func matchEnvLink(a, b ps.Map) bool {
	return mapValEq(a, b, "nick") || matchAddress(a, b)
}

type IdentifierDataset struct{}

func (i IdentifierDataset) CanIdentify(data types.Vtx) bool {
	_, ok := data.(vertexDataset)
	return ok
}

func (i IdentifierDataset) Matches(a types.Vtx, b types.Vtx) bool {
	l, ok := a.(vertexDataset)
	if !ok {
		return false
	}
	r, ok := b.(vertexDataset)
	if !ok {
		return false
	}

	return mapValEq(l.Props(), r.Props(), "name")
}

type IdentifierCommit struct{}

func (i IdentifierCommit) CanIdentify(data types.Vtx) bool {
	_, ok := data.(vertexCommit)
	return ok
}

func (i IdentifierCommit) Matches(a types.Vtx, b types.Vtx) bool {
	l, ok := a.(vertexCommit)
	if !ok {
		return false
	}
	r, ok := b.(vertexCommit)
	if !ok {
		return false
	}

	// TODO mapValEq should be able to handle this
	//lsha, lexists := l.Props().Lookup("sha1")
	//rsha, rexists := r.Props().Lookup("sha1")
	//return rexists && lexists && bytes.Equal(lsha.(Property).Value.([]byte), rsha.(Property).Value.([]byte))
	return mapValEq(l.Props(), r.Props(), "sha1")
}

type IdentifierGitTag struct{}

func (i IdentifierGitTag) CanIdentify(data types.Vtx) bool {
	_, ok := data.(vertexGitTag)
	return ok
}

func (i IdentifierGitTag) Matches(a types.Vtx, b types.Vtx) bool {
	l, ok := a.(vertexGitTag)
	if !ok {
		return false
	}
	r, ok := b.(vertexGitTag)
	if !ok {
		return false
	}

	return mapValEq(l.Props(), r.Props(), "name")
}

type IdentifierGitBranch struct{}

func (i IdentifierGitBranch) CanIdentify(data types.Vtx) bool {
	_, ok := data.(vertexGitBranch)
	return ok
}

func (i IdentifierGitBranch) Matches(a types.Vtx, b types.Vtx) bool {
	l, ok := a.(vertexGitBranch)
	if !ok {
		return false
	}
	r, ok := b.(vertexGitBranch)
	if !ok {
		return false
	}

	return mapValEq(l.Props(), r.Props(), "name")
}

type IdentifierTestResult struct{}

func (i IdentifierTestResult) CanIdentify(data types.Vtx) bool {
	_, ok := data.(vertexTestResult)
	return ok
}

func (i IdentifierTestResult) Matches(a types.Vtx, b types.Vtx) bool {
	_, ok := a.(vertexTestResult)
	if !ok {
		return false
	}
	_, ok = b.(vertexTestResult)
	if !ok {
		return false
	}

	return true // TODO LOLOLOL totally demonstrating how this system is broken
}

type IdentifierParentDataset struct{}

func (i IdentifierParentDataset) CanIdentify(data types.Vtx) bool {
	_, ok := data.(vertexParentDataset)
	return ok
}

func (i IdentifierParentDataset) Matches(a types.Vtx, b types.Vtx) bool {
	l, ok := a.(vertexParentDataset)
	if !ok {
		return false
	}
	r, ok := b.(vertexParentDataset)
	if !ok {
		return false
	}

	return mapValEqAnd(l.Props(), r.Props(), "name", "path")
}

type IdentifierYumPkg struct{}

func (i IdentifierYumPkg) CanIdentify(data types.Vtx) bool {
	_, ok := data.(vertexYumPkg)
	return ok
}

func (i IdentifierYumPkg) Matches(a types.Vtx, b types.Vtx) bool {
	l, ok := a.(vertexYumPkg)
	if !ok {
		return false
	}
	r, ok := b.(vertexYumPkg)
	if !ok {
		return false
	}

	// TODO is this actually the correct matching pattern?
	return mapValEq(l.Props(), r.Props(), "name", "version", "arch", "epoch")
}
