package types

import (
	"github.com/DQNEO/babygo/internal/universe"
	"github.com/DQNEO/babygo/lib/ast"
)

type Type struct {
	E       ast.Expr // original
	PkgName string
	Name    string
	GoType  GoType
}

type TypeKind string

const T_STRING TypeKind = "T_STRING"
const T_INTERFACE TypeKind = "T_INTERFACE"
const T_FUNC TypeKind = "T_FUNC"
const T_SLICE TypeKind = "T_SLICE"
const T_BOOL TypeKind = "T_BOOL"
const T_INT TypeKind = "T_INT"
const T_INT32 TypeKind = "T_INT32"
const T_UINT8 TypeKind = "T_UINT8"
const T_UINT16 TypeKind = "T_UINT16"
const T_UINTPTR TypeKind = "T_UINTPTR"
const T_ARRAY TypeKind = "T_ARRAY"
const T_STRUCT TypeKind = "T_STRUCT"
const T_POINTER TypeKind = "T_POINTER"
const T_MAP TypeKind = "T_MAP"

// --- universal types ---
var Bool *Type = &Type{
	E: &ast.Ident{
		Name:    "bool",
		Obj:     universe.Bool,
		NamePos: 1,
	},
	GoType: &Basic{
		Knd:  GBool,
		name: "bool",
	},
}

var Int *Type = &Type{
	E: &ast.Ident{
		Name:    "int",
		Obj:     universe.Int,
		NamePos: 1,
	},
	GoType: &Basic{
		Knd:  GInt,
		name: "int",
	},
}

// Rune
var Int32 *Type = &Type{
	E: &ast.Ident{
		Name:    "int32",
		Obj:     universe.Int32,
		NamePos: 1,
	},
}

var Uintptr *Type = &Type{
	E: &ast.Ident{
		Name:    "uintptr",
		Obj:     universe.Uintptr,
		NamePos: 1,
	},
}
var Uint8 *Type = &Type{
	E: &ast.Ident{
		Name:    "uint8",
		Obj:     universe.Uint8,
		NamePos: 1,
	},
}

var Uint16 *Type = &Type{
	E: &ast.Ident{
		Name:    "uint16",
		Obj:     universe.Uint16,
		NamePos: 1,
	},
}
var String *Type = &Type{
	E: &ast.Ident{
		Name:    "string",
		Obj:     universe.String,
		NamePos: 1,
	},
}

// Eface means the empty interface type
var Eface *Type = &Type{
	E: &ast.InterfaceType{
		Interface: 1,
	},
}

var GeneralSliceType *Type = &Type{}

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

// GoType emulates go/types.Type
type GoType interface {
	// Underlying returns the underlying type of a type.
	Underlying() GoType

	// String returns a string representation of a type.
	String() string
}

type Basic struct {
	Knd  int
	name string
}

// Kind returns the kind of basic type b.
func (b *Basic) Kind() int { return b.Knd }

// Name returns the name of basic type b.
func (b *Basic) Name() string { return b.name }

func (t *Basic) Underlying() GoType { return t }
func (b *Basic) String() string     { return b.name }
