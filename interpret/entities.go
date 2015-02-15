package interpret

import (
	"encoding/json"
)

type Message struct {
	Nodes []Node
	Edges []Edge
}

type Node interface{} // TODO for now
type Edge interface{} // TODO for now

func (m Message) UnmarshalJSON(data []byte) error {
	tm := struct {
		env []Environment `json:"environments"`
		ls  []LogicState  `json:"logic-states"`
		ds  []Dataset     `json:"datasets"`
		p   []Process     `json:"processes"`
		c   []Commit      `json:"commits"`
		cm  []CommitMeta  `json:"commit-meta"`
	}{}

	json.Unmarshal(data, tm)

	return nil
}

type Environment struct {
	Address  Address `json:"address"`
	Os       string  `json:"os"`
	Provider string  `json:"provider"`
	Type     string  `json:"type"`
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
	Datasets    []Dataset `json:"datasets"`
	Environment EnvLink   `json:"environment"`
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
	Sha1      string   `json:"sha1"`
	Tags      []string `json:"tags"`
	TestState string   `json:"testState"`
}

type Commit struct {
	Author  string     `json:"author"`
	Date    string     `json:"date"`
	Parents [][40]byte `json:"parents"`
	Sha1    [40]byte   `json:"sha1"`
	Subject string     `json:"subject"`
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
		GenesisSelf string `json:"genesis"`
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
		Path  string   `json:"type"`
	} `json:"listen"`
	LogicStates []string `json:"logic-states"`
	User        string   `json:"user"`
}
