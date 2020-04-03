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

func isGlobalVar(obj *ast.Object) bool {
	return getObjData(obj) == -1
}

func setObjData(obj *ast.Object, i int) {
	obj.Data = i
}

func getObjData(obj *ast.Object) int {
	objData, ok := obj.Data.(int)
	if !ok {
		throw(obj.Data)
	}
	return objData
}

func emitVariable(obj *ast.Object) {
	// precondition
	if obj.Kind != ast.Var {
		panic("obj should be ast.Var")
	}
	fmt.Printf("  # obj=%#v\n", obj)

	var typ ast.Expr
	var localOffset int
	switch dcl := obj.Decl.(type) {
	case *ast.ValueSpec:
		typ = dcl.Type
		localOffset = getObjData(obj)
	case *ast.Field:
		typ = dcl.Type
		localOffset = getObjData(obj) // param offset
		fmt.Printf("  # dcl.Names[0].Obj=%#v\n", dcl.Names[0].Obj)
		fmt.Printf("  # param offset=%d\n", localOffset)
	default:
		throw(dcl)
	}

	var scope_comment string
	if isGlobalVar(obj) {
		scope_comment = "global"
	} else {
		scope_comment = "local"
	}
	fmt.Printf("  # eval %s variable \"%s\" T=%T Data=%d\n", scope_comment, obj.Name, typ, obj.Data)

	var addr string
	if isGlobalVar(obj) {
		addr = fmt.Sprintf("%s(%%rip)", obj.Name)
	} else {
		addr = fmt.Sprintf("%d(%%rbp)", localOffset)
	}
	switch getTypeKind(typ) {
	case T_SLICE:
		fmt.Printf("  leaq %s, %%rdx # slice variable\n", addr)
		fmt.Printf("  movq %d(%%rdx), %%rax\n", 0)
		fmt.Printf("  movq %d(%%rdx), %%rcx\n", 8)
		fmt.Printf("  movq %d(%%rdx), %%rdx\n", 16)
		fmt.Printf("  pushq %%rdx # cap\n")
		fmt.Printf("  pushq %%rcx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case T_STRING:
		fmt.Printf("  leaq %s, %%rcx # string variable\n", addr)
		fmt.Printf("  movq %d(%%rcx), %%rax\n", 0)
		fmt.Printf("  movq %d(%%rcx), %%rcx\n", 8)
		fmt.Printf("  pushq %%rcx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case T_INT:
		fmt.Printf("  leaq %s, %%rax # int variable\n", addr)
		fmt.Printf("  movq %d(%%rax), %%rax\n", 0)
		fmt.Printf("  pushq %%rax # int value\n")
	default:
		throw(typ)
	}
}

func emitVariableAddr(obj *ast.Object) {
	// precondition
	if obj.Kind != ast.Var {
		panic("obj should be ast.Var")
	}
	decl,ok := obj.Decl.(*ast.ValueSpec)
	if !ok {
		panic("Unexpected case")
	}

	fmt.Printf("  # emitVariable\n")
	fmt.Printf("  # obj.Data=%d\n", obj.Data)

	switch getTypeKind(decl.Type) {
	case T_SLICE:
		if isGlobalVar(obj) {
			fmt.Printf("  # global\n")
			fmt.Printf("  leaq %s+%d(%%rip), %%rax # ptr\n", obj.Name, 0)
			fmt.Printf("  leaq %s+%d(%%rip), %%rcx # len\n", obj.Name, 8)
			fmt.Printf("  leaq %s+%d(%%rip), %%rdx # cap\n", obj.Name, 16)
		} else {
			fmt.Printf("  # local\n")
			localOffset := (getObjData(obj))
			fmt.Printf("  leaq %d(%%rbp), %%rax # ptr %s \n", localOffset, obj.Name)
			fmt.Printf("  leaq %d(%%rbp), %%rcx # len %s \n", localOffset + 8, obj.Name)
			fmt.Printf("  leaq %d(%%rbp), %%rdx # cap %s \n", localOffset + 16, obj.Name)
		}
		fmt.Printf("  pushq %%rdx # cap\n")
		fmt.Printf("  pushq %%rcx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case T_STRING:
		if isGlobalVar(obj) {
			fmt.Printf("  # global\n")
			fmt.Printf("  leaq %s+%d(%%rip), %%rax # ptr\n", obj.Name, 0)
			fmt.Printf("  leaq %s+%d(%%rip), %%rcx # len\n", obj.Name, 8)
		} else {
			fmt.Printf("  # local\n")
			localOffset := (getObjData(obj))
			fmt.Printf("  leaq %d(%%rbp), %%rax # ptr %s \n", localOffset, obj.Name)
			fmt.Printf("  leaq %d(%%rbp), %%rcx # len %s \n", localOffset + 8, obj.Name)
		}
		fmt.Printf("  pushq %%rcx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case T_INT:
		if isGlobalVar(obj) {
			fmt.Printf("  # global\n")
			fmt.Printf("  leaq %s+0(%%rip), %%rax\n", obj.Name)
		} else {
			fmt.Printf("  # local\n")
			localOffset := (getObjData(obj))
			fmt.Printf("  leaq %d(%%rbp), %%rax # %s \n", localOffset, obj.Name)
		}
		fmt.Printf("  pushq %%rax # int var\n")
	case T_ARRAY:
		if isGlobalVar(obj) {
			fmt.Printf("  # global\n")
			fmt.Printf("  leaq %s+0(%%rip), %%rax\n", obj.Name)
		} else {
			panic("TBI")
			fmt.Printf("  # local\n")
			localOffset := (getObjData(obj))
			fmt.Printf("  leaq %d(%%rbp), %%rax # %s \n", localOffset, obj.Name)
		}
		fmt.Printf("  pushq %%rax # array head\n")
	default:
		throw(decl.Type)
	}
}

func throw(x interface{}) {
	panic(fmt.Sprintf("%#v", x))
}

func getSizeOfType(typeExpr ast.Expr) int {
	switch typ := typeExpr.(type) {
	case *ast.Ident:
		data,ok := typ.Obj.Data.(int)
		if !ok {
			throw(typ.Obj)
		}
		return data
	}
	panic("Unexpected")
	return 0
}

func emitAddr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Ident:
		fmt.Printf("  # ident kind=%v\n", e.Obj.Kind)
		fmt.Printf("  # Obj=%v\n", e.Obj)
		if e.Obj.Kind == ast.Var {
			emitVariableAddr(e.Obj)
		} else {
			panic("Unexpected ident kind")
		}
	case *ast.IndexExpr:
		emitExpr(e.Index)
		emitAddr(e.X)
		elmType := getTypeOfExpr(e)
		size := getSizeOfType(elmType)
		fmt.Printf("  popq %%rax # collection addr\n")
		fmt.Printf("  popq %%rcx # index\n")
		fmt.Printf("  movq $%d, %%rdx # elm size\n", size)
		fmt.Printf("  imulq %%rdx, %%rcx\n")
		fmt.Printf("  addq %%rcx, %%rax\n")
		fmt.Printf("  pushq %%rax # addr of element\n")
	default:
		throw(expr)
	}
}

func emitExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Ident:
		fmt.Printf("  # ident kind=%v\n", e.Obj.Kind)
		fmt.Printf("  # Obj=%v\n", e.Obj)
		if e.Obj.Kind == ast.Var {
			emitVariable(e.Obj)
		} else {
			panic("Unexpected ident kind")
		}
	case *ast.CallExpr:
		fun := e.Fun
		fmt.Printf("  # callExpr=%#v\n", fun)
		switch fn := fun.(type) {
		case *ast.Ident:
			switch fn.Obj.Kind {
			case ast.Typ:
				// Conversion
				fmt.Printf("# Conversion %s => %s\n", fn.Obj, getTypeOfExpr(e.Args[0]))
				switch fn.Obj {
				case gString: // string(e)
					switch getTypeKind(getTypeOfExpr(e.Args[0])) {
					case T_SLICE: // slice -> string
						emitExpr(e.Args[0]) // slice
						fmt.Printf("  popq %%rax # ptr\n")
						fmt.Printf("  popq %%rcx # len\n")
						fmt.Printf("  popq %%rdx # cap (to be abandoned)\n")
						fmt.Printf("  pushq %%rcx # str len\n")
						fmt.Printf("  pushq %%rax # str ptr\n")
						return
					}
				default:
					throw(fn.Obj)
				}
			case ast.Fun:

			}
			if fn.Name == "print" {
				// builtin print
				emitExpr(e.Args[0]) // push ptr, push len
				switch getTypeKind(getTypeOfExpr(e.Args[0]))  {
				case T_STRING:
					symbol := fmt.Sprintf("runtime.printstring")
					fmt.Printf("  callq %s\n", symbol)
					fmt.Printf("  addq $16, %%rsp # revert for one string\n")
				case T_INT:
					symbol := fmt.Sprintf("runtime.printint")
					fmt.Printf("  callq %s\n", symbol)
					fmt.Printf("  addq $8, %%rsp # revert for one int\n")
				default:
					panic("TBI")
				}
			} else {
				// general funcall
				var totalSize int = 0
				for i:=len(e.Args) - 1;i>=0;i-- {
					arg := e.Args[i]
					emitExpr(arg)
					size := getExprSize(arg)
					totalSize += size
				}
				symbol := "main." + fn.Name
				fmt.Printf("  callq %s\n", symbol)
				fmt.Printf("  addq $%d, %%rsp # revert\n", totalSize)

				obj := fn.Obj //.Kind == FN
				fndecl,ok := obj.Decl.(*ast.FuncDecl)
				if !ok {
					throw(fn.Obj)
				}
				if fndecl.Type.Results != nil {
					if len(fndecl.Type.Results.List) > 2 {
						panic("TBI")
					} else if len(fndecl.Type.Results.List) == 1 {
						retval0 := fndecl.Type.Results.List[0]
						switch getTypeKind(retval0.Type) {
						case T_STRING:
							fmt.Printf("  # fn.Obj=%#v\n", obj)
							fmt.Printf("  pushq %%rsi # str len\n")
							fmt.Printf("  pushq %%rax # str ptr\n")
						case T_INT:
							fmt.Printf("  # fn.Obj=%#v\n", obj)
							fmt.Printf("  pushq %%rax\n")
						default:
							throw(retval0.Type)
						}
					}
				}
			}
		case *ast.SelectorExpr:
			emitExpr(e.Args[0])
			symbol := fmt.Sprintf("%s.%s", fn.X, fn.Sel)
			fmt.Printf("  callq %s\n", symbol)
		default:
			throw(fun)
		}
	case *ast.ParenExpr:
		emitExpr(e.X)
	case *ast.BasicLit:
		fmt.Printf("  # start %T\n", e)
		fmt.Printf("  # kind=%s\n", e.Kind)
		switch e.Kind.String() {
		case "CHAR":
			val := e.Value
			char := val[1]
			ival := int(char)
			fmt.Printf("  pushq $%d # convert char literal to int\n", ival)
		case "INT":
			val := e.Value
			ival, _ := strconv.Atoi(val)
			fmt.Printf("  pushq $%d # number literal\n", ival)
		case "STRING":
			// e.Value == ".S%d:%d"
			splitted := strings.Split(e.Value, ":")
			fmt.Printf("  pushq $%s # str len\n", splitted[1])
			fmt.Printf("  leaq %s, %%rax # str ptr\n", splitted[0])
			fmt.Printf("  pushq %%rax # str ptr\n")
		default:
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
			throw(e.Op)
		}
		fmt.Printf("  # end %T\n", e)
	case *ast.CompositeLit:
		panic("TBI")
	case *ast.IndexExpr:
		emitAddr(e) // emit addr of element
		fmt.Printf("  popq %%rax # addr of element\n")
		typ :=getTypeOfExpr(e)
		size := getSizeOfType(typ)
		switch size {
		case 1:
			fmt.Printf("  movb (%%rax), %%al # load 1 byte\n")
		default:
			panic("TBI")
		}
		fmt.Printf("  pushq %%rax #\n")
	case *ast.SliceExpr:
		//e.Index, e.X
		emitAddr(e.X) // array head
		emitExpr(e.Low) // intval
		emitExpr(e.High) // intval
		//emitExpr(e.Max) // @TODO
		fmt.Printf("  popq %%rax # high\n")
		fmt.Printf("  popq %%rcx # low\n")
		fmt.Printf("  popq %%rdx # array\n")
		fmt.Printf("  subq %%rcx, %%rax # high - low\n")
		fmt.Printf("  pushq %%rax # cap\n")
		fmt.Printf("  pushq %%rax # len\n")
		fmt.Printf("  pushq %%rdx # array\n")
	default:
		throw(expr)
	}
}

func emitFuncDecl(pkgPrefix string, fnc *Func) {
	funcDecl := fnc.decl
	fmt.Printf("%s.%s: # args %d, locals %d\n",
		pkgPrefix, funcDecl.Name, fnc.argsarea, fnc.localarea)
	fmt.Printf("  pushq %%rbp\n")
	fmt.Printf("  movq %%rsp, %%rbp\n")
	if len(fnc.localvars) > 0 {
		fmt.Printf("  subq $%d, %%rsp # local area\n", fnc.localarea)
	}
	for _, stmt := range funcDecl.Body.List {
		switch s := stmt.(type) {
		case *ast.ExprStmt:
			expr := s.X
			emitExpr(expr)
		case *ast.DeclStmt:
			continue
		case *ast.AssignStmt:
			fmt.Printf("  # *ast.AssignStmt\n")
			lhs := s.Lhs[0]
			fmt.Printf("# lhs=%#v\n", lhs)
			rhs := s.Rhs[0]
			emitAddr(lhs)
			emitExpr(rhs) // push len, push ptr
			fmt.Printf("# %#v\n", lhs)
			switch getTypeKind(getTypeOfExpr(lhs)) {
			case T_STRING:
				fmt.Printf("  popq %%rcx # rhs ptr\n")
				fmt.Printf("  popq %%rax # rhs len\n")
				fmt.Printf("  popq %%rdx # lhs ptr addr\n")
				fmt.Printf("  popq %%rsi # lhs len addr\n")
				fmt.Printf("  movq %%rcx, (%%rdx) # ptr to ptr\n")
				fmt.Printf("  movq %%rax, (%%rsi) # len to len\n")
			case T_SLICE:
				fmt.Printf("  popq %%rcx # rhs ptr\n")
				fmt.Printf("  popq %%rax # rhs len\n")
				fmt.Printf("  popq %%r8 # rhs cap\n")
				fmt.Printf("  popq %%rdx # lhs ptr addr\n")
				fmt.Printf("  popq %%rsi # lhs len addr\n")
				fmt.Printf("  popq %%r9 # lhs cap\n")
				fmt.Printf("  movq %%rcx, (%%rdx) # ptr to ptr\n")
				fmt.Printf("  movq %%rax, (%%rsi) # len to len\n")
				fmt.Printf("  movq %%r8, (%%r9) # cap to cap\n")

			case T_INT:
				fmt.Printf("  popq %%rdi # rhs evaluated\n")
				fmt.Printf("  popq %%rax # lhs addr\n")
				fmt.Printf("  movq %%rdi, (%%rax) # assign\n")
			case T_UINT8:
				fmt.Printf("  popq %%rdi # rhs evaluated\n")
				fmt.Printf("  popq %%rax # lhs addr\n")
				fmt.Printf("  movb %%dil, (%%rax) # assign byte\n")
			default:
				panic("TBI")
			}
		case *ast.ReturnStmt:
			if len(s.Results) == 1 {
				emitExpr(s.Results[0])
				switch getTypeKind(getTypeOfExpr(s.Results[0])) {
				case T_INT:
					fmt.Printf("  popq %%rax # return int\n")
				case T_STRING:
					fmt.Printf("  popq %%rax # return string (ptr)\n")
					fmt.Printf("  popq %%rsi # return string (len)\n")
				default:
					panic("TBI")
				}


				fmt.Printf("  leave\n")
				fmt.Printf("  ret\n")
			} else if len(s.Results) == 0 {
				fmt.Printf("  leave\n")
				fmt.Printf("  ret\n")
			} else {
				panic("TBI")
			}
		default:
			throw(stmt)
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
		switch e.Kind.String() {
		case "INT":
		case "CHAR":
		case "STRING":
			e.Value = registerStringLiteral(e.Value)
		default:
			panic("Unexpected literal kind:" + e.Kind.String())
		}
	case *ast.BinaryExpr:
		walkExpr(e.X) // left
		walkExpr(e.Y) // right
	case *ast.CompositeLit:
		// what to do ?
	case *ast.IndexExpr:
		walkExpr(e.Index)
	case *ast.SliceExpr:
		if e.Low != nil {
			walkExpr(e.Low)
		}
		if e.High != nil {
			walkExpr(e.High)
		}
		if e.Max != nil {
			walkExpr(e.Max)
		}
		walkExpr(e.X)
	default:
		throw(expr)
	}
}

const sliceSize = 24

var gString = &ast.Object{
	Kind: ast.Typ,
	Name: "string",
	Decl: nil,
	Data: 16,
	Type: nil,
}

var gInt = &ast.Object{
	Kind: ast.Typ,
	Name: "int",
	Decl: nil,
	Data: 8,
	Type: nil,
}

var gUint8 = &ast.Object{
	Kind: ast.Typ,
	Name: "uint8",
	Decl: nil,
	Data: 1,
	Type: nil,
}

var globalVars []*ast.ValueSpec
var globalFuncs []*Func

func semanticAnalyze(fset *token.FileSet, fiile *ast.File) {
	// https://github.com/golang/example/tree/master/gotypes#an-example
	// Type check
	// A Config controls various options of the type checker.
	// The defaults work fine except for one setting:
	// we must specify how to deal with imports.
	conf := types.Config{Importer: importer.Default()}

	// Type-check the package containing only file fiile.
	// Check returns a *types.Package.
	pkg, err := conf.Check("./t", fset, []*ast.File{fiile}, nil)
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
	universe.Insert(gUint8)

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
	ap, _ := ast.NewPackage(fset, map[string]*ast.File{"": fiile}, nil, universe)

	var unresolved []*ast.Ident
	for _, ident := range fiile.Unresolved {
		if obj := universe.Lookup(ident.Name); obj != nil {
			ident.Obj = obj
		} else {
			unresolved = append(unresolved, ident)
		}
	}

	fmt.Printf("# Name:    %s\n", pkg.Name())
	fmt.Printf("# Unresolved: %v\n", unresolved)
	fmt.Printf("# Package:   %s\n", ap.Name)


	for _, decl := range fiile.Decls {
		switch dcl := decl.(type) {
		case *ast.GenDecl:
			switch dcl.Tok {
			case token.VAR:
				spec := dcl.Specs[0]
				valSpec := spec.(*ast.ValueSpec)
				fmt.Printf("# valSpec.type=%#v\n", valSpec.Type)
				nameIdent := valSpec.Names[0]
				nameIdent.Obj.Data = -1 // mark as global
				if len(valSpec.Values) > 0 {
					fmt.Printf("# spec.Name=%s, Value=%v\n", nameIdent, valSpec.Values[0])
					fmt.Printf("# nameIdent.Obj=%v\n", nameIdent.Obj)
					switch getTypeKind(valSpec.Type) {
					case T_STRING:
						lit,ok := valSpec.Values[0].(*ast.BasicLit)
						if !ok {
							throw(valSpec.Type)
						}
						lit.Value = registerStringLiteral(lit.Value)
					case T_INT:
						_,ok := valSpec.Values[0].(*ast.BasicLit)
						if !ok {
							throw(valSpec.Type)
						}
					default:
						throw(valSpec.Type)
					}
				}
				globalVars = append(globalVars, valSpec)
			}
		case *ast.FuncDecl:
			funcDecl := decl.(*ast.FuncDecl)
			var localvars []*ast.ValueSpec = nil
			var localoffset int
			var paramoffset int = 16
			for _, field := range funcDecl.Type.Params.List {
				obj :=field.Names[0].Obj
				var varSize int
				switch getTypeKind(field.Type) {
				case T_STRING:
					varSize = gString.Data.(int)
				case T_INT:
					varSize = gInt.Data.(int)
				default:
					panic("TBI")
				}
				setObjData(obj, paramoffset)
				paramoffset += varSize
				fmt.Printf("# field.Names[0].Obj=%#v\n", obj)
			}
			if funcDecl.Body == nil {
				break
			}
			for _, stmt := range funcDecl.Body.List {
				switch s := stmt.(type) {
				case *ast.ExprStmt:
					expr := s.X
					walkExpr(expr)
				case *ast.DeclStmt:
					decl := s.Decl
					switch dcl := decl.(type) {
					case *ast.GenDecl:
						declSpec := dcl.Specs[0]
						switch ds := declSpec.(type) {
						case *ast.ValueSpec:
							varSpec := ds
							obj := varSpec.Names[0].Obj
							var varSize int
							switch getTypeKind(varSpec.Type)  {
							case T_SLICE:
								varSize = sliceSize
							case T_STRING:
								varSize = gString.Data.(int)
							case T_INT:
								varSize = gInt.Data.(int)
							default:
								throw(varSpec.Type)
							}

							localoffset -= varSize
							setObjData(obj, localoffset)
							localvars = append(localvars, ds)
						}
					default:
						throw(decl)
					}
				case *ast.AssignStmt:
					//lhs := s.Lhs[0]
					rhs := s.Rhs[0]
					walkExpr(rhs)
				case *ast.ReturnStmt:
					for _, r := range s.Results {
						walkExpr(r)
					}
				default:
					throw(stmt)
				}
			}
			fnc := &Func{
				decl:      funcDecl,
				localvars: localvars,
				localarea: -localoffset,
				argsarea: paramoffset,
			}
			globalFuncs = append(globalFuncs, fnc)
		default:
			throw(decl)
		}
	}
}

type Func struct {
	decl      *ast.FuncDecl
	localvars []*ast.ValueSpec
	localarea int
	argsarea  int
}

func getExprSize(valueExpr ast.Expr) int {
	switch getTypeKind(getTypeOfExpr(valueExpr)) {
	case T_STRING:
		return 8*2
	case T_SLICE:
		return 8*3
	case T_INT:
		return 8
	case T_ARRAY:
		panic("TBI")
	default:
		throw(valueExpr)
	}
	return 0
}

const T_STRING = "T_STRING"
const T_SLICE = "T_SLICE"
const T_INT = "T_INT"
const T_UINT8 = "T_UINT8"
const T_ARRAY = "T_ARRAY"

func getTypeOfExpr(expr ast.Expr) ast.Expr {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Obj.Kind == ast.Var {
			switch dcl := e.Obj.Decl.(type) {
			case *ast.ValueSpec:
				return dcl.Type
			case *ast.Field:
				return dcl.Type
			default:
				throw(e.Obj)
			}
		} else {
			throw(e.Obj)
		}
	case *ast.BasicLit:
		switch e.Kind.String() {
		case "STRING":
			return &ast.Ident{
				NamePos: 0,
				Name:    "string",
				Obj:     gString,
			}
		case "INT":
			return &ast.Ident{
				NamePos: 0,
				Name:    "int",
				Obj:     gInt,
			}
		default:
			throw(e)
		}
	case *ast.BinaryExpr:
		return getTypeOfExpr(e.X)
	case *ast.IndexExpr:
		collection := e.X
		typ := getTypeOfExpr(collection)
		switch tp := typ.(type) {
		case *ast.ArrayType:
			return tp.Elt
		default:
			panic(fmt.Sprintf("Unexpected expr type:%#v", typ))
		}
	case *ast.CallExpr: // funcall or conversion
		switch fn := e.Fun.(type) {
		case *ast.Ident:
			switch fn.Obj.Kind {
			case ast.Typ:
				return fn
			}
		}
		panic("TBI")
	default:
		panic(fmt.Sprintf("Unexpected expr type:%#v", expr))
	}
	throw(expr)
	return nil
}

func getTypeKind(typeExpr ast.Expr) string {
	switch e := typeExpr.(type) {
	case *ast.Ident:
		if e.Obj.Kind == ast.Var {
			throw(e.Obj)
		} else if e.Obj.Kind == ast.Typ {
			switch e.Obj {
			case gInt:
				return T_INT
			case gString:
				return T_STRING
			case gUint8:
				return T_UINT8
			default:
				throw(e.Obj)
			}
		}
	case *ast.ArrayType:
		if e.Len == nil {
			return T_SLICE
		} else {
			return T_ARRAY
		}
	default:
		throw(typeExpr)
	}
	return ""
}

func emitData() {
	fmt.Printf(".data\n")
	for i, sl := range stringLiterals {
		fmt.Printf("# string literals\n")
		fmt.Printf(".S%d:\n", i)
		fmt.Printf("  .string %s\n", sl)
	}

	fmt.Printf("# ===== Global Variables =====\n")
	for _, varDecl := range globalVars {
		name := varDecl.Names[0]
		var val ast.Expr
		if len(varDecl.Values) > 0 {
			val = varDecl.Values[0]
		}

		fmt.Printf("%s: # T %s\n", name, getTypeKind(varDecl.Type))
		switch getTypeKind(varDecl.Type) {
		case T_STRING:
			switch vl := val.(type) {
			case *ast.BasicLit:
				var strval string
				strval = vl.Value
				splitted := strings.Split(strval, ":")
				fmt.Printf("  .quad %s\n", splitted[0])
				fmt.Printf("  .quad %s\n", splitted[1])
			default:
				panic("Unexpected case")
			}
		case T_INT:
			switch vl := val.(type) {
			case *ast.BasicLit:
				fmt.Printf("  .quad %s\n", vl.Value)
			default:
				panic("Unexpected case")
			}
		case T_ARRAY:
			if val == nil {
				arrayType,ok :=  varDecl.Type.(*ast.ArrayType)
				if !ok {
					panic("Unexpected")
				}
				basicLit, ok := arrayType.Len.(*ast.BasicLit)
				length, err := strconv.Atoi(basicLit.Value)
				if err != nil {
					panic(fmt.Sprintf("%#v\n", basicLit.Value))
				}
				for i:=0;i<length;i++ {
					fmt.Printf("  .byte %d\n", 0)
				}
			} else {
				panic("TBI")
			}
		default:
			throw(varDecl.Type)
		}
	}
	fmt.Printf("# ==============================\n")
}

func emitText() {
	fmt.Printf(".text\n")
	for _, fnc := range globalFuncs {
		emitFuncDecl("main", fnc)
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
