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

func TestMerge(t *testing.T) {
	var g CoreGraph = &coreGraph{vtuples: ps.NewMap(), vserial: 0}

	g = g.Merge(msgs[0])

	//for _, v := range g.VerticesWith(qbv()) {
	//pretty.Print(v.flat())
	//}

	g = g.Merge(msgs[1])

	//for _, v := range g.VerticesWith(qbv()) {
	//pretty.Print(v.flat())
	//}
}
