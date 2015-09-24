package helpers

import (
	"testing"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/tag1consulting/pipeviz/represent/types"
)

func TestQbv(t *testing.T) {
	// ensure implement both VFilter and EFilter interfaces, and the V chainer
	var _ types.VFilterChain = vertexFilter{}
	var _ types.VEFilter = vertexFilter{}

	assert.Equal(t, Qbv(), vertexFilter{}, "qbv with no args creates an empty vertexFilter")
	assert.Equal(t, Qbv(), vertexFilter{vtype: types.VTypeNone}, "qbv with no args creates equivalent of passing VTypeNone as first arg")
	assert.Equal(t,
		Qbv(types.VType("foo")),
		vertexFilter{vtype: types.VType("foo")},
		"qbv with single arg assigns to VType struct prop")
	assert.Equal(t,
		Qbv(types.VTypeNone, "foo"),
		vertexFilter{vtype: types.VTypeNone},
		"qbv with two args ignores second (unpaired) arg")
	assert.Equal(t,
		Qbv(types.VTypeNone, "foo", "bar"),
		vertexFilter{vtype: types.VTypeNone, props: []types.PropPair{{"foo", "bar"}}},
		"qbv with three args creates one pair of second (key) and third (value) args")
	assert.Equal(t,
		Qbv(types.VTypeNone, "foo", "bar", "baz"),
		vertexFilter{vtype: types.VTypeNone, props: []types.PropPair{{"foo", "bar"}}},
		"qbv with four args creates one pair from 2nd and 3rd args, ignores 4th")

	// ensure that some incorrect things owing to loose typing correctly panic
	assert.Panics(t, func() {
		Qbv("foo")
	}, "qbv panics on type conversion when passing a string instead of VType")

	assert.Panics(t, func() {
		Qbv(types.VTypeNone, 1, "foo")
	}, "qbv panics on type conversion when second argument (with corresponding pair val 3rd arg) is non-string")

	assert.Panics(t, func() {
		Qbv(types.VTypeNone, "foo", "bar", 1, "baz")
	}, "qbv panics on type conversion when Nth even argument (with corresponding pair val N+1 arg) is non-string")
}

func TestQbe(t *testing.T) {
	// ensure implement both VFilter and EFilter interfaces, and the E chainer
	var _ types.EFilterChain = edgeFilter{}
	var _ types.VEFilter = edgeFilter{}

	assert.Equal(t, Qbe(), edgeFilter{}, "qbe with no args creates an empty edgeFilter")
	assert.Equal(t, Qbe(), edgeFilter{etype: types.ETypeNone}, "qbe with no args creates equivalent of passing ETypeNone as first arg")
	assert.Equal(t,
		Qbe(types.EType("foo")),
		edgeFilter{etype: types.EType("foo")},
		"qbe with single arg assigns to EType struct prop")
	assert.Equal(t,
		Qbe(types.ETypeNone, "foo"),
		edgeFilter{etype: types.ETypeNone},
		"qbe with two args ignores second (unpaired) arg")
	assert.Equal(t,
		Qbe(types.ETypeNone, "foo", "bar"),
		edgeFilter{etype: types.ETypeNone, props: []types.PropPair{{"foo", "bar"}}},
		"qbe with three args creates one pair of second (key) and third (value) args")
	assert.Equal(t,
		Qbe(types.ETypeNone, "foo", "bar", "baz"),
		edgeFilter{etype: types.ETypeNone, props: []types.PropPair{{"foo", "bar"}}},
		"qbe with four args creates one pair from 2nd and 3rd args, ignores 4th")

	// ensure that some incorrect things owing to loose typing correctly panic
	assert.Panics(t, func() {
		Qbe("foo")
	}, "qbe panics on type conversion when passing a string instead of EType")

	assert.Panics(t, func() {
		Qbe(types.ETypeNone, 1, "foo")
	}, "qbe panics on type conversion when second argument (with corresponding pair val 3rd arg) is non-string")

	assert.Panics(t, func() {
		Qbe(types.ETypeNone, "foo", "bar", 1, "baz")
	}, "qbe panics on type conversion when Nth even argument (with corresponding pair val N+1 arg) is non-string")
}

func TestBothFilter(t *testing.T) {
	var _ types.VEFilter = bothFilter{}

	assert.Equal(t, Qbe().And(Qbv()), bothFilter{vertexFilter{}, edgeFilter{}}, "Qbe And()'d with Qbv creates an empty bothFilter")
	assert.Equal(t, Qbv().And(Qbe()), bothFilter{vertexFilter{}, edgeFilter{}}, "Qbv And()'d with Qbe creates an empty bothFilter")

	correctv := vertexFilter{vtype: types.VTypeNone, props: []types.PropPair{{"Vfoo", "Vbar"}}}
	correcte := edgeFilter{etype: types.ETypeNone, props: []types.PropPair{{"Efoo", "Ebar"}}}
	correct := bothFilter{correctv, correcte}

	vfirst := Qbv(types.VTypeNone, "Vfoo", "Vbar").And(Qbe(types.ETypeNone, "Efoo", "Ebar"))
	efirst := Qbe(types.ETypeNone, "Efoo", "Ebar").And(Qbv(types.VTypeNone, "Vfoo", "Vbar"))

	assert.Equal(t, vfirst, correct, "Qbv And()d with Qbe creates a correct bothFilter struct instance")
	assert.Equal(t, efirst, correct, "Qbe And()d with Qbv creates a correct bothFilter struct instance")

	assert.Equal(t, correct.VType(), correctv.vtype, "correct bothFilter reports correct VType")
	assert.Equal(t, correct.EType(), correcte.etype, "correct bothFilter reports correct EType")
	assert.Equal(t, correct.VProps(), correctv.props, "correct BothFilter reports correct vertex props")
	assert.Equal(t, correct.EProps(), correcte.props, "correct BothFilter reports correct edge props")
}
