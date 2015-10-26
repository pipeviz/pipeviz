// +build !debug

package main

import (
	"log/syslog"

	"github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	logrus_syslog "github.com/tag1consulting/pipeviz/Godeps/_workspace/src/github.com/Sirupsen/logrus/hooks/syslog"
)

func setUpLogging() {
	// For now, either log to syslog OR stdout
	if *useSyslog {
		hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
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
