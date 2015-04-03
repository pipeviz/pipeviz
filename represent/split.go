package represent

import (
	"errors"

	"github.com/sdboyer/pipeviz/interpret"
)

// TODO for now, no structure to this. change to queryish form later
type EdgeSpec interface{}
type EdgeSpecs []EdgeSpec

type SpecCommit struct {
	Sha1 []byte
}

type SpecLocalLogic struct {
	Path string
}

// TODO unused until plugging/codegen
type Splitter func(data interface{}, id int) (Vertex, EdgeSpecs, error)

// TODO hardcoded for now, till code generation
func Split(d interface{}, id int) (Vertex, EdgeSpecs, error) {
	switch v := d.(type) {
	case interpret.Environment:
		return splitEnvironment(v, id)
	case interpret.LogicState:
		return splitLogicState(v, id)
	case interpret.Process:
		return splitProcess(v, id)
		// TODO missing dataset, commit, and commitmeta
	}

	return Vertex{nil}, nil, errors.New("No handler for object type")
}

func splitEnvironment(d interpret.Environment, id int) (Vertex, EdgeSpecs, error) {
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
	return v, nil, nil
}

func splitLogicState(d interpret.LogicState, id int) (Vertex, EdgeSpecs, error) {
	v := Vertex{props: make([]Property, 0)}
	var edges EdgeSpecs

	// TODO do IDs need different handling?
	v.props = append(v.props, Property{id, "path", d.Path})

	if d.Lgroup != "" {
		// TODO should be an edge, a simple semantic one
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
		edges = append(edges, SpecCommit{[]byte(d.ID.Commit)})
	}
	// FIXME this shouldn't be here, it's a property of the commit
	if d.ID.Repository != "" {
		v.props = append(v.props, Property{id, "repository", d.ID.Repository})
	}
	if d.ID.Version != "" {
		v.props = append(v.props, Property{id, "version", d.ID.Version})
	}
	if d.ID.Semver != "" {
		v.props = append(v.props, Property{id, "semver", d.ID.Semver})
	}

	for _, dl := range d.Datasets {
		edges = append(edges, dl)
	}

	edges = append(edges, d.Environment)

	return v, edges, nil
}
func splitProcess(d interpret.Process, id int) (Vertex, EdgeSpecs, error) {
	v := Vertex{props: make([]Property, 0)}
	var edges EdgeSpecs

	v.props = append(v.props, Property{id, "pid", d.Pid})
	if d.Cwd != "" {
		v.props = append(v.props, Property{id, "Cwd", d.Cwd})
	}
	if d.Group != "" {
		v.props = append(v.props, Property{id, "Group", d.Group})
	}
	if d.User != "" {
		v.props = append(v.props, Property{id, "User", d.User})
	}

	for _, ls := range d.LogicStates {
		edges = append(edges, SpecLocalLogic{ls})
	}

	for _, listen := range d.Listen {
		edges = append(edges, listen)
	}

	edges = append(edges, d.Environment)

	return v, edges, nil
}

// TODO can't do this till refactor interpret.Dataset to transform into something sane
//func splitDataset(d interpret.Dataset, id int) (Vertex, EdgeSpecs, error) {
//v := Vertex{props: make([]Property, 0)}
//var edges EdgeSpecs

//// Props first
//if d.Name != "" {
//v.props = append(v.props, Property{id, "Name", d.Name})
//}
//if d.Name != "" {
//v.props = append(v.props, Property{id, "Name", d.Name})
//}
//}
