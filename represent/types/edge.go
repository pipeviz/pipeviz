package types

import "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"

type StandardEdge struct {
	ID     int
	Source int
	Target int
	EType  EType
	Props  ps.Map
}
