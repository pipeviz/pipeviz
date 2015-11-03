package q

import (
	"testing"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/pipeviz/pipeviz/types/system"
)

func TestQbv(t *testing.T) {
	// ensure implement both VFilter and EFilter interfaces, and the V chainer
	var _ system.VFilterChain = vertexFilter{}
	var _ system.VEFilter = vertexFilter{}

	assert.Equal(t, Qbv(), vertexFilter{}, "qbv with no args creates an empty vertexFilter")
	assert.Equal(t, Qbv(), vertexFilter{vtype: system.VTypeNone}, "qbv with no args creates equivalent of passing VTypeNone as first arg")
	assert.Equal(t,
		Qbv(system.VType("foo")),
		vertexFilter{vtype: system.VType("foo")},
		"qbv with single arg assigns to VType struct prop")
	assert.Equal(t,
		Qbv(system.VTypeNone, "foo"),
		vertexFilter{vtype: system.VTypeNone},
		"qbv with two args ignores second (unpaired) arg")
	assert.Equal(t,
		Qbv(system.VTypeNone, "foo", "bar"),
		vertexFilter{vtype: system.VTypeNone, props: []system.PropPair{{"foo", "bar"}}},
		"qbv with three args creates one pair of second (key) and third (value) args")
	assert.Equal(t,
		Qbv(system.VTypeNone, "foo", "bar", "baz"),
		vertexFilter{vtype: system.VTypeNone, props: []system.PropPair{{"foo", "bar"}}},
		"qbv with four args creates one pair from 2nd and 3rd args, ignores 4th")

	// ensure that some incorrect things owing to loose typing correctly panic
	assert.Panics(t, func() {
		Qbv("foo")
	}, "qbv panics on type conversion when passing a string instead of VType")

	assert.Panics(t, func() {
		Qbv(system.VTypeNone, 1, "foo")
	}, "qbv panics on type conversion when second argument (with corresponding pair val 3rd arg) is non-string")

	assert.Panics(t, func() {
		Qbv(system.VTypeNone, "foo", "bar", 1, "baz")
	}, "qbv panics on type conversion when Nth even argument (with corresponding pair val N+1 arg) is non-string")
}

func TestQbe(t *testing.T) {
	// ensure implement both VFilter and EFilter interfaces, and the E chainer
	var _ system.EFilterChain = edgeFilter{}
	var _ system.VEFilter = edgeFilter{}

	assert.Equal(t, Qbe(), edgeFilter{}, "qbe with no args creates an empty edgeFilter")
	assert.Equal(t, Qbe(), edgeFilter{etype: system.ETypeNone}, "qbe with no args creates equivalent of passing ETypeNone as first arg")
	assert.Equal(t,
		Qbe(system.EType("foo")),
		edgeFilter{etype: system.EType("foo")},
		"qbe with single arg assigns to EType struct prop")
	assert.Equal(t,
		Qbe(system.ETypeNone, "foo"),
		edgeFilter{etype: system.ETypeNone},
		"qbe with two args ignores second (unpaired) arg")
	assert.Equal(t,
		Qbe(system.ETypeNone, "foo", "bar"),
		edgeFilter{etype: system.ETypeNone, props: []system.PropPair{{"foo", "bar"}}},
		"qbe with three args creates one pair of second (key) and third (value) args")
	assert.Equal(t,
		Qbe(system.ETypeNone, "foo", "bar", "baz"),
		edgeFilter{etype: system.ETypeNone, props: []system.PropPair{{"foo", "bar"}}},
		"qbe with four args creates one pair from 2nd and 3rd args, ignores 4th")

	// ensure that some incorrect things owing to loose typing correctly panic
	assert.Panics(t, func() {
		Qbe("foo")
	}, "qbe panics on type conversion when passing a string instead of EType")

	assert.Panics(t, func() {
		Qbe(system.ETypeNone, 1, "foo")
	}, "qbe panics on type conversion when second argument (with corresponding pair val 3rd arg) is non-string")

	assert.Panics(t, func() {
		Qbe(system.ETypeNone, "foo", "bar", 1, "baz")
	}, "qbe panics on type conversion when Nth even argument (with corresponding pair val N+1 arg) is non-string")
}

func TestBothFilter(t *testing.T) {
	var _ system.VEFilter = bothFilter{}

	assert.Equal(t, Qbe().And(Qbv()), bothFilter{vertexFilter{}, edgeFilter{}}, "Qbe And()'d with Qbv creates an empty bothFilter")
	assert.Equal(t, Qbv().And(Qbe()), bothFilter{vertexFilter{}, edgeFilter{}}, "Qbv And()'d with Qbe creates an empty bothFilter")

	correctv := vertexFilter{vtype: system.VTypeNone, props: []system.PropPair{{"Vfoo", "Vbar"}}}
	correcte := edgeFilter{etype: system.ETypeNone, props: []system.PropPair{{"Efoo", "Ebar"}}}
	correct := bothFilter{correctv, correcte}

	vfirst := Qbv(system.VTypeNone, "Vfoo", "Vbar").And(Qbe(system.ETypeNone, "Efoo", "Ebar"))
	efirst := Qbe(system.ETypeNone, "Efoo", "Ebar").And(Qbv(system.VTypeNone, "Vfoo", "Vbar"))

	assert.Equal(t, vfirst, correct, "Qbv And()d with Qbe creates a correct bothFilter struct instance")
	assert.Equal(t, efirst, correct, "Qbe And()d with Qbv creates a correct bothFilter struct instance")

	assert.Equal(t, correct.VType(), correctv.vtype, "correct bothFilter reports correct VType")
	assert.Equal(t, correct.EType(), correcte.etype, "correct bothFilter reports correct EType")
	assert.Equal(t, correct.VProps(), correctv.props, "correct BothFilter reports correct vertex props")
	assert.Equal(t, correct.EProps(), correcte.props, "correct BothFilter reports correct edge props")
}
