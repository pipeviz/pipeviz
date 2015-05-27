package represent

import (
	"bytes"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/interpret"
)

// Performs simple equality comparison on the provided keys between two persistent maps.
// CRUCIAL NOTE - returns false if either OR both keys do not exist.
func mapValEq(l, r ps.Map, keys ...string) bool {
	for _, key := range keys {
		lv, exists := l.Lookup(key)
		if !exists {
			return false
		}

		rv, exists := r.Lookup(key)
		if !exists {
			return false
		}

		// Transparently convert properties into their values
		if lpv, ok := lv.(Property); ok {
			lv = lpv.Value
			if rpv, ok := rv.(Property); ok {
				rv = rpv.Value
			} else {
				return false
			}
		}

		switch tlv := lv.(type) {
		default:
			if rv != lv {
				return false
			}
		case []byte:
			trv, ok := rv.([]byte)
			if ok && bytes.Equal(tlv, trv) { // for readability, instead of ! || !
				continue
			}
			return false
		}
	}

	return true
}

// Performs equality comparison on the provided keys between two persistent maps.
// Returns true IFF all values are equal. Unlike mapValEq, both keys not existing
// *is* considered equality.
func mapValEqAnd(l, r ps.Map, keys ...string) bool {
	for _, key := range keys {
		lv, lexists := l.Lookup(key)
		rv, rexists := r.Lookup(key)

		// bail out if one exists and the other doesn't
		if lexists != rexists {
			return false
		}
		// if neither exist, skip
		if !lexists {
			continue
		}

		// Transparently convert properties into their values
		if lpv, ok := lv.(Property); ok {
			lv = lpv.Value
			if rpv, ok := rv.(Property); ok {
				rv = rpv.Value
			} else {
				return false
			}
		}

		switch tlv := lv.(type) {
		default:
			if rv != lv {
				return false
			}
		case []byte:
			trv, ok := rv.([]byte)
			if ok && bytes.Equal(tlv, trv) { // for readability, instead of ! || !
				continue
			}
			return false
		}
	}

	return true
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
