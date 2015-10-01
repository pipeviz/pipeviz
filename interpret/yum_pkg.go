package interpret

import (
	"github.com/tag1consulting/pipeviz/represent/helpers"
	"github.com/tag1consulting/pipeviz/represent/types"
)

type YumPkg struct {
	Name       string `json:"name,omitempty"`
	Repository string `json:"repository,omitempty"`
	Version    string `json:"version,omitempty"`
	Epoch      int    `json:"epoch,omitempty"`
	Release    string `json:"release,omitempty"`
	Arch       string `json:"arch,omitempty"`
}

func (d YumPkg) UnificationForm(id uint64) []types.UnifyInstructionForm {

	return []types.UnifyInstructionForm{uif{
		v: types.NewVertex("dataset", id,
			types.PropPair{K: "name", V: d.Name},
			types.PropPair{K: "version", V: d.Version},
			types.PropPair{K: "epoch", V: d.Epoch},
			types.PropPair{K: "release", V: d.Release},
			types.PropPair{K: "arch", V: d.Arch},
		),
		u: pkgYumUnify,
	}}
}

func pkgYumUnify(g types.CoreGraph, u types.UnifyInstructionForm) int {
	props := u.Vertex().Properties
	name, _ := props.Lookup("name")
	version, _ := props.Lookup("version")
	arch, _ := props.Lookup("arch")
	epoch, _ := props.Lookup("epoch")

	vtv := g.VerticesWith(helpers.Qbv(types.VType("pkg-yum"),
		"name", name.(types.Property).Value,
		"version", version.(types.Property).Value,
		"arch", arch.(types.Property).Value,
		"epoch", epoch.(types.Property).Value,
	))

	if len(vtv) > 0 {
		return vtv[0].ID
	}
	return 0
}
