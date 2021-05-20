package reflect

import "unsafe"

type Type struct {
	typ *rtype
}

type emptyInterface struct {
	typ  *rtype // dynamic type
	data unsafe.Pointer // pointer to the actual data of the dynamic type
}

type rtype struct {
	id   int
	name string // string representation of type
}

func TypeOf(x interface{}) *Type {
	eface := (*emptyInterface)(unsafe.Pointer(&x))
	return &Type{
		typ: eface.typ,
	}
}

func (t *Type) String() string {
	return t.typ.name
}
