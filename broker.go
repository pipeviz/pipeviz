package main

import (
	"sync"

	"github.com/sdboyer/pipeviz/broker"
	"github.com/sdboyer/pipeviz/represent"
)

type graphBroker struct {
	// synchronizes access to the channel list
	// TODO be lock-free
	lock sync.RWMutex
	subs map[broker.GraphReceiver]broker.GraphSender
}

func (gb *graphBroker) Fanout(input broker.GraphReceiver) {
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

func (gb *graphBroker) Subscribe() broker.GraphReceiver {
	gb.lock.Lock()

	// Unbuffered. Listeners must be careful not to do too much in their receiving
	// goroutine, lest they create a pileup!
	c := make(chan represent.CoreGraph, 0)
	gb.subs[c] = c

	gb.lock.Unlock()
	return c
}

func (gb *graphBroker) Unsubscribe(recv broker.GraphReceiver) {
	gb.lock.Lock()
	if c, exists := gb.subs[recv]; exists {
		delete(gb.subs, recv)
		close(c)
	}
	gb.lock.Unlock()
}
