package maputil_test

import (
	"testing"

	"github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/maputil"
	"github.com/pipeviz/pipeviz/types/system"
)

func TestAnyMatch(t *testing.T) {
	l, r := ps.NewMap(), ps.NewMap()

	l = l.Set("foo", 1)
	r = r.Set("foo", 1)

	l = l.Set("bar", 2)
	r = r.Set("bar", 3)

	l = l.Set("baz", 4)
	r = r.Set("quux", nil)

	// zero inputs, must be false. this is basically testing the identity func
	if maputil.AnyMatch(l, r) {
		t.Error("No inputs, must be no match")
	}

	if !maputil.AnyMatch(l, r, "foo") {
		t.Error("Should have matched on 'foo' key's value")
	}

	if !maputil.AnyMatch(l, r, "foo", "bar") {
		t.Error("Should have matched on 'foo' key's value")
	}

	// test that order has no effect
	if !maputil.AnyMatch(l, r, "bar", "foo") {
		t.Error("Should have matched on 'foo' key's value")
	}

	if maputil.AnyMatch(l, r, "baz", "quux") {
		t.Error("Should not have found a match when no keys present in both maps")
	}

	if !maputil.AnyMatch(l, r, "baz", "foo") {
		t.Error("Should have found a match even if there's some key asymmetry")
	}

	if !maputil.AnyMatch(l.Set("foo", system.Property{MsgSrc: 0, Value: 1}), r, "foo", "bar") {
		t.Error("Should have matched thanks to auto-transform of Property")
	}

	if !maputil.AnyMatch(l, r.Set("foo", system.Property{MsgSrc: 0, Value: 1}), "foo", "bar") {
		t.Error("Property transform works on left or right")
	}

	if !maputil.AnyMatch(l.Set("foo", system.Property{MsgSrc: 0, Value: 1}), r.Set("foo", system.Property{MsgSrc: 0, Value: 1}), "foo", "bar") {
		t.Error("Property eq works when both are Properties")
	}
}

func TestAllMatch(t *testing.T) {
	l, r := ps.NewMap(), ps.NewMap()

	l = l.Set("foo", 1)
	r = r.Set("foo", 1)

	l = l.Set("bar", 2)
	r = r.Set("bar", 3)

	l = l.Set("baz", 4)
	r = r.Set("quux", nil)

	// zero inputs, must be false. this is basically testing the identity func
	if maputil.AllMatch(l, r) {
		t.Error("No inputs, must be no match")
	}

	if !maputil.AllMatch(l, r, "foo") {
		t.Error("Should have matched on 'foo' key's value")
	}

	if !maputil.AllMatch(l, r.Set("baz", system.Property{MsgSrc: 0, Value: 4}), "foo", "baz") {
		t.Error("Should have matched on 'foo' key's value")
	}

	if maputil.AllMatch(l, r, "notexist") {
		t.Error("Should not match if only nonexistent key is passed in")
	}

	if maputil.AllMatch(l, r, "notexist2") {
		t.Error("Should not match if only nonexistent keys are passed in")
	}

	if maputil.AllMatch(l, r, "baz", "quux") {
		t.Error("Should not have found a match when there is any key asymmetry")
	}

	if !maputil.AllMatch(l.Set("baz", system.Property{MsgSrc: 0, Value: 1}), r.Set("baz", system.Property{MsgSrc: 0, Value: 1}), "foo", "baz") {
		t.Error("Property eq works when both are Properties")
	}
}
