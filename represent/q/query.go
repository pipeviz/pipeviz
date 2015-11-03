package q

import "github.com/pipeviz/pipeviz/types/system"

type edgeFilter struct {
	etype system.EType
	props []system.PropPair
}

type vertexFilter struct {
	vtype system.VType
	props []system.PropPair
}

type bothFilter struct {
	system.VFilter
	system.EFilter
}

// all temporary functions to just make query building a little easier for now
func (vf vertexFilter) VType() system.VType {
	return vf.vtype
}

func (vf vertexFilter) VProps() []system.PropPair {
	return vf.props
}

func (vf vertexFilter) EType() system.EType {
	return system.ETypeNone
}

func (vf vertexFilter) EProps() []system.PropPair {
	return nil
}

func (vf vertexFilter) And(ef system.EFilter) system.VEFilter {
	return bothFilter{vf, ef}
}

func (ef edgeFilter) VType() system.VType {
	return system.VTypeNone
}

func (ef edgeFilter) VProps() []system.PropPair {
	return nil
}

func (ef edgeFilter) EType() system.EType {
	return ef.etype
}

func (ef edgeFilter) EProps() []system.PropPair {
	return ef.props
}

func (ef edgeFilter) And(vf system.VFilter) system.VEFilter {
	return bothFilter{vf, ef}
}

// first string is vtype, then pairs after that are props
func Qbv(v ...interface{}) vertexFilter {
	switch len(v) {
	case 0:
		return vertexFilter{system.VTypeNone, nil}
	case 1, 2:
		return vertexFilter{v[0].(system.VType), nil}
	default:
		vf := vertexFilter{v[0].(system.VType), nil}
		v = v[1:]

		var k string
		var v2 interface{}
		for len(v) > 1 {
			k, v2, v = v[0].(string), v[1], v[2:]
			vf.props = append(vf.props, system.PropPair{K: k, V: v2})
		}

		return vf
	}
}

// first string is etype, then pairs after that are props
func Qbe(v ...interface{}) edgeFilter {
	switch len(v) {
	case 0:
		return edgeFilter{system.ETypeNone, nil}
	case 1, 2:
		return edgeFilter{v[0].(system.EType), nil}
	default:
		ef := edgeFilter{v[0].(system.EType), nil}
		v = v[1:]

		var k string
		var v2 interface{}
		for len(v) > 1 {
			k, v2, v = v[0].(string), v[1], v[2:]
			ef.props = append(ef.props, system.PropPair{K: k, V: v2})
		}

		return ef
	}
}
