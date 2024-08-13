package goetf

type EncoderOpt func(*EncoderConfig)

// DefaultEncoderConfig creates a new default encoder configuration.
func DefaultEncoderConfig() *EncoderConfig {
	return &EncoderConfig{
		StringOverAtom: false,
	}
}

// An EncoderConfig struct to handle encoding.
type EncoderConfig struct {
	// Encode data as string over atom type
	StringOverAtom bool
}

// WithStringOverAtom tells the encoder to always encode strings as ETF String.
//
// StringOverAtom default value is false.
func WithStringOverAtom(b bool) EncoderOpt {
	return func(ec *EncoderConfig) {
		ec.StringOverAtom = b
	}
}
