package semantic

import (
	"github.com/pipeviz/pipeviz/represent/q"
	"github.com/pipeviz/pipeviz/types/system"
)

func init() {
	if err := registerUnifier("pkg-yum", unifyPkgYum); err != nil {
		panic("pkg-yum vertex already registered")
	}
}

type PkgYum struct {
	Name       string `json:"name,omitempty"`
	Repository string `json:"repository,omitempty"`
	Version    string `json:"version,omitempty"`
	Epoch      int    `json:"epoch,omitempty"`
	Release    string `json:"release,omitempty"`
	Arch       string `json:"arch,omitempty"`
}

func (d PkgYum) UnificationForm() []system.UnifyInstructionForm {

	return []system.UnifyInstructionForm{uif{
		v: pv{typ: "dataset", props: system.RawProps{
			"name":    d.Name,
			"version": d.Version,
			"epoch":   d.Epoch,
			"release": d.Release,
			"arch":    d.Arch,
		}},
	}}
}

func unifyPkgYum(g system.CoreGraph, u system.UnifyInstructionForm) uint64 {
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
