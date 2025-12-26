package utils

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// DefaultLogLevel is the default logging level.
const DefaultLogLevel = zerolog.ErrorLevel

// Log is the application logger. Use SetLogLevel to configure verbosity.
var Log zerolog.Logger

func init() {
	initLogger(DefaultLogLevel)
}

// initLogger initializes the logger with the specified level.
func initLogger(level zerolog.Level) {
	Log = zerolog.New(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC822Z,
		}).
		Level(level).
		With().
		Timestamp().
		Logger()
}

// SetLogLevel sets the log level based on the number of times the verbose flag is used.
func SetLogLevel(cmd *cobra.Command, _ []string) {
	verbosity := 0
	if v, err := cmd.Flags().GetCount("verbose"); err == nil {
		verbosity = v
	} else if v, err := cmd.PersistentFlags().GetCount("verbose"); err == nil {
		verbosity = v
	} else if v, err := cmd.InheritedFlags().GetCount("verbose"); err == nil {
		verbosity = v
	}
	level := DefaultLogLevel - zerolog.Level(verbosity)
	if level < zerolog.TraceLevel {
		level = zerolog.TraceLevel
	}
	zerolog.SetGlobalLevel(level)
	initLogger(level)
}
