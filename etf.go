package goetf

import (
	"fmt"
	"hash/crc32"
)

const (
	// Erlang external term format version
	Version = byte(131)
	// Erlang distribution header
	DistHeader = byte(68)
)

type ExternalTagType = byte

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

var crc32q = crc32.MakeTable(0xD5828281)

// Term is a piece of data of any data type.
//
// Ref: https://www.erlang.org/doc/system/data_types.html#terms
type Term = any

// Tuple type.
// A tuple is a compound data type with a fixed number of terms.
//
// Ref: https://www.erlang.org/doc/system/data_types.html#tuple
type Tuple = []Term

// List type.
// A list is a compound data type with a variable number of terms.
//
// Ref: https://www.erlang.org/doc/system/data_types.html#list
type List = []Term

// Map type.
// A map is a compound data type with a variable number of key-value associations.
//
// Ref: https://www.erlang.org/doc/system/data_types.html#map
type Map = map[Term]Term

// Alias type.
type Alias = Ref

// ListImproper as a workaround for the Erlang's improper list [a|b].
// Intended to be used to interact with Erlang.
type ListImproper = []Term

// Atom type.
// An atom is a literal, a constant with name.
//
// Ref: https://www.erlang.org/doc/system/data_types.html#atom
type Atom = string

// BitString type.
// A bit string value encodes as a binary (Erlang type: <<...>>)
//
// Ref: https://www.erlang.org/doc/system/data_types.html#bit-strings-and-binaries
type BitString = string

// String type.
// Strings are a shorthan for a character list (Erlang type: [$e, $t, $f]).
//
// Ref: https://www.erlang.org/doc/system/data_types.html#string
type String = string

// Pid type.
//
// Ref: https://www.erlang.org/doc/system/data_types.html#pid
type Pid struct {
	Node     Atom
	ID       uint64
	Serial   uint32
	Creation uint32
}

func (p Pid) String() string {
	if p == (Pid{}) {
		return "<0.0.0>"
	}

	n := uint32(0)
	if p.Node != "" {
		n = crc32.Checksum([]byte(p.Node), crc32q)
	}
	return fmt.Sprintf("<%08X.%d.%d>", n, int32(p.ID>>32), int32(p.ID))
}

// Port type.
//
// Ref: https://www.erlang.org/doc/system/data_types.html#port-identifier
type Port struct {
	Node     Atom
	ID       uint32
	Creation uint32
}

// Ref type.
//
// Link: https://www.erlang.org/doc/system/data_types.html#reference
type Ref struct {
	Node     Atom
	Creation uint32
	ID       [5]uint32
}

func (r Ref) String() string {
	n := uint32(0)
	if r.Node != "" {
		n = crc32.Checksum([]byte(r.Node), crc32q)
	}
	return fmt.Sprintf("Ref#<%08X.%d.%d.%d>", n, r.ID[0], r.ID[1], r.ID[2])
}

// Function type.
//
// Ref: https://www.erlang.org/doc/system/data_types.html#fun
type Function struct {
	Arity  byte
	Unique [16]byte
	Index  uint32
	//	Free      uint32
	Module    Atom
	OldIndex  uint32
	OldUnique uint32
	Pid       Pid
	FreeVars  []Term
}

// Export type.
type Export struct {
	Module   Atom
	Function Atom
	Arity    int
}
