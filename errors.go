package goetf

import "fmt"

var (
	ErrMalformedAtomUTF8      = fmt.Errorf("malformed ETF. EttAtomUTF8")
	ErrMalformedSmallAtomUTF8 = fmt.Errorf("malformed ETF. EttSmallAtomUTF8")
	ErrMalformedString        = fmt.Errorf("malformed ETF. EttString")
	ErrMalformedCacheRef      = fmt.Errorf("malformed ETF. EttAtomCacheRef")
	ErrMalformedNewFloat      = fmt.Errorf("malformed ETF. EttNewFloat")
	ErrMalformedFloat         = fmt.Errorf("malformed ETF. EttFloat")
	ErrMalformedSmallInteger  = fmt.Errorf("malformed ETF. EttSmallInteger")
	ErrMalformedInteger       = fmt.Errorf("malformed ETF. EttInteger")
	ErrMalformedSmallBig      = fmt.Errorf("malformed ETF. EttSmallBig")
	ErrMalformedLargeBig      = fmt.Errorf("malformed ETF. EttLargeBig")
	ErrMalformedList          = fmt.Errorf("malformed ETF. EttList")
	ErrMalformedSmallTuple    = fmt.Errorf("malformed ETF. EttSmallTuple")
	ErrMalformedLargeTuple    = fmt.Errorf("malformed ETF. EttLargeTuple")
	ErrMalformedMap           = fmt.Errorf("malformed ETF. EttMap")
	ErrMalformedBinary        = fmt.Errorf("malformed ETF. EttBinary")
	ErrMalformedBitBinary     = fmt.Errorf("malformed ETF. EttBitBinary")
	ErrMalformedPid           = fmt.Errorf("malformed ETF. EttPid")
	ErrMalformedNewPid        = fmt.Errorf("malformed ETF. EttNewPid")
	ErrMalformedRef           = fmt.Errorf("malformed ETF. EttNewRef")
	ErrMalformedNewRef        = fmt.Errorf("malformed ETF. EttNewerRef")
	ErrMalformedPort          = fmt.Errorf("malformed ETF. EttPort")
	ErrMalformedNewPort       = fmt.Errorf("malformed ETF. EttNewPort")
	ErrMalformedFun           = fmt.Errorf("malformed ETF. EttNewFun")
	ErrMalformedExport        = fmt.Errorf("malformed ETF. EttExport")
	ErrMalformedUnknownType   = fmt.Errorf("malformed ETF. unknown type")
	ErrMalformed              = fmt.Errorf("malformed ETF")

	ErrNilDecodeValue = fmt.Errorf("value to decode is nil")
)
