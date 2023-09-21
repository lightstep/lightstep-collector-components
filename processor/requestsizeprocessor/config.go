package requestsizeprocessor

import (
	"go.opentelemetry.io/collector/component"
)

type Config struct {
	// HeaderName should be set to the same key that
	// services will check for the size of requests.
	// By default lightstep downstream services will
	// check lightstep-uncompressed-bytes for the value.
	HeaderName string `mapstructure:"header_name"`
}

func CreateDefaultConfig() component.Config {
	return &Config{
		HeaderName: "lightstep-uncompressed-bytes",
	}
}