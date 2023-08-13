package reflect

import "unsafe"

type Type struct {
	typ *rtype
}

type iface struct {
	typ         *rtype         // dynamic type
	data        unsafe.Pointer // pointer to the actual data of the dynamic type
	methodName1 string
	methodRef1  unsafe.Pointer // a pointer to a method def e.g. os.$File.Write
	methodName2 string
	methodRef2  unsafe.Pointer
	// methodName3
	// methodRef3
	// ...
	// ...
	NullPointer uintptr // always 0
}

type rtype struct {
	id   int    // dtypeID
	name string // string representation of type
}

func TypeOf(x interface{}) *Type {
	eface := (*iface)(unsafe.Pointer(&x))
	return &Type{
		typ: eface.typ,
	}
}

func (t *Type) String() string {
	return t.typ.String()
}

func (t *rtype) String() string {
	return t.name
}
