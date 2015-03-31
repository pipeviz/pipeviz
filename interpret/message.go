package interpret

import (
	"encoding/json"
	"fmt"

	"github.com/sdboyer/gogl"
)

type Message struct {
	Nodes []Node
	Edges []Edge
	Graph mGraph
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
		m.Graph = mGraph{make(map[int]VertexContainer), 0, 0}
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

	// All vertices are populated now. Try to link things up.
	for _, e := range tm.Ls {
		env, found := findEnv(tm.Env, e.Environment)
		if found != false {
			// No env found, assign false
			m.Graph.AddArcs(StandardEdge{e, false, "contained in", e.Environment})
		} else {
			m.Graph.AddArcs(StandardEdge{e, env, "contained in", e.Environment})
		}

		for _, ds := range e.Datasets {
			// TODO comparison against empty struct literal...works?
			if ds.ConnUnix != (ConnUnix{}) {
				// implicit link within env; can only work if we found an env
				if found != false {
					m.Graph.AddArcs(gogl.NewDataArc(e, false, ds))
				} else {

				}
			}
		}
	}

	return nil
}

func findEnv(envs []Environment, el EnvLink) (Environment, bool) {
	if el.Nick != "" {
		for _, e := range envs {
			if e.Nickname == el.Nick {
				return e, true
			}
		}
	} else if el.Address.Hostname != "" {
		for _, e := range envs {
			if e.Address.Hostname == el.Address.Hostname {
				return e, true
			}
		}
	} else if el.Address.Ipv4 != "" {
		for _, e := range envs {
			if e.Address.Ipv4 == el.Address.Ipv4 {
				return e, true
			}
		}
	} else if el.Address.Ipv6 != "" {
		for _, e := range envs {
			if e.Address.Ipv6 == el.Address.Ipv6 {
				return e, true
			}
		}
	}

	return Environment{}, false
}
