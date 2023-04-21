// Copyright (c) 2023 Remo Ronca 106963724+sobchak-security@users.noreply.github.com
// MIT License

package log

import (
	"encoding/json"
	"fmt"
	"os"
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

// ConfigWrapper is a simple unmarshaling wrapper of zap's Config structure. Let
// cfg be a variable of zap.Config, then, admittedly, the rather clumsy way of
// calling the unmarshaling function would be
// _ = (*ConfigWrapper)(&tt.cfg).Unmarshal(m)
// ... deal with it.
type ConfigWrapper zap.Config

// Unmarshal supports unmarshaling a zap logger configuration from a map. This
// way, in a pre-proccessing step, different configuration file formats can be
// parsed. Also, enhancements, like new encoders, are processed. Lastly,
// environment variables are resolved, including an attempt, at ensuring the
// variable HOSTNAME (not available on every platform), is made.
func (cfg *ConfigWrapper) Unmarshal(m map[string]interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("UnmarshalConfig(): marshaling to JSON failed - %w", err)
	}
	if _, ok := os.LookupEnv("HOSTNAME"); !ok {
		var hostname string
		if hostname, err = os.Hostname(); err != nil || len(hostname) <= 0 {
			hostname, _ = os.LookupEnv("HOST")
		}
		if len(hostname) > 0 {
			os.Setenv("HOSTNAME", hostname)
		}
	}
	b = []byte(os.ExpandEnv(string(b)))
	if err := json.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("UnmarshalConfig(): unmarshaling (full) from JSON failed - %w", err)
	}

	var enhancements struct {
		EncoderConfig struct {
			TimeEncoder string
		}
	}
	if err := json.Unmarshal(b, &enhancements); err != nil {
		return fmt.Errorf("UnmarshalConfig(): unmarshaling (enhancements) from JSON failed - %w", err)
	}
	switch enhancements.EncoderConfig.TimeEncoder {
	case ConfigKeyTimeEncoderShort:
		cfg.EncoderConfig.EncodeTime = EpochShortTimeEncoder
	case ConfigKeyTimeEncoderInt64Seconds:
		cfg.EncoderConfig.EncodeTime = EpochInt64SecondsEncoder
	}
	return nil
}

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
