package represent

import (
	"errors"

	"github.com/tag1consulting/pipeviz/interpret"
	"github.com/tag1consulting/pipeviz/represent/types"
)

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
type Splitter func(data interface{}, id uint64) ([]types.SplitData, error)

// TODO hardcoded for now, till code generation
func Split(d interface{}, id uint64) (interface{}, error) {
	// Temporary shim while converting types to UIF
	if u, ok := d.(types.Unifier); ok {
		return u.UnificationForm(id), nil
	}

	switch v := d.(type) {
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
	case interpret.YumPkg:
		return splitYumPkg(v, id)
	}

	return nil, errors.New("No handler for object type")
}

func splitEnvironment(d interpret.Environment, id uint64) ([]types.SplitData, error) {
	// seven distinct props
	v := types.NewVertex("environment", id,
		types.PropPair{K: "os", V: d.OS},
		types.PropPair{K: "provider", V: d.Provider},
		types.PropPair{K: "type", V: d.Type},
		types.PropPair{K: "nick", V: d.Nick},
		types.PropPair{K: "hostname", V: d.Address.Hostname},
		types.PropPair{K: "ipv4", V: d.Address.Ipv4},
		types.PropPair{K: "ipv6", V: d.Address.Ipv6},
	)

	// By spec, Environments have no outbound edges
	return []types.SplitData{{Vertex: v}}, nil
}

func splitLogicState(d interpret.LogicState, id uint64) ([]types.SplitData, error) {
	v := types.NewVertex("logic-state", id,
		types.PropPair{K: "path", V: d.Path},
		types.PropPair{K: "lgroup", V: d.Lgroup},
		types.PropPair{K: "nick", V: d.Nick},
		types.PropPair{K: "type", V: d.Type},
		types.PropPair{K: "version", V: d.ID.Version},
		types.PropPair{K: "semver", V: d.ID.Semver},
	)

	var edges types.EdgeSpecs

	if !d.ID.Commit.IsEmpty() {
		edges = append(edges, SpecCommit{d.ID.Commit})
	}

	for _, dl := range d.Datasets {
		edges = append(edges, dl)
	}

	edges = append(edges, d.Environment)

	return []types.SplitData{{Vertex: v, EdgeSpecs: edges}}, nil
}

func splitProcess(d interpret.Process, id uint64) ([]types.SplitData, error) {
	sd := make([]types.SplitData, 0)

	v := types.NewVertex("process", id,
		types.PropPair{K: "pid", V: d.Pid},
		types.PropPair{K: "cwd", V: d.Cwd},
		types.PropPair{K: "group", V: d.Group},
		types.PropPair{K: "user", V: d.User},
	)

	var edges types.EdgeSpecs
	edges = append(edges, d.Environment)

	for _, ls := range d.LogicStates {
		edges = append(edges, SpecLocalLogic{ls})
	}

	if d.Dataset != "" {
		edges = append(edges, SpecParentDataset{Name: d.Dataset})
	}

	for _, listen := range d.Listen {
		// TODO change this to use diff vtx types for unix domain sock and network sock
		v2 := types.NewVertex("comm", id,
			types.PropPair{K: "type", V: listen.Type},
		)

		if listen.Type == "unix" {
			edges = append(edges, SpecUnixDomainListener{Path: listen.Path})
			v2.Properties = v2.Properties.Set("path", types.Property{MsgSrc: id, Value: listen.Path})
		} else {
			for _, proto := range listen.Proto {
				edges = append(edges, SpecNetListener{Port: listen.Port, Proto: proto})
			}
			v2.Properties = v2.Properties.Set("port", types.Property{MsgSrc: id, Value: listen.Port})
		}
		sd = append(sd, types.SplitData{v2, types.EdgeSpecs{d.Environment}})
	}

	return append([]types.SplitData{{Vertex: v, EdgeSpecs: edges}}, sd...), nil
}

func splitCommit(d interpret.Commit, id uint64) ([]types.SplitData, error) {
	v := types.NewVertex("commit", id,
		types.PropPair{K: "sha1", V: d.Sha1},
		types.PropPair{K: "author", V: d.Author},
		types.PropPair{K: "date", V: d.Date},
		types.PropPair{K: "subject", V: d.Subject},
		types.PropPair{K: "repository", V: d.Repository},
	)

	var edges types.EdgeSpecs
	for k, parent := range d.Parents {
		edges = append(edges, SpecGitCommitParent{parent, k + 1})
	}

	return []types.SplitData{{Vertex: v, EdgeSpecs: edges}}, nil
}

func splitCommitMeta(d interpret.CommitMeta, id uint64) ([]types.SplitData, error) {
	sd := make([]types.SplitData, 0)

	for _, tag := range d.Tags {
		v := types.NewVertex("git-tag", id, types.PropPair{"name", tag})
		sd = append(sd, types.SplitData{Vertex: v, EdgeSpecs: []types.EdgeSpec{SpecCommit{d.Sha1}}})
	}

	for _, branch := range d.Branches {
		v := types.NewVertex("git-branch", id, types.PropPair{"name", branch})
		sd = append(sd, types.SplitData{Vertex: v, EdgeSpecs: []types.EdgeSpec{SpecCommit{d.Sha1}}})
	}

	if d.TestState != "" {
		v := types.NewVertex("test-result", id, types.PropPair{"result", d.TestState})
		sd = append(sd, types.SplitData{Vertex: v, EdgeSpecs: []types.EdgeSpec{SpecCommit{d.Sha1}}})
	}

	return sd, nil
}

func splitParentDataset(d interpret.ParentDataset, id uint64) ([]types.SplitData, error) {
	v := types.NewVertex("parent-dataset", id,
		types.PropPair{K: "name", V: d.Name},
		types.PropPair{K: "path", V: d.Path},
	)
	var edges types.EdgeSpecs

	edges = append(edges, d.Environment)

	ret := []types.SplitData{{v, edges}}

	for _, sub := range d.Subsets {
		sub.Parent = d.Name
		sdsi, _ := Split(sub, id)
		// can only be one
		sds := sdsi.([]types.SplitData)
		sd := sds[0]
		// FIXME having this here is cool, but the schema does allow top-level datasets, which won't pass thru and so are guaranteed orphans
		sd.EdgeSpecs = append(types.EdgeSpecs{d.Environment}, sd.EdgeSpecs...)
		ret = append(ret, sd)
	}

	return ret, nil
}

func splitDataset(d interpret.Dataset, id uint64) ([]types.SplitData, error) {
	v := types.NewVertex("dataset", id,
		types.PropPair{K: "name", V: d.Name},
		// TODO convert input from string to int and force timestamps. javascript apparently likes
		// ISO 8601, but go doesn't? so, timestamps.
		types.PropPair{K: "create-time", V: d.CreateTime},
	)
	var edges types.EdgeSpecs

	edges = append(edges, SpecDatasetHierarchy{[]string{d.Parent}})
	edges = append(edges, d.Genesis)

	return []types.SplitData{{Vertex: v, EdgeSpecs: edges}}, nil
}

func splitYumPkg(d interpret.YumPkg, id uint64) ([]types.SplitData, error) {
	v := types.NewVertex("dataset", id,
		types.PropPair{K: "name", V: d.Name},
		types.PropPair{K: "version", V: d.Version},
		types.PropPair{K: "epoch", V: d.Epoch},
		types.PropPair{K: "release", V: d.Release},
		types.PropPair{K: "arch", V: d.Arch},
	)

	return []types.SplitData{{Vertex: v}}, nil
}
