package goetf

import "github.com/philpearl/intern"

const DefaultCacheSize = 1024 * 1024

var defaultInternalCache = intern.New(DefaultCacheSize)

type DecoderOpt func(*DecoderConfig)

// DefaultDecoderConfig creates a new default decoder configuration.
func DefaultDecoderConfig() *DecoderConfig {
	return &DecoderConfig{
		CacheSize: DefaultCacheSize,
	}
}

type DecoderConfig struct {
	CacheSize int
}

// WithCacheSize tells the decoder to an specific size for the internal cache.
//
// CacheSize default value is 1048576 (1024*1024).
func WithCacheSize(size int) DecoderOpt {
	return func(ec *DecoderConfig) {
		ec.CacheSize = size
	}
}
