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
	D_msgid    int    = 1
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
					mapPropPairs(D_msgid, p{"hostname", D_hostname}),
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
					mapPropPairs(D_msgid, p{"nick", D_nick}, p{"hostname", D_hostname}),
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
					mapPropPairs(D_msgid, p{"nick", D_nick}, p{"hostname", D_hostname}, p{"ipv4", D_ipv4}, p{"ipv6", D_ipv6}),
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
					mapPropPairs(D_msgid, p{"nick", D_nick}, p{"hostname", D_hostname}, p{"os", D_env.Os}, p{"provider", D_env.Provider}, p{"type", D_env.Type}),
				},
				EdgeSpecs: nil,
			},
		},
	},
}

type FixtureLogicStateSplit struct {
	Summary string
	Input   interpret.LogicState
	Output  []SplitData
}

var F_LogicState []FixtureLogicStateSplit

type FixtureProcessSplit struct {
	Summary string
	Input   interpret.Process
	Output  []SplitData
}

var F_Process []FixtureProcessSplit

func init() {
	lsIds := []struct {
		Commit  string `json:"commit"`
		Version string `json:"version"`
		Semver  string `json:"semver"`
	}{
		{Commit: D_commit},
		{Version: D_version},
		{Semver: D_semver},
	}

	datalinks := []interpret.DataLink{
		{
			Name:        "foo",
			Type:        "mediated",
			Interaction: "rw",
			ConnUnix:    interpret.ConnUnix{"/var/run/db.sock"},
		},
		{
			Type:        "file",
			Interaction: "ro",
			ConnUnix:    interpret.ConnUnix{"/var/run/db/sthlike.sqlite"},
		},
		{
			Name:        "mysql db",
			Type:        "mediated",
			Interaction: "rw",
			Subset:      "appdb",
			ConnNet: interpret.ConnNet{
				Hostname: D_hostname,
				Port:     3306,
				Proto:    "tcp",
			},
		},
	}

	F_LogicState = []FixtureLogicStateSplit{
		{
			Summary: "Commit id, path, envlink. No props or datasets.",
			Input:   interpret.LogicState{ID: lsIds[0], Path: D_ls.Path, Environment: M_envlink[1]},
			Output: []SplitData{
				{
					Vertex: logicStateVertex{
						mapPropPairs(D_msgid, p{"path", D_ls.Path}),
					},
					EdgeSpecs: EdgeSpecs{
						SpecCommit{[]byte(D_commit)},
						M_envlink[1],
					},
				},
			},
		},
		{
			Summary: "Version, path, envlink. No props, datasets.",
			Input:   interpret.LogicState{ID: lsIds[1], Path: D_ls.Path, Environment: M_envlink[1]},
			Output: []SplitData{
				{
					Vertex: logicStateVertex{
						mapPropPairs(D_msgid, p{"path", D_ls.Path}, p{"version", D_version}),
					},
					EdgeSpecs: EdgeSpecs{
						M_envlink[1],
					},
				},
			},
		},
		{
			Summary: "Semver, path, and envlink. No props or datasets.",
			Input:   interpret.LogicState{ID: lsIds[2], Path: D_ls.Path, Environment: M_envlink[1]},
			Output: []SplitData{
				{
					Vertex: logicStateVertex{
						mapPropPairs(D_msgid, p{"path", D_ls.Path}, p{"semver", D_semver}),
					},
					EdgeSpecs: EdgeSpecs{
						M_envlink[1],
					},
				},
			},
		},
		{
			Summary: "Commit id, path, envlink, and all non-edge inducing props.",
			Input: interpret.LogicState{
				ID:          lsIds[0],
				Lgroup:      D_ls.Lgroup,
				Nick:        D_nick,
				Path:        D_ls.Path,
				Type:        D_ls.Type,
				Environment: M_envlink[1],
			},
			Output: []SplitData{
				{
					Vertex: logicStateVertex{
						mapPropPairs(D_msgid, p{"path", D_ls.Path}, p{"type", D_ls.Type}, p{"lgroup", D_ls.Lgroup}, p{"nick", D_nick}),
					},
					EdgeSpecs: EdgeSpecs{
						SpecCommit{[]byte(D_commit)},
						M_envlink[1],
					},
				},
			},
		},
		{
			Summary: "Semver and path, two local datasets, and an envlink.",
			Input: interpret.LogicState{
				ID:          lsIds[2],
				Path:        D_ls.Path,
				Datasets:    datalinks[:2],
				Environment: M_envlink[0],
			},
			Output: []SplitData{
				{
					Vertex: logicStateVertex{
						mapPropPairs(D_msgid, p{"path", D_ls.Path}, p{"semver", D_semver}),
					},
					EdgeSpecs: EdgeSpecs{
						datalinks[0],
						datalinks[1],
						M_envlink[0],
					},
				},
			},
		},
		{
			Summary: "Semver and path, one remote dataset, and an envlink.",
			Input: interpret.LogicState{
				ID:          lsIds[2],
				Path:        D_ls.Path,
				Datasets:    datalinks[2:],
				Environment: M_envlink[6],
			},
			Output: []SplitData{
				{
					Vertex: logicStateVertex{
						mapPropPairs(D_msgid, p{"path", D_ls.Path}, p{"semver", D_semver}),
					},
					EdgeSpecs: EdgeSpecs{
						datalinks[2],
						M_envlink[6],
					},
				},
			},
		},
	}

	F_Process = []FixtureProcessSplit{
		{
			Summary: "Pid, envlink, one logic state, nothing else.",
			Input: interpret.Process{
				Pid:         42,
				Environment: M_envlink[0],
				LogicStates: []string{"/path/to/sth"},
			},
			Output: []SplitData{
				{
					Vertex: processVertex{mapPropPairs(D_msgid, p{"pid", 42})},
					EdgeSpecs: []EdgeSpec{
						SpecLocalLogic{"/path/to/sth"},
						M_envlink[0],
					},
				},
			},
		},
		{
			Summary: "Pid, envlink, one logic state, all props, no listeners.",
			Input: interpret.Process{
				Pid:         42,
				Environment: M_envlink[1],
				LogicStates: []string{"/path/to/sth"},
				Cwd:         "/usr/local/src",
				Group:       "scuba",
				User:        "pooja",
			},
			Output: []SplitData{
				{
					Vertex: processVertex{
						mapPropPairs(D_msgid, p{"pid", 42}, p{"cwd", "/usr/local/src"}, p{"user", "pooja"}, p{"group", "scuba"}),
					},
					EdgeSpecs: []EdgeSpec{
						SpecLocalLogic{"/path/to/sth"},
						M_envlink[1],
					},
				},
			},
		},
		{
			Summary: "Pid, envlink, two logic states, one single-proto net listener.",
			Input: interpret.Process{
				Pid:         42,
				Environment: M_envlink[0],
				LogicStates: []string{"/path/to/sth", "/usr/local/src/imaginationland"},
				Listen: []interpret.ListenAddr{
					{Type: "port", Port: 1025, Proto: []string{"tcp"}},
				},
			},
			Output: []SplitData{
				{
					Vertex: processVertex{mapPropPairs(D_msgid, p{"pid", 42})},
					EdgeSpecs: []EdgeSpec{
						SpecLocalLogic{"/path/to/sth"},
						SpecLocalLogic{"/usr/local/src/imaginationland"},
						interpret.ListenAddr{Type: "port", Port: 1025, Proto: []string{"tcp"}},
						M_envlink[0],
					},
				},
				{
					Vertex: commVertex{mapPropPairs(D_msgid, p{"port", 1025}, p{"proto", []string{"tcp"}}, p{"type", "port"})},
					EdgeSpecs: []EdgeSpec{
						M_envlink[0],
						SpecProc{42},
					},
				},
			},
		},
		{
			Summary: "Pid, envlink, one logic state, one local sock listener.",
			Input: interpret.Process{
				Pid:         42,
				Environment: M_envlink[0],
				LogicStates: []string{"/usr/local/src/imaginationland"},
				Listen: []interpret.ListenAddr{
					{Type: "unix", Path: "/var/run/lookitsa.sock"},
				},
			},
			Output: []SplitData{
				{
					Vertex: processVertex{mapPropPairs(D_msgid, p{"pid", 42})},
					EdgeSpecs: []EdgeSpec{
						SpecLocalLogic{"/usr/local/src/imaginationland"},
						interpret.ListenAddr{Type: "unix", Path: "/var/run/lookitsa.sock"},
						M_envlink[0],
					},
				},
				{
					Vertex: commVertex{mapPropPairs(D_msgid, p{"path", "/var/run/lookitsa.sock"}, p{"type", "unix"})},
					EdgeSpecs: []EdgeSpec{
						M_envlink[0],
						SpecProc{42},
					},
				},
			},
		},
	}

}

// ******** Utility funcs

func compareSplitData(expect, actual []SplitData, t *testing.T) {
	//fmt.Printf("%T %#v\n%T %#v\n\n", expect, expect, actual, actual)
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

		if !mapEq(esd.Vertex.Props(), asd.Vertex.Props(), t, true) {
			continue
		}
		if !assert.Equal(t, esd.EdgeSpecs, asd.EdgeSpecs, "EdgeSpecs are not equal") {
			//t.Errorf("Vertices not equal type mismatch at SplitData index %v: expected %T, actual %T", k, esd.Vertex, asd.Vertex)
			continue
		}
	}
}

func mapEq(expect, actual ps.Map, t *testing.T, emitErr bool) (match bool) {
	match = true
	if expect.Size() != actual.Size() {
		match = false
		if emitErr {
			t.Errorf("Prop maps are different sizes; guaranteed not equal. Expected size %v, actual %v", expect.Size(), actual.Size())
		}
	}

	expect.ForEach(func(k string, val ps.Any) {
		aval, exists := actual.Lookup(k)
		if !exists {
			match = false
			if emitErr {
				t.Errorf("Missing expected key '%v', expected value %v", k, val)
			}
			return
		}

		if emitErr {
			match = assert.Equal(t, val, aval, "Values for key '%v' are not equal: expected %v, actual %v", k, val, aval)
		} else {
			// keep it from erroring
			match = assert.Equal(new(testing.T), val, aval, "Values for key '%v' are not equal: expected %v, actual %v", k, val, aval)
		}
	})
	return
}

// ******** Actual tests

func TestSplitEnvironment(t *testing.T) {
	for _, fixture := range F_Environment {
		t.Log("Split test on environment fixture:", fixture.Summary)

		// by convention we're always using msgid 1 in fixtures
		sd, err := Split(fixture.Input, D_msgid)
		if err != nil {
			t.Error(err)
		}

		compareSplitData(fixture.Output, sd, t)
	}
}

func TestSplitLogicState(t *testing.T) {
	for _, fixture := range F_LogicState {
		t.Log("Split test on logic state fixture:", fixture.Summary)

		// by convention we're always using msgid 1 in fixtures
		sd, err := Split(fixture.Input, D_msgid)
		if err != nil {
			t.Error(err)
		}

		//fmt.Println(fixture.Summary)
		//fmt.Printf("%T %v\n%T %v\n\n", expect, actual, expect, actual)
		compareSplitData(fixture.Output, sd, t)
	}
}
func TestSplitProcess(t *testing.T) {
	for _, fixture := range F_Process {
		t.Log("Split test on process fixture:", fixture.Summary)

		// by convention we're always using msgid 1 in fixtures
		sd, err := Split(fixture.Input, D_msgid)
		if err != nil {
			t.Error(err)
		}

		compareSplitData(fixture.Output, sd, t)
	}
}
