package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
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

func (m *tfmMock) StdoutEquals(t *testing.T, expected string, f func()) bool {
	outC := make(chan string)
	done := make(chan struct{})
	captureStdout(outC, done)

	f()

	// Give the reader chans a tick to catch up
	time.Sleep(50 * time.Millisecond)
	done <- struct{}{}
	out := <-outC

	if g, e := strings.TrimSpace(out), strings.TrimSpace(expected); g != e {
		t.Errorf("got:\n%s\nwant:\n%s\n", g, e)
		return false
	}
	return true
}

//func (m *tfmMock) StderrEquals(t *testing.T, expected string, f func()) bool {

//}

func panicRecover(t *testing.T) {
	err := recover()
	if err != nil {
		if val, ok := err.(int); !ok || val != panicKey {
			t.Fatalf("Unexpected panic: %s\n", err)
		}
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
func captureStdout(outC chan<- string, quit <-chan struct{}) {
	// Capture stdout.
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Stdout = w
	var buf bytes.Buffer
	chanData := chanFromReader(r, 1024)
	go func() {
	loop:
		for {
			select {
			case <-quit:
				// quit signaled from outside - stop reading
				break loop
			case byt := <-chanData:
				// new data came in, append to buffer
				buf.Write(byt)
			}
		}
		// inner reader is definitely done
		// Close pipe, restore stdout
		r.Close()
		w.Close()
		os.Stdout = stdout

		// write whatever we have to the out channel
		outC <- buf.String()
	}()
}

func chanFromReader(reader io.Reader, pCap int) <-chan []byte {
	outChan := make(chan []byte)
	go func() {
		for {
			p := make([]byte, pCap)
			n, err := reader.Read(p)
			if err != nil {
				return
			}
			if n > 0 {
				outChan <- p[0:n]
			}
		}
	}()

	return outChan
}

func TstEmptyTransforms(t *testing.T) {
	m := newMock(1, t)

	// Should err, complaining about no transforms listed
	captureStderr("Must specify at least one transform to apply\n", t, true)

	//defer panicRecover(t)
	if exit := m.tfm.Run(nil); exit != 1 {
		t.Errorf("Expected exit code 1, got %v", exit)
	}
}

func TestListTransforms(t *testing.T) {
	m := newMock(0, t)
	m.tfm.list = true

	m.StdoutEquals(t, "ensure-client\nidentity", func() {
		if exit := m.tfm.Run(nil); exit != 0 {
			t.Errorf("Expected exit code 0, got %v", exit)
		}
	})
}
