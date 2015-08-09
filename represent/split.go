package represent

import (
	"errors"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/interpret"
)

type SplitData struct {
	Vertex    Vertex
	EdgeSpecs EdgeSpecs
}

// TODO for now, no structure to this. change to queryish form later
type EdgeSpec interface {
	// Given a graph and a vtTuple root, searches for an existing edge
	// that this EdgeSpec would supercede
	//FindExisting(*CoreGraph, vtTuple) (StandardEdge, bool)

	// Resolves the spec into a real edge, merging as appropriate with
	// any existing edge (as returned from FindExisting)
	//Resolve(*CoreGraph, vtTuple, int) (StandardEdge, bool)
}

type EdgeSpecs []EdgeSpec

type SpecCommit struct {
	Sha1 interpret.Sha1
}

type SpecGitCommitParent struct {
	Sha1      interpret.Sha1
	ParentNum int
}

type SpecLocalLogic struct {
	Path string
}

type SpecNetListener struct {
	Port  int
	Proto string
}

type SpecUnixDomainListener struct {
	Path string
}

type SpecParentDataset struct {
	Name string
}

type SpecDatasetHierarchy struct {
	NamePath []string // path through the series of names that arrives at the final dataset
}

// TODO unused until plugging/codegen
type Splitter func(data interface{}, id uint64) ([]SplitData, error)

// TODO hardcoded for now, till code generation
func Split(d interface{}, id uint64) ([]SplitData, error) {
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

func splitEnvironment(d interpret.Environment, id uint64) ([]SplitData, error) {
	// seven distinct props
	v := vertexEnvironment{props: ps.NewMap()}
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

func splitLogicState(d interpret.LogicState, id uint64) ([]SplitData, error) {
	v := vertexLogicState{props: ps.NewMap()}
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
	if !d.ID.Commit.IsEmpty() {
		edges = append(edges, SpecCommit{d.ID.Commit})
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

func splitProcess(d interpret.Process, id uint64) ([]SplitData, error) {
	sd := make([]SplitData, 0)

	v := vertexProcess{props: ps.NewMap()}
	var edges EdgeSpecs
	edges = append(edges, d.Environment)

	v.props = v.props.Set("pid", Property{MsgSrc: id, Value: d.Pid})
	if d.Cwd != "" {
		v.props = v.props.Set("cwd", Property{MsgSrc: id, Value: d.Cwd})
	}
	if d.Group != "" {
		v.props = v.props.Set("group", Property{MsgSrc: id, Value: d.Group})
	}
	if d.User != "" {
		v.props = v.props.Set("user", Property{MsgSrc: id, Value: d.User})
	}

	for _, ls := range d.LogicStates {
		edges = append(edges, SpecLocalLogic{ls})
	}

	if d.Dataset != "" {
		edges = append(edges, SpecParentDataset{Name: d.Dataset})
	}

	for _, listen := range d.Listen {
		// TODO change this to use diff vtx types for unix domain sock and network sock
		v2 := vertexComm{props: ps.NewMap()}

		if listen.Type == "unix" {
			edges = append(edges, SpecUnixDomainListener{Path: listen.Path})
			v2.props = v2.props.Set("path", Property{MsgSrc: id, Value: listen.Path})
		} else {
			for _, proto := range listen.Proto {
				edges = append(edges, SpecNetListener{Port: listen.Port, Proto: proto})
			}
			v2.props = v2.props.Set("port", Property{MsgSrc: id, Value: listen.Port})
		}
		v2.props = v2.props.Set("type", Property{MsgSrc: id, Value: listen.Type})
		sd = append(sd, SplitData{v2, EdgeSpecs{d.Environment}})
	}

	return append([]SplitData{{Vertex: v, EdgeSpecs: edges}}, sd...), nil
}

func splitCommit(d interpret.Commit, id uint64) ([]SplitData, error) {
	v := vertexCommit{props: ps.NewMap()}

	v.props = v.props.Set("sha1", Property{MsgSrc: id, Value: d.Sha1})
	v.props = v.props.Set("author", Property{MsgSrc: id, Value: d.Author})
	v.props = v.props.Set("date", Property{MsgSrc: id, Value: d.Date})
	v.props = v.props.Set("subject", Property{MsgSrc: id, Value: d.Subject})
	v.props = v.props.Set("repository", Property{MsgSrc: id, Value: d.Repository})

	var edges EdgeSpecs
	for k, parent := range d.Parents {
		edges = append(edges, SpecGitCommitParent{parent, k + 1})
	}

	return []SplitData{{Vertex: v, EdgeSpecs: edges}}, nil
}

func splitCommitMeta(d interpret.CommitMeta, id uint64) ([]SplitData, error) {
	sd := make([]SplitData, 0)

	for _, tag := range d.Tags {
		v := vertexGitTag{ps.NewMap()}
		v.props = v.props.Set("name", Property{MsgSrc: id, Value: tag})
		sd = append(sd, SplitData{Vertex: v, EdgeSpecs: []EdgeSpec{SpecCommit{d.Sha1}}})
	}

	for _, branch := range d.Branches {
		v := vertexGitBranch{ps.NewMap()}
		v.props = v.props.Set("name", Property{MsgSrc: id, Value: branch})
		sd = append(sd, SplitData{Vertex: v, EdgeSpecs: []EdgeSpec{SpecCommit{d.Sha1}}})
	}

	if d.TestState != "" {
		v := vertexTestResult{ps.NewMap()}
		v.props = v.props.Set("result", Property{MsgSrc: id, Value: d.TestState})
		sd = append(sd, SplitData{Vertex: v, EdgeSpecs: []EdgeSpec{SpecCommit{d.Sha1}}})
	}

	return sd, nil
}

func splitParentDataset(d interpret.ParentDataset, id uint64) ([]SplitData, error) {
	v := vertexParentDataset{props: ps.NewMap()}
	var edges EdgeSpecs

	v.props = v.props.Set("name", Property{MsgSrc: id, Value: d.Name})
	if d.Path != "" {
		v.props = v.props.Set("path", Property{MsgSrc: id, Value: d.Path})
	}

	edges = append(edges, d.Environment)

	ret := []SplitData{{v, edges}}

	for _, sub := range d.Subsets {
		sub.Parent = d.Name
		sds, _ := Split(sub, id)
		// can only be one
		sd := sds[0]
		// FIXME having this here is cool, but the schema does allow top-level datasets, which won't pass thru and so are guaranteed orphans
		sd.EdgeSpecs = append(EdgeSpecs{d.Environment}, sd.EdgeSpecs...)
		ret = append(ret, sd)
	}

	return ret, nil
}

func splitDataset(d interpret.Dataset, id uint64) ([]SplitData, error) {
	v := vertexDataset{props: ps.NewMap()}
	var edges EdgeSpecs

	v.props = v.props.Set("name", Property{MsgSrc: id, Value: d.Name})

	// TODO convert input from string to int and force timestamps. javascript apparently likes
	// ISO 8601, but go doesn't? so, timestamps.
	if d.CreateTime != "" {
		v.props = v.props.Set("create-time", Property{MsgSrc: id, Value: d.CreateTime})
	}

	edges = append(edges, SpecDatasetHierarchy{[]string{d.Parent}})
	edges = append(edges, d.Genesis)

	return []SplitData{{v, edges}}, nil
}
