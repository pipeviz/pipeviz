// +build debug

package broker

import "expvar"

func init() {
	// Publish count of websocket clients
	expvar.Publish("BrokerCount", expvar.Func(func() interface{} { return brokerCount }))
	expvar.Publish("MainBrokerSubs", expvar.Func(func() interface{} { return len(singletonBroker.subs) }))
}
