// +build debug

package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func runProfiler() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}
