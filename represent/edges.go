package represent

import (
	"github.com/sdboyer/gogl"
)

type StandardEdge struct {
	Source interface{}
	Target interface{}
	Label  string
	Data   interface{}
}

type MessageArc interface {
	gogl.Arc
	Data() interface{}
	Label() string
}
