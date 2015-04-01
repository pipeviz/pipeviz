package interpret

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Id  int
	Env []Environment `json:"environments"`
	Ls  []LogicState  `json:"logic-states"`
	Ds  []Dataset     `json:"datasets"`
	P   []Process     `json:"processes"`
	C   []Commit      `json:"commits"`
	Cm  []CommitMeta  `json:"commit-meta"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &m)
	if err != nil {
		// FIXME logging
		fmt.Println(err)
	}

	// TODO separate all of this into pluggable/generated structures
	// first, dump all top-level objects into the graph.
	for _, e := range m.Env {
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
			m.Ls = append(m.Ls, ls)
		}
		for _, p := range e.Processes {
			p.Environment = envlink
			m.P = append(m.P, p)
		}
		for _, ds := range e.Datasets {
			ds.Environment = envlink
			m.Ds = append(m.Ds, ds)
		}
	}

	// All vertices are populated now. Try to link things up.
	//for _, e := range m.Ls {
	//env, found := findEnv(m.Env, e.Environment)
	//if found != false {
	//// No env found, assign false
	//m.Graph.AddArcs(StandardEdge{e, false, "contained in", e.Environment})
	//} else {
	//m.Graph.AddArcs(StandardEdge{e, env, "contained in", e.Environment})
	//}

	//for _, ds := range e.Datasets {
	//// TODO comparison against empty struct literal...works?
	//if ds.ConnUnix != (ConnUnix{}) {
	//// implicit link within env; can only work if we found an env
	//if found != false {
	//m.Graph.AddArcs(gogl.NewDataArc(e, false, ds))
	//} else {

	//}
	//}
	//}
	//}

	return nil
}

func (m *Message) Each(f func(vertex interface{})) {
	for _, e := range m.Env {
		f(e)
	}
	for _, e := range m.Ls {
		f(e)
	}
	for _, e := range m.Ds {
		f(e)
	}
	for _, e := range m.P {
		f(e)
	}
	for _, e := range m.C {
		f(e)
	}
	for _, e := range m.Cm {
		f(e)
	}
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
