package semantic

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/maputil"
	"github.com/pipeviz/pipeviz/represent/q"
	"github.com/pipeviz/pipeviz/types/system"
)

func init() {
	if err := registerUnifier("dataset", unifyDataset); err != nil {
		panic("dataset vertex already registered")
	}
	if err := registerUnifier("parent-dataset", unifyParentDataset); err != nil {
		panic("parent-dataset vertex already registered")
	}
	if err := registerResolver("dataset-hierarchy", resolveSpecDatasetHierarchy); err != nil {
		panic("dataset-hierarchy edge already registered")
	}
	if err := registerResolver("data-provenance", resolveDataProvenance); err != nil {
		panic("data-provenance edge already registered")
	}
}

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
	system.EdgeSpec
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

func (d Dataset) UnificationForm() []system.UnifyInstructionForm {
	v := pv{typ: "dataset", props: system.RawProps{
		"name":        d.Name,
		"create-time": d.CreateTime,
	}}
	var edges []system.EdgeSpec

	edges = append(edges, d.Genesis)

	return []system.UnifyInstructionForm{uif{
		v: v,
		u: unifyDataset,
		se: []system.EdgeSpec{specDatasetHierarchy{
			Environment: d.Environment,
			NamePath:    []string{d.Parent},
		}},
		e: edges,
	}}
}

// Unmarshaling a Dataset involves resolving whether it has α genesis (string), or
// a provenancial one (struct). So we have to decode directly, here.
func (ds *Dataset) UnmarshalJSON(data []byte) (err error) {
	type αDataset struct {
		Name       string    `json:"name"`
		CreateTime string    `json:"create-time"`
		Genesis    DataAlpha `json:"genesis"`
	}
	type provDataset struct {
		Name       string         `json:"name"`
		CreateTime string         `json:"create-time"`
		Genesis    DataProvenance `json:"genesis"`
	}

	a, b := αDataset{}, provDataset{}
	// use α first, as that can match the case where it's not specified (though schema
	// currently does not allow that)
	if err = json.Unmarshal(data, &a); err == nil {
		ds.Name, ds.CreateTime, ds.Genesis = a.Name, a.CreateTime, a.Genesis
	} else if err = json.Unmarshal(data, &b); err == nil {
		ds.Name, ds.CreateTime, ds.Genesis = b.Name, b.CreateTime, b.Genesis
	} else {
		err = errors.New("JSON genesis did not match either alpha or provenancial forms.")
	}

	return err
}

func unifyDataset(g system.CoreGraph, u system.UnifyInstructionForm) uint64 {
	vtv := g.VerticesWith(q.Qbv(system.VType("dataset"), "name", u.Vertex().Properties()["name"]))
	if len(vtv) == 0 {
		return 0
	}

	spec := u.ScopingSpecs()[0].(specDatasetHierarchy)
	el, success := spec.Environment.Resolve(g, 0, emptyVT(u.Vertex()))
	// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
	if success {
		for _, vt := range vtv {
			if id := findMatchingEnvId(g, el, g.SuccessorsWith(vt.ID, q.Qbe(system.EType("dataset-hierarchy")))); id != 0 {
				return vt.ID
			}
		}
	}

	return 0
}

func (d ParentDataset) UnificationForm() []system.UnifyInstructionForm {
	ret := []system.UnifyInstructionForm{uif{
		v: pv{typ: "parent-dataset", props: system.RawProps{
			"name": d.Name,
			"path": d.Path,
		}},
		u:  unifyParentDataset,
		se: []system.EdgeSpec{d.Environment},
	}}

	// TODO make recursive. which also means getting rid of the whole parent type...
	for _, sub := range d.Subsets {
		sub.Parent = d.Name
		sub.Environment = d.Environment
		ret = append(ret, sub.UnificationForm()...)
	}

	return ret
}

func unifyParentDataset(g system.CoreGraph, u system.UnifyInstructionForm) uint64 {
	edge, success := u.ScopingSpecs()[0].(EnvLink).Resolve(g, 0, emptyVT(u.Vertex()))
	if !success {
		// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
		return 0
	}

	props := u.Vertex().Properties()
	return findMatchingEnvId(g, edge, g.VerticesWith(q.Qbv(system.VType("parent-dataset"), "path", props["path"], "name", props["name"])))
}

type specDatasetHierarchy struct {
	Environment EnvLink
	NamePath    []string // path through the series of names that arrives at the final dataset
}

func resolveSpecDatasetHierarchy(e system.EdgeSpec, g system.CoreGraph, mid uint64, src system.VertexTuple) (system.StdEdge, bool) {
	return e.(specDatasetHierarchy).Resolve(g, mid, src)
}

func (spec specDatasetHierarchy) Resolve(g system.CoreGraph, mid uint64, src system.VertexTuple) (e system.StdEdge, success bool) {
	e = system.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "dataset-hierarchy",
	}
	e.Props = e.Props.Set("parent", system.Property{MsgSrc: mid, Value: spec.NamePath[0]})

	// check for existing link - there can be only be one
	re := g.OutWith(src.ID, q.Qbe(system.EType("dataset-hierarchy")))
	if len(re) == 1 {
		success = true
		e = re[0]
		// TODO semantics should preclude this from being able to change, but doing it dirty means force-setting it anyway for now
		e.Props = e.Props.Set("parent", system.Property{MsgSrc: mid, Value: spec.NamePath[0]})
		return
	}

	// no existing link found; search for proc directly
	envlink, success := spec.Environment.Resolve(g, 0, src)
	if success {
		rv := g.PredecessorsWith(envlink.Target, q.Qbv(system.VType("parent-dataset"), "name", spec.NamePath[0]))
		if len(rv) != 0 { // >1 shouldn't be possible
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}

// Type indicates the EType the EdgeSpec will produce. This is necessarily invariant.
func (spec specDatasetHierarchy) Type() system.EType {
	return "dataset-hierarchy"
}

func resolveDataProvenance(e system.EdgeSpec, g system.CoreGraph, mid uint64, src system.VertexTuple) (system.StdEdge, bool) {
	switch typ := e.(type) {
	case DataProvenance, DataAlpha:
		return typ.Resolve(g, mid, src)
	default:
		// Hitting this branch guarantees there's some incorrect hardcoding somewhere
		panic(fmt.Sprintf("Invalid dynamic type %T passed to resolveDataProvenance"))
	}
}

func (spec DataProvenance) Resolve(g system.CoreGraph, mid uint64, src system.VertexTuple) (e system.StdEdge, success bool) {
	// FIXME this presents another weird case where "success" is not binary. We *could*
	// find an already-existing data-provenance edge, but then have some net-addr params
	// change which cause it to fail to resolve to an environment. If we call that successful,
	// then it won't try to resolve again later...though, hm, just call it unsuccessful and
	// then try again one more time. Maybe it is fine. THINK IT THROUGH.

	e = system.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "data-provenance",
	}
	e.Props = assignAddress(mid, spec.Address, e.Props, false)

	re := g.OutWith(src.ID, q.Qbe(system.EType("data-provenance")))
	if len(re) == 1 {
		reresolve := maputil.AnyMatch(e.Props, re[0].Props, "hostname", "ipv4", "ipv6")

		e = re[0]
		if spec.SnapTime != "" {
			e.Props = e.Props.Set("snap-time", system.Property{MsgSrc: mid, Value: spec.SnapTime})
		}

		if reresolve {
			e.Props = assignAddress(mid, spec.Address, e.Props, true)
		} else {
			return e, true
		}
	}

	envid, found := findEnvironment(g, e.Props)
	if !found {
		// TODO returning this already-modified edge necessitates that the core system
		// disregard 'failed' edges. which should be fine, that should be a guarantee
		return e, false
	}

	e.Target, success = findDataset(g, envid, spec.Dataset)
	return
}

// Type indicates the EType the EdgeSpec will produce. This is necessarily invariant.
func (spec DataProvenance) Type() system.EType {
	return "data-provenance"
}

func (spec DataAlpha) Resolve(g system.CoreGraph, mid uint64, src system.VertexTuple) (e system.StdEdge, success bool) {
	// TODO this makes a loop...are we cool with that?
	success = true // impossible to fail here
	e = system.StdEdge{
		Source: src.ID,
		Target: src.ID,
		Props:  ps.NewMap(),
		EType:  "data-provenance",
	}

	re := g.OutWith(src.ID, q.Qbe(system.EType("data-provenance")))
	if len(re) == 1 {
		e = re[0]
	}

	return
}

// Type indicates the EType the EdgeSpec will produce. This is necessarily invariant.
func (spec DataAlpha) Type() system.EType {
	return "data-provenance"
}
