package represent

import (
	"bytes"

	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
)

func Identify(g *CoreGraph, sd SplitData) int {
	switch v := sd.Vertex.(type) {
	// TODO special casing here...
	default:
		ids := identifyDefault(g, sd)
		if ids == nil {
			return 0
		} else if len(ids) == 1 {
			return ids[0]
		} else {
			panic("how can have more than one w/out env namespacer?")
		}
	}
}

func identifyDefault(g *CoreGraph, sd SplitData) []int {
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
		panic("missing identify checker")
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

	if hasEl {
		newvt := vtTuple{v: sd.Vertex, ie: ps.NewMap(), oe: ps.NewMap()}
		edge, success := Resolve(g, 0, newvt, envlink)

		if !success {
			// FIXME failure to resolve envlink doesn't necessarily mean no match
			return nil
		}

		var match bool
		for _, candidate := range filtered {
			for _, edge2 := range g.OutWith(candidate.id, qbe("envlink")) {
				if edge2.Target == edge.Target {
					return []int{candidate.id}
				}
			}
		}
	}

	var ret []int
	for _, vt := range filtered {
		ret = append(ret, vt.id)
	}

	return ret
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

	if !mapValEq(l.Props(), r.Props(), "path") {
		return false
	}

	// Path matches; env has to match, too.
	// TODO matching like this assumes that envlinks are always directly resolved, with no bounding context
	return matchEnvLink(l.Props(), r.Props())
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

	if !mapValEq(l.Props(), r.Props(), "name") {
		return false
	}

	// Name matches; env has to match, too.
	// TODO matching like this assumes that envlinks are always directly resolved, with no bounding context
	return matchEnvLink(l.Props(), r.Props())
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

	lsha, lexists := l.Props().Lookup("sha1")
	rsha, rexists := r.Props().Lookup("sha1")
	return rexists && lexists && bytes.Equal(lsha.([]byte), rsha.([]byte))
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
	return mapValEq(l.Props(), r.Props(), "pid") && matchEnvLink(l.Props(), r.Props())
}
