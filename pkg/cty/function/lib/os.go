package lib

import (
	"os"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var (
	// Hostname attempts to determine the current system's hostname as the
	// environment variable HOSTNAME is not available on every platform.
	Hostname = function.New(&function.Spec{
		VarParam: nil,
		Params:   nil,
		Type:     func(args []cty.Value) (cty.Type, error) { return cty.String, nil },
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			if hostname, ok := os.LookupEnv("HOSTNAME"); ok {
				return cty.StringVal(hostname), nil
			}
			if hostname, err := os.Hostname(); err == nil && len(hostname) > 0 {
				return cty.StringVal(hostname), nil
			}
			hostname, _ := os.LookupEnv("HOST")
			return cty.StringVal(hostname), nil
		},
	})
)
