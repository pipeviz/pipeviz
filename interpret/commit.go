package interpret

import (
	"bytes"
	"encoding/hex"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/represent/helpers"
	"github.com/tag1consulting/pipeviz/represent/types"
)

type Commit struct {
	Author     string `json:"author,omitempty"`
	Date       string `json:"date,omitempty"`
	Parents    []Sha1
	ParentsStr []string `json:"parents,omitempty"`
	Sha1       Sha1
	Sha1Str    string `json:"sha1,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Repository string `json:"repository,omitempty"`
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

func (d Commit) UnificationForm(id uint64) []types.UnifyInstructionForm {
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

	return []types.UnifyInstructionForm{uif{v: v, u: commitUnify, e: edges}}
}

func commitUnify(g types.CoreGraph, u types.UnifyInstructionForm) int {
	sha1, _ := u.Vertex().Properties.Lookup("sha1")
	candidates := g.VerticesWith(helpers.Qbv(types.VType("commit"), "sha1", sha1.(types.Property).Value))

	if len(candidates) > 0 { // there can be only one
		return candidates[0].ID
	}

	return 0
}

type SpecGitCommitParent struct {
	Sha1      Sha1
	ParentNum int
}

func (spec SpecGitCommitParent) Resolve(g types.CoreGraph, mid uint64, src types.VertexTuple) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "parent-commit",
	}

	re := g.OutWith(src.ID, helpers.Qbe(types.EType("parent-commit"), "pnum", spec.ParentNum))
	if len(re) > 0 {
		success = true
		e.Target = re[0].Target
		e.ID = re[0].ID
	} else {
		rv := g.VerticesWith(helpers.Qbv(types.VType("commit"), "sha1", spec.Sha1))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].ID
			e.Props = e.Props.Set("pnum", types.Property{MsgSrc: mid, Value: spec.ParentNum})
			e.Props = e.Props.Set("sha1", types.Property{MsgSrc: mid, Value: spec.Sha1})
		}
	}

	return
}
