package types

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
var Bool Type = &Basic{
	Knd:  GBool,
	name: "bool",
}

var Int Type = &Basic{
	Knd:  GInt,
	name: "int",
}

// Rune
var Int32 Type = &Basic{
	Knd:  GInt32,
	name: "int32",
}

var Uintptr Type = &Basic{
	Knd:  GUintptr,
	name: "uintptr",
}

var Uint8 Type = &Basic{
	Knd:  GUint8,
	name: "uint8",
}

var Uint16 Type = &Basic{
	Knd:  GUint16,
	name: "uint16",
}
var String Type = &Basic{
	Knd:  GString,
	name: "string",
}

var EmptyInterface Type = &Interface{}
var GeneralSliceType Type = &Slice{}

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

// Type emulates go/types.Type
type Type interface {
	// Underlying returns the underlying type of a type.
	Underlying() Type

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

func (t *Basic) Underlying() Type { return t }
func (b *Basic) String() string   { return b.name }

// An Array represents an array type.
type Array struct {
	len  int
	elem Type
}

// NewArray returns a new array type for the given element type and length.
// A negative length indicates an unknown length.
func NewArray(elem Type, len int) *Array { return &Array{len: len, elem: elem} }

// Len returns the length of array a.
// A negative result indicates an unknown length.
func (a *Array) Len() int { return a.len }

// Elem returns element type of array a.
func (a *Array) Elem() Type { return a.elem }

func (t *Array) Underlying() Type { return t }
func (t *Array) String() string   { return "@TBI" }

type Slice struct {
	elem   Type
	IsElps bool
}

// NewSlice returns a new slice type for the given element type.
func NewSlice(elem Type) *Slice { return &Slice{elem: elem} }

// Elem returns the element Type of slice s.
func (s *Slice) Elem() Type { return s.elem }

func (t *Slice) Underlying() Type { return t }
func (t *Slice) String() string   { return "@TBI" }

// A Pointer represents a pointer type.
type Pointer struct {
	base Type // element Type
}

// NewPointer returns a new pointer Type for the given element (base) Type.
func NewPointer(elem Type) *Pointer { return &Pointer{base: elem} }

// Elem returns the element Type for the given pointer p.
func (p *Pointer) Elem() Type { return p.base }

func (t *Pointer) Underlying() Type { return t }
func (t *Pointer) String() string   { return "@TBI" }

// A Map represents a map type.
type Map struct {
	key  Type
	elem Type
}

// NewMap returns a new map for the given key and element types.
func NewMap(key Type, elem Type) *Map {
	return &Map{key: key, elem: elem}
}

// Key returns the key type of map m.
func (m *Map) Key() Type { return m.key }

// Elem returns the element type of map m.
func (m *Map) Elem() Type { return m.elem }

func (t *Map) Underlying() Type { return t }
func (t *Map) String() string   { return "@TBI" }

type Interface struct {
	Methods []*Func
}

func NewInterfaceType(methods []*Func) *Interface {
	i := &Interface{
		Methods: methods,
	}
	return i
}

func (t *Interface) Underlying() Type { return t }
func (t *Interface) String() string   { return "@TBI" }

type Func struct {
	Typ  *Signature
	Name string
}

func NewFunc(sig *Signature) *Func {
	return &Func{
		Typ: sig,
	}
}

func (t *Func) Underlying() Type { return t.Typ }
func (t *Func) String() string   { return t.Typ.String() }

type Tuple struct {
	Types []Type
}

func (t *Tuple) Underlying() Type { return t }
func (t *Tuple) String() string   { return "@TBI" }

type Signature struct {
	Params  *Tuple
	Results *Tuple
}

func (t *Signature) Underlying() Type { return t }
func (t *Signature) String() string   { return "@TBI" }

type Var struct {
	Name   string
	Type   Type
	Offset int
}

type Struct struct {
	Fields []*Var // Fields != nil indicates the struct is set up (possibly with len(Fields) == 0)
	//AstFields    []*ast.Field // @TODO: Replace this by Fields
	IsCalculated bool // the offsets are calculated ?
}

func NewStruct(fields []*Var) *Struct {
	return &Struct{
		Fields: fields,
	}
}

func (t *Struct) Underlying() Type { return t }
func (t *Struct) String() string   { return "@TBI" }

type Named struct {
	name    string
	UT      Type
	PkgName string
}

func NewNamed(name string, typ Type) *Named {
	return &Named{
		name: name,
		UT:   typ,
	}
}

func (t *Named) Underlying() Type {
	if t.UT == nil {
		panic("Named type " + t.PkgName + "." + t.name + ": Underlying is nil")
	}
	return t.UT
}
func (t *Named) String() string { return t.name }
