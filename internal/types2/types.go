package types2

import (
	"github.com/DQNEO/babygo/lib/fmt"
)

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

// Type emulates go/types.Type
type Type interface {
	// Underlying returns the underlying type of a type.
	Underlying() Type

	// String returns a string representation of a type.
	String() string
}

func Kind(typ Type) TypeKind {
	if typ == nil {
		panic(fmt.Sprintf("[Kind] Unexpected nil:\n"))
	}

	switch gt := typ.(type) {
	case *Basic:
		switch gt.Kind() {
		case GBool:
			return T_BOOL
		case GInt:
			return T_INT
		case GInt32:
			return T_INT32
		case GUint8:
			return T_UINT8
		case GUint16:
			return T_UINT16
		case GUintptr:
			return T_UINTPTR
		case GString:
			return T_STRING
		default:
			panic("Unknown kind")
		}
	case *Array:
		return T_ARRAY
	case *Slice:
		return T_SLICE
	case *Struct:
		return T_STRUCT
	case *Pointer:
		return T_POINTER
	case *Map:
		return T_MAP
	case *Interface:
		return T_INTERFACE
	case *Func:
		return T_FUNC
	case *Signature:
		return T_FUNC
	case *Tuple:
		panic(fmt.Sprintf("Tuple is not expected: type %T\n", typ))
	case *Named:
		ut := gt.Underlying()
		if ut == nil {
			panic(fmt.Sprintf("no underlying type for NamedType %s\n", gt.String()))
		}
		return Kind(ut)
	default:
		panic(fmt.Sprintf("[Kind] Unexpected type: %T\n", typ))
	}
	return "UNKNOWN_KIND"
}
