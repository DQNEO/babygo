package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
	"strings"
)

func emitVariable(ident *ast.Object) {
	if ident.Kind != ast.Var {
		panic("ident should be ast.Var")
	}
	valSpec,ok := ident.Decl.(*ast.ValueSpec)
	if !ok {
		panic("Unexpected case")
	}

	fmt.Printf("  # emitVariable string ident.Decl.Type=%#v\n", valSpec.Type)
	if getPrimType(valSpec.Type) == gString {
		fmt.Printf("  movq %s+0(%%rip), %%rax\n", ident.Name)
		fmt.Printf("  pushq %%rax # ptr\n")
		fmt.Printf("  mov %s+8(%%rip), %%rax\n", ident.Name)
		fmt.Printf("  pushq %%rax # len\n")
	} else if getPrimType(valSpec.Type) == gInt {
		fmt.Printf("  movq %s+0(%%rip), %%rax\n", ident.Name)
		fmt.Printf("  pushq %%rax # ptr\n")
	} else {
		panic("Unexpected type")
	}
}

func emitExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Ident:
		fmt.Printf("# ident kind=%v\n", e.Obj.Kind)
		fmt.Printf("# Obj=%v\n", e.Obj)
		if e.Obj.Kind == ast.Var {
			emitVariable(e.Obj)
		} else {
			panic("Unexpected ident kind")
		}
	case *ast.CallExpr:
		fun := e.Fun
		fmt.Printf("  # funcall=%T\n", fun)
		switch fn := fun.(type) {
		case *ast.Ident:
			if fn.Name == "print" {
				// builtin print
				emitExpr(e.Args[0]) // push ptr, push len
				symbol := fmt.Sprintf("runtime.printstring")
				fmt.Printf("  callq %s\n", symbol)
			} else {
				panic("Unexpected fn.Name:" + fn.Name)
			}
		case *ast.SelectorExpr:
			emitExpr(e.Args[0])
			symbol := fmt.Sprintf("%s.%s", fn.X, fn.Sel)
			fmt.Printf("  callq %s\n", symbol)
		default:
			panic(fmt.Sprintf("Unexpected expr type %T", fun))
		}
	case *ast.ParenExpr:
		emitExpr(e.X)
	case *ast.BasicLit:
		fmt.Printf("  # start %T\n", e)
		fmt.Printf("  # kind=%s\n", e.Kind)
		if e.Kind.String() == "INT" {
			val := e.Value
			ival, _ := strconv.Atoi(val)
			fmt.Printf("  movq $%d, %%rax\n", ival)
			fmt.Printf("  pushq %%rax\n")
		} else if e.Kind.String() == "STRING" {
			// e.Value == ".S%d:%d"
			splitted := strings.Split(e.Value, ":")
			fmt.Printf("  leaq %s, %%rax\n", splitted[0]) // str.ptr
			fmt.Printf("  pushq %%rax\n")
			fmt.Printf("  pushq $%s\n", splitted[1]) // str.len
		} else {
			panic("Unexpected literal kind:" + e.Kind.String())
		}
		fmt.Printf("  # end %T\n", e)
	case *ast.BinaryExpr:
		fmt.Printf("  # start %T\n", e)
		emitExpr(e.X) // left
		emitExpr(e.Y) // right
		if e.Op.String() == "+" {
			fmt.Printf("  popq %%rdi # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  addq %%rdi, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		} else if e.Op.String() == "-" {
			fmt.Printf("  popq %%rdi # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  subq %%rdi, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		} else if e.Op.String() == "*" {
			fmt.Printf("  popq %%rdi # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  imulq %%rdi, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		} else {
			panic(fmt.Sprintf("Unexpected binary operator %s", e.Op))
		}
		fmt.Printf("  # end %T\n", e)
	default:
		panic(fmt.Sprintf("Unexpected expr type %T", expr))
	}
}

func emitFuncDecl(pkgPrefix string, funcDecl *ast.FuncDecl) {
	fmt.Printf("%s.%s:\n", pkgPrefix, funcDecl.Name)
	fmt.Printf("push %%rbp\n")
	fmt.Printf("movq %%rsp, %%rbp\n")
	for _, stmt := range funcDecl.Body.List {
		switch stmt.(type) {
		case *ast.ExprStmt:
			expr := stmt.(*ast.ExprStmt).X
			emitExpr(expr)
		default:
			panic("Unexpected stmt type")
		}
	}
	fmt.Printf("  leave\n")
	fmt.Printf("  ret\n")
}

var stringLiterals []string
var stringIndex int

func registerStringLiteral(s string) string {
	rawStringLiteal := s
	stringLiterals = append(stringLiterals, rawStringLiteal)
	r := fmt.Sprintf(".S%d:%d", stringIndex, len(rawStringLiteal) - 2 -1) // \n is counted as 2 ?
	stringIndex++
	return r
}

func walkExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Ident:
		// what to do ?
	case *ast.CallExpr:
		for _, arg := range e.Args {
			walkExpr(arg)
		}
	case *ast.ParenExpr:
		walkExpr(e.X)
	case *ast.BasicLit:
		if e.Kind.String() == "INT" {
		} else if e.Kind.String() == "STRING" {
			e.Value = registerStringLiteral(e.Value)
		} else {
			panic("Unexpected literal kind:" + e.Kind.String())
		}
	case *ast.BinaryExpr:
		walkExpr(e.X) // left
		walkExpr(e.Y) // right
	default:
		panic(fmt.Sprintf("Unexpected expr type %T", expr))
	}
}

var gString = &ast.Object{
	Kind: ast.Typ,
	Name: "string",
	Decl: nil,
	Data: nil,
	Type: nil,
}

var gInt = &ast.Object{
	Kind: ast.Typ,
	Name: "int",
	Decl: nil,
	Data: nil,
	Type: nil,
}

var globalVars []*ast.ValueSpec
var globalFuncs []*ast.FuncDecl

func semanticAnalyze(fset *token.FileSet, f *ast.File) {
	// https://github.com/golang/example/tree/master/gotypes#an-example
	// Type check
	// A Config controls various options of the type checker.
	// The defaults work fine except for one setting:
	// we must specify how to deal with imports.
	conf := types.Config{Importer: importer.Default()}

	// Type-check the package containing only file f.
	// Check returns a *types.Package.
	pkg, err := conf.Check("./t", fset, []*ast.File{f}, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("# Package  %q\n", pkg.Path())
	universe := &ast.Scope{
		Outer:   nil,
		Objects: make(map[string]*ast.Object),
	}

	universe.Insert(gString)
	universe.Insert(gInt)
	universe.Insert(&ast.Object{
		Kind: ast.Fun,
		Name: "print",
		Decl: nil,
		Data: nil,
		Type: nil,
	})

	universe.Insert(&ast.Object{
		Kind: ast.Pkg,
		Name: "os", // why ???
		Decl: nil,
		Data: nil,
		Type: nil,
	})
	//fmt.Printf("Universer:    %v\n", types.Universe)
	ap, _ := ast.NewPackage(fset, map[string]*ast.File{"":f}, nil, universe)

	var unresolved []*ast.Ident
	for _, ident := range f.Unresolved {
		if obj := universe.Lookup(ident.Name); obj != nil {
			ident.Obj = obj
		} else {
			unresolved = append(unresolved, ident)
		}
	}

	fmt.Printf("# Name:    %s\n", pkg.Name())
	fmt.Printf("# Unresolved: %v\n", unresolved)
	fmt.Printf("# Package:   %s\n", ap.Name)


	for _, decl := range f.Decls {
		switch dcl := decl.(type) {
		case *ast.GenDecl:
			switch dcl.Tok {
			case token.VAR:
				spec := dcl.Specs[0]
				valSpec := spec.(*ast.ValueSpec)
				fmt.Printf("# valSpec.type=%#v\n", valSpec.Type)
				typeIdent , ok := valSpec.Type.(*ast.Ident)
				if !ok {
					panic("Unexpected case")
				}
				if typeIdent.Obj == gString {
					fmt.Printf("# spec.Name=%v, Value=%v\n", valSpec.Names[0], valSpec.Values[0])
					lit,ok := valSpec.Values[0].(*ast.BasicLit)
					if !ok {
						panic("Unexpected type")
					}
					lit.Value = registerStringLiteral(lit.Value)
				} else if typeIdent.Obj == gInt {
					_,ok := valSpec.Values[0].(*ast.BasicLit)
					if !ok {
						panic("Unexpected type")
					}
				} else {
					panic("Unexpected global ident")
				}
				globalVars = append(globalVars, valSpec)
			}
		case *ast.FuncDecl:
			funcDecl := decl.(*ast.FuncDecl)
			for _, stmt := range funcDecl.Body.List {
				switch stmt.(type) {
				case *ast.ExprStmt:
					expr := stmt.(*ast.ExprStmt).X
					walkExpr(expr)
				default:
					panic("Unexpected stmt type")
				}
			}
			globalFuncs = append(globalFuncs, funcDecl)
		default:
			panic("unexpected decl type")
		}
	}
}


func getPrimType(expr ast.Expr) *ast.Object {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Obj
	default:
		panic("Unexpected expr type")
	}
}

func emitData() {
	fmt.Printf(".data\n")
	for i, sl := range stringLiterals {
		fmt.Printf("# string literals\n")
		fmt.Printf(".S%d:\n", i)
		fmt.Printf("  .string %s\n", sl)
	}

	fmt.Printf("# Global Variables\n")
	for _, valSpec := range globalVars {
		name := valSpec.Names[0]
		val := valSpec.Values[0]
		var strval string
		switch vl := val.(type) {
		case *ast.BasicLit:
			if getPrimType(valSpec.Type) == gString {
				strval = vl.Value
				splitted := strings.Split(strval,":")
				fmt.Printf("%s:  #\n", name)
				fmt.Printf("  .quad %s\n", splitted[0])
				fmt.Printf("  .quad %s\n", splitted[1])
			} else if getPrimType(valSpec.Type) == gInt {
				fmt.Printf("%s:  #\n", name)
				fmt.Printf("  .quad %s\n", vl.Value)
			} else {
				panic("Unexpected case")
			}

		default:
			panic("Only BasicLit is allowed in Global var's value")
		}
	}
}

func emitText() {
	fmt.Printf(".text\n")
	for _, funcDecl := range globalFuncs {
		fmt.Printf("# funcDecl %s\n", funcDecl.Name)
		emitFuncDecl("main", funcDecl)
	}
}

func generateCode(f *ast.File) {
	emitData()
	emitText()
}

func main() {
	fset := &token.FileSet{}
	f, err := parser.ParseFile(fset, "./t/source.go", nil, 0)
	if err != nil {
		panic(err)
	}

	semanticAnalyze(fset, f)
	generateCode(f)
}
