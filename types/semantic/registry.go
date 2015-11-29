package semantic

import (
	"fmt"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/pipeviz/pipeviz/types/system"
)

var (
	unifyMap     = make(map[system.VType]system.UnifyFunc)
	unifyEdgeMap = make(map[system.EType]system.UnifyEdgeFunc)
	resolveMap   = make(map[system.EType]system.ResolveFunc)
)

func registerUnifier(typ system.VType, f system.UnifyFunc) error {
	if _, exists := unifyMap[typ]; exists {
		logrus.WithFields(logrus.Fields{
			"system": "semantic",
			"type":   typ,
		}).Error("Attempt to register UnifyFunc for type more than once")
		return fmt.Errorf("a UnifyFunc for vertex type %q is already registered", typ)
	}

	unifyMap[typ] = f
	return nil
}

// Unify unifies a UIF with the provided CoreGraph using a UnifyFunc preregistered by the semantic
// system for the type of vertex contained in the UIF. An error is returned iff no UnifyFunc was
// registered for the vertex type.
func Unify(g system.CoreGraph, uif system.UnifyInstructionForm) (uint64, error) {
	typ := uif.Vertex().Type()
	if f, exists := unifyMap[typ]; exists {
		return f(g, uif), nil
	} else {
		return 0, fmt.Errorf("no unifier exists for vertices of type %q", typ)
	}
}

func registerEdgeUnifier(typ system.EType, f system.UnifyEdgeFunc) error {
	if _, exists := unifyEdgeMap[typ]; exists {
		logrus.WithFields(logrus.Fields{
			"system": "semantic",
			"type":   typ,
		}).Error("Attempt to register UnifyEdgeFunc for type more than once")
		return fmt.Errorf("a UnifyEdgeFunc for edge type %q is already registered", typ)
	}

	unifyEdgeMap[typ] = f
	return nil
}

func UnifyEdge(vt system.VertexTuple, e system.EdgeSpec) (uint64, error) {
	typ := e.Type()
	if f, exists := unifyEdgeMap[typ]; exists {
		return f(vt, e), nil
	} else {
		return 0, fmt.Errorf("no edge unifier exists for edges of type %q", typ)
	}
}

func registerResolver(typ system.EType, f system.ResolveFunc) error {
	if _, exists := resolveMap[typ]; exists {
		logrus.WithFields(logrus.Fields{
			"system": "semantic",
			"type":   typ,
		}).Error("Attempt to register ResolverFunc for type more than once")
		return fmt.Errorf("a ResolverFunc for edge type %q is already registered", typ)
	}

	resolveMap[typ] = f
	return nil
}

// Resolve resolves an EdgeSpec with the provided CoreGraph and VertexTuple using a ResolveFunc
// preregistered by the semantic system for the type of edge contained in the UIF. An error is
// returned iff no ResolveFunc was registered for the edge type.
func Resolve(e system.EdgeSpec, g system.CoreGraph, msgid uint64, vt system.VertexTuple) (system.StdEdge, bool, error) {
	typ := e.Type()
	if f, exists := resolveMap[typ]; exists {
		id, succ := f(e, g, msgid, vt)
		return id, succ, nil
	} else {
		return system.StdEdge{}, false, fmt.Errorf("no resolver exists for edges of type %q", typ)
	}
}
