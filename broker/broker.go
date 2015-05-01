// Provides a broker between the main state machine event loop and the
// listeners/consumers of that data. Package is kept separate in part to avoid
// cyclic dependencies.
//
// Logic here is also responsible for providing any necessary fanout to
// the various listeners.
package broker

import (
	"sync"

	"github.com/sdboyer/pipeviz/represent"
)

// TODO switch to doing this all with DI instead, i think
var singletonBroker *GraphBroker = New()

// Get returns the singleton graph broker that (is assumed to) consume from the
// main state machine's processing loop.
func Get() *GraphBroker {
	return singletonBroker
}

type GraphSender chan<- represent.CoreGraph
type GraphReceiver <-chan represent.CoreGraph

type GraphBroker struct {
	// synchronizes access to the channel list
	// TODO be lock-free
	lock sync.RWMutex
	subs map[GraphReceiver]GraphSender
}

// New creates a pointer to a new, fully initialized GraphBroker.
func New() *GraphBroker {
	return &GraphBroker{lock: sync.RWMutex{}, subs: make(map[GraphReceiver]GraphSender)}
}

// Fanout initiates a goroutine that fans out each graph passed through the
// provided channel to all of its subscribers. New subscribers added after
// the fanout goroutine starts are automatically incorporated.
func (gb *GraphBroker) Fanout(input GraphReceiver) {
	go func() {
		for in := range input {
			// for now we just iterate straight through and send in one goroutine
			for k, c := range gb.subs {
				// take a read lock at each stage of the loop. this guarantees that a
				// sub/unsub can interrupt at each point; mostly this is crucial because
				// otherwise we run the risk of sending on a closed channel.
				gb.lock.RLock()
				if _, exists := gb.subs[k]; exists {
					c <- in
				}
				gb.lock.RUnlock()
			}
		}
	}()
}

// Subscribe creates a channel that receives all graphs passed into the
// broker. In general, subscribers should avoid doing a lot of work in the
// receiving goroutine, as it could block other subscribers.
func (gb *GraphBroker) Subscribe() GraphReceiver {
	gb.lock.Lock()

	// Unbuffered. Listeners must be careful not to do too much in their receiving
	// goroutine, lest they create a pileup!
	c := make(chan represent.CoreGraph, 0)
	gb.subs[c] = c

	gb.lock.Unlock()
	return c
}

// Unsubscribe immediately removes the provided channel from the subscribers
// list, and closes the channel.
//
// If the provided channel is not in the subscribers list, it will have no
// effect on the brokering behavior, and the channel will not be closed.
func (gb *GraphBroker) Unsubscribe(recv GraphReceiver) {
	gb.lock.Lock()
	if c, exists := gb.subs[recv]; exists {
		delete(gb.subs, recv)
		close(c)
	}
	gb.lock.Unlock()
}
