package util

import (
	log "github.com/sirupsen/logrus"
	"time"
)

// LogFormatter is used to override the default time location in the logger.
type LogFormatter struct {
	log.Formatter
	Loc *time.Location
}

func (u LogFormatter) Format(e *log.Entry) ([]byte, error) {
	e.Time = e.Time.In(u.Loc)
	return u.Formatter.Format(e)
}
