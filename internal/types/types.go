package types

import (
	"github.com/DQNEO/babygo/internal/universe"
	"github.com/DQNEO/babygo/lib/ast"
)

type Type struct {
	E       ast.Expr // original
	PkgName string
	Name    string
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
}

var Int *Type = &Type{
	E: &ast.Ident{
		Name:    "int",
		Obj:     universe.Int,
		NamePos: 1,
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
