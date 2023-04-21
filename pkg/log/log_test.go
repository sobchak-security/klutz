// Copyright (c) 2023 Remo Ronca 106963724+sobchak-security@users.noreply.github.com
// MIT License

package log

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"go.uber.org/zap"
)

func testTmpFile(t *testing.T, content string) (*os.File, func()) {
	t.Helper()

	f, err := os.CreateTemp("", "test_*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(content)); err != nil {
		t.Error(err)
	}
	return f, func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
		if err := os.Remove(f.Name()); err != nil {
			t.Error(err)
		}
	}
}

func TestAdHoc(t *testing.T) {
	conf := []byte(`{
		"disableStacktrace": true,
		"encoding": "console",
		"level": "info",
		"outputPaths": ["stderr"],
		"encoderConfig": {
		  "levelKey": "level",
		  "levelEncoder": "capital",
		  "messageKey": "short_message",
		  "timeEncoder": "short",
		  "timeKey": "timestamp"
		},
		"initialFields": {
		  "host": "${HOSTNAME}"
		}
	}`)
	var m map[string]interface{}
	var cfg zap.Config

	if err := json.Unmarshal(conf, &m); err != nil {
		t.Fatal(err)
	}
	if err := (*ConfigWrapper)(&cfg).Unmarshal(m); err != nil {
		t.Fatal(err)
	}

	logger, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}

	logger.Debug("not included with info level")
	logger.Warn("a warning")

	// t.Error("intentional")
}

func TestUnmarshal(t *testing.T) {
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
			  }`, logfile.Name(), ConfigKeyTimeEncoderShort)
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
			  }`, logfile.Name(), ConfigKeyTimeEncoderInt64Seconds)
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
			if err := (*ConfigWrapper)(&tt.cfg).Unmarshal(m); err != nil {
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

func TestDevConfig(t *testing.T) {
	f, closer := testTmpFile(t, "")
	defer closer()

	cfg := DevConfig()
	cfg.OutputPaths = append(cfg.OutputPaths, f.Name())

	l, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}

	// 2023-04-21T09:41:58.212+0200	DEBUG	log/log_test.go:179	debug
	wantLog := `\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d\.\d{3}[+|-]\d{4}\s+DEBUG\s+log/log_test\.go:\d+\s+debug\s*$`

	l.Debug("debug")

	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	ok, err := regexp.MatchString(wantLog, string(b))
	if err != nil {
		t.Errorf("UnmarshalConfig(): MatchString failed %v", err)
	}
	if !ok {
		t.Errorf("UnmarshalConfig() sample log output: %q, want %q",
			string(b), wantLog)
	}

	// t.Error("intentional")
}

func TestStdConfig(t *testing.T) {
	f, closer := testTmpFile(t, "")
	defer closer()

	cfg := StdConfig()
	cfg.OutputPaths = append(cfg.OutputPaths, f.Name())

	l, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}

	// 2023-04-21T09:41:58.212+0200	[34mINFO[0m	info
	wantLog := `\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d\.\d{3}[+|-]\d{4}\s+.*INFO.*\s+info\s*$`

	l.Debug("debug")
	l.Info("info")

	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	ok, err := regexp.MatchString(wantLog, string(b))
	if err != nil {
		t.Errorf("UnmarshalConfig(): MatchString failed %v", err)
	}
	if !ok {
		t.Errorf("UnmarshalConfig() sample log output: %q, want %q",
			string(b), wantLog)
	}

	// t.Error("intentional")
}
