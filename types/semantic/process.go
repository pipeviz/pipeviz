package semantic

import (
	"fmt"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/represent/q"
	"github.com/pipeviz/pipeviz/types/system"
)

func init() {
	if err := registerUnifier("process", unifyProcess); err != nil {
		panic("process vertex already registered")
	}
	if err := registerUnifier("comm", unifyComm); err != nil {
		panic("comm vertex already registered")
	}
	if err := registerResolver("listening", resolveListening); err != nil {
		panic("version edge already registered")
	}
	if err := registerResolver("logic-link", resolveSpecLocalLogic); err != nil {
		panic("logic-link edge already registered")
	}
	if err := registerResolver("dataset-gateway", resolveSpecParentDataset); err != nil {
		panic("dataset-gateway edge already registered")
	}
}

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

func (d Process) UnificationForm() []system.UnifyInstructionForm {
	ret := make([]system.UnifyInstructionForm, 0)

	v := pv{
		typ: "process",
		props: system.RawProps{
			"pid":   d.Pid,
			"cwd":   d.Cwd,
			"group": d.Group,
			"user":  d.User,
		},
	}

	var edges []system.EdgeSpec

	for _, ls := range d.LogicStates {
		edges = append(edges, specLocalLogic{ls})
	}

	if d.Dataset != "" {
		edges = append(edges, specParentDataset{Name: d.Dataset})
	}

	for _, listen := range d.Listen {
		// TODO change this to use diff vtx types for unix domain sock and network sock
		v2 := pv{typ: "comm", props: system.RawProps{
			"type": listen.Type,
		}}

		if listen.Type == "unix" {
			edges = append(edges, specUnixDomainListener{Path: listen.Path})
			v2.props["path"] = listen.Path
		} else {
			for _, proto := range listen.Proto {
				edges = append(edges, specNetListener{Port: listen.Port, Proto: proto})
			}
			v2.props["port"] = listen.Port
		}
		ret = append(ret, uif{v: v2, u: unifyComm, se: []system.EdgeSpec{d.Environment}})
	}

	return append([]system.UnifyInstructionForm{uif{
		v:  v,
		u:  unifyProcess,
		e:  edges,
		se: []system.EdgeSpec{d.Environment},
	}}, ret...)
}

func unifyProcess(g system.CoreGraph, u system.UnifyInstructionForm) uint64 {
	// only one scoping edge - the envlink
	edge, success := u.ScopingSpecs()[0].(EnvLink).Resolve(g, 0, emptyVT(u.Vertex()))
	if !success {
		// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
		return 0
	}

	return findMatchingEnvId(g, edge, g.VerticesWith(q.Qbv(system.VType("process"), "pid", u.Vertex().Properties()["pid"])))
}

func unifyComm(g system.CoreGraph, u system.UnifyInstructionForm) uint64 {
	// only one scoping edge - the envlink
	edge, success := u.ScopingSpecs()[0].(EnvLink).Resolve(g, 0, emptyVT(u.Vertex()))
	if !success {
		// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
		return 0
	}

	vp := u.Vertex().Properties()
	typ, _ := vp["type"]
	path, haspath := vp["path"]
	if haspath {
		return findMatchingEnvId(g, edge, g.VerticesWith(q.Qbv(system.VType("comm"),
			"type", typ,
			"path", path)))
	} else {
		port, _ := vp["port"]
		return findMatchingEnvId(g, edge, g.VerticesWith(q.Qbv(system.VType("comm"),
			"type", typ,
			"port", port)))
	}
}

type specLocalLogic struct {
	Path string
}

func resolveSpecLocalLogic(e system.EdgeSpec, g system.CoreGraph, mid uint64, src system.VertexTuple) (system.StdEdge, bool) {
	return e.(specLocalLogic).Resolve(g, mid, src)
}

func (spec specLocalLogic) Resolve(g system.CoreGraph, mid uint64, src system.VertexTuple) (e system.StdEdge, success bool) {
	e = system.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "logic-link",
	}

	// search for existing link
	re := g.OutWith(src.ID, q.Qbe(system.EType("logic-link"), "path", spec.Path))
	if len(re) == 1 {
		// TODO don't set the path prop again, it's the unique id...meh, same question here w/uniqueness as above
		success = true
		e = re[0]
		return
	}

	// no existing link found, search for proc directly
	envid, _, _ := findEnv(g, src)
	rv := g.PredecessorsWith(envid, q.Qbv(system.VType("logic-state"), "path", spec.Path))
	if len(rv) == 1 {
		success = true
		e.Target = rv[0].ID
	}

	return
}

// Type indicates the EType the EdgeSpec will produce. This is necessarily invariant.
func (spec specLocalLogic) Type() system.EType {
	return "logic-link"
}

type specParentDataset struct {
	Name string
}

func resolveSpecParentDataset(e system.EdgeSpec, g system.CoreGraph, mid uint64, src system.VertexTuple) (system.StdEdge, bool) {
	return e.(specParentDataset).Resolve(g, mid, src)
}

func (spec specParentDataset) Resolve(g system.CoreGraph, mid uint64, src system.VertexTuple) (e system.StdEdge, success bool) {
	e = system.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "dataset-gateway",
	}
	e.Props = e.Props.Set("name", system.Property{MsgSrc: mid, Value: spec.Name})

	// check for existing link - there can be only be one
	re := g.OutWith(src.ID, q.Qbe(system.EType("dataset-gateway")))
	if len(re) == 1 {
		success = true
		e = re[0]
		// TODO semantics should preclude this from being able to change, but doing it dirty means force-setting it anyway for now
	} else {

		// no existing link found; search for proc directly
		envid, _, _ := findEnv(g, src)
		rv := g.PredecessorsWith(envid, q.Qbv(system.VType("parent-dataset"), "name", spec.Name))
		if len(rv) != 0 { // >1 shouldn't be possible
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}

// Type indicates the EType the EdgeSpec will produce. This is necessarily invariant.
func (spec specParentDataset) Type() system.EType {
	return "dataset-gateway"
}

type specNetListener struct {
	Port  int
	Proto string
}

func resolveListening(e system.EdgeSpec, g system.CoreGraph, mid uint64, src system.VertexTuple) (system.StdEdge, bool) {
	switch typ := e.(type) {
	case specNetListener, specUnixDomainListener:
		return typ.Resolve(g, mid, src)
	default:
		// Hitting this branch guarantees there's some incorrect hardcoding somewhere
		panic(fmt.Sprintf("Invalid dynamic type %T passed to resolveListening"))
	}
}

func (spec specNetListener) Resolve(g system.CoreGraph, mid uint64, src system.VertexTuple) (e system.StdEdge, success bool) {
	// check for existing edge; this one is quite straightforward
	re := g.OutWith(src.ID, q.Qbe(system.EType("listening"), "type", "port", "port", spec.Port, "proto", spec.Proto))
	if len(re) == 1 {
		return re[0], true
	}

	e = system.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "listening",
	}

	e.Props = e.Props.Set("port", system.Property{MsgSrc: mid, Value: spec.Port})
	e.Props = e.Props.Set("proto", system.Property{MsgSrc: mid, Value: spec.Proto})

	envid, _, hasenv := findEnv(g, src)
	if hasenv {
		rv := g.PredecessorsWith(envid, q.Qbv(system.VType("comm"), "type", "port", "port", spec.Port))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}

// Type indicates the EType the EdgeSpec will produce. This is necessarily invariant.
func (spec specNetListener) Type() system.EType {
	return "listening"
}

type specUnixDomainListener struct {
	Path string
}

func (spec specUnixDomainListener) Resolve(g system.CoreGraph, mid uint64, src system.VertexTuple) (e system.StdEdge, success bool) {
	// check for existing edge; this one is quite straightforward
	re := g.OutWith(src.ID, q.Qbe(system.EType("listening"), "type", "unix", "path", spec.Path))
	if len(re) == 1 {
		return re[0], true
	}

	e = system.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "listening",
	}

	e.Props = e.Props.Set("path", system.Property{MsgSrc: mid, Value: spec.Path})

	envid, _, hasenv := findEnv(g, src)
	if hasenv {
		rv := g.PredecessorsWith(envid, q.Qbv(system.VType("comm"), "type", "unix", "path", spec.Path))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}

// Type indicates the EType the EdgeSpec will produce. This is necessarily invariant.
func (spec specUnixDomainListener) Type() system.EType {
	return "listening"
}
