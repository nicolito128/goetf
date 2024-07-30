package goetf

const (
	// Erlang external term format version.
	EtVersion = byte(131)
	// Erlang distribution header.
	EtDistHeader = byte(68)
)

type ExternalTagType byte

// Erlang external term tags.
const (
	EttAtomCacheRef ExternalTagType = 82

	EttAtomUTF8      ExternalTagType = (118)
	EttSmallAtomUTF8 ExternalTagType = (119)

	EttSmallInteger ExternalTagType = (97)
	EttInteger      ExternalTagType = (98)
	EttSmallBig     ExternalTagType = (110)
	EttLargeBig     ExternalTagType = (111)

	EttNewFloat ExternalTagType = (70)
	EttFloat    ExternalTagType = (99)

	EttNewPort ExternalTagType = (89)
	EttPort    ExternalTagType = (102) // since OTP 23, only when BIG_CREATION flag is set

	EttV4Port ExternalTagType = (120)

	EttSmallTuple ExternalTagType = (104)
	EttLargeTuple ExternalTagType = (105)

	EttMap ExternalTagType = (116) // 116 Arity Pairs | K1,V1,K2,V2,...

	EttNil ExternalTagType = (106) // Empyu list: []

	EttListImproper ExternalTagType = (18)  // to be able to encode improper lists like [a|b]
	EttString       ExternalTagType = (107) // used for lists with integers in the range 0..255
	EttList         ExternalTagType = (108)

	EttBitBinary ExternalTagType = (77)
	EttBinary    ExternalTagType = (109)

	EttNewPid         ExternalTagType = (88) // since OTP 23, only when BIG_CREATION flag is set
	EttNewerReference ExternalTagType = (90) // since OTP 21, only when BIG_CREATION flag is set
	EttPid            ExternalTagType = (103)
	EttNewReference   ExternalTagType = (114)

	EttNewFun ExternalTagType = (112)
	EttExport ExternalTagType = (113)
	EttFun    ExternalTagType = (117) // legacy

	EttLocal ExternalTagType = (121) // OTP 26.0

	EttAtom      ExternalTagType = (100) // deprecated
	EttRef       ExternalTagType = (101) // deprecated
	EttSmallAtom ExternalTagType = (115) // deprecated
)

func (ett ExternalTagType) Byte() byte {
	return byte(ett)
}

var tagNames = map[ExternalTagType]string{
	EttAtom:          "ATOM_EXT",
	EttAtomUTF8:      "ATOM_UTF8_EXT",
	EttBinary:        "BINARY_EXT",
	EttBitBinary:     "BIT_BINARY_EXT",
	EttAtomCacheRef:  "ATOM_CACHE_REF",
	EttExport:        "EXPORT_EXT",
	EttFloat:         "FLOAT_EXT",
	EttFun:           "FUN_EXT",
	EttInteger:       "INTEGER_EXT",
	EttLargeBig:      "LARGE_BIG_EXT",
	EttLargeTuple:    "LARGE_TUPLE_EXT",
	EttList:          "LIST_EXT",
	EttNewFloat:      "NEW_FLOAT_EXT",
	EttNewFun:        "NEW_FUN_EXT",
	EttNewReference:  "NEW_REFERENCE_EXT",
	EttNil:           "NIL_EXT",
	EttPid:           "PID_EXT",
	EttPort:          "PORT_EXT",
	EttRef:           "REFERENCE_EXT",
	EttSmallAtom:     "SMALL_ATOM_EXT",
	EttSmallAtomUTF8: "SMALL_ATOM_UTF8_EXT",
	EttSmallBig:      "SMALL_BIG_EXT",
	EttSmallInteger:  "SMALL_INTEGER_EXT",
	EttSmallTuple:    "SMALL_TUPLE_EXT",
	EttString:        "STRING_EXT",
	EttV4Port:        "V4_PORT_EXT",
	EttLocal:         "LOCAL_EXT",
}

func TagString(ett ExternalTagType) string {
	if tag, ok := tagNames[ett]; ok {
		return tag
	}

	return ""
}

func IsValidEtt(b byte) bool {
	tag := TagString(ExternalTagType(b))
	return tag != ""
}
