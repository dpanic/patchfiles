package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// configure will return instance of zap logger configuration, configured to be verbose or to use JSON formatting
func Setup(verbose bool) (logger *zap.Logger, err error) {
	level := zapcore.InfoLevel
	if verbose {
		level = zapcore.DebugLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "time",
			NameKey:        "logger",
			CallerKey:      "go",
			StacktraceKey:  "trace",
			LineEnding:     "\n",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
				callerName := caller.TrimmedPath()
				callerName = minWidth(callerName, " ", 20)

				enc.AppendString(callerName)
			},
			EncodeName: zapcore.FullNameEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: nil,
		InitialFields:    nil,
	}

	return config.Build()
}
