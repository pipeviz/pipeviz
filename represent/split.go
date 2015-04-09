package represent

import (
	"errors"

	"github.com/mndrix/ps"
	"github.com/sdboyer/pipeviz/interpret"
)

type SplitData struct {
	Vertex
	EdgeSpecs
}

// TODO for now, no structure to this. change to queryish form later
type EdgeSpec interface{}
type EdgeSpecs []EdgeSpec

type SpecCommit struct {
	Sha1 []byte
}

type SpecLocalLogic struct {
	Path string
}

type SpecProc struct {
	Pid int
}

// TODO unused until plugging/codegen
type Splitter func(data interface{}, id int) ([]SplitData, error)

// TODO hardcoded for now, till code generation
func Split(d interface{}, id int) ([]SplitData, error) {
	switch v := d.(type) {
	case interpret.Environment:
		return splitEnvironment(v, id)
	case interpret.LogicState:
		return splitLogicState(v, id)
	case interpret.Process:
		return splitProcess(v, id)
		// TODO missing dataset, commit, and commitmeta
	}

	return nil, errors.New("No handler for object type")
}

func splitEnvironment(d interpret.Environment, id int) ([]SplitData, error) {
	// seven distinct props
	v := environmentVertex{props: ps.NewMap()}
	if d.Os != "" {
		v.props = v.props.Set("os", Property{MsgSrc: id, Value: d.Os})
	}
	if d.Provider != "" {
		v.props = v.props.Set("provider", Property{MsgSrc: id, Value: d.Provider})
	}
	if d.Type != "" {
		v.props = v.props.Set("type", Property{MsgSrc: id, Value: d.Type})
	}
	if d.Nickname != "" {
		v.props = v.props.Set("nickname", Property{MsgSrc: id, Value: d.Nickname})
	}
	if d.Address.Hostname != "" {
		v.props = v.props.Set("hostname", Property{MsgSrc: id, Value: d.Address.Hostname})
	}
	if d.Address.Ipv4 != "" {
		v.props = v.props.Set("ipv4", Property{MsgSrc: id, Value: d.Address.Ipv4})
	}
	if d.Address.Ipv6 != "" {
		v.props = v.props.Set("ipv6", Property{MsgSrc: id, Value: d.Address.Ipv6})
	}

	// By spec, Environments have no outbound edges
	return []SplitData{{Vertex: v}}, nil
}

func splitLogicState(d interpret.LogicState, id int) ([]SplitData, error) {
	v := logicStateVertex{props: ps.NewMap()}
	var edges EdgeSpecs

	// TODO do IDs need different handling?
	v.props = v.props.Set("path", Property{MsgSrc: id, Value: d.Path})

	if d.Lgroup != "" {
		// TODO should be an edge, a simple semantic one
		v.props = v.props.Set("lgroup", Property{MsgSrc: id, Value: d.Lgroup})
	}
	if d.Nick != "" {
		v.props = v.props.Set("nick", Property{MsgSrc: id, Value: d.Nick})
	}
	if d.Type != "" {
		v.props = v.props.Set("type", Property{MsgSrc: id, Value: d.Type})
	}

	// TODO should do anything with mutually exclusive properties here?
	if d.ID.Commit != "" {
		edges = append(edges, SpecCommit{[]byte(d.ID.Commit)})
	}
	// FIXME this shouldn't be here, it's a property of the commit
	if d.ID.Repository != "" {
		v.props = v.props.Set("repository", Property{MsgSrc: id, Value: d.ID.Repository})
	}
	if d.ID.Version != "" {
		v.props = v.props.Set("version", Property{MsgSrc: id, Value: d.ID.Version})
	}
	if d.ID.Semver != "" {
		v.props = v.props.Set("semver", Property{MsgSrc: id, Value: d.ID.Semver})
	}

	for _, dl := range d.Datasets {
		edges = append(edges, dl)
	}

	edges = append(edges, d.Environment)

	return []SplitData{{Vertex: v, EdgeSpecs: edges}}, nil
}

func splitProcess(d interpret.Process, id int) ([]SplitData, error) {
	sd := make([]SplitData, 0)

	v := processVertex{props: ps.NewMap()}
	var edges EdgeSpecs

	v.props = v.props.Set("pid", Property{MsgSrc: id, Value: d.Pid})
	if d.Cwd != "" {
		v.props = v.props.Set("Cwd", Property{MsgSrc: id, Value: d.Cwd})
	}
	if d.Group != "" {
		v.props = v.props.Set("Group", Property{MsgSrc: id, Value: d.Group})
	}
	if d.User != "" {
		v.props = v.props.Set("User", Property{MsgSrc: id, Value: d.User})
	}

	for _, ls := range d.LogicStates {
		edges = append(edges, SpecLocalLogic{ls})
	}

	for _, listen := range d.Listen {
		v2 := commVertex{props: ps.NewMap()}

		if listen.Type == "unix" {
			v2.props = v2.props.Set("path", listen.Path)
		} else {
			v2.props = v2.props.Set("port", listen.Port)
			// FIXME protos are a slice, wtf do we do about this
			v2.props = v2.props.Set("proto", listen.Proto)
		}
		v2.props = v2.props.Set("type", listen.Type)
		sd = append(sd, SplitData{v2, EdgeSpecs{d.Environment, SpecProc{d.Pid}}})

		edges = append(edges, listen)
	}

	edges = append(edges, d.Environment)
	return append([]SplitData{{Vertex: v, EdgeSpecs: edges}}, sd...), nil
}

// TODO can't do this till refactor interpret.Dataset to transform into something sane
//func splitDataset(d interpret.Dataset, id int) (VtxI, EdgeSpecs, error) {
//v := datasetVertex{props: ps.NewMap()}
//var edges EdgeSpecs

//// Props first
//if d.Name != "" {
//v.props = v.props.Set("Name", Property{MsgSrc: id, Value: d.Name})
//}
//if d.Name != "" {
//v.props = v.props.Set("Name", Property{MsgSrc: id, Value: d.Name})
//}
//}
