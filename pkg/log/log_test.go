// Copyright (c) 2023 Remo Ronca 106963724+sobchak-security@users.noreply.github.com
// MIT License

package log_test

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"

	"github.com/sobchak-security/klutz/pkg/log"
	"go.uber.org/zap"
)

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
	if err := (*log.ConfigWrapper)(&cfg).UnmarshalMap(m); err != nil {
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

func TestDevConfig(t *testing.T) {
	f, closer := testTmpFile(t, "")
	defer closer()

	cfg := log.DevConfig()
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

	cfg := log.StdConfig()
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

func TestLogWriterWrapperConfig(t *testing.T) {
	f, closer := testTmpFile(t, "")
	defer closer()

	cfg := log.StdConfig()
	cfg.OutputPaths = append(cfg.OutputPaths, f.Name())

	l, err := cfg.Build()
	if err != nil {
		t.Fatal(err)
	}

	// 2023-04-21T09:41:58.212+0200	[34mINFO[0m	info
	wantLog := `\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d\.\d{3}[+|-]\d{4}\s+.*INFO.*\s+info\s*$`

	cfg.Level.SetLevel(zap.InfoLevel)
	(*log.LogWriterWrapper)(l).Write([]byte("msg"))

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
