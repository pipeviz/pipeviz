package interpret

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Nodes []Node
	Edges []Edge
}

type Node interface{} // TODO for now
type Edge interface{} // TODO for now

func (m *Message) UnmarshalJSON(data []byte) error {
	tm := struct {
		Env []Environment `json:"environments"`
		Ls  []LogicState  `json:"logic-states"`
		Ds  []Dataset     `json:"datasets"`
		P   []Process     `json:"processes"`
		C   []Commit      `json:"commits"`
		Cm  []CommitMeta  `json:"commit-meta"`
	}{}

	err := json.Unmarshal(data, &tm)
	if err != nil {
		fmt.Println(err)
	}

	// first, dump all top-level objects into the node list. ugh type system that we can't append
	for _, e := range tm.Env {
		m.Nodes = append(m.Nodes, e)
	}
	for _, e := range tm.Ls {
		m.Nodes = append(m.Nodes, e)
	}
	for _, e := range tm.Ds {
		m.Nodes = append(m.Nodes, e)
	}
	for _, e := range tm.P {
		m.Nodes = append(m.Nodes, e)
	}
	for _, e := range tm.C {
		m.Nodes = append(m.Nodes, e)
	}
	for _, e := range tm.Cm {
		m.Nodes = append(m.Nodes, e)
	}

	// now do nested from env
	// FIXME all needs refactoring, but the important guarantee is order of interpretation, including nested
	// structures. this approach allows earlier nested structures to overwrite later top-level structures.
	for _, e := range tm.Env {
		for _, ne := range e.LogicStates {
			m.Nodes = append(m.Nodes, ne)
		}
		for _, ne := range e.Processes {
			m.Nodes = append(m.Nodes, ne)
		}
		for _, ne := range e.Datasets {
			m.Nodes = append(m.Nodes, ne)
		}
	}

	return nil
}

type Environment struct {
	Address     Address      `json:"address"`
	Os          string       `json:"os"`
	Provider    string       `json:"provider"`
	Type        string       `json:"type"`
	LogicStates []LogicState `json:"logic-states"`
	Datasets    []Dataset    `json:"datasets"`
	Processes   []Process    `json:"processes"`
}

type EnvLink struct {
	Address Address `json:"address"`
	Nick    string  `json:"nick"`
}

type Address struct {
	Hostname string `json:"hostname"`
	Ipv4     string `json:"ipv4"`
	Ipv6     string `json:"ipv6"`
}

type LogicState struct {
	Datasets []struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Subset      string `json:"subset"`
		Interaction string `json:"interaction"`
		Path        string `json:"path"`
		Rel         struct {
			Hostname string `json:"hostname"`
			Port     int    `json:"port"`
			Proto    string `json:"proto"`
			Type     string `json:"type"`
			Path     string `json:"path"`
		} `json:"rel"`
	} `json:"datasets"`
	Environment EnvLink `json:"environment"`
	ID          struct {
		Commit     string `json:"commit"`
		Repository string `json:"repository"`
		Version    string `json:"version"`
		Semver     string `json:"semver"`
	} `json:"id"`
	Lgroup string `json:"lgroup"`
	Nick   string `json:"nick"`
	Path   string `json:"path"`
	Type   string `json:"type"`
}

type CommitMeta struct {
	Sha1      []byte   `json:"sha1"`
	Tags      []string `json:"tags"`
	TestState string   `json:"testState"`
}

type Commit struct {
	Author  string   `json:"author"`
	Date    string   `json:"date"`
	Parents [][]byte `json:"parents"`
	Sha1    []byte   `json:"sha1"`
	Subject string   `json:"subject"`
}

type Dataset struct {
	Environment EnvLink `json:"environment"`
	Name        string  `json:"name"`
	Subsets     []struct {
		Name       string `json:"name"`
		CreateTime string `json:"create-time"`
		Genesis    struct {
			Address  Address  `json:"address"`
			Dataset  []string `json:"dataset"`
			SnapTime string   `json:"snap-time"`
		} `json:"genesis"`
	} `json:"subsets"`
}

type Process struct {
	Pid         int     `json:"pid"`
	Cwd         string  `json:"cwd"`
	Environment EnvLink `json:"environment"`
	Group       string  `json:"group"`
	Listen      []struct {
		Port  int      `json:"port"`
		Proto []string `json:"proto"`
		Type  string   `json:"type"`
		Path  string   `json:"path"`
	} `json:"listen"`
	LogicStates []string `json:"logic-states"`
	User        string   `json:"user"`
}
