package semantic

import (
	"bytes"
	"encoding/hex"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/represent/q"
	"github.com/pipeviz/pipeviz/types/system"
)

func init() {
	if err := registerUnifier("commit", unifyCommit); err != nil {
		panic("commit vertex already registered")
	}
	if err := registerEdgeUnifier("parent-commit", eunifyGitCommitParent); err != nil {
		panic("parent-commit edge unifier already registered")
	}
	if err := registerResolver("parent-commit", resolveSpecGitCommitParent); err != nil {
		panic("parent-commit edge already registered")
	}
}

type Commit struct {
	Author     string   `json:"author,omitempty"`
	Date       string   `json:"date,omitempty"`
	ParentsStr []string `json:"parents"`
	Sha1       Sha1     `json:"-"`
	Sha1Str    string   `json:"sha1,omitempty"`
	Subject    string   `json:"subject,omitempty"`
	Repository string   `json:"repository,omitempty"`
}

type Sha1 [20]byte

func (s Sha1) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 42))

	_, err := buf.WriteString(`"`)
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(hex.EncodeToString(s[:]))
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(`"`)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// IsEmpty checks to see if the Sha1 is equal to the null Sha1 (20 zero bytes/40 ASCII zeroes)
func (s Sha1) IsEmpty() bool {
	// TODO this may have some odd semantic side effects, as git uses this, the "null sha1", to indicate a nonexistent head. not in any places pipeviz will forseeably interact with it, though...
	return s == [20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}

// String converts a sha1 to the standard 40-char lower-case hexadecimal representation.
func (s Sha1) String() string {
	return hex.EncodeToString(s[:])
}

func (d Commit) UnificationForm() []system.UnifyInstructionForm {
	byts, err := hex.DecodeString(d.Sha1Str)
	if err != nil {
		return nil
	}
	copy(d.Sha1[:], byts[0:20])

	v := pv{typ: "commit", props: system.RawProps{
		"sha1":       d.Sha1,
		"author":     d.Author,
		"date":       d.Date,
		"subject":    d.Subject,
		"repository": d.Repository,
	}}

	var edges []system.EdgeSpec

	for k, pstr := range d.ParentsStr {
		byts, err := hex.DecodeString(pstr)
		if err != nil {
			continue
		}

		var sha1 Sha1
		copy(sha1[:], byts[0:20])
		edges = append(edges, specGitCommitParent{Sha1: sha1, ParentNum: k + 1})
	}

	return []system.UnifyInstructionForm{uif{v: v, e: edges}}
}

func unifyCommit(g system.CoreGraph, u system.UnifyInstructionForm) uint64 {
	candidates := g.VerticesWith(q.Qbv(system.VType("commit"), "sha1", u.Vertex().Properties()["sha1"]))

	if len(candidates) > 0 { // there can be only one
		return candidates[0].ID
	}

	return 0
}

type specGitCommitParent struct {
	Sha1      Sha1
	ParentNum int
}

func eunifyGitCommitParent(vt system.VertexTuple, e system.EdgeSpec) uint64 {
	spec := e.(specGitCommitParent)
	return faofEdgeId(vt, q.Qbe("parent-commit", "pnum", spec.ParentNum))
}

func resolveSpecGitCommitParent(e system.EdgeSpec, g system.CoreGraph, mid uint64, src system.VertexTuple) (system.StdEdge, bool) {
	return e.(specGitCommitParent).Resolve(g, mid, src)
}

func (spec specGitCommitParent) Resolve(g system.CoreGraph, mid uint64, src system.VertexTuple) (e system.StdEdge, success bool) {
	e = system.StdEdge{
		Source:     src.ID,
		Props:      ps.NewMap(),
		Incomplete: true,
		EType:      "parent-commit",
	}

	re := g.OutWith(src.ID, q.Qbe(system.EType("parent-commit"), "pnum", spec.ParentNum))
	if len(re) > 0 {
		e.Incomplete, success = false, true
		e.Target = re[0].Target
		e.Props = re[0].Props
		// FIXME evidence of a problem here - since we're using pnum as the deduping identifier, there's no
		// way it could also sensibly change its MsgSrc value. This is very much a product of the intensional/extensional
		// identity problem: what does it mean to have the identifying data change? is it now a new thing? was it the old thing,
		// and it underwent a transition into the new thing? or is there no distinction between the old and new thing?
		e.Props = e.Props.Set("sha1", system.Property{MsgSrc: mid, Value: spec.Sha1})
		e.ID = re[0].ID
	} else {
		rv := g.VerticesWith(q.Qbv(system.VType("commit"), "sha1", spec.Sha1))
		if len(rv) == 1 {
			e.Incomplete, success = false, true
			e.Target = rv[0].ID
			e.Props = e.Props.Set("pnum", system.Property{MsgSrc: mid, Value: spec.ParentNum})
			e.Props = e.Props.Set("sha1", system.Property{MsgSrc: mid, Value: spec.Sha1})
		}
	}

	return
}

// Type indicates the EType the EdgeSpec will produce. This is necessarily invariant.
func (spec specGitCommitParent) Type() system.EType {
	return "parent-commit"
}
