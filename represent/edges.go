package represent

import (
	"github.com/mndrix/ps"
	"github.com/sdboyer/gogl"
)

type StandardEdge struct {
	id     int
	Source int
	Target int
	// TODO do these *actually* need to be persistent structures?
	Spec  ps.Map
	Props ps.Map
	Label string
}

type MessageArc interface {
	gogl.Arc
	Data() interface{}
	Label() string
}
