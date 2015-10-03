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
		v: pv{typ: "dataset", props: system.RawProps{
			"name":    d.Name,
			"version": d.Version,
			"epoch":   d.Epoch,
			"release": d.Release,
			"arch":    d.Arch,
		}},
		u: pkgYumUnify,
	}}
}

func pkgYumUnify(g system.CoreGraph, u system.UnifyInstructionForm) int {
	props := u.Vertex().Properties()
	vtv := g.VerticesWith(q.Qbv(system.VType("pkg-yum"),
		"name", props["name"],
		"version", props["version"],
		"arch", props["arch"],
		"epoch", props["epoch"],
	))

	if len(vtv) > 0 {
		return vtv[0].ID
	}
	return 0
}
