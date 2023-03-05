package api

import (
	"testing"
)

func TestLogging(t *testing.T) {
	SetupConfig()
	// we are just going to call the logger a bunch of times; we don't capture
	// any output, so pretty much just make sure it doesn't NPE or anything

	// set up a recoverer for panic mode logs (!)
	defer func() {
		if r := recover(); r != nil {
			Log(LogLevelTrace, "test", "recovered from the panic", nil)
		}
	}()

	options := &LogOptions{
		CallingFile: "log_test.go",
		CallingFunc: "TestLogging",
		ExtraData: map[string]interface{}{
			"test": true,
		},
	}
	Log(LogLevelTrace, "test", "trace level", options)
	Log(LogLevelDebug, "test", "debug level", options)
	Log(LogLevelInfo, "test", "info level", options)
	Log(LogLevelWarn, "test", "warn level", options)
	Log(LogLevelError, "test", "error level", options)
	Log(LogLevelFatal, "test", "fatal level", options)
	Log(LogLevelPanic, "test", "panic level", options)
	Log("unknown", "test", "unknown, so warn level", options)
}
