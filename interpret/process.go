package interpret

import (
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/represent/helpers"
	"github.com/tag1consulting/pipeviz/represent/types"
)

type Process struct {
	Pid         int          `json:"pid,omitempty"`
	Cwd         string       `json:"cwd,omitempty"`
	Dataset     string       `json:"dataset,omitempty"`
	Environment EnvLink      `json:"environment,omitempty"`
	Group       string       `json:"group,omitempty"`
	Listen      []ListenAddr `json:"listen,omitempty"`
	LogicStates []string     `json:"logic-states,omitempty"`
	User        string       `json:"user,omitempty"`
}

type ListenAddr struct {
	Port  int      `json:"port,omitempty"`
	Proto []string `json:"proto,omitempty"`
	Type  string   `json:"type,omitempty"`
	Path  string   `json:"path,omitempty"`
}

func (d Process) UnificationForm(id uint64) []types.UnifyInstructionForm {
	ret := make([]types.UnifyInstructionForm, 0)

	v := types.NewVertex("process", id,
		types.PropPair{K: "pid", V: d.Pid},
		types.PropPair{K: "cwd", V: d.Cwd},
		types.PropPair{K: "group", V: d.Group},
		types.PropPair{K: "user", V: d.User},
	)

	var edges types.EdgeSpecs

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
		ret = append(ret, uif{v: v2, u: commUnify, se: types.EdgeSpecs{d.Environment}})
	}

	return append([]types.UnifyInstructionForm{uif{
		v:  v,
		u:  processUnify,
		e:  edges,
		se: []types.EdgeSpec{d.Environment},
	}}, ret...)
}

func processUnify(g types.CoreGraph, u types.UnifyInstructionForm) int {
	// only one scoping edge - the envlink
	edge, success := u.ScopingSpecs()[0].(EnvLink).Resolve(g, 0, emptyVT(u.Vertex()))
	if !success {
		// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
		return 0
	}

	pid, _ := u.Vertex().Properties.Lookup("pid")
	return hasMatchingEnv(g, edge, g.VerticesWith(helpers.Qbv(types.VType("process"), "pid", pid.(types.Property).Value)))
}

func commUnify(g types.CoreGraph, u types.UnifyInstructionForm) int {
	// only one scoping edge - the envlink
	edge, success := u.ScopingSpecs()[0].(EnvLink).Resolve(g, 0, emptyVT(u.Vertex()))
	if !success {
		// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
		return 0
	}

	vp := u.Vertex().Properties
	typ, _ := vp.Lookup("type")
	path, haspath := vp.Lookup("path")
	if haspath {
		return hasMatchingEnv(g, edge, g.VerticesWith(helpers.Qbv(types.VType("comm"),
			"type", typ.(types.Property).Value,
			"path", path.(types.Property).Value)))
	} else {
		port, _ := vp.Lookup("port")
		return hasMatchingEnv(g, edge, g.VerticesWith(helpers.Qbv(types.VType("comm"),
			"type", typ.(types.Property).Value,
			"port", port.(types.Property).Value)))
	}
}

type SpecLocalLogic struct {
	Path string
}

func (spec SpecLocalLogic) Resolve(g types.CoreGraph, mid uint64, src types.VertexTuple) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "logic-link",
	}

	// search for existing link
	re := g.OutWith(src.ID, helpers.Qbe(types.EType("logic-link"), "path", spec.Path))
	if len(re) == 1 {
		// TODO don't set the path prop again, it's the unique id...meh, same question here w/uniqueness as above
		success = true
		e = re[0]
		return
	}

	// no existing link found, search for proc directly
	envid, _, _ := findEnv(g, src)
	rv := g.PredecessorsWith(envid, helpers.Qbv(types.VType("logic-state"), "path", spec.Path))
	if len(rv) == 1 {
		success = true
		e.Target = rv[0].ID
	}

	return
}

type SpecParentDataset struct {
	Name string
}

func (spec SpecParentDataset) Resolve(g types.CoreGraph, mid uint64, src types.VertexTuple) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "dataset-gateway",
	}
	e.Props = e.Props.Set("name", types.Property{MsgSrc: mid, Value: spec.Name})

	// check for existing link - there can be only be one
	re := g.OutWith(src.ID, helpers.Qbe(types.EType("dataset-gateway")))
	if len(re) == 1 {
		success = true
		e = re[0]
		// TODO semantics should preclude this from being able to change, but doing it dirty means force-setting it anyway for now
	} else {

		// no existing link found; search for proc directly
		envid, _, _ := findEnv(g, src)
		rv := g.PredecessorsWith(envid, helpers.Qbv(types.VType("parent-dataset"), "name", spec.Name))
		if len(rv) != 0 { // >1 shouldn't be possible
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}

type SpecNetListener struct {
	Port  int
	Proto string
}

func (spec SpecNetListener) Resolve(g types.CoreGraph, mid uint64, src types.VertexTuple) (e types.StdEdge, success bool) {
	// check for existing edge; this one is quite straightforward
	re := g.OutWith(src.ID, helpers.Qbe(types.EType("listening"), "type", "port", "port", spec.Port, "proto", spec.Proto))
	if len(re) == 1 {
		return re[0], true
	}

	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "listening",
	}

	e.Props = e.Props.Set("port", types.Property{MsgSrc: mid, Value: spec.Port})
	e.Props = e.Props.Set("proto", types.Property{MsgSrc: mid, Value: spec.Proto})

	envid, _, hasenv := findEnv(g, src)
	if hasenv {
		rv := g.PredecessorsWith(envid, helpers.Qbv(types.VType("comm"), "type", "port", "port", spec.Port))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}

type SpecUnixDomainListener struct {
	Path string
}

func (spec SpecUnixDomainListener) Resolve(g types.CoreGraph, mid uint64, src types.VertexTuple) (e types.StdEdge, success bool) {
	// check for existing edge; this one is quite straightforward
	re := g.OutWith(src.ID, helpers.Qbe(types.EType("listening"), "type", "unix", "path", spec.Path))
	if len(re) == 1 {
		return re[0], true
	}

	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "listening",
	}

	e.Props = e.Props.Set("path", types.Property{MsgSrc: mid, Value: spec.Path})

	envid, _, hasenv := findEnv(g, src)
	if hasenv {
		rv := g.PredecessorsWith(envid, helpers.Qbv(types.VType("comm"), "type", "unix", "path", spec.Path))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}
