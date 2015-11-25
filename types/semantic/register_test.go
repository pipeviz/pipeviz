package semantic

import (
	"reflect"
	"testing"

	"github.com/pipeviz/pipeviz/types/system"
)

// TestUnifierExclusivity attempts to ensure that unifier functions are exclusive to
// a given type. Exceptions are allowed.
//
// A lot of the value of this test is that it gives us one (executable/provable) place
// in which to keep track of the exceptions. This is particularly true because
// Go1 functions *do not have* identity; reflect is not a part of the language
// specification, so while this may work, it is not guaranteed to. See
// http://stackoverflow.com/questions/9643205/how-do-i-compare-two-functions-for-pointer-equality-in-the-latest-go-weekly
func TestUnifierExclusivity(t *testing.T) {
	exceptions := []map[system.VType]bool{
		{
			"git-tag":    true,
			"git-branch": true,
			"git-result": true,
		},
	}

	// Loop over the unifyMap and ensure uniqueness.
	for typ, f := range unifyMap {
		fval := reflect.ValueOf(f).Pointer()

	inner:
		for ctyp, cf := range unifyMap {
			if typ != ctyp && fval == reflect.ValueOf(cf).Pointer() {
				for _, group := range exceptions {
					_, e1 := group[typ]
					_, e2 := group[ctyp]

					if e1 && e2 {
						continue inner
					}
				}
				t.Errorf("Types %q and %q share a unifier function", typ, ctyp)
			}
		}
	}
}

// TestResolverExclusivity attempts to ensure that resolver functions are exclusive to
// a given type. Exceptions are allowed.
//
// This has the same possibility of false positives (passes when it shouldn't) as
// TestUnifierExclusivity.
//
// TODO can we test for, or evidence, the case where multiple edge logics are encompassed
// under a single edge? As in the cases of resolveDataProvenance() and resolveListening()
func TestResolverExclusivity(t *testing.T) {
	// Loop over the resolveMap and ensure uniqueness.
	for typ, f := range resolveMap {
		fval := reflect.ValueOf(f).Pointer()
		for ctyp, cf := range resolveMap {
			if typ != ctyp && fval == reflect.ValueOf(cf).Pointer() {
				t.Errorf("Types %q and %q share unifier function", typ, ctyp)
			}
		}
	}
}
