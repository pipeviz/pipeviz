package ingest_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	gjs "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
	"github.com/tag1consulting/pipeviz/fixtures"
	"github.com/tag1consulting/pipeviz/ingest"
	"github.com/tag1consulting/pipeviz/schema"
)

var (
	masterSchema *gjs.Schema
	Msgs         []*ingest.Message
	MsgJSON      [][]byte
)

func init() {
	for i := range make([]struct{}, 8) {
		m := &ingest.Message{}

		path := fmt.Sprintf("../fixtures/ein/%v.json", i+1)
		f, err := ioutil.ReadFile(path)
		if err != nil {
			panic("json fnf: " + path)
		}

		MsgJSON = append(MsgJSON, f)
		_ = json.Unmarshal(f, m)
		Msgs = append(Msgs, m)
	}

	src, err := schema.Master()
	if err != nil {
		panic(fmt.Sprint("Failed to open master schema file, test must abort. Note: this test must be run from the pipeviz repo root. message:", err.Error()))
	}

	masterSchema, err = gjs.NewSchema(gjs.NewStringLoader(string(src)))
	if err != nil {
		panic(fmt.Sprint("Failed to create a schema object from the master schema.json:", err.Error()))
	}
}

// Reads all message fixtures from fixtures/ein and validates them
// against the master message schema (schema.json).
func TestMessageValidity(t *testing.T) {
	files, err := ioutil.ReadDir("../fixtures/ein/")
	if err != nil {
		t.Error("Failed to scan message fixtures dir:", err.Error())
		t.FailNow()
	}

	for _, f := range files {
		if testing.Verbose() {
			t.Log("Beginning validation on", f.Name())
		}

		src, _ := ioutil.ReadFile("../fixtures/ein/" + f.Name())
		msg := gjs.NewStringLoader(string(src))
		result, err := masterSchema.Validate(msg)

		if err != nil {
			panic(err.Error())
		}

		if result.Valid() {
			if testing.Verbose() {
				t.Log(f.Name(), "passed validation")
			}
		} else {
			for _, desc := range result.Errors() {
				t.Errorf("%s\n", strings.Replace(desc.String(), "root", f.Name(), 1))
			}
		}
	}
}

func BenchmarkUnmarshalMessageOne(b *testing.B) {
	d, _ := fixtures.Asset("1.json")

	for i := 0; i < b.N; i++ {
		m := &ingest.Message{}
		json.Unmarshal(d, m)
	}
}

func BenchmarkUnmarshalMessageTwo(b *testing.B) {
	d, _ := fixtures.Asset("2.json")

	for i := 0; i < b.N; i++ {
		m := &ingest.Message{}
		json.Unmarshal(d, m)
	}
}

func BenchmarkUnmarshalMessageOneAndTwo(b *testing.B) {
	d1 := MsgJSON[0]
	d2 := MsgJSON[1]

	for i := 0; i < b.N; i++ {
		m1 := &ingest.Message{}
		json.Unmarshal(d1, m1)

		m2 := &ingest.Message{}
		json.Unmarshal(d2, m2)
	}
}

func BenchmarkValidateMessageOne(b *testing.B) {
	d := MsgJSON[0]
	msg := gjs.NewStringLoader(string(d))

	for i := 0; i < b.N; i++ {
		masterSchema.Validate(msg)
	}
}

func BenchmarkValidateMessageTwo(b *testing.B) {
	d := MsgJSON[0]
	msg := gjs.NewStringLoader(string(d))

	for i := 0; i < b.N; i++ {
		masterSchema.Validate(msg)
	}
}

func BenchmarkValidateMessageOneAndTwo(b *testing.B) {
	d1 := MsgJSON[0]
	d2 := MsgJSON[1]
	msg1 := gjs.NewStringLoader(string(d1))
	msg2 := gjs.NewStringLoader(string(d2))

	for i := 0; i < b.N; i++ {
		masterSchema.Validate(msg1)
		masterSchema.Validate(msg2)
	}
}
