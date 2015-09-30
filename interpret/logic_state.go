package interpret

import (
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/represent/helpers"
	"github.com/tag1consulting/pipeviz/represent/types"
)

type LogicState struct {
	Datasets    []DataLink `json:"datasets,omitempty"`
	Environment EnvLink    `json:"environment,omitempty"`
	ID          struct {
		Commit    Sha1   `json:"-"`
		CommitStr string `json:"commit,omitempty"`
		Version   string `json:"version,omitempty"`
		Semver    string `json:"semver,omitempty"`
	} `json:"id,omitempty"`
	Lgroup string `json:"lgroup,omitempty"`
	Nick   string `json:"nick,omitempty"`
	Path   string `json:"path,omitempty"`
	Type   string `json:"type,omitempty"`
}

type DataLink struct {
	Name        string   `json:"name,omitempty"`
	Type        string   `json:"type,omitempty"`
	Subset      string   `json:"subset,omitempty"`
	Interaction string   `json:"interaction,omitempty"`
	ConnUnix    ConnUnix `json:"connUnix,omitempty"`
	ConnNet     ConnNet  `json:"connNet,omitempty"`
}

func (d LogicState) UnificationForm(id uint64) []types.UnifyInstructionForm {
	v := types.NewVertex("logic-state", id,
		pp("path", d.Path),
		pp("lgroup", d.Lgroup),
		pp("nick", d.Nick),
		pp("type", d.Type),
		pp("version", d.ID.Version),
		pp("semver", d.ID.Semver),
	)

	var edges types.EdgeSpecs

	if !d.ID.Commit.IsEmpty() {
		edges = append(edges, SpecCommit{d.ID.Commit})
	}

	for _, dl := range d.Datasets {
		edges = append(edges, dl)
	}

	return []types.UnifyInstructionForm{uif{v: v, u: lsUnify, e: edges, se: types.EdgeSpecs{d.Environment}}}
}

func lsUnify(g types.CoreGraph, u types.UnifyInstructionForm) int {
	// only one scoping edge - the envlink
	edge, success := u.ScopingSpecs()[0].(EnvLink).Resolve(g, 0, emptyVT(u.Vertex()))
	if !success {
		return 0
	}

	vp := u.Vertex().Properties
	path, _ := vp.Lookup("path")
	lss := g.VerticesWith(helpers.Qbv(types.VType("logic-state"), "path", path.(types.Property).Value))

	for _, candidate := range lss {
		for _, edge2 := range g.OutWith(candidate.ID, helpers.Qbe(types.EType("envlink"))) {
			if edge2.Target == edge.Target {
				return candidate.ID
			}
		}
	}

	// no match
	return 0
}

type SpecCommit struct {
	Sha1 Sha1
}

func (spec SpecCommit) Resolve(g types.CoreGraph, mid uint64, src types.VertexTuple) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "version",
	}
	e.Props = e.Props.Set("sha1", types.Property{MsgSrc: mid, Value: spec.Sha1})

	re := g.OutWith(src.ID, helpers.Qbe(types.EType("version")))
	if len(re) > 0 {
		sha1, _ := re[0].Props.Lookup("sha1")
		e.ID = re[0].ID // FIXME setting the id to non-0 AND failing is currently unhandled
		if sha1.(types.Property).Value == spec.Sha1 {
			success = true
			e.Target = re[0].Target
		} else {
			rv := g.VerticesWith(helpers.Qbv(types.VType("commit"), "sha1", spec.Sha1))
			if len(rv) == 1 {
				success = true
				e.Target = rv[0].ID
			}
		}
	} else {
		rv := g.VerticesWith(helpers.Qbv(types.VType("commit"), "sha1", spec.Sha1))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}
