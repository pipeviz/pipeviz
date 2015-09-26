package interpret

import "github.com/tag1consulting/pipeviz/represent/types"

// pp is a convenience function to create a types.PropPair. The compiler
// always inlines it, so there's no cost.
func pp(k string, v interface{}) types.PropPair {
	return types.PropPair{K: k, V: v}
}
