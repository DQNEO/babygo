package compiler

import (
	"os"
	"os/exec"

	"github.com/DQNEO/babygo/internal/codegen"
	"github.com/DQNEO/babygo/internal/ir"
	"github.com/DQNEO/babygo/internal/sema"
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/parser"
	"github.com/DQNEO/babygo/lib/path"
	"github.com/DQNEO/babygo/lib/token"
)

type PackageToCompile struct {
	Path     string
	Name     string
	Imports  []string
	GoFiles  []string
	AsmFiles []string
}

func CompileDecl(universe *ast.Scope, fset *token.FileSet, importPath string, declFilePath string) {
	f, err := os.Open(declFilePath)
	if err != nil {
		panic("cannot open decl file " + declFilePath)
	}
	f.Close()
	pkgName := path.Base(importPath)
	pkg := &ir.PkgContainer{Name: pkgName, Path: importPath, Fset: fset}
	pkg.FileNoMap = make(map[string]int)

	//	fmt.Fprintf(fout, "#=== Package %s\n", pkg.Path)

	pkgScope := ast.NewScope(universe)
	files := []string{declFilePath}
	for i, file := range files {
		fileno := i + 1
		pkg.FileNoMap[file] = fileno
		//		fmt.Fprintf(fout, "  .file %d \"%s\"\n", fileno, file) // For DWARF debug info
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

	sema.Walk(pkg)

}

// Compile compiles go files of a package into an assembly file, and copy input assembly files into it.
func Compile(universe *ast.Scope, fset *token.FileSet, pkgc *PackageToCompile, outAsmPath string, outObjPath string, declFilePath string) *ir.AnalyzedPackage {
	// Path string, pkgName string, gofiles []string, asmfiles []string,
	pkg := &ir.PkgContainer{Name: pkgc.Name, Path: pkgc.Path, Fset: fset}
	pkg.FileNoMap = make(map[string]int)
	fout, err := os.Create(outAsmPath)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(fout, "#=== Package %s\n", pkg.Path)

	pkgScope := ast.NewScope(universe)
	for i, file := range pkgc.GoFiles {
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

	apkg := sema.Walk(pkg)
	apkg.Imports = pkgc.Imports
	codegen.GenerateDecls(apkg, declFilePath)
	codegen.GenerateCode(apkg, fout)

	// append static asm files
	AppendAsmFiles(pkgc.AsmFiles, fout)

	// cleanup
	fout.Close()
	sema.Clear()

	// Do assembling
	out, err := exec.Command("/usr/bin/as", "-o", outObjPath, outAsmPath).CombinedOutput()
	if err != nil {
		panic(string(out))
	}
	return apkg
}

func AppendAsmFiles(asmFiles []string, fout *os.File) {
	for _, file := range asmFiles {
		fmt.Fprintf(fout, "# === static assembly %s ====\n", file)
		asmContents, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		fout.Write(asmContents)
	}
}

func resolveImports(file *ast.File) {
	mapImports := make(map[string]bool)
	for _, imprt := range file.Imports {
		// unwrap double quote "..."
		rawValue := imprt.Path.Value
		pth := rawValue[1 : len(rawValue)-1]
		base := path.Base(pth)
		mapImports[base] = true
	}
	for _, ident := range file.Unresolved {
		// lookup imported package name
		_, ok := mapImports[ident.Name]
		if ok {
			ident.Obj = &ast.Object{
				Kind: ast.Pkg,
				Name: ident.Name,
			}
		}
	}
}
