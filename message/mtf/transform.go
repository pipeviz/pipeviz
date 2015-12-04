/*
Package mtf provides a common interface for performing transformations on
pipeviz messages. These transformations can be used to perform structured
upgrades, fixes, or other modifications.
*/
package mtf

import (
	"fmt"
	"io"
)

// A TransformerFunc takes an input pipeviz message, applies a transformation,
// and returns the result.
//
// TransformerFuncs should be tightly focused on a single transformation, with
// the expectation that a series of them may be applied end-to-end in a
// pipeline. As such, transformations should generally be idempotent (though it
// is not required).
//
// Further, TransformerFuncs should not assume that it is appropriate to make a
// change to a message simply because they have been called. They should first
// verify a transformation is appropriate; if not, the input data should be
// passed through unchanged, and false returned for the second parameter.
//
// If the input was invalid/unrecognized by the TransformerFunc, an
// InvalidInputMessageError should be returned; such an error will not halt the
// processing pipeline. Any other type of error will halt processing.
type TransformerFunc func(io.Reader) (result []byte, changed bool, err error)

type InvalidInputMessageError struct {
	Message string
}

type NamedTransformer interface {
	// Name returns the name of the transformer. This is human-oriented and
	// guaranteed to be unique.
	Name() string
	Transform(io.Reader) (result []byte, changed bool, err error)
}

type namedTransformer struct {
	name string
	t    TransformerFunc
}

func (nt namedTransformer) Name() string {
	return nt.name
}

func (nt namedTransformer) Transform(in io.Reader) (result []byte, changed bool, err error) {
	return nt.t(in)
}

var tfmap = make(map[string]namedTransformer)

// Register receives transformers and an associated human-readable name. Panics
// if an attempt is made to register a name twice.
func Register(name string, t TransformerFunc) {
	if _, exists := tfmap[name]; exists {
		panic(fmt.Sprintf("transformer is already registered under name %q", name))
	}
	tfmap[name] = namedTransformer{name: name, t: t}
}

// GetTransformers retrieves each transformer given in the
func GetTransformers(names ...string) (tf []NamedTransformer, missing []string) {
	for _, name := range names {
		if nt, exists := tfmap[name]; exists {
			tf = append(tf, nt)
		} else {
			missing = append(missing, name)
		}
	}
	return
}
