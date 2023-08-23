package universe

import (
	"github.com/DQNEO/babygo/lib/ast"
)

var Nil = &ast.Object{
	Kind: ast.Con, // is nil a constant ?
	Name: "nil",
}

var True = &ast.Object{
	Kind: ast.Con,
	Name: "true",
}
var False = &ast.Object{
	Kind: ast.Con,
	Name: "false",
}

var String = &ast.Object{
	Kind: ast.Typ,
	Name: "string",
}

var Uintptr = &ast.Object{
	Kind: ast.Typ,
	Name: "uintptr",
}
var Bool = &ast.Object{
	Kind: ast.Typ,
	Name: "bool",
}
var Int = &ast.Object{
	Kind: ast.Typ,
	Name: "int",
}

var Int32 = &ast.Object{
	Kind: ast.Typ,
	Name: "int32",
}

var Uint8 = &ast.Object{
	Kind: ast.Typ,
	Name: "uint8",
}

var Uint16 = &ast.Object{
	Kind: ast.Typ,
	Name: "uint16",
}

var Error = &ast.Object{
	Kind: ast.Typ,
	Name: "error",
}

var New = &ast.Object{
	Kind: ast.Fun,
	Name: "new",
}

var Make = &ast.Object{
	Kind: ast.Fun,
	Name: "make",
}
var Append = &ast.Object{
	Kind: ast.Fun,
	Name: "append",
}

var Len = &ast.Object{
	Kind: ast.Fun,
	Name: "len",
}

var Cap = &ast.Object{
	Kind: ast.Fun,
	Name: "cap",
}
var Panic = &ast.Object{
	Kind: ast.Fun,
	Name: "panic",
}
var Delete = &ast.Object{
	Kind: ast.Fun,
	Name: "delete",
}

func CreateUniverse() *ast.Scope {
	universe := ast.NewScope(nil)
	objects := []*ast.Object{
		Nil,
		// constants
		True, False,
		// types
		String, Uintptr, Bool, Int, Uint8, Uint16, Int32, Error,
		// funcs
		New, Make, Append, Len, Cap, Panic, Delete,
	}

	for _, obj := range objects {
		universe.Insert(obj)
	}

	// setting aliases
	universe.Objects["byte"] = Uint8

	return universe
}

// T exists for For pointer address test
type T struct {
	A int
}

// X exists for For pointer address test
var X = &T{}
