// +build debug

package main

import (
	log "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}
