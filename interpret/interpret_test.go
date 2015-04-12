package interpret_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/sdboyer/pipeviz/interpret"
)

func TestUnmarshal(t *testing.T) {
	m := interpret.Message{}

	f, err := ioutil.ReadFile("../fixtures/ein/6.json")
	if err != nil {
		t.Error("fnf")
	}

	err = json.Unmarshal(f, &m)
}
