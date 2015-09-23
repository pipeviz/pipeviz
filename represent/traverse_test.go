package represent

import (
	"strconv"
	"testing"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/tag1consulting/pipeviz/represent/types"
)

func tprops(pairs ...interface{}) []types.PropPair {
	ret := make([]types.PropPair, 0)
	for k, v := range pairs {
		if k%2 != 0 {
			continue
		}
		ret = append(ret, types.PropPair{K: v.(string), V: pairs[k+1]})
	}
	return ret
}

// utility func to create a vtTuple. puts edges in the right place by
// checking source/target ids. panics if they don't line up!
func mkTuple(vid int, vtx types.Vertex, edges ...types.StandardEdge) types.VertexTuple {
	vt := types.VertexTuple{
		ID:       vid,
		Vertex:   vtx,
		InEdges:  ps.NewMap(),
		OutEdges: ps.NewMap(),
	}

	for _, e := range edges {
		if e.Source == vid {
			vt.OutEdges = vt.OutEdges.Set(strconv.Itoa(e.ID), e)
		} else if e.Target == vid {
			vt.InEdges = vt.InEdges.Set(strconv.Itoa(e.ID), e)
		} else {
			panic("edge had neither source nor target of vid")
		}
	}

	return vt
}

// utility func to create a StandardEdge.
func mkEdge(id, source, target int, msgid uint64, etype string, props ...interface{}) types.StandardEdge {
	e := types.StandardEdge{
		ID:     id,
		Source: source,
		Target: target,
		EType:  types.EType(etype),
		Props:  ps.NewMap(),
	}

	var k string
	var v interface{}
	for len(props) > 1 {
		k, v, props = props[0].(string), props[1], props[2:]
		e.Props = e.Props.Set(k, types.Property{MsgSrc: msgid, Value: v})
	}

	return e
}

func getGraphFixture() *coreGraph {
	g := &coreGraph{vtuples: ps.NewMap(), vserial: 0}

	// Manually populate the graph with some dummy vertices and edges.
	// These don't necessarily line up with any real schemas, on purpose.

	// edge, id 10, connects vid 1 to vid 2. msgid 2. type "dummy-edge-type1". one prop - "eprop1": "foo".
	edge10 := mkEdge(10, 1, 2, 2, "dummy-edge-type1", "eprop1", "foo")
	// edge, id 11, connects vid 3 to vid 1. msgid 3. type "dummy-edge-type2". one prop - "eprop2": "bar".
	edge11 := mkEdge(11, 3, 1, 3, "dummy-edge-type2", "eprop2", "bar")
	// edge, id 12, connects vid 3 to vid 4. msgid 4. type "dummy-edge-type2". one prop - "eprop2": "baz".
	edge12 := mkEdge(12, 3, 4, 3, "dummy-edge-type2", "eprop2", "baz")
	// edge, id 13, connects vid 3 to vid 4. msgid 4. type "dummy-edge-type3". two props - "eprop2": "qux", "eprop3": 42.
	edge13 := mkEdge(13, 3, 4, 4, "dummy-edge-type3", "eprop2", "bar", "eprop3", 42)

	// vid 1, type "env". two props - "prop1": "bar", "prop2": 42. msgid 1
	vt1 := mkTuple(1, types.NewVertex("env", 1, tprops("prop1", "foo", "prop2", 42)...), edge10, edge11) // one in, one out
	g.vtuples = g.vtuples.Set(strconv.Itoa(1), vt1)

	// vid 2, type "env". , "one prop - "prop1", "foo". msgid 2
	vt2 := mkTuple(2, types.NewVertex("env", 2, tprops("prop1", "bar")...), edge10) // one in
	g.vtuples = g.vtuples.Set(strconv.Itoa(2), vt2)

	// vid 3, type "vt2". two props - "prop1", "bar", "bowser", "moo". msgid 3
	vt3 := mkTuple(3, types.NewVertex("vt2", 3, tprops("prop1", "bar", "bowser", "moo")...), edge11, edge12, edge13) // three out
	g.vtuples = g.vtuples.Set(strconv.Itoa(3), vt3)

	// vid 4, type "vt3". three props - "prop1", "baz", "prop2", 42, "prop3", "qux". msgid 4
	vt4 := mkTuple(4, types.NewVertex("vt3", 4, tprops("prop1", "baz", "prop2", 42, "prop3", "qux")...), edge12, edge13) // two in, same origin
	g.vtuples = g.vtuples.Set(strconv.Itoa(4), vt4)

	// vid 5, type "vt3". no props, no edges. msgid 5
	vt5 := mkTuple(5, types.NewVertex("vt3", 5)) // none in or out
	g.vtuples = g.vtuples.Set(strconv.Itoa(5), vt5)
	g.vserial = 13

	return g
}

func TestQbv(t *testing.T) {
	// ensure implement both VFilter and EFilter interfaces
	var _ VEFilter = vertexFilter{}

	assert.Equal(t, Qbv(), vertexFilter{}, "qbv with no args creates an empty vertexFilter")
	assert.Equal(t, Qbv(), vertexFilter{vtype: VTypeNone}, "qbv with no args creates equivalent of passing VTypeNone as first arg")
	assert.Equal(t, Qbv(types.VType("foo")), vertexFilter{vtype: types.VType("foo")}, "qbv with single arg assigns to VType struct prop")
	assert.Equal(t, Qbv(VTypeNone, "foo"), vertexFilter{vtype: VTypeNone}, "qbv with two args ignores second (unpaired) arg")
	assert.Equal(t, Qbv(VTypeNone, "foo", "bar"), vertexFilter{vtype: VTypeNone, props: []types.PropPair{{"foo", "bar"}}}, "qbv with three args creates one pair of second (key) and third (value) args")
	assert.Equal(t, Qbv(VTypeNone, "foo", "bar", "baz"), vertexFilter{vtype: VTypeNone, props: []types.PropPair{{"foo", "bar"}}}, "qbv with four args creates one pair from 2nd and 3rd args, ignores 4th")

	// ensure that some incorrect things owing to loose typing correctly panic
	assert.Panics(t, func() {
		Qbv("foo")
	}, "qbv panics on type conversion when passing a string instead of VType")

	assert.Panics(t, func() {
		Qbv(VTypeNone, 1, "foo")
	}, "qbv panics on type conversion when second argument (with corresponding pair val 3rd arg) is non-string")

	assert.Panics(t, func() {
		Qbv(VTypeNone, "foo", "bar", 1, "baz")
	}, "qbv panics on type conversion when Nth even argument (with corresponding pair val N+1 arg) is non-string")
}

func TestQbe(t *testing.T) {
	// ensure implement both VFilter and EFilter interfaces
	var _ VEFilter = edgeFilter{}

	assert.Equal(t, Qbe(), edgeFilter{}, "qbe with no args creates an empty edgeFilter")
	assert.Equal(t, Qbe(), edgeFilter{etype: ETypeNone}, "qbe with no args creates equivalent of passing ETypeNone as first arg")
	assert.Equal(t, Qbe(types.EType("foo")), edgeFilter{etype: types.EType("foo")}, "qbe with single arg assigns to EType struct prop")
	assert.Equal(t, Qbe(ETypeNone, "foo"), edgeFilter{etype: ETypeNone}, "qbe with two args ignores second (unpaired) arg")
	assert.Equal(t, Qbe(ETypeNone, "foo", "bar"), edgeFilter{etype: ETypeNone, props: []types.PropPair{{"foo", "bar"}}}, "qbe with three args creates one pair of second (key) and third (value) args")
	assert.Equal(t, Qbe(ETypeNone, "foo", "bar", "baz"), edgeFilter{etype: ETypeNone, props: []types.PropPair{{"foo", "bar"}}}, "qbe with four args creates one pair from 2nd and 3rd args, ignores 4th")

	// ensure that some incorrect things owing to loose typing correctly panic
	assert.Panics(t, func() {
		Qbe("foo")
	}, "qbe panics on type conversion when passing a string instead of EType")

	assert.Panics(t, func() {
		Qbe(ETypeNone, 1, "foo")
	}, "qbe panics on type conversion when second argument (with corresponding pair val 3rd arg) is non-string")

	assert.Panics(t, func() {
		Qbe(ETypeNone, "foo", "bar", 1, "baz")
	}, "qbe panics on type conversion when Nth even argument (with corresponding pair val N+1 arg) is non-string")
}

func TestVerticesWith(t *testing.T) {
	g := getGraphFixture()
	var result []types.VertexTuple

	result = g.VerticesWith(Qbv())
	if len(result) != 5 {
		t.Errorf("Should find 4 vertices with no filter; found %v", len(result))
	}

	result = g.VerticesWith(Qbv(types.VType("env")))
	if len(result) != 2 {
		t.Errorf("Should find 2 vertices when filtering to type env; found %v", len(result))
	}

	result = g.VerticesWith(Qbv(types.VType("nonexistent-type")))
	if len(result) != 0 {
		t.Errorf("Should find no vertices when filtering on type that's not present; found %v", len(result))
	}

	result = g.VerticesWith(Qbv(VTypeNone, "prop1", "bar"))
	if len(result) != 2 {
		t.Errorf("Should find two vertices with prop1 == \"bar\"; found %v", len(result))
	}

	result = g.VerticesWith(Qbv(VTypeNone, "none-have-this-prop-key", "doesn't matter"))
	if len(result) != 0 {
		t.Errorf("Should find no vertices when filtering on nonexistent prop key; found %v", len(result))
	}

	result = g.VerticesWith(Qbv(types.VType("env"), "prop1", "foo"))
	if len(result) != 1 {
		t.Errorf("Should find one vertex when filtering to env types and with prop1 == \"foo\"; found %v", len(result))
	}

	result = g.VerticesWith(Qbv(types.VType("env"), "prop2", 42))
	if len(result) != 1 {
		t.Errorf("Should find one vertex when filtering to env types and with prop2 == 42; found %v", len(result))
	}
}

// Tests arcWith(), which effectively tests both OutWith() and InWith().
func TestOutInArcWith(t *testing.T) {
	g := getGraphFixture()
	var result []types.StandardEdge

	// first test zero-case - vtx 5 has no edges
	result = g.arcWith(5, Qbe(), false)
	if len(result) != 0 {
		t.Errorf("Vertex 5 has no edges at all, but still got %v out-edge results", len(result))
	}

	result = g.arcWith(5, Qbe(), true)
	if len(result) != 0 {
		t.Errorf("Vertex 5 has no edges at all, but still got %v in-edge results", len(result))
	}

	// next test single case - vtx 1 has one in, one out
	result = g.arcWith(1, Qbe(), true)
	if len(result) != 1 {
		t.Errorf("Vertex 1 should have one in-edge, but got %v edges", len(result))
	}

	// ensure InWith behaves same as arcWith + arg
	result = g.InWith(1, Qbe())
	if len(result) != 1 {
		t.Errorf("Vertex 1 should have one in-edge, but got %v edges (InWith calls arcWith correctly)", len(result))
	}

	result = g.arcWith(1, Qbe(), false)
	if len(result) != 1 {
		t.Errorf("Vertex 1 should have one out-edge, but got %v edges", len(result))
	}

	result = g.OutWith(1, Qbe())
	if len(result) != 1 {
		t.Errorf("Vertex 1 should have one out-edge, but got %v edges (OutWith calls arcWith correctly)", len(result))
	}

	// last of basic tests - N>1 number of edges
	result = g.arcWith(4, Qbe(), true)
	if len(result) != 2 {
		t.Errorf("Vertex 4 has two in-edges, but got %v in-edge results", len(result))
	}

	result = g.InWith(4, Qbe())
	if len(result) != 2 {
		t.Errorf("Vertex 4 has two in-edges, but got %v in-edge results (InWith calls arcWith correctly)", len(result))
	}

	result = g.arcWith(3, Qbe(), false)
	if len(result) != 3 {
		t.Errorf("Vertex 3 should have three out-edges, but got %v edges", len(result))
	}

	result = g.OutWith(3, Qbe())
	if len(result) != 3 {
		t.Errorf("Vertex 3 should have three out-edges, but got %v edges (OutWith calls arcWith correctly)", len(result))
	}

	result = g.InWith(3, Qbe())
	if len(result) != 0 {
		t.Errorf("Vertex 3 has out-edges but no in-edge; still got %v in-edge results", len(result))
	}

	// now, tests that actually exercise the filter
	result = g.OutWith(3, Qbe(ETypeNone))
	if len(result) != 3 {
		t.Errorf("ETypeNone does not correctly matches all edge types - should've gotten 3 out-edges, but got %v edges", len(result))
	}

	// basic edge type filtering
	result = g.OutWith(3, Qbe(types.EType("dummy-edge-type2")))
	if len(result) != 2 {
		t.Errorf("Vertex 2 should have two \"dummy-edge-type2\"-typed out-edges, but got %v edges", len(result))
	}

	// nonexistent type means no results
	result = g.InWith(2, Qbe(types.EType("nonexistent-type")))
	if len(result) != 0 {
		t.Errorf("Vertex 2 should have no edges of a nonexistent type, but got %v edges", len(result))
	}

	// existing edge type, but not one this vt has
	result = g.InWith(3, Qbe(types.EType("dummy-edge-type1")))
	if len(result) != 0 {
		t.Errorf("Vertex 3 has none of the \"dummy-edge-type1\" edges (though it is a real type in the graph); however, got %v edges", len(result))
	}

	// test prop-checking
	result = g.OutWith(3, Qbe(ETypeNone, "eprop2", "baz"))
	if len(result) != 1 {
		t.Errorf("Vertex 3 should have one out-edge with \"eprop2\" at \"baz\", but got %v edges", len(result))
	}

	result = g.OutWith(3, Qbe(ETypeNone, "eprop2", "bar"))
	if len(result) != 2 {
		t.Errorf("Vertex 3 should have two out-edges with \"eprop2\" at \"bar\", but got %v edges", len(result))
	}

	// test multi-prop checking - ensure they\"re ANDed
	result = g.OutWith(3, Qbe(ETypeNone, "eprop2", "bar", "eprop3", 42))
	if len(result) != 1 {
		t.Errorf("Vertex 3 should have one out-edge with \"eprop2\" at \"bar\" AND \"eprop3\" at 42, but got %v edges", len(result))
	}

	result = g.OutWith(3, Qbe(ETypeNone, "eprop2", "baz", "eprop3", 42))
	if len(result) != 0 {
		// OR would\"ve here would produce 2 edges
		t.Errorf("Vertex 3 should have no out-edges with \"eprop2\" at \"baz\" AND \"eprop3\" at 42 , but got %v edges", len(result))
	}

	result = g.OutWith(3, Qbe(types.EType("dummy-edge-type2"), "eprop2", "bar"))
	if len(result) != 1 {
		t.Errorf("Vertex 3 should have one out-edges that is dummy type2 AND has \"eprop2\" at \"bar\", but got %v edges", len(result))
	}
}

// Tests adjacentWith(), which effectively tests SuccessorsWith() and PredecessorsWith()
func TestAdjacentWith(t *testing.T) {
	g := getGraphFixture()
	var result []types.VertexTuple

	// basic, unfiltered tests first to ensure the right data is coming through
	// vtx 2 has just one in-edge
	result = g.adjacentWith(2, Qbv(), true)
	if len(result) != 1 {
		t.Errorf("Vertex 2 has one predecessor, but got %v vertices", len(result))
	}

	result = g.PredecessorsWith(2, Qbv())
	if len(result) != 1 {
		t.Errorf("Vertex 2 has one predecessor, but got %v vertices", len(result))
	}

	// vtx 1 has one out-edge and one in-edge
	result = g.adjacentWith(1, Qbv(), false)
	if len(result) != 1 {
		t.Errorf("Vertex 1 has one successor, but got %v vertices", len(result))
	}

	result = g.SuccessorsWith(1, Qbv())
	if len(result) != 1 {
		t.Errorf("Vertex 1 has one successor, but got %v vertices", len(result))
	}

	// vtx 5 is an isolate
	result = g.adjacentWith(5, Qbv(), true)
	if len(result) != 0 {
		t.Errorf("Vertex 5 has no predecessors, but got %v vertices", len(result))
	}

	result = g.PredecessorsWith(5, Qbv())
	if len(result) != 0 {
		t.Errorf("Vertex 5 has no predecessors, but got %v vertices", len(result))
	}

	result = g.adjacentWith(5, Qbv(), false)
	if len(result) != 0 {
		t.Errorf("Vertex 5 has no successors, but got %v vertices", len(result))
	}

	result = g.SuccessorsWith(5, Qbv())
	if len(result) != 0 {
		t.Errorf("Vertex 5 has no successors, but got %v vertices", len(result))
	}

	// qbe w/out args should be equivalent
	result = g.PredecessorsWith(2, Qbe())
	if len(result) != 1 {
		t.Errorf("Vertex 2 has one predecessor, but got %v vertices (qbe)", len(result))
	}

	result = g.SuccessorsWith(1, Qbe())
	if len(result) != 1 {
		t.Errorf("Vertex 1 has one successor, but got %v vertices (qbe)", len(result))
	}

	// deduping: vtx 4 has two in-edges and none out, but those edges are parallel so only one unique vtx
	result = g.PredecessorsWith(4, Qbv())
	if len(result) != 1 {
		t.Errorf("Vertex 4 has two in-edges, but only one unique predecessor; however, got %v vertices", len(result))
	}

	// vtx 3 is on the other side of vtx 4 - three out-edges, but only two uniques
	result = g.SuccessorsWith(3, Qbv())
	if len(result) != 2 {
		t.Errorf("Vertex 4 has three out-edges, but only two unique successors; however, got %v vertices", len(result))
	}

	// filter checks, beginning with edge and/or vertex typing
	result = g.SuccessorsWith(3, Qbv(types.VType("vt3")))
	if len(result) != 1 {
		t.Errorf("Vertex 4 has only one unique successor of type \"vt3\"; however, got %v vertices", len(result))
	}

	result = g.SuccessorsWith(3, Qbe(types.EType("dummy-edge-type2")))
	if len(result) != 2 {
		t.Errorf("Vertex 4 has two out-edges of \"dummy-edge-type2\" and both point to different vertices, so expecting 2, but got %v vertices", len(result))
	}

	result = g.SuccessorsWith(3, Qbe(types.EType("dummy-edge-type2")).and(Qbv(types.VType("env"))))
	if len(result) != 1 {
		t.Errorf("Vertex 4 has two unique successors along \"dummy-edge-type2\" out-edges, but only one is vtype \"env\". However, got %v vertices", len(result))
	}

	result = g.SuccessorsWith(3, Qbe(types.EType("dummy-edge-type3")).and(Qbv(types.VType("env"))))
	if len(result) != 0 {
		t.Errorf("Vertex 4 has one unique successor along \"dummy-edge-type3\" out-edges, but it is not an \"env\" type. However, got %v vertices", len(result))
	}

	// prop-filtering checks
	result = g.SuccessorsWith(3, Qbv(VTypeNone, "prop2", 42))
	if len(result) != 2 {
		t.Errorf("Vertex 4 has only two unique successors with \"prop2\" at 42; however, got %v vertices", len(result))
	}

	result = g.SuccessorsWith(3, Qbe(ETypeNone, "eprop2", "bar"))
	if len(result) != 2 {
		t.Errorf("Vertex 4 has two unique successors connected by two out-edges with \"eprop2\" at \"bar\"; however, got %v vertices", len(result))
	}

	result = g.SuccessorsWith(3, Qbv(VTypeNone, "prop1", "baz", "prop2", 42))
	if len(result) != 1 {
		t.Errorf("Vertex 4 has only one unique successor with \"prop1\" at \"baz\" and \"prop2\" at 42; however, got %v vertices", len(result))
	}

	result = g.SuccessorsWith(3, Qbv(VTypeNone, "prop3", "qux", "prop2", 42))
	if len(result) != 1 {
		t.Errorf("Vertex 4 has only one unique successor with BOTH \"prop3\" at \"qux\" and \"prop2\" at 42; however, got %v vertices", len(result))
	}

	result = g.SuccessorsWith(3, Qbe(ETypeNone, "eprop2", "bar").and(Qbv(VTypeNone, "prop1", "baz")))
	if len(result) != 1 {
		t.Errorf("Vertex 4 has only one unique successor with \"prop1\" at \"baz\" along an out-edge with \"eprop2\" at \"bar\"; however, got %v vertices", len(result))
	}

	result = g.SuccessorsWith(3, Qbe(ETypeNone, "eprop2", "bar").and(Qbv(types.VType("vt3"), "prop1", "baz")))
	if len(result) != 1 {
		t.Errorf("Vertex 4 has one unique successor of type \"vt3\" with \"prop1\" at \"baz\" along an out-edge with \"eprop2\" at \"bar\"; however, got %v vertices", len(result))
	}
}
