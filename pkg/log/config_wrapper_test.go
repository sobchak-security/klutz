package log_test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	klib "github.com/sobchak-security/klutz/pkg/cty/function/lib"
	"github.com/sobchak-security/klutz/pkg/log"
)

func TestUnmarshalMap(t *testing.T) {
	type args struct {
		m map[string]interface{}
	}
	tests := []struct {
		name    string
		cfg     zap.Config
		conf    func(logfile *os.File) string
		args    args
		logF    func(*zap.Logger)
		wantLog string
		wantErr bool
	}{
		{
			name: "success: custom configuration with short time encoder",
			cfg:  zap.Config{},
			conf: func(logfile *os.File) string {
				return fmt.Sprintf(`{
				  "disableCaller": false,
				  "disableStacktrace": false,
				  "encoding": "console",
				  "errorOutputPaths": [
					"stderr"
				  ],
				  "level": "warn",
				  "outputPaths": [
					"stdout",
					%q
				  ],
				  "encoderConfig": {
					"callerEncoder": "full",
					"durationEncoder": "string",
					"levelEncoder": "capital",
					"levelKey": "level",
					"messageKey": "short_message",
		
					"timeEncoder": %q,
					"timeKey": "timestamp"
				  },
				  "initialFields": {
					"host": "${HOSTNAME}",
					"version": "1.1"
				  }
			  }`, logfile.Name(), log.ConfigKeyTimeEncoderShort)
			},
			logF: func(l *zap.Logger) {
				l.Info("info")
				l.Error("error", zap.Duration("_duration", 3*time.Hour+5*time.Minute+7*time.Second))
			},
			wantLog: `\d\d-\d\d-\d\d \d\d:\d\d:\d\d\s+ERROR\s+error\s+{"host": ".+", "version": "1.1", "_duration": "3h5m7s"}`,
		},
		{
			name: "success: custom configuration with int64 time encoder",
			cfg:  zap.Config{},
			conf: func(logfile *os.File) string {
				return fmt.Sprintf(`{
				  "disableCaller": true,
				  "disableStacktrace": true,
				  "encoding": "console",
				  "level": "info",
				  "outputPaths": [
					%q
				  ],
				  "encoderConfig": {
					"durationEncoder": "string",
					"levelKey": "level",
					"messageKey": "short_message",
		
					"timeEncoder": %q,
					"timeKey": "timestamp"
				  }
			  }`, logfile.Name(), log.ConfigKeyTimeEncoderInt64Seconds)
			},
			logF: func(l *zap.Logger) {
				l.Info("info")
			},
			wantLog: `\d{8}	info`,
		},
		{
			name: "failure: invalid level",
			cfg:  zap.Config{},
			args: args{
				m: map[string]interface{}{
					"level": "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, closer := testTmpFile(t, "")
			defer closer()

			m := tt.args.m
			if m == nil {
				if err := json.Unmarshal([]byte(tt.conf(f)), &m); err != nil {
					t.Fatal(err)
				}
			}
			if err := (*log.ConfigWrapper)(&tt.cfg).UnmarshalMap(m); err != nil {
				if !tt.wantErr {
					t.Errorf("UnmarshalConfig() error = %v, wantErr %v", err, true)
				}
				return
			}

			logger, err := tt.cfg.Build()
			if err != nil {
				t.Fatal(err)
			}
			tt.logF(logger)

			b, err := os.ReadFile(f.Name())
			if err != nil {
				t.Fatal(err)
			}
			ok, err := regexp.MatchString(tt.wantLog, string(b))
			if err != nil {
				t.Errorf("UnmarshalConfig(): MatchString failed %v", err)
			}
			if !ok {
				t.Errorf("UnmarshalConfig() sample log output: %q, want %q",
					string(b), tt.wantLog)
			}
			// t.Error("intentional")
		})
	}
}

func TestEncoderConfigWrapperUnmarshalHCL(t *testing.T) {
	ctx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"hostname": klib.Hostname,
		},
		Variables: map[string]cty.Value{},
	}
	defaultEncoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig := zapcore.EncoderConfig{
		EncodeCaller:   defaultEncoderConfig.EncodeCaller,
		EncodeDuration: defaultEncoderConfig.EncodeDuration,
		EncodeLevel:    defaultEncoderConfig.EncodeLevel,
		EncodeTime:     defaultEncoderConfig.EncodeTime,
	}

	type args struct {
		ctx *hcl.EvalContext
		// body hcl.Body
	}
	tests := []struct {
		name    string
		conf    string
		args    args
		want    zapcore.EncoderConfig
		wantErr bool
	}{
		{
			name: "success: empty encoder config",
			args: args{ctx: ctx},
			conf: ``,
			want: encoderConfig,
		},
		{
			name: "success: default development config",
			args: args{ctx: ctx},
			conf: func() string {
				return fmt.Sprintf(`
					time_key = "T"
					level_key = "L"
					name_key = "N"
					caller_key = "C"
					function_key = %q
					message_key = "M"
					stacktrace_key = "S"
					line_ending = %q
					level_encoder = "capital"
					time_encoder = "iso8601"
					duration_encoder = "string"
					caller_encoder = "short"
				`, zapcore.OmitKey, zapcore.DefaultLineEnding)
			}(),
			want: zap.NewDevelopmentEncoderConfig(),
		},
		{
			name: "success: custom short time encoder config",
			args: args{ctx: ctx},
			conf: func() string {
				return fmt.Sprintf(`time_key = "T"
					time_encoder = %q`, log.ConfigKeyTimeEncoderShort)
			}(),
			want: zapcore.EncoderConfig{
				TimeKey:    "T",
				EncodeTime: log.EpochShortTimeEncoder,

				EncodeCaller:   defaultEncoderConfig.EncodeCaller,
				EncodeDuration: defaultEncoderConfig.EncodeDuration,
				EncodeLevel:    defaultEncoderConfig.EncodeLevel,
			},
		},
		{
			name: "success: custom int64 time encoder config",
			args: args{ctx: ctx},
			conf: func() string {
				return fmt.Sprintf(`name_encoder = "full"
					time_encoder = %q`, log.ConfigKeyTimeEncoderInt64Seconds)
			}(),
			want: zapcore.EncoderConfig{
				EncodeName: zapcore.FullNameEncoder,
				EncodeTime: log.EpochInt64SecondsEncoder,

				EncodeCaller:   defaultEncoderConfig.EncodeCaller,
				EncodeDuration: defaultEncoderConfig.EncodeDuration,
				EncodeLevel:    defaultEncoderConfig.EncodeLevel,
			},
		},
		{
			name:    "failure: invalid syntax",
			args:    args{ctx: ctx},
			conf:    `invalid = "invalid"`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags hcl.Diagnostics
			parser := hclparse.NewParser()

			hf, d := parser.ParseHCL([]byte(tt.conf), "")
			diags = append(diags, d...)
			if diags.HasErrors() {
				t.Fatalf("parsing config failed %v", diags)
			}

			var got zapcore.EncoderConfig

			if err := (*log.EncoderConfigWrapper)(&got).UnmarshalHCL(tt.args.ctx, hf.Body); err != nil {
				if !tt.wantErr {
					t.Errorf("UnmarshalHCL() error = %v, wantErr %v", err, true)
				}
				return
			}
			testCompareEncoderConfig(t, got, tt.want)
		})
	}
}

func TestConfigWrapperUnmarshalHCL(t *testing.T) {
	ctx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"hostname": klib.Hostname,
		},
		Variables: map[string]cty.Value{},
	}
	defaultEncoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig := zapcore.EncoderConfig{
		EncodeCaller:   defaultEncoderConfig.EncodeCaller,
		EncodeDuration: defaultEncoderConfig.EncodeDuration,
		EncodeLevel:    defaultEncoderConfig.EncodeLevel,
		EncodeTime:     defaultEncoderConfig.EncodeTime,
	}

	type args struct {
		ctx *hcl.EvalContext
		// body hcl.Body
	}
	tests := []struct {
		name    string
		conf    string
		args    args
		want    zap.Config
		wantErr bool
	}{
		{
			name: "success: empty config",
			args: args{ctx: ctx},
			conf: ``,
			want: zap.Config{
				Level:         zap.NewAtomicLevel(),
				Encoding:      zap.NewProductionConfig().Encoding,
				EncoderConfig: encoderConfig,
			},
		},
		{
			name: "success: only initial fields",
			args: args{ctx: ctx},
			conf: `initial_fields = {
				version = "1.1"
				host = hostname()
				}`,
			want: zap.Config{
				Level:         zap.NewAtomicLevel(),
				Encoding:      zap.NewProductionConfig().Encoding,
				EncoderConfig: encoderConfig,
			},
		},
		{
			name: "success: default development config",
			args: args{ctx: ctx},
			conf: func() string {
				return fmt.Sprintf(`
					level = "debug"
					development = true
					encoding = "console"
					output_paths = [ "stderr" ]
					error_output_paths = [ "stderr" ]
					encoder_config {
						time_key = "T"
						level_key = "L"
						name_key = "N"
						caller_key = "C"
						function_key = %q
						message_key = "M"
						stacktrace_key = "S"
						line_ending = %q
						level_encoder = "capital"
						time_encoder = "iso8601"
						duration_encoder = "string"
						caller_encoder = "short"
					}`, zapcore.OmitKey, zapcore.DefaultLineEnding)
			}(),
			want: zap.NewDevelopmentConfig(),
		},
		{
			name:    "failure: invalid log level",
			args:    args{ctx: ctx},
			conf:    `level = "invalid"`,
			wantErr: true,
		},
		{
			name:    "failure: invalid encoder config syntax",
			args:    args{ctx: ctx},
			conf:    `encoder_config { invalid = "invalid" }`,
			wantErr: true,
		},
		{
			name:    "failure: invalid config syntax",
			args:    args{ctx: ctx},
			conf:    `output_paths = "invalid"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags hcl.Diagnostics
			parser := hclparse.NewParser()

			hf, d := parser.ParseHCL([]byte(tt.conf), "")
			diags = append(diags, d...)
			if diags.HasErrors() {
				t.Fatalf("parsing config failed %v", diags)
			}

			var got zap.Config

			if err := (*log.ConfigWrapper)(&got).UnmarshalHCL(tt.args.ctx, hf.Body); err != nil {
				if !tt.wantErr {
					t.Errorf("UnmarshalHCL() error = %v, wantErr %v", err, true)
				}
				return
			}
			if got.Level.String() != tt.want.Level.String() {
				t.Errorf("NewConfigFromHCL [Level]: %+v, want %+v",
					got, tt.want)
			}
			if got.Development != tt.want.Development {
				t.Errorf("NewConfigFromHCL [Development]: %+v, want %+v",
					got, tt.want)
			}
			if got.Encoding != tt.want.Encoding {
				t.Errorf("NewConfigFromHCL [Encoding]: %+v, want %+v",
					got, tt.want)
			}
			if !reflect.DeepEqual(got.OutputPaths, tt.want.OutputPaths) {
				t.Errorf("NewConfigFromHCL [OutputPaths]: %+v, want %+v",
					got, tt.want)
			}
			if !reflect.DeepEqual(got.ErrorOutputPaths, tt.want.ErrorOutputPaths) {
				t.Errorf("NewConfigFromHCL [ErrorOutputPaths]: %+v, want %+v",
					got, tt.want)
			}
			zap.NewDevelopmentConfig()

			testCompareEncoderConfig(t, got.EncoderConfig, tt.want.EncoderConfig)

			if _, err := got.Build(); err != nil {
				t.Errorf("NewConfigFromHCL: resulting config unable to build logger - %v", err)
			}
		})
	}
}
