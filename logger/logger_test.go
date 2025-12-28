package logger

import (
	"testing"

	"github.com/rs/zerolog"
)

func Test_GetLogger(t *testing.T) {
	logger := GetLogger()

	if logger == nil {
		t.Fatal("GetLogger() returned nil")
	}

	// Verify it's actually a zerolog logger by using it
	// This should not panic
	logger.Debug().Msg("test message")
}

func Test_GetLogger_returnsSameInstance(t *testing.T) {
	logger1 := GetLogger()
	logger2 := GetLogger()

	if logger1 != logger2 {
		t.Error("GetLogger() should return the same instance")
	}
}

func Test_SetLogLevel(t *testing.T) {
	tests := []struct {
		name          string
		verboseCount  int
		expectedLevel zerolog.Level
	}{
		{
			name:          "verbosity 0 sets error level",
			verboseCount:  0,
			expectedLevel: zerolog.ErrorLevel,
		},
		{
			name:          "verbosity 1 sets warn level",
			verboseCount:  1,
			expectedLevel: zerolog.WarnLevel,
		},
		{
			name:          "verbosity 2 sets info level",
			verboseCount:  2,
			expectedLevel: zerolog.InfoLevel,
		},
		{
			name:          "verbosity 3 sets debug level",
			verboseCount:  3,
			expectedLevel: zerolog.DebugLevel,
		},
		{
			name:          "verbosity 4 sets trace level",
			verboseCount:  4,
			expectedLevel: zerolog.TraceLevel,
		},
		{
			name:          "verbosity 5+ sets trace level",
			verboseCount:  10,
			expectedLevel: zerolog.TraceLevel,
		},
		{
			name:          "negative verbosity sets error level",
			verboseCount:  -1,
			expectedLevel: zerolog.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLogLevel(tt.verboseCount)

			// Check global level since we use zerolog.SetGlobalLevel()
			if zerolog.GlobalLevel() != tt.expectedLevel {
				t.Errorf("Expected level %v, got %v", tt.expectedLevel, zerolog.GlobalLevel())
			}
		})
	}
}
