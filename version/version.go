// Package version exists solely so that any part of the application can access the compile time-determined version.
package version

var v = "dev"

// Version returns the version set at compile time via ldflags.
//
// For this to behave correctly, `-ldflags "-X github.com/pipeviz/pipeviz/version.v=$VERSION"`
// should be passed, where $VERSION is the appropriate version string.
func Version() string {
	return v
}
