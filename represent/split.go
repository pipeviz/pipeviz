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
	case interpret.YumPkg:
		return splitYumPkg(v, id)
	}

	return nil, errors.New("No handler for object type")
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
