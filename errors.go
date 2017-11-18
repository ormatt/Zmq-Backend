package main

import log "github.com/sirupsen/logrus"
import "errors"

func reportError(errStr string, logger *log.Entry) error{
	logger.Error(errStr)
	return errors.New(errStr)
}