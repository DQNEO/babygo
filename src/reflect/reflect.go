package reflect

import "unsafe"

type Type struct {
	X interface{}
}

func TypeOf(x interface{}) *Type {
	return &Type{
		X: x,
	}
}

type eface struct {
	_type *_type
	data  unsafe.Pointer
}

type _type struct {
	id   int
	name string
}

func (t *Type) String() string {
	switch t.X.(type) {
	//case int:
	//	return "int"
	//case string:
	//	return "string"
	default:
		ifc := t.X
		var ifcAddr uintptr = uintptr(unsafe.Pointer(&ifc))
		var pEface *eface = (*eface)(unsafe.Pointer(ifcAddr))
		typ := pEface._type
		return typ.name
	}
}
