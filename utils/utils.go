/*
Copyright © 2023 Jake Rogers <code@supportoss.org>
*/
package utils

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
	DefaultLogLevel = zerolog.ErrorLevel
)

var (
	Log = Logger(DefaultLogLevel)
)

// Logger returns a zerolog logger with a console writer.
func Logger(level zerolog.Level) zerolog.Logger {
	return zerolog.New(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC822Z,
		}).
		Level(level).
		With().
		Timestamp().
		// Caller().
		Logger()
}

// SetLogLevel sets the log level based on the number of times the verbose flag is used.
func SetLogLevel(cmd *cobra.Command, args []string) {
	verbosity, _ := cmd.Flags().GetCount("verbose")
	level := Log.GetLevel()
	Log = Logger(level - zerolog.Level(verbosity))
}
