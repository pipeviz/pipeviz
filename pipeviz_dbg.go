// +build debug

package main

import (
	"expvar"
	"log"
	"log/syslog"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	logrus_syslog "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/tag1consulting/pipeviz/broker"
	"github.com/tag1consulting/pipeviz/represent"
	"github.com/tag1consulting/pipeviz/types/system"
)

func init() {
	startTime := time.Now().UTC()
	expvar.Publish("Uptime", expvar.Func(func() interface{} { return int64(time.Since(startTime)) }))

	expvar.Publish("Goroutines", expvar.Func(func() interface{} { return runtime.NumGoroutine() }))

	// subscribe to the broker in order to report data about current graph
	c := broker.Get().Subscribe()

	// Instantiate a real, empty graph to ensure the interface type is never nil when it might be called
	var g system.CoreGraph = represent.NewGraph()
	go func() {
		for latest := range c {
			g = latest
		}
	}()
	expvar.Publish("MsgId", expvar.Func(func() interface{} { return g.MsgId() }))

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func setUpLogging() {
	logrus.SetLevel(logrus.DebugLevel)

	// For now, either log to syslog OR stdout
	if *useSyslog {
		hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_DEBUG, "")
		if err == nil {
			logrus.AddHook(hook)
		} else {
			logrus.WithFields(logrus.Fields{
				"system": "main",
				"err":    err,
			}).Fatal("Could not connect to syslog, exiting")
		}
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:  true,
			DisableSorting: true,
		})
	}
}
