// Package broker provides a broker between the main state machine event loop
// and the listeners/consumers of that data. Package is kept separate in part
// to avoid cyclic dependencies.
//
// Logic here is also responsible for providing any necessary fanout to
// the various listeners.
package broker

import (
	"sync"
	"sync/atomic"

	log "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tag1consulting/pipeviz/types/system"
)

// TODO switch to doing this all with DI instead, i think
var singletonBroker = New()

// Package-level var to count the total number of brokers that have been created.
// Purely for instrumentation, logging purposes.
var brokerCount uint64

// Get returns the singleton graph broker that (is assumed to) consume from the
// main state machine's processing loop.
func Get() *GraphBroker {
	return singletonBroker
}

type (
	// GraphSender is a channel over which graphs can be sent. They are
	// the input end to a fanout broker.
	GraphSender chan<- system.CoreGraph
	// GraphReceiver is a channel over which graphs are received. They are the
	// output end of a fanout broker.
	GraphReceiver <-chan system.CoreGraph
	// GraphBroker acts as a broker between a single sending channel and
	// multiple receiving channels,ensuring that a single message sent on the
	// input side is propagated to all the channels on the receiving side.
	GraphBroker struct {
		// synchronizes access to the channel list
		// TODO be lock-free
		lock sync.RWMutex
		subs map[GraphReceiver]GraphSender
		id   uint64 // internal id, just for instrumentation
	}
)

// New creates a pointer to a new, fully initialized GraphBroker.
func New() *GraphBroker {
	gb := &GraphBroker{lock: sync.RWMutex{}, subs: make(map[GraphReceiver]GraphSender), id: atomic.AddUint64(&brokerCount, 1)}

	log.WithFields(log.Fields{
		"system":    "broker",
		"broker-id": gb.id,
	}).Debug("New graph broker created")
	return gb
}

// Fanout initiates a goroutine that fans out each graph passed through the
// provided channel to all of its subscribers. New subscribers added after
// the fanout goroutine starts are automatically incorporated.
func (gb *GraphBroker) Fanout(input GraphReceiver) {
	go func() {
		log.WithFields(log.Fields{
			"system":    "broker",
			"broker-id": gb.id,
		}).Debug("New fanout initiated from graph broker")

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
// TODO nonblocking impl
func (gb *GraphBroker) Subscribe() GraphReceiver {
	gb.lock.Lock()

	log.WithFields(log.Fields{
		"system":    "broker",
		"broker-id": gb.id,
	}).Debug("New subscriber to graph broker")

	// Unbuffered. Listeners must be careful not to do too much in their receiving
	// goroutine, lest they create a pileup!
	c := make(chan system.CoreGraph, 0)
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

	log.WithFields(log.Fields{
		"system":    "broker",
		"broker-id": gb.id,
	}).Debug("Receiver unsubscribed from graph broker")

	if c, exists := gb.subs[recv]; exists {
		delete(gb.subs, recv)
		close(c)
	}
	gb.lock.Unlock()
}
