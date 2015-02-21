package interpret

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sdboyer/gogl"
	"github.com/sdboyer/gogl/graph/al"
)

type MessageGraph interface {
	gogl.DataDigraph
	gogl.VertexSetMutator
	gogl.DataArcSetMutator
}

type Message struct {
	Nodes []Node
	Edges []Edge
	Graph MessageGraph
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

	if m.Graph == nil {
		m.Graph = gogl.Spec().Mutable().Directed().DataEdges().Create(al.G).(MessageGraph)
	}

	// first, dump all top-level objects into the graph.
	for _, e := range tm.Env {
		envlink := EnvLink{Address: Address{}}

		// Create an envlink for any nested items, preferring nick, then hostname, ipv4, ipv6.
		if e.Nickname != "" {
			envlink.Nick = e.Nickname
		} else if e.Address.Hostname != "" {
			envlink.Address.Hostname = e.Address.Hostname
		} else if e.Address.Ipv4 != "" {
			envlink.Address.Ipv4 = e.Address.Ipv4
		} else if e.Address.Ipv6 != "" {
			envlink.Address.Ipv6 = e.Address.Ipv6
		}

		// manage the little environment hierarchy
		for _, ls := range e.LogicStates {
			ls.Environment = envlink
			tm.Ls = append(tm.Ls, ls)
		}
		for _, p := range e.Processes {
			p.Environment = envlink
			tm.P = append(tm.P, p)
		}
		for _, ds := range e.Datasets {
			ds.Environment = envlink
			tm.Ds = append(tm.Ds, ds)
		}

		m.Graph.EnsureVertex(e)
	}
	for _, e := range tm.Ls {
		// TODO deduping/overwriting on ID needs to be done here
		m.Graph.EnsureVertex(e)
	}
	for _, e := range tm.Ds {
		// TODO deduping/overwriting on ID needs to be done here
		m.Graph.EnsureVertex(e)
	}
	for _, e := range tm.P {
		// TODO deduping/overwriting on ID needs to be done here
		m.Graph.EnsureVertex(e)
	}
	for _, e := range tm.C {
		m.Graph.EnsureVertex(e)
	}
	for _, e := range tm.Cm {
		m.Graph.EnsureVertex(e)
	}

	findEnv := func(el EnvLink) (Environment, error) {
		if el.Nick != "" {
			for _, e := range tm.Env {
				if e.Nickname == el.Nick {
					return e, nil
				}
			}
		} else if el.Address.Hostname != "" {
			for _, e := range tm.Env {
				if e.Address.Hostname == el.Address.Hostname {
					return e, nil
				}
			}
		} else if el.Address.Ipv4 != "" {
			for _, e := range tm.Env {
				if e.Address.Ipv4 == el.Address.Ipv4 {
					return e, nil
				}
			}
		} else if el.Address.Ipv6 != "" {
			for _, e := range tm.Env {
				if e.Address.Ipv6 == el.Address.Ipv6 {
					return e, nil
				}
			}
		}

		return Environment{}, errors.New("Env not found")
	}

	// All vertices are populated now. Try to link things up.
	for _, e := range tm.Ls {
		env, err := findEnv(e.Environment)
	}

	return nil
}

type Environment struct {
	Address     Address      `json:"address"`
	Os          string       `json:"os"`
	Provider    string       `json:"provider"`
	Type        string       `json:"type"`
	Nickname    string       `json:"nickname"`
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
