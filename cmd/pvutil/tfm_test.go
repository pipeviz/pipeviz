package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// StdoutEquals verifies that a string equal to the provided string is emitted
// during the execution of the provided function.
func StdoutEquals(t *testing.T, expected string, f func()) bool {
	outC := make(chan string)
	done := make(chan struct{})
	captureStdout(outC, done)

	f()

	// Give the reader chans a tick to catch up
	time.Sleep(1 * time.Millisecond)
	done <- struct{}{}
	out := <-outC

	if g, e := strings.TrimSpace(out), strings.TrimSpace(expected); g != e {
		t.Errorf("got:\n%s\nwant:\n%s\n", g, e)
		return false
	}
	return true
}

func StderrEquals(t *testing.T, expected string, f func()) bool {
	outC := make(chan string)
	done := make(chan struct{})
	captureStderr(outC, done)

	f()

	// Give the reader chans a tick to catch up
	time.Sleep(1 * time.Millisecond)
	done <- struct{}{}
	out := <-outC

	if g, e := strings.TrimSpace(out), strings.TrimSpace(expected); g != e {
		t.Errorf("got:\n%s\nwant:\n%s\n", g, e)
		return false
	}
	return true
}

// Adapted from runExample() in stdlib testing library
func captureStderr(outC chan<- string, quit <-chan struct{}) {
	// Capture stderr.
	stderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Stderr = w
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
		// Close pipe, restore stderr
		r.Close()
		w.Close()
		os.Stderr = stderr

		// write whatever we have to the out channel
		outC <- buf.String()
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

func TestEmptyTransforms(t *testing.T) {
	tfm := tfm{}

	// Should err, complaining about no transforms listed
	StderrEquals(t, "Must specify at least one transform to apply\n", func() {
		if exit := tfm.Run(nil); exit != 1 {
			t.Errorf("Expected exit code 1, got %v", exit)
		}
	})
}

func TestListTransforms(t *testing.T) {
	tfm := tfm{}
	tfm.list = true

	StdoutEquals(t, "ensure-client\nidentity", func() {
		if exit := tfm.Run(nil); exit != 0 {
			t.Errorf("Expected exit code 0, got %v", exit)
		}
	})
}

func TestOnlyWrongTransformsFails(t *testing.T) {
	tfm := tfm{}
	// a random name; won't match anything real
	tfm.transforms = strconv.Itoa(rand.Int())

	if exit := tfm.Run(nil); exit != 1 {
		t.Errorf("Expected exit code 1, got %v", exit)
	}
}
