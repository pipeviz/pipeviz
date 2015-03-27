package interpret

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

	// Current logic: match if *any* non-empty of hostname, ipv4, or ipv6 match
	if l.Address.Hostname == r.Address.Hostname && r.Address.Hostname != "" {
		return true
	}
	if l.Address.Ipv4 == r.Address.Ipv4 && r.Address.Ipv4 != "" {
		return true
	}
	if l.Address.Ipv6 == r.Address.Ipv6 && r.Address.Ipv6 != "" {
		return true
	}

	return false
}
