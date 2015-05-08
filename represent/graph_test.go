package represent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/interpret"
)

var msgs []interpret.Message

func init() {
	for i := range make([]struct{}, 8) {
		m := interpret.Message{Id: i + 1}

		path := fmt.Sprintf("../fixtures/ein/%v.json", i+1)
		f, err := ioutil.ReadFile(path)
		if err != nil {
			panic("json fnf: " + path)
		}

		err = json.Unmarshal(f, &m)
		msgs = append(msgs, m)
	}
}

func TestClone(t *testing.T) {
	var g *coreGraph = &coreGraph{vtuples: ps.NewMap(), vserial: 0}
	g.vserial = 2
	g.vtuples = g.vtuples.Set("foo", "val")

	var g2 *coreGraph = g.clone()
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
	var g CoreGraph = &coreGraph{vtuples: ps.NewMap()}
	for i := 0; i < b.N; i++ {
		g.Merge(msgs[0])
	}
}

func BenchmarkMergeMessageOneAndTwo(b *testing.B) {
	var g CoreGraph = &coreGraph{vtuples: ps.NewMap()}

	for i := 0; i < b.N; i++ {
		g.Merge(msgs[0])
		g.Merge(msgs[1])
	}
}
