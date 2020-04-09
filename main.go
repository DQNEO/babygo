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

const globalFlag int = 99999

func isGlobalVar(obj *ast.Object) bool {
	return getObjData(obj) == globalFlag
}

func setObjData(obj *ast.Object, i int) {
	obj.Data = i
}

func getObjData(obj *ast.Object) int {
	objData, ok := obj.Data.(int)
	if !ok {
		throw(obj)
	}
	return objData
}

func emitLoad(typ ast.Expr) {
	fmt.Printf("  popq %%rdx\n")
	switch getTypeKind(typ) {
	case T_SLICE:
		fmt.Printf("  movq %d(%%rdx), %%rax\n", 0)
		fmt.Printf("  movq %d(%%rdx), %%rcx\n", 8)
		fmt.Printf("  movq %d(%%rdx), %%rdx\n", 16)

		fmt.Printf("  pushq %%rdx # cap\n")
		fmt.Printf("  pushq %%rcx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case T_STRING:
		fmt.Printf("  movq %d(%%rdx), %%rax\n", 0)
		fmt.Printf("  movq %d(%%rdx), %%rdx\n", 8)

		fmt.Printf("  pushq %%rdx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case T_UINT8:
		fmt.Printf("  movzbq %d(%%rdx), %%rdx # load uint8\n", 0)
		fmt.Printf("  pushq %%rdx\n")
	case T_UINT16:
		fmt.Printf("  movzwq %d(%%rdx), %%rdx # load uint16\n", 0)
		fmt.Printf("  pushq %%rdx\n")
	case T_INT, T_BOOL, T_UINTPTR:
		fmt.Printf("  movq %d(%%rdx), %%rdx # load int\n", 0)
		fmt.Printf("  pushq %%rdx\n")
	default:
		throw(typ)
	}
}

func emitVariableAddr(obj *ast.Object) {
	assert(obj.Kind == ast.Var, "obj should be ast.Var")

	var localOffset int
	localOffset = getObjData(obj)

	var scope_comment string
	if isGlobalVar(obj) {
		scope_comment = "global"
	} else {
		scope_comment = "local"
	}
	fmt.Printf("  # obj %#v\n", obj)
	fmt.Printf("  # emitVariableAddr %s \"%s\" Data=%d %#v\n", scope_comment, obj.Name, obj.Data, obj.Decl)


	var addr string
	if isGlobalVar(obj) {
		addr = fmt.Sprintf("%s(%%rip)", obj.Name)
	} else {
		addr = fmt.Sprintf("%d(%%rbp)", localOffset)
	}

	fmt.Printf("  leaq %s, %%rax # addr\n", addr)
	fmt.Printf("  pushq %%rax\n")
}

func assert(bol bool, msg string) {
	if !bol {
		panic(msg)
	}
}

func throw(x interface{}) {
	panic(fmt.Sprintf("%#v", x))
}

func getSizeOfType(typeExpr ast.Expr) int {
	switch typ := typeExpr.(type) {
	case *ast.Ident:
		if typ.Obj == nil {
			throw(typ)
		}
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
		if e.Obj.Kind == ast.Var {
			emitVariableAddr(e.Obj)
		} else {
			panic("Unexpected ident kind")
		}
	case *ast.IndexExpr:
		elmType := getTypeOfExpr(e)
		size := getSizeOfType(elmType)

		emitExpr(e.Index) // index number
		switch getTypeKind(getTypeOfExpr(e.X)) {
		case T_ARRAY:
			emitAddr(e.X) // array head
		case T_SLICE:
			emitExpr(e.X)
			fmt.Printf("  popq %%rax # slice.ptr\n")
			fmt.Printf("  popq %%rcx # garbage\n")
			fmt.Printf("  popq %%rcx # garbage\n")
			fmt.Printf("  pushq %%rax # slice.ptr\n")
		}
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

// Conversion slice(string)
func emitConversionToSlice(fn *ast.ArrayType, arg0 ast.Expr) {
	assert(fn.Len == nil, "fn should be slice")
	assert(getTypeKind(getTypeOfExpr(arg0)) == T_STRING, "fn should be slice")
	fmt.Printf("  # Conversion to slice %s <= %s\n", fn.Elt, getTypeOfExpr(arg0))
	emitExpr(arg0)
	fmt.Printf("  popq %%rax # ptr\n")
	fmt.Printf("  popq %%rcx # len\n")
	fmt.Printf("  pushq %%rcx # cap\n")
	fmt.Printf("  pushq %%rcx # len\n")
	fmt.Printf("  pushq %%rax # ptr\n")
}

func emitConversion(fn *ast.Ident, arg0 ast.Expr) {
	fmt.Printf("  # general Conversion %s <= %s\n", fn.Obj, getTypeOfExpr(arg0))
	switch fn.Obj {
	case gString: // string(e)
		switch getTypeKind(getTypeOfExpr(arg0)) {
		case T_SLICE: // string(slice)
			emitExpr(arg0) // slice
			fmt.Printf("  popq %%rax # ptr\n")
			fmt.Printf("  popq %%rcx # len\n")
			fmt.Printf("  popq %%rdx # cap (to be abandoned)\n")
			fmt.Printf("  pushq %%rcx # str len\n")
			fmt.Printf("  pushq %%rax # str ptr\n")
		}
	case gInt, gUint8, gUint16, gUintptr: // int(e)
		emitExpr(arg0)
	default:
		throw(fn.Obj)
	}
	return
}

func emitExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Obj == nil {
			panic(fmt.Sprintf("ident %s is unresolved", e.Name))
		}
		if e.Obj.Kind == ast.Var {
			switch e.Obj {
			case gTrue:
				fmt.Printf("  pushq $1 # true\n")
				return
			case gFalse:
				fmt.Printf("  pushq $0 # false\n")
				return
			}
			emitVariableAddr(e.Obj)
			var typ ast.Expr
			switch dcl := e.Obj.Decl.(type) {
			case *ast.ValueSpec:
				typ = dcl.Type
			case *ast.Field:
				typ = dcl.Type
			default:
				throw(e.Obj)
			}
			emitLoad(typ)
		} else {
			panic("Unexpected ident kind:" + e.Obj.Kind.String())
		}
	case *ast.SelectorExpr:
		symbol := fmt.Sprintf("%s.%s", e.X, e.Sel) // e.g. os.Stdout
		panic(symbol)
	case *ast.CallExpr:
		fun := e.Fun
		fmt.Printf("  # callExpr=%#v\n", fun)
		switch fn := fun.(type) {
		case *ast.ArrayType: // Conversion to slice
			emitConversionToSlice(fn, e.Args[0])
		case *ast.Ident:
			if fn.Obj == nil {
				panic("unresolved ident: " + fn.String())
			}
			switch fn.Obj {
			case gMake:
				var typeArg ast.Expr = e.Args[0]
				switch getTypeKind(typeArg) {
				case T_SLICE:
					// make([]T, ...)
					arrayType, ok := typeArg.(*ast.ArrayType)
					assert(ok, "should be *ast.ArrayType")


					var lenArg ast.Expr = e.Args[1]
					var capArg ast.Expr = e.Args[2]
					emitExpr(capArg)
					emitExpr(lenArg)
					elmSize := getSizeOfType(arrayType.Elt)
					fmt.Printf("  pushq $%d # elm size\n", elmSize)
					symbol := "runtime.makeSlice"
					fmt.Printf("  callq %s\n", symbol)
					fmt.Printf("  addq $24, %%rsp # revert for one string\n")
					fmt.Printf("  pushq %%rsi # slice cap\n")
					fmt.Printf("  pushq %%rdi # slice len\n")
					fmt.Printf("  pushq %%rax # slice ptr\n")
					return
				default:
					throw(typeArg)
				}
			}
			switch fn.Obj.Kind {
			case ast.Typ:
				// Conversion
				emitConversion(fn, e.Args[0])
				return
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
				symbol := pkgName + "." + fn.Name
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
							fmt.Printf("  pushq %%rdi # str len\n")
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
			symbol := fmt.Sprintf("%s.%s", fn.X, fn.Sel) // syscall.Write() or unsafe.Pointer(x)
			switch symbol {
			case "unsafe.Pointer":
				emitExpr(e.Args[0])
			case "syscall.Write":
				emitExpr(e.Args[1])
				emitExpr(e.Args[0])
				fmt.Printf("  callq %s\n", symbol) // func decl is in runtime
			default:
				panic(symbol)
			}
		default:
			throw(fun)
		}
	case *ast.ParenExpr:
		emitExpr(e.X)
	case *ast.BasicLit:
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
	case *ast.UnaryExpr:
		switch e.Op.String() {
		case "-":
			emitExpr(e.X)
			fmt.Printf("  popq %%rax # e.X\n")
			fmt.Printf("  imulq $-1, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "&":
			emitAddr(e.X)
		default:
			throw(e.Op.String())
		}
	case *ast.BinaryExpr:
		fmt.Printf("  # start %T\n", e)
		emitExpr(e.X) // left
		emitExpr(e.Y) // right
		switch e.Op.String()  {
		case "+":
			fmt.Printf("  popq %%rdi # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  addq %%rdi, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "-":
			fmt.Printf("  popq %%rdi # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  subq %%rdi, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "*":
			fmt.Printf("  popq %%rdi # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  imulq %%rdi, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "%":
			fmt.Printf("  popq %%rcx # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  movq $0, %%rdx # init %%rdx\n")
			fmt.Printf("  divq %%rcx\n")
			fmt.Printf("  movq %%rdx, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "/":
			fmt.Printf("  popq %%rcx # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  movq $0, %%rdx # init %%rdx\n")
			fmt.Printf("  divq %%rcx\n")
			fmt.Printf("  pushq %%rax\n")
		case "==":
			emitCompExpr("sete")
		case "!=":
			emitCompExpr("setne")
		case "<":
			emitCompExpr("setl")
		case "<=":
			emitCompExpr("setle")
		case ">":
			emitCompExpr("setg")
		case ">=":
			emitCompExpr("setge")
		default:
			panic(fmt.Sprintf("TBI: binary operation for '%s'", e.Op.String()))
		}
		fmt.Printf("  # end %T\n", e)
	case *ast.CompositeLit:
		panic("TBI")
	case *ast.IndexExpr:
		emitAddr(e) // emit addr of element
		typ := getTypeOfExpr(e)
		emitLoad(typ)
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

//@TODO handle larger types than int
func emitCompExpr(inst string) {
	fmt.Printf("  popq %%rcx # right\n")
	fmt.Printf("  popq %%rax # left\n")
	fmt.Printf("  cmpq %%rcx, %%rax\n")
	fmt.Printf("  %s %%al\n", inst)
	fmt.Printf("  movzbq %%al, %%rax\n")
	fmt.Printf("  pushq %%rax\n")
}

func emitStore(typ ast.Expr) {
	switch getTypeKind(typ) {
	case T_SLICE:
		fmt.Printf("  popq %%rcx # rhs ptr\n")
		fmt.Printf("  popq %%rax # rhs len\n")
		fmt.Printf("  popq %%r8 # rhs cap\n")

		fmt.Printf("  popq %%rdx # lhs ptr addr\n")
		fmt.Printf("  leaq %d(%%rdx), %%rsi #len \n", 8)
		fmt.Printf("  leaq %d(%%rdx), %%r9 # cap\n", 16)

		fmt.Printf("  movq %%rcx, (%%rdx) # ptr to ptr\n")
		fmt.Printf("  movq %%rax, (%%rsi) # len to len\n")
		fmt.Printf("  movq %%r8, (%%r9) # cap to cap\n")
	case T_STRING:
		fmt.Printf("  popq %%rcx # rhs ptr\n")
		fmt.Printf("  popq %%rax # rhs len\n")

		fmt.Printf("  popq %%rdx # lhs ptr addr\n")
		fmt.Printf("  leaq %d(%%rdx), %%rsi #len \n", 8)

		fmt.Printf("  movq %%rcx, (%%rdx) # ptr to ptr\n")
		fmt.Printf("  movq %%rax, (%%rsi) # len to len\n")
	case T_INT,T_BOOL, T_UINTPTR:
		fmt.Printf("  popq %%rdi # rhs evaluated\n")
		fmt.Printf("  popq %%rax # lhs addr\n")
		fmt.Printf("  movq %%rdi, (%%rax) # assign\n")
	case T_UINT8:
		fmt.Printf("  popq %%rdi # rhs evaluated\n")
		fmt.Printf("  popq %%rax # lhs addr\n")
		fmt.Printf("  movb %%dil, (%%rax) # assign byte\n")
	case T_UINT16:
		fmt.Printf("  popq %%rdi # rhs evaluated\n")
		fmt.Printf("  popq %%rax # lhs addr\n")
		fmt.Printf("  movw %%di, (%%rax) # assign word\n")
	default:
		panic("TBI:" + getTypeKind(typ))
	}
}

func emitStmt(stmt ast.Stmt) {
	fmt.Printf("  # == Stmt %T ==\n", stmt)
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		expr := s.X
		emitExpr(expr)
	case *ast.DeclStmt:
		return // do nothing
	case *ast.AssignStmt:
		switch s.Tok.String() {
		case "=":
			lhs := s.Lhs[0]
			rhs := s.Rhs[0]
			emitAddr(lhs)
			emitExpr(rhs)
			emitStore(getTypeOfExpr(lhs))
		default:
			panic("TBI: assignment of " + s.Tok.String())
		}
	case *ast.ReturnStmt:
		if len(s.Results) == 1 {
			emitExpr(s.Results[0])
			switch getTypeKind(getTypeOfExpr(s.Results[0])) {
			case T_INT, T_UINTPTR:
				fmt.Printf("  popq %%rax # return int\n")
			case T_STRING:
				fmt.Printf("  popq %%rax # return string (ptr)\n")
				fmt.Printf("  popq %%rdi # return string (len)\n")
			case T_SLICE:
				fmt.Printf("  popq %%rax # return string (ptr)\n")
				fmt.Printf("  popq %%rdi # return string (len)\n")
				fmt.Printf("  popq %%rsi # return string (cap)\n")
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
	case *ast.IfStmt:
		fmt.Printf("  # if\n")

		labelid++
		labelEndif := fmt.Sprintf(".L.endif.%d", labelid)
		labelElse := fmt.Sprintf(".L.else.%d", labelid)

		if s.Else != nil {
			emitExpr(s.Cond)
			fmt.Printf("  popq %%rax\n")
			fmt.Printf("  cmpq $0, %%rax\n")
			fmt.Printf("  je %s # jmp if false\n", labelElse)
			emitStmt(s.Body) // then
			fmt.Printf("  jmp %s\n", labelEndif)
			fmt.Printf("  %s:\n", labelElse)
			emitStmt(s.Else) // then
		} else {
			emitExpr(s.Cond)
			fmt.Printf("  popq %%rax\n")
			fmt.Printf("  cmpq $0, %%rax\n")
			fmt.Printf("  je %s # jmp if false\n", labelEndif)
			emitStmt(s.Body) // then
		}
		fmt.Printf("  %s:\n", labelEndif)
		fmt.Printf("  # end if\n")
	case *ast.BlockStmt:
		for _, stmt := range s.List {
			emitStmt(stmt)
		}
	case *ast.ForStmt:
		labelid++
		labelCond := fmt.Sprintf(".L.loop.cond.%d", labelid)
		labelExit := fmt.Sprintf(".L.loop.exit.%d", labelid)

		emitStmt(s.Init)

		fmt.Printf("  %s:\n", labelCond)
		emitExpr(s.Cond)
		fmt.Printf("  popq %%rax\n")
		fmt.Printf("  cmpq $0, %%rax\n")
		fmt.Printf("  je %s # jmp if false\n", labelExit)
		emitStmt(s.Body)
		emitStmt(s.Post)
		fmt.Printf("  jmp %s\n", labelCond)
		fmt.Printf("  %s:\n", labelExit)


	default:
		throw(stmt)
	}
}
var labelid int

func emitFuncDecl(pkgPrefix string, fnc *Func) {
	fmt.Printf("\n")
	funcDecl := fnc.decl
	fmt.Printf("%s.%s: # args %d, locals %d\n",
		pkgPrefix, funcDecl.Name, fnc.argsarea, fnc.localarea)
	fmt.Printf("  pushq %%rbp\n")
	fmt.Printf("  movq %%rsp, %%rbp\n")
	if len(fnc.localvars) > 0 {
		fmt.Printf("  subq $%d, %%rsp # local area\n", fnc.localarea)
	}
	for _, stmt := range funcDecl.Body.List {
		emitStmt(stmt)
	}
	fmt.Printf("  leave\n")
	fmt.Printf("  ret\n")
}

func registerStringLiteral(s string) string {
	rawStringLiteal := s
	stringLiterals = append(stringLiterals, rawStringLiteal)
	var strlen int
	for _, c := range []uint8(rawStringLiteal) {
		if c != '\\' {
			strlen++
		}
	}

	if pkgName == "" {
		panic("no pkgName")
	}
	r := fmt.Sprintf(".%s.S%d:%d", pkgName, stringIndex, strlen - 2)
	stringIndex++
	return r
}

var localvars []*ast.ValueSpec
var localoffset int

func walkStmt(stmt ast.Stmt) {
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
				case T_INT, T_UINTPTR:
					varSize = gInt.Data.(int)
				case T_UINT8:
					varSize = gUint8.Data.(int)
				case T_BOOL:
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
	case *ast.IfStmt:
		walkExpr(s.Cond)
		walkStmt(s.Body)
		if s.Else != nil {
			walkStmt(s.Else)
		}
	case *ast.BlockStmt:
		for _, stmt := range s.List {
			walkStmt(stmt)
		}
	case *ast.ForStmt:
		walkStmt(s.Init)
		walkExpr(s.Cond)
		walkStmt(s.Post)
		walkStmt(s.Body)
	default:
		throw(stmt)
	}

}

func walkExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Ident:
		// what to do ?
	case *ast.SelectorExpr:
		walkExpr(e.X)
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
	case *ast.UnaryExpr:
		walkExpr(e.X)
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
	case *ast.ArrayType: // first argument of builtin func like make()
		// do nothing
	default:
		throw(expr)
	}
}

const sliceSize = 24

var gTrue = &ast.Object{
	Kind: ast.Var,
	Name: "true",
	Decl: nil,
	Data: nil,
	Type: nil,
}

var gFalse = &ast.Object{
	Kind: ast.Var,
	Name: "false",
	Decl: nil,
	Data: nil,
	Type: nil,
}

var gString = &ast.Object{
	Kind: ast.Typ,
	Name: "string",
	Decl: nil,
	Data: 16,
	Type: nil,
}

var gUintptr = &ast.Object{
	Kind: ast.Typ,
	Name: "uintptr",
	Decl: nil,
	Data: 8,
	Type: nil,
}

var gBool = &ast.Object{
	Kind: ast.Typ,
	Name: "bool",
	Decl: nil,
	Data: 8, // same as int for now
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

var gUint16 = &ast.Object{
	Kind: ast.Typ,
	Name: "uint16",
	Decl: nil,
	Data: 2,
	Type: nil,
}



var gMake = &ast.Object{
	Kind: ast.Fun,
	Name: "make",
	Decl: nil,
	Data: nil,
	Type: nil,
}

func semanticAnalyze(fset *token.FileSet, fiile *ast.File) *types.Package {
	// https://github.com/golang/example/tree/master/gotypes#an-example
	// Type check
	// A Config controls various options of the type checker.
	// The defaults work fine except for one setting:
	// we must specify how to deal with imports.
	conf := types.Config{Importer: importer.Default()}

	// Type-check the package containing only file.
	// Check returns a *types.Package.
	pkg, err := conf.Check("./t", fset, []*ast.File{fiile}, nil)
	if err != nil {
		panic(err)
	}
	pkgName = pkg.Name()

	fmt.Printf("# Package  %q\n", pkg.Path())
	universe := &ast.Scope{
		Outer:   nil,
		Objects: make(map[string]*ast.Object),
	}

	// predeclared types
	universe.Insert(gString)
	universe.Insert(gUintptr)
	universe.Insert(gBool)
	universe.Insert(gInt)
	universe.Insert(gUint8)
	universe.Insert(gUint16)

	// predeclared variables
	universe.Insert(gTrue)
	universe.Insert(gFalse)

	// predeclare funcs
	universe.Insert(gMake)

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
				nameIdent.Obj.Data = globalFlag
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
					case T_INT,T_UINT8, T_UINT16, T_UINTPTR:
						_,ok := valSpec.Values[0].(*ast.BasicLit)
						if !ok {
							throw(valSpec.Type) // allow only literal
						}
					default:
						throw(valSpec.Type)
					}
				}
				globalVars = append(globalVars, valSpec)
			}
		case *ast.FuncDecl:
			funcDecl := decl.(*ast.FuncDecl)
			localvars = nil
			localoffset  = 0
			var paramoffset int = 16
			for _, field := range funcDecl.Type.Params.List {
				obj :=field.Names[0].Obj
				var varSize int
				switch getTypeKind(field.Type) {
				case T_STRING:
					varSize = gString.Data.(int)
				case T_INT, T_UINTPTR:
					varSize = gInt.Data.(int)
				default:
					panic(getTypeKind(field.Type))
				}
				setObjData(obj, paramoffset)
				paramoffset += varSize
				fmt.Printf("# field.Names[0].Obj=%#v\n", obj)
			}
			if funcDecl.Body == nil {
				break
			}
			for _, stmt := range funcDecl.Body.List {
				walkStmt(stmt)
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

	return pkg
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
	case T_UINT8:
		return 1
	case T_ARRAY:
		panic("TBI")
	default:
		throw(valueExpr)
	}
	return 0
}

const T_STRING = "T_STRING"
const T_SLICE = "T_SLICE"
const T_BOOL = "T_BOOL"
const T_INT = "T_INT"
const T_UINT8 = "T_UINT8"
const T_UINT16 = "T_UINT16"
const T_UINTPTR = "T_UINTPTR"
const T_ARRAY = "T_ARRAY"
const T_POINTER = "T_POINTER"

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
		case "CHAR":
			return &ast.Ident{
				NamePos: 0,
				Name:    "int",
				Obj:     gInt,
			}
		default:
			throw(e.Kind.String())
		}
	case *ast.UnaryExpr:
		switch e.Op.String() {
		case "-":
			return getTypeOfExpr(e.X)
		default:
			throw(e.Op.String())
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
			case ast.Typ: // conversion
				return fn
			case ast.Fun:
				switch decl := fn.Obj.Decl.(type) {
				case *ast.FuncDecl:
					assert(len(decl.Type.Results.List) == 1, "func is expected to return a single value")
					return decl.Type.Results.List[0].Type
				default:
					throw(fn.Obj.Decl)
				}
			}
		case *ast.SelectorExpr:
			xIdent, ok := fn.X.(*ast.Ident)
			if !ok {
				throw(fn)
			}
			if xIdent.Name == "unsafe" && fn.Sel.Name == "Pointer" {
				// unsafe.Pointer(x)
				return &ast.Ident{
					NamePos: 0,
					Name:    "uintptr",
					Obj:     gUintptr,
				}
			}
			throw(fmt.Sprintf("%#v, %#v\n", xIdent, fn.Sel))
		default:
			throw(e.Fun)
		}
	case *ast.SliceExpr:
		underlyingCollectionType := getTypeOfExpr(e.X)
		var elementTyp ast.Expr
		switch colType := underlyingCollectionType.(type) {
		case *ast.ArrayType:
			elementTyp = colType.Elt
		}
		r := &ast.ArrayType{
			Len: nil,
			Elt: elementTyp,
		}
		return r

	default:
		panic(fmt.Sprintf("Unexpected expr type:%#v", expr))
	}
	throw(expr)
	return nil
}

func getTypeKind(typeExpr ast.Expr) string {
	switch e := typeExpr.(type) {
	case *ast.Ident:
		if e.Obj == nil {
			panic("Unresolved identifier:" +e.Name)
		}
		if e.Obj.Kind == ast.Var {
			throw(e.Obj)
		} else if e.Obj.Kind == ast.Typ {
			switch e.Obj {
			case gUintptr:
				return T_UINTPTR
			case gInt:
				return T_INT
			case gString:
				return T_STRING
			case gUint8:
				return T_UINT8
			case gUint16:
				return T_UINT16
			case gBool:
				return T_BOOL
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
	case *ast.StarExpr:
		return T_POINTER
	default:
		throw(typeExpr)
	}
	return ""
}

func emitData(pkgName string) {
	fmt.Printf(".data\n")
	for i, sl := range stringLiterals {
		fmt.Printf("# string literals\n")
		fmt.Printf(".%s.S%d:\n", pkgName, i)
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
			case nil:
				fmt.Printf("  .quad 0\n")
				fmt.Printf("  .quad 0\n")
			default:
				panic("Unexpected case")
			}
		case T_POINTER:
			fmt.Printf("  .quad 0 # pointer \n") // @TODO
		case T_UINTPTR:
			switch vl := val.(type) {
			case *ast.BasicLit:
				fmt.Printf("  .quad %s\n", vl.Value)
			case nil:
				fmt.Printf("  .quad 0\n")
			default:
				throw(val)
			}
		case T_INT:
			switch vl := val.(type) {
			case *ast.BasicLit:
				fmt.Printf("  .quad %s\n", vl.Value)
			case nil:
				fmt.Printf("  .quad 0\n")
			default:
				throw(val)
			}
		case T_UINT8:
			switch vl := val.(type) {
			case *ast.BasicLit:
				fmt.Printf("  .byte %s\n", vl.Value)
			case nil:
				fmt.Printf("  .byte 0\n")
			default:
				throw(val)
			}
		case T_UINT16:
			switch vl := val.(type) {
			case *ast.BasicLit:
				fmt.Printf("  .word %s\n", vl.Value)
			case nil:
				fmt.Printf("  .word 0\n")
			default:
				throw(val)
			}
		case T_SLICE:
			fmt.Printf("  .quad 0 # ptr\n")
			fmt.Printf("  .quad 0 # len\n")
			fmt.Printf("  .quad 0 # cap\n")
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
			throw(getTypeKind(varDecl.Type))
		}
	}
	fmt.Printf("# ==============================\n")
}

func emitText(pkgName string) {
	fmt.Printf(".text\n")
	for _, fnc := range globalFuncs {
		emitFuncDecl(pkgName, fnc)
	}
}

func generateCode(pkgName string) {
	emitData(pkgName)
	emitText(pkgName)
}

var pkgName string
var stringLiterals []string
var stringIndex int

var globalVars []*ast.ValueSpec
var globalFuncs []*Func

func main() {
	for _, file := range []string{"./runtime.go", "./t/a.go"} {
		globalVars = nil
		globalFuncs = nil
		stringLiterals = nil
		stringIndex = 0
		fset := &token.FileSet{}
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			panic(err)
		}

		pkg := semanticAnalyze(fset, f)
		generateCode(pkg.Name())
	}
}
