package semantic

import (
	"encoding/hex"

	"github.com/tag1consulting/pipeviz/represent/q"
	"github.com/tag1consulting/pipeviz/types/system"
)

type CommitMeta struct {
	Sha1Str   string   `json:"sha1,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Branches  []string `json:"branches,omitempty"`
	TestState string   `json:"testState,omitempty"`
}

func (d CommitMeta) UnificationForm(id uint64) []system.UnifyInstructionForm {
	ret := make([]system.UnifyInstructionForm, 0)

	var commit Sha1
	byts, err := hex.DecodeString(d.Sha1Str)
	if err != nil {
		return nil
	}
	copy(commit[:], byts[0:20])

	for _, tag := range d.Tags {
		v := system.NewVertex("git-tag", id, pp("name", tag))
		ret = append(ret, uif{v: v, u: commitMetaUnify, se: []system.EdgeSpec{SpecCommit{commit}}})
	}

	for _, branch := range d.Branches {
		v := system.NewVertex("git-branch", id, pp("name", branch))
		ret = append(ret, uif{v: v, u: commitMetaUnify, se: []system.EdgeSpec{SpecCommit{commit}}})
	}

	if d.TestState != "" {
		v := system.NewVertex("test-result", id, pp("result", d.TestState))
		ret = append(ret, uif{v: v, u: commitMetaUnify, se: []system.EdgeSpec{SpecCommit{commit}}})
	}

	return ret
}

func commitMetaUnify(g system.CoreGraph, u system.UnifyInstructionForm) int {
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
		vtv := g.VerticesWith(q.Qbv(system.VType(u.Vertex().Type), "name", name.(system.Property).Value))
		if len(vtv) > 0 {
			return vtv[0].ID
		}

	case "test-result":
		// TODO no additional matching with test result...meaning that it's locked in to one result per commit
		for _, vt := range g.VerticesWith(q.Qbv(system.VType(u.Vertex().Type))) {
			if len(g.OutWith(vt.ID, q.Qbe(system.EType("version"), "sha1", spec.Sha1))) == 1 {
				return vt.ID
			}
		}
	}

	return 0
}
