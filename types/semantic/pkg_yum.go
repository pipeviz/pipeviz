package semantic

import (
	"github.com/tag1consulting/pipeviz/represent/q"
	"github.com/tag1consulting/pipeviz/types/system"
)

type PkgYum struct {
	Name       string `json:"name,omitempty"`
	Repository string `json:"repository,omitempty"`
	Version    string `json:"version,omitempty"`
	Epoch      int    `json:"epoch,omitempty"`
	Release    string `json:"release,omitempty"`
	Arch       string `json:"arch,omitempty"`
}

func (d PkgYum) UnificationForm(id uint64) []system.UnifyInstructionForm {

	return []system.UnifyInstructionForm{uif{
		v: system.NewVertex("dataset", id,
			system.PropPair{K: "name", V: d.Name},
			system.PropPair{K: "version", V: d.Version},
			system.PropPair{K: "epoch", V: d.Epoch},
			system.PropPair{K: "release", V: d.Release},
			system.PropPair{K: "arch", V: d.Arch},
		),
		u: pkgYumUnify,
	}}
}

func pkgYumUnify(g system.CoreGraph, u system.UnifyInstructionForm) int {
	props := u.Vertex().Properties
	name, _ := props.Lookup("name")
	version, _ := props.Lookup("version")
	arch, _ := props.Lookup("arch")
	epoch, _ := props.Lookup("epoch")

	vtv := g.VerticesWith(q.Qbv(system.VType("pkg-yum"),
		"name", name.(system.Property).Value,
		"version", version.(system.Property).Value,
		"arch", arch.(system.Property).Value,
		"epoch", epoch.(system.Property).Value,
	))

	if len(vtv) > 0 {
		return vtv[0].ID
	}
	return 0
}
