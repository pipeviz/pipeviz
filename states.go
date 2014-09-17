package pipeviz

import (
	"fmt"
)

type BuildState struct {
	Code CodeStateIdentifier
	Data DataStateIdentifier
	//Configuration ConfState
}

type EnvironmentState struct {
}

type ApplicationInstance struct {
	E EnvironmentState
	B BuildState
}

type CodeState []CodeStateIdentifier
type DataState []DataStateIdentifier

type CodeStateIdentifier fmt.Stringer
type DataStateIdentifier fmt.Stringer

/*
First, we need something to represent the collected code states. This can be just a map.
*/
type CodeStateSet []CodeState

/*
We also need a similar structure for tracking DataState.
*/
type DataStateSet []DataState

type StateDescriptor map[string]string

/*
Example ApplicationInstance:

Environment:
	Runtime: go 1.3.1 <statemod timestamp>
	Context: docker <statemod timestamp>
	Address: proto://foo.bar.baz

Build:
	Code: aa12345, <statemod timestamp>
	Data: <created timestamp>, <statemod timestamp>
*/
