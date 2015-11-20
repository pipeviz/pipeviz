package ingest_test

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/xeipuuv/gojsonschema"
	"github.com/pipeviz/pipeviz/fixtures"
	"github.com/pipeviz/pipeviz/ingest"
	"github.com/pipeviz/pipeviz/schema"
)

var msgJSON [][]byte

func init() {
	assets, err := fixtures.AssetDir("ein")
	if err != nil {
		panic("Could not load assets; has code generation (`make gen`) been done?")
	}

	sort.Strings(assets)
	for _, name := range assets {
		msgJSON = append(msgJSON, fixtures.MustAsset("ein/"+name))
	}
}

func BenchmarkUnmarshalMessageOne(b *testing.B) {
	d := msgJSON[0]

	for i := 0; i < b.N; i++ {
		m := &ingest.Message{}
		json.Unmarshal(d, m)
	}
}

func BenchmarkUnmarshalMessageTwo(b *testing.B) {
	d := msgJSON[1]

	for i := 0; i < b.N; i++ {
		m := &ingest.Message{}
		json.Unmarshal(d, m)
	}
}

func BenchmarkUnmarshalMessageOneAndTwo(b *testing.B) {
	d1, d2 := msgJSON[0], msgJSON[1]

	for i := 0; i < b.N; i++ {
		m1 := &ingest.Message{}
		json.Unmarshal(d1, m1)

		m2 := &ingest.Message{}
		json.Unmarshal(d2, m2)
	}
}

func BenchmarkValidateMessageOne(b *testing.B) {
	d := msgJSON[0]
	msg := gojsonschema.NewStringLoader(string(d))

	for i := 0; i < b.N; i++ {
		schema.Master().Validate(msg)
	}
}

func BenchmarkValidateMessageTwo(b *testing.B) {
	d := msgJSON[0]
	msg := gojsonschema.NewStringLoader(string(d))

	for i := 0; i < b.N; i++ {
		schema.Master().Validate(msg)
	}
}

func BenchmarkValidateMessageOneAndTwo(b *testing.B) {
	d1 := msgJSON[0]
	d2 := msgJSON[1]
	msg1 := gojsonschema.NewStringLoader(string(d1))
	msg2 := gojsonschema.NewStringLoader(string(d2))

	for i := 0; i < b.N; i++ {
		schema.Master().Validate(msg1)
		schema.Master().Validate(msg2)
	}
}
