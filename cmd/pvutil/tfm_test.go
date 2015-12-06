package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"
)

type tfmMock struct {
	tfm        *tfm
	expectExit int
	tst        *testing.T
}

// panicKey is used to differentiate expected panics from non-expected panics.
var panicKey = rand.Int()

func newMock(ee int, tst *testing.T) (mock *tfmMock) {
	mock = &tfmMock{
		expectExit: ee,
		tst:        tst,
		tfm: &tfm{
			exitFunc: func(i int) {
				if mock.expectExit != i {
					tst.Errorf("Expected os.Exit to be called with code %v, got %v\n", mock.expectExit, i)
					panic(panicKey)
				}
				panic(panicKey)
			},
		},
	}

	return
}

func panicRecover(t *testing.T) {
	err := recover()
	if val, ok := err.(int); !ok || val != panicKey {
		t.Fatalf("Unexpected panic: %s\n", err)
	}
}

// Adapted from runExample() in stdlib testing library
func captureStderr(expected string, t *testing.T, fatal bool) {
	// Capture stderr.
	stderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		fmt.Fprintln(stderr, err)
		os.Exit(1)
	}
	os.Stderr = w
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		r.Close()
		if err != nil {
			fmt.Fprintf(stderr, "error copying stderr pipe: %v\n", err)
			os.Exit(1)
		}
		outC <- buf.String()
	}()

	go func() {
		// Wait until we receive output before restoring global state
		out := <-outC

		// Close pipe, restore stderr
		w.Close()
		os.Stderr = stderr

		if g, e := strings.TrimSpace(out), strings.TrimSpace(expected); g != e {
			if fatal {
				t.Fatalf("got:\n%s\nwant:\n%s\n", g, e)
			} else {
				t.Errorf("got:\n%s\nwant:\n%s\n", g, e)
			}
		}
	}()
}

// Adapted from runExample() in stdlib testing library
func captureStdout(expected string, t *testing.T, fatal bool) {
	// Capture stdout.
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Stdout = w
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		r.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error copying stdout pipe: %v\n", err)
			os.Exit(1)
		}
		outC <- buf.String()
	}()

	go func() {
		// Wait until we receive output before restoring global state
		out := <-outC

		// Close pipe, restore stdout
		w.Close()
		os.Stdout = stdout

		if g, e := strings.TrimSpace(out), strings.TrimSpace(expected); g != e {
			if fatal {
				t.Fatalf("got:\n%s\nwant:\n%s\n", g, e)
			} else {
				t.Errorf("got:\n%s\nwant:\n%s\n", g, e)
			}
		}
	}()
}

func TestEmptyTransforms(t *testing.T) {
	m := newMock(1, t)

	// Should err, complaining about no transforms listed
	captureStderr("Must specify at least one transform to aply\n", t, true)

	defer panicRecover(t)
	m.tfm.Run(nil, nil)

	t.Fatalf("Expected exit from Run call")
}

func TstListTransforms(t *testing.T) {
	m := newMock(1, t)
	m.tfm.list = true

	// Should err, complaining about no transforms listed
	captureStderr("Must specify at least one transform to apply\n", t, true)

	defer panicRecover(t)
	m.tfm.Run(nil, nil)

	t.Fatalf("Expected exit from Run call")
}
