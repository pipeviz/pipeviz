package represent

import (
	"strconv"
	"testing"

	"github.com/mndrix/ps"
	"github.com/stretchr/testify/assert"
)

func getGraphFixture() *CoreGraph {
	g := &CoreGraph{vtuples: ps.NewMap(), vserial: 0}
	var vt vtTuple

	// Manually populate the graph with a couple envs and logic states
	// env #1
	sd, _ := Split(F_Environment[0].Input, 1)
	vt = vtTuple{
		id: 1,
		v:  sd[0].Vertex,
		ie: ps.NewMap(),
		oe: ps.NewMap(),
	}
	g.vtuples = g.vtuples.Set(strconv.Itoa(1), vt)

	// env #2
	sd, _ = Split(F_Environment[1].Input, 2)
	// create the edge w/map that will connect to this env
	edge := StandardEdge{
		id:     4,
		Source: 3,
		Target: 2,
		EType:  EType("envlink"),
		Props:  mapPropPairs(3, p{"ipv4", D_ipv4}),
	}

	vt = vtTuple{
		id: 2,
		v:  sd[0].Vertex,
		ie: ps.NewMap(),
		oe: ps.NewMap(),
	}
	vt.ie = vt.ie.Set(strconv.Itoa(4), edge)

	g.vtuples = g.vtuples.Set(strconv.Itoa(2), vt)

	// and the logic state
	sd, _ = Split(F_LogicState[1].Input, 3)

	vt = vtTuple{
		id: 3,
		v:  sd[0].Vertex,
		ie: ps.NewMap(),
		oe: ps.NewMap(),
	}

	vt.oe = vt.oe.Set(strconv.Itoa(4), edge)

	g.vtuples = g.vtuples.Set(strconv.Itoa(3), vt)
	g.vserial = 4

	return g
}

func TestQbv(t *testing.T) {
	// ensure implement both VFilter and EFilter interfaces
	var _ VEFilter = vertexFilter{}

	assert.Equal(t, qbv(), vertexFilter{}, "qbv with no args creates an empty vertexFilter")
	assert.Equal(t, qbv(), vertexFilter{vtype: VTypeNone}, "qbv with no args creates equivalent of passing VTypeNone as first arg")
	assert.Equal(t, qbv(VType("foo")), vertexFilter{vtype: VType("foo")}, "qbv with single arg assigns to VType struct prop")
	assert.Equal(t, qbv(VTypeNone, "foo"), vertexFilter{vtype: VTypeNone}, "qbv with two args ignores second (unpaired) arg")
	assert.Equal(t, qbv(VTypeNone, "foo", "bar"), vertexFilter{vtype: VTypeNone, props: []PropQ{{"foo", "bar"}}}, "qbv with three args creates one pair of second (key) and third (value) args")
	assert.Equal(t, qbv(VTypeNone, "foo", "bar", "baz"), vertexFilter{vtype: VTypeNone, props: []PropQ{{"foo", "bar"}}}, "qbv with four args creates one pair from 2nd and 3rd args, ignores 4th")

	// ensure that some incorrect things owing to loose typing correctly panic
	assert.Panics(t, func() {
		qbv("foo")
	}, "qbv panics on type conversion when passing a string instead of VType")

	assert.Panics(t, func() {
		qbv(VTypeNone, 1, "foo")
	}, "qbv panics on type conversion when second argument (with corresponding pair val 3rd arg) is non-string")

	assert.Panics(t, func() {
		qbv(VTypeNone, "foo", "bar", 1, "baz")
	}, "qbv panics on type conversion when Nth even argument (with corresponding pair val N+1 arg) is non-string")
}

func TestQbe(t *testing.T) {
	// ensure implement both VFilter and EFilter interfaces
	var _ VEFilter = edgeFilter{}

	assert.Equal(t, qbe(), edgeFilter{}, "qbe with no args creates an empty edgeFilter")
	assert.Equal(t, qbe(), edgeFilter{etype: ETypeNone}, "qbe with no args creates equivalent of passing ETypeNone as first arg")
	assert.Equal(t, qbe(EType("foo")), edgeFilter{etype: EType("foo")}, "qbe with single arg assigns to EType struct prop")
	assert.Equal(t, qbe(ETypeNone, "foo"), edgeFilter{etype: ETypeNone}, "qbe with two args ignores second (unpaired) arg")
	assert.Equal(t, qbe(ETypeNone, "foo", "bar"), edgeFilter{etype: ETypeNone, props: []PropQ{{"foo", "bar"}}}, "qbe with three args creates one pair of second (key) and third (value) args")
	assert.Equal(t, qbe(ETypeNone, "foo", "bar", "baz"), edgeFilter{etype: ETypeNone, props: []PropQ{{"foo", "bar"}}}, "qbe with four args creates one pair from 2nd and 3rd args, ignores 4th")

	// ensure that some incorrect things owing to loose typing correctly panic
	assert.Panics(t, func() {
		qbe("foo")
	}, "qbe panics on type conversion when passing a string instead of EType")

	assert.Panics(t, func() {
		qbe(ETypeNone, 1, "foo")
	}, "qbe panics on type conversion when second argument (with corresponding pair val 3rd arg) is non-string")

	assert.Panics(t, func() {
		qbe(ETypeNone, "foo", "bar", 1, "baz")
	}, "qbe panics on type conversion when Nth even argument (with corresponding pair val N+1 arg) is non-string")
}

func TestVerticesWith(t *testing.T) {
	g := getGraphFixture()
	var result []vtTuple

	result = g.VerticesWith(qbv())
	if len(result) != 3 {
		t.Errorf("Should find 3 vertices with no filter; found %v", len(result))
	}

	result = g.VerticesWith(qbv(VType("environment")))
	if len(result) != 2 {
		t.Errorf("Should find 2 vertices when filtering to type env; found %v", len(result))
	}

	result = g.VerticesWith(qbv(VType("commit")))
	if len(result) != 0 {
		t.Errorf("Should find no vertices when filtering to type env; found %v", len(result))
	}

	result = g.VerticesWith(qbv(VTypeNone, "ipv4", D_ipv4))
	if len(result) != 1 {
		t.Errorf("Should find one vertex when filtering on ipv4 prop; found %v", len(result))
	}

	result = g.VerticesWith(qbv(VTypeNone, "ipv4", D_ipv6))
	if len(result) != 0 {
		t.Errorf("Should find no vertices when filtering on bad value for ipv4 prop; found %v", len(result))
	}

	result = g.VerticesWith(qbv(VType("environment"), "ipv4", D_ipv4))
	if len(result) != 1 {
		t.Errorf("Should find one vertex when filtering to env types and on ipv4 prop; found %v", len(result))
	}
}

// Tests arcWith(), which effectively tests both OutWith() and InWith().
func TestOutInArcWith(t *testing.T) {
	g := getGraphFixture()
	var result []StandardEdge

	result = g.arcWith(1, qbe(), false)
	if len(result) != 0 {
		t.Errorf("Vertex 1 has no edges at all, but still got %v out-edge results", len(result))
	}

	result = g.arcWith(1, qbe(), true)
	if len(result) != 0 {
		t.Errorf("Vertex 1 has no edges at all, but still got %v in-edge results", len(result))
	}

	result = g.arcWith(2, qbe(), true)
	if len(result) != 1 {
		t.Errorf("Vertex 2 should have one in-edge, but got %v edges", len(result))
	}

	result = g.InWith(2, qbe())
	if len(result) != 1 {
		t.Errorf("Vertex 2 should have one in-edge, but got %v edges (InWith calls arcWith correctly)", len(result))
	}

	result = g.arcWith(2, qbe(), false)
	if len(result) != 0 {
		t.Errorf("Vertex 2 has an in-edge, but no out-edge; still got %v in-edge results", len(result))
	}

	result = g.arcWith(3, qbe(), false)
	if len(result) != 1 {
		t.Errorf("Vertex 3 should have one out-edge, but got %v edges", len(result))
	}

	result = g.arcWith(3, qbe(ETypeNone), false)
	if len(result) != 1 {
		t.Errorf("ETypeNone does not correctly matches all edge types - should've gotten 1 out-edge, but got %v edges", len(result))
	}

	result = g.OutWith(3, qbe())
	if len(result) != 1 {
		t.Errorf("Vertex 3 should have one out-edge, but got %v edges (OutWith calls arcWith correctly)", len(result))
	}

	result = g.InWith(3, qbe())
	if len(result) != 0 {
		t.Errorf("Vertex 3 has an out-edge, but no in-edge; still got %v in-edge results", len(result))
	}

	result = g.InWith(2, qbe(EType("envlink")))
	if len(result) != 1 {
		t.Errorf("Vertex 2 should have one 'envlink'-typed in-edge, but got %v edges", len(result))
	}

	result = g.InWith(2, qbe(EType("not-envlink")))
	if len(result) != 0 {
		t.Errorf("Vertex 2 should have no 'not-envlink'-typed in-edges, but got %v edges", len(result))
	}

	result = g.OutWith(3, qbe(EType("envlink")))
	if len(result) != 1 {
		t.Errorf("Vertex 3 should have one 'envlink'-typed out-edge, but got %v edges", len(result))
	}

	result = g.OutWith(3, qbe(EType("not-envlink")))
	if len(result) != 0 {
		t.Errorf("Vertex 3 should have no 'not-envlink'-typed out-edges, but got %v edges", len(result))
	}

	result = g.OutWith(3, qbe(ETypeNone, "ipv4", D_ipv4))
	if len(result) != 1 {
		t.Errorf("Vertex 3 should have one out-edge with prop 'ipv4' at correct value, but got %v edges", len(result))
	}

	result = g.OutWith(3, qbe(ETypeNone, "ipv4", "wrong"))
	if len(result) != 0 {
		t.Errorf("Vertex 3 should have zero out-edges with prop 'ipv4' at the wrong value, but got %v edges", len(result))
	}

	// TODO add some cases to ensure N>1 works for number of edges and number of prop queries
}
