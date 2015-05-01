package broker

import (
	"testing"
	"time"

	"github.com/sdboyer/pipeviz/represent"
)

// Make sure brokers are initialized correctly
func TestNew(t *testing.T) {
	br := New()
	if br.subs == nil {
		t.Error("map in new broker is not initialized")
	}
}

// Ensures that all broker behavior works as expected
func TestBroker(t *testing.T) {
	g := represent.NewGraph()
	br := New()

	input := make(chan represent.CoreGraph, 0)

	tallyChan := make(chan represent.CoreGraph, 0)
	var tally int

	lstn := func(c <-chan represent.CoreGraph) {
		for g := range c {
			tallyChan <- g
		}
	}

	// kick off the tallying channel, the final target
	go func() {
		for _ = range tallyChan {
			tally++
		}
	}()

	// Kick off fanout goroutine with no listeners registered
	br.Fanout(input)
	input <- g
	// give a bit of time for the channels to pass it through
	time.Sleep(time.Millisecond)

	if tally != 0 {
		t.Fatalf("Broker has no listeners, fanout should not have increased the tally, but somehow %v message made it through", tally)
	}

	// Now add one listener
	go lstn(br.Subscribe())

	input <- g
	time.Sleep(time.Millisecond)

	if tally != 1 {
		t.Fatalf("Fanout to single listener failed; should have gotten one message, but got %v", tally)
	}
	tally = 0

	// Add a second listener, which should be incorporated by the already-running fanout goroutine
	l2 := br.Subscribe()
	go lstn(l2)
	if len(br.subs) != 2 {
		t.Fatal("Should be two channels in the subs list after second subscription, but got %v", len(br.subs))
	}

	// Kick off second fanout goroutine
	input2 := make(chan represent.CoreGraph, 0)
	br.Fanout(input2)
	input <- g
	input2 <- g
	time.Sleep(time.Millisecond)

	if tally != 4 {
		t.Fatalf("Two fanouts to two listeners failed; should have gotten four messages, but got %v", tally)
	}
	tally = 0

	// Now unsubscribe the second listener
	br.Unsubscribe(l2)
	if len(br.subs) != 1 {
		t.Fatal("Should be just one channel in the subs list after unsubscription, but got %v", len(br.subs))
	}

	input <- g
	input2 <- g
	time.Sleep(time.Millisecond)

	if tally != 2 {
		t.Fatalf("Two fanouts to what should be one listener after unsubscribing failed; should have gotten one message, but got %v", tally)
	}

	close(input)
	close(input2)
	close(tallyChan)
}
