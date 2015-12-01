// +build !debug

package main

import (
	"log/syslog"

	"github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
)

func setUpLogging(s *srv) {
	// For now, either log to syslog OR stdout
	if s.useSyslog {
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
