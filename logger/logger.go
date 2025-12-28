// Copyright © 2025 Jake Rogers <code@supportoss.org>

// Package logger provides a zerolog-based logger with verbosity levels.
//
// Verbosity mapping:
//   - 0 (default): Error level
//   - -v (1): Warn level
//   - -vv (2): Info level
//   - -vvv (3): Debug level
//   - -vvvv (4+): Trace level
//
// The logger outputs to stderr with colored console formatting and RFC3339 timestamps.
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// sharedLogger is the package-level logger instance.
// It is initialized once at package load time and shared across all callers.
var sharedLogger zerolog.Logger

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	var output io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	sharedLogger = zerolog.New(output).
		With().
		Timestamp().
		Logger()
}

// GetLogger returns a pointer to the shared logger instance.
// All callers receive the same logger, so level changes via SetLogLevel
// affect all users of this logger.
func GetLogger() *zerolog.Logger {
	return &sharedLogger
}

// SetLogLevel adjusts the shared logger's level based on verbosity count.
// This is typically called from CLI flag processing.
//
// Mapping:
//   - 0: Error (default, quietest)
//   - 1: Warn
//   - 2: Info
//   - 3: Debug
//   - 4+: Trace (most verbose)
func SetLogLevel(verboseCount int) {
	var level zerolog.Level
	switch verboseCount {
	case 1:
		level = zerolog.WarnLevel
	case 2:
		level = zerolog.InfoLevel
	case 3:
		level = zerolog.DebugLevel
	default:
		if verboseCount >= 4 {
			level = zerolog.TraceLevel
		} else {
			level = zerolog.ErrorLevel
		}
	}
	zerolog.SetGlobalLevel(level)
}

// Disable disables all logging output.
// This is useful for interactive modes (e.g., TUI) where log output
// would interfere with the display.
func Disable() {
	zerolog.SetGlobalLevel(zerolog.Disabled)
}
