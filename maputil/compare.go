package maputil

import (
	"bytes"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/types/system"
)

// AnyMatch performs simple equality comparisons between two maps on
// the values associated with each of the provided keys. False is returned if there
// are no matches, true if there are any matches. One or both of the maps missing a
// particular key is considered a non-match.
//
// The expectation is generally that values will be of types.Property. In this
// case, it is the Property.Value that is compared, not the whole Property.
// However, if a non-Property value is found at a provided key, then it is used
// directly in the equality check.
func AnyMatch(l, r ps.Map, keys ...string) bool {
	var lv, rv interface{}
	var lexists, rexists bool

	lm, lrm := l.(system.RawProps)
	rm, rrm := r.(system.RawProps)

	for _, key := range keys {
		if lrm {
			lv, lexists = lm[key]
		} else {
			lv, lexists = l.Lookup(key)
		}

		if rrm {
			rv, rexists = rm[key]
		} else {
			rv, rexists = r.Lookup(key)
		}

		// skip if either or both don't exist
		if !lexists || !rexists {
			continue
		}

		// Transparently convert properties into their values
		if lpv, ok := lv.(system.Property); ok {
			lv = lpv.Value
		}
		if rpv, ok := rv.(system.Property); ok {
			rv = rpv.Value
		}

		switch tlv := lv.(type) {
		default:
			if rv == lv {
				return true
			}
		case []byte:
			trv, ok := rv.([]byte)
			if ok && bytes.Equal(tlv, trv) {
				return true
			}
		}
	}

	return false
}

// AllMatch performs simple equality comparisons between two maps on
// the values associated with each of the provided keys. True is returned if
// all of the values are equal (including the case where both maps are missing
// a particular key), and false otherwise.
//
// Properties values are transformed into their contained Value in the
// same manner as in AnyMatch.
func AllMatch(l, r ps.Map, keys ...string) bool {
	// keep track of our both-missing count
	var nep int

	var lv, rv interface{}
	var lexists, rexists bool

	lm, lrm := l.(system.RawProps)
	rm, rrm := r.(system.RawProps)

	for _, key := range keys {
		if lrm {
			lv, lexists = lm[key]
		} else {
			lv, lexists = l.Lookup(key)
		}

		if rrm {
			rv, rexists = rm[key]
		} else {
			rv, rexists = r.Lookup(key)
		}

		// bail out if one exists and the other doesn't
		if lexists != rexists {
			return false
		}
		// if neither exist, skip
		if !lexists {
			nep++
			continue
		}

		// Transparently convert properties into their values
		if lpv, ok := lv.(system.Property); ok {
			lv = lpv.Value
		}
		if rpv, ok := rv.(system.Property); ok {
			rv = rpv.Value
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

	return len(keys) > nep
}
