package main

import (
	"os"
	"syscall"

	"github.com/DQNEO/babygo/lib/myfmt"
	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/path"
	"github.com/DQNEO/babygo/lib/strconv"
	"github.com/DQNEO/babygo/lib/strings"
)

// --- foundation ---
var __func__ = "__func__"

func assert(bol bool, msg string, caller string) {
	if !bol {
		panic2(caller, msg)
	}
}

func throw(s string) {
	panic(s)
}

func panic2(caller string, x string) {
	panic("[" + caller + "] " + x)
}

var debugFrontEnd bool

func logf(format string, a ...interface{}) {
	if !debugFrontEnd {
		return
	}
	var f = "# " + format
	var s = myfmt.Sprintf(f, a...)
	syscall.Write(1, []uint8(s))
}

// --- codegen ---
var debugCodeGen bool

func emitComment(indent int, format string, a ...interface{}) {
	if !debugCodeGen {
		return
	}
	var spaces []uint8
	var i int
	for i = 0; i < indent; i++ {
		spaces = append(spaces, ' ')
	}
	var format2 = string(spaces) + "# " + format
	var s = myfmt.Sprintf(format2, a...)
	syscall.Write(1, []uint8(s))
}

func evalInt(expr astExpr) int {
	switch e := expr.(type)  {
	case *astBasicLit:
		return strconv.Atoi(e.Value)
	default:
		panic("Unknown type")
	}
	return 0
}

func emitPopPrimitive(comment string) {
	myfmt.Printf("  popq %%rax # result of %s\n", comment)
}

func emitPopBool(comment string) {
	myfmt.Printf("  popq %%rax # result of %s\n", comment)
}

func emitPopAddress(comment string) {
	myfmt.Printf("  popq %%rax # address of %s\n", comment)
}

func emitPopString() {
	myfmt.Printf("  popq %%rax # string.ptr\n")
	myfmt.Printf("  popq %%rcx # string.len\n")
}

func emitPopInterFace() {
	myfmt.Printf("  popq %%rax # eface.dtype\n")
	myfmt.Printf("  popq %%rcx # eface.data\n")
}

func emitPopSlice() {
	myfmt.Printf("  popq %%rax # slice.ptr\n")
	myfmt.Printf("  popq %%rcx # slice.len\n")
	myfmt.Printf("  popq %%rdx # slice.cap\n")
}

func emitPushStackTop(condType *Type, comment string) {
	switch kind(condType) {
	case T_STRING:
		myfmt.Printf("  movq 8(%%rsp), %%rcx # copy str.len from stack top (%s)\n", comment)
		myfmt.Printf("  movq 0(%%rsp), %%rax # copy str.ptr from stack top (%s)\n", comment)
		myfmt.Printf("  pushq %%rcx # str.len\n")
		myfmt.Printf("  pushq %%rax # str.ptr\n")
	case T_POINTER, T_UINTPTR, T_BOOL, T_INT, T_UINT8, T_UINT16:
		myfmt.Printf("  movq (%%rsp), %%rax # copy stack top value (%s) \n", comment)
		myfmt.Printf("  pushq %%rax\n")
	default:
		throw(kind(condType))
	}
}

func emitRevertStackPointer(size int) {
	myfmt.Printf("  addq $%d, %%rsp # revert stack pointer\n", size)
}

func emitAddConst(addValue int, comment string) {
	emitComment(2, "Add const: %s\n", comment)
	myfmt.Printf("  popq %%rax\n")
	myfmt.Printf("  addq $%d, %%rax\n", addValue)
	myfmt.Printf("  pushq %%rax\n")
}

func emitLoad(t *Type) {
	if t == nil {
		panic2(__func__, "nil type error\n")
	}
	emitPopAddress(kind(t))
	switch kind(t) {
	case T_SLICE:
		myfmt.Printf("  movq %d(%%rax), %%rdx\n", 16)
		myfmt.Printf("  movq %d(%%rax), %%rcx\n", 8)
		myfmt.Printf("  movq %d(%%rax), %%rax\n", 0)
		myfmt.Printf("  pushq %%rdx # cap\n")
		myfmt.Printf("  pushq %%rcx # len\n")
		myfmt.Printf("  pushq %%rax # ptr\n")
	case T_STRING:
		myfmt.Printf("  movq %d(%%rax), %%rdx # len\n", 8)
		myfmt.Printf("  movq %d(%%rax), %%rax # ptr\n", 0)
		myfmt.Printf("  pushq %%rdx # len\n")
		myfmt.Printf("  pushq %%rax # ptr\n")
	case T_INTERFACE:
		myfmt.Printf("  movq %d(%%rax), %%rdx # data\n", 8)
		myfmt.Printf("  movq %d(%%rax), %%rax # dtype\n", 0)
		myfmt.Printf("  pushq %%rdx # data\n")
		myfmt.Printf("  pushq %%rax # dtype\n")
	case T_UINT8:
		myfmt.Printf("  movzbq %d(%%rax), %%rax # load uint8\n", 0)
		myfmt.Printf("  pushq %%rax\n")
	case T_UINT16:
		myfmt.Printf("  movzwq %d(%%rax), %%rax # load uint16\n", 0)
		myfmt.Printf("  pushq %%rax\n")
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		myfmt.Printf("  movq %d(%%rax), %%rax # load int\n", 0)
		myfmt.Printf("  pushq %%rax\n")
	case T_ARRAY, T_STRUCT:
		// pure proxy
		myfmt.Printf("  pushq %%rax\n")
	default:
		panic2(__func__, "TBI:kind="+kind(t))
	}
}

func emitVariableAddr(variable *Variable) {
	emitComment(2, "emit Addr of variable \"%s\" \n", variable.name)

	if variable.isGlobal {
		myfmt.Printf("  leaq %s(%%rip), %%rax # global variable \"%s\"\n", variable.globalSymbol,  variable.name)
	} else {
		myfmt.Printf("  leaq %d(%%rbp), %%rax # local variable \"%s\"\n", variable.localOffset,  variable.name)
	}

	myfmt.Printf("  pushq %%rax # variable address\n")
}

func emitListHeadAddr(list astExpr) {
	var t = getTypeOfExpr(list)
	switch kind(t) {
	case T_ARRAY:
		emitAddr(list) // array head
	case T_SLICE:
		emitExpr(list, nil)
		emitPopSlice()
		myfmt.Printf("  pushq %%rax # slice.ptr\n")
	case T_STRING:
		emitExpr(list, nil)
		emitPopString()
		myfmt.Printf("  pushq %%rax # string.ptr\n")
	default:
		panic2(__func__, "kind="+kind(getTypeOfExpr(list)))
	}
}

func emitAddr(expr astExpr) {
	emitComment(2, "[emitAddr] %s\n", dtypeOf(expr))
	switch e := expr.(type)  {
	case *astIdent:
		if e.Name == "_" {
			panic(" \"_\" has no address")
		}
		if e.Obj == nil {
			throw("e.Obj is nil: " + e.Name)
		}
		if e.Obj.Kind == astVar {
			assert(e.Obj.Variable != nil,
				"ERROR: Obj.Variable is not set for ident : "+e.Obj.Name, __func__)
			emitVariableAddr(e.Obj.Variable)
		} else {
			panic2(__func__, "Unexpected Kind "+e.Obj.Kind)
		}
	case *astIndexExpr:
		emitExpr(e.Index, nil) // index number
		var list = e.X
		var elmType = getTypeOfExpr(expr)
		emitListElementAddr(list, elmType)
	case *astStarExpr:
		emitExpr(expr2StarExpr(expr).X, nil)
	case *astSelectorExpr: // (X).Sel
		var typeOfX = getTypeOfExpr(e.X)
		var structType *Type
		switch kind(typeOfX) {
		case T_STRUCT:
			// strct.field
			structType = typeOfX
			emitAddr(e.X)
		case T_POINTER:
			// ptr.field
			var ptrType = expr2StarExpr(typeOfX.e)
			structType = e2t(ptrType.X)
			emitExpr(e.X, nil)
		default:
			panic2(__func__, "TBI:"+kind(typeOfX))
		}
		var field = lookupStructField(getStructTypeSpec(structType), e.Sel.Name)
		var offset = getStructFieldOffset(field)
		emitAddConst(offset, "struct head address + struct.field offset")
	case *astCompositeLit:
		var knd = kind(getTypeOfExpr(expr))
		switch knd {
		case T_STRUCT:
			// result of evaluation of a struct literal is its address
			emitExpr(expr, nil)
		default:
			panic2(__func__, "TBI "+ knd)
		}
	default:
		panic2(__func__, "TBI "+ dtypeOf(expr))
	}
}

func isType(expr astExpr) bool {
	switch e := expr.(type)  {
	case *astArrayType:
		return true
	case *astIdent:
		if e == nil {
			panic2(__func__, "ident should not be nil")
		}
		if e.Obj == nil {
			panic2(__func__, " unresolved ident:"+e.Name)
		}
		emitComment(2, "[isType][DEBUG] e.Name = %s\n", e.Name)
		emitComment(2, "[isType][DEBUG] e.Obj = %s,%s\n",
			e.Obj.Name, e.Obj.Kind)
		return e.Obj.Kind == astTyp
	case *astParenExpr:
		return isType(e.X)
	case *astStarExpr:
		return isType(expr2StarExpr(expr).X)
	case *astInterfaceType:
		return true
	default:
		emitComment(2, "[isType][%s] is not considered a type\n", dtypeOf(expr))
	}

	return false

}

// explicit conversion T(e)
func emitConversion(toType *Type, arg0 astExpr) {
	emitComment(2, "[emitConversion]\n")
	var to = toType.e
	switch tt := to.(type)  {
	case *astIdent:
		switch tt.Obj {
		case gString: // string(e)
			switch kind(getTypeOfExpr(arg0)) {
			case T_SLICE: // string(slice)
				emitExpr(arg0, nil) // slice
				emitPopSlice()
				myfmt.Printf("  pushq %%rcx # str len\n")
				myfmt.Printf("  pushq %%rax # str ptr\n")
			}
		case gInt, gUint8, gUint16, gUintptr: // int(e)
			emitComment(2, "[emitConversion] to int \n")
			emitExpr(arg0, nil)
		default:
			if tt.Obj.Kind == astTyp {
				var isTypeSpec bool
				_, isTypeSpec = tt.Obj.Decl.(*astTypeSpec)
				if !isTypeSpec {
					panic2(__func__, "Something is wrong")
				}
				//e2t(tt.Obj.Decl.typeSpec.Type))
				emitExpr(arg0, nil)
			} else{
				panic2(__func__, "[*astIdent] TBI : "+tt.Obj.Name)
			}
		}
	case *astArrayType: // Conversion to slice
		var arrayType = expr2ArrayType(to)
		if arrayType.Len != nil {
			panic2(__func__, "internal error")
		}
		if (kind(getTypeOfExpr(arg0))) != T_STRING {
			panic2(__func__, "source type should be string")
		}
		emitComment(2, "Conversion of string => slice \n")
		emitExpr(arg0, nil)
		emitPopString()
		myfmt.Printf("  pushq %%rcx # cap\n")
		myfmt.Printf("  pushq %%rcx # len\n")
		myfmt.Printf("  pushq %%rax # ptr\n")
	case *astParenExpr:
		emitConversion(e2t(tt.X), arg0)
	case *astStarExpr: // (*T)(e)
		emitComment(2, "[emitConversion] to pointer \n")
		emitExpr(arg0, nil)
	case *astInterfaceType:
		emitExpr(arg0, nil)
		if isInterface(getTypeOfExpr(arg0))  {
			// do nothing
		} else {
			// Convert dynamic value to interface
			emitConvertToInterface(getTypeOfExpr(arg0))
		}
	default:
		panic2(__func__, "TBI :"+ dtypeOf(to))
	}
}

func emitZeroValue(t *Type) {
	switch kind(t) {
	case T_SLICE:
		myfmt.Printf("  pushq $0 # slice cap\n")
		myfmt.Printf("  pushq $0 # slice len\n")
		myfmt.Printf("  pushq $0 # slice ptr\n")
	case T_STRING:
		myfmt.Printf("  pushq $0 # string len\n")
		myfmt.Printf("  pushq $0 # string ptr\n")
	case T_INTERFACE:
		myfmt.Printf("  pushq $0 # interface data\n")
		myfmt.Printf("  pushq $0 # interface dtype\n")
	case T_INT, T_UINTPTR, T_UINT8, T_POINTER, T_BOOL:
		myfmt.Printf("  pushq $0 # %s zero value\n", kind(t))
	case T_STRUCT:
		var structSize = getSizeOfType(t)
		emitComment(2, "zero value of a struct. size=%d (allocating on heap)\n", structSize)
		emitCallMalloc(structSize)
	default:
		panic2(__func__, "TBI:"+kind(t))
	}
}

func emitLen(arg astExpr) {
	emitComment(2, "[%s] begin\n", __func__)
	switch kind(getTypeOfExpr(arg)) {
	case T_ARRAY:
		var typ = getTypeOfExpr(arg)
		var arrayType = expr2ArrayType(typ.e)
		emitExpr(arrayType.Len, nil)
	case T_SLICE:
		emitExpr(arg, nil)
		emitPopSlice()
		myfmt.Printf("  pushq %%rcx # len\n")
	case T_STRING:
		emitExpr(arg, nil)
		emitPopString()
		myfmt.Printf("  pushq %%rcx # len\n")
	default:
		throw(kind(getTypeOfExpr(arg)))
	}
	emitComment(2, "[%s] end\n", __func__)
}

func emitCap(arg astExpr) {
	switch kind(getTypeOfExpr(arg)) {
	case T_ARRAY:
		var typ = getTypeOfExpr(arg)
		var arrayType = expr2ArrayType(typ.e)
		emitExpr(arrayType.Len, nil)
	case T_SLICE:
		emitExpr(arg, nil)
		emitPopSlice()
		myfmt.Printf("  pushq %%rdx # cap\n")
	case T_STRING:
		panic("cap() cannot accept string type")
	default:
		throw(kind(getTypeOfExpr(arg)))
	}
}

func emitCallMalloc(size int) {
	myfmt.Printf("  pushq $%d\n", size)
	// call malloc and return pointer
	var resultList = []*astField{
		&astField{
			Type:    tUintptr.e,
		},
	}
	myfmt.Printf("  callq runtime.malloc\n") // no need to invert args orders
	emitRevertStackPointer(intSize)
	emitReturnedValue(resultList)
}

func emitStructLiteral(e *astCompositeLit) {
	// allocate heap area with zero value
	myfmt.Printf("  # Struct literal\n")
	var structType = e2t(e.Type)
	emitZeroValue(structType) // push address of the new storage
	var kvExpr *astKeyValueExpr
	for i, elm := range e.Elts {
		kvExpr = expr2KeyValueExpr(elm)
		fieldName := expr2Ident(kvExpr.Key)
		emitComment(2,"  - [%d] : key=%s, value=%s\n", i, fieldName.Name, dtypeOf(kvExpr.Value))
		var field = lookupStructField(getStructTypeSpec(structType), fieldName.Name)
		var fieldType = e2t(field.Type)
		var fieldOffset = getStructFieldOffset(field)
		// push lhs address
		emitPushStackTop(tUintptr, "address of struct heaad")
		emitAddConst(fieldOffset, "address of struct field")
		// push rhs value
		ctx := &evalContext{
			_type: fieldType,
		}
		emitExprIfc(kvExpr.Value, ctx)
		// assign
		emitStore(fieldType, true, false)
	}
}

func emitArrayLiteral(arrayType *astArrayType, arrayLen int, elts []astExpr) {
	var elmType = e2t(arrayType.Elt)
	var elmSize = getSizeOfType(elmType)
	var memSize = elmSize * arrayLen
	emitCallMalloc(memSize) // push
	for i, elm := range elts {
		// emit lhs
		emitPushStackTop(tUintptr, "malloced address")
		emitAddConst(elmSize*i, "malloced address + elmSize * index ("+strconv.Itoa(i)+")")
		ctx := &evalContext{
			_type: elmType,
		}
		emitExprIfc(elm, ctx)
		emitStore(elmType, true, false)
	}
}

func emitInvertBoolValue() {
	emitPopBool("")
	myfmt.Printf("  xor $1, %%rax\n")
	myfmt.Printf("  pushq %%rax\n")
}

func emitTrue() {
	myfmt.Printf("  pushq $1 # true\n")
}

func emitFalse() {
	myfmt.Printf("  pushq $0 # false\n")
}

type Arg struct {
	e      astExpr
	t      *Type // expected type
	offset int
}

func emitArgs(args []*Arg) int {
	emitComment(2, "emitArgs len=%d\n", len(args))
	var totalPushedSize int
	//var arg astExpr
	for _, arg := range args {
		var t *Type
		if arg.t != nil {
			t = arg.t
		} else {
			t = getTypeOfExpr(arg.e)
		}
		arg.offset = totalPushedSize
		totalPushedSize = totalPushedSize + getPushSizeOfType(t)
	}
	myfmt.Printf("  subq $%d, %%rsp # for args\n", totalPushedSize)
	for _, arg := range args {
		ctx := &evalContext{
			_type: arg.t,
		}
		emitExprIfc(arg.e, ctx)
	}
	myfmt.Printf("  addq $%d, %%rsp # for args\n", totalPushedSize)

	for _, arg := range args {
		var t *Type
		if arg.t != nil {
			t = arg.t
		} else {
			t = getTypeOfExpr(arg.e)
		}
		switch kind(t) {
		case T_BOOL, T_INT, T_UINT8, T_POINTER, T_UINTPTR:
			myfmt.Printf("  movq %d-8(%%rsp) , %%rax # load\n", -arg.offset)
			myfmt.Printf("  movq %%rax, %d(%%rsp) # store\n", +arg.offset)
		case T_STRING, T_INTERFACE:
			myfmt.Printf("  movq %d-16(%%rsp), %%rax\n", -arg.offset)
			myfmt.Printf("  movq %d-8(%%rsp), %%rcx\n", -arg.offset)
			myfmt.Printf("  movq %%rax, %d(%%rsp)\n", +arg.offset)
			myfmt.Printf("  movq %%rcx, %d+8(%%rsp)\n", +arg.offset)
		case T_SLICE:
			myfmt.Printf("  movq %d-24(%%rsp), %%rax\n", -arg.offset) // arg1: slc.ptr
			myfmt.Printf("  movq %d-16(%%rsp), %%rcx\n", -arg.offset) // arg1: slc.len
			myfmt.Printf("  movq %d-8(%%rsp), %%rdx\n", -arg.offset)  // arg1: slc.cap
			myfmt.Printf("  movq %%rax, %d+0(%%rsp)\n", +arg.offset)  // arg1: slc.ptr
			myfmt.Printf("  movq %%rcx, %d+8(%%rsp)\n", +arg.offset)  // arg1: slc.len
			myfmt.Printf("  movq %%rdx, %d+16(%%rsp)\n", +arg.offset) // arg1: slc.cap
		default:
			throw(kind(t))
		}
	}

	return totalPushedSize
}

func prepareArgs(funcType *astFuncType, receiver astExpr, eArgs []astExpr, expandElipsis bool) []*Arg {
	if funcType == nil {
		panic("no funcType")
	}
	var params = funcType.Params.List
	var variadicArgs []astExpr
	var variadicElmType astExpr
	var args []*Arg
	var param *astField
	var arg *Arg
	lenParams := len(params)
	for argIndex, eArg := range eArgs {
		emitComment(2, "[%s][*astIdent][default] loop idx %d, len params %d\n", __func__, argIndex, lenParams)
		if argIndex < lenParams {
			param = params[argIndex]
			if isExprEllipsis(param.Type) {
				ellipsis := expr2Ellipsis(param.Type)
				variadicElmType = ellipsis.Elt
				variadicArgs = make([]astExpr, 0, 20)
			}
		}
		if variadicElmType != nil && !expandElipsis {
			variadicArgs = append(variadicArgs, eArg)
			continue
		}

		var paramType = e2t(param.Type)
		arg = &Arg{}
		arg.e = eArg
		arg.t = paramType
		args = append(args, arg)
	}

	if variadicElmType != nil && !expandElipsis {
		// collect args as a slice
		var sliceType = &astArrayType{}
		sliceType.Elt = variadicElmType
		var eSliceType = newExpr(sliceType)
		var sliceLiteral = &astCompositeLit{}
		sliceLiteral.Type = eSliceType
		sliceLiteral.Elts = variadicArgs
		var _arg = &Arg{
			e : newExpr(sliceLiteral),
			t : e2t(eSliceType),
		}
		args = append(args, _arg)
	} else if len(args) < len(params) {
		// Add nil as a variadic arg
		emitComment(2, "len(args)=%d, len(params)=%d\n", len(args), len(params))
		var param = params[len(args)]
		if param == nil {
			panic2(__func__, "param should not be nil")
		}
		if param.Type == nil {
			panic2(__func__, "param.Type should not be nil")
		}
		assert(isExprEllipsis(param.Type), "internal error", __func__)

		var _arg = &Arg{}
		_arg.e = eNil
		_arg.t = e2t(param.Type)
		args = append(args, _arg)
	}

	if receiver != nil {
		var receiverAndArgs []*Arg = []*Arg{
			&Arg{
				e: receiver,
				t: getTypeOfExpr(receiver),
			},
		}

		for _, a := range args {
			receiverAndArgs = append(receiverAndArgs, a)
		}
		return receiverAndArgs
	}

	return args
}

func emitCall(symbol string, args []*Arg, results []*astField) {
	emitComment(2, "[%s] %s\n", __func__, symbol)
	var totalPushedSize = emitArgs(args)
	myfmt.Printf("  callq %s\n", symbol)
	emitRevertStackPointer(totalPushedSize)
	emitReturnedValue(results)
}

func emitReturnedValue(resultList []*astField) {
	switch len(resultList) {
	case 0:
		// do nothing
	case 1:
		var retval0 = resultList[0]
		var knd = kind(e2t(retval0.Type))
		switch knd {
		case T_STRING:
			myfmt.Printf("  pushq %%rdi # str len\n")
			myfmt.Printf("  pushq %%rax # str ptr\n")
		case T_INTERFACE:
			myfmt.Printf("  pushq %%rdi # ifc data\n")
			myfmt.Printf("  pushq %%rax # ifc dtype\n")
		case T_BOOL, T_UINT8, T_INT, T_UINTPTR, T_POINTER:
			myfmt.Printf("  pushq %%rax\n")
		case T_SLICE:
			myfmt.Printf("  pushq %%rsi # slice cap\n")
			myfmt.Printf("  pushq %%rdi # slice len\n")
			myfmt.Printf("  pushq %%rax # slice ptr\n")
		default:
			panic2(__func__, "Unexpected kind="+knd)
		}
	default:
		panic2(__func__,"multipul returned values is not supported ")
	}
}

func emitFuncall(fun astExpr, eArgs []astExpr, hasEllissis bool) {
	var symbol string
	var receiver astExpr
	var funcType *astFuncType
	switch fn := fun.(type)  {
	case *astIdent:
		emitComment(2, "[%s][*astIdent]\n", __func__)
		var fnIdent = fn
		switch fnIdent.Obj {
		case gLen:
			var arg = eArgs[0]
			emitLen(arg)
			return
		case gCap:
			var arg = eArgs[0]
			emitCap(arg)
			return
		case gNew:
			var typeArg = e2t(eArgs[0])
			var size = getSizeOfType(typeArg)
			emitCallMalloc(size)
			return
		case gMake:
			var typeArg = e2t(eArgs[0])
			switch kind(typeArg) {
			case T_SLICE:
				// make([]T, ...)
				var arrayType = expr2ArrayType(typeArg.e)
				//assert(ok, "should be *ast.ArrayType")
				var elmSize = getSizeOfType(e2t(arrayType.Elt))
				var numlit = newNumberLiteral(elmSize)
				var eNumLit = newExpr(numlit)

				var args []*Arg = []*Arg{
					// elmSize
					&Arg{
						e: eNumLit,
						t: tInt,
					},
					// len
					&Arg{
						e: eArgs[1],
						t: tInt,
					},
					// cap
					&Arg{
						e: eArgs[2],
						t: tInt,
					},
				}


				var resultList = []*astField{
					&astField{
						Type: generalSlice,
					},
				}
				emitCall("runtime.makeSlice", args, resultList)
				return
			default:
				panic2(__func__, "TBI")
			}

			return
		case gAppend:
			var sliceArg = eArgs[0]
			var elemArg = eArgs[1]
			var elmType = getElementTypeOfListType(getTypeOfExpr(sliceArg))
			var elmSize = getSizeOfType(elmType)

			var args []*Arg = []*Arg{
				// slice
				&Arg{
					e: sliceArg,
				},
				// element
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
				panic2(__func__, "Unexpected elmSize")
			}
			var resultList = []*astField{
				&astField{
					Type: generalSlice,
				},
			}
			emitCall(symbol, args, resultList)
			return
		case gPanic:
			symbol = "runtime.panic"
			_args := []*Arg{&Arg{
				e: eArgs[0],
			}}
			emitCall(symbol, _args, nil)
			return
		}

		if fn.Name == "print" {
			emitExpr(eArgs[0], nil)
			myfmt.Printf("  callq runtime.printstring\n")
			myfmt.Printf("  addq $%d, %%rsp # revert \n", 16)
			return
		}

		if fn.Name == "makeSlice1" || fn.Name == "makeSlice8" || fn.Name == "makeSlice16" || fn.Name == "makeSlice24" {
			fn.Name = "makeSlice"
		}
		// general function call
		symbol = getPackageSymbol(pkg.name, fn.Name)
		emitComment(2, "[%s][*astIdent][default] start\n", __func__)
		if pkg.name == "os" && fn.Name == "runtime_args" {
			symbol = "runtime.runtime_args"
		} else if pkg.name == "os" && fn.Name == "runtime_getenv" {
			symbol = "runtime.runtime_getenv"
		}

		var obj = fn.Obj
		if obj.Decl == nil {
			panic2(__func__, "[*astCallExpr] decl is nil")
		}

		var fndecl *astFuncDecl // = decl.funcDecl
		var isFuncDecl bool
		fndecl, isFuncDecl = obj.Decl.(*astFuncDecl)
		if !isFuncDecl || fndecl == nil {
			panic2(__func__, "[*astCallExpr] fndecl is nil")
		}
		if fndecl.Type == nil {
			panic2(__func__, "[*astCallExpr] fndecl.Type is nil")
		}
		funcType = fndecl.Type
	case *astSelectorExpr:
		var selectorExpr = fn
		if !isExprIdent(selectorExpr.X) {
			panic2(__func__, "TBI selectorExpr.X="+ dtypeOf(selectorExpr.X))
		}
		xIdent := expr2Ident(selectorExpr.X)
		symbol = xIdent.Name + "." + selectorExpr.Sel.Name
		switch symbol {
		case "unsafe.Pointer":
			emitExpr(eArgs[0], nil)
			return
		default:
			// Assume method call
			fn := selectorExpr
			xIdent := expr2Ident(fn.X)
			if xIdent.Obj == nil {
				throw("xIdent.Obj  should not be nil:" + xIdent.Name)
			}
			if xIdent.Obj.Kind == astPkg {
				// pkg.Sel()
				funcdecl := lookupForeignFunc(xIdent.Name, fn.Sel.Name)
				funcType = funcdecl.Type
			} else {
				receiver = selectorExpr.X
				var receiverType = getTypeOfExpr(receiver)
				var method = lookupMethod(receiverType, selectorExpr.Sel)
				funcType = method.funcType
				symbol = getMethodSymbol(method)
			}
		}
	case *astParenExpr:
		panic2(__func__, "[astParenExpr] TBI ")
	default:
		panic2(__func__, "TBI fun.dtype="+ dtypeOf(fun))
	}

	var args = prepareArgs(funcType, receiver, eArgs, hasEllissis)
	var resultList []*astField
	if funcType.Results != nil {
		resultList = funcType.Results.List
	}
	emitCall(symbol, args, resultList)
}

func emitNil(targetType *Type) {
	if targetType == nil {
		panic2(__func__, "Type is required to emit nil")
	}
	switch kind(targetType) {
	case T_SLICE, T_POINTER, T_INTERFACE:
		emitZeroValue(targetType)
	default:
		panic2(__func__, "Unexpected kind="+kind(targetType))
	}
}

func emitNamedConst(ident *astIdent, ctx *evalContext) {
	var valSpec *astValueSpec
	var ok bool
	valSpec, ok = ident.Obj.Decl.(*astValueSpec)
	assert(ok, "valSpec should not be nil", __func__)
	assert(valSpec != nil, "valSpec should not be nil", __func__)
	assert(valSpec.Value != nil, "valSpec should not be nil", __func__)
	assert(isExprBasicLit(valSpec.Value), "const value should be a literal", __func__)
	emitExprIfc(valSpec.Value, ctx)
}

type okContext struct {
	needMain bool
	needOk   bool
}

type evalContext struct {
	okContext *okContext
	_type     *Type
}

func emitExpr(expr astExpr, ctx *evalContext) bool {
	var isNilObject bool
	emitComment(2, "[emitExpr] dtype=%s\n", dtypeOf(expr))
	switch e := expr.(type)  {
	case *astIdent:
		var ident = e
		if ident.Obj == nil {
			panic2(__func__, "ident unresolved:"+ident.Name)
		}
		switch e.Obj {
		case gTrue:
			emitTrue()
		case gFalse:
			emitFalse()
		case gNil:
			if ctx._type == nil {
				panic2(__func__, "context of nil is not passed")
			}
			emitNil(ctx._type)
			isNilObject = true
		default:
			switch ident.Obj.Kind {
			case astVar:
				emitAddr(expr)
				var t = getTypeOfExpr(expr)
				emitLoad(t)
			case astCon:
				emitNamedConst(ident, ctx)
			case astTyp:
				panic2(__func__, "[*astIdent] Kind Typ should not come here")
			default:
				panic2(__func__, "[*astIdent] unknown Kind="+ident.Obj.Kind+" Name="+ident.Obj.Name)
			}
		}
	case *astIndexExpr:
		emitAddr(expr)
		emitLoad(getTypeOfExpr(expr))
	case *astStarExpr:
		emitAddr(expr)
		emitLoad(getTypeOfExpr(expr))
	case *astSelectorExpr:
		x := e.X
		if isExprIdent(x) && expr2Ident(x).Obj.Kind == astPkg {
			ident := lookupForeignVar(expr2Ident(x).Name, e.Sel.Name)
			e := newExpr(ident)
			emitExpr(e, ctx)
		} else {
			emitAddr(expr)
			emitLoad(getTypeOfExpr(expr))
		}
	case *astCallExpr:
		var fun = e.Fun
		emitComment(2, "[%s][*astCallExpr]\n", __func__)
		if isType(fun) {
			emitConversion(e2t(fun), e.Args[0])
		} else {
			emitFuncall(fun, e.Args, e.Ellipsis)
		}
	case *astParenExpr:
		emitExpr(e.X, ctx)
	case *astBasicLit:
		//		emitComment(0, "basicLit.Kind = %s \n", expr.basicLit.Kind)
		basicLit := expr2BasicLit(expr)
		switch basicLit.Kind {
		case "INT":
			var ival = strconv.Atoi(basicLit.Value)
			myfmt.Printf("  pushq $%d # number literal\n", ival)
		case "STRING":
			var sl = getStringLiteral(basicLit)
			if sl.strlen == 0 {
				// zero value
				emitZeroValue(tString)
			} else {
				myfmt.Printf("  pushq $%d # str len\n", sl.strlen)
				myfmt.Printf("  leaq %s, %%rax # str ptr\n", sl.label)
				myfmt.Printf("  pushq %%rax # str ptr\n")
			}
		case "CHAR":
			var val = basicLit.Value
			var char = val[1]
			if val[1] == '\\' {
				switch val[2] {
				case '\'':
					char = '\''
				case 'n':
					char = '\n'
				case '\\':
					char = '\\'
				case 't':
					char = '\t'
				case 'r':
					char = '\r'
				}
			}
			myfmt.Printf("  pushq $%d # convert char literal to int\n", int(char))
		default:
			panic2(__func__, "[*astBasicLit] TBI : "+basicLit.Kind)
		}
	case *astSliceExpr:
		var list = e.X
		var listType = getTypeOfExpr(list)

		// For convenience, any of the indices may be omitted.
		// A missing low index defaults to zero;
		var low astExpr
		if e.Low != nil {
			low = e.Low
		} else {
			low = eZeroInt
		}

		// a missing high index defaults to the length of the sliced operand:
		// @TODO

		switch kind(listType) {
		case T_SLICE, T_ARRAY:
			if e.Max == nil {
				// new cap = cap(operand) - low
				emitCap(e.X)
				emitExpr(low, nil)
				myfmt.Printf("  popq %%rcx # low\n")
				myfmt.Printf("  popq %%rax # orig_cap\n")
				myfmt.Printf("  subq %%rcx, %%rax # orig_cap - low\n")
				myfmt.Printf("  pushq %%rax # new cap\n")

				// new len = high - low
				if e.High != nil {
					emitExpr(e.High, nil)
				} else {
					// high = len(orig)
					emitLen(e.X)
				}
				emitExpr(low, nil)
				myfmt.Printf("  popq %%rcx # low\n")
				myfmt.Printf("  popq %%rax # high\n")
				myfmt.Printf("  subq %%rcx, %%rax # high - low\n")
				myfmt.Printf("  pushq %%rax # new len\n")
			} else {
				// new cap = max - low
				emitExpr(e.Max, nil)
				emitExpr(low, nil)
				myfmt.Printf("  popq %%rcx # low\n")
				myfmt.Printf("  popq %%rax # max\n")
				myfmt.Printf("  subq %%rcx, %%rax # new cap = max - low\n")
				myfmt.Printf("  pushq %%rax # new cap\n")
				// new len = high - low
				emitExpr(e.High, nil)
				emitExpr(low, nil)
				myfmt.Printf("  popq %%rcx # low\n")
				myfmt.Printf("  popq %%rax # high\n")
				myfmt.Printf("  subq %%rcx, %%rax # new len = high - low\n")
				myfmt.Printf("  pushq %%rax # new len\n")
			}
		case T_STRING:
			// new len = high - low
			if e.High != nil {
				emitExpr(e.High, nil)
			} else {
				emitLen(e.X)
			}
			emitExpr(low, nil)
			myfmt.Printf("  popq %%rcx # low\n")
			myfmt.Printf("  popq %%rax # high\n")
			myfmt.Printf("  subq %%rcx, %%rax # high - low\n")
			myfmt.Printf("  pushq %%rax # len\n")
			// no cap
		default:
			panic2(__func__, "Unknown kind="+kind(listType))
		}

		emitExpr(low, nil)
		var elmType = getElementTypeOfListType(listType)
		emitListElementAddr(list, elmType)
	case *astUnaryExpr:
		emitComment(2, "[DEBUG] unary op = %s\n", e.Op)
		switch e.Op {
		case "+":
			emitExpr(e.X, nil)
		case "-":
			emitExpr(e.X, nil)
			myfmt.Printf("  popq %%rax # e.X\n")
			myfmt.Printf("  imulq $-1, %%rax\n")
			myfmt.Printf("  pushq %%rax\n")
		case "&":
			emitAddr(e.X)
		case "!":
			emitExpr(e.X, nil)
			emitInvertBoolValue()
		default:
			panic2(__func__, "TBI:astUnaryExpr:"+e.Op)
		}
	case *astBinaryExpr:
		binaryExpr := expr2BinaryExpr(expr)

		if kind(getTypeOfExpr(binaryExpr.X)) == T_STRING {
			var args []*Arg
			var argX = &Arg{}
			var argY = &Arg{}
			argX.e = binaryExpr.X
			argY.e = binaryExpr.Y
			args = append(args, argX)
			args = append(args, argY)
			switch binaryExpr.Op {
			case "+":
				var resultList = []*astField{
					&astField{
						Type:    tString.e,
					},
				}
				emitCall("runtime.catstrings", args, resultList)
			case "==":
				emitArgs(args)
				emitCompEq(getTypeOfExpr(binaryExpr.X))
			case "!=":
				emitArgs(args)
				emitCompEq(getTypeOfExpr(binaryExpr.X))
				emitInvertBoolValue()
			default:
				panic2(__func__, "[emitExpr][*astBinaryExpr] string : TBI T_STRING")
			}
		} else {
			switch binaryExpr.Op {
			case "&&":
				labelid++
				var labelExitWithFalse = myfmt.Sprintf(".L.%d.false", labelid)
				var labelExit = myfmt.Sprintf(".L.%d.exit", labelid)
				emitExpr(binaryExpr.X, nil) // left
				emitPopBool("left")
				myfmt.Printf("  cmpq $1, %%rax\n")
				// exit with false if left is false
				myfmt.Printf("  jne %s\n", labelExitWithFalse)

				// if left is true, then eval right and exit
				emitExpr(binaryExpr.Y, nil) // right
				myfmt.Printf("  jmp %s\n", labelExit)

				myfmt.Printf("  %s:\n", labelExitWithFalse)
				emitFalse()
				myfmt.Printf("  %s:\n", labelExit)
			case "||":
				labelid++
				var labelExitWithTrue = myfmt.Sprintf(".L.%d.true", labelid)
				var labelExit = myfmt.Sprintf(".L.%d.exit", labelid)
				emitExpr(binaryExpr.X, nil) // left
				emitPopBool("left")
				myfmt.Printf("  cmpq $1, %%rax\n")
				// exit with true if left is true
				myfmt.Printf("  je %s\n", labelExitWithTrue)

				// if left is false, then eval right and exit
				emitExpr(binaryExpr.Y, nil) // right
				myfmt.Printf("  jmp %s\n", labelExit)

				myfmt.Printf("  %s:\n", labelExitWithTrue)
				emitTrue()
				myfmt.Printf("  %s:\n", labelExit)
			case "+":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil)   // right
				myfmt.Printf("  popq %%rcx # right\n")
				myfmt.Printf("  popq %%rax # left\n")
				myfmt.Printf("  addq %%rcx, %%rax\n")
				myfmt.Printf("  pushq %%rax\n")
			case "-":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil)   // right
				myfmt.Printf("  popq %%rcx # right\n")
				myfmt.Printf("  popq %%rax # left\n")
				myfmt.Printf("  subq %%rcx, %%rax\n")
				myfmt.Printf("  pushq %%rax\n")
			case "*":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil)   // right
				myfmt.Printf("  popq %%rcx # right\n")
				myfmt.Printf("  popq %%rax # left\n")
				myfmt.Printf("  imulq %%rcx, %%rax\n")
				myfmt.Printf("  pushq %%rax\n")
			case "%":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil)   // right
				myfmt.Printf("  popq %%rcx # right\n")
				myfmt.Printf("  popq %%rax # left\n")
				myfmt.Printf("  movq $0, %%rdx # init %%rdx\n")
				myfmt.Printf("  divq %%rcx\n")
				myfmt.Printf("  movq %%rdx, %%rax\n")
				myfmt.Printf("  pushq %%rax\n")
			case "/":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil)   // right
				myfmt.Printf("  popq %%rcx # right\n")
				myfmt.Printf("  popq %%rax # left\n")
				myfmt.Printf("  movq $0, %%rdx # init %%rdx\n")
				myfmt.Printf("  divq %%rcx\n")
				myfmt.Printf("  pushq %%rax\n")
			case "==":
				var t = getTypeOfExpr(binaryExpr.X)
				emitExpr(binaryExpr.X, nil) // left
				ctx := &evalContext{_type: t}
				emitExpr(binaryExpr.Y, ctx)   // right
				emitCompEq(t)
			case "!=":
				var t = getTypeOfExpr(binaryExpr.X)
				emitExpr(binaryExpr.X, nil) // left
				ctx := &evalContext{_type: t}
				emitExpr(binaryExpr.Y, ctx)   // right
				emitCompEq(t)
				emitInvertBoolValue()
			case "<":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil)   // right
				emitCompExpr("setl")
			case "<=":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil)   // right
				emitCompExpr("setle")
			case ">":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil)   // right
				emitCompExpr("setg")
			case ">=":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil)   // right
				emitCompExpr("setge")
			default:
				panic2(__func__, "# TBI: binary operation for "+binaryExpr.Op)
			}
		}
	case *astCompositeLit:
		// slice , array, map or struct
		var k = kind(e2t(e.Type))
		switch k {
		case T_STRUCT:
			emitStructLiteral(e)
		case T_ARRAY:
			arrayType := expr2ArrayType(e.Type)
			var arrayLen = evalInt(arrayType.Len)
			emitArrayLiteral(arrayType, arrayLen, e.Elts)
		case T_SLICE:
			arrayType := expr2ArrayType(e.Type)
			var length = len(e.Elts)
			emitArrayLiteral(arrayType, length, e.Elts)
			emitPopAddress("malloc")
			myfmt.Printf("  pushq $%d # slice.cap\n", length)
			myfmt.Printf("  pushq $%d # slice.len\n", length)
			myfmt.Printf("  pushq %%rax # slice.ptr\n")
		default:
			panic2(__func__, "Unexpected kind="+k)
		}
	case *astTypeAssertExpr:
		emitExpr(expr2TypeAssertExpr(expr).X, nil)
		myfmt.Printf("  popq  %%rax # ifc.dtype\n")
		myfmt.Printf("  popq  %%rcx # ifc.data\n")
		myfmt.Printf("  pushq %%rax # ifc.data\n")
		typ := e2t(expr2TypeAssertExpr(expr).Type)
		sType := serializeType(typ)
		_id := getTypeId(sType)
		typeSymbol := typeIdToSymbol(_id)
		// check if type matches
		myfmt.Printf("  leaq %s(%%rip), %%rax # ifc.dtype\n", typeSymbol)
		myfmt.Printf("  pushq %%rax           # ifc.dtype\n")

		emitCompExpr("sete") // this pushes 1 or 0 in the end
		emitPopBool("type assertion ok value")
		myfmt.Printf("  cmpq $1, %%rax\n")

		labelid++
		labelTypeAssertionEnd := myfmt.Sprintf(".L.end_type_assertion.%d", labelid)
		labelElse := myfmt.Sprintf(".L.unmatch.%d", labelid)
		myfmt.Printf("  jne %s # jmp if false\n", labelElse)

		// if matched
		if ctx.okContext != nil {
			emitComment(2, " double value context\n")
			if ctx.okContext.needMain {
				emitExpr(expr2TypeAssertExpr(expr).X, nil)
				myfmt.Printf("  popq %%rax # garbage\n")
				emitLoad(e2t(expr2TypeAssertExpr(expr).Type)) // load dynamic data
			}
			if ctx.okContext.needOk {
				myfmt.Printf("  pushq $1 # ok = true\n")
			}
		} else {
			emitComment(2, " single value context\n")
			emitExpr(expr2TypeAssertExpr(expr).X, nil)
			myfmt.Printf("  popq %%rax # garbage\n")
			emitLoad(e2t(expr2TypeAssertExpr(expr).Type)) // load dynamic data
		}

		// exit
		myfmt.Printf("  jmp %s\n", labelTypeAssertionEnd)

		// if not matched
		myfmt.Printf("  %s:\n", labelElse)
		if ctx.okContext != nil {
			emitComment(2, " double value context\n")
			if ctx.okContext.needMain {
				emitZeroValue(typ)
			}
			if ctx.okContext.needOk {
				myfmt.Printf("  pushq $0 # ok = false\n")
			}
		} else {
			emitComment(2, " single value context\n")
			emitZeroValue(typ)
		}

		myfmt.Printf("  %s:\n", labelTypeAssertionEnd)
	default:
		panic2(__func__, "[emitExpr] `TBI:"+dtypeOf(expr))
	}

	return isNilObject
}

// convert stack top value to interface
func emitConvertToInterface(fromType *Type) {
	emitComment(2, "ConversionToInterface\n")
	memSize := getSizeOfType(fromType)
	// copy data to heap
	emitCallMalloc(memSize)
	emitStore(fromType, false, true) // heap addr pushed
	// push type id
	emitDtypeSymbol(fromType)
}

func emitExprIfc(expr astExpr, ctx *evalContext) {
	isNilObj := emitExpr(expr, ctx)
	if !isNilObj && ctx != nil && ctx._type != nil && isInterface(ctx._type) && !isInterface(getTypeOfExpr(expr)) {
		emitConvertToInterface(getTypeOfExpr(expr))
	}
}

var typeMap []*typeEntry
type typeEntry struct {
	serialized string
	id int
}

var typeId int = 1

func typeIdToSymbol(id int) string {
	return "dtype." + strconv.Itoa(id)
}

func getTypeId(serialized string) int {
	for _, te := range typeMap{
		if te.serialized == serialized {
			return te.id
		}
	}
	r := typeId
	te := &typeEntry{
		serialized: serialized,
		id:         typeId,
	}
	typeMap = append(typeMap, te)
	typeId++
	return r
}

func emitDtypeSymbol(t *Type) {
	str := serializeType(t)
	typeId := getTypeId(str)
	typeSymbol := typeIdToSymbol(typeId)
	myfmt.Printf("  leaq %s(%%rip), %%rax # type symbol \"%s\"\n", typeSymbol, str)
	myfmt.Printf("  pushq %%rax           # type symbol\n")
}

func newNumberLiteral(x int) *astBasicLit {
	var r = &astBasicLit{}
	r.Kind = "INT"
	r.Value = strconv.Itoa(x)
	return r
}

func emitListElementAddr(list astExpr, elmType *Type) {
	emitListHeadAddr(list)
	emitPopAddress("list head")
	myfmt.Printf("  popq %%rcx # index id\n")
	myfmt.Printf("  movq $%d, %%rdx # elm size\n", getSizeOfType(elmType))
	myfmt.Printf("  imulq %%rdx, %%rcx\n")
	myfmt.Printf("  addq %%rcx, %%rax\n")
	myfmt.Printf("  pushq %%rax # addr of element\n")
}

func emitCompEq(t *Type) {
	switch kind(t) {
	case T_STRING:
		var resultList = []*astField{
			&astField{
				Type:    tBool.e,
			},
		}
		myfmt.Printf("  callq runtime.cmpstrings\n")
		emitRevertStackPointer(stringSize * 2)
		emitReturnedValue(resultList)
	case T_INTERFACE:
		var resultList = []*astField{
			&astField{
				Type:    tBool.e,
			},
		}
		myfmt.Printf("  callq runtime.cmpinterface\n")
		emitRevertStackPointer(interfaceSize * 2)
		emitReturnedValue(resultList)
	case T_INT, T_UINT8, T_UINT16, T_UINTPTR, T_POINTER:
		emitCompExpr("sete")
	case T_SLICE:
		emitCompExpr("sete") // @FIXME this is not correct
	default:
		panic2(__func__, "Unexpected kind="+kind(t))
	}
}

//@TODO handle larger types than int
func emitCompExpr(inst string) {
	myfmt.Printf("  popq %%rcx # right\n")
	myfmt.Printf("  popq %%rax # left\n")
	myfmt.Printf("  cmpq %%rcx, %%rax\n")
	myfmt.Printf("  %s %%al\n", inst)
	myfmt.Printf("  movzbq %%al, %%rax\n") // true:1, false:0
	myfmt.Printf("  pushq %%rax\n")
}

func emitPop(knd string) {
	switch knd {
	case T_SLICE:
		emitPopSlice()
	case T_STRING:
		emitPopString()
	case T_INTERFACE:
		emitPopInterFace()
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		emitPopPrimitive(knd)
	case T_UINT16:
		emitPopPrimitive(knd)
	case T_UINT8:
		emitPopPrimitive(knd)
	case T_STRUCT, T_ARRAY:
		emitPopPrimitive(knd)
	default:
		panic("TBI:" + knd)
	}
}

func emitStore(t *Type, rhsTop bool, pushLhs bool) {
	knd := kind(t)
	emitComment(2, "emitStore(%s)\n", knd)
	if rhsTop {
		emitPop(knd) // rhs
		myfmt.Printf("  popq %%rsi # lhs addr\n")
	} else {
		myfmt.Printf("  popq %%rsi # lhs addr\n")
		emitPop(knd) // rhs
	}
	if pushLhs {
		myfmt.Printf("  pushq %%rsi # lhs addr\n")
	}
	switch knd {
	case T_SLICE:
		myfmt.Printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", 0)
		myfmt.Printf("  movq %%rcx, %d(%%rsi) # len to len\n", 8)
		myfmt.Printf("  movq %%rdx, %d(%%rsi) # cap to cap\n", 16)
	case T_STRING:
		myfmt.Printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", 0)
		myfmt.Printf("  movq %%rcx, %d(%%rsi) # len to len\n", 8)
	case T_INTERFACE:
		myfmt.Printf("  movq %%rax, %d(%%rsi) # store dtype\n", 0)
		myfmt.Printf("  movq %%rcx, %d(%%rsi) # store data\n", 8)
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		myfmt.Printf("  movq %%rax, (%%rsi) # assign\n")
	case T_UINT16:
		myfmt.Printf("  movw %%ax, (%%rsi) # assign word\n")
	case T_UINT8:
		myfmt.Printf("  movb %%al, (%%rsi) # assign byte\n")
	case T_STRUCT, T_ARRAY:
		myfmt.Printf("  pushq $%d # size\n", getSizeOfType(t))
		myfmt.Printf("  pushq %%rsi # dst lhs\n")
		myfmt.Printf("  pushq %%rax # src rhs\n")
		myfmt.Printf("  callq runtime.memcopy\n")
		emitRevertStackPointer(ptrSize*2 + intSize)
	default:
		panic2(__func__, "TBI:"+kind(t))
	}
}

func isBlankIdentifier(e astExpr) bool {
	if !isExprIdent(e) {
		return false
	}
	return expr2Ident(e).Name == "_"
}

func emitAssignWithOK(lhss []astExpr, rhs astExpr) {
	lhsMain := lhss[0]
	lhsOK := lhss[1]

	needMain := !isBlankIdentifier(lhsMain)
	needOK := !isBlankIdentifier(lhsOK)
	emitComment(2, "Assignment: emitAssignWithOK rhs\n")
	ctx := &evalContext{
		okContext: &okContext{
			needMain: needMain,
			needOk:   needOK,
		},
	}
	emitExprIfc(rhs, ctx) // {push data}, {push bool}
	if needOK {
		emitComment(2, "Assignment: ok variable\n")
		emitAddr(lhsOK)
		emitStore(getTypeOfExpr(lhsOK), false, false)
	}

	if needMain {
		emitAddr(lhsMain)
		emitComment(2, "Assignment: emitStore(getTypeOfExpr(lhs))\n")
		emitStore(getTypeOfExpr(lhsMain), false, false)
	}
}

func emitAssign(lhs astExpr, rhs astExpr) {
	emitComment(2, "Assignment: emitAddr(lhs:%s)\n", dtypeOf(lhs))
	emitAddr(lhs)
	emitComment(2, "Assignment: emitExpr(rhs)\n")
	ctx := &evalContext{
		_type: getTypeOfExpr(lhs),
	}
	emitExprIfc(rhs, ctx)
	emitStore(getTypeOfExpr(lhs), true, false)
}

func emitStmt(stmt astStmt) {
	emitComment(2, "== Statement %s ==\n", dtypeOf(stmt))
	switch s := stmt.(type)  {
	case *astBlockStmt:
		for _, stmt2 := range s.List {
			emitStmt(stmt2)
		}
	case *astExprStmt:
		emitExpr(s.X, nil)
	case *astDeclStmt:
		var decl astDecl = s.Decl
		var genDecl *astGenDecl
		var isGenDecl bool
		genDecl, isGenDecl = decl.(*astGenDecl)
		if !isGenDecl {
			panic2(__func__, "[*astDeclStmt] internal error")
		}

		var valSpec *astValueSpec
		var ok bool
		valSpec, ok = genDecl.Spec.(*astValueSpec)
		assert(ok, "should be ok", __func__)
		var t = e2t(valSpec.Type)
		var ident = valSpec.Name
		var lhs = newExpr(ident)
		var rhs astExpr
		if valSpec.Value == nil {
			emitComment(2, "lhs addresss\n")
			emitAddr(lhs)
			emitComment(2, "emitZeroValue for %s\n", dtypeOf(t.e))
			emitZeroValue(t)
			emitComment(2, "Assignment: zero value\n")
			emitStore(t, true, false)
		} else {
			rhs = valSpec.Value
			emitAssign(lhs, rhs)
		}

		//var valueSpec *astValueSpec = genDecl.Specs[0]
		//var obj *astObject = valueSpec.Name.Obj
		//var typ astExpr = valueSpec.Type
		//mylib.Printf("[emitStmt] TBI declSpec:%s\n", valueSpec.Name.Name)
		//os.Exit(1)

	case *astAssignStmt:
		switch s.Tok {
		case "=":
		case ":=":
		default:
		}
		var lhs = s.Lhs[0]
		var rhs = s.Rhs[0]
		if len(s.Lhs) == 2 && isExprTypeAssertExpr(rhs) {
			emitAssignWithOK(s.Lhs, rhs)
		} else {
			emitAssign(lhs, rhs)
		}
	case *astReturnStmt:
		node := s.node
		funcType := node.fnc.funcType
		if len(s.Results) == 0 {
			myfmt.Printf("  leave\n")
			myfmt.Printf("  ret\n")
		} else if len(s.Results) == 1 {
			targetType := e2t(funcType.Results.List[0].Type)
			ctx := &evalContext{
				_type:     targetType,
			}
			emitExprIfc(s.Results[0], ctx)
			var knd = kind(targetType)
			switch knd {
			case T_BOOL, T_UINT8, T_INT, T_UINTPTR, T_POINTER:
				myfmt.Printf("  popq %%rax # return 64bit\n")
			case T_STRING, T_INTERFACE:
				myfmt.Printf("  popq %%rax # return string (head)\n")
				myfmt.Printf("  popq %%rdi # return string (tail)\n")
			case T_SLICE:
				myfmt.Printf("  popq %%rax # return string (head)\n")
				myfmt.Printf("  popq %%rdi # return string (body)\n")
				myfmt.Printf("  popq %%rsi # return string (tail)\n")
			default:
				panic2(__func__, "[*astReturnStmt] TBI:"+knd)
			}
			myfmt.Printf("  leave\n")
			myfmt.Printf("  ret\n")
		} else if len(s.Results) == 3 {
			// Special treatment to return a slice
			emitExpr(s.Results[2], nil) // @FIXME
			emitExpr(s.Results[1], nil) // @FIXME
			emitExpr(s.Results[0], nil) // @FIXME
			myfmt.Printf("  popq %%rax # return 64bit\n")
			myfmt.Printf("  popq %%rdi # return 64bit\n")
			myfmt.Printf("  popq %%rsi # return 64bit\n")
		} else {
			panic2(__func__, "[*astReturnStmt] TBI\n")
		}
	case *astIfStmt:
		emitComment(2, "if\n")

		labelid++
		var labelEndif = ".L.endif." + strconv.Itoa(labelid)
		var labelElse = ".L.else." + strconv.Itoa(labelid)

		emitExpr(s.Cond, nil)
		emitPopBool("if condition")
		myfmt.Printf("  cmpq $1, %%rax\n")
		bodyStmt := newStmt(s.Body)
		if s.Else != nil {
			myfmt.Printf("  jne %s # jmp if false\n", labelElse)
			emitStmt(bodyStmt) // then
			myfmt.Printf("  jmp %s\n", labelEndif)
			myfmt.Printf("  %s:\n", labelElse)
			emitStmt(s.Else) // then
		} else {
			myfmt.Printf("  jne %s # jmp if false\n", labelEndif)
			emitStmt(bodyStmt) // then
		}
		myfmt.Printf("  %s:\n", labelEndif)
		emitComment(2, "end if\n")
	case *astForStmt:
		labelid++
		var labelCond = ".L.for.cond." + strconv.Itoa(labelid)
		var labelPost = ".L.for.post." + strconv.Itoa(labelid)
		var labelExit = ".L.for.exit." + strconv.Itoa(labelid)
		//forStmt, ok := mapForNodeToFor[s]
		//assert(ok, "map value should exist")
		s.labelPost = labelPost
		s.labelExit = labelExit

		if s.Init != nil {
			emitStmt(s.Init)
		}

		myfmt.Printf("  %s:\n", labelCond)
		if s.Cond != nil {
			emitExpr(s.Cond, nil)
			emitPopBool("for condition")
			myfmt.Printf("  cmpq $1, %%rax\n")
			myfmt.Printf("  jne %s # jmp if false\n", labelExit)
		}
		emitStmt(blockStmt2Stmt(s.Body))
		myfmt.Printf("  %s:\n", labelPost) // used for "continue"
		if s.Post != nil {
			emitStmt(s.Post)
		}
		myfmt.Printf("  jmp %s\n", labelCond)
		myfmt.Printf("  %s:\n", labelExit)
	case *astRangeStmt: // only for array and slice
		labelid++
		var labelCond = ".L.range.cond." + strconv.Itoa(labelid)
		var labelPost = ".L.range.post." + strconv.Itoa(labelid)
		var labelExit = ".L.range.exit." + strconv.Itoa(labelid)

		s.labelPost = labelPost
		s.labelExit = labelExit
		// initialization: store len(rangeexpr)
		emitComment(2, "ForRange Initialization\n")

		emitComment(2, "  assign length to lenvar\n")
		// lenvar = len(s.X)
		emitVariableAddr(s.lenvar)
		emitLen(s.X)
		emitStore(tInt, true, false)

		emitComment(2, "  assign 0 to indexvar\n")
		// indexvar = 0
		emitVariableAddr(s.indexvar)
		emitZeroValue(tInt)
		emitStore(tInt, true, false)

		// init key variable with 0
		if s.Key != nil {
			keyIdent := expr2Ident(s.Key)
			if keyIdent.Name != "_" {
				emitAddr(s.Key) // lhs
				emitZeroValue(tInt)
				emitStore(tInt, true, false)
			}
		}

		// Condition
		// if (indexvar < lenvar) then
		//   execute body
		// else
		//   exit
		emitComment(2, "ForRange Condition\n")
		myfmt.Printf("  %s:\n", labelCond)

		emitVariableAddr(s.indexvar)
		emitLoad(tInt)
		emitVariableAddr(s.lenvar)
		emitLoad(tInt)
		emitCompExpr("setl")
		emitPopBool(" indexvar < lenvar")
		myfmt.Printf("  cmpq $1, %%rax\n")
		myfmt.Printf("  jne %s # jmp if false\n", labelExit)

		emitComment(2, "assign list[indexvar] value variables\n")
		var elemType = getTypeOfExpr(s.Value)
		emitAddr(s.Value) // lhs

		emitVariableAddr(s.indexvar)
		emitLoad(tInt) // index value
		emitListElementAddr(s.X, elemType)

		emitLoad(elemType)
		emitStore(elemType, true, false)

		// Body
		emitComment(2, "ForRange Body\n")
		emitStmt(blockStmt2Stmt(s.Body))

		// Post statement: Increment indexvar and go next
		emitComment(2, "ForRange Post statement\n")
		myfmt.Printf("  %s:\n", labelPost)        // used for "continue"
		emitVariableAddr(s.indexvar) // lhs
		emitVariableAddr(s.indexvar) // rhs
		emitLoad(tInt)
		emitAddConst(1, "indexvar value ++")
		emitStore(tInt, true, false)

		if s.Key != nil {
			keyIdent := expr2Ident(s.Key)
			if keyIdent.Name != "_" {
				emitAddr(s.Key)              // lhs
				emitVariableAddr(s.indexvar) // rhs
				emitLoad(tInt)
				emitStore(tInt, true, false)
			}
		}

		myfmt.Printf("  jmp %s\n", labelCond)

		myfmt.Printf("  %s:\n", labelExit)

	case *astIncDecStmt:
		var addValue int
		switch s.Tok {
		case "++":
			addValue = 1
		case "--":
			addValue = -1
		default:
			panic2(__func__, "Unexpected Tok="+s.Tok)
		}
		emitAddr(s.X)
		emitExpr(s.X, nil)
		emitAddConst(addValue, "rhs ++ or --")
		emitStore(getTypeOfExpr(s.X), true, false)
	case *astSwitchStmt:
		labelid++
		var labelEnd = myfmt.Sprintf(".L.switch.%d.exit", labelid)
		if s.Tag == nil {
			panic2(__func__, "Omitted tag is not supported yet")
		}
		emitExpr(s.Tag, nil)
		var condType = getTypeOfExpr(s.Tag)
		var cases = s.Body.List
		emitComment(2, "[DEBUG] cases len=%d\n", len(cases))
		var labels = make([]string, len(cases), len(cases))
		var defaultLabel string
		emitComment(2, "Start comparison with cases\n")
		for i, c := range cases {
			emitComment(2, "CASES idx=%d\n", i)
			assert(isStmtCaseClause(c), "should be *astCaseClause", __func__)
			cc := stmt2CaseClause(c)
			labelid++
			var labelCase = ".L.case." + strconv.Itoa(labelid)
			labels[i] = labelCase
			if len(cc.List) == 0 { // @TODO implement slice nil comparison
				defaultLabel = labelCase
				continue
			}
			for _, e := range cc.List {
				assert(getSizeOfType(condType) <= 8 || kind(condType) == T_STRING, "should be one register size or string", __func__)
				emitPushStackTop(condType, "switch expr")
				emitExpr(e, nil)
				emitCompEq(condType)
				emitPopBool(" of switch-case comparison")
				myfmt.Printf("  cmpq $1, %%rax\n")
				myfmt.Printf("  je %s # jump if match\n", labelCase)
			}
		}
		emitComment(2, "End comparison with cases\n")

		// if no case matches, then jump to
		if defaultLabel != "" {
			// default
			myfmt.Printf("  jmp %s\n", defaultLabel)
		} else {
			// exit
			myfmt.Printf("  jmp %s\n", labelEnd)
		}

		emitRevertStackTop(condType)
		for i, c := range cases {
			assert(isStmtCaseClause(c), "should be *astCaseClause", __func__)
			var cc = stmt2CaseClause(c)
			myfmt.Printf("%s:\n", labels[i])
			for _, _s := range cc.Body {
				emitStmt(_s)
			}
			myfmt.Printf("  jmp %s\n", labelEnd)
		}
		myfmt.Printf("%s:\n", labelEnd)
	case *astTypeSwitchStmt:
		typeSwitch := s.node
//		assert(ok, "should exist")
		labelid++
		labelEnd := myfmt.Sprintf(".L.typeswitch.%d.exit", labelid)

		// subjectVariable = subject
		emitVariableAddr(typeSwitch.subjectVariable)
		emitExpr(typeSwitch.subject, nil)
		emitStore(tEface, true, false)

		cases := s.Body.List
		var labels = make([]string, len(cases), len(cases))
		var defaultLabel string
		emitComment(2, "Start comparison with cases\n")
		for i, c := range cases {
			cc := stmt2CaseClause(c)
			//assert(ok, "should be *ast.CaseClause")
			labelid++
			labelCase := ".L.case." + strconv.Itoa(labelid)
			labels[i] = labelCase
			if len(cc.List) == 0 { // @TODO implement slice nil comparison
				defaultLabel = labelCase
				continue
			}
			for _, e := range cc.List {
				emitVariableAddr(typeSwitch.subjectVariable)
				emitLoad(tEface)

				emitDtypeSymbol(e2t(e))
				emitCompExpr("sete") // this pushes 1 or 0 in the end

				emitPopBool(" of switch-case comparison")
				myfmt.Printf("  cmpq $1, %%rax\n")
				myfmt.Printf("  je %s # jump if match\n", labelCase)
			}
		}
		emitComment(2, "End comparison with cases\n")

		// if no case matches, then jump to
		if defaultLabel != "" {
			// default
			myfmt.Printf("  jmp %s\n", defaultLabel)
		} else {
			// exit
			myfmt.Printf("  jmp %s\n", labelEnd)
		}

		for i, typeSwitchCaseClose := range typeSwitch.cases {
			if typeSwitchCaseClose.variable != nil {
				typeSwitch.assignIdent.Obj.Variable = typeSwitchCaseClose.variable
			}
			myfmt.Printf("%s:\n", labels[i])

			for _, _s := range typeSwitchCaseClose.orig.Body {
				if typeSwitchCaseClose.variable != nil {
					// do assignment
					expr := newExpr(typeSwitch.assignIdent)
					emitAddr(expr)
					emitVariableAddr(typeSwitch.subjectVariable)
					emitLoad(tEface)
					myfmt.Printf("  popq %%rax # ifc.dtype\n")
					myfmt.Printf("  popq %%rcx # ifc.data\n")
					myfmt.Printf("  push %%rcx # ifc.data\n")
					emitLoad(typeSwitchCaseClose.variableType)

					emitStore(typeSwitchCaseClose.variableType, true, false)
				}

				emitStmt(_s)
			}
			myfmt.Printf("  jmp %s\n", labelEnd)
		}
		myfmt.Printf("%s:\n", labelEnd)

	case *astBranchStmt:
		var containerFor = s.currentFor
		var labelToGo string
		switch s.Tok {
		case "continue":
			switch s := containerFor.(type)  {
			case *astForStmt:
				labelToGo = s.labelPost
			case *astRangeStmt:
				labelToGo = s.labelPost
			default:
				panic2(__func__, "unexpected container dtype="+dtypeOf(containerFor))
			}
			myfmt.Printf("jmp %s # continue\n", labelToGo)
		case "break":
			switch s := containerFor.(type)  {
			case *astForStmt:
				labelToGo = s.labelExit
			case *astRangeStmt:
				labelToGo = s.labelExit
			default:
				panic2(__func__, "unexpected container dtype="+dtypeOf(containerFor))
			}
			myfmt.Printf("jmp %s # break\n", labelToGo)
		default:
			panic2(__func__, "unexpected tok="+s.Tok)
		}
	default:
		panic2(__func__, "TBI:"+dtypeOf(stmt))
	}
}

func blockStmt2Stmt(block *astBlockStmt) astStmt {
	return newStmt(block)
}

func emitRevertStackTop(t *Type) {
	myfmt.Printf("  addq $%d, %%rsp # revert stack top\n", getSizeOfType(t))
}

var labelid int

func getMethodSymbol(method *Method) string {
	var rcvTypeName = method.rcvNamedType
	var subsymbol string
	if method.isPtrMethod {
		subsymbol =  "$" + rcvTypeName.Name + "." + method.name // pointer
	} else {
		subsymbol =  rcvTypeName.Name + "." + method.name // value
	}

	return getPackageSymbol(method.pkgName, subsymbol)
}

func getPackageSymbol(pkgName string, subsymbol string) string {
	return pkgName + "." + subsymbol
}

func emitFuncDecl(pkgName string, fnc *Func) {
	myfmt.Printf("# emitFuncDecl\n")
	var localarea = fnc.localarea
	var symbol string
	if fnc.method != nil {
		symbol = getMethodSymbol(fnc.method)
	} else {
		symbol = getPackageSymbol(pkgName, fnc.name)
	}
	myfmt.Printf("%s: # args %d, locals %d\n",
		symbol, int(fnc.argsarea), int(fnc.localarea))

	myfmt.Printf("  pushq %%rbp\n")
	myfmt.Printf("  movq %%rsp, %%rbp\n")
	if localarea != 0 {
		myfmt.Printf("  subq $%d, %%rsp # local area\n", -localarea)
	}

	if fnc.Body != nil {
		emitStmt(blockStmt2Stmt(fnc.Body))
	}

	myfmt.Printf("  leave\n")
	myfmt.Printf("  ret\n")
}

func emitGlobalVariableComplex(name *astIdent, t *Type, val astExpr) {
	typeKind := kind(t)
	switch typeKind {
	case T_POINTER:
		myfmt.Printf("# init global %s:\n", name.Name)
		lhs := newExpr(name)
		emitAssign(lhs, val)
	}
}

func emitGlobalVariable(pkg *PkgContainer, name *astIdent, t *Type, val astExpr) {
	typeKind := kind(t)
	myfmt.Printf("%s.%s: # T %s\n", pkg.name, name.Name, typeKind)
	switch typeKind {
	case T_STRING:
		if val == nil {
			myfmt.Printf("  .quad 0\n")
			myfmt.Printf("  .quad 0\n")
			return
		}
		switch vl := val.(type)  {
		case *astBasicLit:
			var sl = getStringLiteral(vl)
			myfmt.Printf("  .quad %s\n", sl.label)
			myfmt.Printf("  .quad %d\n", sl.strlen)
		default:
			panic("Unsupported global string value")
		}
	case T_INTERFACE:
		// only zero value
		myfmt.Printf("  .quad 0 # dtype\n")
		myfmt.Printf("  .quad 0 # data\n")
	case T_BOOL:
		if val == nil {
			myfmt.Printf("  .quad 0 # bool zero value\n")
			return
		}
		switch vl := val.(type)  {
		case *astIdent:
			switch vl.Obj {
			case gTrue:
				myfmt.Printf("  .quad 1 # bool true\n")
			case gFalse:
				myfmt.Printf("  .quad 0 # bool false\n")
			default:
				panic2(__func__, "")
			}
		default:
			panic2(__func__, "")
		}
	case T_INT:
		if val == nil {
			myfmt.Printf("  .quad 0\n")
			return
		}
		switch vl := val.(type)  {
		case *astBasicLit:
			myfmt.Printf("  .quad %s\n", vl.Value)
		default:
			panic("Unsupported global value")
		}
	case T_UINT8:
		if val == nil {
			myfmt.Printf("  .byte 0\n")
			return
		}
		switch vl := val.(type)  {
		case *astBasicLit:
			myfmt.Printf("  .byte %s\n", vl.Value)
		default:
			panic("Unsupported global value")
		}
	case T_UINT16:
		if val == nil {
			myfmt.Printf("  .word 0\n")
			return
		}
		switch val.(type)  {
		case *astBasicLit:
			myfmt.Printf("  .word %s\n", expr2BasicLit(val).Value)
		default:
			panic("Unsupported global value")
		}
	case T_POINTER:
		// will be set in the initGlobal func
		myfmt.Printf("  .quad 0\n")
	case T_UINTPTR:
		if val != nil {
			panic("Unsupported global value")
		}
		// only zero value
		myfmt.Printf("  .quad 0\n")
	case T_SLICE:
		if val != nil {
			panic("Unsupported global value")
		}
		// only zero value
		myfmt.Printf("  .quad 0 # ptr\n")
		myfmt.Printf("  .quad 0 # len\n")
		myfmt.Printf("  .quad 0 # cap\n")
	case T_ARRAY:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		var arrayType = expr2ArrayType(t.e)
		if arrayType.Len == nil {
			panic2(__func__, "global slice is not supported")
		}
		bl := expr2BasicLit(arrayType.Len)
		var length = evalInt(newExpr(bl))
		emitComment(0, "[emitGlobalVariable] array length uint8=%d\n", length)
		var zeroValue string
		var kind string = kind(e2t(arrayType.Elt))
		switch kind {
		case T_INT:
			zeroValue = "  .quad 0 # int zero value\n"
		case T_UINT8:
			zeroValue = "  .byte 0 # uint8 zero value\n"
		case T_STRING:
			zeroValue = "  .quad 0 # string zero value (ptr)\n"
			zeroValue = zeroValue + "  .quad 0 # string zero value (len)\n"
		case T_INTERFACE:
			zeroValue = "  .quad 0 # eface zero value (dtype)\n"
			zeroValue = zeroValue + "  .quad 0 # eface zero value (data)\n"
		default:
			panic2(__func__, "Unexpected kind:"+kind)
		}

		var i int
		for i = 0; i < length; i++ {
			myfmt.Printf(zeroValue)
		}
	default:
		panic2(__func__, "TBI:kind="+typeKind)
	}
}

func generateCode(pkg *PkgContainer) {
	myfmt.Printf("#===================== generateCode %s =====================\n", pkg.name)
	myfmt.Printf(".data\n")
	emitComment(0, "string literals len = %d\n", len(pkg.stringLiterals))
	for _, con := range pkg.stringLiterals {
		emitComment(0, "string literals\n")
		myfmt.Printf("%s:\n", con.sl.label)
		myfmt.Printf("  .string %s\n", con.sl.value)
	}

	for _, spec := range pkg.vars {
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariable(pkg, spec.Name, t, spec.Value)
	}

	myfmt.Printf("\n")
	myfmt.Printf(".text\n")
	myfmt.Printf("%s.__initGlobals:\n", pkg.name)
	for _, spec := range pkg.vars {
		if spec.Value == nil{
			continue
		}
		val := spec.Value
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariableComplex(spec.Name, t, val)
	}
	myfmt.Printf("  ret\n")

	for _, fnc := range pkg.funcs {
		emitFuncDecl(pkg.name, fnc)
	}

	myfmt.Printf("\n")
}

// --- type ---
const sliceSize int = 24
const stringSize int = 16
const intSize int = 8
const ptrSize int = 8
const interfaceSize int = 16

type Type struct {
	//kind string
	e astExpr
}

const T_STRING string = "T_STRING"
const T_INTERFACE string = "T_INTERFACE"
const T_SLICE string = "T_SLICE"
const T_BOOL string = "T_BOOL"
const T_INT string = "T_INT"
const T_INT32 string = "T_INT32"
const T_UINT8 string = "T_UINT8"
const T_UINT16 string = "T_UINT16"
const T_UINTPTR string = "T_UINTPTR"
const T_ARRAY string = "T_ARRAY"
const T_STRUCT string = "T_STRUCT"
const T_POINTER string = "T_POINTER"

func getTypeOfExpr(expr astExpr) *Type {
	//emitComment(0, "[%s] start\n", __func__)
	switch e := expr.(type)  {
	case *astIdent:
		if e.Obj == nil {
			panic(e.Name)
		}
		switch e.Obj.Kind {
		case astVar:
			// injected type is the 1st priority
			// this use case happens in type switch with short decl var
			// switch ident := x.(type) {
			// case T:
			//    y := ident // <= type of ident cannot be associated directly with ident
			//
			if e.Obj.Variable != nil {
				return e.Obj.Variable.typ
			}
			switch decl := e.Obj.Decl.(type) {
			case *astValueSpec:
				var t = &Type{}
				t.e = decl.Type
				return t
			case *astField:
				var t = &Type{}
				t.e = decl.Type
				return t
			case *astAssignStmt: // lhs := rhs
				return getTypeOfExpr(decl.Rhs[0])
			default:
				panic2(__func__, "unkown dtype ")
			}
		case astCon:
			switch e.Obj {
			case gTrue, gFalse:
				return tBool
			}
			switch decl2 := e.Obj.Decl.(type) {
			case *astValueSpec:
				return e2t(decl2.Type)
			default:
				panic2(__func__, "cannot decide type of cont ="+e.Obj.Name)
			}
		default:
			panic2(__func__, "2:Obj=" + e.Obj.Name + e.Obj.Kind)
		}
	case *astBasicLit:
		basicLit := expr2BasicLit(expr)
		switch basicLit.Kind {
		case "STRING":
			return tString
		case "INT":
			return tInt
		case "CHAR":
			return tInt32
		default:
			panic2(__func__, "TBI:"+basicLit.Kind)
		}
	case *astIndexExpr:
		var list = e.X
		return getElementTypeOfListType(getTypeOfExpr(list))
	case *astUnaryExpr:
		switch e.Op {
		case "+":
			return getTypeOfExpr(e.X)
		case "-":
			return getTypeOfExpr(e.X)
		case "!":
			return tBool
		case "&":
			var starExpr = &astStarExpr{}
			var t = getTypeOfExpr(e.X)
			starExpr.X = t.e
			return e2t(newExpr(starExpr))
		case "range":
			listType := getTypeOfExpr(e.X)
			elmType := getElementTypeOfListType(listType)
			return elmType
		default:
			panic2(__func__, "TBI: Op="+e.Op)
		}
	case *astCallExpr:
		emitComment(2, "[%s] *astCallExpr\n", __func__)
		var fun = e.Fun
		switch fn := fun.(type)  {
		case *astIdent:
			if fn.Obj == nil {
				panic2(__func__, "[astCallExpr] nil Obj is not allowed")
			}
			switch fn.Obj.Kind {
			case astTyp:
				return e2t(fun)
			case astFun:
				switch fn.Obj {
				case gLen, gCap:
					return tInt
				case gNew:
					var starExpr = &astStarExpr{}
					starExpr.X = e.Args[0]
					return e2t(newExpr(starExpr))
				case gMake:
					return e2t(e.Args[0])
				case gAppend:
					return e2t(e.Args[0])
				}
				var decl = fn.Obj.Decl
				if decl == nil {
					panic2(__func__, "decl of function "+fn.Name+" is  nil")
				}
				switch dcl := decl.(type) {
				case *astFuncDecl:
					var resultList = dcl.Type.Results.List
					if len(resultList) != 1 {
						panic2(__func__, "[astCallExpr] len results.List is not 1")
					}
					return e2t(dcl.Type.Results.List[0].Type)
				default:
					panic2(__func__, "[astCallExpr] unknown dtype")
				}
				panic2(__func__, "[astCallExpr] Fun ident "+fn.Name)
			}
		case *astParenExpr: // (X)(e) funcall or conversion
			if isType(fn.X) {
				return e2t(fn.X)
			} else {
				panic("TBI: what should we do ?")
			}
		case *astArrayType:
			return e2t(fun)
		case *astSelectorExpr: // (X).Sel()
			xIdent := expr2Ident(fn.X)
			symbol := xIdent.Name + "." + fn.Sel.Name
			switch symbol {
			case "unsafe.Pointer":
				// unsafe.Pointer(x)
				return tUintptr
			default:
				xIdent := expr2Ident(fn.X)
				if xIdent.Obj == nil {
					panic2(__func__,  "xIdent.Obj should not be nil")
				}
				if xIdent.Obj.Kind == astPkg {
					funcdecl := lookupForeignFunc(xIdent.Name, fn.Sel.Name)
					return e2t(funcdecl.Type.Results.List[0].Type)
				} else {
					var xType = getTypeOfExpr(fn.X)
					var method = lookupMethod(xType, fn.Sel)
					assert(len(method.funcType.Results.List) == 1, "func is expected to return a single value", __func__)
					return e2t(method.funcType.Results.List[0].Type)
				}
			}
		case *astInterfaceType:
			return tEface
		default:
			panic2(__func__, "[astCallExpr] dtype="+ dtypeOf(e.Fun))
		}
	case *astSliceExpr:
		var underlyingCollectionType = getTypeOfExpr(e.X)
		if kind(underlyingCollectionType) == T_STRING {
			// str2 = str1[n:m]
			return tString
		}
		var elementTyp astExpr
		switch underlyingCollectionType.e.(type) {
		case *astArrayType:
			elementTyp = expr2ArrayType(underlyingCollectionType.e).Elt
		}
		var t = &astArrayType{}
		t.Len = nil
		t.Elt = elementTyp
		return e2t(newExpr(t))
	case *astStarExpr:
		var t = getTypeOfExpr(expr2StarExpr(expr).X)
		var ptrType = expr2StarExpr(t.e)
		if ptrType == nil {
			panic2(__func__, "starExpr shoud not be nil")
		}
		return e2t(ptrType.X)
	case *astBinaryExpr:
		binaryExpr := expr2BinaryExpr(expr)
		switch binaryExpr.Op {
		case "==", "!=", "<", ">", "<=", ">=":
			return tBool
		default:
			return getTypeOfExpr(binaryExpr.X)
		}
	case *astSelectorExpr:
		x := e.X
		if isExprIdent(x) && expr2Ident(x).Obj.Kind == astPkg {
			ident := lookupForeignVar(expr2Ident(x).Name, e.Sel.Name)
			return getTypeOfExpr(newExpr(ident))
		} else {
			var structType = getStructTypeOfX(e)
			var field = lookupStructField(getStructTypeSpec(structType), e.Sel.Name)
			return e2t(field.Type)
		}
	case *astCompositeLit:
		return e2t(e.Type)
	case *astParenExpr:
		return getTypeOfExpr(e.X)
	case *astTypeAssertExpr:
		return e2t(expr2TypeAssertExpr(expr).Type)
	case *astInterfaceType:
		return tEface
	default:
		panic2(__func__, "TBI:dtype="+dtypeOf(expr))
	}

	panic2(__func__, "nil type is not allowed\n")
	var r *Type
	return r
}

func e2t(typeExpr astExpr) *Type {
	if typeExpr == nil {
		panic2(__func__, "nil is not allowed")
	}
	var r = &Type{}
	r.e = typeExpr
	return r
}

func serializeType(t *Type) string {
	if t == nil {
		panic("nil type is not expected")
	}
	if t.e == generalSlice {
		panic("TBD: generalSlice")
	}

	switch e := t.e.(type) {
	case *astIdent:
		if e.Obj == nil {
			panic("Unresolved identifier:" + e.Name)
		}
		if e.Obj.Kind == astVar {
			throw("bug?")
		} else if e.Obj.Kind == astTyp {
			switch e.Obj {
			case gUintptr:
				return "uintptr"
			case gInt:
				return "int"
			case gString:
				return "string"
			case gUint8:
				return "uint8"
			case gUint16:
				return "uint16"
			case gBool:
				return "bool"
			default:
				// named type
				decl := e.Obj.Decl
				var typeSpec *astTypeSpec
				var ok bool
				typeSpec, ok = decl.(*astTypeSpec)
				if !ok {
					panic("unexpected dtype")
				}
				return "main." + typeSpec.Name.Name
			}
		}
	case *astStructType:
		return "struct"
	case *astArrayType:
		if e.Len == nil {
			if e.Elt == nil {
				panic(e)
			}
			return "[]" + serializeType(e2t(e.Elt))
		} else {
			return "[" + strconv.Itoa(evalInt(e.Len)) + "]" + serializeType(e2t(e.Elt))
		}
	case *astStarExpr:
		return "*" + serializeType(e2t(e.X))
	case *astEllipsis: // x ...T
		panic("TBD: Ellipsis")
	case *astInterfaceType:
		return "interface"
	default:
		throw(dtypeOf(t.e))
	}
	return ""
}


func kind(t *Type) string {
	if t == nil {
		panic2(__func__, "nil type is not expected\n")
	}
	if t.e == generalSlice {
		return T_SLICE
	}

	switch e := t.e.(type)  {
	case *astIdent:
		var ident = e
		switch ident.Name {
		case "uintptr":
			return T_UINTPTR
		case "int":
			return T_INT
		case "int32":
			return T_INT32
		case "string":
			return T_STRING
		case "uint8", "byte":
			return T_UINT8
		case "uint16":
			return T_UINT16
		case "bool":
			return T_BOOL
		default:
			// named type
			var decl = ident.Obj.Decl
			var typeSpec *astTypeSpec
			var ok bool
			typeSpec , ok = decl.(*astTypeSpec)
			if !ok {
				panic2(__func__, "unsupported decl :")
			}
			return kind(e2t(typeSpec.Type))
		}
	case *astStructType:
		return T_STRUCT
	case *astArrayType:
		if e.Len == nil {
			return T_SLICE
		} else {
			return T_ARRAY
		}
	case *astStarExpr:
		return T_POINTER
	case *astEllipsis: // x ...T
		return T_SLICE // @TODO is this right ?
	case *astInterfaceType:
		return T_INTERFACE
	case *astParenExpr:
		panic(dtypeOf(e))
	case *astSelectorExpr:
		full := expr2Ident(e.X).Name + "." + e.Sel.Name
		if full == "unsafe.Pointer" {
			return T_POINTER
		} else {
			panic2(__func__, "Unkown type")
		}
	default:
		panic2(__func__, "Unkown dtype:"+dtypeOf(t.e))
	}
	panic2(__func__, "error")
	return ""
}

func isInterface(t *Type) bool {
	return kind(t) == T_INTERFACE
}

func getStructTypeOfX(e *astSelectorExpr) *Type {
	var typeOfX = getTypeOfExpr(e.X)
	var structType *Type
	switch kind(typeOfX) {
	case T_STRUCT:
		// strct.field => e.X . e.Sel
		structType = typeOfX
	case T_POINTER:
		// ptr.field => e.X . e.Sel
		ptrType := expr2StarExpr(typeOfX.e)
		structType = e2t(ptrType.X)
	default:
		panic2(__func__, "TBI")
	}
	return structType
}

func getElementTypeOfListType(t *Type) *Type {
	switch kind(t) {
	case T_SLICE, T_ARRAY:
		switch e := t.e.(type) {
	case *astArrayType:
		return e2t(e.Elt)
	case *astEllipsis:
		return e2t(e.Elt)
	default:
		throw(dtypeOf(t.e))
	}
	case T_STRING:
		return tUint8
	default:
		panic2(__func__, "TBI kind="+kind(t))
	}
	var r *Type
	return r
}

func getSizeOfType(t *Type) int {
	var knd = kind(t)
	switch kind(t) {
	case T_SLICE:
		return sliceSize
	case T_STRING:
		return 16
	case T_ARRAY:
		arrayType := expr2ArrayType(t.e)
		var elemSize = getSizeOfType(e2t(arrayType.Elt))
		return elemSize * evalInt(arrayType.Len)
	case T_INT, T_UINTPTR, T_POINTER:
		return 8
	case T_UINT8:
		return 1
	case T_UINT16:
		return 2
	case T_BOOL:
		return 8
	case T_STRUCT:
		return calcStructSizeAndSetFieldOffset(getStructTypeSpec(t))
	case T_INTERFACE:
		return interfaceSize
	default:
		panic2(__func__, "TBI:"+knd)
	}
	return 0
}

func getPushSizeOfType(t *Type) int {
	switch kind(t) {
	case T_SLICE:
		return sliceSize
	case T_STRING:
		return stringSize
	case T_INTERFACE:
		return interfaceSize
	case T_UINT8, T_UINT16, T_INT, T_BOOL:
		return intSize
	case T_UINTPTR, T_POINTER:
		return ptrSize
	case T_ARRAY, T_STRUCT:
		return ptrSize
	default:
		throw(kind(t))
	}
	throw(kind(t))
	return 0
}

func getStructFieldOffset(field *astField) int {
	var offset = field.Offset
	return offset
}

func setStructFieldOffset(field *astField, offset int) {
	field.Offset = offset
}

func getStructFields(structTypeSpec *astTypeSpec) []*astField {
	var structType = expr2StructType(structTypeSpec.Type)
	return structType.Fields.List
}

func getStructTypeSpec(namedStructType *Type) *astTypeSpec {
	if kind(namedStructType) != T_STRUCT {
		panic2(__func__, "not T_STRUCT")
	}
	ident := expr2Ident(namedStructType.e)
	var typeSpec *astTypeSpec
	var ok bool
	typeSpec , ok = ident.Obj.Decl.(*astTypeSpec)
	if !ok {
		panic2(__func__, "not *astTypeSpec")
	}

	return typeSpec
}

func lookupStructField(structTypeSpec *astTypeSpec, selName string) *astField {
	var field *astField
	for _, field := range getStructFields(structTypeSpec) {
		if field.Name.Name == selName {
			return field
		}
	}
	panic("Unexpected flow: struct field not found:" + selName)
	return field
}

func calcStructSizeAndSetFieldOffset(structTypeSpec *astTypeSpec) int {
	var offset int = 0

	var fields = getStructFields(structTypeSpec)
	for _, field := range fields {
		setStructFieldOffset(field, offset)
		var size = getSizeOfType(e2t(field.Type))
		offset = offset + size
	}
	return offset
}

// --- walk ---
type sliteral struct {
	label  string
	strlen int
	value  string // raw value/pre/precompiler.go:2150
}

type stringLiteralsContainer struct {
	lit *astBasicLit
	sl  *sliteral
}

type Func struct {
	localvars []*string
	localarea int
	argsarea  int
	funcType  *astFuncType
	rcvType   astExpr
	name      string
	Body      *astBlockStmt
	method    *Method
}

type Method struct {
	pkgName string
	rcvNamedType *astIdent
	isPtrMethod  bool
	name string
	funcType *astFuncType
}

type Variable struct {
	name         string
	isGlobal     bool
	globalSymbol string
	localOffset  int
	typ          *Type
}

//type localoffsetint int //@TODO

func (fnc *Func) registerParamVariable(name string, t *Type) *Variable {
	vr := newLocalVariable(name, fnc.argsarea, t)
	fnc.argsarea = fnc.argsarea + getSizeOfType(t)
	return vr
}

func (fnc *Func) registerLocalVariable(name string, t *Type) *Variable {
	assert(t != nil && t.e != nil, "type of local var should not be nil", __func__)
	fnc.localarea = fnc.localarea - getSizeOfType(t)
	return newLocalVariable(name, currentFunc.localarea, t)
}

var currentFunc *Func

func getStringLiteral(lit *astBasicLit) *sliteral {
	for _, container := range pkg.stringLiterals {
		if container.lit == lit {
			return container.sl
		}
	}

	panic2(__func__, "string literal not found:"+lit.Value)
	var r *sliteral
	return r
}

func registerStringLiteral(lit *astBasicLit) {
	logf(" [registerStringLiteral] begin\n")

	if pkg.name == "" {
		panic2(__func__, "no pkgName")
	}

	var strlen int
	var vl = []uint8(lit.Value)
	for _, c := range vl {
		if c != '\\' {
			strlen++
		}
	}

	label := myfmt.Sprintf(".%s.S%d", pkg.name, pkg.stringIndex)
	pkg.stringIndex++

	sl := &sliteral{
		label : label,
		strlen : strlen - 2,
		value: lit.Value,
	}
	logf(" [registerStringLiteral] label=%s, strlen=%d %s\n", sl.label, sl.strlen, sl.value)
	cont := &stringLiteralsContainer{}
	cont.sl = sl
	cont.lit = lit
	pkg.stringLiterals = append(pkg.stringLiterals, cont)
}

func newGlobalVariable(pkgName string, name string, t *Type) *Variable {
	vr := &Variable{
		name:         name,
		isGlobal:     true,
		globalSymbol: pkgName + "." + name,
		typ: t,
	}
	return vr
}

func newLocalVariable(name string, localoffset int, t *Type) *Variable {
	vr := &Variable{
		name : name,
		isGlobal : false,
		localOffset : localoffset,
		typ: t,
	}
	return vr
}


type methodEntry struct {
	name string
	method *Method
}

type namedTypeEntry struct {
	name    string
	methods []*methodEntry
}

var typesWithMethods []*namedTypeEntry

func findNamedType(typeName string) *namedTypeEntry {
	var typ *namedTypeEntry
	for _, typ := range typesWithMethods {
		if typ.name == typeName {
			return typ
		}
	}
	typ = nil
	return typ
}

func newMethod(pkgName string, funcDecl *astFuncDecl) *Method {
	var rcvType = funcDecl.Recv.List[0].Type
	var isPtr bool
	if isExprStarExpr(rcvType) {
		isPtr = true
		rcvType = expr2StarExpr(rcvType).X
	}

	rcvNamedType := expr2Ident(rcvType)
	var method = &Method{
		pkgName: pkgName,
		rcvNamedType: rcvNamedType,
		isPtrMethod : isPtr,
		name: funcDecl.Name.Name,
		funcType: funcDecl.Type,
	}
	return method
}

func registerMethod(method *Method) {
	var nt = findNamedType(method.rcvNamedType.Name)
	if nt == nil {
		nt = &namedTypeEntry{
			name:    method.rcvNamedType.Name,
			methods: nil,
		}
		typesWithMethods = append(typesWithMethods, nt)
	}

	var me *methodEntry = &methodEntry{
		name: method.name,
		method: method,
	}
	nt.methods = append(nt.methods, me)
}

func lookupMethod(rcvT *Type, methodName *astIdent) *Method {
	var rcvType = rcvT.e
	if isExprStarExpr(rcvType) {
		rcvType = expr2StarExpr(rcvType).X
	}

	var rcvTypeName = expr2Ident(rcvType)
	var nt = findNamedType(rcvTypeName.Name)
	if nt == nil {
		panic(rcvTypeName.Name + " has no moethodeiverTypeName:")
	}

	for _, me := range nt.methods {
		if me.name == methodName.Name {
			return me.method
		}
	}

	panic("method not found: " + methodName.Name)
	return nil
}

func walkStmt(stmt astStmt) {
	logf(" [%s] begin dtype=%s\n", __func__, dtypeOf(stmt))
	switch s := stmt.(type) {
	case *astDeclStmt:
		logf(" [%s] *ast.DeclStmt\n", __func__)
		if s == nil {
			panic2(__func__, "nil pointer exception\n")
		}
		var declStmt = s
		if declStmt.Decl == nil {
			panic2(__func__, "ERROR\n")
		}
		var dcl = declStmt.Decl
		var genDecl *astGenDecl
		var ok bool
		genDecl, ok = dcl.(*astGenDecl)
		if !ok {
			panic2(__func__, "[dcl.dtype] internal error")
		}
		var valSpec *astValueSpec
		valSpec, ok = genDecl.Spec.(*astValueSpec)
		if valSpec.Type == nil {
			if valSpec.Value == nil {
				panic2(__func__, "type inference requires a value")
			}
			var _typ = getTypeOfExpr(valSpec.Value)
			if _typ != nil && _typ.e != nil {
				valSpec.Type = _typ.e
			} else {
				panic2(__func__, "type inference failed")
			}
		}
		var typ = valSpec.Type // Type can be nil
		logf(" [walkStmt] valSpec Name=%s, Type=%s\n",
			valSpec.Name.Name, dtypeOf(typ))

		t := e2t(typ)
		valSpec.Name.Obj.Variable = currentFunc.registerLocalVariable(valSpec.Name.Name, t)
		logf(" var %s offset = %d\n", valSpec.Name.Obj.Name,
			valSpec.Name.Obj.Variable.localOffset)
		if valSpec.Value != nil {
			walkExpr(valSpec.Value)
		}
	case *astAssignStmt:
		var lhs = s.Lhs[0]
		var rhs = s.Rhs[0]
		if s.Tok == ":=" {
			assert(isExprIdent(lhs), "should be ident", __func__)
			var obj = expr2Ident(lhs).Obj
			assert(obj.Kind == astVar, obj.Name + " should be ast.Var", __func__)
			walkExpr(rhs)
			// infer type
			var typ = getTypeOfExpr(rhs)
			if typ != nil && typ.e != nil {
			} else {
				panic("type inference is not supported: " + obj.Name)
			}
			logf("infered type of %s is %s, rhs=%s\n", obj.Name, dtypeOf(typ.e), dtypeOf(rhs))
			obj.Variable = currentFunc.registerLocalVariable(obj.Name, typ)
		} else {
			walkExpr(rhs)
		}
	case *astExprStmt:
		walkExpr(s.X)
	case *astReturnStmt:
		s.node = &nodeReturnStmt{
			fnc: currentFunc,
		}
		for _, rt := range s.Results {
			walkExpr(rt)
		}
	case *astIfStmt:
		if s.Init != nil {
			walkStmt(s.Init)
		}
		walkExpr(s.Cond)
		for _, s := range s.Body.List {
			walkStmt(s)
		}
		if s.Else != nil {
			walkStmt(s.Else)
		}
	case *astForStmt:
		s.Outer = currentFor
		currentFor = stmt
		if s.Init != nil {
			walkStmt(s.Init)
		}
		if s.Cond != nil {
			walkExpr(s.Cond)
		}
		if s.Post != nil {
			walkStmt(s.Post)
		}
		walkStmt(newStmt(s.Body))
		currentFor = s.Outer
	case *astRangeStmt:
		walkExpr(s.X)
		s.Outer = currentFor
		currentFor = stmt
		var _s = blockStmt2Stmt(s.Body)
		walkStmt(_s)
		var lenvar = currentFunc.registerLocalVariable(".range.len", tInt)
		var indexvar = currentFunc.registerLocalVariable(".range.index", tInt)

		if s.Tok == ":=" {
			listType := getTypeOfExpr(s.X)

			keyIdent := expr2Ident(s.Key)
			//@TODO map key can be any type
			//keyType := getKeyTypeOfListType(listType)
			var keyType *Type = tInt
			keyIdent.Obj.Variable =  currentFunc.registerLocalVariable(keyIdent.Name, keyType)

			// determine type of Value
			elmType := getElementTypeOfListType(listType)
			valueIdent := expr2Ident(s.Value)
			valueIdent.Obj.Variable = currentFunc.registerLocalVariable(valueIdent.Name, elmType)
		}
		s.lenvar = lenvar
		s.indexvar = indexvar
		currentFor = s.Outer
	case *astIncDecStmt:
		walkExpr(s.X)
	case *astBlockStmt:
		for _, s := range s.List {
			walkStmt(s)
		}
	case *astBranchStmt:
		s.currentFor = currentFor
	case *astSwitchStmt:
		if s.Tag != nil {
			walkExpr(s.Tag)
		}
		walkStmt(blockStmt2Stmt(s.Body))
	case *astTypeSwitchStmt:
		typeSwitch := &nodeTypeSwitchStmt{}
		s.node = typeSwitch
		var assignIdent *astIdent
		switch s2 := s.Assign.(type) {
		case *astExprStmt:
			typeAssertExpr := expr2TypeAssertExpr(s2.X)
			//assert(ok, "should be *ast.TypeAssertExpr")
			typeSwitch.subject = typeAssertExpr.X
			walkExpr(typeAssertExpr.X)
		case *astAssignStmt:
			lhs := s2.Lhs[0]
			//var ok bool
			assignIdent = expr2Ident(lhs)
			//assert(ok, "lhs should be ident")
			typeSwitch.assignIdent = assignIdent
			// ident will be a new local variable in each case clause
			typeAssertExpr := expr2TypeAssertExpr(s2.Rhs[0])
			//assert(ok, "should be *ast.TypeAssertExpr")
			typeSwitch.subject = typeAssertExpr.X
			walkExpr(typeAssertExpr.X)
		default:
			throw(dtypeOf(s.Assign))
		}

		typeSwitch.subjectVariable = currentFunc.registerLocalVariable(".switch_expr", tEface)
		for _, _case := range s.Body.List {
			cc := stmt2CaseClause(_case)
			tscc := &TypeSwitchCaseClose{
				orig: cc,
			}
			typeSwitch.cases = append(typeSwitch.cases, tscc)
			if  assignIdent != nil && len(cc.List) > 0 {
				// inject a variable of that type
				varType := e2t(cc.List[0])
				vr := currentFunc.registerLocalVariable(assignIdent.Name, varType)
				tscc.variable = vr
				tscc.variableType = varType
				assignIdent.Obj.Variable = vr
			}

			for _, s_ := range cc.Body {
				walkStmt(s_)
			}
			if assignIdent != nil {
				assignIdent.Obj.Variable = nil
			}
		}
	case *astCaseClause:
		for _, e_ := range s.List {
			walkExpr(e_)
		}
		for _, s_ := range s.Body {
			walkStmt(s_)
		}
	default:
		panic2(__func__, "TBI: s="+dtypeOf(stmt))
	}
}

var currentFor astStmt

func walkExpr(expr astExpr) {
	logf(" [walkExpr] dtype=%s\n", dtypeOf(expr))
	switch e := expr.(type) {
	case *astIdent:
		// what to do ?
	case *astCallExpr:
		walkExpr(e.Fun)
		// Replace __func__ ident by a string literal
		var basicLit *astBasicLit
		var newArg astExpr
		for i, arg := range e.Args {
			if isExprIdent(arg) {
				ident := expr2Ident(arg)
				if ident.Name == "__func__" && ident.Obj.Kind == astVar {
					basicLit = &astBasicLit{}
					basicLit.Kind = "STRING"
					basicLit.Value = "\"" + currentFunc.name + "\""
					newArg = newExpr(basicLit)
					e.Args[i] = newArg
					arg = newArg
				}
			}
			walkExpr(arg)
		}
	case *astBasicLit:
		basicLit := e
		switch basicLit.Kind {
		case "STRING":
			registerStringLiteral(basicLit)
		}
	case *astCompositeLit:
		for _, v := range e.Elts {
			walkExpr(v)
		}
	case *astUnaryExpr:
		walkExpr(e.X)
	case *astBinaryExpr:
		binaryExpr := e
		walkExpr(binaryExpr.X)
		walkExpr(binaryExpr.Y)
	case *astIndexExpr:
		walkExpr(e.Index)
		walkExpr(e.X)
	case *astSliceExpr:
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
	case *astStarExpr:
		walkExpr(e.X)
	case *astSelectorExpr:
		walkExpr(e.X)
	case *astArrayType: // []T(e)
		// do nothing ?
	case *astParenExpr:
		walkExpr(e.X)
	case *astKeyValueExpr:
		walkExpr(e.Key)
		walkExpr(e.Value)
	case *astInterfaceType:
		// interface{}(e)  conversion. Nothing to do.
	case *astTypeAssertExpr:
		walkExpr(e.X)
	default:
		panic2(__func__, "TBI:"+dtypeOf(expr))
	}
}

var ExportedQualifiedIdents []*exportEntry

type exportEntry struct {
	qi string
	any interface{} // *astFuncDecl|*astIdent(variable)
}

func walk(pkg *PkgContainer) {
	var typeSpecs []*astTypeSpec
	var funcDecls []*astFuncDecl
	var varSpecs []*astValueSpec
	var constSpecs []*astValueSpec

	for _, decl := range pkg.Decls {
		switch dcl := decl.(type) {
		case *astGenDecl:
			switch spec := dcl.Spec.(type) {
			case *astTypeSpec:
				typeSpecs = append(typeSpecs, spec)
			case *astValueSpec:
				if spec.Name.Obj.Kind == astVar {
					varSpecs = append(varSpecs, spec)
				} else if spec.Name.Obj.Kind == astCon {
					constSpecs = append(constSpecs, spec)
				} else {
					panic("Unexpected")
				}
			}
		case *astFuncDecl:
			funcDecls = append(funcDecls, dcl)
		default:
			panic("Unexpected")
		}
	}

	for _, typeSpec := range typeSpecs {
		switch kind(e2t(typeSpec.Type)) {
		case T_STRUCT:
			calcStructSizeAndSetFieldOffset(typeSpec)
		}
	}

	// collect methods in advance
	for _, funcDecl := range funcDecls {
		exportEntry := &exportEntry{
			qi: pkg.name + "." + funcDecl.Name.Name,
			any: funcDecl,
		}
		ExportedQualifiedIdents = append(ExportedQualifiedIdents, exportEntry)
		if funcDecl.Body != nil {
			if funcDecl.Recv != nil { // is Method
				var method = newMethod(pkg.name, funcDecl)
				registerMethod(method)
			}
		}
	}

	for _, constSpec := range constSpecs {
		walkExpr(constSpec.Value)
	}

	for _, valSpec := range varSpecs {
		var nameIdent = valSpec.Name
		assert(nameIdent.Obj.Kind == astVar, "should be Var", __func__)
		if valSpec.Type == nil {
			var val = valSpec.Value
			var t = getTypeOfExpr(val)
			valSpec.Type = t.e
		}
		nameIdent.Obj.Variable = newGlobalVariable(pkg.name, nameIdent.Obj.Name, e2t(valSpec.Type))
		pkg.vars = append(pkg.vars, valSpec)
		exportEntry := &exportEntry{
			qi: pkg.name + "." + nameIdent.Name,
			any: nameIdent,
		}
		ExportedQualifiedIdents = append(ExportedQualifiedIdents, exportEntry)
		if valSpec.Value != nil {
			walkExpr(valSpec.Value)
		}
	}

		for _, funcDecl := range funcDecls {
			fnc := &Func{
				name:      funcDecl.Name.Name,
				funcType:  funcDecl.Type,
				localarea: 0,
				argsarea:  16,
			}
			currentFunc = fnc
			logf(" [sema] == astFuncDecl %s ==\n", funcDecl.Name.Name)
			//var paramoffset = 16
			var paramFields []*astField

			if funcDecl.Recv != nil { // Method
				paramFields = append(paramFields, funcDecl.Recv.List[0])
			}
			for _, field := range funcDecl.Type.Params.List {
				paramFields = append(paramFields, field)
			}

			for _, field := range paramFields {
				obj := field.Name.Obj
				obj.Variable = fnc.registerParamVariable(obj.Name, e2t(field.Type))
			}

			if funcDecl.Body != nil {
				for _, stmt := range funcDecl.Body.List {
					walkStmt(stmt)
				}
				fnc.Body = funcDecl.Body

				if funcDecl.Recv != nil { // Method
					fnc.method = newMethod(pkg.name, funcDecl)
				}
				pkg.funcs = append(pkg.funcs, fnc)
			}
		}
}

// --- universe ---
var gNil = &astObject{
	Kind : astCon, // is it Con ?
	Name : "nil",
}

var identNil = &astIdent{
	Obj:  gNil,
	Name: "nil",
}

var eNil astExpr
var eZeroInt astExpr

var gTrue = &astObject{
	Kind: astCon,
	Name: "true",
}
var gFalse = &astObject{
	Kind: astCon,
	Name: "false",
}

var gString = &astObject{
	Kind: astTyp,
	Name: "string",
}

var gInt = &astObject{
	Kind: astTyp,
	Name: "int",
}

var gInt32 = &astObject{
	Kind: astTyp,
	Name: "int32",
}


var gUint8 = &astObject{
	Kind: astTyp,
	Name: "uint8",
}

var gUint16 = &astObject{
	Kind: astTyp,
	Name: "uint16",
}
var gUintptr = &astObject{
	Kind: astTyp,
	Name: "uintptr",
}
var gBool = &astObject{
	Kind: astTyp,
	Name: "bool",
}

var gNew = &astObject{
	Kind: astFun,
	Name: "new",
}

var gMake = &astObject{
	Kind: astFun,
	Name: "make",
}
var gAppend = &astObject{
	Kind: astFun,
	Name: "append",
}

var gLen = &astObject{
	Kind: astFun,
	Name: "len",
}

var gCap = &astObject{
	Kind: astFun,
	Name: "cap",
}
var gPanic = &astObject{
	Kind: astFun,
	Name: "panic",
}

var tInt *Type
var tInt32 *Type // Rune
var tUint8 *Type
var tUint16 *Type
var tUintptr *Type
var tString *Type
var tEface *Type
var tBool *Type
var generalSlice astExpr

func createUniverse() *astScope {
	var universe = new(astScope)

	universe.Insert(gInt)
	universe.Insert(gUint8)

	universe.Objects = append(universe.Objects, &objectEntry{
		name: "byte",
		obj:  gUint8,
	})

	universe.Insert(gUint16)
	universe.Insert(gUintptr)
	universe.Insert(gString)
	universe.Insert(gBool)
	universe.Insert(gNil)
	universe.Insert(gTrue)
	universe.Insert(gFalse)
	universe.Insert(gNew)
	universe.Insert(gMake)
	universe.Insert(gAppend)
	universe.Insert(gLen)
	universe.Insert(gCap)
	universe.Insert(gPanic)

	logf(" [%s] scope insertion of predefined identifiers complete\n", __func__)

	// @FIXME package names should not be be in universe

	universe.Insert(&astObject{
		Kind: "Pkg",
		Name: "os",
	})

	universe.Insert(&astObject{
		Kind: "Pkg",
		Name: "syscall",
	})

	universe.Insert(&astObject{
		Kind: "Pkg",
		Name: "unsafe",
	})
	logf(" [%s] scope insertion complete\n", __func__)
	return universe
}

func resolveImports(file *astFile) {
	var mapImports []string
	for _, imprt := range file.Imports {
		// unwrap double quote "..."
		rawPath := imprt.Path[1:(len(imprt.Path) - 1)]
		base := path.Base(rawPath)
		mapImports = append(mapImports, base)
	}
	for _, ident := range file.Unresolved {
		if mylib.InArray(ident.Name, mapImports) {
			ident.Obj = &astObject{
				Kind: astPkg,
				Name: ident.Name,
			}
			logf("# resolved: %s\n", ident.Name)
		}
	}
}

func lookupForeignVar(pkg string, identifier string) *astIdent {
	key := pkg + "." + identifier
	logf("lookupForeignVar... %s\n", key)
	for _, entry := range ExportedQualifiedIdents {
		logf("  looking into %s\n", entry.qi)
		if entry.qi == key {
			var ident *astIdent
			var ok bool
			ident, ok = entry.any.(*astIdent)
			if !ok  {
				panic("not ident")
			}
			return ident
		}
	}
	return nil
}

func lookupForeignFunc(pkg string, identifier string) *astFuncDecl {
	key := pkg + "." + identifier
	logf("lookupForeignFunc... %s\n", key)
	for _, entry := range ExportedQualifiedIdents {
		logf("  looking into %s\n", entry.qi)
		if entry.qi == key {
			var fdecl *astFuncDecl
			var ok bool
			fdecl, ok = entry.any.(*astFuncDecl)
			if !ok  {
				panic("not fdecl")
			}
			return fdecl
		}
	}
	panic("function not found: " + key)
	return nil
}


var pkg *PkgContainer

type PkgContainer struct {
	path     string
	name     string
	files    []string
	astFiles []*astFile
	vars     []*astValueSpec
	funcs    []*Func
	stringLiterals []*stringLiteralsContainer
	stringIndex int
	Decls    []astDecl
}

func showHelp() {
	myfmt.Printf("Usage:\n")
	myfmt.Printf("    babygo version:  show version\n")
	myfmt.Printf("    babygo [-DF] [-DG] filename\n")
}

const GOPATH string = "/root/go"

// "foo/bar" => "bar.go"
// "some/dir" => []string{"a.go", "b.go"}
func findFilesInDir(dir string) []string {
	//fname := path2.Base(dir) + ".go"
	//return []string{fname}
	dirents := mylib.GetDirents(dir)
	var r []string
	for _, dirent := range dirents {
		if dirent == "." || dirent == ".." {
			continue
		}
		r = append(r, dirent)
	}
	return r
}

func isStdLib(pth string) bool {
	return !strings.Contains(pth, "/")
}

func getImportPathsFromFile(file string) []string {
	astFile0 := parseImports(file)
	var importPaths []string
	for _, importSpec := range astFile0.Imports {
		rawValue := importSpec.Path
		logf("import %s\n", rawValue)
		pth :=  rawValue[1:len(rawValue)-1]
		importPaths = append(importPaths, pth)
	}
	return 	importPaths
}

func isInTree(tree []*depEntry, pth string) bool {
	for _, entry := range tree {
		if entry.path == pth {
			return true
		}
	}
	return false
}

func getPackageDir(importPath string) string {
	if isStdLib(importPath) {
		return srcPath + "/github.com/DQNEO/babygo/src/" + importPath
	} else {
		return srcPath + "/" + importPath
	}
}

func collectDependency(tree []*depEntry, paths []string) []*depEntry {
	logf(" collectDependency\n")
	for _, pkgPath := range paths {
		if isInTree(tree, pkgPath) {
			continue
		}
		logf("   in pkgPath=%s\n", pkgPath)
		packageDir := getPackageDir(pkgPath)
		fnames := findFilesInDir(packageDir)
		var children []string
		for _, fname := range fnames {
			_paths := getImportPathsFromFile(packageDir + "/" + fname)
			for _, p := range _paths {
				children = append(children, p)
			}
		}

		newEntry := &depEntry{
			path:     pkgPath,
			children: children,
		}
		tree = append(tree, newEntry)
		tree = collectDependency(tree, children)
	}
	return tree
}

func removeLeafNode(tree []*depEntry, sortedPaths []string) []*depEntry {
	// remove leaf node
	var newTree []*depEntry
	for _, entry := range tree {
		if mylib.InArray(entry.path, sortedPaths) {
			continue
		}
		de := &depEntry{
			path:     entry.path,
			children: nil,
		}
		for _, child := range entry.children {
			if mylib.InArray(child, sortedPaths) {
				continue
			}
			de.children = append(de.children, child)
		}
		newTree = append(newTree, de)
	}
	return newTree
}

func collectLeafNode(sortedPaths []string, tree []*depEntry) []string {
	for _, entry := range tree {
		if len(entry.children) == 0 {
			// leaf node
			logf("Found leaf node: %s\n", entry.path)
			logf("  num children: %d\n", len(entry.children))
			sortedPaths = append(sortedPaths, entry.path)
		}
	}
	return sortedPaths
}

func sortDepTree(tree []*depEntry) []string {
	var sortedPaths []string

	var keys []string
	for _, entry := range tree {
		keys = append(keys, entry.path)
	}
	mylib.SortStrings(keys)
	var newTree []*depEntry
	for _, key := range keys {
		for _, entry := range tree {
			if entry.path == key {
				newTree = append(newTree, entry)
			}
		}
	}
	tree = newTree
	logf("====TREE====\n")
	for {
		if len(tree) == 0 {
			break
		}
		sortedPaths = collectLeafNode(sortedPaths, tree)
		tree = removeLeafNode(tree, sortedPaths)
	}
	return sortedPaths
}

var srcPath string

func main() {
	if len(os.Args) == 1 {
		showHelp()
		return
	}

	if os.Args[1] == "version" {
		myfmt.Printf("babygo version 0.1.0  linux/amd64\n")
		return
	} else if os.Args[1] == "help" {
		showHelp()
		return
	} else if os.Args[1] == "panic" {
		panicVersion := strconv.Itoa(mylib.Sum(1 , 1))
		panic("I am panic version " + panicVersion)
	}

	logf("Build start\n")
	srcPath = os.Getenv("GOPATH") + "/src"

	eNil = newExpr(identNil)
	eZeroInt = newExpr(&astBasicLit{
		Value: "0",
		Kind:  "INT",
	})
	generalSlice = newExpr(&astIdent{})
	tInt = &Type{
		e: newExpr(&astIdent{
			Name: "int",
			Obj:  gInt,
		}),
	}
	tInt32 = &Type{
		e: newExpr(&astIdent{
			Name: "int32",
			Obj:  gInt32,
		}),
	}
	tUint8 = &Type{
		e: newExpr(&astIdent{
			Name: "uint8",
			Obj:  gUint8,
		}),
	}

	tUint16 = &Type{
		e: newExpr(&astIdent{
			Name: "uint16",
			Obj:  gUint16,
		}),
	}
	tUintptr = &Type{
		e: newExpr(&astIdent{
			Name: "uintptr",
			Obj:  gUintptr,
		}),
	}

	tString = &Type{
		e: newExpr(&astIdent{
			Name: "string",
			Obj:  gString,
		}),
	}

	tEface = &Type{
		e: newExpr(&astInterfaceType{}),
	}

	tBool = &Type{
		e: newExpr(&astIdent{
			Name: "bool",
			Obj:  gBool,
		}),
	}

	var universe = createUniverse()
	var arg string
	var inputFiles []string
	for _, arg = range os.Args[1:] {
		switch arg {
		case "-DF":
			debugFrontEnd = true
		case "-DG":
			debugCodeGen = true
		default:
			inputFiles = append(inputFiles, arg)
		}
	}

	var importPaths []string

	for _, inputFile := range inputFiles {
		logf("input file: \"%s\"\n", inputFile)
		logf("Parsing imports\n")
		_paths := getImportPathsFromFile(inputFile)
		for _ , p := range _paths {
			if !mylib.InArray(p, importPaths) {
				importPaths = append(importPaths, p)
			}
		}
	}

	var stdPackagesUsed []string
	var extPackagesUsed []string
	var tree []*depEntry
	tree = collectDependency(tree, importPaths)
	logf("====TREE====\n")
	for _, _pkg := range tree {
		logf("pkg: %s\n", _pkg.path)
		for _, child := range _pkg.children {
			logf("  %s\n", child)
		}
	}

	sortedPaths := sortDepTree(tree)
	for _, pth := range sortedPaths {
		if isStdLib(pth) {
			stdPackagesUsed = append(stdPackagesUsed, pth)
		} else {
			extPackagesUsed = append(extPackagesUsed, pth)
		}
	}

	pkgRuntime := &PkgContainer{
		path: "runtime",
	}
	var packagesToBuild =  []*PkgContainer{pkgRuntime}
	myfmt.Printf("# === sorted stdPackagesUsed ===\n")
	for _, _path := range stdPackagesUsed {
		myfmt.Printf("#  %s\n", _path)
		packagesToBuild = append(packagesToBuild, &PkgContainer{
			path : _path,
		})
	}

	myfmt.Printf("# === sorted extPackagesUsed ===\n")
	for _, _path := range extPackagesUsed {
		myfmt.Printf("#  %s\n", _path)
		packagesToBuild = append(packagesToBuild, &PkgContainer{
			path : _path,
		})
	}
	mainPkg := &PkgContainer{
		name: "main",
		files: inputFiles,
	}
	packagesToBuild = append(packagesToBuild,mainPkg)
	//[]string{"runtime.go"}
	for _, _pkg := range packagesToBuild {
		if len(_pkg.files)  == 0 {
			pkgDir := getPackageDir(_pkg.path)
			fnames := findFilesInDir(pkgDir)
			var files []string
			for _, fname := range fnames {
				srcFile := pkgDir + "/" + fname
				files = append(files, srcFile)
			}
			_pkg.files = files
		}
		pkgScope := astNewScope(universe)
		for _, file := range _pkg.files {
			logf("Parsing file: %s\n", file)
			af := parseFile(file, false)
			_pkg.name = af.Name
			_pkg.astFiles = append(_pkg.astFiles, af)
			for _, oe := range af.scope.Objects {
				pkgScope.Objects = append(pkgScope.Objects, oe)
			}
		}
		for _, af := range _pkg.astFiles {
			resolveImports(af)
			logf("[%s] start\n", __func__)
			// inject predeclared identifers
			var unresolved []*astIdent
			logf(" [SEMA] resolving af.Unresolved (n=%d)\n", len(af.Unresolved))
			for _, ident := range af.Unresolved {
				logf(" [SEMA] resolving ident %s ... \n", ident.Name)
				var obj *astObject = pkgScope.Lookup(ident.Name)
				if obj != nil {
					logf(" matched\n")
					ident.Obj = obj
				} else {
					obj  = universe.Lookup(ident.Name)
					if obj != nil {
						logf(" matched\n")
						ident.Obj = obj
					} else {
						// we should allow unresolved for now.
						// e.g foo in X{foo:bar,}
						logf("Unresolved (maybe struct field name in composite literal): "+ident.Name)
						unresolved = append(unresolved, ident)
					}
				}
			}
			for _, dcl := range af.Decls {
				_pkg.Decls = append(_pkg.Decls, dcl)
			}
		}
		pkg = _pkg
		logf("Walking package: %s\n", pkg.name)
		walk(pkg)
		generateCode(pkg)
	}

	// emitting dynamic types
	myfmt.Printf("# ------- Dynamic Types ------\n")
	myfmt.Printf(".data\n")
	for _, te := range typeMap {
		id := te.id
		name := te.serialized
		symbol := typeIdToSymbol(id)
		myfmt.Printf("%s: # %s\n", symbol, name)
		myfmt.Printf("  .quad %d\n", id)
		myfmt.Printf("  .quad .S.dtype.%d\n", id)
		myfmt.Printf("  .quad %d\n", len(name))
		myfmt.Printf(".S.dtype.%d:\n", id)
		myfmt.Printf("  .string \"%s\"\n", name)
	}
	myfmt.Printf("\n")

}

type depEntry struct {
	path string
	children []string
}

func newStmt(x interface{}) astStmt {
	return x
}

func isStmtAssignStmt(s astStmt) bool {
	var ok bool
	_, ok = s.(*astAssignStmt)
	return ok
}

func isStmtCaseClause(s astStmt) bool {
	var ok bool
	_, ok = s.(*astCaseClause)
	return ok
}

func stmt2AssignStmt(s astStmt) *astAssignStmt {
	var r *astAssignStmt
	var ok bool
	r, ok = s.(*astAssignStmt)
	if !ok {
		panic("Not *astAssignStmt")
	}
	return r
}

func stmt2ExprStmt(s astStmt) *astExprStmt {
	var r *astExprStmt
	var ok bool
	r, ok = s.(*astExprStmt)
	if !ok {
		panic("Not *astExprStmt")
	}
	return r
}

func stmt2CaseClause(s astStmt) *astCaseClause {
	var r *astCaseClause
	var ok bool
	r, ok = s.(*astCaseClause)
	if !ok {
		panic("Not *astCaseClause")
	}
	return r
}

func newExpr(expr interface{}) astExpr {
	return expr
}

func expr2Ident(e astExpr) *astIdent {
	var r *astIdent
	var ok bool
	r, ok = e.(*astIdent)
	if ! ok {
		panic(myfmt.Sprintf("Not *astIdent but got: %T", e))
	}
	return r
}

func expr2BinaryExpr(e astExpr) *astBinaryExpr {
	var r *astBinaryExpr
	var ok bool
	r, ok = e.(*astBinaryExpr)
	if ! ok {
		panic("Not *astBinaryExpr")
	}
	return r
}

func expr2UnaryExpr(e astExpr) *astUnaryExpr {
	var r *astUnaryExpr
	var ok bool
	r, ok = e.(*astUnaryExpr)
	if ! ok {
		panic("Not *astUnaryExpr")
	}
	return r
}

func expr2Ellipsis(e astExpr) *astEllipsis {
	var r *astEllipsis
	var ok bool
	r, ok = e.(*astEllipsis)
	if ! ok {
		panic("Not *astEllipsis")
	}
	return r
}

func expr2TypeAssertExpr(e astExpr) *astTypeAssertExpr {
	var r *astTypeAssertExpr
	var ok bool
	r, ok = e.(*astTypeAssertExpr)
	if ! ok {
		panic("Not *astTypeAssertExpr")
	}
	return r
}

func expr2ArrayType(e astExpr) *astArrayType {
	var r *astArrayType
	var ok bool
	r, ok = e.(*astArrayType)
	if ! ok {
		panic("Not *astArrayType")
	}
	return r
}

func expr2BasicLit(e astExpr) *astBasicLit {
	var r *astBasicLit
	var ok bool
	r, ok = e.(*astBasicLit)
	if ! ok {
		panic("Not *astBasicLit")
	}
	return r
}

func expr2StarExpr(e astExpr) *astStarExpr {
	var r *astStarExpr
	var ok bool
	r, ok = e.(*astStarExpr)
	if ! ok {
		panic("Not *astStarExpr")
	}
	return r
}

func expr2KeyValueExpr(e astExpr) *astKeyValueExpr {
	var r *astKeyValueExpr
	var ok bool
	r, ok = e.(*astKeyValueExpr)
	if ! ok {
		panic("Not *astKeyValueExpr")
	}
	return r
}

func expr2StructType(e astExpr) *astStructType {
	var r *astStructType
	var ok bool
	r, ok = e.(*astStructType)
	if ! ok {
		panic("Not *astStructType")
	}
	return r
}

func isExprBasicLit(e astExpr) bool {
	var ok bool
	_, ok = e.(*astBasicLit)
	return ok
}


func isExprStarExpr(e astExpr) bool {
	var ok bool
	_, ok = e.(*astStarExpr)
	return ok
}

func isExprEllipsis(e astExpr) bool {
	var ok bool
	_, ok = e.(*astEllipsis)
	return ok
}

func isExprTypeAssertExpr(e astExpr) bool {
	var ok bool
	_, ok = e.(*astTypeAssertExpr)
	return ok
}

func isExprIdent(e astExpr) bool {
	var ok bool
	_, ok = e.(*astIdent)
	return ok
}

func dtypeOf(x interface{}) string {
	return myfmt.Sprintf("%T", x)
}

