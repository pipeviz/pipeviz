package main_test

import (
	"fmt"
	gjs "github.com/xeipuuv/gojsonschema"
	"testing"
)

func TestMessageValidity(t *testing.T) {
	schema := gjs.NewReferenceLoader("file:///Users/sdboyer/ws/go/src/github.com/sdboyer/pipeviz/schema.json")
	f1 := gjs.NewReferenceLoader("file:///Users/sdboyer/ws/go/src/github.com/sdboyer/pipeviz/fixtures/ein/1.json")

	result, err := gjs.Validate(schema, f1)
	if err != nil {
		panic(err.Error())
	}

	if result.Valid() {
		fmt.Printf("The document is valid\n")
	} else {
		fmt.Printf("The document is not valid. see errors :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}
}
