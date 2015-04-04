package represent

import (
	"github.com/mndrix/ps"
	"github.com/sdboyer/gogl"
)

type StandardEdge struct {
	id     int
	Source int
	Target int
	Spec   ps.Map
	Props  ps.Map
}

type MessageArc interface {
	gogl.Arc
	Data() interface{}
	Label() string
}
