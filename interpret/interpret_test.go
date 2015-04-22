package interpret_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/sdboyer/pipeviz/interpret"
)

var Msgs []*interpret.Message

func init() {
	for i := range make([]struct{}, 8) {
		m := &interpret.Message{Id: i + 1}

		path := fmt.Sprintf("../fixtures/ein/%v.json", i+1)
		f, err := ioutil.ReadFile(path)
		if err != nil {
			panic("json fnf: " + path)
		}

		err = json.Unmarshal(f, m)
		Msgs = append(Msgs, m)
	}
}

func TestUnmarshal(t *testing.T) {
	m := interpret.Message{}

	f, err := ioutil.ReadFile("../fixtures/ein/6.json")
	if err != nil {
		t.Error("fnf")
	}

	err = json.Unmarshal(f, &m)
}
