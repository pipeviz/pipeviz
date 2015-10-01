package interpret

import (
	"github.com/tag1consulting/pipeviz/maputil"
	"github.com/tag1consulting/pipeviz/represent/helpers"
	"github.com/tag1consulting/pipeviz/represent/types"
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

func (d Environment) UnificationForm(id uint64) []types.UnifyInstructionForm {
	// seven distinct props
	v := types.NewVertex("environment", id,
		pp("os", d.OS),
		pp("provider", d.Provider),
		pp("type", d.Type),
		pp("nick", d.Nick),
		pp("hostname", d.Address.Hostname),
		pp("ipv4", d.Address.Ipv4),
		pp("ipv6", d.Address.Ipv6),
	)

	// By spec, Environments have no outbound edges
	return []types.UnifyInstructionForm{uif{v: v, u: envUnify}}
}

func envUnify(g types.CoreGraph, u types.UnifyInstructionForm) int {
	matches := g.VerticesWith(helpers.Qbv(types.VType("environment")))

	for _, e := range matches {
		if maputil.AnyMatch(e.Vertex.Properties, u.Vertex().Properties, "hostname", "ipv4", "ipv6") {
			return e.ID
		}
	}

	return 0
}

type EnvLink struct {
	Address Address `json:"address,omitempty"`
	Nick    string  `json:"nick,omitempty"`
}

func (spec EnvLink) Resolve(g types.CoreGraph, mid uint64, src types.VertexTuple) (e types.StdEdge, success bool) {
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

	rv := g.VerticesWith(helpers.Qbv(types.VType("environment")))
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
