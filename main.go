package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strconv"
)

type Type struct {
	e ast.Expr // original expr
}

type localoffsetint int

func emitLoad(t *Type) {
	fmt.Printf("  popq %%rdx\n")
	switch kind(t) {
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
	case T_ARRAY:
		// pure proxy
		fmt.Printf("  pushq %%rdx\n")
	default:
		throw(t)
	}
}

func emitVariableAddr(variable *Variable) {
	fmt.Printf("  # variable %#v\n", variable)

	var addr string
	if variable.isGlobal {
		addr = fmt.Sprintf("%s(%%rip)", variable.globalSymbol)
	} else {
		addr = fmt.Sprintf("%d(%%rbp)", variable.localOffset)
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

func evalInt(expr ast.Expr) int {
	switch e := expr.(type) {
	case *ast.BasicLit:
		i, err := strconv.Atoi(e.Value)
		if err != nil {
			panic(err)
		}
		return i
	}
	return 0
}

func getSizeOfType(t *Type) int {
	var varSize int
	switch kind(t) {
	case T_SLICE:
		return sliceSize
	case T_STRING:
		return gString.Data.(int)
	case T_INT, T_UINTPTR, T_POINTER:
		return gInt.Data.(int)
	case T_UINT8:
		return gUint8.Data.(int)
	case T_BOOL:
		return gInt.Data.(int)
	case T_ARRAY:
		arrayType, ok := t.e.(*ast.ArrayType)
		assert(ok, "expect *ast.ArrayType")
		elmSize := getSizeOfType(e2t(arrayType.Elt))
		return elmSize * evalInt(arrayType.Len)
	case T_STRUCT:
		return calcStructSizeAndSetFieldOffset(getStructTypeSpec(t))
	default:
		throw(t)
	}
	return varSize
}

type Variable struct {
	name         string
	isGlobal     bool
	globalSymbol string
	localOffset  localoffsetint
}

func emitListHeadAddr(list ast.Expr) {
	t := getTypeOfExpr(list)
	switch kind(t) {
	case T_ARRAY:
		emitAddr(list) // array head
	case T_SLICE:
		emitExpr(list, t)
		fmt.Printf("  popq %%rax # slice.ptr\n")
		fmt.Printf("  popq %%rcx # garbage\n")
		fmt.Printf("  popq %%rcx # garbage\n")
		fmt.Printf("  pushq %%rax # slice.ptr\n")
	case T_STRING:
		emitExpr(list, t)
		fmt.Printf("  popq %%rax # string.ptr\n")
		fmt.Printf("  popq %%rcx # garbage\n")
		fmt.Printf("  pushq %%rax # string.ptr\n")
	default:
		panic(kind(getTypeOfExpr(list)))
	}

}

func emitAddr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Obj.Kind == ast.Var {
			vr, ok := e.Obj.Data.(*Variable)
			assert(ok, "should be *Variable")
			emitVariableAddr(vr)
		} else {
			panic("Unexpected ident kind")
		}
	case *ast.IndexExpr:
		emitExpr(e.Index, tInt) // index number

		list := e.X
		elmType := getTypeOfExpr(e)
		emitListElementAddr(list, elmType)
	case *ast.StarExpr:
		emitExpr(e.X, nil)
	case *ast.SelectorExpr: // X.Sel
		typeOfX := getTypeOfExpr(e.X)
		var structType *Type
		switch kind(typeOfX) {
		case T_STRUCT:
			// strct.field
			structType = typeOfX
			emitAddr(e.X)
		case T_POINTER:
			// ptr.field
			ptrType, ok := typeOfX.e.(*ast.StarExpr)
			assert(ok, "should be *ast.StarExpr")
			structType = e2t(ptrType.X)
			emitExpr(e.X, nil)
		default:
			throw("TBI")
		}
		field := lookupStructField(getStructTypeSpec(structType), e.Sel.Name)
		offset := getStructFieldOffset(field)
		fmt.Printf("  popq %%rax # addr of struct head\n")
		fmt.Printf("  addq $%d, %%rax # add offset to \"%s\"\n", offset, e.Sel.Name)
		fmt.Printf("  pushq %%rax # addr of struct.field\n")
	default:
		throw(expr)
	}
}

func emitConversion(tp *Type, arg0 ast.Expr) {
	fmt.Printf("  # Conversion %s <= %s\n", tp.e, getTypeOfExpr(arg0))
	switch typeExpr := tp.e.(type) {
	case *ast.Ident:
		ident := typeExpr
		switch ident.Obj {
		case gString: // string(e)
			switch kind(getTypeOfExpr(arg0)) {
			case T_SLICE: // string(slice)
				emitExpr(arg0, e2t(ident)) // slice
				fmt.Printf("  popq %%rax # ptr\n")
				fmt.Printf("  popq %%rcx # len\n")
				fmt.Printf("  popq %%rdx # cap (to be abandoned)\n")
				fmt.Printf("  pushq %%rcx # str len\n")
				fmt.Printf("  pushq %%rax # str ptr\n")
			}
		case gInt, gUint8, gUint16, gUintptr: // int(e)
			emitExpr(arg0, e2t(ident))
		default:
			throw(ident.Obj)
		}
	case *ast.ArrayType: // Conversion to slice
		arrayType := typeExpr
		if arrayType.Len == nil {
			assert(kind(getTypeOfExpr(arg0)) == T_STRING, "arrayType should be slice")
			fmt.Printf("  # Conversion to slice %s <= %s\n", arrayType.Elt, getTypeOfExpr(arg0))
			emitExpr(arg0, tp)
			fmt.Printf("  popq %%rax # ptr\n")
			fmt.Printf("  popq %%rcx # len\n")
			fmt.Printf("  pushq %%rcx # cap\n")
			fmt.Printf("  pushq %%rcx # len\n")
			fmt.Printf("  pushq %%rax # ptr\n")
		} else {
			throw(typeExpr)
		}
	case *ast.ParenExpr: // (T)(arg0)
		emitConversion(e2t(typeExpr.X), arg0)
	case *ast.StarExpr: // (*T)(arg0)
		// go through
		emitExpr(arg0, nil)
	default:
		throw(typeExpr)
	}
	return
}

func getStructTypeOfX(e *ast.SelectorExpr) *Type {
	typeOfX := getTypeOfExpr(e.X)
	var structType *Type
	switch kind(typeOfX) {
	case T_STRUCT:
		// strct.field => e.X . e.Sel
		structType = typeOfX
	case T_POINTER:
		// ptr.field => e.X . e.Sel
		ptrType, ok := typeOfX.e.(*ast.StarExpr)
		assert(ok, "should be *ast.StarExpr")
		structType = e2t(ptrType.X)
	default:
		throw("TBI")
	}
	return structType
}

func emitZeroValue(t *Type) {
	switch kind(t) {
	case T_SLICE:
		fmt.Printf("  pushq $0 # slice zero value\n")
		fmt.Printf("  pushq $0 # slice zero value\n")
		fmt.Printf("  pushq $0 # slice zero valuer\n")
	case T_STRING:
		fmt.Printf("  pushq $0 # string zero value\n")
		fmt.Printf("  pushq $0 # string zero value\n")
	case T_INT, T_UINTPTR, T_UINT8, T_POINTER, T_BOOL:
		fmt.Printf("  pushq $0 # %s zero value\n", kind(t))
	case T_STRUCT:
		//@FIXME
	default:
		throw(t)
	}
}

func emitLen(arg ast.Expr) {
	switch kind(getTypeOfExpr(arg)) {
	case T_ARRAY:
		arrayType, ok := getTypeOfExpr(arg).e.(*ast.ArrayType)
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
		throw(kind(getTypeOfExpr(arg)))
	}
}

func emitCallMalloc(size int) {
	fmt.Printf("  pushq $%d\n", size)
	// call malloc and return pointer
	fmt.Printf("  callq runtime.malloc\n")
	fmt.Printf("  addq $8, %%rsp # revert stack pointer\n")
	fmt.Printf("  pushq %%rax # addr\n")
}

func emitArrayLiteral(arrayType *ast.ArrayType, arrayLen int, elts []ast.Expr) {
	elmType := e2t(arrayType.Elt)
	elmSize := getSizeOfType(elmType)
	memSize := elmSize * arrayLen
	emitCallMalloc(memSize) // push
	for i, elm := range elts {
		// emit lhs
		fmt.Printf("  popq %%rax # addr\n")
		fmt.Printf("  pushq %%rax # backup malloced addr\n")
		fmt.Printf("  addq $%d, %%rax # add offset\n", elmSize * i)
		fmt.Printf("  pushq %%rax # elm addr\n")
		emitExpr(elm, elmType)
		emitStore(elmType)
	}
}

func emitExpr(expr ast.Expr, forceType *Type) {
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
			if forceType == nil {
				panic("Type is required for nil")
			}
			switch kind(forceType) {
			case T_SLICE:
				emitZeroValue(forceType)
			case T_POINTER:
				emitZeroValue(forceType)
			default:
				throw(kind(forceType))
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
		case *ast.ArrayType: // Conversion
			emitConversion(e2t(fn), e.Args[0])
		case *ast.Ident:
			if fn.Obj == nil {
				panic("unresolved ident: " + fn.String())
			}
			if fn.Obj.Kind == ast.Typ {
				// Conversion
				emitConversion(e2t(fn), e.Args[0])
				return
			}
			switch fn.Obj {
			case gLen:
				assert(len(e.Args) == 1, "builtin len should take only 1 args")
				var arg ast.Expr = e.Args[0]
				emitLen(arg)
			case gCap:
				assert(len(e.Args) == 1, "builtin len should take only 1 args")
				var arg ast.Expr = e.Args[0]
				switch kind(getTypeOfExpr(arg)) {
				case T_ARRAY:
					arrayType, ok := getTypeOfExpr(arg).e.(*ast.ArrayType)
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
					throw(kind(getTypeOfExpr(arg)))
				}
			case gNew:
				typeArg := e2t(e.Args[0])
				// size to malloc
				size := getSizeOfType(typeArg)
				emitCallMalloc(size)
			case gMake:
				var typeArg = e2t(e.Args[0])
				switch kind(typeArg) {
				case T_SLICE:
					// make([]T, ...)
					arrayType, ok := typeArg.e.(*ast.ArrayType)
					assert(ok, "should be *ast.ArrayType")

					var lenArg ast.Expr = e.Args[1]
					var capArg ast.Expr = e.Args[2]
					emitExpr(capArg, typeArg)
					emitExpr(lenArg, typeArg)
					elmSize := getSizeOfType(e2t(arrayType.Elt))
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
			case gAppend:
				var sliceArg ast.Expr = e.Args[0]
				var elemArg ast.Expr = e.Args[1]
				emitExpr(elemArg, nil)  // various size
				emitExpr(sliceArg, nil) // size 24
				var stackForElm int
				var symbol string
				var elmSize int = getSizeOfType(getElementTypeOfListType(getTypeOfExpr(sliceArg)))
				switch elmSize {
				case 1:
					symbol = "runtime.append1"
					stackForElm = 8
				case 8:
					symbol = "runtime.append8"
					stackForElm = 8
				case 16:
					symbol = "runtime.append16"
					stackForElm = 16
				case 24:
					symbol = "runtime.append24"
					stackForElm = 24
				default:
					throw(elmSize)
				}

				fmt.Printf("  callq %s\n", symbol)
				fmt.Printf("  addq $24+%d, %%rsp # revert\n", stackForElm)
				fmt.Printf("  pushq %%rsi # slice cap\n")
				fmt.Printf("  pushq %%rdi # slice len\n")
				fmt.Printf("  pushq %%rax # slice ptr\n")
			default:
				if fn.Name == "print" {
					// builtin print
					emitExpr(e.Args[0], nil) // push ptr, push len
					switch kind(getTypeOfExpr(e.Args[0])) {
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
					if fn.Name == "makeSlice1" || fn.Name == "makeSlice8" || fn.Name == "makeSlice16" || fn.Name == "makeSlice24" {
						fn.Name = "makeSlice"
					}
					// general funcall
					var totalSize int = 0
					for i := len(e.Args) - 1; i >= 0; i-- {
						arg := e.Args[i]
						emitExpr(arg, nil)
						size := getSizeOfType(getTypeOfExpr(arg))
						totalSize += size
					}
					symbol := pkgName + "." + fn.Name
					fmt.Printf("  callq %s\n", symbol)
					fmt.Printf("  addq $%d, %%rsp # revert\n", totalSize)

					obj := fn.Obj //.Kind == FN
					fndecl, ok := obj.Decl.(*ast.FuncDecl)
					if !ok {
						throw(fn.Obj)
					}
					if fndecl.Type.Results != nil {
						if len(fndecl.Type.Results.List) > 2 {
							panic("TBI")
						} else if len(fndecl.Type.Results.List) == 1 {
							retval0 := fndecl.Type.Results.List[0]
							switch kind(e2t(retval0.Type)) {
							case T_STRING:
								fmt.Printf("  # fn.Obj=%#v\n", obj)
								fmt.Printf("  pushq %%rdi # str len\n")
								fmt.Printf("  pushq %%rax # str ptr\n")
							case T_BOOL, T_INT, T_UINTPTR, T_POINTER:
								fmt.Printf("  # fn.Obj=%#v\n", obj)
								fmt.Printf("  pushq %%rax\n")
							case T_SLICE:
								fmt.Printf("  pushq %%rsi # slice cap\n")
								fmt.Printf("  pushq %%rdi # slice len\n")
								fmt.Printf("  pushq %%rax # slice ptr\n")
							default:
								throw(kind(e2t(retval0.Type)))
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
				// @FIXME revert rsp
			default:
				panic(symbol)
			}
		case *ast.ParenExpr: // (T)(e)
			// we assume this is conversion
			emitConversion(e2t(fn), e.Args[0])
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
			sl := getStringLiteral(e)
			if sl.strlen == 0 {
				// zero value
				emitZeroValue(tString)
			} else {
				fmt.Printf("  pushq $%d # str len\n", sl.strlen)
				fmt.Printf("  leaq %s, %%rax # str ptr\n", sl.label)
				fmt.Printf("  pushq %%rax # str ptr\n")
			}
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
		case "!":
			emitExpr(e.X, nil)
			fmt.Printf("  popq %%rax # e.X\n")
			fmt.Printf("  not %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		default:
			throw(e)
		}
	case *ast.BinaryExpr:
		if kind(getTypeOfExpr(e.X)) == T_STRING {
			switch e.Op.String() {
			case "+":
				emitExpr(e.Y, nil)
				emitExpr(e.X, nil)
				fmt.Printf("  callq runtime.catstrings\n")
				fmt.Printf("  addq $32, %%rsp # revert for one string\n")
				fmt.Printf("  pushq %%rdi # slice len\n")
				fmt.Printf("  pushq %%rax # slice ptr\n")
			case "==":
				panic("TBI")
			default:
				throw(e.Op.String())
			}
			return
		}
		fmt.Printf("  # start %T\n", e)
		emitExpr(e.X, nil) // left
		emitExpr(e.Y, getTypeOfExpr(e.X)) // right
		switch e.Op.String() {
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
		// slice , array, map or struct
		switch kind(e2t(e.Type)) {
		case T_ARRAY:
			arrayType, ok := e.Type.(*ast.ArrayType)
			assert(ok, "expect *ast.ArrayType")
			arrayLen := evalInt(arrayType.Len)
			emitArrayLiteral(arrayType, arrayLen, e.Elts)
		case T_SLICE:
			arrayType, ok := e.Type.(*ast.ArrayType)
			assert(ok, "expect *ast.ArrayType")
			length := len(e.Elts)
			emitArrayLiteral(arrayType, length, e.Elts)
			fmt.Printf("  popq  %%rax # malloced addr\n")
			fmt.Printf("  pushq $%d # slice.cap\n", length)
			fmt.Printf("  pushq $%d # slice.len\n", length)
			fmt.Printf("  pushq %%rax # slice.ptr\n")
		default:
			throw(e.Type)
		}
	case *ast.SliceExpr: // list[low:high]
		list := e.X
		listType := getTypeOfExpr(list)
		emitExpr(e.Low, tInt)  // intval
		emitExpr(e.High, tInt) // intval
		//emitExpr(e.Max) // @TODO
		fmt.Printf("  popq %%rax # high\n")
		fmt.Printf("  popq %%rcx # low\n")
		fmt.Printf("  subq %%rcx, %%rax # high - low\n")
		switch kind(listType) {
		case T_SLICE, T_ARRAY:
			fmt.Printf("  pushq %%rax # cap\n")
			fmt.Printf("  pushq %%rax # len\n")
		case T_STRING:
			fmt.Printf("  pushq %%rax # len\n")
			// no cap
		default:
			throw(list)
		}

		emitExpr(e.Low, tInt) // index number
		elmType := getElementTypeOfListType(listType)
		emitListElementAddr(list, elmType)
	default:
		throw(expr)
	}
}

func emitListElementAddr(list ast.Expr, elmType *Type) {
	emitListHeadAddr(list)
	fmt.Printf("  popq %%rax # list addr\n")
	fmt.Printf("  popq %%rcx # index\n")
	fmt.Printf("  movq $%d, %%rdx # elm size %s\n", getSizeOfType(elmType), elmType)
	fmt.Printf("  imulq %%rdx, %%rcx\n")
	fmt.Printf("  addq %%rcx, %%rax\n")
	fmt.Printf("  pushq %%rax # addr of element\n")
}

func getElementTypeOfListType(t *Type) *Type {
	switch kind(t) {
	case T_SLICE, T_ARRAY:
		arrayType, ok := t.e.(*ast.ArrayType)
		assert(ok, "expect *ast.ArrayType")
		return e2t(arrayType.Elt)
	case T_STRING:
		return tUint8
	default:
		throw(t)
	}
	return nil
}

func emitConcateString(left ast.Expr, right ast.Expr) {

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

func emitStore(t *Type) {
	fmt.Printf("  # emitStore(%s)\n", kind(t))
	switch kind(t) {
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
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
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
	case T_STRUCT:
		// @FXIME
	case T_ARRAY:
		fmt.Printf("  popq %%rdi # rhs: addr of data\n")
		fmt.Printf("  popq %%rax # lhs: addr to store\n")
		fmt.Printf("  pushq $%d # size\n", getSizeOfType(t))
		fmt.Printf("  pushq %%rax # dst lhs\n")
		fmt.Printf("  pushq %%rdi # src rhs\n")
		fmt.Printf("  callq runtime.memcopy\n")
		fmt.Printf("  addq $24, %%rsp\n # rewind stack pointer")
	default:
		panic("TBI:" + kind(t))
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
				t := e2t(ds.Type)
				fmt.Printf("  # Decl.Specs[0]: Names[0]=%#v, Type=%#v\n", ds.Names[0], t.e)
				valSpec := ds
				lhs := valSpec.Names[0]
				var rhs ast.Expr
				if len(valSpec.Values) == 0 {
					fmt.Printf("  # lhs addresss\n")
					emitAddr(lhs)
					fmt.Printf("  # emitZeroValue\n")
					emitZeroValue(t)
					fmt.Printf("  # Assignment: zero value\n")
					emitStore(t)
				} else if len(valSpec.Values) == 1 {
					// assignment
					rhs = valSpec.Values[0]
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
			switch kind(getTypeOfExpr(s.Results[0])) {
			case T_BOOL, T_INT, T_UINTPTR, T_POINTER:
				fmt.Printf("  popq %%rax # return 64bit\n")
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
		} else if len(s.Results) == 3 {
			// Special treatment to return a slice
			for _, result := range s.Results {
				assert(getSizeOfType(getTypeOfExpr(result)) == 8, "TBI")
			}
			emitExpr(s.Results[2], nil) // @FIXME
			emitExpr(s.Results[1], nil) // @FIXME
			emitExpr(s.Results[0], nil) // @FIXME
			fmt.Printf("  popq %%rax # return 64bit\n")
			fmt.Printf("  popq %%rdi # return 64bit\n")
			fmt.Printf("  popq %%rsi # return 64bit\n")

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
	case *ast.RangeStmt: // only for array and slice
		labelid++
		labelCond := fmt.Sprintf(".L.loop.cond.%d", labelid)
		labelEndFor := fmt.Sprintf(".L.loop.exit.%d", labelid)

		// initialization: store len(rangeexpr)
		fmt.Printf("  # ForRange Initialization\n")

		fmt.Printf("  #   assign to lenvar\n")
		// emit len of rangeexpr
		rngMisc, ok := mapRangeStmt[s]
		assert(ok, "lenVar should exist")
		// lenvar = len(s.X)
		emitVariableAddr(rngMisc.lenvar)
		emitLen(s.X)
		emitStore(tInt)

		fmt.Printf("  #   assign to indexvar\n")
		// indexvar = 0
		emitVariableAddr(rngMisc.indexvar)
		emitZeroValue(tInt)
		emitStore(tInt)

		// Condition
		// if (indexvar < lenvar) then
		//   execute body
		// else
		//   exit
		fmt.Printf("  # ForRange Condition\n")
		fmt.Printf("  %s:\n", labelCond)

		emitVariableAddr(rngMisc.indexvar)
		emitLoad(tInt)
		emitVariableAddr(rngMisc.lenvar)
		emitLoad(tInt)
		emitCompExpr("setl")
		fmt.Printf("  popq %%rax\n")
		fmt.Printf("  cmpq $0, %%rax\n")
		fmt.Printf("  je %s # jmp if false\n", labelEndFor)

		fmt.Printf("  # assign list[indexvar] value variables\n")
		elemType := getTypeOfExpr(s.Value)
		emitAddr(s.Value) // lhs

		emitVariableAddr(rngMisc.indexvar)
		emitLoad(tInt) // index value
		emitListElementAddr(s.X, elemType)

		emitLoad(elemType)
		emitStore(elemType)

		// Body
		fmt.Printf("  # ForRange Body\n")
		emitStmt(s.Body)

		// Post statement: Increment indexvar and go next
		fmt.Printf("  # ForRange Post statement\n")
		emitVariableAddr(rngMisc.indexvar) // lhs
		emitVariableAddr(rngMisc.indexvar) // rhs
		emitLoad(tInt)
		fmt.Printf("  popq %%rax # indexvar value\n")
		fmt.Printf("  addq $1, %%rax # ++\n")
		fmt.Printf("  pushq %%rax #\n")
		emitStore(tInt)

		fmt.Printf("  jmp %s\n", labelCond)

		fmt.Printf("  %s:\n", labelEndFor)
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
	fmt.Printf("%s.%s: # args %d, locals %d\n",
		pkgPrefix, fnc.name, fnc.argsarea, fnc.localarea)
	fmt.Printf("  pushq %%rbp\n")
	fmt.Printf("  movq %%rsp, %%rbp\n")
	if fnc.localarea != 0 {
		fmt.Printf("  subq $%d, %%rsp # local area\n", -fnc.localarea)
	}
	for _, stmt := range fnc.stmts {
		emitStmt(stmt)
	}
	fmt.Printf("  leave\n")
	fmt.Printf("  ret\n")
}

type sliteral struct {
	label  string
	strlen int
	value  string // raw value
}

func getStringLiteral(lit *ast.BasicLit) *sliteral {
	sl, ok := mapStringLiterals[lit]
	if !ok {
		panic("StringLiteral is not registered")
	}

	return sl
}

var mapStringLiterals map[*ast.BasicLit]*sliteral = map[*ast.BasicLit]*sliteral{}

func registerStringLiteral(lit *ast.BasicLit) {
	if pkgName == "" {
		panic("no pkgName")
	}

	var strlen int
	for _, c := range []uint8(lit.Value) {
		if c != '\\' {
			strlen++
		}
	}

	label := fmt.Sprintf(".%s.S%d", pkgName, stringIndex)
	stringIndex++

	sl := &sliteral{
		label:  label,
		strlen: strlen - 2,
		value:  lit.Value,
	}
	mapStringLiterals[lit] = sl
	stringLiterals = append(stringLiterals, sl)
}

var localoffset localoffsetint

func getStructFieldOffset(field *ast.Field) int {
	if field.Doc == nil {
		panic("Doc is nil:" + field.Names[0].Name)
	}
	text := field.Doc.List[0].Text
	offset, err := strconv.Atoi(text)
	if err != nil {
		panic(text)
	}
	return offset
}

func setStructFieldOffset(field *ast.Field, offset int) {
	comment := &ast.Comment{
		Text: strconv.Itoa(offset),
	}
	commentGroup := &ast.CommentGroup{
		List: []*ast.Comment{comment},
	}
	field.Doc = commentGroup
}

func getStructFields(structTypeSpec *ast.TypeSpec) []*ast.Field {
	structType, ok := structTypeSpec.Type.(*ast.StructType)
	if !ok {
		throw(structTypeSpec.Type)
	}

	return structType.Fields.List
}

func getStructTypeSpec(namedStructType *Type) *ast.TypeSpec {
	if kind(namedStructType) != T_STRUCT {
		throw(namedStructType)
	}
	ident, ok := namedStructType.e.(*ast.Ident)
	if !ok {
		throw(namedStructType)
	}
	typeSpec, ok := ident.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		throw(ident.Obj.Decl)
	}
	return typeSpec
}

func lookupStructField(structTypeSpec *ast.TypeSpec, selName string) *ast.Field {
	for _, field := range getStructFields(structTypeSpec) {
		if field.Names[0].Name == selName {
			return field
		}
	}
	panic("Unexpected flow")
	return nil
}

func calcStructSizeAndSetFieldOffset(structTypeSpec *ast.TypeSpec) int {
	var offset int = 0
	for _, field := range getStructFields(structTypeSpec) {
		setStructFieldOffset(field, offset)
		size := getSizeOfType(e2t(field.Type))
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
				if varSpec.Type == nil {
					panic("type inference is not supported: " + obj.Name)
				}
				t := e2t(varSpec.Type)
				localoffset -= localoffsetint(getSizeOfType(t))
				obj.Data = newLocalVariable(obj.Name, localoffset)
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
	case *ast.RangeStmt:
		walkExpr(s.X)
		walkStmt(s.Body)
		localoffset -= localoffsetint(gInt.Data.(int))
		lenvar := newLocalVariable(".range.len", localoffset)
		localoffset -= localoffsetint(gInt.Data.(int))
		indexvar := newLocalVariable(".range.index", localoffset)
		mapRangeStmt[s] = &RangeStmtMisc{
			lenvar:   lenvar,
			indexvar: indexvar,
		}
	case *ast.IncDecStmt:
		walkExpr(s.X)
	default:
		throw(stmt)
	}
}

type RangeStmtMisc struct {
	lenvar   *Variable
	indexvar *Variable
}

var mapRangeStmt map[*ast.RangeStmt]*RangeStmtMisc = map[*ast.RangeStmt]*RangeStmtMisc{}

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
			registerStringLiteral(e)
		default:
			panic("Unexpected literal kind:" + e.Kind.String())
		}
	case *ast.CompositeLit:
		for _, v := range e.Elts {
			walkExpr(v)
		}
	case *ast.UnaryExpr:
		walkExpr(e.X)
	case *ast.BinaryExpr:
		walkExpr(e.X) // left
		walkExpr(e.Y) // right
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

var gNew = &ast.Object{
	Kind: ast.Fun,
	Name: "new",
	Decl: nil,
	Data: nil,
	Type: nil,
}

var gMake = &ast.Object{
	Kind: ast.Fun,
	Name: "make",
	Decl: nil,
	Data: nil,
	Type: nil,
}

var gAppend = &ast.Object{
	Kind: ast.Fun,
	Name: "append",
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

type isGlobal bool

func newGlobalVariable(name string) *Variable {
	return &Variable{
		name:         name,
		isGlobal:     true,
		globalSymbol: name,
		localOffset:  0,
	}
}

func newLocalVariable(name string, localoffset localoffsetint) *Variable {
	return &Variable{
		name:         name,
		isGlobal:     false,
		globalSymbol: "",
		localOffset:  localoffset,
	}
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
	universe.Insert(gNew)
	universe.Insert(gMake)
	universe.Insert(gAppend)
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
			spec := dcl.Specs[0]
			switch spc := spec.(type) {
			case *ast.ValueSpec:
				valSpec := spc
				//println(fmt.Sprintf("spec=%s", dcl.Tok))
				//fmt.Printf("# valSpec.type=%#v\n", valSpec.Type)
				t := e2t(valSpec.Type)
				nameIdent := valSpec.Names[0]
				nameIdent.Obj.Data = newGlobalVariable(nameIdent.Obj.Name)
				if len(valSpec.Values) > 0 {
					switch kind(t) {
					case T_STRING:
						lit, ok := valSpec.Values[0].(*ast.BasicLit)
						if !ok {
							throw(t)
						}
						registerStringLiteral(lit)
					case T_INT, T_UINT8, T_UINT16, T_UINTPTR:
						_, ok := valSpec.Values[0].(*ast.BasicLit)
						if !ok {
							throw(t)// allow only literal
						}
					default:
						throw(t)
					}
				}
				globalVars = append(globalVars, valSpec)
			case *ast.ImportSpec:
			case *ast.TypeSpec:
				typeSpec := spc
				assert(kind(e2t(typeSpec.Type)) == T_STRUCT, "should be T_STRUCT")
				calcStructSizeAndSetFieldOffset(typeSpec)
			default:
				throw(spec)
			}

		case *ast.FuncDecl:
			funcDecl := decl.(*ast.FuncDecl)
			localoffset = 0
			var paramoffset localoffsetint = 16
			for _, field := range funcDecl.Type.Params.List {
				obj := field.Names[0].Obj
				obj.Data = newLocalVariable(obj.Name, paramoffset)
				var varSize int = getSizeOfType(e2t(field.Type))
				paramoffset += localoffsetint(varSize)
				fmt.Printf("# field.Names[0].Obj=%#v\n", obj)
			}
			if funcDecl.Body == nil {
				break
			}
			for _, stmt := range funcDecl.Body.List {
				walkStmt(stmt)
			}
			fnc := &Func{
				name:      funcDecl.Name.Name,
				stmts:     funcDecl.Body.List,
				localarea: localoffset,
				argsarea:  paramoffset,
			}
			globalFuncs = append(globalFuncs, fnc)
		default:
			throw(decl)
		}
	}

	return pkg
}

type Func struct {
	name      string
	stmts     []ast.Stmt
	localarea localoffsetint
	argsarea  localoffsetint
}

type TypeKind string
const T_STRING TypeKind = "T_STRING"
const T_SLICE TypeKind = "T_SLICE"
const T_BOOL TypeKind = "T_BOOL"
const T_INT TypeKind = "T_INT"
const T_UINT8 TypeKind = "T_UINT8"
const T_UINT16 TypeKind = "T_UINT16"
const T_UINTPTR TypeKind = "T_UINTPTR"
const T_ARRAY TypeKind = "T_ARRAY"
const T_STRUCT TypeKind = "T_STRUCT"
const T_POINTER TypeKind = "T_POINTER"

var tBool *Type = &Type{
	e: &ast.Ident{
		NamePos: 0,
		Name:    "bool",
		Obj:     gBool,
	},
}


var tInt *Type = &Type{
	e: &ast.Ident{
		NamePos: 0,
		Name:    "int",
		Obj:     gInt,
	},
}

var tUintptr *Type = &Type{
	e: &ast.Ident{
		NamePos: 0,
		Name:    "uintptr",
		Obj:     gUintptr,
	},
}

var tUint8 *Type = &Type{
	e:&ast.Ident{
		NamePos: 0,
		Name:    "int",
		Obj:     gUint8,
	},
}

var tString *Type = &Type{
	e:&ast.Ident{
		NamePos: 0,
		Name:    "string",
		Obj:     gString,
	},
}

func getTypeOfExpr(expr ast.Expr) *Type {
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Obj.Kind == ast.Typ {
			panic("expression expected, but got Type")
		}
		if e.Obj.Kind == ast.Var {
			switch dcl := e.Obj.Decl.(type) {
			case *ast.ValueSpec:
				return e2t(dcl.Type)
			case *ast.Field:
				return e2t(dcl.Type)
			default:
				throw(e.Obj)
			}
		} else if e.Obj.Kind == ast.Con {
			if e.Obj == gTrue {
				return tBool
			} else if e.Obj == gFalse {
				return tBool
			} else {
				throw(e.Obj)
			}
		}
	case *ast.BasicLit:
		switch e.Kind.String() {
		case "STRING":
			return tString
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
		case "!":
			return tBool
		default:
			throw(e.Op.String())
		}
	case *ast.BinaryExpr:
		return getTypeOfExpr(e.X)
	case *ast.IndexExpr:
		list := e.X
		return getElementTypeOfListType(getTypeOfExpr(list))
	case *ast.CallExpr: // funcall or conversion
		switch fn := e.Fun.(type) {
		case *ast.Ident:
			if fn.Obj == nil {
				throw(fn)
			}
			switch fn.Obj.Kind {
			case ast.Typ: // conversion
				return e2t(fn)
			case ast.Fun:
				switch fn.Obj {
				case gLen, gCap:
					return tInt
				}
				switch decl := fn.Obj.Decl.(type) {
				case *ast.FuncDecl:
					assert(len(decl.Type.Results.List) == 1, "func is expected to return a single value")
					return e2t(decl.Type.Results.List[0].Type)
				default:
					throw(fn.Obj.Decl)
				}
			}
		case *ast.ArrayType: // conversion [n]T(e) or []T(e)
			return e2t(fn)
		case *ast.SelectorExpr: // X.Sel()
			xIdent, ok := fn.X.(*ast.Ident)
			if !ok {
				throw(fn)
			}
			if xIdent.Name == "unsafe" {
				if fn.Sel.Name == "Pointer" {
					// unsafe.Pointer(x)
					return tUintptr
				} else {
					panic("TBI")
				}
			}
			throw(fmt.Sprintf("%#v, %#v\n", xIdent, fn.Sel))
		default:
			throw(e.Fun)
		}
	case *ast.SliceExpr:
		underlyingCollectionType := getTypeOfExpr(e.X)
		var elementTyp ast.Expr
		switch colType := underlyingCollectionType.e.(type) {
		case *ast.ArrayType:
			elementTyp = colType.Elt
		}
		r := &ast.ArrayType{
			Len: nil,
			Elt: elementTyp,
		}
		return e2t(r)
	case *ast.StarExpr:
		t := getTypeOfExpr(e.X)
		ptrType, ok := t.e.(*ast.StarExpr)
		if !ok {
			throw(t)
		}
		return e2t(ptrType.X)
	case *ast.SelectorExpr:
		fmt.Printf("  # getTypeOfExpr(%s.%s)\n", e.X, e.Sel)
		structType := getStructTypeOfX(e)
		field := lookupStructField(getStructTypeSpec(structType), e.Sel.Name)
		return e2t(field.Type)
	case *ast.CompositeLit:
		return e2t(e.Type)
	default:
		panic(fmt.Sprintf("Unexpected expr type:%#v", expr))
	}
	throw(expr)
	return nil
}

func e2t(typeExpr ast.Expr) *Type {
	if typeExpr == nil {
		panic("nil is not allowed")
	}
	return &Type{
		e: typeExpr,
	}
}

func kind(t *Type) TypeKind {
	if t == nil {
		panic("nil type is not expected")
	}
	switch e := t.e.(type) {
	case *ast.Ident:
		if e.Obj == nil {
			panic("Unresolved identifier:" + e.Name)
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
				return kind(e2t(typeSpec.Type))
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
		throw(t)
	}
	return ""
}

func emitGlobalVariable(name *ast.Ident, t *Type, val ast.Expr) {
	typeKind := kind(t)
	fmt.Printf("%s: # T %s\n", name, typeKind)
	switch typeKind {
	case T_STRING:
		switch vl := val.(type) {
		case *ast.BasicLit:
			sl := getStringLiteral(vl)
			fmt.Printf("  .quad %s\n", sl.label)
			fmt.Printf("  .quad %d\n", sl.strlen)
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
		arrayType, ok := t.e.(*ast.ArrayType)
		assert(ok, "should be *ast.ArrayType")
		assert(arrayType.Len != nil, "slice type is not expected")
		basicLit, ok := arrayType.Len.(*ast.BasicLit)
		assert(ok, "should be *ast.BasicLit")
		length, err := strconv.Atoi(basicLit.Value)
		if err != nil {
			panic(err)
		}
		var zeroValue string
		switch kind(e2t(arrayType.Elt)) {
		case T_INT:
			zeroValue = fmt.Sprintf("  .quad 0 # int zero value\n")
		case T_UINT8:
			zeroValue = fmt.Sprintf("  .byte 0 # uint8 zero value\n")
		case T_STRING:
			zeroValue = fmt.Sprintf("  .quad 0 # string zero value (ptr)\n")
			zeroValue += fmt.Sprintf("  .quad 0 # string zero value (len)\n")
		default:
			throw(arrayType.Elt)
		}
		for i := 0; i < length; i++ {
			fmt.Printf(zeroValue)
		}
	default:
		throw(typeKind)
	}
}

func emitData(pkgName string) {
	fmt.Printf(".data\n")
	for _, sl := range stringLiterals {
		fmt.Printf("# string literals\n")
		fmt.Printf("%s:\n", sl.label)
		fmt.Printf("  .string %s\n", sl.value)
	}

	fmt.Printf("# ===== Global Variables =====\n")
	for _, spec := range globalVars {
		var val ast.Expr
		if len(spec.Values) > 0 {
			val = spec.Values[0]
		}
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariable(spec.Names[0], t, val)
	}
	fmt.Printf("# ==============================\n")
}

func emitText(pkgName string) {
	fmt.Printf(".text\n")
	var fnc *Func
	for _, fnc = range globalFuncs {
		emitFuncDecl(pkgName, fnc)
	}
}

func generateCode(pkgName string) {
	emitData(pkgName)
	emitText(pkgName)
}

var pkgName string
var stringLiterals []*sliteral
var stringIndex int

var globalVars []*ast.ValueSpec
var globalFuncs []*Func

var sourceFiles [2]string

func main() {
	sourceFiles[0] = "./runtime.go"
	sourceFiles[1] = "./t/source.go"
	var sourceFile string
	for _, sourceFile = range sourceFiles {
		globalVars = nil
		globalFuncs = nil
		stringLiterals = nil
		stringIndex = 0
		fset := &token.FileSet{}
		f, err := parser.ParseFile(fset, sourceFile, nil, 0)
		if err != nil {
			panic(err)
		}

		pkg := semanticAnalyze(fset, f)
		generateCode(pkg.Name())
	}
}
