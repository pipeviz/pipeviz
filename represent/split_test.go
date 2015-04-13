package represent

import (
	"reflect"
	"testing"

	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
	"github.com/stretchr/testify/assert"
)

// For tests that rely on a simple, single vertex value, this provides a standard set
const (
	D_hostname string = "foohost"
	D_ipv4     string = "33.31.155.2"
	D_ipv6     string = "2001:4860:0:2001::68"
	D_nick     string = "bar-nick"
	D_commit   string = "f36becb37b195dcc7dbe191a55ac3b5b65e64f19"
	D_version  string = "2.2"
	D_semver   string = "0.12.9"
)

// Default values for use in environments; these complement the constants
var D_env interpret.Environment = interpret.Environment{
	Address:  M_addr[0],
	Os:       "unix",
	Provider: "vagrant",
	Type:     "physical",
	Nick:     D_nick,
}

// Default values for logic states; as defined, will induce no edges.
var D_ls interpret.LogicState = interpret.LogicState{
	ID: struct {
		Commit  string `json:"commit"`
		Version string `json:"version"`
		Semver  string `json:"semver"`
	}{Version: D_version},
	Lgroup: "bigparty",
	Nick:   D_nick,
	Path:   "/usr/local/src/imaginationland",
	Type:   "code",
}

// A matrix of standard interpret.Address values, including all permutations of field presence
var M_addr []interpret.Address = []interpret.Address{
	{Hostname: D_hostname},
	{Ipv4: D_ipv4},
	{Ipv6: D_ipv6},
	{Hostname: D_hostname, Ipv4: D_ipv4},
	{Hostname: D_hostname, Ipv6: D_ipv6},
	{Ipv4: D_ipv4, Ipv6: D_ipv6},
	{Hostname: D_hostname, Ipv4: D_ipv4, Ipv6: D_ipv6},
}

// A matrix of standard interpret.EnvLink values
var M_envlink []interpret.EnvLink = []interpret.EnvLink{
	{Address: M_addr[0]},
	{Address: M_addr[1]},
	{Address: M_addr[2]},
	{Address: M_addr[0], Nick: D_nick},
	{Address: M_addr[1], Nick: D_nick},
	{Address: M_addr[2], Nick: D_nick},
	{Address: interpret.Address{}, Nick: D_nick},
}

type FixtureEnvironmentSplit struct {
	Summary string
	Input   interpret.Environment
	Output  []SplitData
}

var F_Environment []FixtureEnvironmentSplit = []FixtureEnvironmentSplit{
	{
		Summary: "Minimal environment with only hostname",
		Input: interpret.Environment{
			Address: M_addr[0],
		},
		Output: []SplitData{
			{
				Vertex: environmentVertex{
					mapPropPairs(1, p{"hostname", D_hostname}),
				},
				EdgeSpecs: nil,
			},
		},
	},
	{
		Summary: "Minimal environment with hostname and nick",
		Input:   interpret.Environment{Address: M_addr[0], Nick: D_nick},
		Output: []SplitData{
			{
				Vertex: environmentVertex{
					mapPropPairs(1, p{"nick", D_nick}, p{"hostname", D_hostname}),
				},
				EdgeSpecs: nil,
			},
		},
	},
	{
		Summary: "Minimal environment with hostname, ipv4, ipv6, and nick",
		Input:   interpret.Environment{Address: M_addr[6], Nick: D_nick},
		Output: []SplitData{
			{
				Vertex: environmentVertex{
					mapPropPairs(1, p{"nick", D_nick}, p{"hostname", D_hostname}, p{"ipv4", D_ipv4}, p{"ipv6", D_ipv6}),
				},
				EdgeSpecs: nil,
			},
		},
	},
	{
		Summary: "Environment with all props set, plus hostname and nick",
		Input:   D_env,
		Output: []SplitData{
			{
				Vertex: environmentVertex{
					mapPropPairs(1, p{"nick", D_nick}, p{"hostname", D_hostname}, p{"os", D_env.Os}, p{"provider", D_env.Provider}, p{"type", D_env.Type}),
				},
				EdgeSpecs: nil,
			},
		},
	},
}

func compareSplitData(expect, actual []SplitData, t *testing.T) {
	if len(expect) != len(actual) {
		t.Errorf("SplitData slices are different lengths; guaranteed not equal. Expected len %v, actual %v", len(expect), len(actual))
	}

	for k, esd := range expect {
		asd := actual[k]
		et := reflect.TypeOf(esd.Vertex)
		at := reflect.TypeOf(asd.Vertex)

		if et.Name() != at.Name() {
			t.Errorf("Vertex type mismatch at SplitData index %v: expected %T, actual %T", k, esd.Vertex, asd.Vertex)
		}

		mapEq(esd.Vertex.Props(), asd.Vertex.Props(), t)
	}
}

func mapEq(expect, actual ps.Map, t *testing.T) {
	if expect.Size() != actual.Size() {
		t.Errorf("Prop maps are different sizes; guaranteed not equal. Expected size %v, actual %v", expect.Size(), actual.Size())
	}

	expect.ForEach(func(k string, val ps.Any) {
		aval, exists := actual.Lookup(k)
		if !exists {
			t.Errorf("Missing expected key '%v', expected value %v", k, val)
			return
		}

		assert.Equal(t, val, aval, "Values for key '%v' are not equal: expected %v, actual %v", k, val, aval)
	})
	// TODO if keys are missing/nonequal length, walk the actual map to find out what's not there and dump it?
}

func TestSplitEnvironment(t *testing.T) {
	for _, fixture := range F_Environment {
		t.Log("Split test on environment fixture:", fixture.Summary)

		// by convention we're always using msgid 1 in fixtures
		sd, err := Split(fixture.Input, 1)
		if err != nil {
			t.Error(err)
		}

		compareSplitData(fixture.Output, sd, t)
	}
}
