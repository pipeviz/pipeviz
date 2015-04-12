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

type SpecLocalDataset struct {
	NamePath []string // path through the series of names that arrives at the final dataset
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
	case interpret.Commit:
		return splitCommit(v, id)
	case interpret.CommitMeta:
		return splitCommitMeta(v, id)
	case interpret.ParentDataset:
		return splitParentDataset(v, id)
	case interpret.Dataset:
		return splitDataset(v, id)
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
	if d.Nick != "" {
		v.props = v.props.Set("nick", Property{MsgSrc: id, Value: d.Nick})
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
			v2.props = v2.props.Set("path", Property{MsgSrc: id, Value: listen.Path})
		} else {
			v2.props = v2.props.Set("port", Property{MsgSrc: id, Value: listen.Port})
			// FIXME protos are a slice, wtf do we do about this
			v2.props = v2.props.Set("proto", Property{MsgSrc: id, Value: listen.Proto})
		}
		v2.props = v2.props.Set("type", listen.Type)
		sd = append(sd, SplitData{v2, EdgeSpecs{d.Environment, SpecProc{d.Pid}}})

		edges = append(edges, listen)
	}

	edges = append(edges, d.Environment)
	return append([]SplitData{{Vertex: v, EdgeSpecs: edges}}, sd...), nil
}

func splitCommit(d interpret.Commit, id int) ([]SplitData, error) {
	v := commitVertex{props: ps.NewMap()}

	v.props = v.props.Set("sha1", Property{MsgSrc: id, Value: d.Sha1})
	if d.Author != "" {
		v.props = v.props.Set("author", Property{MsgSrc: id, Value: d.Author})
	}
	if d.Date != "" {
		v.props = v.props.Set("date", Property{MsgSrc: id, Value: d.Date})
	}
	if d.Subject != "" {
		v.props = v.props.Set("subject", Property{MsgSrc: id, Value: d.Subject})
	}

	var edges EdgeSpecs
	for _, parent := range d.Parents {
		edges = append(edges, SpecCommit{parent})
	}

	return []SplitData{{Vertex: v, EdgeSpecs: edges}}, nil
}

func splitCommitMeta(d interpret.CommitMeta, id int) ([]SplitData, error) {
	sd := make([]SplitData, 0)

	for _, tag := range d.Tags {
		v := vcsLabelVertex{ps.NewMap()}
		v.props = v.props.Set("name", Property{MsgSrc: id, Value: tag})
		sd = append(sd, SplitData{Vertex: v, EdgeSpecs: []EdgeSpec{SpecCommit{d.Sha1}}})
	}

	if d.TestState != "" {
		v := testResultVertex{ps.NewMap()}
		v.props = v.props.Set("result", Property{MsgSrc: id, Value: d.TestState})
		sd = append(sd, SplitData{Vertex: v, EdgeSpecs: []EdgeSpec{SpecCommit{d.Sha1}}})
	}

	return sd, nil
}

func splitParentDataset(d interpret.ParentDataset, id int) ([]SplitData, error) {
	v := parentDatasetVertex{props: ps.NewMap()}
	var edges EdgeSpecs

	v.props = v.props.Set("name", Property{MsgSrc: id, Value: d.Name})
	if d.Path != "" {
		v.props = v.props.Set("path", Property{MsgSrc: id, Value: d.Path})
	}

	edges = append(edges, d.Environment)

	for _, sds := range d.Subsets {
		edges = append(edges, SpecLocalDataset{[]string{d.Name, sds.Name}})
	}

	return []SplitData{{v, edges}}, nil
}

func splitDataset(d interpret.Dataset, id int) ([]SplitData, error) {
	v := datasetVertex{props: ps.NewMap()}
	var edges EdgeSpecs

	v.props = v.props.Set("name", Property{MsgSrc: id, Value: d.Name})

	// TODO convert input from string to int and force timestamps. javascript apparently likes
	// ISO 8601, but go doesn't? so, timestamps.
	if d.CreateTime != "" {
		v.props = v.props.Set("create-time", Property{MsgSrc: id, Value: d.CreateTime})
	}

	edges = append(edges, SpecLocalDataset{[]string{d.Parent}})
	edges = append(edges, d.Genesis)

	return []SplitData{{v, edges}}, nil
}
