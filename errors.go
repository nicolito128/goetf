package goetf

import "fmt"

var (
	errMalformedAtomUTF8      = fmt.Errorf("malformed ETF. EttAtomUTF8")
	errMalformedSmallAtomUTF8 = fmt.Errorf("malformed ETF. EttSmallAtomUTF8")
	errMalformedString        = fmt.Errorf("malformed ETF. EttString")
	errMalformedNewFloat      = fmt.Errorf("malformed ETF. EttNewFloat")
	errMalformedFloat         = fmt.Errorf("malformed ETF. EttFloat")
	errMalformedSmallInteger  = fmt.Errorf("malformed ETF. EttSmallInteger")
	errMalformedInteger       = fmt.Errorf("malformed ETF. EttInteger")
	errMalformedSmallBig      = fmt.Errorf("malformed ETF. EttSmallBig")
	errMalformedLargeBig      = fmt.Errorf("malformed ETF. EttLargeBig")
	errMalformedList          = fmt.Errorf("malformed ETF. EttList")
	errMalformedSmallTuple    = fmt.Errorf("malformed ETF. EttSmallTuple")
	errMalformedLargeTuple    = fmt.Errorf("malformed ETF. EttLargeTuple")
	errMalformedMap           = fmt.Errorf("malformed ETF. EttMap")
	errMalformedBinary        = fmt.Errorf("malformed ETF. EttBinary")
	errMalformedBitBinary     = fmt.Errorf("malformed ETF. EttBitBinary")
	errMalformed              = fmt.Errorf("malformed ETF")
)
