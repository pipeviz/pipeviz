package helpers

import "github.com/tag1consulting/pipeviz/represent/types"

type edgeFilter struct {
	etype types.EType
	props []types.PropPair
}

type vertexFilter struct {
	vtype types.VType
	props []types.PropPair
}

type bothFilter struct {
	types.VFilter
	types.EFilter
}

// all temporary functions to just make query building a little easier for now
func (vf vertexFilter) VType() types.VType {
	return vf.vtype
}

func (vf vertexFilter) VProps() []types.PropPair {
	return vf.props
}

func (vf vertexFilter) EType() types.EType {
	return types.ETypeNone
}

func (vf vertexFilter) EProps() []types.PropPair {
	return nil
}

func (vf vertexFilter) And(ef types.EFilter) types.VEFilter {
	return bothFilter{vf, ef}
}

func (ef edgeFilter) VType() types.VType {
	return types.VTypeNone
}

func (ef edgeFilter) VProps() []types.PropPair {
	return nil
}

func (ef edgeFilter) EType() types.EType {
	return ef.etype
}

func (ef edgeFilter) EProps() []types.PropPair {
	return ef.props
}

func (ef edgeFilter) And(vf types.VFilter) types.VEFilter {
	return bothFilter{vf, ef}
}

// first string is vtype, then pairs after that are props
func Qbv(v ...interface{}) vertexFilter {
	switch len(v) {
	case 0:
		return vertexFilter{types.VTypeNone, nil}
	case 1, 2:
		return vertexFilter{v[0].(types.VType), nil}
	default:
		vf := vertexFilter{v[0].(types.VType), nil}
		v = v[1:]

		var k string
		var v2 interface{}
		for len(v) > 1 {
			k, v2, v = v[0].(string), v[1], v[2:]
			vf.props = append(vf.props, types.PropPair{k, v2})
		}

		return vf
	}
}

// first string is etype, then pairs after that are props
func Qbe(v ...interface{}) edgeFilter {
	switch len(v) {
	case 0:
		return edgeFilter{types.ETypeNone, nil}
	case 1, 2:
		return edgeFilter{v[0].(types.EType), nil}
	default:
		ef := edgeFilter{v[0].(types.EType), nil}
		v = v[1:]

		var k string
		var v2 interface{}
		for len(v) > 1 {
			k, v2, v = v[0].(string), v[1], v[2:]
			ef.props = append(ef.props, types.PropPair{k, v2})
		}

		return ef
	}
}
