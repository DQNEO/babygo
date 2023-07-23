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

// compile compiles go files of a package into an assembly file, and copy input assembly files into it.
func compile(universe *ast.Scope, fset *token.FileSet, pkgPath string, pkgName string, gofiles []string, asmfiles []string, outFilePath string) *ir.AnalyzedPackage {
	pkg := &ir.PkgContainer{Name: pkgName, Path: pkgPath, Fset: fset}
	pkg.FileNoMap = make(map[string]int)
	fout, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(fout, "#=== Package %s\n", pkg.Path)

	pkgScope := ast.NewScope(universe)
	for i, file := range gofiles {
		fileno := i + 1
		pkg.FileNoMap[file] = fileno
		fmt.Fprintf(fout, "  .file %d \"%s\"\n", fileno, file) // For DWARF debug info
		astFile, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			panic(err.Error())
		}

		pkg.AstFiles = append(pkg.AstFiles, astFile)
		for name, obj := range astFile.Scope.Objects {
			pkgScope.Objects[name] = obj
		}
	}
	for _, astFile := range pkg.AstFiles {
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
					// we should allow unresolved in this stage.
					// e.g foo in X{foo:bar,}
					unresolved = append(unresolved, ident)
				}
			}
		}
		for _, dcl := range astFile.Decls {
			pkg.Decls = append(pkg.Decls, dcl)
		}
	}

	fmt.Fprintf(fout, "#--- walk \n")
	apkg := sema.Walk(pkg)
	codegen.GenerateCode(apkg, fout)

	// append static asm files
	for _, file := range asmfiles {
		fmt.Fprintf(fout, "# === static assembly %s ====\n", file)
		asmContents, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		fout.Write(asmContents)
	}

	// cleanup
	fout.Close()
	sema.CurrentPkg = nil
	return apkg
}
