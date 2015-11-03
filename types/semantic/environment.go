package semantic

import (
	"github.com/pipeviz/pipeviz/maputil"
	"github.com/pipeviz/pipeviz/represent/q"
	"github.com/pipeviz/pipeviz/types/system"
)

type Environment struct {
	Address     Address         `json:"address,omitempty"`
	OS          string          `json:"os,omitempty"`
	Provider    string          `json:"provider,omitempty"`
	Type        string          `json:"type,omitempty"`
	Nick        string          `json:"nick,omitempty"`
	LogicStates []LogicState    `json:"logic-states,omitempty"`
	Datasets    []ParentDataset `json:"datasets,omitempty"`
	Processes   []Process       `json:"processes,omitempty"`
}

type Address struct {
	Hostname string `json:"hostname,omitempty"`
	Ipv4     string `json:"ipv4,omitempty"`
	Ipv6     string `json:"ipv6,omitempty"`
}

func (d Environment) UnificationForm() []system.UnifyInstructionForm {
	// seven distinct props
	ret := []system.UnifyInstructionForm{uif{
		v: pv{typ: "environment", props: system.RawProps{
			"os":       d.OS,
			"provider": d.Provider,
			"type":     d.Type,
			"nick":     d.Nick,
			"hostname": d.Address.Hostname,
			"ipv4":     d.Address.Ipv4,
			"ipv6":     d.Address.Ipv6,
		}},
		u: envUnify,
	}}

	envlink := EnvLink{Address: Address{}}
	// Create an envlink for any nested items, preferring nick, then hostname, ipv4, ipv6.
	if d.Nick != "" {
		envlink.Nick = d.Nick
	} else if d.Address.Hostname != "" {
		envlink.Address.Hostname = d.Address.Hostname
	} else if d.Address.Ipv4 != "" {
		envlink.Address.Ipv4 = d.Address.Ipv4
	} else if d.Address.Ipv6 != "" {
		envlink.Address.Ipv6 = d.Address.Ipv6
	}

	for _, ls := range d.LogicStates {
		ls.Environment = envlink
		ret = append(ret, ls.UnificationForm()...)
	}
	for _, p := range d.Processes {
		p.Environment = envlink
		ret = append(ret, p.UnificationForm()...)
	}
	for _, pds := range d.Datasets {
		pds.Environment = envlink
		ret = append(ret, pds.UnificationForm()...)
	}

	return ret
}

func envUnify(g system.CoreGraph, u system.UnifyInstructionForm) uint64 {
	matches := g.VerticesWith(q.Qbv(system.VType("environment")))

	for _, e := range matches {
		if maputil.AnyMatch(e.Vertex.Properties, u.Vertex().Properties(), "hostname", "ipv4", "ipv6") {
			return e.ID
		}
	}

	return 0
}

type EnvLink struct {
	Address Address `json:"address,omitempty"`
	Nick    string  `json:"nick,omitempty"`
}

func (spec EnvLink) Resolve(g system.CoreGraph, mid uint64, src system.VertexTuple) (e system.StdEdge, success bool) {
	_, e, success = findEnv(g, src)

	// Whether we find a match or not, have to merge in the EnvLink
	e.Props = maputil.FillPropMap(mid, false,
		pp("hostname", spec.Address.Hostname),
		pp("ipv4", spec.Address.Ipv4),
		pp("ipv6", spec.Address.Ipv6),
		pp("nick", spec.Nick),
	)

	// If we already found the matching edge, bail out now
	if success {
		return
	}

	rv := g.VerticesWith(q.Qbv(system.VType("environment")))
	for _, vt := range rv {
		// TODO this'll be cross-package eventually - reorg needed
		if maputil.AnyMatch(e.Props, vt.Vertex.Properties, "nick", "hostname", "ipv4", "ipv6") {
			success = true
			e.Target = vt.ID
			break
		}
	}

	return
}
