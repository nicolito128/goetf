package goetf

import "fmt"

var (
	errMalformedAtomUTF8      = fmt.Errorf("malformed ETF. EttAtomUTF8")
	errMalformedSmallAtomUTF8 = fmt.Errorf("malformed ETF. EttSmallAtomUTF8")
	errMalformedString        = fmt.Errorf("malformed ETF. EttString")
	errMalformedCacheRef      = fmt.Errorf("malformed ETF. EttAtomCacheRef")
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
	errMalformedPid           = fmt.Errorf("malformed ETF. EttPid")
	errMalformedNewPid        = fmt.Errorf("malformed ETF. EttNewPid")
	errMalformedRef           = fmt.Errorf("malformed ETF. EttNewRef")
	errMalformedNewRef        = fmt.Errorf("malformed ETF. EttNewerRef")
	errMalformedPort          = fmt.Errorf("malformed ETF. EttPort")
	errMalformedNewPort       = fmt.Errorf("malformed ETF. EttNewPort")
	errMalformedFun           = fmt.Errorf("malformed ETF. EttNewFun")
	errMalformedExport        = fmt.Errorf("malformed ETF. EttExport")
	errMalformedUnknownType   = fmt.Errorf("malformed ETF. unknown type")
	errMalformed              = fmt.Errorf("malformed ETF")

	errInvalidDecodeV = fmt.Errorf("invalid decoding value or something is wrong")
)
