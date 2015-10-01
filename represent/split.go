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

	return nil, errors.New("No handler for object type")
}
