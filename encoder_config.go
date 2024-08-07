package goetf

type EncoderOpt func(*EncoderConfig)

func DefaultEncoderConfig() *EncoderConfig {
	c := new(EncoderConfig)
	c.StringOverAtom = false
	return c
}

// An EncoderConfig struct to handle encoding.
type EncoderConfig struct {
	// Encode data as string over atom type
	StringOverAtom bool
}

// WithStringOverAtom tells the encoder to always encode strings as ETF String.
func WithStringOverAtom(c *EncoderConfig) {
	c.StringOverAtom = true
}
