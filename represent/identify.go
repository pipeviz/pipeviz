package represent

import (
	"bytes"

	"github.com/mndrix/ps"
)

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
	_, ok := data.(environmentVertex)
	return ok
}

func (i IdentifierEnvironment) Matches(a Vertex, b Vertex) bool {
	l, ok := a.(environmentVertex)
	if !ok {
		return false
	}
	r, ok := b.(environmentVertex)
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
	_, ok := data.(logicStateVertex)
	return ok
}

func (i IdentifierLogicState) Matches(a Vertex, b Vertex) bool {
	l, ok := a.(logicStateVertex)
	if !ok {
		return false
	}
	r, ok := b.(logicStateVertex)
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
	_, ok := data.(datasetVertex)
	return ok
}

func (i IdentifierDataset) Matches(a Vertex, b Vertex) bool {
	l, ok := a.(datasetVertex)
	if !ok {
		return false
	}
	r, ok := b.(datasetVertex)
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
	_, ok := data.(commitVertex)
	return ok
}

func (i IdentifierCommit) Matches(a Vertex, b Vertex) bool {
	l, ok := a.(commitVertex)
	if !ok {
		return false
	}
	r, ok := b.(commitVertex)
	if !ok {
		return false
	}

	lsha, lexists := l.Props().Lookup("sha1")
	rsha, rexists := r.Props().Lookup("sha1")
	return rexists && lexists && bytes.Equal(lsha.([]byte), rsha.([]byte))
}

type IdentifierProcess struct{}

func (i IdentifierProcess) CanIdentify(data Vertex) bool {
	_, ok := data.(processVertex)
	return ok
}

func (i IdentifierProcess) Matches(a Vertex, b Vertex) bool {
	l, ok := a.(processVertex)
	if !ok {
		return false
	}
	r, ok := b.(processVertex)
	if !ok {
		return false
	}

	// TODO numeric id within the 2^16 ring buffer that is pids is a horrible way to do this
	return mapValEq(l.Props(), r.Props(), "pid") && matchEnvLink(l.Props(), r.Props())
}
