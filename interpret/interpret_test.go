package interpret_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/kr/pretty"
	"github.com/sdboyer/pipeviz/interpret"
)

func TestUnmarshal(t *testing.T) {
	m := &interpret.Message{
		Nodes: make([]interpret.Node, 0),
	}

	f, err := ioutil.ReadFile("../fixtures/ein/2.json")
	if err != nil {
		t.Error("fnf")
	}

	json.Unmarshal(f, m)
	pretty.Print(m)
}
