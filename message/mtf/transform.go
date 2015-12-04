/*
Package mtf provides a common interface for performing transformations on
pipeviz messages. These transformations can be used to perform structured
upgrades, fixes, or other modifications.
*/
package mtf

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
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

func (e InvalidInputMessageError) Error() string {
	return e.Message
}

type Transformer interface {
	// Name returns the name of the transformer. This is human-oriented and
	// guaranteed to be unique.
	Name() string
	// Transform is the main transformation function.
	Transform(io.Reader) (result []byte, changed bool, err error)
}

// TransformList is a slice of Transformers, and is itself also (recursively) a
// Transformer.
type TransformList []Transformer

// Name returns a list of all the contained Transforms' names, joined with a comma.
func (tl TransformList) Name() string {
	// If this ever goes critical path,
	var names []string
	for _, t := range tl {
		names = append(names, t.Name())
	}
	return strings.Join(names, ",")
}

// Transform performs each transform contained in the TransformList on the
// input and returns the final results.
func (tl TransformList) Transform(in io.Reader) (result []byte, changed bool, err error) {
	var didchange bool
	//bb := make([]byte, 0)
	var w io.Reader
	//w := bytes.NewBuffer(bb)
	w = in
	for _, t := range tl {
		// There is definitely a more elegant way of doing this, probably with an io.Pipe
		result, didchange, err = t.Transform(w)
		if err != nil {
			if _, ok := err.(InvalidInputMessageError); ok {
				// This is the only kind of error we skip
				continue
			}
			return nil, false, fmt.Errorf("transform aborted due to error while applying transform %q: %s\n", t.Name(), err)
		}
		if didchange {
			changed = true
		}

		w = bytes.NewBuffer(result)
	}

	//return w.Bytes(), changed, nil
	return
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

// Get retrieves the transformer associated with each name in the provided string slice,
// maintaining the order. Any names that could not be resolved to transformers are returned
// in the second return value.
func Get(names ...string) (tf []Transformer, missing []string) {
	for _, name := range names {
		if nt, exists := tfmap[name]; exists {
			tf = append(tf, nt)
		} else {
			missing = append(missing, name)
		}
	}
	return
}

// List enumerates the names of all registered transformers.
func List() (names []string) {
	for name := range tfmap {
		names = append(names, name)
	}
	sort.Strings(names)
	return
}
