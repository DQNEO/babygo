package main

import (
	"os"

	"github.com/DQNEO/babygo/internal/codegen"
	"github.com/DQNEO/babygo/internal/ir"
	"github.com/DQNEO/babygo/internal/sema"
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/parser"
	"github.com/DQNEO/babygo/lib/token"
)

func parseFile(fset *token.FileSet, filename string) *ast.File {
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		panic(err.Error())
	}
	return f
}

// compile compiles go files of a package into an assembly file, and copy input assembly files into it.
func compile(universe *ast.Scope, fset *token.FileSet, pkgPath string, name string, gofiles []string, asmfiles []string, outFilePath string) *ir.PkgContainer {
	_pkg := &ir.PkgContainer{Name: name, Path: pkgPath, Fset: fset}
	_pkg.FileNoMap = make(map[string]int)
	outAsmFile, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	fout = outAsmFile
	codegen.Fout = fout
	printf("#=== Package %s\n", _pkg.Path)

	codegen.TypesMap = make(map[string]*codegen.DtypeEntry)
	codegen.TypeId = 1

	pkgScope := ast.NewScope(universe)
	for i, file := range gofiles {
		fileno := i + 1
		_pkg.FileNoMap[file] = fileno
		printf("  .file %d \"%s\"\n", fileno, file)

		astFile := parseFile(fset, file)
		//		logf("[main]package decl lineno = %s\n", fset.Position(astFile.Package))
		_pkg.Name = astFile.Name.Name
		_pkg.AstFiles = append(_pkg.AstFiles, astFile)
		for name, obj := range astFile.Scope.Objects {
			pkgScope.Objects[name] = obj
		}
	}
	for _, astFile := range _pkg.AstFiles {
		resolveImports(astFile)
		var unresolved []*ast.Ident
		for _, ident := range astFile.Unresolved {
			obj := pkgScope.Lookup(ident.Name)
			if obj != nil {
				ident.Obj = obj
			} else {
				obj := universe.Lookup(ident.Name)
				if obj != nil {

					ident.Obj = obj
				} else {

					// we should allow unresolved for now.
					// e.g foo in X{foo:bar,}
					unresolved = append(unresolved, ident)
				}
			}
		}
		for _, dcl := range astFile.Decls {
			_pkg.Decls = append(_pkg.Decls, dcl)
		}
	}

	printf("#--- walk \n")
	sema.Walk(_pkg)
	codegen.GenerateCode(_pkg)

	// append static asm files
	for _, file := range asmfiles {
		fmt.Fprintf(outAsmFile, "# === static assembly %s ====\n", file)
		asmContents, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		outAsmFile.Write(asmContents)
	}

	outAsmFile.Close()
	fout = nil
	codegen.Fout = nil
	sema.CurrentPkg = nil
	return _pkg
}
