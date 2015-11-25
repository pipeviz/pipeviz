package semantic

import (
	"encoding/hex"

	"github.com/pipeviz/pipeviz/represent/q"
	"github.com/pipeviz/pipeviz/types/system"
)

func init() {
	if err := registerUnifier("git-tag", unifyCommitMeta); err != nil {
		panic("git-tag vertex already registered")
	}
	if err := registerUnifier("git-branch", unifyCommitMeta); err != nil {
		panic("git-branch vertex already registered")
	}
	if err := registerUnifier("git-result", unifyCommitMeta); err != nil {
		panic("git-result vertex already registered")
	}
}

type CommitMeta struct {
	Sha1Str   string   `json:"sha1,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Branches  []string `json:"branches,omitempty"`
	TestState string   `json:"testState,omitempty"`
}

func (d CommitMeta) UnificationForm() []system.UnifyInstructionForm {
	ret := make([]system.UnifyInstructionForm, 0)

	var commit Sha1
	byts, err := hex.DecodeString(d.Sha1Str)
	if err != nil {
		return nil
	}
	copy(commit[:], byts[0:20])

	for _, tag := range d.Tags {
		v := pv{typ: "git-tag", props: system.RawProps{"name": tag}}
		ret = append(ret, uif{v: v, se: []system.EdgeSpec{specCommit{commit}}})
	}

	for _, branch := range d.Branches {
		v := pv{typ: "git-branch", props: system.RawProps{"name": branch}}
		ret = append(ret, uif{v: v, se: []system.EdgeSpec{specCommit{commit}}})
	}

	if d.TestState != "" {
		v := pv{typ: "git-result", props: system.RawProps{"result": d.TestState}}
		ret = append(ret, uif{v: v, se: []system.EdgeSpec{specCommit{commit}}})
	}

	return ret
}

func unifyCommitMeta(g system.CoreGraph, u system.UnifyInstructionForm) uint64 {
	// the commit is the only scoping edge
	spec := u.ScopingSpecs()[0].(specCommit)
	_, success := spec.Resolve(g, 0, emptyVT(u.Vertex()))
	if !success {
		// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
		return 0
	}

	switch u.Vertex().Type() {
	case "git-tag", "git-branch":
		vtv := g.VerticesWith(q.Qbv(system.VType(u.Vertex().Type()), "name", u.Vertex().Properties()["name"]))
		if len(vtv) > 0 {
			return vtv[0].ID
		}

	case "test-result":
		// TODO no additional matching with test result...meaning that it's locked in to one result per commit
		for _, vt := range g.VerticesWith(q.Qbv(system.VType(u.Vertex().Type()))) {
			if len(g.OutWith(vt.ID, q.Qbe(system.EType("version"), "sha1", spec.Sha1))) == 1 {
				return vt.ID
			}
		}
	}

	return 0
}
