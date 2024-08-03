package goetf

// SizeType refers to the length in bytes associated with an fixed external term type.
type SizeType = int

// ETF fixed type sizes.
const (
	SizeNil SizeType = 0

	SizeSmallInteger    SizeType = 1
	SizeAtomCacheRef    SizeType = 1
	SizeSmallTupleArity SizeType = 1
	SizeSmallBigN       SizeType = 1
	SizeSmallBigSign    SizeType = 1
	SizeLargeBigSign    SizeType = 1
	SizeBitBinaryBits   SizeType = 1
	SizeSmallAtom       SizeType = 1
	SizeSmallAtomUTF8   SizeType = 1

	SizeAtom         SizeType = 2
	SizeAtomUTF8     SizeType = 2
	SizeStringLength SizeType = 2

	SizeLargeBigN       SizeType = 4
	SizeInteger         SizeType = 4
	SizeMapArity        SizeType = 4
	SizeBinaryLen       SizeType = 4
	SizeListLength      SizeType = 4
	SizeLargeTupleArity SizeType = 4
	SizeBitBinaryLen    SizeType = 4

	SizeNewFloat SizeType = 8

	SizeFloat SizeType = 31
)
