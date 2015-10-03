package semantic

import (
	"encoding/hex"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/maputil"
	"github.com/tag1consulting/pipeviz/represent/q"
	"github.com/tag1consulting/pipeviz/types/system"
)

type LogicState struct {
	Datasets    []DataLink `json:"datasets,omitempty"`
	Environment EnvLink    `json:"environment,omitempty"`
	ID          struct {
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

func (d LogicState) UnificationForm(id uint64) []system.UnifyInstructionForm {
	v := system.NewVertex("logic-state", id,
		pp("path", d.Path),
		pp("lgroup", d.Lgroup),
		pp("nick", d.Nick),
		pp("type", d.Type),
		pp("version", d.ID.Version),
		pp("semver", d.ID.Semver),
	)

	var edges system.EdgeSpecs

	if d.ID.CommitStr != "" {
		var commit Sha1
		byts, err := hex.DecodeString(d.ID.CommitStr)
		if err == nil {
			copy(commit[:], byts[0:20])
			edges = append(edges, SpecCommit{commit})
		}
	}

	for _, dl := range d.Datasets {
		edges = append(edges, dl)
	}

	return []system.UnifyInstructionForm{uif{v: v, u: lsUnify, e: edges, se: system.EdgeSpecs{d.Environment}}}
}

func lsUnify(g system.CoreGraph, u system.UnifyInstructionForm) int {
	// only one scoping edge - the envlink
	edge, success := u.ScopingSpecs()[0].(EnvLink).Resolve(g, 0, emptyVT(u.Vertex()))
	if !success {
		// FIXME scoping edge resolution failure does not mean no match - there could be an orphan
		return 0
	}

	vp := u.Vertex().Properties
	path, _ := vp.Lookup("path")
	return findMatchingEnvId(g, edge, g.VerticesWith(q.Qbv(system.VType("logic-state"), "path", path.(system.Property).Value)))
}

type SpecCommit struct {
	Sha1 Sha1
}

func (spec SpecCommit) Resolve(g system.CoreGraph, mid uint64, src system.VertexTuple) (e system.StdEdge, success bool) {
	e = system.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "version",
	}
	e.Props = e.Props.Set("sha1", system.Property{MsgSrc: mid, Value: spec.Sha1})

	re := g.OutWith(src.ID, q.Qbe(system.EType("version")))
	if len(re) > 0 {
		sha1, _ := re[0].Props.Lookup("sha1")
		e.ID = re[0].ID // FIXME setting the id to non-0 AND failing is currently unhandled
		if sha1.(system.Property).Value == spec.Sha1 {
			success = true
			e.Target = re[0].Target
		} else {
			rv := g.VerticesWith(q.Qbv(system.VType("commit"), "sha1", spec.Sha1))
			if len(rv) == 1 {
				success = true
				e.Target = rv[0].ID
			}
		}
	} else {
		rv := g.VerticesWith(q.Qbv(system.VType("commit"), "sha1", spec.Sha1))
		if len(rv) == 1 {
			success = true
			e.Target = rv[0].ID
		}
	}

	return
}

func (spec DataLink) Resolve(g system.CoreGraph, mid uint64, src system.VertexTuple) (e system.StdEdge, success bool) {
	e = system.StdEdge{
		Source: src.ID,
		Props:  ps.NewMap(),
		EType:  "datalink",
	}

	// DataLinks have a 'name' field that is expected to be unique for the source, if present
	if spec.Name != "" {
		// TODO 'name' is a traditional unique key; a change in it inherently denotes a new edge. how to handle this?
		// FIXME this approach just always updates the mid, which is weird?
		e.Props = e.Props.Set("name", system.Property{MsgSrc: mid, Value: spec.Name})

		re := g.OutWith(src.ID, q.Qbe(system.EType("datalink"), "name", spec.Name))
		if len(re) == 1 {
			success = true
			e = re[0]
		}
	}

	if spec.Type != "" {
		e.Props = e.Props.Set("type", system.Property{MsgSrc: mid, Value: spec.Type})
	}
	if spec.Subset != "" {
		e.Props = e.Props.Set("subset", system.Property{MsgSrc: mid, Value: spec.Subset})
	}
	if spec.Interaction != "" {
		e.Props = e.Props.Set("interaction", system.Property{MsgSrc: mid, Value: spec.Interaction})
	}

	// Special bits: if we have ConnUnix data, eliminate ConnNet data, and vice-versa.
	var isLocal bool
	if spec.ConnUnix.Path != "" {
		isLocal = true
		e.Props = e.Props.Set("path", system.Property{MsgSrc: mid, Value: spec.ConnUnix.Path})
		e.Props = e.Props.Delete("hostname")
		e.Props = e.Props.Delete("ipv4")
		e.Props = e.Props.Delete("ipv6")
		e.Props = e.Props.Delete("port")
		e.Props = e.Props.Delete("proto")
	} else {
		e.Props = e.Props.Set("port", system.Property{MsgSrc: mid, Value: spec.ConnNet.Port})
		e.Props = e.Props.Set("proto", system.Property{MsgSrc: mid, Value: spec.ConnNet.Proto})

		// can only be one of hostname, ipv4 or ipv6
		if spec.ConnNet.Hostname != "" {
			e.Props = e.Props.Set("hostname", system.Property{MsgSrc: mid, Value: spec.ConnNet.Hostname})
		} else if spec.ConnNet.Ipv4 != "" {
			e.Props = e.Props.Set("ipv4", system.Property{MsgSrc: mid, Value: spec.ConnNet.Ipv4})
		} else {
			e.Props = e.Props.Set("ipv6", system.Property{MsgSrc: mid, Value: spec.ConnNet.Ipv6})
		}
	}

	if success {
		return
	}

	var sock system.VertexTuple
	var rv system.VertexTupleVector // just for reuse
	// If net, must scan; if local, a bit easier.
	if !isLocal {
		// First, find the environment vertex
		rv = g.VerticesWith(q.Qbv(system.VType("environment")))
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
		rv = g.PredecessorsWith(envid, q.Qbv(system.VType("comm"), "type", "port", "port", spec.ConnNet.Port).And(q.Qbe(system.EType("envlink"))))

		if len(rv) != 1 {
			return
		}
		sock = rv[0]

		// With sock in hand, now find its proc
		rv = g.PredecessorsWith(sock.ID, q.Qbe(system.EType("listening"), "proto", spec.ConnNet.Proto).And(q.Qbv(system.VType("process"))))
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
		rv = g.PredecessorsWith(envid, q.Qbv(system.VType("comm"), "path", spec.ConnUnix.Path).And(q.Qbe(system.EType("envlink"))))
		if len(rv) != 1 {
			return
		}
		sock = rv[0]

		// With sock in hand, now find its proc
		rv = g.PredecessorsWith(sock.ID, q.Qbv(system.VType("process")).And(q.Qbe(system.EType("listening"))))
		if len(rv) != 1 {
			// TODO could/will we ever allow >1?
			return
		}
	}

	rv = g.SuccessorsWith(rv[0].ID, q.Qbv(system.VType("parent-dataset")))
	// FIXME this absolutely could be more than 1
	if len(rv) != 1 {
		return
	}
	dataset := rv[0]

	// if the spec indicates a subset, find it
	if spec.Subset != "" {
		rv = g.PredecessorsWith(rv[0].ID, q.Qbv(system.VType("dataset"), "name", spec.Subset).And(q.Qbe(system.EType("dataset-hierarchy"))))
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
