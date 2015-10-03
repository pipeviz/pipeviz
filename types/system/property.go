package system

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
