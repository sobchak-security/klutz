// Copyright (c) 2023 Remo Ronca 106963724+sobchak-security@users.noreply.github.com
// MIT License

package log

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ConfigWrapper is a simple unmarshaling wrapper of zap's Config structure. Let
// cfg be a variable of zap.Config, then, admittedly, the rather clumsy way of
// calling the unmarshaling function would be
// _ = (*ConfigWrapper)(&tt.cfg).Unmarshal(m)
// ... deal with it.
type ConfigWrapper zap.Config

// UnmarshalMap supports unmarshaling a zap logger configuration from a map. This
// way, in a pre-proccessing step, different configuration file formats can be
// parsed. Also, enhancements, like new encoders, are processed. Lastly,
// environment variables are resolved, including an attempt, at ensuring the
// variable HOSTNAME (not available on every platform), is made.
func (cw *ConfigWrapper) UnmarshalMap(m map[string]interface{}) error {
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
	if err := json.Unmarshal(b, cw); err != nil {
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
		cw.EncoderConfig.EncodeTime = EpochShortTimeEncoder
	case ConfigKeyTimeEncoderInt64Seconds:
		cw.EncoderConfig.EncodeTime = EpochInt64SecondsEncoder
	}
	return nil
}

// UnmarshalHCL processes a HCL configuration and returns a zap.Config and
// the unparsed remainder of body.
func (cw *ConfigWrapper) UnmarshalHCL(ctx *hcl.EvalContext, body hcl.Body) error {
	var cfg configHCL

	if diags := gohcl.DecodeBody(body, ctx, &cfg); diags.HasErrors() {
		return fmt.Errorf("UnmarshalHCL(): parsing log configuration failed - %w", diags)
	}

	if err := cfg.initZapConfig((*zap.Config)(cw)); err != nil {
		return fmt.Errorf("UnmarshalHCL(): initializing configuration failed - %w", err)
	}

	return nil
}

// EncoderConfigWrapper ...
type EncoderConfigWrapper zapcore.EncoderConfig

// UnmarshalHCL ...
func (ecw *EncoderConfigWrapper) UnmarshalHCL(ctx *hcl.EvalContext, body hcl.Body) error {
	var ec encoderConfigHCL

	if diags := gohcl.DecodeBody(body, ctx, &ec); diags.HasErrors() {
		return fmt.Errorf("UnmarshalHCL(): parsing log configuration failed - %w", diags)
	}

	ec.initZapEncoderConfig((*zapcore.EncoderConfig)(ecw))

	return nil
}
