package represent

import (
	"bytes"

	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
)

// Performs simple equality comparison on the provided key between two persistent maps.
// CRUCIAL NOTE - returns false if either OR both keys do not exist.
func mapValEq(key string, l, r ps.Map) bool {
	lv, exists := l.Lookup(key)
	if !exists {
		return false
	}

	rv, exists := r.Lookup(key)
	if !exists {
		return false
	}

	switch tlv := lv.(type) {
	default:
		return rv == lv
	case []byte:
		trv, ok := rv.([]byte)
		return ok && bytes.Equal(tlv, trv)
	}
}

func envLinkEqProps(e interpret.EnvLink, m ps.Map) bool {
	nick, exists := m.Lookup("nick")
	if (!exists && e.Nick != "") || (exists && e.Nick != nick) {
		return false
	}

	return addressEqProps(e.Address, m)
}

func addressEqProps(a interpret.Address, m ps.Map) bool {
	hostname, exists := m.Lookup("hostname")
	if (!exists && a.Hostname != "") || (exists && a.Hostname != hostname) {
		return false
	}

	ipv4, exists := m.Lookup("ipv4")
	if (!exists && a.Ipv4 != "") || (exists && a.Ipv4 != ipv4) {
		return false
	}

	ipv6, exists := m.Lookup("ipv6")
	if (!exists && a.Ipv6 != "") || (exists && a.Ipv6 != ipv6) {
		return false
	}

	return true
}
