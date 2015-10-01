package interpret

import (
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/maputil"
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

type ConnNet struct {
	Hostname string `json:"hostname,omitempty"`
	Ipv4     string `json:"ipv4,omitempty"`
	Ipv6     string `json:"ipv6,omitempty"`
	Port     int    `json:"port,omitempty"`
	Proto    string `json:"proto,omitempty"`
}

type ConnUnix struct {
	Path string `json:"path,omitempty"`
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
		// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
		return 0
	}

	vp := u.Vertex().Properties
	path, _ := vp.Lookup("path")
	return hasMatchingEnv(g, edge, g.VerticesWith(helpers.Qbv(types.VType("logic-state"), "path", path.(types.Property).Value)))
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

func (spec DataLink) Resolve(g types.CoreGraph, mid uint64, src types.VertexTuple) (e types.StdEdge, success bool) {
	e = types.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "datalink",
	}

	// DataLinks have a 'name' field that is expected to be unique for the source, if present
	if spec.Name != "" {
		// TODO 'name' is a traditional unique key; a change in it inherently denotes a new edge. how to handle this?
		// FIXME this approach just always updates the mid, which is weird?
		e.Props = e.Props.Set("name", types.Property{MsgSrc: mid, Value: spec.Name})

		re := g.OutWith(src.ID, helpers.Qbe(types.EType("datalink"), "name", spec.Name))
		if len(re) == 1 {
			success = true
			e = re[0]
		}
	}

	if spec.Type != "" {
		e.Props = e.Props.Set("type", types.Property{MsgSrc: mid, Value: spec.Type})
	}
	if spec.Subset != "" {
		e.Props = e.Props.Set("subset", types.Property{MsgSrc: mid, Value: spec.Subset})
	}
	if spec.Interaction != "" {
		e.Props = e.Props.Set("interaction", types.Property{MsgSrc: mid, Value: spec.Interaction})
	}

	// Special bits: if we have ConnUnix data, eliminate ConnNet data, and vice-versa.
	var isLocal bool
	if spec.ConnUnix.Path != "" {
		isLocal = true
		e.Props = e.Props.Set("path", types.Property{MsgSrc: mid, Value: spec.ConnUnix.Path})
		e.Props = e.Props.Delete("hostname")
		e.Props = e.Props.Delete("ipv4")
		e.Props = e.Props.Delete("ipv6")
		e.Props = e.Props.Delete("port")
		e.Props = e.Props.Delete("proto")
	} else {
		e.Props = e.Props.Set("port", types.Property{MsgSrc: mid, Value: spec.ConnNet.Port})
		e.Props = e.Props.Set("proto", types.Property{MsgSrc: mid, Value: spec.ConnNet.Proto})

		// can only be one of hostname, ipv4 or ipv6
		if spec.ConnNet.Hostname != "" {
			e.Props = e.Props.Set("hostname", types.Property{MsgSrc: mid, Value: spec.ConnNet.Hostname})
		} else if spec.ConnNet.Ipv4 != "" {
			e.Props = e.Props.Set("ipv4", types.Property{MsgSrc: mid, Value: spec.ConnNet.Ipv4})
		} else {
			e.Props = e.Props.Set("ipv6", types.Property{MsgSrc: mid, Value: spec.ConnNet.Ipv6})
		}
	}

	if success {
		return
	}

	var sock types.VertexTuple
	var rv types.VertexTupleVector // just for reuse
	// If net, must scan; if local, a bit easier.
	if !isLocal {
		// First, find the environment vertex
		rv = g.VerticesWith(helpers.Qbv(types.VType("environment")))
		var envid int
		for _, vt := range rv {
			if maputil.AnyMatch(e.Props, vt.Vertex.Properties, "hostname", "ipv4", "ipv6") {
				envid = vt.ID
				break
			}
		}

		// No matching env found, bail out
		if envid == 0 {
			return
		}

		// Now, walk the environment's edges to find the vertex representing the port
		rv = g.PredecessorsWith(envid, helpers.Qbv(types.VType("comm"), "type", "port", "port", spec.ConnNet.Port).And(helpers.Qbe(types.EType("envlink"))))

		if len(rv) != 1 {
			return
		}
		sock = rv[0]

		// With sock in hand, now find its proc
		rv = g.PredecessorsWith(sock.ID, helpers.Qbe(types.EType("listening"), "proto", spec.ConnNet.Proto).And(helpers.Qbv(types.VType("process"))))
		if len(rv) != 1 {
			// TODO could/will we ever allow >1?
			return
		}
	} else {
		envid, _, exists := findEnv(g, src)

		if !exists {
			// this is would be a pretty weird case
			return
		}

		// Walk the graph to find the vertex representing the unix socket
		rv = g.PredecessorsWith(envid, helpers.Qbv(types.VType("comm"), "path", spec.ConnUnix.Path).And(helpers.Qbe(types.EType("envlink"))))
		if len(rv) != 1 {
			return
		}
		sock = rv[0]

		// With sock in hand, now find its proc
		rv = g.PredecessorsWith(sock.ID, helpers.Qbv(types.VType("process")).And(helpers.Qbe(types.EType("listening"))))
		if len(rv) != 1 {
			// TODO could/will we ever allow >1?
			return
		}
	}

	rv = g.SuccessorsWith(rv[0].ID, helpers.Qbv(types.VType("parent-dataset")))
	// FIXME this absolutely could be more than 1
	if len(rv) != 1 {
		return
	}
	dataset := rv[0]

	// if the spec indicates a subset, find it
	if spec.Subset != "" {
		rv = g.PredecessorsWith(rv[0].ID, helpers.Qbv(types.VType("dataset"), "name", spec.Subset).And(helpers.Qbe(types.EType("dataset-hierarchy"))))
		if len(rv) != 1 {
			return
		}
		dataset = rv[0]
	}

	// FIXME only recording the final target id is totally broken; see https://github.com/tag1consulting/pipeviz/issues/37

	// Aaaand we found our target.
	success = true
	e.Target = dataset.ID
	return
}
