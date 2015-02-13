package main_test

import (
	gjs "github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"strings"
	"testing"
)

func TestMessageValidity(t *testing.T) {
	src, err := ioutil.ReadFile("./schema.json")
	if err != nil {
		t.Error("Failed to open master schema file, test must abort. Note: this test must be run from the pipeviz repo root. message:", err.Error())
		t.FailNow()
	}

	schema, err := gjs.NewSchema(gjs.NewStringLoader(string(src)))
	if err != nil {
		t.Error("Failed to create a schema object from the master schema.json:", err.Error())
	}

	files, err := ioutil.ReadDir("./fixtures/ein/")
	if err != nil {
		t.Error("Failed to scan message fixtures dir:", err.Error())
		t.FailNow()
	}

	for _, f := range files {
		if testing.Verbose() {
			t.Log("Beginning validation on", f.Name())
		}

		src, _ = ioutil.ReadFile("./fixtures/ein/" + f.Name())
		msg := gjs.NewStringLoader(string(src))
		result, err := schema.Validate(msg)

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
