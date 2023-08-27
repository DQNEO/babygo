package types2

type Basic struct {
	Knd  int
	name string
}

const GBool = 1
const GInt = 2
const GInt8 = 3
const GInt16 = 4
const GInt32 = 5
const GInt64 = 6
const GUint = 7
const GUint8 = 8
const GUint16 = 9
const GUint32 = 10
const GUint64 = 11
const GUintptr = 12
const GFloat32 = 13
const GFloat64 = 14
const GComplex64 = 15
const GComplex128 = 16
const GString = 17
const GUnsafePointer = 18

var Bool = &Basic{
	Knd:  GBool,
	name: "bool",
}

var Int = &Basic{
	Knd:  GInt,
	name: "int",
}

// Rune
var Int32 = &Basic{
	Knd:  GInt32,
	name: "int32",
}

var Uintptr = &Basic{
	Knd:  GUintptr,
	name: "uintptr",
}

var Uint8 = &Basic{
	Knd:  GUint8,
	name: "uint8",
}

var Uint16 = &Basic{
	Knd:  GUint16,
	name: "uint16",
}
var String = &Basic{
	Knd:  GString,
	name: "string",
}

// Kind returns the kind of basic type b.
func (b *Basic) Kind() int { return b.Knd }

// Name returns the Name of basic type b.
func (b *Basic) Name() string { return b.name }

func (t *Basic) Underlying() Type { return t }
func (b *Basic) String() string   { return b.name }
