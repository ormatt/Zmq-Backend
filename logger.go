package main

import (
	"os"
	log "github.com/sirupsen/logrus"
)
func getNewLogger(loggerName string) *log.Entry{
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{QuoteEmptyFields:true,})

	contextLogger := log.WithFields(log.Fields{
		"Name": loggerName,
	})
	return contextLogger
}
