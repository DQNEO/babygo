package types

import (
	"github.com/DQNEO/babygo/lib/ast"
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

// --- universal types ---
var Bool GoType = &Basic{
	Knd:  GBool,
	name: "bool",
}

var Int GoType = &Basic{
	Knd:  GInt,
	name: "int",
}

// Rune
var Int32 GoType = &Basic{
	Knd:  GInt32,
	name: "int32",
}

var Uintptr GoType = &Basic{
	Knd:  GUintptr,
	name: "uintptr",
}

var Uint8 GoType = &Basic{
	Knd:  GUint8,
	name: "uint8",
}

var Uint16 GoType = &Basic{
	Knd:  GUint16,
	name: "uint16",
}
var String GoType = &Basic{
	Knd:  GString,
	name: "string",
}

var EmptyInterface GoType = &Interface{}
var GGeneralSliceType GoType = &Slice{}

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

// Name returns the Name of basic type b.
func (b *Basic) Name() string { return b.name }

func (t *Basic) Underlying() GoType { return t }
func (b *Basic) String() string     { return b.name }

// An Array represents an array type.
type Array struct {
	len  int
	elem GoType
}

// NewArray returns a new array type for the given element type and length.
// A negative length indicates an unknown length.
func NewArray(elem GoType, len int) *Array { return &Array{len: len, elem: elem} }

// Len returns the length of array a.
// A negative result indicates an unknown length.
func (a *Array) Len() int { return a.len }

// Elem returns element type of array a.
func (a *Array) Elem() GoType { return a.elem }

func (t *Array) Underlying() GoType { return t }
func (t *Array) String() string     { return "@TBI" }

type Slice struct {
	elem GoType
	Elp  bool
}

// NewSlice returns a new slice type for the given element type.
func NewSlice(elem GoType) *Slice { return &Slice{elem: elem} }

// Elem returns the element GoType of slice s.
func (s *Slice) Elem() GoType { return s.elem }

func (t *Slice) Underlying() GoType { return t }
func (t *Slice) String() string     { return "@TBI" }

// A Pointer represents a pointer type.
type Pointer struct {
	base GoType // element GoType
}

// NewPointer returns a new pointer GoType for the given element (base) GoType.
func NewPointer(elem GoType) *Pointer { return &Pointer{base: elem} }

// Elem returns the element GoType for the given pointer p.
func (p *Pointer) Elem() GoType { return p.base }

func (t *Pointer) Underlying() GoType { return t }
func (t *Pointer) String() string     { return "@TBI" }

// A Map represents a map type.
type Map struct {
	key  GoType
	elem GoType
}

// NewMap returns a new map for the given key and element types.
func NewMap(key GoType, elem GoType) *Map {
	return &Map{key: key, elem: elem}
}

// Key returns the key type of map m.
func (m *Map) Key() GoType { return m.key }

// Elem returns the element type of map m.
func (m *Map) Elem() GoType { return m.elem }

func (t *Map) Underlying() GoType { return t }
func (t *Map) String() string     { return "@TBI" }

type Interface struct {
	Methods []*Func
	//EMethods *ast.FieldList
}

func NewInterfaceType(methods []*Func) *Interface {
	i := &Interface{
		Methods: methods,
	}
	return i
}

func (t *Interface) Underlying() GoType { return t }
func (t *Interface) String() string     { return "@TBI" }

type Func struct {
	Typ  GoType
	Name string
}

func NewFunc(sig *Signature) *Func {
	return &Func{
		Typ: sig,
	}
}

func (t *Func) Underlying() GoType { return t.Typ }
func (t *Func) String() string     { return t.Typ.String() }

type Tuple struct {
	Types []GoType
}

func (t *Tuple) Underlying() GoType { return t }
func (t *Tuple) String() string     { return "@TBI" }

type Signature struct {
	Params  *Tuple
	Results *Tuple
}

func (t *Signature) Underlying() GoType { return t }
func (t *Signature) String() string     { return "@TBI" }

type Var struct {
	Name string
	Typ  GoType
}

type Struct struct {
	Fields    []*Var // Fields != nil indicates the struct is set up (possibly with len(Fields) == 0)
	AstFields []*ast.Field
}

func NewStruct(fields []*Var, astFields []*ast.Field) *Struct {
	return &Struct{
		Fields:    fields,
		AstFields: astFields,
	}
}

func (t *Struct) Underlying() GoType { return t }
func (t *Struct) String() string     { return "@TBI" }

type Named struct {
	name        string
	Uunderlying GoType
	PkgName     string
}

func NewNamed(name string, typ GoType) *Named {
	return &Named{
		name:        name,
		Uunderlying: typ,
	}
}

func (t *Named) Underlying() GoType { return t.Uunderlying }
func (t *Named) String() string     { return t.name }
