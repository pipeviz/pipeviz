package represent

import (
	"fmt"

	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
)

func Identify(g CoreGraph, sd SplitData) int {
	ids := identifyDefault(g, sd)
	if ids == nil {
		return 0
	} else if len(ids) == 1 {
		return ids[0]
	}

	// default one didn't do it, use specialists to narrow further

	switch sd.Vertex.(type) {
	default:
		panic("ZOMG CAN'T IDENTIFY")
	case vertexTestResult:
		return identifyByGitHashSpec(g, sd, ids)
	}
}

func identifyDefault(g CoreGraph, sd SplitData) []int {
	matches := g.VerticesWith(qbv(sd.Vertex.Typ()))
	if len(matches) == 0 {
		// no vertices of this type, safe to bail early
		return nil
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
		panic(fmt.Sprintf("missing identify checker for type %T", sd.Vertex))
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
		newvt := vtTuple{v: sd.Vertex, ie: ps.NewMap(), oe: ps.NewMap()}
		edge, success := Resolve(g, 0, newvt, envlink)

		if !success {
			// FIXME failure to resolve envlink doesn't necessarily mean no match
			return nil
		}

		for _, candidate := range filtered {
			for _, edge2 := range g.OutWith(candidate.id, qbe(EType("envlink"))) {
				filtered2 = append(filtered2, candidate)
				if edge2.Target == edge.Target {
					return []int{candidate.id}
				}
			}
		}
	} else {
		// no el, suggesting our candidate list is good
		filtered2 = filtered
	}

	var ret []int
	for _, vt := range filtered2 {
		ret = append(ret, vt.id)
	}

	return ret
}

// Narrow a match list by looking for alignment on a git commit sha1
func identifyByGitHashSpec(g CoreGraph, sd SplitData, matches []int) int {
	for _, es := range sd.EdgeSpecs {
		// first find the commit spec
		if spec, ok := es.(SpecCommit); ok {
			// then search otherwise-matching vertices for a corresponding sha1 edge
			for _, matchvid := range matches {
				if len(g.OutWith(matchvid, qbe(EType("version"), "sha1", spec.Sha1))) == 1 {
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
		IdentifierEnvironment{},
		IdentifierLogicState{},
		IdentifierDataset{},
		IdentifierProcess{},
		IdentifierCommit{},
		IdentifierVcsLabel{},
		IdentifierTestResult{},
		IdentifierParentDataset{},
		IdentifierComm{},
	}
}

// Identifiers represent the logic for identifying specific types of objects
// that may be contained within the graph, and finding matches between these
// types of objects
type Identifier interface {
	CanIdentify(data Vertex) bool
	Matches(a Vertex, b Vertex) bool
}

// Identifier for Environments
type IdentifierEnvironment struct{}

func (i IdentifierEnvironment) CanIdentify(data Vertex) bool {
	_, ok := data.(vertexEnvironment)
	return ok
}

func (i IdentifierEnvironment) Matches(a Vertex, b Vertex) bool {
	l, ok := a.(vertexEnvironment)
	if !ok {
		return false
	}
	r, ok := b.(vertexEnvironment)
	if !ok {
		return false
	}

	return matchAddress(l.Props(), r.Props())
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

type IdentifierLogicState struct{}

func (i IdentifierLogicState) CanIdentify(data Vertex) bool {
	_, ok := data.(vertexLogicState)
	return ok
}

func (i IdentifierLogicState) Matches(a Vertex, b Vertex) bool {
	l, ok := a.(vertexLogicState)
	if !ok {
		return false
	}
	r, ok := b.(vertexLogicState)
	if !ok {
		return false
	}

	return mapValEq(l.Props(), r.Props(), "path")
}

type IdentifierDataset struct{}

func (i IdentifierDataset) CanIdentify(data Vertex) bool {
	_, ok := data.(vertexDataset)
	return ok
}

func (i IdentifierDataset) Matches(a Vertex, b Vertex) bool {
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

func (i IdentifierCommit) CanIdentify(data Vertex) bool {
	_, ok := data.(vertexCommit)
	return ok
}

func (i IdentifierCommit) Matches(a Vertex, b Vertex) bool {
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

type IdentifierProcess struct{}

func (i IdentifierProcess) CanIdentify(data Vertex) bool {
	_, ok := data.(vertexProcess)
	return ok
}

func (i IdentifierProcess) Matches(a Vertex, b Vertex) bool {
	l, ok := a.(vertexProcess)
	if !ok {
		return false
	}
	r, ok := b.(vertexProcess)
	if !ok {
		return false
	}

	// TODO numeric id within the 2^16 ring buffer that is pids is a horrible way to do this
	return mapValEq(l.Props(), r.Props(), "pid")
}

type IdentifierVcsLabel struct{}

func (i IdentifierVcsLabel) CanIdentify(data Vertex) bool {
	_, ok := data.(vertexVcsLabel)
	return ok
}

func (i IdentifierVcsLabel) Matches(a Vertex, b Vertex) bool {
	l, ok := a.(vertexVcsLabel)
	if !ok {
		return false
	}
	r, ok := b.(vertexVcsLabel)
	if !ok {
		return false
	}

	return mapValEq(l.Props(), r.Props(), "name")
}

type IdentifierTestResult struct{}

func (i IdentifierTestResult) CanIdentify(data Vertex) bool {
	_, ok := data.(vertexTestResult)
	return ok
}

func (i IdentifierTestResult) Matches(a Vertex, b Vertex) bool {
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

func (i IdentifierParentDataset) CanIdentify(data Vertex) bool {
	_, ok := data.(vertexParentDataset)
	return ok
}

func (i IdentifierParentDataset) Matches(a Vertex, b Vertex) bool {
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

type IdentifierComm struct{}

func (i IdentifierComm) CanIdentify(data Vertex) bool {
	_, ok := data.(vertexComm)
	return ok
}

func (i IdentifierComm) Matches(a Vertex, b Vertex) bool {
	l, ok := a.(vertexComm)
	if !ok {
		return false
	}
	r, ok := b.(vertexComm)
	if !ok {
		return false
	}

	return mapValEqAnd(l.Props(), r.Props(), "name", "path")
}
