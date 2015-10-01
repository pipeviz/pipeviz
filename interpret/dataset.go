package interpret

import (
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/represent/helpers"
	"github.com/tag1consulting/pipeviz/represent/types"
)

// FIXME this metaset/set design is not recursive, but it will need to be
type ParentDataset struct {
	Environment EnvLink   `json:"environment,omitempty"`
	Path        string    `json:"path,omitempty"`
	Name        string    `json:"name,omitempty"`
	Subsets     []Dataset `json:"subsets,omitempty"`
}

type Dataset struct {
	Name        string      `json:"name,omitempty"`
	Parent      string      // TODO yechhhh...how do we qualify the hierarchy?
	CreateTime  string      `json:"create-time,omitempty"`
	Genesis     DataGenesis `json:"genesis,omitempty"`
	Environment EnvLink     `json:"-"` // hidden from json output; just used by parent
}

type DataGenesis interface {
	_dg() // dummy method, avoid propagating the interface
}

// TODO this form implies a snap that only existed while 'in flight' - that is, no
// dump artfiact that exists on disk anywhere. Need to incorporate that case, though.
type DataProvenance struct {
	Address  Address  `json:"address,omitempty"`
	Dataset  []string `json:"dataset,omitempty"`
	SnapTime string   `json:"snap-time,omitempty"`
}

type DataAlpha string

func (d DataAlpha) _dg()      {}
func (d DataProvenance) _dg() {}

func (d Dataset) UnificationForm(id uint64) []types.UnifyInstructionForm {
	v := types.NewVertex("dataset", id,
		types.PropPair{K: "name", V: d.Name},
		// TODO convert input from string to int and force timestamps. javascript apparently likes
		// ISO 8601, but go doesn't? so, timestamps.
		types.PropPair{K: "create-time", V: d.CreateTime},
	)
	var edges types.EdgeSpecs

	edges = append(edges, d.Genesis)

	return []types.UnifyInstructionForm{uif{
		v: v,
		u: datasetUnify,
		se: types.EdgeSpecs{SpecDatasetHierarchy{
			Environment: d.Environment,
			NamePath:    []string{d.Parent},
		}},
		e: edges,
	}}
}

func datasetUnify(g types.CoreGraph, u types.UnifyInstructionForm) int {
	name, _ := u.Vertex().Properties.Lookup("name")
	vtv := g.VerticesWith(helpers.Qbv(types.VType("dataset"), "name", name.(types.Property).Value))
	if len(vtv) == 0 {
		return 0
	}

	spec := u.ScopingSpecs()[0].(SpecDatasetHierarchy)
	el, success := spec.Environment.Resolve(g, 0, emptyVT(u.Vertex()))
	//pretty.Println(spec, el, success)
	// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
	if success {
		for _, vt := range vtv {
			//pretty.Println(vt.Flat())
			//for _, tmp := range g.SuccessorsWith(vt.ID, helpers.Qbe(types.EType("dataset-hierarchy"))) {
			//pretty.Println(tmp.Flat())
			//}
			if id := hasMatchingEnv(g, el, g.SuccessorsWith(vt.ID, helpers.Qbe(types.EType("dataset-hierarchy")))); id != 0 {
				return vt.ID
			}
		}
	}

	return 0
}

func (d ParentDataset) UnificationForm(id uint64) []types.UnifyInstructionForm {
	ret := []types.UnifyInstructionForm{uif{
		v: types.NewVertex("parent-dataset", id,
			types.PropPair{K: "name", V: d.Name},
			types.PropPair{K: "path", V: d.Path},
		),
		u:  parentDatasetUnify,
		se: []types.EdgeSpec{d.Environment},
	}}

	// TODO make recursive. which also means getting rid of the whole parent type...
	for _, sub := range d.Subsets {
		sub.Parent = d.Name
		sub.Environment = d.Environment
		ret = append(ret, sub.UnificationForm(id)...)
	}

	return ret
}

func parentDatasetUnify(g types.CoreGraph, u types.UnifyInstructionForm) int {
	edge, success := u.ScopingSpecs()[0].(EnvLink).Resolve(g, 0, emptyVT(u.Vertex()))
	if !success {
		// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
		return 0
	}

	path, _ := u.Vertex().Properties.Lookup("path")
	name, _ := u.Vertex().Properties.Lookup("name")
	return hasMatchingEnv(g, edge, g.VerticesWith(helpers.Qbv(types.VType("parent-dataset"),
		"path", path.(types.Property).Value,
		"name", name.(types.Property).Value,
	)))

	//path, _ := u.Vertex().Properties.Lookup("path")
	//name, _ := u.Vertex().Properties.Lookup("name")
	//vtv := g.VerticesWith(helpers.Qbv(types.VType("parent-dataset"),
	//"path", path.(types.Property).Value,
	//"name", name.(types.Property).Value,
	//))
	//pretty.Println(path, name, edge)
	//for _, vt := range vtv {
	//pretty.Println(vt.Flat())
	//}
	//id := hasMatchingEnv(g, edge, g.VerticesWith(helpers.Qbv(types.VType("parent-dataset"),
	//"path", path.(types.Property).Value,
	//"name", name.(types.Property).Value,
	//)))
	//fmt.Println("RETURNING ID:", id)
	//return id
}

type SpecDatasetHierarchy struct {
	Environment EnvLink
	NamePath    []string // path through the series of names that arrives at the final dataset
}

func (spec SpecDatasetHierarchy) Resolve(g types.CoreGraph, mid uint64, src types.VertexTuple) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "dataset-hierarchy",
	}
	e.Props = e.Props.Set("parent", types.Property{MsgSrc: mid, Value: spec.NamePath[0]})

	// check for existing link - there can be only be one
	re := g.OutWith(src.ID, helpers.Qbe(types.EType("dataset-hierarchy")))
	if len(re) == 1 {
		success = true
		e = re[0]
		// TODO semantics should preclude this from being able to change, but doing it dirty means force-setting it anyway for now
		e.Props = e.Props.Set("parent", types.Property{MsgSrc: mid, Value: spec.NamePath[0]})
		return
	}

	// no existing link found; search for proc directly
	envlink, success := spec.Environment.Resolve(g, 0, emptyVT(src.Vertex))
	if success {
		rv := g.PredecessorsWith(envlink.Target, helpers.Qbv(types.VType("parent-dataset"), "name", spec.NamePath[0]))
		if len(rv) != 0 { // >1 shouldn't be possible
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}
