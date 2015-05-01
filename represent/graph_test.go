package represent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
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
