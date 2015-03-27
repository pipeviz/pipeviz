package interpret

import (
	"bytes"
)

// Identifiers represent the logic for identifying specific types of objects
// that may be contained within the graph, and finding matches between these
// types of objects
type Identifier interface {
	CanIdentify(data interface{}) bool
	Matches(a interface{}, b interface{}) bool
}

// Identifier for Environments
type IdentifierEnvironment struct{}

func (i IdentifierEnvironment) CanIdentify(data interface{}) bool {
	_, ok := data.(Environment)
	return ok
}

func (i IdentifierEnvironment) Matches(a interface{}, b interface{}) bool {
	l, ok := a.(Environment)
	if !ok {
		return false
	}
	r, ok := b.(Environment)
	if !ok {
		return false
	}

	return matchAddress(l.Address, r.Address)
}

// Helper func to match addresses
func matchAddress(a Address, b Address) bool {
	// For now, match if *any* non-empty of hostname, ipv4, or ipv6 match
	// TODO this needs moar thinksies
	if a.Hostname == b.Hostname && b.Hostname != "" {
		return true
	}
	if a.Ipv4 == b.Ipv4 && b.Ipv4 != "" {
		return true
	}
	if a.Ipv6 == b.Ipv6 && b.Ipv6 != "" {
		return true
	}

	return false
}

// Helper func to match env links
func matchEnvLink(a EnvLink, b EnvLink) bool {
	return matchAddress(a.Address, b.Address) || a.Nick == b.Nick
}

type IdentifierLogicState struct{}

func (i IdentifierLogicState) CanIdentify(data interface{}) bool {
	_, ok := data.(LogicState)
	return ok
}

func (i IdentifierLogicState) Matches(a interface{}, b interface{}) bool {
	l, ok := a.(LogicState)
	if !ok {
		return false
	}
	r, ok := b.(LogicState)
	if !ok {
		return false
	}

	if l.Path != r.Path {
		return false
	}

	// Path matches; env has to match, too.
	// TODO matching like this assumes that envlinks are always directly resolved, with no bounding context
	return matchEnvLink(l.Environment, r.Environment)
}

type IdentifierDataset struct{}

func (i IdentifierDataset) CanIdentify(data interface{}) bool {
	_, ok := data.(Dataset)
	return ok
}

func (i IdentifierDataset) Matches(a interface{}, b interface{}) bool {
	l, ok := a.(Dataset)
	if !ok {
		return false
	}
	r, ok := b.(Dataset)
	if !ok {
		return false
	}

	if l.Name != r.Name {
		return false
	}

	// Name matches; env has to match, too.
	// TODO matching like this assumes that envlinks are always directly resolved, with no bounding context
	return matchEnvLink(l.Environment, r.Environment)
}

type IdentifierCommit struct{}

func (i IdentifierCommit) CanIdentify(data interface{}) bool {
	_, ok := data.(Commit)
	return ok
}

func (i IdentifierCommit) Matches(a interface{}, b interface{}) bool {
	l, ok := a.(Commit)
	if !ok {
		return false
	}
	r, ok := b.(Commit)
	if !ok {
		return false
	}

	return bytes.Equal(l.Sha1, r.Sha1)
}
