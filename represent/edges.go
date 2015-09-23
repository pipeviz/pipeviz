package represent

import (
	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/tag1consulting/pipeviz/represent/types"
)

type StandardEdge struct {
	id     int
	Source int
	Target int
	EType  types.EType
	Props  ps.Map
}
