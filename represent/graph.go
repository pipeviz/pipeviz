package represent

import (
	"github.com/sdboyer/pipeviz/interpret"
)

// the main graph construct
type CoreGraph struct {
	vlist   map[int]Vertex
	vserial int
}

type Vertex struct {
	props []Property
}

type Property struct {
	MsgSrc int
	Key    string
	Value  string
}

// some kind of structure representing the delta introduced by a message
type Delta struct{}

// the method to merge a message into the graph
func (g *CoreGraph) Merge(interpret.Message) Delta {

}

// the func we eventually aim to fulfill, replacing Merge for integrating messages
//func (g CoreGraph) Cons(interpret.Message) CoreGraph, Delta, error {}
