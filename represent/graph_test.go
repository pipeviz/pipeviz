package represent

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
)

func TestMerge(t *testing.T) {
	var g CoreGraph = &coreGraph{vtuples: ps.NewMap(), vserial: 0}

	m := interpret.Message{Id: 1}

	f, err := ioutil.ReadFile("../fixtures/ein/1.json")
	if err != nil {
		t.Error("fnf")
	}

	err = json.Unmarshal(f, &m)
	//pretty.Print(m)
	g = g.Merge(m)
}
