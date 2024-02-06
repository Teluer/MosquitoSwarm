package util

import (
	log "github.com/sirupsen/logrus"
)

func RecoverAndLog(flowName string) {
	err := recover()
	if err != nil {
		log.Error(flowName, " flow recovered from error: ", err)
	}
}

// RecoverAndLogAndDo recovers from panic, logs the error, and executes an action.
// flowName - a prefix for the log message.
// action - any function to execute after recovering, e.g. transaction rollback.
func RecoverAndLogAndDo(flowName string, action func()) {
	err := recover()
	if err != nil {
		log.Error(flowName, " flow recovered from error: ", err)
		action()
	}
}
