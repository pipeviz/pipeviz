package represent

import (
	"errors"

	"github.com/sdboyer/pipeviz/interpret"
)

type EdgeSpec struct{}

// TODO unused until plugging/codegen
type Splitter func(data interface{}, id int) (Vertex, EdgeSpec, error)

// TODO hardcoded for now, till code generation
func Split(d interface{}, id int) (Vertex, EdgeSpec, error) {
	switch v := d.(type) {
	case interpret.Environment:
		return splitEnvironment(v, id)
	case interpret.LogicState:
		return splitLogicState(v, id)
	}

	return Vertex{nil}, EdgeSpec{}, errors.New("No handler for object type")
}

func splitEnvironment(d interpret.Environment, id int) (Vertex, EdgeSpec, error) {
	// seven distinct props
	v := Vertex{props: make([]Property, 0)}
	if d.Os != "" {
		v.props = append(v.props, Property{id, "os", d.Os})
	}
	if d.Provider != "" {
		v.props = append(v.props, Property{id, "provider", d.Provider})
	}
	if d.Type != "" {
		v.props = append(v.props, Property{id, "type", d.Type})
	}
	if d.Nickname != "" {
		v.props = append(v.props, Property{id, "nickname", d.Nickname})
	}
	if d.Address.Hostname != "" {
		v.props = append(v.props, Property{id, "hostname", d.Address.Hostname})
	}
	if d.Address.Ipv4 != "" {
		v.props = append(v.props, Property{id, "ipv4", d.Address.Ipv4})
	}
	if d.Address.Ipv6 != "" {
		v.props = append(v.props, Property{id, "ipv6", d.Address.Ipv6})
	}

	// By spec, Environments have no outbound edges
	return v, EdgeSpec{}, nil
}

func splitLogicState(d interpret.LogicState, id int) (Vertex, EdgeSpec, error) {
	v := Vertex{props: make([]Property, 0)}

	// TODO do IDs need different handling?
	v.props = append(v.props, Property{id, "path", d.Path})

	if d.Lgroup != "" {
		v.props = append(v.props, Property{id, "lgroup", d.Lgroup})
	}
	if d.Nick != "" {
		v.props = append(v.props, Property{id, "nick", d.Nick})
	}
	if d.Type != "" {
		v.props = append(v.props, Property{id, "type", d.Type})
	}

	// TODO should do anything with mutually exclusive properties here?
	if d.ID.Commit != "" {
		v.props = append(v.props, Property{id, "commit", d.ID.Commit})
	}
	if d.ID.Repository != "" {
		v.props = append(v.props, Property{id, "repository", d.ID.Repository})
	}
	if d.ID.Version != "" {
		v.props = append(v.props, Property{id, "version", d.ID.Version})
	}
	if d.ID.Semver != "" {
		v.props = append(v.props, Property{id, "semver", d.ID.Semver})
	}

	return v, EdgeSpec{}, nil
}
