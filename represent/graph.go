package represent

import (
	"fmt"
	"strconv"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
	"github.com/pipeviz/pipeviz/types/system"
)

var i2a = func(i uint64) string {
	return strconv.FormatUint(i, 10)
}

// the main graph construct
type coreGraph struct {
	msgid, vserial uint64
	vtuples        intMapV
	partials       ps.Map
	orphans        edgeSpecSet // FIXME breaks immut
}

// NewGraph creates a new in-memory coreGraph and returns it as a system.CoreGraph.
func NewGraph() system.CoreGraph {
	logrus.WithFields(logrus.Fields{
		"system": "engine",
	}).Debug("New coreGraph created")

	return &coreGraph{vtuples: newIntMapV(), partials: ps.NewMap(), vserial: 0}
}

type veProcessingInfo struct {
	vt    system.VertexTuple
	uif   system.UnifyInstructionForm // TODO REMOVE
	e     []system.EdgeSpec
	msgid uint64
}

type edgeSpecSet []*veProcessingInfo

func (ess edgeSpecSet) EdgeCount() (i int) {
	for _, tuple := range ess {
		i = i + len(tuple.e)
	}
	return
}

// Clones the graph object, and the map pointers it contains.
func (g *coreGraph) clone() *coreGraph {
	var cp coreGraph
	cp = *g
	return &cp
}

func (g *coreGraph) MsgID() uint64 {
	return g.msgid
}

// Gets the vtTuple for a given vertex id.
func (g *coreGraph) Get(id uint64) (system.VertexTuple, error) {
	if id > g.vserial {
		return system.VertexTuple{}, fmt.Errorf("Graph has only %d elements, no vertex yet exists with id %d", g.vserial, id)
	}

	vtx, exists := g.vtuples.Get(id)
	if exists {
		return vtx, nil
	}

	return system.VertexTuple{}, fmt.Errorf("No vertex exists with id %d at the present revision of the graph", id)
}
