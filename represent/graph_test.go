package represent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/ingest"
	"github.com/pipeviz/pipeviz/types/system"
)

var msgs []ingest.Message

func init() {
	for i := range make([]struct{}, 8) {
		m := ingest.Message{}

		path := fmt.Sprintf("../fixtures/ein/%v.json", i+1)
		f, err := ioutil.ReadFile(path)
		if err != nil {
			panic("json fnf: " + path)
		}

		_ = json.Unmarshal(f, &m)
		msgs = append(msgs, m)
	}
}

func TestClone(t *testing.T) {
	g := &coreGraph{vtuples: ps.NewMap(), vserial: 0}
	g.vserial = 2
	g.vtuples = g.vtuples.Set("foo", "val")

	g2 := g.clone()
	g2.vserial = 4
	g2.vtuples = g2.vtuples.Set("foo", "newval")

	if g.vserial != 2 {
		t.Errorf("changes in cloned graph propagated back to original")
	}
	if val, _ := g2.vtuples.Lookup("foo"); val != "newval" {
		t.Errorf("map somehow propagated changes back up to original map")
	}
}

func BenchmarkMergeMessageOne(b *testing.B) {
	g := &coreGraph{vtuples: ps.NewMap()}
	for i := 0; i < b.N; i++ {
		g.Merge(0, msgs[0].UnificationForm())
	}
}

func BenchmarkMergeMessageOneAndTwo(b *testing.B) {
	var g system.CoreGraph = &coreGraph{vtuples: ps.NewMap()}

	for i := 0; i < b.N; i++ {
		g.Merge(0, msgs[0].UnificationForm())
		g.Merge(0, msgs[1].UnificationForm())
	}
}
