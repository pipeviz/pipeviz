package represent

import (
	"github.com/sdboyer/pipeviz/interpret"
)

// the main graph construct
type CoreGraph struct{}

// some kind of structure representing the delta introduced by a message
type Delta struct{}

// the method to merge a message into the graph
func (g *CoreGraph) Merge(interpret.Message) Delta {

}
