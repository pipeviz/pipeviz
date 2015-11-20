package represent

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/fixtures"
	"github.com/pipeviz/pipeviz/ingest"
	"github.com/pipeviz/pipeviz/types/system"
)

var msgs []ingest.Message

func init() {
	assets, err := fixtures.AssetDir("ein")
	if err != nil {
		panic("Could not load assets; has code generation (`make gen`) been done?")
	}

	sort.Strings(assets)
	for _, name := range assets {
		m := ingest.Message{}
		_ = json.Unmarshal(fixtures.MustAsset("ein/"+name), &m)
		msgs = append(msgs, m)
	}
}

func TestClone(t *testing.T) {
	g := &coreGraph{vtuples: newIntMap(), vserial: 0}
	g.vserial = 2
	g.vtuples = g.vtuples.Set(42, system.VertexTuple{
		Vertex: system.StdVertex{
			Type: system.VType("foo"),
		},
		InEdges:  ps.NewMap(),
		OutEdges: ps.NewMap(),
	})

	var g2 *coreGraph = g.clone()
	g2.vserial = 4
	g2.vtuples = g2.vtuples.Set(42, system.VertexTuple{
		Vertex: system.StdVertex{
			Type: system.VType("newval"),
		},
		InEdges:  ps.NewMap(),
		OutEdges: ps.NewMap(),
	})

	if g.vserial != 2 {
		t.Errorf("changes in cloned graph propagated back to original")
	}
	if val, _ := g2.vtuples.Get(42); val.Vertex.Type != system.VType("newval") {
		t.Errorf("map somehow propagated changes back up to original map")
	}
}

func BenchmarkMergeMessageOne(b *testing.B) {
	g := &coreGraph{vtuples: newIntMap()}
	m := msgs[0]

	for i := 0; i < b.N; i++ {
		g.Merge(0, m.UnificationForm())
	}
}

func BenchmarkMergeMessageOneAndTwo(b *testing.B) {
	var g system.CoreGraph = &coreGraph{vtuples: newIntMap()}
	m1 := msgs[0]
	m2 := msgs[1]

	for i := 0; i < b.N; i++ {
		g.Merge(0, m1.UnificationForm())
		g.Merge(0, m2.UnificationForm())
	}
}
