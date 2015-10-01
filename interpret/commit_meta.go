package interpret

import (
	"github.com/tag1consulting/pipeviz/represent/helpers"
	"github.com/tag1consulting/pipeviz/represent/types"
)

type CommitMeta struct {
	Sha1      Sha1
	Sha1Str   string   `json:"sha1,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Branches  []string `json:"branches,omitempty"`
	TestState string   `json:"testState,omitempty"`
}

func (d CommitMeta) UnificationForm(id uint64) []types.UnifyInstructionForm {
	ret := make([]types.UnifyInstructionForm, 0)

	for _, tag := range d.Tags {
		v := types.NewVertex("git-tag", id, pp("name", tag))
		ret = append(ret, uif{v: v, u: commitMetaUnify, se: []types.EdgeSpec{SpecCommit{d.Sha1}}})
	}

	for _, branch := range d.Branches {
		v := types.NewVertex("git-branch", id, pp("name", branch))
		ret = append(ret, uif{v: v, u: commitMetaUnify, se: []types.EdgeSpec{SpecCommit{d.Sha1}}})
	}

	if d.TestState != "" {
		v := types.NewVertex("test-result", id, pp("result", d.TestState))
		ret = append(ret, uif{v: v, u: commitMetaUnify, se: []types.EdgeSpec{SpecCommit{d.Sha1}}})
	}

	return ret
}

func commitMetaUnify(g types.CoreGraph, u types.UnifyInstructionForm) int {
	// the commit is the only scoping edge
	spec := u.ScopingSpecs()[0].(SpecCommit)
	_, success := spec.Resolve(g, 0, emptyVT(u.Vertex()))
	if !success {
		// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
		return 0
	}

	switch u.Vertex().Type {
	case "git-tag", "git-branch":
		name, _ := u.Vertex().Properties.Lookup("name")
		vtv := g.VerticesWith(helpers.Qbv(types.VType(u.Vertex().Type), "name", name.(types.Property).Value))
		if len(vtv) > 0 {
			return vtv[0].ID
		}

	case "test-result":
		// TODO no additional matching with test result...meaning that it's locked in to one result per commit
		for _, vt := range g.VerticesWith(helpers.Qbv(types.VType(u.Vertex().Type))) {
			if len(g.OutWith(vt.ID, helpers.Qbe(types.EType("version"), "sha1", spec.Sha1))) == 1 {
				return vt.ID
			}
		}
	}

	return 0
}
