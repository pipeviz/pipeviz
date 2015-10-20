// +build debug

package webapp

import "expvar"

func init() {
	// Publish count of websocket clients
	expvar.Publish("WebsockClients", expvar.Func(func() interface{} { return clientCount }))
}
