package system

import (
	"fmt"

	"github.com/pipeviz/pipeviz/Godeps/_workspace/src/github.com/mndrix/ps"
)

type Property struct {
	MsgSrc uint64      `json:"msgsrc"`
	Value  interface{} `json:"value"`
}

type PropPair struct {
	K string
	V interface{}
}

// emptyChecker is an interface that raw property values can implement to
// indicate if they are empty.
type emptyChecker interface {
	IsEmpty() bool
}

// RawProps is an alias for map[string]interface{}, used to allow mimicking
// the ps.Map interface, and also to make the use pattern a little clearer.
//
// Implementations are not expected to use the provided methods at all -
// doing so would be far more awkward than simply using Go's built-in
// map utilities.
type RawProps map[string]interface{}

func (m RawProps) IsNil() bool {
	return m == nil
}

func (m RawProps) Set(key string, value ps.Any) ps.Map {
	m[key] = value
	return m
}

func (m RawProps) Delete(key string) ps.Map {
	delete(m, key)
	return m
}

func (m RawProps) Lookup(key string) (ps.Any, bool) {
	v, ok := m[key]
	return v, ok
}

func (m RawProps) Size() int {
	return len(m)
}

func (m RawProps) ForEach(f func(key string, val ps.Any)) {
	for k, v := range m {
		f(k, v)
	}
}

func (m RawProps) Keys() (ret []string) {
	for k, _ := range m {
		ret = append(ret, k)
	}

	return
}

func (m RawProps) String() string {
	// not sure if this is the right way to pass through, but...
	return fmt.Sprintf("%v", m)
}
