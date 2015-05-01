// Provides a broker between the main state machine event loop and the
// listeners/consumers of that data. Package is kept separate in part to avoid
// cyclic dependencies.
//
// Logic here is also responsible for providing any necessary fanout to
// the various listeners.
package broker

import "github.com/sdboyer/pipeviz/represent"

// TODO switch to doing this all with DI instead, i think
var singletonBroker GraphBroker

// Get returns the singleton graph broker that (is assumed to) consume from the
// main state machine's processing loop.
func Get() GraphBroker {
	return singletonBroker
}

// Set injects a GraphBroker for use as the main, singleton broker.
//
// Panics if it's called twice, and the main binary package calls it from
// main() - cheap, dirty way to maintain control (initially) :)
func Set(gb GraphBroker) {
	if singletonBroker != nil {
		panic("attempted to set the singleton broker twice")
	}
	singletonBroker = gb
}

type GraphSender chan<- represent.CoreGraph
type GraphReceiver <-chan represent.CoreGraph

type GraphBroker interface {
	// Subscribe creates a channel that receives all graphs passed into the
	// broker. In general, subscribers should avoid doing a lot of work in the
	// receiving goroutine, as it could block other subscribers.
	Subscribe() GraphReceiver

	// Unsubscribe immediately removes the provided channel from the subscribers
	// list, and closes the channel.
	//
	// If the provided channel is not in the subscribers list, it will have no
	// effect on the brokering behavior, and the channel will not be closed.
	Unsubscribe(GraphReceiver)

	// Fanout initiates a goroutine that fans out each graph passed through the
	// provided channel to all of its subscribers. New subscribers added after
	// the fanout goroutine starts are automatically incorporated.
	Fanout(GraphReceiver)
}
