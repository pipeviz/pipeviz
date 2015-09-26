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

type environmentUIF struct{}

func (d Environment) UnificationForm(id uint64) []types.SplitData {
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
	return []types.SplitData{{Vertex: v}}
}

func (_ environmentUIF) Unify(g roCoreGraph, sd types.SplitData) int {
	matches := g.VerticesWith(helpers.Qbv(types.VType("environment")))

	for _, e := range matches {
		if maputil.AnyMatch(e.Vertex.Properties, sd.Vertex.Properties, "hostname", "ipv4", "ipv6") {
			return e.ID
		}
	}

	return 0
}
