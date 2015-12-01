package fixtures_test

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/xeipuuv/gojsonschema"
	"github.com/pipeviz/pipeviz/fixtures"
	"github.com/pipeviz/pipeviz/schema"
)

// Tests that all message fixtures pass schema validation against the master
// message schema (../schema/schema.json).
func TestFixtureValidity(t *testing.T) {
	// Ensure stable order, for user's sanity
	names := fixtures.AssetNames()
	sort.Strings(names)

	// If this test is run without having generated assets via go-bindata, it could
	// pass. This should be precluded because the test won't compile without the
	// functions created by go-bindata, but this is an easy guard to put in place, so
	// might as well.
	if len(names) == 0 {
		t.Error("No asset files found for validation - has code generation been run correctly? (should be `make gen`)")
	}

	for _, name := range names {
		if testing.Verbose() {
			t.Log("Beginning validation on", name)
		}

		b, err := fixtures.Asset(name)
		result, err := schema.Master().Validate(gojsonschema.NewStringLoader(string(b)))

		if err != nil {
			t.Errorf("Error while validating fixture %v: %s", name, err)
			panic(err.Error())
		}

		if result.Valid() {
			if testing.Verbose() {
				t.Log(name, "passed validation")
			}
		} else {
			buf := bytes.NewBufferString(fmt.Sprintf("%s failed validation with the following errors:\n", name))
			for _, desc := range result.Errors() {
				buf.WriteString(fmt.Sprintf("%s\n", strings.Replace(desc.String(), "root", name, 1)))
			}

			t.Errorf(buf.String())
		}
	}
}
