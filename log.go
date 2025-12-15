package main

import (
	"fmt"
	"github.com/go-logr/logr"
	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
	"strings"
)

// WailsLogSink implements the logr.LogSink interface by directing logs to the Wails logger.
// Since the Wails logger is string-based, we handle logr's structured key/value pairs by formatting them into a single string.
type WailsLogSink struct {
	wailsLog wailsLogger.Logger
	name     string
	values   []interface{} // Stored keys and values from WithValues
}

// NewWailsLogr returns a logr.Logger instance that writes its output to the provided Wails Logger.
func NewWailsLogr(wailsLog wailsLogger.Logger) logr.Logger {
	return logr.New(WailsLogSink{wailsLog: wailsLog})
}

// Init implements logr.LogSink.
// It is called once when the logger is constructed, but logr.RuntimeInfo currently only provides CallDepth, not the name.
func (s WailsLogSink) Init(info logr.RuntimeInfo) {
	// The name is accumulated via WithName, not set here, as logr.RuntimeInfo does not contain a name field.
}

// Enabled implements logr.LogSink. It always returns true, relying on the Wails logger's internal logic
// or the application's configured Wails LogLevel to filter messages.
// NOTE: For better performance, you could try to check the Wails logger's active log level here.
func (s WailsLogSink) Enabled(level int) bool {
	// Wails logger interface doesn't expose its current LogLevel for runtime checking,
	// so we assume it's always enabled for simplicity and rely on Wails' own filtering.
	return true
}

// Info implements logr.LogSink for non-error messages.
func (s WailsLogSink) Info(level int, msg string, keysAndValues ...interface{}) {

	// Combine stored values and current values
	allKeysAndValues := append(s.values, keysAndValues...)

	formattedMsg := s.formatMessage(msg, allKeysAndValues...)

	// Map logr V-levels to Wails levels:
	switch level {
	case 0:
		s.wailsLog.Info(formattedMsg)
	case 1:
		s.wailsLog.Debug(formattedMsg)
	case 2:
		s.wailsLog.Trace(formattedMsg)
	default:
		// Use Trace for V-levels > 2
		if level > 2 {
			s.wailsLog.Trace(formattedMsg)
		} else {
			// Negative or unusual V-levels default to Info
			s.wailsLog.Info(formattedMsg)
		}
	}
}

// Error implements logr.LogSink for error messages.
func (s WailsLogSink) Error(err error, msg string, keysAndValues ...interface{}) {

	// Combine stored values, current values, and the error object
	allKeysAndValues := append(s.values, keysAndValues...)
	allKeysAndValues = append(allKeysAndValues, "error", err.Error())

	formattedMsg := s.formatMessage(msg, allKeysAndValues...)

	s.wailsLog.Error(formattedMsg)
}

// WithValues implements logr.LogSink to return a new LogSink with additional key/value pairs.
func (s WailsLogSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	newValues := make([]interface{}, 0, len(s.values)+len(keysAndValues))
	newValues = append(newValues, s.values...)
	newValues = append(newValues, keysAndValues...)

	return WailsLogSink{
		wailsLog: s.wailsLog,
		name:     s.name,
		values:   newValues,
	}
}

// WithName implements logr.LogSink to return a new LogSink with the specified name appended.
func (s WailsLogSink) WithName(name string) logr.LogSink {
	newName := s.name
	if newName != "" {
		newName = newName + "/"
	}
	newName += name

	newSink := WailsLogSink{
		wailsLog: s.wailsLog,
		name:     newName,
		values:   s.values,
	}
	// Init is not called on WithName return, so we set the name directly
	return newSink
}

// V is required by logr but not used in this Wails adapter as we rely on the
// level passed to the Info method. We return a clone.
func (s WailsLogSink) V(level int) logr.LogSink {
	return s
}

// formatMessage converts the logr structured message into a Wails compatible string.
func (s WailsLogSink) formatMessage(msg string, keysAndValues ...interface{}) string {
	var builder strings.Builder

	// 1. Write Logger Name
	if s.name != "" {
		builder.WriteString("[")
		builder.WriteString(s.name)
		builder.WriteString("] ")
	}

	// 2. Write Message
	builder.WriteString(msg)

	// 3. Write Key/Value pairs
	if len(keysAndValues) > 0 {
		builder.WriteString(" | ")
		for i := 0; i < len(keysAndValues); i += 2 {
			key := keysAndValues[i]
			value := interface{}("")
			if i+1 < len(keysAndValues) {
				value = keysAndValues[i+1]
			}
			builder.WriteString(fmt.Sprintf("%v=%v", key, value))
			if i+2 < len(keysAndValues) {
				builder.WriteString(", ")
			}
		}
	}
	return builder.String()
}
