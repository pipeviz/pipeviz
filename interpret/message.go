package interpret

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Message struct {
	Id int
	m  *message
}

type message struct {
	Env []Environment   `json:"environments"`
	Ls  []LogicState    `json:"logic-states"`
	Pds []ParentDataset `json:"datasets"`
	Ds  []Dataset
	P   []Process    `json:"processes"`
	C   []Commit     `json:"commits"`
	Cm  []CommitMeta `json:"commit-meta"`
}

func (m *Message) UnmarshalJSON(data []byte) error {
	m.m = new(message)
	err := json.Unmarshal(data, m.m)
	if err != nil {
		// FIXME logging
		fmt.Println(err)
	}

	// TODO separate all of this into pluggable/generated structures
	// first, dump all top-level objects into the graph.
	for _, e := range m.m.Env {
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
			m.m.Ls = append(m.m.Ls, ls)
		}
		for _, p := range e.Processes {
			p.Environment = envlink
			m.m.P = append(m.m.P, p)
		}
		for _, pds := range e.Datasets {
			for _, ds := range pds.Subsets {
				// FIXME this doesn't really work as a good linkage
				ds.Parent = pds.Name
				m.m.Ds = append(m.m.Ds, ds)
			}

			pds.Environment = envlink
			m.m.Pds = append(m.m.Pds, pds)
		}
	}

	return nil
}

func (m *Message) Each(f func(vertex interface{})) {
	for _, e := range m.m.Env {
		f(e)
	}
	for _, e := range m.m.Ls {
		f(e)
	}
	for _, e := range m.m.Pds {
		f(e)
	}
	for _, e := range m.m.Ds {
		f(e)
	}
	for _, e := range m.m.P {
		f(e)
	}
	for _, e := range m.m.C {
		f(e)
	}
	for _, e := range m.m.Cm {
		f(e)
	}
}

// Unmarshalers for variant subtypes

// Unmarshaling a Dataset involves resolving whether it has α genesis (string), or
// a provenancial one (struct). So we have to decode directly, here.
func (ds *Dataset) UnmarshalJSON(data []byte) (err error) {
	type αDataset struct {
		Name       string    `json:"name"`
		CreateTime string    `json:"create-time"`
		Genesis    DataAlpha `json:"genesis"`
	}
	type provDataset struct {
		Name       string         `json:"name"`
		CreateTime string         `json:"create-time"`
		Genesis    DataProvenance `json:"genesis"`
	}

	a, b := αDataset{}, provDataset{}
	// use α first, as that can match the case where it's not specified (though schema
	// currently does not allow that)
	if err = json.Unmarshal(data, &a); err == nil {
		ds.Name, ds.CreateTime, ds.Genesis = a.Name, a.CreateTime, a.Genesis
	} else if err = json.Unmarshal(data, &b); err == nil {
		ds.Name, ds.CreateTime, ds.Genesis = b.Name, b.CreateTime, b.Genesis
	} else {
		err = errors.New("JSON genesis did not match either alpha or provenancial forms.")
	}

	return err
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
