package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"

	"golang.org/x/tools/go/loader"
)

func main() {
	handsOn()
}

const sourceCode = `package main

import "fmt"
import "os"
import "go/token"

var stdout = os.Stdout
type tfs token.FileSet 
func main() {
        fmt.Println("Hello, world")
}`

func main3() {
	var conf loader.Config
	conf.CreateFromFilenames("main", "../../t/test.go", "../../t/another.go")
	prog, err := conf.Load()
	if err != nil {
		panic(err)
	}
	fmt.Println("AllPackages:")
	for pkg, pi := range prog.AllPackages {
		if pkg.Name() == "main" {
			fmt.Printf("%s\n", pi)
			fmt.Printf("Path  %q\n", pkg.Path())
			fmt.Printf("Name:    %s\n", pkg.Name())
			fmt.Printf("Imports: %s\n", pkg.Imports())
			fmt.Printf("Scope:   %s\n", pkg.Scope())
			for _, name := range pkg.Scope().Names() {
				fmt.Printf("name:  %s\n", name)
			}
			break
		}
	}
}
func main2() {
	fset := token.NewFileSet()

	// Parse the input string, []byte, or io.Reader,
	// recording position information in fset.
	// ParseFile returns an *ast.File, a syntax tree.
	testFile, err := os.Open("./t/test.go")
	if err != nil {
		log.Fatal(err) // parse error
	}
	astFile, err := parser.ParseFile(fset, "test.go", testFile, 0)
	if err != nil {
		log.Println("ParseFile failed")
		log.Fatal(err) // parse error
	}
	log.Println("ParseFile succeeded")

	// A Config controls various options of the type checker.
	// The defaults work fine except for one setting:
	// we must specify how to deal with imports.
	conf := types.Config{Importer: importer.Default()}

	info := types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Instances:  nil,
		Defs:       nil,
		Uses:       nil,
		Implicits:  nil,
		Selections: nil,
		Scopes:     nil,
		InitOrder:  nil,
	}
	// Type-check the package containing only file astFile.
	// Check returns a *types.Package.
	pkg, err := conf.Check("cmd/hello", fset, []*ast.File{astFile}, &info)
	if err != nil {
		log.Fatal(err) // type error
	}

	fmt.Printf("Package  %q\n", pkg.Path())
	fmt.Printf("Name:    %s\n", pkg.Name())
	fmt.Printf("Imports: %s\n", pkg.Imports())
	fmt.Printf("Scope:   %s\n", pkg.Scope())
	for _, name := range pkg.Scope().Names() {
		fmt.Printf("name:  %s\n", name)
	}
	object := pkg.Scope().Lookup("stdout")
	fmt.Printf("object :  %s\n", object)
	object2 := pkg.Scope().Lookup("tfs")
	fmt.Printf("object2 :  %s\n", object2)
}

const code = `package main

import "fmt"

type MyInt int
func main() {
	const c = 10
	var i MyInt = c
	var ifc interface{} = i
	fmt.Println(i, ifc)
}

`

func handsOn() {
	fset := token.NewFileSet()

	conf := types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			fmt.Printf("ERROR: %#v\n", err)
		},
	}

	f, err := parser.ParseFile(fset, "code", code, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// Info.ObjectOfやInfo.TypeOfのspecに書いてあるPreconditionを読んで初期化が必要なフィールドを把握しておく
	info := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Defs:  map[*ast.Ident]types.Object{},
		// Uses:  map[*ast.Ident]types.Object{},
	}

	_, err = conf.Check("p", fset, []*ast.File{f}, info)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== ast.Inspect ...")
	ast.Inspect(f, func(node ast.Node) bool {
		if expr, ok := node.(ast.Expr); ok {
			typeAndValue, ok := info.Types[expr]
			if ok {
				typ := typeAndValue.Type
				val := typeAndValue.Value
				fmt.Printf("[%s] typeAndValue=%s\n", fset.Position(node.Pos()), typeAndValue)
				fmt.Printf("      type of expr=%T\n", expr)
				fmt.Printf("      type of type=%T  %s\n", typ, typ)
				fmt.Printf("      type of val=%T %s\n", val, val)
			}
		}
		switch e := node.(type) {
		case *ast.Ident:
			ident := e
			fmt.Printf("[%s] ident found %s\n", fset.Position(ident.Pos()), ident)
			obj := info.ObjectOf(ident)
			if obj == nil {
				fmt.Println("    not Object")
				return true
			}
			typ := obj.Type()
			fmt.Printf("    T of Type='%T' obj='%s', object.Type='%s'\n", typ, obj, typ)

			typeAndValue, ok := info.Types[ident]
			if ok {
				fmt.Printf("    typeAndValue %v\n", typeAndValue)
			}
		}
		return true
	})
}
