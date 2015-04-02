package represent

import (
	"github.com/sdboyer/gogl"
)

type StandardEdge struct {
	Source int
	Target int
	Label  string
	Data   interface{}
}

type MessageArc interface {
	gogl.Arc
	Data() interface{}
	Label() string
}
