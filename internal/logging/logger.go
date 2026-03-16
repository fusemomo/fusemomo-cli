package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// Init initialises the global Zap logger.
// debug=true: development logger (colored console, debug+ level)
// debug=false: production logger (JSON, warn+ level)
func Init(debug bool) {
	var cfg zap.Config
	if debug {
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		cfg = zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	}
	// Always log to stderr.
	cfg.OutputPaths = []string{"stderr"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	var err error
	logger, err = cfg.Build()
	if err != nil {
		// Fallback to a no-op logger rather than panicking.
		logger = zap.NewNop()
	}
}

// Get returns the global logger. Initialises a production logger if Init
// has not been called yet.
func Get() *zap.Logger {
	if logger == nil {
		Init(false)
	}
	return logger
}

// RedactKey returns a redacted representation of an API key safe for logging.
// "fm_live_abcdef..." → "fm_***..."
func RedactKey(key string) string {
	if len(key) <= 7 {
		return "fm_***"
	}
	return key[:3] + "***"
}

// Sync flushes any buffered log entries. Call on shutdown.
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}
