package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"syscall"
)

func fmtSprintf(format string, a []string) string {
	var buf []uint8
	var inPercent bool
	var argIndex int
	var c uint8
	for _, c = range []uint8(format) {
		if inPercent {
			if c == '%' {
				buf = append(buf, c)
			} else {
				var arg string = a[argIndex]
				argIndex++
				var s string = arg // // p.printArg(arg, c)
				var _c uint8
				for _, _c = range []uint8(s) {
					buf = append(buf, _c)
				}
			}
			inPercent = false
		} else {
			if c == '%' {
				inPercent = true
			} else {
				buf = append(buf, c)
			}
		}
	}

	return string(buf)
}

func fmtPrintf(format string, a ...string) {
	var s string = fmtSprintf(format, a)
	syscall.Write(1, []uint8(s))
}

var __itoa_buf [100]uint8
var __itoa_r [100]uint8

func Itoa(ival int) string {
	var next int
	var right int
	var ix int = 0
	if ival == 0 {
		return "0"
	}
	var minus bool
	minus = false
	for ix = 0; ival != 0; ix = ix + 1 {
		if ival < 0 {
			ival = -1 * ival
			minus = true
			__itoa_r[0] = '-'
		} else {
			next = ival / 10
			right = ival - next*10
			ival = next
			__itoa_buf[ix] = uint8('0' + right)
		}
	}

	var j int
	var c uint8
	for j = 0; j < ix; j = j + 1 {
		c = __itoa_buf[ix-j-1]
		if minus {
			__itoa_r[j+1] = c
		} else {
			__itoa_r[j] = c
		}
	}

	return string(__itoa_r[0:ix])
}

func emitPopBool(comment string) {
	fmtPrintf("  popq %%rax # result of %s\n", comment)
}

func emitPopAddress(comment string) {
	fmtPrintf("  popq %%rax # address of %s\n", comment)
}

func emitPopString() {
	fmtPrintf("  popq %%rax # string.ptr\n")
	fmtPrintf("  popq %%rcx # string.len\n")
}

func emitPopSlice() {
	fmtPrintf("  popq %%rax # slice.ptr\n")
	fmtPrintf("  popq %%rcx # slice.len\n")
	fmtPrintf("  popq %%rdx # slice.cap\n")
}

func emitPushStackTop(condType *Type, comment string) {
	switch kind(condType) {
	case T_STRING:
		fmtPrintf("  movq 8(%%rsp), %%rcx # copy str.len from stack top (%s)\n", comment)
		fmtPrintf("  movq 0(%%rsp), %%rax # copy str.ptr from stack top (%s)\n", comment)
		fmtPrintf("  pushq %%rcx # str.len\n")
		fmtPrintf("  pushq %%rax # str.ptr\n")
	case T_POINTER, T_UINTPTR, T_BOOL, T_INT, T_UINT8, T_UINT16:
		fmtPrintf("  movq (%%rsp), %%rax # copy stack top value (%s) \n", comment)
		fmtPrintf("  pushq %%rax\n")
	default:
		throw(condType)
	}
}

func emitRevertStackPointer(size int) {
	fmtPrintf("  addq $%s, %%rsp # revert stack pointer\n", Itoa(size))
}

func emitAddConst(addValue int, comment string) {
	fmtPrintf("  # Add const: %s\n", comment)
	fmtPrintf("  popq %%rax\n")
	fmtPrintf("  addq $%s, %%rax\n", Itoa(addValue))
	fmtPrintf("  pushq %%rax\n")
}

type Type struct {
	e ast.Expr // original expr
}

type Arg struct {
	e      ast.Expr
	t      *Type // expected type
	offset int
}

func getPushSizeOfType(t *Type) int {
	switch kind(t) {
	case T_SLICE:
		return sliceSize
	case T_STRING:
		return stringSize
	case T_UINT8, T_UINT16, T_INT, T_BOOL:
		return intSize
	case T_UINTPTR, T_POINTER:
		return ptrSize
	case T_ARRAY, T_STRUCT:
		return ptrSize
	default:
		throw(t)
	}
	throw(t)
	return 0
}

func emitLoad(t *Type) {
	emitPopAddress(string(kind(t)))
	switch kind(t) {
	case T_SLICE:
		fmt.Printf("  movq %d(%%rax), %%rdx\n", 16)
		fmt.Printf("  movq %d(%%rax), %%rcx\n", 8)
		fmt.Printf("  movq %d(%%rax), %%rax\n", 0)
		fmt.Printf("  pushq %%rdx # cap\n")
		fmt.Printf("  pushq %%rcx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case T_STRING:
		fmt.Printf("  movq %d(%%rax), %%rdx\n", 8)
		fmt.Printf("  movq %d(%%rax), %%rax\n", 0)

		fmt.Printf("  pushq %%rdx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case T_UINT8:
		fmt.Printf("  movzbq %d(%%rax), %%rax # load uint8\n", 0)
		fmt.Printf("  pushq %%rax\n")
	case T_UINT16:
		fmt.Printf("  movzwq %d(%%rax), %%rax # load uint16\n", 0)
		fmt.Printf("  pushq %%rax\n")
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		fmt.Printf("  movq %d(%%rax), %%rax # load int\n", 0)
		fmt.Printf("  pushq %%rax\n")
	case T_ARRAY:
		// pure proxy
		fmt.Printf("  pushq %%rax\n")
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

type localoffsetint int

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
		emitPopSlice()
		fmt.Printf("  pushq %%rax # slice.ptr\n")
	case T_STRING:
		emitExpr(list, t)
		emitPopString()
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
		emitAddConst(offset, "struct head address + struct.field offset")
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
				emitPopSlice()
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
			emitPopString()
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
		emitPopSlice()
		fmt.Printf("  pushq %%rcx # len\n")
	case T_STRING:
		emitExpr(arg, nil)
		emitPopString()
		fmt.Printf("  pushq %%rcx # len\n")
	default:
		throw(kind(getTypeOfExpr(arg)))
	}
}

func emitCallMalloc(size int) {
	fmt.Printf("  pushq $%d\n", size)
	// call malloc and return pointer
	fmt.Printf("  callq runtime.malloc\n") // no need to invert args orders
	emitRevertStackPointer(intSize)
	fmt.Printf("  pushq %%rax # addr\n")
}

func emitArrayLiteral(arrayType *ast.ArrayType, arrayLen int, elts []ast.Expr) {
	elmType := e2t(arrayType.Elt)
	elmSize := getSizeOfType(elmType)
	memSize := elmSize * arrayLen
	emitCallMalloc(memSize) // push
	for i, elm := range elts {
		// emit lhs
		emitPushStackTop(tUintptr, "malloced address")
		emitAddConst(elmSize*i, "malloced address + elmSize * index")
		emitExpr(elm, elmType)
		emitStore(elmType)
	}
}

func emitInvertBoolValue() {
	emitPopBool("")
	fmt.Printf("  xor $1, %%rax\n")
	fmt.Printf("  pushq %%rax\n")
}

func emitTrue() {
	fmt.Printf("  pushq $1 # true\n")
}

func emitFalse() {
	fmt.Printf("  pushq $0 # false\n")
}

func newNumberLiteral(x int) *ast.BasicLit {
	e := &ast.BasicLit{
		ValuePos: 0,
		Kind:     token.INT,
		Value:    fmt.Sprintf("%d", x),
	}
	return e
}

func emitPushArgs2(arg0 ast.Expr, t0 *Type, arg1 ast.Expr, t1 *Type) {
	emitExpr(arg0, t0)
	emitExpr(arg1, t1)
	// pop arg1
	// pop arg0
	// push arg1
	// push arg0
}

func emitFuncallArgs(args []*Arg) int {
	var totalPushedSize int
	for _, arg := range args {
		arg.offset = totalPushedSize
		var t *Type
		if arg.t != nil {
			t = arg.t
		} else {
			t = getTypeOfExpr(arg.e)
		}
		totalPushedSize += getPushSizeOfType(t)
	}
	fmt.Printf("  subq $%d, %%rsp # for args\n", totalPushedSize)
	for _, arg := range args {
		emitExpr(arg.e, arg.t)
	}
	fmt.Printf("  addq $%d, %%rsp # for args\n", totalPushedSize)
	return totalPushedSize
}

// copy args on stack in reversed order
func emitReverseArgs(args []*Arg) {
	for _, arg := range args {
		var t *Type
		if arg.t != nil {
			t = arg.t
		} else {
			t = getTypeOfExpr(arg.e)
		}
		switch kind(t) {
		case T_INT, T_UINT8, T_POINTER, T_UINTPTR:
			fmt.Printf("  movq %d-8(%%rsp) , %%rax # load\n", -arg.offset)
			fmt.Printf("  movq %%rax, %d(%%rsp) # store\n", arg.offset)
		case T_STRING:
			fmt.Printf("  movq %d-16(%%rsp), %%rax\n", -arg.offset)
			fmt.Printf("  movq %d-8(%%rsp), %%rcx\n", -arg.offset)
			fmt.Printf("  movq %%rax, %d(%%rsp)\n", arg.offset)
			fmt.Printf("  movq %%rcx, %d+8(%%rsp)\n", arg.offset)
		case T_SLICE:
			fmt.Printf("  movq %d-24(%%rsp), %%rax\n", -arg.offset) // arg1: slc.ptr
			fmt.Printf("  movq %d-16(%%rsp), %%rcx\n", -arg.offset) // arg1: slc.len
			fmt.Printf("  movq %d-8(%%rsp), %%rdx\n", -arg.offset)  // arg1: slc.cap
			fmt.Printf("  movq %%rax, %d+0(%%rsp)\n", +arg.offset)  // arg1: slc.ptr
			fmt.Printf("  movq %%rcx, %d+8(%%rsp)\n", +arg.offset)  // arg1: slc.len
			fmt.Printf("  movq %%rdx, %d+16(%%rsp)\n", +arg.offset) // arg1: slc.cap
		default:
			throw(kind(t))
		}
	}
}

// Call without using func decl
func emitCallNonDecl(symbol string, eArgs []ast.Expr) {
	var args []*Arg
	for _, eArg := range eArgs {
		arg := &Arg{
			e: eArg,
			t: nil,
		}
		args = append(args, arg)
	}
	emitCall(symbol, args)
}

func emitCall(symbol string, args []*Arg) {
	totalPushedSize := emitFuncallArgs(args)
	emitReverseArgs(args)
	fmt.Printf("  callq %s\n", symbol)
	emitRevertStackPointer(totalPushedSize)
}

// ABI of stack layout
//
// string:
//   str.ptr
//   str.len
// slice:
//   slc.ptr
//   slc.len
//   slc.cap
//
// ABI of function call
//
// call f(i1 int, i2 int)
//   -- stack top
//   i1
//   i2
//   --
//
// call f(i int, s string, slc []T)
//   -- stack top
//   i
//   s.ptr
//   s.len
//   slc.ptr
//   slc.len
//   slc.cap
//   --
func emitExpr(expr ast.Expr, forceType *Type) {
	switch e := expr.(type) {
	case *ast.Ident:
		switch e.Obj {
		case gTrue: // true constant
			emitTrue()
			return
		case gFalse: // false constant
			emitFalse()
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

		switch e.Obj.Kind {
		case ast.Var:
			emitAddr(e)
			emitLoad(getTypeOfExpr(e))
		case ast.Con:
			valSpec, ok := e.Obj.Decl.(*ast.ValueSpec)
			assert(ok, "should be *ast.ValueSpec")
			lit, ok := valSpec.Values[0].(*ast.BasicLit)
			assert(ok, "should be *ast.BasicLit")
			var t *Type
			if valSpec.Type != nil {
				t = e2t(valSpec.Type)
			} else {
				t = forceType
			}
			emitExpr(lit, t)
		default:
			panic("Unexpected ident kind:" + e.Obj.Kind.String())
		}
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
					emitPopSlice()
					fmt.Printf("  pushq %%rdx # cap\n")
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
					elmSize := getSizeOfType(e2t(arrayType.Elt))
					numlit := newNumberLiteral(elmSize)

					var args []*Arg = []*Arg{
						// elmSize
						&Arg{
							e: numlit,
							t: tInt,
						},
						// len
						&Arg{
							e: e.Args[1],
							t: tInt,
						},
						// cap
						&Arg{
							e: e.Args[2],
							t: tInt,
						},
					}

					emitCall("runtime.makeSlice", args)
					fmt.Printf("  pushq %%rsi # slice cap\n")
					fmt.Printf("  pushq %%rdi # slice len\n")
					fmt.Printf("  pushq %%rax # slice ptr\n")
				default:
					throw(typeArg)
				}
			case gAppend:
				var sliceArg ast.Expr = e.Args[0]
				var elemArg ast.Expr = e.Args[1]
				var elmType *Type = getElementTypeOfListType(getTypeOfExpr(sliceArg))
				var elmSize int = getSizeOfType(elmType)

				var args []*Arg = []*Arg{
					// slice
					&Arg{
						e: sliceArg,
						t: nil,
					},
					// elm
					&Arg{
						e: elemArg,
						t: elmType,
					},
				}

				var symbol string
				switch elmSize {
				case 1:
					symbol = "runtime.append1"
				case 8:
					symbol = "runtime.append8"
				case 16:
					symbol = "runtime.append16"
				case 24:
					symbol = "runtime.append24"
				default:
					throw(elmSize)
				}
				emitCall(symbol, args)
				fmt.Printf("  pushq %%rsi # slice cap\n")
				fmt.Printf("  pushq %%rdi # slice len\n")
				fmt.Printf("  pushq %%rax # slice ptr\n")
			default:
				if fn.Name == "print" {
					// builtin print
					args := []*Arg{&Arg{
						e: e.Args[0],
					}}
					var symbol string
					switch kind(getTypeOfExpr(e.Args[0])) {
					case T_STRING:
						symbol = "runtime.printstring"
					case T_INT:
						symbol = "runtime.printint"
					default:
						panic("TBI")
					}
					emitCall(symbol, args)
				} else {
					if fn.Name == "makeSlice1" || fn.Name == "makeSlice8" || fn.Name == "makeSlice16" || fn.Name == "makeSlice24" {
						fn.Name = "makeSlice"
					}
					// general function call
					obj := fn.Obj //.Kind == FN
					fndecl, ok := obj.Decl.(*ast.FuncDecl)
					if !ok {
						throw(fn.Obj)
					}
					params := fndecl.Type.Params.List
					var variadicArgs []ast.Expr // nil means there is no varadic in funcdecl
					var variadicElp *ast.Ellipsis
					var args []*Arg
					var argIndex int
					var eArg ast.Expr
					var param *ast.Field
					for argIndex, eArg = range e.Args {
						if argIndex < len(params) {
							param = params[argIndex]
							elp, ok := param.Type.(*ast.Ellipsis)
							if ok {
								variadicElp = elp
								variadicArgs = make([]ast.Expr, 0)
							}
						}
						if variadicArgs != nil {
							variadicArgs = append(variadicArgs, eArg)
							continue
						}

						paramType := e2t(param.Type)
						arg := &Arg{
							e: eArg,
							t: paramType,
						}
						args = append(args, arg)
					}

					if variadicArgs != nil {
						sliceType := &ast.ArrayType{Elt: variadicElp.Elt}
						slicelite := &ast.CompositeLit{
							Type:       sliceType,
							Lbrace:     0,
							Elts:       variadicArgs,
							Rbrace:     0,
							Incomplete: false,
						}
						args = append(args, &Arg{
							e:      slicelite,
							t:      e2t(sliceType),
							offset: 0,
						})
					} else if len(e.Args) < len(params) {
						// Add nil as a variadic arg
						param := params[argIndex+1]
						elp, ok := param.Type.(*ast.Ellipsis)
						assert(ok, "compile error")
						//variadicArgs = make([]ast.Expr, 0,0)
						args = append(args, &Arg{
							e:      eNil,
							t:      e2t(elp),
							offset: 0,
						})
					}

					symbol := pkgName + "." + fn.Name
					emitCall(symbol, args)

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
				// func decl is in runtime
				emitCallNonDecl(symbol, e.Args)
			case "syscall.Open":
				// func decl is in runtime
				emitCallNonDecl(symbol, e.Args)
				fmt.Printf("  pushq %%rax # fd\n")
			case "syscall.Read":
				// func decl is in runtime
				emitCallNonDecl(symbol, e.Args)
				fmt.Printf("  pushq %%rax # fd\n")
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
			emitInvertBoolValue()
		default:
			throw(e)
		}
	case *ast.BinaryExpr:
		if kind(getTypeOfExpr(e.X)) == T_STRING {
			var args []*Arg = []*Arg{
				&Arg{
					e:      e.X,
					t:      nil,
					offset: 0,
				},
				&Arg{
					e:      e.Y,
					t:      nil,
					offset: 0,
				},
			}

			switch e.Op.String() {
			case "+":
				emitCall("runtime.catstrings", args)
				fmt.Printf("  pushq %%rdi # slice len\n")
				fmt.Printf("  pushq %%rax # slice ptr\n")
			case "==":
				emitFuncallArgs(args)
				emitReverseArgs(args)
				emitCompEq(getTypeOfExpr(e.X))
			case "!=":
				emitFuncallArgs(args)
				emitReverseArgs(args)
				emitCompEq(getTypeOfExpr(e.X))
				emitInvertBoolValue()
			default:
				throw(e.Op.String())
			}
			return
		}

		switch e.Op.String() {
		case "&&":
			labelid++
			labelExitWithFalse := fmt.Sprintf(".L.%d.false", labelid)
			labelExit := fmt.Sprintf(".L.%d.exit", labelid)
			emitExpr(e.X, nil) // left
			emitPopBool("left")
			fmt.Printf("  cmpq $1, %%rax\n")
			// exit with false if left is false
			fmt.Printf("  jne %s\n", labelExitWithFalse)

			// if left is true, then eval right and exit
			emitExpr(e.Y, nil) // right
			fmt.Printf("  jmp %s\n", labelExit)

			fmt.Printf("  %s:\n", labelExitWithFalse)
			emitFalse()
			fmt.Printf("  %s:\n", labelExit)
			return
		case "||":
			labelid++
			labelExitWithTrue := fmt.Sprintf(".L.%d.true", labelid)
			labelExit := fmt.Sprintf(".L.%d.exit", labelid)
			emitExpr(e.X, nil) // left
			emitPopBool("left")
			fmt.Printf("  cmpq $1, %%rax\n")
			// exit with true if left is true
			fmt.Printf("  je %s\n", labelExitWithTrue)

			// if left is false, then eval right and exit
			emitExpr(e.Y, nil) // right
			fmt.Printf("  jmp %s\n", labelExit)

			fmt.Printf("  %s:\n", labelExitWithTrue)
			emitTrue()
			fmt.Printf("  %s:\n", labelExit)
			return
		}

		t := getTypeOfExpr(e.X)
		fmt.Printf("  # start %T\n", e)
		emitExpr(e.X, nil) // left
		emitExpr(e.Y, t)   // right
		switch e.Op.String() {
		case "+":
			fmt.Printf("  popq %%rcx # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  addq %%rcx, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "-":
			fmt.Printf("  popq %%rcx # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  subq %%rcx, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "*":
			fmt.Printf("  popq %%rcx # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  imulq %%rcx, %%rax\n")
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
			emitCompEq(t)
		case "!=":
			emitCompEq(t)
			emitInvertBoolValue()
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
			emitPopAddress("malloc")
			fmt.Printf("  pushq $%d # slice.cap\n", length)
			fmt.Printf("  pushq $%d # slice.len\n", length)
			fmt.Printf("  pushq %%rax # slice.ptr\n")
		default:
			throw(e.Type)
		}
	case *ast.SliceExpr: // list[low:high]
		list := e.X
		listType := getTypeOfExpr(list)
		emitExpr(e.High, tInt) // intval
		emitExpr(e.Low, tInt)  // intval
		//emitExpr(e.Max) // @TODO
		fmt.Printf("  popq %%rcx # low\n")
		fmt.Printf("  popq %%rax # high\n")
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
	emitPopAddress("list head")
	fmt.Printf("  popq %%rcx # index id\n")
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

func emitCompEq(t *Type) {
	switch kind(t) {
	case T_STRING:
		fmt.Printf("  callq runtime.cmpstrings\n")
		emitRevertStackPointer(stringSize * 2)
		fmt.Printf("  pushq %%rax # cmp result (1 or 0)\n")
	case T_INT, T_UINT8, T_UINT16, T_UINTPTR, T_POINTER:
		emitCompExpr("sete")
	case T_SLICE:
		emitCompExpr("sete") // @FIXME this is not correct
	default:
		throw(kind(t))
	}
}

//@TODO handle larger types than int
func emitCompExpr(inst string) {
	fmt.Printf("  popq %%rcx # right\n")
	fmt.Printf("  popq %%rax # left\n")
	fmt.Printf("  cmpq %%rcx, %%rax\n")
	fmt.Printf("  %s %%al\n", inst)
	fmt.Printf("  movzbq %%al, %%rax\n") // true:1, false:0
	fmt.Printf("  pushq %%rax\n")
}

func emitStore(t *Type) {
	fmt.Printf("  # emitStore(%s)\n", kind(t))
	switch kind(t) {
	case T_SLICE:
		emitPopSlice()
		fmt.Printf("  popq %%rsi # lhs ptr addr\n")
		fmt.Printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", 0)
		fmt.Printf("  movq %%rcx, %d(%%rsi) # len to len\n", 8)
		fmt.Printf("  movq %%rdx, %d(%%rsi) # cap to cap\n", 16)
	case T_STRING:
		emitPopString()
		fmt.Printf("  popq %%rsi # lhs ptr addr\n")
		fmt.Printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", 0)
		fmt.Printf("  movq %%rcx, %d(%%rsi) # len to len\n", 8)
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
		emitRevertStackPointer(ptrSize*2 + intSize)
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

		emitExpr(s.Cond, nil)
		emitPopBool("if condition")
		fmt.Printf("  cmpq $1, %%rax\n")
		if s.Else != nil {
			fmt.Printf("  jne %s # jmp if false\n", labelElse)
			emitStmt(s.Body) // then
			fmt.Printf("  jmp %s\n", labelEndif)
			fmt.Printf("  %s:\n", labelElse)
			emitStmt(s.Else) // then
		} else {
			fmt.Printf("  jne %s # jmp if false\n", labelEndif)
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
		labelCond := fmt.Sprintf(".L.for.cond.%d", labelid)
		labelPost := fmt.Sprintf(".L.for.post.%d", labelid)
		labelExit := fmt.Sprintf(".L.for.exit.%d", labelid)
		forStmt, ok := mapForNodeToFor[s]
		assert(ok, "map value should exist")
		forStmt.labelPost = labelPost
		forStmt.labelExit = labelExit

		emitStmt(s.Init)

		fmt.Printf("  %s:\n", labelCond)
		emitExpr(s.Cond, nil)
		emitPopBool("for condition")
		fmt.Printf("  cmpq $1, %%rax\n")
		fmt.Printf("  jne %s # jmp if false\n", labelExit)
		emitStmt(s.Body)
		fmt.Printf("  %s:\n", labelPost) // used for "continue"
		emitStmt(s.Post)
		fmt.Printf("  jmp %s\n", labelCond)
		fmt.Printf("  %s:\n", labelExit)
	case *ast.RangeStmt: // only for array and slice
		labelid++
		labelCond := fmt.Sprintf(".L.range.cond.%d", labelid)
		labelPost := fmt.Sprintf(".L.range.post.%d", labelid)
		labelExit := fmt.Sprintf(".L.range.exit.%d", labelid)

		forStmt, ok := mapRangeNodeToFor[s]
		assert(ok, "map value should exist")
		forStmt.labelPost = labelPost
		forStmt.labelExit = labelExit
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
		emitPopBool(" indexvar < lenvar")
		fmt.Printf("  cmpq $1, %%rax\n")
		fmt.Printf("  jne %s # jmp if false\n", labelExit)

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
		fmt.Printf("  %s:\n", labelPost)   // used for "continue"
		emitVariableAddr(rngMisc.indexvar) // lhs
		emitVariableAddr(rngMisc.indexvar) // rhs
		emitLoad(tInt)
		emitAddConst(1, "indexvar value ++")
		emitStore(tInt)

		fmt.Printf("  jmp %s\n", labelCond)

		fmt.Printf("  %s:\n", labelExit)
	case *ast.IncDecStmt:
		var addValue int
		switch s.Tok.String() {
		case "++":
			addValue = 1
		case "--":
			addValue = -1
		default:
			throw(s.Tok.String())
		}

		emitAddr(s.X)
		emitExpr(s.X, nil)
		emitAddConst(addValue, "rhs ++ or --")
		emitStore(getTypeOfExpr(s.X))
	case *ast.SwitchStmt:
		labelid++
		labelEnd := fmt.Sprintf(".L.switch.%d.exit", labelid)
		if s.Init != nil {
			panic("TBI")
		}
		if s.Tag == nil {
			panic("TBI")
		}
		emitExpr(s.Tag, nil)
		condType := getTypeOfExpr(s.Tag)
		cases := s.Body.List
		var labels []string = make([]string, len(cases))
		var defaultLabel string
		for i, c := range cases {
			cc, ok := c.(*ast.CaseClause)
			assert(ok, "should be *ast.CaseClause")
			labelid++
			labelCase := fmt.Sprintf(".L.case.%d", labelid)
			labels[i] = labelCase
			if cc.List == nil {
				defaultLabel = labelCase
				continue
			}
			for _, e := range cc.List {
				assert(getSizeOfType(condType) <= 8 || kind(condType) == T_STRING, "should be one register size or string")
				emitPushStackTop(condType, "switch expr")
				emitExpr(e, nil)
				emitCompEq(condType)
				emitPopBool(" of switch-case comparison")
				fmt.Printf("  cmpq $1, %%rax\n")
				fmt.Printf("  je %s # jump if match\n", labelCase)
			}

		}
		if defaultLabel != "" {
			// default
			fmt.Printf("  jmp %s\n", defaultLabel)
		}

		emitRevertStackTop(condType)
		for i, c := range cases {
			cc, ok := c.(*ast.CaseClause)
			assert(ok, "should be *ast.CaseClause")
			fmt.Printf("%s:\n", labels[i])
			for _, _s := range cc.Body {
				emitStmt(_s)
			}
			fmt.Printf("  jmp %s\n", labelEnd)
		}
		fmt.Printf("%s:\n", labelEnd)
		/*
			case *ast.CaseClause:
		*/
	case *ast.BranchStmt:
		containerFor, ok := mapBranchToFor[s]
		assert(ok, "map value should exist")
		switch s.Tok {
		case token.CONTINUE:
			fmt.Printf("jmp %s # continue\n", containerFor.labelPost)
		case token.BREAK:
			fmt.Printf("jmp %s # break\n", containerFor.labelExit)
		default:
			throw(s.Tok)
		}
	default:
		throw(stmt)
	}
}

func emitRevertStackTop(t *Type) {
	fmt.Printf("  addq $%d, %%rsp # revert stack top\n", getSizeOfType(t))
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
		forStmt := new(ForStmt)
		forStmt.astFor = s
		forStmt.outer = currentFor
		currentFor = forStmt
		mapForNodeToFor[s] = forStmt
		walkStmt(s.Init)
		walkExpr(s.Cond)
		walkStmt(s.Post)
		walkStmt(s.Body)
		currentFor = forStmt.outer
	case *ast.RangeStmt:
		forStmt := new(ForStmt)
		forStmt.astRange = s
		forStmt.outer = currentFor
		currentFor = forStmt
		mapRangeNodeToFor[s] = forStmt
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
		currentFor = forStmt.outer
	case *ast.IncDecStmt:
		walkExpr(s.X)
	case *ast.SwitchStmt:
		if s.Init != nil {
			walkStmt(s.Init)
		}
		if s.Tag != nil {
			walkExpr(s.Tag)
		}
		walkStmt(s.Body)
	case *ast.CaseClause:
		for _, e := range s.List {
			walkExpr(e)
		}
		for _, stmt := range s.Body {
			walkStmt(stmt)
		}
	case *ast.BranchStmt:
		assert(currentFor != nil, "break or continue should be in for body")
		mapBranchToFor[s] = currentFor
	default:
		throw(stmt)
	}
}

type ForStmt struct {
	kind      int    // 1:for 2:range
	labelPost string // for continue
	labelExit string // for break
	outer     *ForStmt
	astFor    *ast.ForStmt
	astRange  *ast.RangeStmt
}

var currentFor *ForStmt

var mapForNodeToFor map[*ast.ForStmt]*ForStmt = map[*ast.ForStmt]*ForStmt{}
var mapRangeNodeToFor map[*ast.RangeStmt]*ForStmt = map[*ast.RangeStmt]*ForStmt{}
var mapBranchToFor map[*ast.BranchStmt]*ForStmt = map[*ast.BranchStmt]*ForStmt{}

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

const sliceSize int = 24
const stringSize int = 16
const intSize int = 8
const ptrSize int = 8

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

var eNil = &ast.Ident{
	Obj: gNil,
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

func semanticAnalyze(fset *token.FileSet, fiile *ast.File) string {
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
	pkgName = ap.Name

	var unresolved []*ast.Ident
	for _, ident := range fiile.Unresolved {
		if obj := universe.Lookup(ident.Name); obj != nil {
			ident.Obj = obj
		} else {
			unresolved = append(unresolved, ident)
		}
	}

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
							throw(t) // allow only literal
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
			fmt.Printf("# funcdef %s\n", funcDecl.Name.Name)
			localoffset = 0
			var paramoffset localoffsetint = 16
			for _, field := range funcDecl.Type.Params.List {
				obj := field.Names[0].Obj
				obj.Data = newLocalVariable(obj.Name, paramoffset)
				var varSize int = getSizeOfType(e2t(field.Type))
				paramoffset += localoffsetint(varSize)
				fmt.Printf("# field.Names[0].Obj=%#v\n", obj)
				fmt.Printf("#   field.Type=%#v\n", field.Type)
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

	return pkgName
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
	e: &ast.Ident{
		NamePos: 0,
		Name:    "int",
		Obj:     gUint8,
	},
}

var tString *Type = &Type{
	e: &ast.Ident{
		NamePos: 0,
		Name:    "string",
		Obj:     gString,
	},
}

func getTypeOfExpr(expr ast.Expr) *Type {
	switch e := expr.(type) {
	case *ast.Ident:
		switch e.Obj.Kind {
		case ast.Var:
			switch dcl := e.Obj.Decl.(type) {
			case *ast.ValueSpec:
				return e2t(dcl.Type)
			case *ast.Field:
				return e2t(dcl.Type)
			default:
				throw(e.Obj)
			}
		case ast.Con:
			if e.Obj == gTrue {
				return tBool
			} else if e.Obj == gFalse {
				return tBool
			} else {
				switch dcl := e.Obj.Decl.(type) {
				case *ast.ValueSpec:
					return e2t(dcl.Type)
				default:
					throw(e.Obj)
				}
			}
		default:
			throw(e.Obj.Kind)
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
			if xIdent.Name == "unsafe" && fn.Sel.Name == "Pointer" {
				// unsafe.Pointer(x)
				return tUintptr
			} else {
				panic("TBI")
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
	case *ast.Ellipsis: // x ...T
		return T_SLICE // @TODO is this right ?
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

func parseFile(fset *token.FileSet, filename string) *ast.File {
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		panic(err)
	}
	return f
}

func main() {
	var sourceFiles []string = []string{"./runtime.go", "./t/source.go"}
	var sourceFile string
	for _, sourceFile = range sourceFiles {
		globalVars = nil
		globalFuncs = nil
		stringLiterals = nil
		stringIndex = 0
		fset := &token.FileSet{}
		f := parseFile(fset, sourceFile)
		pkgName := semanticAnalyze(fset, f)
		generateCode(pkgName)
	}
}
