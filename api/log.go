package api

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Note: Once structured logging is confirmed in Go, we will want to consider what that looks like and
// maybe remove logrus

const (
	LogLevelTrace = "TRACE"
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
	LogLevelFatal = "FATAL"
	LogLevelPanic = "PANIC"
)

// LogOptions are optional options and fields for logging
type LogOptions struct {
	ExtraData map[string]interface{}

	// the two can be helpful when things are tough to debug, but shouldn't be required
	CallingFile string
	CallingFunc string
}

// Log sends a log to output
func Log(level string, key, message string, options *LogOptions) {
	log.SetFormatter((&log.JSONFormatter{}))
	if message == "" {
		message = key
	}
	log.WithFields(log.Fields{
		"key":     key,
		"message": message,
	})
	if options != nil {
		for k, v := range options.ExtraData {
			log.WithField(k, v)
		}
		if options.CallingFile != "" {
			log.WithField("file", options.CallingFile)
		}
		if options.CallingFunc != "" {
			log.WithField("func", options.CallingFunc)
		}
	}
	logOut := fmt.Sprintf("%s - %s", key, message)
	switch config.LogLevelOutput {
	case LogLevelTrace:
		log.Trace(logOut)
	case LogLevelDebug:
		log.Debug(logOut)
	case LogLevelInfo:
		log.Info(logOut)
	case LogLevelWarn:
		log.Warn(logOut)
	case LogLevelError:
		log.Error(logOut)
	case LogLevelFatal:
		log.Fatal(logOut)
	case LogLevelPanic:
		log.Panic(logOut)
	default:
		log.Warn(logOut)
	}

	// TODO: if we ever integrate with an external logger, if the key is a test key,
	// bail out before
}
