package goetf

type EncoderOpt func(*EncoderConfig)

func DefaultEncoderConfig() *EncoderConfig {
	c := new(EncoderConfig)
	c.StringOverAtom = false
	return c
}

type EncoderConfig struct {
	// Encode data as string over atom type
	StringOverAtom bool
}

func WithStringOverAtom(c *EncoderConfig) {
	c.StringOverAtom = true
}
