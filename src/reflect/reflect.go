package reflect

import "unsafe"

type ReflectType struct {
	X interface{}
}

func TypeOf(x interface{}) *ReflectType {
	return &ReflectType{
		X:x,
	}
}

type eface struct {
	_type *_type
	data  unsafe.Pointer
}

type _type struct {
	id int
	name string
}

func (t *ReflectType) String() string {
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
		//return "Unknown"
	}
}

