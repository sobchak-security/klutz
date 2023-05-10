package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// encoderConfigHCL is a HCL-compatible representation of zapcore.encoderConfigHCL.
type encoderConfigHCL struct {
	MessageKey       string `hcl:"message_key,optional"`
	LevelKey         string `hcl:"level_key,optional"`
	TimeKey          string `hcl:"time_key,optional"`
	NameKey          string `hcl:"name_key,optional"`
	CallerKey        string `hcl:"caller_key,optional"`
	FunctionKey      string `hcl:"function_key,optional"`
	StacktraceKey    string `hcl:"stacktrace_key,optional"`
	LineEnding       string `hcl:"line_ending,optional"`
	EncodeLevel      string `hcl:"level_encoder,optional"`
	EncodeTime       string `hcl:"time_encoder,optional"`
	EncodeDuration   string `hcl:"duration_encoder,optional"`
	EncodeCaller     string `hcl:"caller_encoder,optional"`
	EncodeName       string `hcl:"name_encoder,optional"`
	ConsoleSeparator string `hcl:"console_separator,optional"`
	SkipLineEnding   bool   `hcl:"skip_line_ending,optional"`
}

func defaultZapEncoderConfig() zapcore.EncoderConfig {

	defaultEncoderConfig := zap.NewProductionEncoderConfig()

	return zapcore.EncoderConfig{
		EncodeCaller:   defaultEncoderConfig.EncodeCaller,
		EncodeDuration: defaultEncoderConfig.EncodeDuration,
		EncodeLevel:    defaultEncoderConfig.EncodeLevel,
		EncodeTime:     defaultEncoderConfig.EncodeTime,
	}
}
func (ech encoderConfigHCL) initZapEncoderConfig(zec *zapcore.EncoderConfig) {
	defaultEncoderConfig := defaultZapEncoderConfig()

	zec.MessageKey = ech.MessageKey
	zec.LevelKey = ech.LevelKey
	zec.TimeKey = ech.TimeKey
	zec.NameKey = ech.NameKey
	zec.CallerKey = ech.CallerKey
	zec.FunctionKey = ech.FunctionKey
	zec.StacktraceKey = ech.StacktraceKey
	zec.SkipLineEnding = ech.SkipLineEnding
	zec.LineEnding = ech.LineEnding
	zec.ConsoleSeparator = ech.ConsoleSeparator

	// see zap/zapcore/encoder.go
	// > Configure the primitive representations of common complex types. For
	// > example, some users may want all time.Times serialized as floating-point
	// > seconds since epoch, while others may prefer ISO8601 strings.
	// ...
	// > Unlike the other primitive type encoders, EncodeName is optional. The
	// > zero value falls back to FullNameEncoder.

	zec.EncodeCaller = defaultEncoderConfig.EncodeCaller
	zec.EncodeDuration = defaultEncoderConfig.EncodeDuration
	zec.EncodeLevel = defaultEncoderConfig.EncodeLevel
	// zec.EncodeName = defaultEncoderConfig.EncodeName
	zec.EncodeTime = defaultEncoderConfig.EncodeTime

	if len(ech.EncodeLevel) > 0 {
		_ = (&zec.EncodeLevel).UnmarshalText([]byte(ech.EncodeLevel))
	}
	if len(ech.EncodeTime) > 0 {
		switch ech.EncodeTime {
		case ConfigKeyTimeEncoderShort:
			zec.EncodeTime = EpochShortTimeEncoder
		case ConfigKeyTimeEncoderInt64Seconds:
			zec.EncodeTime = EpochInt64SecondsEncoder
		default:
			_ = (&zec.EncodeTime).UnmarshalText([]byte(ech.EncodeTime))
		}
	}
	if len(ech.EncodeDuration) > 0 {
		_ = (&zec.EncodeDuration).UnmarshalText([]byte(ech.EncodeDuration))
	}
	if len(ech.EncodeCaller) > 0 {
		_ = (&zec.EncodeCaller).UnmarshalText([]byte(ech.EncodeCaller))
	}
	if len(ech.EncodeName) > 0 {
		_ = (&zec.EncodeName).UnmarshalText([]byte(ech.EncodeName))
	}
}

// configHCL is a HCL-compatible representation of zap.configHCL.
type configHCL struct {
	// Sampling *zap.SamplingConfig `json:"sampling" yaml:"sampling"`
	Level    string `hcl:"level,optional"`
	Encoding string `hcl:"encoding,optional"`
	// EncoderConfig hcl.Body `hcl:"encoder_config,remain"`
	EncoderConfig     *encoderConfigHCL `hcl:"encoder_config,block"`
	OutputPaths       []string          `hcl:"output_paths,optional"`
	ErrorOutputPaths  []string          `hcl:"error_output_paths,optional"`
	InitialFields     map[string]string `hcl:"initial_fields,optional"`
	Development       bool              `hcl:"development,optional"`
	DisableCaller     bool              `hcl:"disable_caller,optional"`
	DisableStacktrace bool              `hcl:"disable_stacktrace,optional"`
}

func (ec configHCL) initZapConfig(zc *zap.Config) error {
	zc.OutputPaths = ec.OutputPaths
	zc.ErrorOutputPaths = ec.ErrorOutputPaths
	zc.Development = ec.Development
	zc.DisableCaller = ec.DisableCaller
	zc.DisableStacktrace = ec.DisableStacktrace

	if len(ec.Encoding) > 0 {
		zc.Encoding = ec.Encoding
	} else {
		// encoding has to be present for the config to be "buildable"
		zc.Encoding = zap.NewProductionConfig().Encoding
	}

	if len(ec.Level) > 0 {
		lvl, err := zapcore.ParseLevel(ec.Level)
		if err != nil {
			return fmt.Errorf(
				"UnmarshalHCL(): parsing log level %q failed - %w", ec.Level, err)
		}
		zc.Level = zap.NewAtomicLevelAt(lvl)
	} else {
		// it is tricky to figure out, whether an AtomicLevel is properly
		// initialized; better to always make sure a zap.Config contains
		// one
		zc.Level = zap.NewAtomicLevel()
	}

	if len(ec.InitialFields) > 0 {
		zc.InitialFields = make(map[string]interface{}, len(ec.InitialFields))
		for k, v := range ec.InitialFields {
			zc.InitialFields[k] = v
		}
	}

	if ec.EncoderConfig != nil {
		ec.EncoderConfig.initZapEncoderConfig(&zc.EncoderConfig)
	} else {
		zc.EncoderConfig = defaultZapEncoderConfig()
	}

	return nil
}
