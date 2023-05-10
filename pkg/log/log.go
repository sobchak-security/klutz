// Copyright (c) 2023 Remo Ronca 106963724+sobchak-security@users.noreply.github.com
// MIT License

package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// configuration keys for enhanced encoders
const (
	ConfigKeyTimeEncoderInt64Seconds = "int64seconds"
	ConfigKeyTimeEncoderShort        = "short"
)

var (
	// Levels provide a convenient way to list all supported log level strings.
	Levels = []string{
		// omitting DPanic and Panic
		zapcore.DebugLevel.String(),
		zapcore.InfoLevel.String(),
		zapcore.WarnLevel.String(),
		zapcore.ErrorLevel.String(),
		zapcore.FatalLevel.String(),
	}
)

// EpochInt64SecondsEncoder serializes a time.Time to a int64 number of seconds
// since the Unix epoch.
func EpochInt64SecondsEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(t.Unix())
}

// EpochShortTimeEncoder serializes a time.Time to short time string.
func EpochShortTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("06-01-02 15:04:05"))
}

// DevConfig returns a logger and atomic log level aimed at development
// environments with a highly opinionated configuration.
// NOTE this factory function panics if an error occurs.
func DevConfig() zap.Config {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return cfg
}

// StdConfig returns a logger and atomic log level aimed at development
// environments with a highly opinionated configuration.
// NOTE this factory function panics if an error occurs.
func StdConfig() zap.Config {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	return cfg
}

// LogWriterWrapper is a simple io.Writer wrapper of zap's Logger.
type LogWriterWrapper zap.Logger

// Write implements the io.Writer interface.
func (w *LogWriterWrapper) Write(p []byte) (int, error) {
	switch (*zap.Logger)(w).Level() {
	case zapcore.PanicLevel:
		(*zap.Logger)(w).Panic(string(p))
	case zapcore.ErrorLevel:
		(*zap.Logger)(w).Error(string(p))
	case zapcore.WarnLevel:
		(*zap.Logger)(w).Warn(string(p))
	case zapcore.DebugLevel:
		(*zap.Logger)(w).Debug(string(p))
	default:
		(*zap.Logger)(w).Info(string(p))
	}
	return len(p), nil
}
