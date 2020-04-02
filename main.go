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
		panic("obj.Data is not int")
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
		panic("unexpected")
	}

	fmt.Printf("  # emitVariable\n")
	fmt.Printf("  # Type=%#v\n", typ)
	fmt.Printf("  # obj.Data=%d\n", obj.Data)

	switch getTypeKind(typ) {
	case T_STRING:
		if isGlobalVar(obj) {
			fmt.Printf("  # global variable %s\n", obj.Name)
			fmt.Printf("  movq %s+0(%%rip), %%rax\n", obj.Name)
			fmt.Printf("  movq %s+8(%%rip), %%rcx\n", obj.Name)
		} else {
			fmt.Printf("  # load local variable %s\n", obj.Name)
			fmt.Printf("  movq %d(%%rbp), %%rax # ptr\n", localOffset)
			fmt.Printf("  movq %d(%%rbp), %%rcx # len\n", (localOffset + 8))
		}
		fmt.Printf("  pushq %%rcx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case T_INT:
		if isGlobalVar(obj) {
			fmt.Printf("  # global\n")
			fmt.Printf("  movq %s+0(%%rip), %%rax\n", obj.Name)
		} else {
			fmt.Printf("  # local\n")
			fmt.Printf("  movq %d(%%rbp), %%rax # %s \n", localOffset, obj.Name)
		}
		fmt.Printf("  pushq %%rax # int val\n")
	default:
		panic(fmt.Sprintf("Unexpected type:%#v", getTypeKind(typ)))
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
	case T_STRING:
		if isGlobalVar(obj) {
			fmt.Printf("  # global\n")
			fmt.Printf("  leaq %s+0(%%rip), %%rax # ptr\n", obj.Name)
			fmt.Printf("  leaq %s+8(%%rip), %%rcx # len\n", obj.Name)
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

	default:
		panic(fmt.Sprintf("Unexpected type:%s :%#v", getTypeKind(decl.Type), decl.Type))
	}
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
		fmt.Printf("  # funcall=%T\n", fun)
		switch fn := fun.(type) {
		case *ast.Ident:
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
					panic("Unexpectred")
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
							panic("TBI")
						}
					}
				}
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
			panic(fmt.Sprintf("Unexpected binary operator %s", e.Op))
		}
		fmt.Printf("  # end %T\n", e)
	case *ast.CompositeLit:
		panic("TBI")
	default:
		panic(fmt.Sprintf("Unexpected expr type %T", expr))
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
			rhs := s.Rhs[0]
			emitAddr(lhs)
			emitExpr(rhs) // push len, push ptr
			switch getTypeKind(getTypeOfExpr(lhs)) {
			case T_STRING:
				fmt.Printf("  popq %%rcx # rhs ptr\n")
				fmt.Printf("  popq %%rax # rhs len\n")
				fmt.Printf("  popq %%rdx # lhs ptr addr\n")
				fmt.Printf("  popq %%rsi # lhs len addr\n")
				fmt.Printf("  movq %%rcx, (%%rdx) # ptr to ptr\n")
				fmt.Printf("  movq %%rax, (%%rsi) # len to len\n")
			case T_INT:
				fmt.Printf("  popq %%rdi # rhs evaluated\n")
				fmt.Printf("  popq %%rax # lhs addr\n")
				fmt.Printf("  movq %%rdi, (%%rax)\n")
			default:
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
			panic(fmt.Sprintf("Unexpected stmt type %T", stmt))
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

var gUint8 = &ast.Object{
	Kind: ast.Typ,
	Name: "uint8",
	Decl: nil,
	Data: nil,
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
							panic("Unexpected type")
						}
						lit.Value = registerStringLiteral(lit.Value)
					case T_INT:
						_,ok := valSpec.Values[0].(*ast.BasicLit)
						if !ok {
							panic("Unexpected type")
						}
					default:
						panic(fmt.Sprintf("Unexpected type:%T",valSpec.Type))
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
					varSize = 16
				case T_INT:
					varSize = 8
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
							case T_STRING:
								varSize = 16
							case T_INT:
								varSize = 8
							default:
								panic("TBI")
							}

							localoffset -= varSize
							setObjData(obj, localoffset)
							localvars = append(localvars, ds)
						}
					default:
						panic(fmt.Sprintf("Unexpected type:%T", decl))
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
					panic(fmt.Sprintf("Unexpected stmt type:%T", stmt))
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
			panic("unexpected decl type")
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
	case T_INT:
		return 8
	case T_ARRAY:
		panic(fmt.Sprintf("TBI: type %#v",(valueExpr) ))

	default:
		panic(fmt.Sprintf("TBI: type %#v",(valueExpr) ))
	}
	return 0
}

const T_STRING = "T_STRING"
const T_INT = "T_INT"
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
				panic(fmt.Sprintf("Unexpected type:%T", dcl))
			}
		} else {
			panic(fmt.Sprintf("TBI:%#v", e.Obj))
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
			panic(fmt.Sprintf("%v", e.Kind))
		}
	case *ast.BinaryExpr:
		return getTypeOfExpr(e.X)
	default:
		panic(fmt.Sprintf("Unexpected expr type:%#v", expr))
	}
}

func getTypeKind(typeExpr ast.Expr) string {
	switch e := typeExpr.(type) {
	case *ast.Ident:
		if e.Obj.Kind == ast.Var {
			panic(fmt.Sprintf("Unexpected type:%#v", e.Obj))
		} else if e.Obj.Kind == ast.Typ {
			switch e.Obj {
			case gInt:
				return T_INT
			case gString:
				return T_STRING
			default:
				panic(fmt.Sprintf("TBI:%#v", e.Obj))
			}
		}
	case *ast.ArrayType:
		return T_ARRAY
	default:
		panic(fmt.Sprintf("Unexpected typeExpr type:%#v", typeExpr))
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
			panic(fmt.Sprintf("Unexpected type: %s", getTypeKind(varDecl.Type)))
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
