package log_test

import (
	"fmt"
	"os"
	"testing"

	"go.uber.org/zap/zapcore"
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

func testCompareEncoderConfig(t *testing.T, got, want zapcore.EncoderConfig) {
	t.Helper()

	if got.TimeKey != want.TimeKey {
		t.Errorf("NewEncoderConfigFromHCL [TimeKey]: %+v, want %+v",
			got, want)
	}
	if got.MessageKey != want.MessageKey {
		t.Errorf("NewEncoderConfigFromHCL [MessageKey]: %+v, want %+v",
			got, want)
	}
	if got.LevelKey != want.LevelKey {
		t.Errorf("NewEncoderConfigFromHCL [LevelKey]: %+v, want %+v",
			got, want)
	}
	if got.NameKey != want.NameKey {
		t.Errorf("NewEncoderConfigFromHCL [NameKey]: %+v, want %+v",
			got, want)
	}
	if got.CallerKey != want.CallerKey {
		t.Errorf("NewEncoderConfigFromHCL [CallerKey]: %+v, want %+v",
			got, want)
	}
	if got.FunctionKey != want.FunctionKey {
		t.Errorf("NewEncoderConfigFromHCL [FunctionKey]: %+v, want %+v",
			got, want)
	}
	if got.StacktraceKey != want.StacktraceKey {
		t.Errorf("NewEncoderConfigFromHCL [StacktraceKey]: %+v, want %+v",
			got, want)
	}
	if got.SkipLineEnding != want.SkipLineEnding {
		t.Errorf("NewEncoderConfigFromHCL [SkipLineEnding]: %+v, want %+v",
			got, want)
	}
	if got.LineEnding != want.LineEnding {
		t.Errorf("NewEncoderConfigFromHCL [LineEnding]: %+v, want %+v",
			got, want)
	}
	if got.ConsoleSeparator != want.ConsoleSeparator {
		t.Errorf("NewEncoderConfigFromHCL [ConsoleSeparator]: %+v, want %+v",
			got, want)
	}

	// encoder checks work as long as the encoders involved are "static"

	if x, y := fmt.Sprintf("%p", got.EncodeLevel),
		fmt.Sprintf("%p", want.EncodeLevel); x != y {
		t.Errorf("NewEncoderConfigFromHCL [EncodeLevel]: %+v, want %+v",
			x, y)
	}
	if x, y := fmt.Sprintf("%p", got.EncodeTime),
		fmt.Sprintf("%p", want.EncodeTime); x != y {
		t.Errorf("NewEncoderConfigFromHCL [EncodeTime]: %+v, want %+v",
			x, y)
	}
	if x, y := fmt.Sprintf("%p", got.EncodeDuration),
		fmt.Sprintf("%p", want.EncodeDuration); x != y {
		t.Errorf("NewEncoderConfigFromHCL [EncodeDuration]: %+v, want %+v",
			x, y)
	}
	if x, y := fmt.Sprintf("%p", got.EncodeCaller),
		fmt.Sprintf("%p", want.EncodeCaller); x != y {
		t.Errorf("NewEncoderConfigFromHCL [EncodeCaller]: %+v, want %+v",
			x, y)
	}
	if x, y := fmt.Sprintf("%p", got.EncodeName),
		fmt.Sprintf("%p", want.EncodeName); x != y {
		t.Errorf("NewEncoderConfigFromHCL [EncodeName]: %+v, want %+v",
			x, y)
	}
}
