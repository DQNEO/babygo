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

const sourceCode = `package main

import "fmt"
import "os"
import "go/token"

var stdout = os.Stdout
type tfs token.FileSet 
func main() {
        fmt.Println("Hello, world")
}`

func main() {
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

func example() {
	//parser.ParseFile()
	config := types.Config{}
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
	pkg, err := config.Check("github.com/DQNEO/babygo/src/unsafe", nil, nil, &info)
	if err != nil {
		panic(err)
	}
	fmt.Printf("pkg=%#v\n", pkg)
	fmt.Printf("pkg.scope=%#v\n", pkg.Scope())
	fmt.Printf("info=%#v\n", info)
}
