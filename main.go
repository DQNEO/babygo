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
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
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
	case *ast.StarExpr:
		return 8
	}
	throw(typeExpr)
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

		emitExpr(e.Index, tInt) // index number
		typ := getTypeOfExpr(e.X)
		switch getTypeKind(typ) {
		case T_ARRAY:
			emitAddr(e.X) // array head
		case T_SLICE:
			emitExpr(e.X, typ)
			fmt.Printf("  popq %%rax # slice.ptr\n")
			fmt.Printf("  popq %%rcx # garbage\n")
			fmt.Printf("  popq %%rcx # garbage\n")
			fmt.Printf("  pushq %%rax # slice.ptr\n")
		case T_STRING:
			emitExpr(e.X, typ)
			fmt.Printf("  popq %%rax # string.ptr\n")
			fmt.Printf("  popq %%rcx # garbage\n")
			fmt.Printf("  pushq %%rax # string.ptr\n")
		default:
			panic(getTypeKind(getTypeOfExpr(e.X)))
		}
		fmt.Printf("  popq %%rax # collection addr\n")
		fmt.Printf("  popq %%rcx # index\n")
		fmt.Printf("  movq $%d, %%rdx # elm size\n", size)
		fmt.Printf("  imulq %%rdx, %%rcx\n")
		fmt.Printf("  addq %%rcx, %%rax\n")
		fmt.Printf("  pushq %%rax # addr of element\n")
	case *ast.StarExpr:
		emitExpr(e.X, nil)
	case *ast.SelectorExpr:// X.Sel
		typeOfX := getTypeOfExpr(e.X)
		var structType ast.Expr
		switch getTypeKind(typeOfX) {
		case T_STRUCT:
			// strct.field
			structType = typeOfX
			emitAddr(e.X)
		case T_POINTER:
			// ptr.field
			ptrType, ok := typeOfX.(*ast.StarExpr)
			assert(ok, "should be *ast.StarExpr")
			structType = ptrType.X
			emitExpr(e.X, nil)
		default:
			throw("TBI")
		}
		field := lookupStructField(structType, e.Sel.Name)
		offset := getStructFieldOffset(field)
		fmt.Printf("  popq %%rax # addr of struct head\n")
		fmt.Printf("  addq $%d, %%rax # add offset to \"%s\"\n", offset, e.Sel.Name)
		fmt.Printf("  pushq %%rax # addr of struct.field\n")
	default:
		throw(expr)
	}
}

// Conversion slice(string)
func emitConversionToSlice(arrayType *ast.ArrayType, arg0 ast.Expr) {
	assert(arrayType.Len == nil, "arrayType should be slice")
	assert(getTypeKind(getTypeOfExpr(arg0)) == T_STRING, "arrayType should be slice")
	fmt.Printf("  # Conversion to slice %s <= %s\n", arrayType.Elt, getTypeOfExpr(arg0))
	emitExpr(arg0, arrayType)
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
			emitExpr(arg0, fn) // slice
			fmt.Printf("  popq %%rax # ptr\n")
			fmt.Printf("  popq %%rcx # len\n")
			fmt.Printf("  popq %%rdx # cap (to be abandoned)\n")
			fmt.Printf("  pushq %%rcx # str len\n")
			fmt.Printf("  pushq %%rax # str ptr\n")
		}
	case gInt, gUint8, gUint16, gUintptr: // int(e)
		emitExpr(arg0, fn)
	default:
		throw(fn.Obj)
	}
	return
}

func getStructTypeOfX(e *ast.SelectorExpr) ast.Expr {
	typeOfX := getTypeOfExpr(e.X)
	var structType ast.Expr
	switch getTypeKind(typeOfX) {
	case T_STRUCT:
		// strct.field => e.X . e.Sel
		structType = typeOfX
	case T_POINTER:
		// ptr.field => e.X . e.Sel
		ptrType, ok := typeOfX.(*ast.StarExpr)
		assert(ok, "should be *ast.StarExpr")
		structType = ptrType.X
	default:
		throw("TBI")
	}
	return structType
}

func emitZeroValue(typeExpr ast.Expr) {
	switch getTypeKind(typeExpr) {
	case T_SLICE:
		fmt.Printf("  pushq $0 # slice.cap\n")
		fmt.Printf("  pushq $0 # slice.len\n")
		fmt.Printf("  pushq $0 # slice.ptr\n")
	default:
		throw(typeExpr)
	}
}

func emitExpr(expr ast.Expr, forceType ast.Expr) {
	switch e := expr.(type) {
	case *ast.Ident:
		switch e.Obj {
		case gTrue: // true constant
			fmt.Printf("  pushq $1 # true\n")
			return
		case gFalse: // false constant
			fmt.Printf("  pushq $0 # false\n")
			return
		case gNil:
			switch getTypeKind(forceType) {
			case T_SLICE:
				emitZeroValue(forceType)
			default:
				throw(forceType)
			}
			return
		}

		if e.Obj == nil {
			panic(fmt.Sprintf("ident %s is unresolved", e.Name))
		}

		if e.Obj.Kind != ast.Var {
			panic("Unexpected ident kind:" + e.Obj.Kind.String())
		} else {
		}
		emitAddr(e)
		emitLoad(getTypeOfExpr(e))
	case *ast.IndexExpr:
		emitAddr(e)
		emitLoad(getTypeOfExpr(e))
	case *ast.StarExpr:
		emitAddr(e)
		emitLoad(getTypeOfExpr(e))
	case *ast.SelectorExpr: // X.Sel
		fmt.Printf("  # emitExpr *ast.SelectorExpr %s.%s\n", e.X, e.Sel)
		emitAddr(e)
		emitLoad(getTypeOfExpr(e))
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
			case gLen:
				assert(len(e.Args) == 1, "builtin len should take only 1 args")
				var arg ast.Expr = e.Args[0]
				switch getTypeKind(getTypeOfExpr(arg)) {
				case T_ARRAY:
					arrayType, ok := getTypeOfExpr(arg).(*ast.ArrayType)
					assert(ok, "should be *ast.ArrayType")
					emitExpr(arrayType.Len, tInt)
				case T_SLICE:
					emitExpr(arg, nil)
					fmt.Printf("  popq %%rax # throw away ptr\n")
					fmt.Printf("  popq %%rcx # len\n")
					fmt.Printf("  popq %%rax # throw away cap\n")
					fmt.Printf("  pushq %%rcx # len\n")
				case T_STRING:
					emitExpr(arg, nil)
					fmt.Printf("  popq %%rax # throw away ptr\n")
					fmt.Printf("  popq %%rcx # len\n")
					fmt.Printf("  pushq %%rcx # len\n")
				default:
					throw(getTypeKind(getTypeOfExpr(arg)))
				}
			case gCap:
				assert(len(e.Args) == 1, "builtin len should take only 1 args")
				var arg ast.Expr = e.Args[0]
				switch getTypeKind(getTypeOfExpr(arg)) {
				case T_ARRAY:
					arrayType, ok := getTypeOfExpr(arg).(*ast.ArrayType)
					assert(ok, "should be *ast.ArrayType")
					emitExpr(arrayType.Len, tInt)
				case T_SLICE:
					emitExpr(arg, nil)
					fmt.Printf("  popq %%rax # throw away ptr\n")
					fmt.Printf("  popq %%rcx # len\n")
					fmt.Printf("  popq %%rax # throw away cap\n")
					fmt.Printf("  pushq %%rax # cap\n")
				case T_STRING:
					panic("cap() cannot accept string type")
				default:
					throw(getTypeKind(getTypeOfExpr(arg)))
				}
			case gMake:
				var typeArg ast.Expr = e.Args[0]
				switch getTypeKind(typeArg) {
				case T_SLICE:
					// make([]T, ...)
					arrayType, ok := typeArg.(*ast.ArrayType)
					assert(ok, "should be *ast.ArrayType")

					var lenArg ast.Expr = e.Args[1]
					var capArg ast.Expr = e.Args[2]
					emitExpr(capArg, typeArg)
					emitExpr(lenArg, typeArg)
					elmSize := getSizeOfType(arrayType.Elt)
					fmt.Printf("  pushq $%d # elm size\n", elmSize)
					symbol := "runtime.makeSlice"
					fmt.Printf("  callq %s\n", symbol)
					fmt.Printf("  addq $24, %%rsp # revert for one string\n")
					fmt.Printf("  pushq %%rsi # slice cap\n")
					fmt.Printf("  pushq %%rdi # slice len\n")
					fmt.Printf("  pushq %%rax # slice ptr\n")
				default:
					throw(typeArg)
				}
			default:
				switch fn.Obj.Kind {
				case ast.Typ:
					// Conversion
					emitConversion(fn, e.Args[0])
					return
				}
				if fn.Name == "print" {
					// builtin print
					emitExpr(e.Args[0], nil) // push ptr, push len
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
						emitExpr(arg, nil)
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

			}
		case *ast.SelectorExpr:
			symbol := fmt.Sprintf("%s.%s", fn.X, fn.Sel) // syscall.Write() or unsafe.Pointer(x)
			switch symbol {
			case "unsafe.Pointer":
				emitExpr(e.Args[0], nil)
			case "syscall.Write":
				emitExpr(e.Args[1], nil)
				emitExpr(e.Args[0], nil)
				fmt.Printf("  callq %s\n", symbol) // func decl is in runtime
			default:
				panic(symbol)
			}
		default:
			throw(fun)
		}
	case *ast.ParenExpr:
		emitExpr(e.X, getTypeOfExpr(e))
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
			emitExpr(e.X, nil)
			fmt.Printf("  popq %%rax # e.X\n")
			fmt.Printf("  imulq $-1, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "&":
			emitAddr(e.X)
		default:
			throw(e.Op.String())
		}
	case *ast.BinaryExpr:
		if getTypeKind(getTypeOfExpr(e.X)) == T_STRING {
			emitConcateString(e.X, e.Y)
			return
		}
		fmt.Printf("  # start %T\n", e)
		emitExpr(e.X, nil) // left
		emitExpr(e.Y, nil) // right
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
	case *ast.SliceExpr:
		//e.Index, e.X
		emitAddr(e.X) // array head
		emitExpr(e.Low, nil) // intval
		emitExpr(e.High, nil) // intval
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

func emitConcateString(left ast.Expr, right ast.Expr) {
	emitExpr(right, nil)
	emitExpr(left, nil)
	fmt.Printf("  callq runtime.catstrings\n")
	fmt.Printf("  addq $32, %%rsp # revert for one string\n")
	fmt.Printf("  pushq %%rdi # slice len\n")
	fmt.Printf("  pushq %%rax # slice ptr\n")
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
	fmt.Printf("  # emitStore(%s)\n", getTypeKind(typ))
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
	case T_INT,T_BOOL, T_UINTPTR, T_POINTER:
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

func emitAssign(lhs ast.Expr, rhs ast.Expr) {
	fmt.Printf("  # Assignment: emitAddr(lhs)\n")
	emitAddr(lhs)
	fmt.Printf("  # Assignment: emitExpr(rhs)\n")
	emitExpr(rhs, getTypeOfExpr(lhs))
	fmt.Printf("  # Assignment: emitStore(getTypeOfExpr(lhs))\n")
	emitStore(getTypeOfExpr(lhs))
}

func emitStmt(stmt ast.Stmt) {
	fmt.Printf("  \n")
	fmt.Printf("  # == Stmt %T ==\n", stmt)
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		expr := s.X
		emitExpr(expr, nil)
	case *ast.DeclStmt:
		decl := s.Decl
		switch dcl := decl.(type) {
		case *ast.GenDecl:
			declSpec := dcl.Specs[0]
			switch ds := declSpec.(type) {
			case *ast.ValueSpec:
				fmt.Printf("  # Decl.Specs[0]: Names[0]=%#v, Type=%#v\n", ds.Names[0], ds.Type)
				varSpec := ds
				if len(varSpec.Values) == 0 {
					fmt.Printf("  # Init with zero value\n")
				} else if len(varSpec.Values) == 1 {
					// assignment
					lhs := varSpec.Names[0]
					rhs := varSpec.Values[0]
					emitAssign(lhs, rhs)
				} else {
					panic("TBI")
				}
			default:
				throw(declSpec)
			}
		default:
			return
		}
		return // do nothing
	case *ast.AssignStmt:
		switch s.Tok.String() {
		case "=":
			lhs := s.Lhs[0]
			rhs := s.Rhs[0]
			emitAssign(lhs, rhs)
		default:
			panic("TBI: assignment of " + s.Tok.String())
		}
	case *ast.ReturnStmt:
		if len(s.Results) == 1 {
			emitExpr(s.Results[0], nil) // @FIXME
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
			emitExpr(s.Cond, nil)
			fmt.Printf("  popq %%rax\n")
			fmt.Printf("  cmpq $0, %%rax\n")
			fmt.Printf("  je %s # jmp if false\n", labelElse)
			emitStmt(s.Body) // then
			fmt.Printf("  jmp %s\n", labelEndif)
			fmt.Printf("  %s:\n", labelElse)
			emitStmt(s.Else) // then
		} else {
			emitExpr(s.Cond, nil)
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
		emitExpr(s.Cond, nil)
		fmt.Printf("  popq %%rax\n")
		fmt.Printf("  cmpq $0, %%rax\n")
		fmt.Printf("  je %s # jmp if false\n", labelExit)
		emitStmt(s.Body)
		emitStmt(s.Post)
		fmt.Printf("  jmp %s\n", labelCond)
		fmt.Printf("  %s:\n", labelExit)
	case *ast.IncDecStmt:
		var inst string
		switch s.Tok.String() {
		case "++":
			inst = "addq"
		case "--":
			inst = "subq"
		default:
			throw(s.Tok.String())
		}

		emitAddr(s.X)
		emitExpr(s.X, nil)
		fmt.Printf("  popq %%rax\n")
		fmt.Printf("  %s $1, %%rax\n", inst)
		fmt.Printf("  pushq %%rax\n")
		emitStore(getTypeOfExpr(s.X))
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


func getStructFieldOffset(field *ast.Field) int {
	text := field.Doc.List[0].Text
	offset, err := strconv.Atoi(text)
	if err != nil {
		panic(text)
	}
	return offset
}

func setStructFieldOffset(field *ast.Field, offset int) {
	comment := &ast.Comment{
		Text:  strconv.Itoa(offset),
	}
	commentGroup := &ast.CommentGroup{
		List: []*ast.Comment{comment},
	}
	field.Doc = commentGroup
}

func getStructFields(namedStructType ast.Expr) []*ast.Field {
	if getTypeKind(namedStructType) != T_STRUCT {
			throw(namedStructType)
	}
	ident, ok := namedStructType.(*ast.Ident)
	if !ok {
		throw(namedStructType)
	}
	typeSpec,ok := ident.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		throw(ident.Obj.Decl)
	}
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		throw(typeSpec.Type)
	}

	return structType.Fields.List
}

func lookupStructField(namedStructType ast.Expr, selName string) *ast.Field {
	for _, field := range getStructFields(namedStructType) {
		if field.Names[0].Name == selName {
			return field
		}
	}
	panic("Unexpected flow")
	return nil
}

func calcStructTypeSize(namedStructType ast.Expr) int {
	var offset int = 0
	for _, field := range getStructFields(namedStructType) {
		setStructFieldOffset(field, offset)
		size := getSizeOfType(field.Type)
		offset += size
	}
	return offset
}

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
				case T_INT, T_UINTPTR, T_POINTER:
					varSize = gInt.Data.(int)
				case T_UINT8:
					varSize = gUint8.Data.(int)
				case T_BOOL:
					varSize = gInt.Data.(int)
				case T_STRUCT:
					varSize = calcStructTypeSize(varSpec.Type)
				default:
					throw(varSpec.Type)
				}

				localoffset -= varSize
				setObjData(obj, localoffset)
				localvars = append(localvars, ds)
				for _, v := range ds.Values {
					walkExpr(v)
				}
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
	case *ast.IncDecStmt:
		walkExpr(s.X)
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
	case *ast.StarExpr:
		walkExpr(e.X)
	default:
		throw(expr)
	}
}

const sliceSize = 24

var gTrue = &ast.Object{
	Kind: ast.Con,
	Name: "true",
	Decl: nil,
	Data: nil,
	Type: nil,
}

var gNil = &ast.Object{
	Kind: ast.Con, // is nil a constant ?
	Name: "nil",
	Decl: nil,
	Data: nil,
	Type: nil,
}

var gFalse = &ast.Object{
	Kind: ast.Con,
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

var gLen = &ast.Object{
	Kind: ast.Fun,
	Name: "len",
	Decl: nil,
	Data: nil,
	Type: nil,
}

var gCap = &ast.Object{
	Kind: ast.Fun,
	Name: "cap",
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

	// predeclared constants
	universe.Insert(gTrue)
	universe.Insert(gFalse)

	universe.Insert(gNil)

	// predeclared funcs
	universe.Insert(gMake)
	universe.Insert(gLen)
	universe.Insert(gCap)

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
	//fmt.Printf("Universer:    %#v\n", types.Universe)
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
	fmt.Printf("# Unresolved: %#v\n", unresolved)
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
const T_STRUCT = "T_STRUCT"
const T_POINTER = "T_POINTER"

var tInt *ast.Ident = &ast.Ident{
	NamePos: 0,
	Name:    "int",
	Obj:     gInt,
}

var tString *ast.Ident = &ast.Ident{
	NamePos: 0,
	Name:    "string",
	Obj:     gString,
}

func getTypeOfExpr(expr ast.Expr) ast.Expr {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Obj.Kind == ast.Typ {
			panic("expression expected, but got Type")
		}
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
			return tInt
		case "CHAR":
			return tInt
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
			if getTypeKind(typ) == T_STRING {
				return  &ast.Ident{
					NamePos: 0,
					Name:    "uint8",
					Obj:     gUint8,
				}
			} else {
				panic(fmt.Sprintf("Unexpected expr type:%#v", typ))
			}
		}
	case *ast.CallExpr: // funcall or conversion
		switch fn := e.Fun.(type) {
		case *ast.Ident:
			if fn.Obj == nil {
				throw(fn)
			}
			switch fn.Obj.Kind {
			case ast.Typ: // conversion
				return fn
			case ast.Fun:
				switch fn.Obj {
				case gLen, gCap:
					return tInt
				}
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
	case *ast.StarExpr:
		typ := getTypeOfExpr(e.X)
		ptrType, ok := typ.(*ast.StarExpr)
		if !ok {
			throw(typ)
		}
		return ptrType.X
	case *ast.SelectorExpr:
		fmt.Printf("  # getTypeOfExpr(%s.%s)\n", e.X, e.Sel)
		structType := getStructTypeOfX(e)
		field := lookupStructField(structType, e.Sel.Name)
		return field.Type
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
				// named type
				decl := e.Obj.Decl
				typeSpec, ok := decl.(*ast.TypeSpec)
				if !ok {
					throw(decl)
				}
				return getTypeKind(typeSpec.Type)
			}
		}
	case *ast.StructType:
		return T_STRUCT
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
			// emit global zero values
			if val != nil {
				panic("TBI")
			}
			arrayType,ok :=  varDecl.Type.(*ast.ArrayType)
			assert(ok, "should be *ast.ArrayType")
			assert(arrayType.Len != nil, "slice type is not expected")
			basicLit, ok := arrayType.Len.(*ast.BasicLit)
			assert(ok, "should be *ast.BasicLit")
			length, err := strconv.Atoi(basicLit.Value)
			if err != nil {
				panic(err)
			}
			var zeroValue string
			switch getTypeKind(arrayType.Elt) {
			case T_INT:
				zeroValue = fmt.Sprintf("  .quad 0 # int\n")
			case T_UINT8:
				zeroValue = fmt.Sprintf("  .byte 0 # uint8\n")
			case T_STRING:
				zeroValue = fmt.Sprintf("  .quad 0 # str.ptr\n")
				zeroValue += fmt.Sprintf("  .quad 0 # str.len\n")
			default:
				throw(arrayType.Elt)
			}
			for i:=0;i<length;i++ {
				fmt.Printf(zeroValue)
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

var sourceFiles [2]string
func main() {
	sourceFiles[0] = "./runtime.go"
	sourceFiles[1] = "./t/source.go"
	var i int
	for i=0;i<len(sourceFiles); i++ {
		globalVars = nil
		globalFuncs = nil
		stringLiterals = nil
		stringIndex = 0
		fset := &token.FileSet{}
		f, err := parser.ParseFile(fset, sourceFiles[i], nil, 0)
		if err != nil {
			panic(err)
		}

		pkg := semanticAnalyze(fset, f)
		generateCode(pkg.Name())
	}
}
