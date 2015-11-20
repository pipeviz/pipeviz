// Package fixtures contains all JSON fixtures, converted into easily accessed
// and guaranteed-to-be-present Go source files. The conversion is performed
// by go-bindata, and is automatically triggered when appropriate from
// pipeviz's root Makefile.
package fixtures

// UGH UGH UGH this go-bindata directive doesn't work because go generate bypasses globbing
//go:generate go-bindata -pkg $GOPACKAGE -o bindata_fixtures.go */*.json
