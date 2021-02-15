package main

import (
	"os"
	"syscall"

	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/path"
	"github.com/DQNEO/babygo/lib/strconv"
	"github.com/DQNEO/babygo/lib/strings"
)

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
	var s = fmt.Sprintf(f, a...)
	syscall.Write(1, []uint8(s))
}

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
	fmt.Printf(format2, a...)
}

func evalInt(expr astExpr) int {
	switch e := expr.(type) {
	case *astBasicLit:
		return strconv.Atoi(e.Value)
	default:
		panic("Unknown type")
	}
	return 0
}

func emitPopPrimitive(comment string) {
	fmt.Printf("  popq %%rax # result of %s\n", comment)
}

func emitPopBool(comment string) {
	fmt.Printf("  popq %%rax # result of %s\n", comment)
}

func emitPopAddress(comment string) {
	fmt.Printf("  popq %%rax # address of %s\n", comment)
}

func emitPopString() {
	fmt.Printf("  popq %%rax # string.ptr\n")
	fmt.Printf("  popq %%rcx # string.len\n")
}

func emitPopInterFace() {
	fmt.Printf("  popq %%rax # eface.dtype\n")
	fmt.Printf("  popq %%rcx # eface.data\n")
}

func emitPopSlice() {
	fmt.Printf("  popq %%rax # slice.ptr\n")
	fmt.Printf("  popq %%rcx # slice.len\n")
	fmt.Printf("  popq %%rdx # slice.cap\n")
}


func emitPushStackTop(condType *Type, offset int, comment string) {
	switch kind(condType) {
	case T_STRING:
		fmt.Printf("  movq %d+8(%%rsp), %%rcx # copy str.len from stack top (%s)\n", offset, comment)
		fmt.Printf("  movq %d+0(%%rsp), %%rax # copy str.ptr from stack top (%s)\n", offset, comment)
		fmt.Printf("  pushq %%rcx # str.len\n")
		fmt.Printf("  pushq %%rax # str.ptr\n")
	case T_POINTER, T_UINTPTR, T_BOOL, T_INT, T_UINT8, T_UINT16:
		fmt.Printf("  movq %d(%%rsp), %%rax # copy stack top value (%s) \n", offset, comment)
		fmt.Printf("  pushq %%rax\n")
	default:
		throw(kind(condType))
	}
}

func emitAllocReturnVarsArea(size int) {
	if size == 0 {
		return
	}
	fmt.Printf("  subq $%d, %%rsp # alloc return vars area\n", size)
}

func emitFreeParametersArea(size int) {
	if size == 0 {
		return
	}
	fmt.Printf("  addq $%d, %%rsp # free parameters area\n", size)
}

func emitAddConst(addValue int, comment string) {
	emitComment(2, "Add const: %s\n", comment)
	fmt.Printf("  popq %%rax\n")
	fmt.Printf("  addq $%d, %%rax\n", addValue)
	fmt.Printf("  pushq %%rax\n")
}

// "Load" means copy data from memory to registers
func emitLoadAndPush(t *Type) {
	assert(t != nil, "type should not be nil", __func__)
	emitPopAddress(kind(t))
	switch kind(t) {
	case T_SLICE:
		fmt.Printf("  movq %d(%%rax), %%rdx\n", 16)
		fmt.Printf("  movq %d(%%rax), %%rcx\n", 8)
		fmt.Printf("  movq %d(%%rax), %%rax\n", 0)
		fmt.Printf("  pushq %%rdx # cap\n")
		fmt.Printf("  pushq %%rcx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case T_STRING:
		fmt.Printf("  movq %d(%%rax), %%rdx # len\n", 8)
		fmt.Printf("  movq %d(%%rax), %%rax # ptr\n", 0)
		fmt.Printf("  pushq %%rdx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case T_INTERFACE:
		fmt.Printf("  movq %d(%%rax), %%rdx # data\n", 8)
		fmt.Printf("  movq %d(%%rax), %%rax # dtype\n", 0)
		fmt.Printf("  pushq %%rdx # data\n")
		fmt.Printf("  pushq %%rax # dtype\n")
	case T_UINT8:
		fmt.Printf("  movzbq %d(%%rax), %%rax # load uint8\n", 0)
		fmt.Printf("  pushq %%rax\n")
	case T_UINT16:
		fmt.Printf("  movzwq %d(%%rax), %%rax # load uint16\n", 0)
		fmt.Printf("  pushq %%rax\n")
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		fmt.Printf("  movq %d(%%rax), %%rax # load int\n", 0)
		fmt.Printf("  pushq %%rax\n")
	case T_ARRAY, T_STRUCT:
		// pure proxy
		fmt.Printf("  pushq %%rax\n")
	default:
		panic2(__func__, "TBI:kind="+kind(t))
	}
}

func emitVariableAddr(variable *Variable) {
	emitComment(2, "emit Addr of variable \"%s\" \n", variable.name)

	if variable.isGlobal {
		fmt.Printf("  leaq %s(%%rip), %%rax # global variable \"%s\"\n", variable.globalSymbol, variable.name)
	} else {
		fmt.Printf("  leaq %d(%%rbp), %%rax # local variable \"%s\"\n", variable.localOffset, variable.name)
	}

	fmt.Printf("  pushq %%rax # variable address\n")
}

func emitListHeadAddr(list astExpr) {
	var t = getTypeOfExpr(list)
	switch kind(t) {
	case T_ARRAY:
		emitAddr(list) // array head
	case T_SLICE:
		emitExpr(list, nil)
		emitPopSlice()
		fmt.Printf("  pushq %%rax # slice.ptr\n")
	case T_STRING:
		emitExpr(list, nil)
		emitPopString()
		fmt.Printf("  pushq %%rax # string.ptr\n")
	default:
		panic2(__func__, "kind="+kind(getTypeOfExpr(list)))
	}
}

func emitAddr(expr astExpr) {
	emitComment(2, "[emitAddr] %T\n", expr)
	switch e := expr.(type) {
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
		emitExpr(e.X, nil)
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
			panic2(__func__, "TBI "+knd)
		}
	default:
		panic2(__func__, "TBI "+dtypeOf(expr))
	}
}

func isType(expr astExpr) bool {
	switch e := expr.(type) {
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
		return isType(e.X)
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
	switch tt := to.(type) {
	case *astIdent:
		switch tt.Obj {
		case gString: // string(e)
			switch kind(getTypeOfExpr(arg0)) {
			case T_SLICE: // string(slice)
				emitExpr(arg0, nil) // slice
				emitPopSlice()
				fmt.Printf("  pushq %%rcx # str len\n")
				fmt.Printf("  pushq %%rax # str ptr\n")
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
			} else {
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
		fmt.Printf("  pushq %%rcx # cap\n")
		fmt.Printf("  pushq %%rcx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case *astParenExpr:
		emitConversion(e2t(tt.X), arg0)
	case *astStarExpr: // (*T)(e)
		emitComment(2, "[emitConversion] to pointer \n")
		emitExpr(arg0, nil)
	case *astInterfaceType:
		emitExpr(arg0, nil)
		if isInterface(getTypeOfExpr(arg0)) {
			// do nothing
		} else {
			// Convert dynamic value to interface
			emitConvertToInterface(getTypeOfExpr(arg0))
		}
	default:
		panic2(__func__, "TBI :"+dtypeOf(to))
	}
}

func emitZeroValue(t *Type) {
	switch kind(t) {
	case T_SLICE:
		fmt.Printf("  pushq $0 # slice cap\n")
		fmt.Printf("  pushq $0 # slice len\n")
		fmt.Printf("  pushq $0 # slice ptr\n")
	case T_STRING:
		fmt.Printf("  pushq $0 # string len\n")
		fmt.Printf("  pushq $0 # string ptr\n")
	case T_INTERFACE:
		fmt.Printf("  pushq $0 # interface data\n")
		fmt.Printf("  pushq $0 # interface dtype\n")
	case T_INT, T_UINTPTR, T_UINT8, T_POINTER, T_BOOL:
		fmt.Printf("  pushq $0 # %s zero value\n", kind(t))
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
		fmt.Printf("  pushq %%rcx # len\n")
	case T_STRING:
		emitExpr(arg, nil)
		emitPopString()
		fmt.Printf("  pushq %%rcx # len\n")
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
		fmt.Printf("  pushq %%rdx # cap\n")
	case T_STRING:
		panic("cap() cannot accept string type")
	default:
		throw(kind(getTypeOfExpr(arg)))
	}
}

func emitCallMalloc(size int) {
	// call malloc and return pointer
	ff := lookupForeignFunc("runtime", "malloc")
	emitAllocReturnVarsAreaFF(ff)
	fmt.Printf("  pushq $%d\n", size)
	emitCallFF(ff)
}

func emitStructLiteral(e *astCompositeLit) {
	// allocate heap area with zero value
	fmt.Printf("  # Struct literal\n")
	var structType = e2t(e.Type)
	emitZeroValue(structType) // push address of the new storage
	var kvExpr *astKeyValueExpr
	for i, elm := range e.Elts {
		kvExpr = expr2KeyValueExpr(elm)
		fieldName := expr2Ident(kvExpr.Key)
		emitComment(2, "  - [%d] : key=%s, value=%s\n", i, fieldName.Name, dtypeOf(kvExpr.Value))
		var field = lookupStructField(getStructTypeSpec(structType), fieldName.Name)
		var fieldType = e2t(field.Type)
		var fieldOffset = getStructFieldOffset(field)
		// push lhs address
		emitPushStackTop(tUintptr, 0, "address of struct heaad")
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
		emitPushStackTop(tUintptr, 0, "malloced address")
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
	fmt.Printf("  xor $1, %%rax\n")
	fmt.Printf("  pushq %%rax\n")
}

func emitTrue() {
	fmt.Printf("  pushq $1 # true\n")
}

func emitFalse() {
	fmt.Printf("  pushq $0 # false\n")
}

type Arg struct {
	e         astExpr
	paramType *Type // expected type
	offset    int
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
		arg.paramType = paramType
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
			e:         newExpr(sliceLiteral),
			paramType: e2t(eSliceType),
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
		_arg.paramType = e2t(param.Type)
		args = append(args, _arg)
	}

	if receiver != nil {
		var receiverAndArgs []*Arg = []*Arg{
			&Arg{
				e:         receiver,
				paramType: getTypeOfExpr(receiver),
			},
		}

		for _, a := range args {
			receiverAndArgs = append(receiverAndArgs, a)
		}
		return receiverAndArgs
	}

	return args
}

// see "ABI of stack layout" in the emitFuncall comment
func emitCall(symbol string, args []*Arg, results []*astField) {
	emitComment(2, "emitArgs len=%d\n", len(args))
	var totalParamSize int
	for _, arg := range args {
		arg.offset = totalParamSize
		totalParamSize = totalParamSize + getSizeOfType(arg.paramType)
	}
	var totalReturnSize int
	for _, r := range results {
		totalReturnSize = totalReturnSize + getSizeOfType(e2t(r.Type))
	}
	emitAllocReturnVarsArea(totalReturnSize)
	fmt.Printf("  subq $%d, %%rsp # alloc parameters area\n", totalParamSize)
	for _, arg := range args {
		paramType := arg.paramType
		ctx := &evalContext{
			_type: paramType,
		}
		emitExprIfc(arg.e, ctx)
		emitPop(kind(paramType))
		fmt.Printf("  leaq %d(%%rsp), %%rsi # place to save\n", arg.offset)
		fmt.Printf("  pushq %%rsi # place to save\n")
		emitRegiToMem(paramType)
	}


	emitCallQ(symbol, totalParamSize, results)
}

func emitAllocReturnVarsAreaFF(ff *ForeignFunc) {
	emitAllocReturnVarsArea(getTotalFieldsSize(ff.decl.Type.Results))
}

func getTotalFieldsSize(flist  *astFieldList) int {
	if flist == nil {
		return 0
	}
	var r int
	for _, fld := range flist.List {
		r = r + getSizeOfType(e2t(fld.Type))
	}
	return r
}

func emitCallFF(ff *ForeignFunc) {
	var totalParamSize int = getTotalFieldsSize(ff.decl.Type.Params)
	if ff.decl.Type.Results == nil {
		emitCallQ(ff.symbol, totalParamSize, nil)
	} else {
		emitCallQ(ff.symbol, totalParamSize, ff.decl.Type.Results.List)
	}
}

func emitCallQ(symbol string, totalParamSize int, results []*astField) {
	var totalReturnSize int
	for _, r := range results {
		totalReturnSize = totalReturnSize + getSizeOfType(e2t(r.Type))
	}

	fmt.Printf("  callq %s\n", symbol)
	emitFreeParametersArea(totalParamSize)
	//fmt.Printf("#  totalReturnSize=%d\n", totalReturnSize)
	emitFreeAndPushReturnedValue(results)
}

// callee
func emitReturnStmt(s *astReturnStmt) {
	node := s.node
	fnc := node.fnc
	if len(fnc.retvars) != len(s.Results) {
		panic("length of return and func type do not match")
	}

	var i int
	_len := len(s.Results)
	for i=0;i<_len;i++ {
		emitAssignToVar(fnc.retvars[i], s.Results[i])
	}
	fmt.Printf("  leave\n")
	fmt.Printf("  ret\n")
}

// caller
func emitFreeAndPushReturnedValue(resultList []*astField) {
	switch len(resultList) {
	case 0:
		// do nothing
	case 1:
		var retval0 = resultList[0]
		var knd = kind(e2t(retval0.Type))
		switch knd {
		case T_STRING, T_INTERFACE:
		case T_UINT8:
			fmt.Printf("  movzbq (%%rsp), %%rax # load uint8\n")
			fmt.Printf("  addq $%d, %%rsp # free returnvars area\n", 1)
			fmt.Printf("  pushq %%rax\n")
		case T_BOOL, T_INT, T_UINTPTR, T_POINTER:
		case T_SLICE:
		default:
			panic2(__func__, "Unexpected kind="+knd)
		}
	default:
		//panic("TBI")
	}
}

// ABI of stack layout in function call
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
// call f(i1 int, i2 int) (r1 int, r2 int)
//   -- stack top
//   i1
//   i2
//   r1
//   r2
//
// call f(i int, s string, slc []T) int
//   -- stack top
//   i
//   s.ptr
//   s.len
//   slc.ptr
//   slc.len
//   slc.cap
//   r
//   --
func emitFuncall(fun astExpr, eArgs []astExpr, hasEllissis bool) {
	var symbol string
	var receiver astExpr
	var funcType *astFuncType
	switch fn := fun.(type) {
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
						e:         eNumLit,
						paramType: tInt,
					},
					// len
					&Arg{
						e:         eArgs[1],
						paramType: tInt,
					},
					// cap
					&Arg{
						e:         eArgs[2],
						paramType: tInt,
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
					e:         sliceArg,
					paramType: e2t(generalSlice),
				},
				// element
				&Arg{
					e:         elemArg,
					paramType: elmType,
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
				e:         eArgs[0],
				paramType: tEface,
			}}
			emitCall(symbol, _args, nil)
			return
		}

		if fn.Name == "print" {
			emitExpr(eArgs[0], nil)
			fmt.Printf("  callq runtime.printstring\n")
			fmt.Printf("  addq $%d, %%rsp # revert \n", 16)
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
			panic2(__func__, "TBI selectorExpr.X="+dtypeOf(selectorExpr.X))
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
				ff := lookupForeignFunc(xIdent.Name, fn.Sel.Name)
				funcType = ff.decl.Type
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
		panic2(__func__, "TBI fun.dtype="+dtypeOf(fun))
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
	switch e := expr.(type) {
	case *astIdent:
		ident := e
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
				emitLoadAndPush(t)
			case astCon:
				emitNamedConst(ident, ctx)
			case astTyp:
				panic2(__func__, "[*astIdent] Kind Typ should not come here")
			default:
				panic2(__func__, "[*astIdent] unknown Kind="+ident.Obj.Kind+" Name="+ident.Obj.Name)
			}
		}
	case *astIndexExpr:
		emitAddr(e)
		emitLoadAndPush(getTypeOfExpr(e))
	case *astStarExpr:
		emitAddr(e)
		emitLoadAndPush(getTypeOfExpr(e))
	case *astSelectorExpr:
		x := e.X
		if isExprIdent(x) && expr2Ident(x).Obj.Kind == astPkg {
			ident := lookupForeignVar(expr2Ident(x).Name, e.Sel.Name)
			e := newExpr(ident)
			emitExpr(e, ctx)
		} else {
			emitAddr(expr)
			emitLoadAndPush(getTypeOfExpr(expr))
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
		basicLit := e
		switch basicLit.Kind {
		case "INT":
			var ival = strconv.Atoi(basicLit.Value)
			fmt.Printf("  pushq $%d # number literal\n", ival)
		case "STRING":
			var sl = getStringLiteral(basicLit)
			if sl.strlen == 0 {
				// zero value
				emitZeroValue(tString)
			} else {
				fmt.Printf("  pushq $%d # str len\n", sl.strlen)
				fmt.Printf("  leaq %s, %%rax # str ptr\n", sl.label)
				fmt.Printf("  pushq %%rax # str ptr\n")
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
			fmt.Printf("  pushq $%d # convert char literal to int\n", int(char))
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
				fmt.Printf("  popq %%rcx # low\n")
				fmt.Printf("  popq %%rax # orig_cap\n")
				fmt.Printf("  subq %%rcx, %%rax # orig_cap - low\n")
				fmt.Printf("  pushq %%rax # new cap\n")

				// new len = high - low
				if e.High != nil {
					emitExpr(e.High, nil)
				} else {
					// high = len(orig)
					emitLen(e.X)
				}
				emitExpr(low, nil)
				fmt.Printf("  popq %%rcx # low\n")
				fmt.Printf("  popq %%rax # high\n")
				fmt.Printf("  subq %%rcx, %%rax # high - low\n")
				fmt.Printf("  pushq %%rax # new len\n")
			} else {
				// new cap = max - low
				emitExpr(e.Max, nil)
				emitExpr(low, nil)
				fmt.Printf("  popq %%rcx # low\n")
				fmt.Printf("  popq %%rax # max\n")
				fmt.Printf("  subq %%rcx, %%rax # new cap = max - low\n")
				fmt.Printf("  pushq %%rax # new cap\n")
				// new len = high - low
				emitExpr(e.High, nil)
				emitExpr(low, nil)
				fmt.Printf("  popq %%rcx # low\n")
				fmt.Printf("  popq %%rax # high\n")
				fmt.Printf("  subq %%rcx, %%rax # new len = high - low\n")
				fmt.Printf("  pushq %%rax # new len\n")
			}
		case T_STRING:
			// new len = high - low
			if e.High != nil {
				emitExpr(e.High, nil)
			} else {
				emitLen(e.X)
			}
			emitExpr(low, nil)
			fmt.Printf("  popq %%rcx # low\n")
			fmt.Printf("  popq %%rax # high\n")
			fmt.Printf("  subq %%rcx, %%rax # high - low\n")
			fmt.Printf("  pushq %%rax # len\n")
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
			fmt.Printf("  popq %%rax # e.X\n")
			fmt.Printf("  imulq $-1, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "&":
			emitAddr(e.X)
		case "!":
			emitExpr(e.X, nil)
			emitInvertBoolValue()
		default:
			panic2(__func__, "TBI:astUnaryExpr:"+e.Op)
		}
	case *astBinaryExpr:
		binaryExpr := e
			switch binaryExpr.Op {
			case "&&":
				labelid++
				var labelExitWithFalse = fmt.Sprintf(".L.%d.false", labelid)
				var labelExit = fmt.Sprintf(".L.%d.exit", labelid)
				emitExpr(binaryExpr.X, nil) // left
				emitPopBool("left")
				fmt.Printf("  cmpq $1, %%rax\n")
				// exit with false if left is false
				fmt.Printf("  jne %s\n", labelExitWithFalse)

				// if left is true, then eval right and exit
				emitExpr(binaryExpr.Y, nil) // right
				fmt.Printf("  jmp %s\n", labelExit)

				fmt.Printf("  %s:\n", labelExitWithFalse)
				emitFalse()
				fmt.Printf("  %s:\n", labelExit)
			case "||":
				labelid++
				var labelExitWithTrue = fmt.Sprintf(".L.%d.true", labelid)
				var labelExit = fmt.Sprintf(".L.%d.exit", labelid)
				emitExpr(binaryExpr.X, nil) // left
				emitPopBool("left")
				fmt.Printf("  cmpq $1, %%rax\n")
				// exit with true if left is true
				fmt.Printf("  je %s\n", labelExitWithTrue)

				// if left is false, then eval right and exit
				emitExpr(binaryExpr.Y, nil) // right
				fmt.Printf("  jmp %s\n", labelExit)

				fmt.Printf("  %s:\n", labelExitWithTrue)
				emitTrue()
				fmt.Printf("  %s:\n", labelExit)
			case "+":
				if kind(getTypeOfExpr(e.X)) == T_STRING {
					emitCatStrings(e.X, e.Y)
				} else {
					emitExpr(binaryExpr.X, nil) // left
					emitExpr(binaryExpr.Y, nil) // right
					fmt.Printf("  popq %%rcx # right\n")
					fmt.Printf("  popq %%rax # left\n")
					fmt.Printf("  addq %%rcx, %%rax\n")
					fmt.Printf("  pushq %%rax\n")
				}
			case "-":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil) // right
				fmt.Printf("  popq %%rcx # right\n")
				fmt.Printf("  popq %%rax # left\n")
				fmt.Printf("  subq %%rcx, %%rax\n")
				fmt.Printf("  pushq %%rax\n")
			case "*":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil) // right
				fmt.Printf("  popq %%rcx # right\n")
				fmt.Printf("  popq %%rax # left\n")
				fmt.Printf("  imulq %%rcx, %%rax\n")
				fmt.Printf("  pushq %%rax\n")
			case "%":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil) // right
				fmt.Printf("  popq %%rcx # right\n")
				fmt.Printf("  popq %%rax # left\n")
				fmt.Printf("  movq $0, %%rdx # init %%rdx\n")
				fmt.Printf("  divq %%rcx\n")
				fmt.Printf("  movq %%rdx, %%rax\n")
				fmt.Printf("  pushq %%rax\n")
			case "/":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil) // right
				fmt.Printf("  popq %%rcx # right\n")
				fmt.Printf("  popq %%rax # left\n")
				fmt.Printf("  movq $0, %%rdx # init %%rdx\n")
				fmt.Printf("  divq %%rcx\n")
				fmt.Printf("  pushq %%rax\n")
			case "==":
				emitBinaryExprComparison(e.X, e.Y)
			case "!=":
				emitBinaryExprComparison(e.X, e.Y)
				emitInvertBoolValue()
			case "<":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil) // right
				emitCompExpr("setl")
			case "<=":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil) // right
				emitCompExpr("setle")
			case ">":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil) // right
				emitCompExpr("setg")
			case ">=":
				emitExpr(binaryExpr.X, nil) // left
				emitExpr(binaryExpr.Y, nil) // right
				emitCompExpr("setge")
			default:
				panic2(__func__, "# TBI: binary operation for "+binaryExpr.Op)
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
			fmt.Printf("  pushq $%d # slice.cap\n", length)
			fmt.Printf("  pushq $%d # slice.len\n", length)
			fmt.Printf("  pushq %%rax # slice.ptr\n")
		default:
			panic2(__func__, "Unexpected kind="+k)
		}
	case *astTypeAssertExpr:
		emitExpr(e.X, nil)
		fmt.Printf("  popq  %%rax # ifc.dtype\n")
		fmt.Printf("  popq  %%rcx # ifc.data\n")
		fmt.Printf("  pushq %%rax # ifc.data\n")
		typ := e2t(e.Type)
		sType := serializeType(typ)
		_id := getTypeId(sType)
		typeSymbol := typeIdToSymbol(_id)
		// check if type matches
		fmt.Printf("  leaq %s(%%rip), %%rax # ifc.dtype\n", typeSymbol)
		fmt.Printf("  pushq %%rax           # ifc.dtype\n")

		emitCompExpr("sete") // this pushes 1 or 0 in the end
		emitPopBool("type assertion ok value")
		fmt.Printf("  cmpq $1, %%rax\n")

		labelid++
		labelTypeAssertionEnd := fmt.Sprintf(".L.end_type_assertion.%d", labelid)
		labelElse := fmt.Sprintf(".L.unmatch.%d", labelid)
		fmt.Printf("  jne %s # jmp if false\n", labelElse)

		// if matched
		if ctx.okContext != nil {
			emitComment(2, " double value context\n")
			if ctx.okContext.needMain {
				emitExpr(expr2TypeAssertExpr(expr).X, nil)
				fmt.Printf("  popq %%rax # garbage\n")
				emitLoadAndPush(e2t(expr2TypeAssertExpr(expr).Type)) // load dynamic data
			}
			if ctx.okContext.needOk {
				fmt.Printf("  pushq $1 # ok = true\n")
			}
		} else {
			emitComment(2, " single value context\n")
			emitExpr(expr2TypeAssertExpr(expr).X, nil)
			fmt.Printf("  popq %%rax # garbage\n")
			emitLoadAndPush(e2t(expr2TypeAssertExpr(expr).Type)) // load dynamic data
		}

		// exit
		fmt.Printf("  jmp %s\n", labelTypeAssertionEnd)

		// if not matched
		fmt.Printf("  %s:\n", labelElse)
		if ctx.okContext != nil {
			emitComment(2, " double value context\n")
			if ctx.okContext.needMain {
				emitZeroValue(typ)
			}
			if ctx.okContext.needOk {
				fmt.Printf("  pushq $0 # ok = false\n")
			}
		} else {
			emitComment(2, " single value context\n")
			emitZeroValue(typ)
		}

		fmt.Printf("  %s:\n", labelTypeAssertionEnd)
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
	id         int
}

var typeId int = 1

func typeIdToSymbol(id int) string {
	return "dtype." + strconv.Itoa(id)
}

func getTypeId(serialized string) int {
	for _, te := range typeMap {
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
	fmt.Printf("  leaq %s(%%rip), %%rax # type symbol \"%s\"\n", typeSymbol, str)
	fmt.Printf("  pushq %%rax           # type symbol\n")
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
	fmt.Printf("  popq %%rcx # index id\n")
	fmt.Printf("  movq $%d, %%rdx # elm size\n", getSizeOfType(elmType))
	fmt.Printf("  imulq %%rdx, %%rcx\n")
	fmt.Printf("  addq %%rcx, %%rax\n")
	fmt.Printf("  pushq %%rax # addr of element\n")
}

func emitCatStrings(left astExpr, right astExpr) {
	args := []*Arg{
		&Arg{
			e:         left,
			paramType: tString,
		}, &Arg{
			e:         right,
			paramType: tString,
		},
	}
	var resultList = []*astField{
		&astField{
			Type: tString.e,
		},
	}
	emitCall("runtime.catstrings", args, resultList)
}

func emitCompStrings(left astExpr, right astExpr) {
	args := []*Arg{
		&Arg{
			e:         left,
			paramType: tString,
			offset:    0,
		},
		&Arg{
			e:         right,
			paramType: tString,
			offset:    0,
		},
	}
	var resultList = []*astField{
		&astField{
			Type:  tBool.e,
		},
	}
	emitCall("runtime.cmpstrings", args, resultList)
}

func emitBinaryExprComparison(left astExpr, right astExpr) {
	if kind(getTypeOfExpr(left)) == T_STRING {
		emitCompStrings(left, right)
	} else if kind(getTypeOfExpr(left)) == T_INTERFACE {
		var t = getTypeOfExpr(left)
		ff := lookupForeignFunc("runtime", "cmpinterface")
		emitAllocReturnVarsAreaFF(ff)
		emitExpr(left, nil) // left
		ctx := &evalContext{_type: t}
		emitExprIfc(right, ctx) // right
		emitCallFF(ff)
	} else {
		var t = getTypeOfExpr(left)
		emitExpr(left, nil) // left
		ctx := &evalContext{_type: t}
		emitExprIfc(right, ctx) // right
		emitCompExpr("sete")
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
		fmt.Printf("  popq %%rsi # lhs addr\n")
	} else {
		fmt.Printf("  popq %%rsi # lhs addr\n")
		emitPop(knd) // rhs
	}
	if pushLhs {
		fmt.Printf("  pushq %%rsi # lhs addr\n")
	}

	fmt.Printf("  pushq %%rsi # place to save\n")
	emitRegiToMem(t)
}

func emitRegiToMem(t *Type) {
	fmt.Printf("  popq %%rsi # place to save\n")
	k := kind(t)
	switch k {
	case T_SLICE:
		fmt.Printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", 0)
		fmt.Printf("  movq %%rcx, %d(%%rsi) # len to len\n", 8)
		fmt.Printf("  movq %%rdx, %d(%%rsi) # cap to cap\n", 16)
	case T_STRING:
		fmt.Printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", 0)
		fmt.Printf("  movq %%rcx, %d(%%rsi) # len to len\n", 8)
	case T_INTERFACE:
		fmt.Printf("  movq %%rax, %d(%%rsi) # store dtype\n", 0)
		fmt.Printf("  movq %%rcx, %d(%%rsi) # store data\n", 8)
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		fmt.Printf("  movq %%rax, %d(%%rsi) # assign\n", 0)
	case T_UINT16:
		fmt.Printf("  movw %%ax, %d(%%rsi) # assign word\n", 0)
	case T_UINT8:
		fmt.Printf("  movb %%al, %d(%%rsi) # assign byte\n", 0)
	case T_STRUCT, T_ARRAY:
		fmt.Printf("  pushq $%d # size\n", getSizeOfType(t))
		fmt.Printf("  pushq %%rsi # dst lhs\n")
		fmt.Printf("  pushq %%rax # src rhs\n")
		ff := lookupForeignFunc("runtime", "memcopy")
		emitCallFF(ff)
	default:
		panic2(__func__, "TBI:"+k)
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

func emitAssignToVar(vr *Variable, rhs astExpr) {
	emitComment(2, "Assignment: emitAddr(lhs)\n")
	emitVariableAddr(vr)
	emitComment(2, "Assignment: emitExpr(rhs)\n")
	ctx := &evalContext{
		_type: vr.typ,
	}
	emitExprIfc(rhs, ctx)
	emitComment(2, "Assignment: emitStore(getTypeOfExpr(lhs))\n")
	emitStore(vr.typ, true, false)
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
	switch s := stmt.(type) {
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
		var rhs0 = s.Rhs[0]
		if len(s.Lhs) == 2 && isExprTypeAssertExpr(rhs0) {
			emitAssignWithOK(s.Lhs, rhs0)
		} else {
			if len(s.Lhs) == 1 && len(s.Rhs) == 1 {
				// 1 to 1 assignment
				// x = e
				lhs0 := s.Lhs[0]
				var ident *astIdent
				var isIdent bool
				ident, isIdent = lhs0.(*astIdent)
				if isIdent && ident.Name == "_" {
					panic(" _ is not supported yet")
				}
				emitAssign(lhs0, rhs0)
			} else if len(s.Lhs) >= 1 && len(s.Rhs) == 1 {
				// multi-values expr
				// a, b, c = f()
				emitExpr(rhs0, nil) // @TODO interface conversion
				var _callExpr *astCallExpr
				var ok bool
				_callExpr,ok = rhs0.(*astCallExpr)
				assert(ok, "should be a CallExpr", __func__)
				returnTypes := getCallResultTypes(_callExpr)
				fmt.Printf("# returnTypes=%d\n" , len(returnTypes))
				fmt.Printf("# len lhs=%d\n" , len(s.Lhs))
				fmt.Printf("# returnTypes=%d\n" , len(returnTypes))
				assert(len(returnTypes) == len(s.Lhs), fmt.Sprintf("length unmatches %d <=> %d", len(s.Lhs), len(returnTypes) ), __func__)
				length := len(returnTypes)
				for i:=0; i<length;i++ {
					lhs := s.Lhs[i]
					rhsType := returnTypes[i]
					if isBlankIdentifier(lhs) {
						emitPop(kind(rhsType))
					} else {
						switch kind(rhsType) {
						case  T_UINT8:
							// repush stack top
							fmt.Printf("  movzbq (%%rsp), %%rax # load uint8\n")
							fmt.Printf("  addq $%d, %%rsp # free returnvars area\n", 1)
							fmt.Printf("  pushq %%rax\n")
						}
						emitAddr(lhs)
						emitStore(getTypeOfExpr(lhs), false, false)
					}
				}

			}
		}
	case *astReturnStmt:
		emitReturnStmt(s)
	case *astIfStmt:
		emitComment(2, "if\n")

		labelid++
		var labelEndif = ".L.endif." + strconv.Itoa(labelid)
		var labelElse = ".L.else." + strconv.Itoa(labelid)

		emitExpr(s.Cond, nil)
		emitPopBool("if condition")
		fmt.Printf("  cmpq $1, %%rax\n")
		bodyStmt := newStmt(s.Body)
		if s.Else != nil {
			fmt.Printf("  jne %s # jmp if false\n", labelElse)
			emitStmt(bodyStmt) // then
			fmt.Printf("  jmp %s\n", labelEndif)
			fmt.Printf("  %s:\n", labelElse)
			emitStmt(s.Else) // then
		} else {
			fmt.Printf("  jne %s # jmp if false\n", labelEndif)
			emitStmt(bodyStmt) // then
		}
		fmt.Printf("  %s:\n", labelEndif)
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

		fmt.Printf("  %s:\n", labelCond)
		if s.Cond != nil {
			emitExpr(s.Cond, nil)
			emitPopBool("for condition")
			fmt.Printf("  cmpq $1, %%rax\n")
			fmt.Printf("  jne %s # jmp if false\n", labelExit)
		}
		emitStmt(blockStmt2Stmt(s.Body))
		fmt.Printf("  %s:\n", labelPost) // used for "continue"
		if s.Post != nil {
			emitStmt(s.Post)
		}
		fmt.Printf("  jmp %s\n", labelCond)
		fmt.Printf("  %s:\n", labelExit)
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
		fmt.Printf("  %s:\n", labelCond)

		emitVariableAddr(s.indexvar)
		emitLoadAndPush(tInt)
		emitVariableAddr(s.lenvar)
		emitLoadAndPush(tInt)
		emitCompExpr("setl")
		emitPopBool(" indexvar < lenvar")
		fmt.Printf("  cmpq $1, %%rax\n")
		fmt.Printf("  jne %s # jmp if false\n", labelExit)

		emitComment(2, "assign list[indexvar] value variables\n")
		var elemType = getTypeOfExpr(s.Value)
		emitAddr(s.Value) // lhs

		emitVariableAddr(s.indexvar)
		emitLoadAndPush(tInt) // index value
		emitListElementAddr(s.X, elemType)

		emitLoadAndPush(elemType)
		emitStore(elemType, true, false)

		// Body
		emitComment(2, "ForRange Body\n")
		emitStmt(blockStmt2Stmt(s.Body))

		// Post statement: Increment indexvar and go next
		emitComment(2, "ForRange Post statement\n")
		fmt.Printf("  %s:\n", labelPost) // used for "continue"
		emitVariableAddr(s.indexvar)     // lhs
		emitVariableAddr(s.indexvar)     // rhs
		emitLoadAndPush(tInt)
		emitAddConst(1, "indexvar value ++")
		emitStore(tInt, true, false)

		if s.Key != nil {
			keyIdent := expr2Ident(s.Key)
			if keyIdent.Name != "_" {
				emitAddr(s.Key)              // lhs
				emitVariableAddr(s.indexvar) // rhs
				emitLoadAndPush(tInt)
				emitStore(tInt, true, false)
			}
		}

		fmt.Printf("  jmp %s\n", labelCond)

		fmt.Printf("  %s:\n", labelExit)

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
		var labelEnd = fmt.Sprintf(".L.switch.%d.exit", labelid)
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
				switch kind(condType) {
				case T_STRING:
					ff := lookupForeignFunc("runtime", "cmpstrings")
					emitAllocReturnVarsAreaFF(ff)
					emitPushStackTop(condType, intSize, "switch expr")
					emitExpr(e, nil)

					emitCallFF(ff)
				case T_INTERFACE:
					ff := lookupForeignFunc("runtime", "cmpinterface")
					emitAllocReturnVarsAreaFF(ff)
					emitPushStackTop(condType, intSize, "switch expr")
					emitExpr(e, nil)
					emitCallFF(ff)
				case T_INT, T_UINT8, T_UINT16, T_UINTPTR, T_POINTER:
					emitPushStackTop(condType, 0, "switch expr")
					emitExpr(e, nil)
					emitCompExpr("sete")
				default:
					panic2(__func__, "Unexpected kind="+kind(condType))
				}

				emitPopBool(" of switch-case comparison")
				fmt.Printf("  cmpq $1, %%rax\n")
				fmt.Printf("  je %s # jump if match\n", labelCase)
			}
		}
		emitComment(2, "End comparison with cases\n")

		// if no case matches, then jump to
		if defaultLabel != "" {
			// default
			fmt.Printf("  jmp %s\n", defaultLabel)
		} else {
			// exit
			fmt.Printf("  jmp %s\n", labelEnd)
		}

		emitRevertStackTop(condType)
		for i, c := range cases {
			cc := stmt2CaseClause(c)
			fmt.Printf("%s:\n", labels[i])
			for _, _s := range cc.Body {
				emitStmt(_s)
			}
			fmt.Printf("  jmp %s\n", labelEnd)
		}
		fmt.Printf("%s:\n", labelEnd)
	case *astTypeSwitchStmt:
		typeSwitch := s.node
		//		assert(ok, "should exist")
		labelid++
		labelEnd := fmt.Sprintf(".L.typeswitch.%d.exit", labelid)

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
				emitPopAddress("type switch subject")
				fmt.Printf("  movq (%%rax), %%rax # dtype\n")
				fmt.Printf("  pushq %%rax # dtype\n")

				emitDtypeSymbol(e2t(e))
				emitCompExpr("sete") // this pushes 1 or 0 in the end
				emitPopBool(" of switch-case comparison")

				fmt.Printf("  cmpq $1, %%rax\n")
				fmt.Printf("  je %s # jump if match\n", labelCase)
			}
		}
		emitComment(2, "End comparison with cases\n")

		// if no case matches, then jump to
		if defaultLabel != "" {
			// default
			fmt.Printf("  jmp %s\n", defaultLabel)
		} else {
			// exit
			fmt.Printf("  jmp %s\n", labelEnd)
		}

		for i, typeSwitchCaseClose := range typeSwitch.cases {
			if typeSwitchCaseClose.variable != nil {
				typeSwitch.assignIdent.Obj.Variable = typeSwitchCaseClose.variable
			}
			fmt.Printf("%s:\n", labels[i])

			for _, _s := range typeSwitchCaseClose.orig.Body {
				if typeSwitchCaseClose.variable != nil {
					// do assignment
					expr := newExpr(typeSwitch.assignIdent)
					emitAddr(expr)
					emitVariableAddr(typeSwitch.subjectVariable)
					emitLoadAndPush(tEface)
					fmt.Printf("  popq %%rax # ifc.dtype\n")
					fmt.Printf("  popq %%rcx # ifc.data\n")
					fmt.Printf("  push %%rcx # ifc.data\n")
					emitLoadAndPush(typeSwitchCaseClose.variableType)

					emitStore(typeSwitchCaseClose.variableType, true, false)
				}

				emitStmt(_s)
			}
			fmt.Printf("  jmp %s\n", labelEnd)
		}
		fmt.Printf("%s:\n", labelEnd)

	case *astBranchStmt:
		var containerFor = s.currentFor
		var labelToGo string
		switch s.Tok {
		case "continue":
			switch s := containerFor.(type) {
			case *astForStmt:
				labelToGo = s.labelPost
			case *astRangeStmt:
				labelToGo = s.labelPost
			default:
				panic2(__func__, "unexpected container dtype="+dtypeOf(containerFor))
			}
			fmt.Printf("jmp %s # continue\n", labelToGo)
		case "break":
			switch s := containerFor.(type) {
			case *astForStmt:
				labelToGo = s.labelExit
			case *astRangeStmt:
				labelToGo = s.labelExit
			default:
				panic2(__func__, "unexpected container dtype="+dtypeOf(containerFor))
			}
			fmt.Printf("jmp %s # break\n", labelToGo)
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
	fmt.Printf("  addq $%d, %%rsp # revert stack top\n", getSizeOfType(t))
}

var labelid int

func getMethodSymbol(method *Method) string {
	var rcvTypeName = method.rcvNamedType
	var subsymbol string
	if method.isPtrMethod {
		subsymbol = "$" + rcvTypeName.Name + "." + method.name // pointer
	} else {
		subsymbol = rcvTypeName.Name + "." + method.name // value
	}

	return getPackageSymbol(method.pkgName, subsymbol)
}

func getPackageSymbol(pkgName string, subsymbol string) string {
	return pkgName + "." + subsymbol
}

func emitFuncDecl(pkgName string, fnc *Func) {
	fmt.Printf("# emitFuncDecl\n")
	var i int
	if len(fnc.params) > 0 {
		for i =0; i<len(fnc.params);  i++{
			v := fnc.params[i]
			logf("  #       params %d %d \"%s\" %s\n", (v.localOffset), getSizeOfType(v.typ), v.name, (kind(v.typ)))
		}
	}
	if len(fnc.retvars) > 0 {
		for i := 0; i < len(fnc.retvars); i++ {
			v := fnc.retvars[i]
			logf("  #       retvars %d %d \"%s\" %s\n", (v.localOffset), getSizeOfType(v.typ), v.name, (kind(v.typ)))
		}
	}

	var localarea = fnc.localarea
	var symbol string
	if fnc.method != nil {
		symbol = getMethodSymbol(fnc.method)
	} else {
		symbol = getPackageSymbol(pkgName, fnc.name)
	}
	fmt.Printf("%s: # args %d, locals %d\n", symbol, int(fnc.argsarea), int(fnc.localarea))
	fmt.Printf("  pushq %%rbp\n")
	fmt.Printf("  movq %%rsp, %%rbp\n")
	if len(fnc.localvars) > 0 {
		for i=len(fnc.vars) -1; i>=0; i--{
			v := fnc.vars[i]
			logf("  # -%d(%%rbp) local variable %d \"%s\"\n", -v.localOffset, getSizeOfType(v.typ), v.name)
		}
	}
	logf("  #  0(%%rbp) previous rbp\n")
	logf("  #  8(%%rbp) return address\n")

	if localarea != 0 {
		fmt.Printf("  subq $%d, %%rsp # local area\n", -localarea)
	}

	if fnc.Body != nil {
		emitStmt(blockStmt2Stmt(fnc.Body))
	}

	fmt.Printf("  leave\n")
	fmt.Printf("  ret\n")
}

func emitGlobalVariableComplex(name *astIdent, t *Type, val astExpr) {
	typeKind := kind(t)
	switch typeKind {
	case T_POINTER:
		fmt.Printf("# init global %s:\n", name.Name)
		lhs := newExpr(name)
		emitAssign(lhs, val)
	}
}

func emitGlobalVariable(pkg *PkgContainer, name *astIdent, t *Type, val astExpr) {
	typeKind := kind(t)
	fmt.Printf("%s.%s: # T %s\n", pkg.name, name.Name, typeKind)
	switch typeKind {
	case T_STRING:
		if val == nil {
			fmt.Printf("  .quad 0\n")
			fmt.Printf("  .quad 0\n")
			return
		}
		switch vl := val.(type) {
		case *astBasicLit:
			var sl = getStringLiteral(vl)
			fmt.Printf("  .quad %s\n", sl.label)
			fmt.Printf("  .quad %d\n", sl.strlen)
		default:
			panic("Unsupported global string value")
		}
	case T_INTERFACE:
		// only zero value
		fmt.Printf("  .quad 0 # dtype\n")
		fmt.Printf("  .quad 0 # data\n")
	case T_BOOL:
		if val == nil {
			fmt.Printf("  .quad 0 # bool zero value\n")
			return
		}
		switch vl := val.(type) {
		case *astIdent:
			switch vl.Obj {
			case gTrue:
				fmt.Printf("  .quad 1 # bool true\n")
			case gFalse:
				fmt.Printf("  .quad 0 # bool false\n")
			default:
				panic2(__func__, "")
			}
		default:
			panic2(__func__, "")
		}
	case T_INT:
		if val == nil {
			fmt.Printf("  .quad 0\n")
			return
		}
		switch vl := val.(type) {
		case *astBasicLit:
			fmt.Printf("  .quad %s\n", vl.Value)
		default:
			panic("Unsupported global value")
		}
	case T_UINT8:
		if val == nil {
			fmt.Printf("  .byte 0\n")
			return
		}
		switch vl := val.(type) {
		case *astBasicLit:
			fmt.Printf("  .byte %s\n", vl.Value)
		default:
			panic("Unsupported global value")
		}
	case T_UINT16:
		if val == nil {
			fmt.Printf("  .word 0\n")
			return
		}
		switch val.(type) {
		case *astBasicLit:
			fmt.Printf("  .word %s\n", expr2BasicLit(val).Value)
		default:
			panic("Unsupported global value")
		}
	case T_POINTER:
		// will be set in the initGlobal func
		fmt.Printf("  .quad 0\n")
	case T_UINTPTR:
		if val != nil {
			panic("Unsupported global value")
		}
		// only zero value
		fmt.Printf("  .quad 0\n")
	case T_SLICE:
		if val != nil {
			panic("Unsupported global value")
		}
		// only zero value
		fmt.Printf("  .quad 0 # ptr\n")
		fmt.Printf("  .quad 0 # len\n")
		fmt.Printf("  .quad 0 # cap\n")
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
			fmt.Printf(zeroValue)
		}
	default:
		panic2(__func__, "TBI:kind="+typeKind)
	}
}

func generateCode(pkg *PkgContainer) {
	fmt.Printf("#===================== generateCode %s =====================\n", pkg.name)
	fmt.Printf(".data\n")
	emitComment(0, "string literals len = %d\n", len(pkg.stringLiterals))
	for _, con := range pkg.stringLiterals {
		emitComment(0, "string literals\n")
		fmt.Printf("%s:\n", con.sl.label)
		fmt.Printf("  .string %s\n", con.sl.value)
	}

	for _, spec := range pkg.vars {
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariable(pkg, spec.Name, t, spec.Value)
	}

	fmt.Printf("\n")
	fmt.Printf(".text\n")
	fmt.Printf("%s.__initGlobals:\n", pkg.name)
	for _, spec := range pkg.vars {
		if spec.Value == nil {
			continue
		}
		val := spec.Value
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariableComplex(spec.Name, t, val)
	}
	fmt.Printf("  ret\n")

	for _, fnc := range pkg.funcs {
		emitFuncDecl(pkg.name, fnc)
	}

	fmt.Printf("\n")
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
	switch e := expr.(type) {
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
			panic2(__func__, "2:Obj="+e.Obj.Name+e.Obj.Kind)
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
		types := getCallResultTypes(e)
		assert(len(types) == 1, "single value is expected", __func__)
		return types[0]
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
		var t = getTypeOfExpr(e.X)
		var ptrType = expr2StarExpr(t.e)
		if ptrType == nil {
			panic2(__func__, "starExpr shoud not be nil")
		}
		return e2t(ptrType.X)
	case *astBinaryExpr:
		binaryExpr := e
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

func fieldList2Types(fldlist *astFieldList) []*Type {
	var r []*Type
	for _ , e2 := range fldlist.List {
		t := e2t(e2.Type)
		r = append(r, t)
	}
	return r
}

func getCallResultTypes(e *astCallExpr) []*Type {
	emitComment(2, "[%s] *astCallExpr\n", __func__)
	var fun = e.Fun
	switch fn := fun.(type) {
	case *astIdent:
		if fn.Obj == nil {
			panic2(__func__, "[astCallExpr] nil Obj is not allowed")
		}
		switch fn.Obj.Kind {
		case astTyp:
			return []*Type{e2t(fun)}
		case astFun:
			switch fn.Obj {
			case gLen, gCap:
				return []*Type{tInt}
			case gNew:
				var starExpr = &astStarExpr{}
				starExpr.X = e.Args[0]
				return []*Type{e2t(newExpr(starExpr))}
			case gMake:
				return []*Type{e2t(e.Args[0])}
			case gAppend:
				return []*Type{e2t(e.Args[0])}
			}
			var decl = fn.Obj.Decl
			if decl == nil {
				panic2(__func__, "decl of function "+fn.Name+" is  nil")
			}
			switch dcl := decl.(type) {
			case *astFuncDecl:
				return fieldList2Types(dcl.Type.Results)
			default:
				panic2(__func__, "[astCallExpr] unknown dtype")
			}
			panic2(__func__, "[astCallExpr] Fun ident "+fn.Name)
		}
	case *astParenExpr: // (X)(e) funcall or conversion
		if isType(fn.X) {
			return []*Type{e2t(fn.X)}
		} else {
			panic("TBI: what should we do ?")
		}
	case *astArrayType:
		return []*Type{e2t(fun)}
	case *astSelectorExpr: // (X).Sel()
		xIdent := expr2Ident(fn.X)
		symbol := xIdent.Name + "." + fn.Sel.Name
		switch symbol {
		case "unsafe.Pointer":
			// unsafe.Pointer(x)
			return []*Type{tUintptr}
		default:
			xIdent := expr2Ident(fn.X)
			if xIdent.Obj == nil {
				panic2(__func__, "xIdent.Obj should not be nil")
			}
			if xIdent.Obj.Kind == astPkg {
				ff := lookupForeignFunc(xIdent.Name, fn.Sel.Name)
				return fieldList2Types(ff.decl.Type.Results)
			} else {
				var xType = getTypeOfExpr(fn.X)
				var method = lookupMethod(xType, fn.Sel)
				assert(len(method.funcType.Results.List) == 1, "func is expected to return a single value", __func__)
				return fieldList2Types(method.funcType.Results)
			}
		}
	case *astInterfaceType:
		return []*Type{tEface}
	default:
		panic2(__func__, "[astCallExpr] dtype="+dtypeOf(e.Fun))
	}
	return nil
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

	switch e := t.e.(type) {
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
			typeSpec, ok = decl.(*astTypeSpec)
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
	if t == nil {
		panic("arg.paramType should not be nil")
	}
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
	typeSpec, ok = ident.Obj.Decl.(*astTypeSpec)
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
	vars []*Variable
	params    []*Variable
	retvars   []*Variable
	funcType  *astFuncType
	rcvType   astExpr
	name      string
	Body      *astBlockStmt
	method    *Method
}

type Method struct {
	pkgName      string
	rcvNamedType *astIdent
	isPtrMethod  bool
	name         string
	funcType     *astFuncType
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
	fnc.params = append(fnc.params, vr)
	return vr
}

func (fnc *Func) registerReturnVariable(name string, t *Type) *Variable {
	vr := newLocalVariable(name, fnc.argsarea, t)
	size := getSizeOfType(t)
	fnc.argsarea = fnc.argsarea + size
	fnc.retvars = append(fnc.retvars, vr)
	return vr
}

func (fnc *Func) registerLocalVariable(name string, t *Type) *Variable {
	assert(t != nil && t.e != nil, "type of local var should not be nil", __func__)
	fnc.localarea = fnc.localarea - getSizeOfType(t)
	vr := newLocalVariable(name, currentFunc.localarea, t)
	fnc.vars = append(fnc.vars, vr)
	return vr
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

	label := fmt.Sprintf(".%s.S%d", pkg.name, pkg.stringIndex)
	pkg.stringIndex++

	sl := &sliteral{
		label:  label,
		strlen: strlen - 2,
		value:  lit.Value,
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
		typ:          t,
	}
	return vr
}

func newLocalVariable(name string, localoffset int, t *Type) *Variable {
	vr := &Variable{
		name:        name,
		isGlobal:    false,
		localOffset: localoffset,
		typ:         t,
	}
	return vr
}

type methodEntry struct {
	name   string
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
		pkgName:      pkgName,
		rcvNamedType: rcvNamedType,
		isPtrMethod:  isPtr,
		name:         funcDecl.Name.Name,
		funcType:     funcDecl.Type,
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
		name:   method.name,
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
			assert(obj.Kind == astVar, obj.Name+" should be ast.Var", __func__)
			walkExpr(rhs)
			// infer type
			var _callExpr *astCallExpr
			var ok bool
			_callExpr, ok = rhs.(*astCallExpr)
			var typ *Type
			if ok {
				types := getCallResultTypes(_callExpr)
				typ = types[0]
			}else  {
				typ = getTypeOfExpr(rhs)
			}
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
			keyIdent.Obj.Variable = currentFunc.registerLocalVariable(keyIdent.Name, keyType)

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
		for _, _s := range s.List {
			walkStmt(_s)
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
			if assignIdent != nil && len(cc.List) > 0 {
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
	qi  string
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
			qi:  pkg.name + "." + funcDecl.Name.Name,
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
			qi:  pkg.name + "." + nameIdent.Name,
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
		var resultFields []*astField

		if funcDecl.Recv != nil { // Method
			paramFields = append(paramFields, funcDecl.Recv.List[0])
		}
		for _, field := range funcDecl.Type.Params.List {
			paramFields = append(paramFields, field)
		}

		if funcDecl.Type.Results != nil {
			for _, field := range funcDecl.Type.Results.List {
				resultFields = append(resultFields, field)
			}
		}

		for _, field := range paramFields {
			obj := field.Name.Obj
			obj.Variable = fnc.registerParamVariable(obj.Name, e2t(field.Type))
		}

		for i, field := range resultFields {
			if field.Name == nil {
				// unnamed retval
				fnc.registerReturnVariable(".r" + strconv.Itoa(i), e2t(field.Type))
			} else {
				panic("TBI: named return variable is not supported")
			}
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
	Kind: astCon, // is it Con ?
	Name: "nil",
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
			if !ok {
				panic("not ident")
			}
			return ident
		}
	}
	return nil
}

type ForeignFunc struct {
	symbol string
	decl *astFuncDecl
}

func lookupForeignFunc(pkg string, identifier string) *ForeignFunc {
	symbol := pkg + "." + identifier
	logf("lookupForeignFunc... %s\n", symbol)
	for _, entry := range ExportedQualifiedIdents {
		logf("  looking into %s\n", entry.qi)
		if entry.qi == symbol {
			var fdecl *astFuncDecl
			var ok bool
			fdecl, ok = entry.any.(*astFuncDecl)
			if !ok {
				panic("not fdecl")
			}
			return &ForeignFunc{
				symbol: symbol,
				decl:   fdecl,
			}
		}
	}
	panic("function not found: " + symbol)
	return nil
}

var pkg *PkgContainer

type PkgContainer struct {
	path           string
	name           string
	files          []string
	astFiles       []*astFile
	vars           []*astValueSpec
	funcs          []*Func
	stringLiterals []*stringLiteralsContainer
	stringIndex    int
	Decls          []astDecl
}

func showHelp() {
	fmt.Printf("Usage:\n")
	fmt.Printf("    babygo version:  show version\n")
	fmt.Printf("    babygo [-DF] [-DG] filename\n")
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
		pth := rawValue[1 : len(rawValue)-1]
		importPaths = append(importPaths, pth)
	}
	return importPaths
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
		fmt.Printf("babygo version 0.1.0  linux/amd64\n")
		return
	} else if os.Args[1] == "help" {
		showHelp()
		return
	} else if os.Args[1] == "panic" {
		panicVersion := strconv.Itoa(mylib.Sum(1, 1))
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

	var importPaths []string = []string{"os"}

	for _, inputFile := range inputFiles {
		logf("input file: \"%s\"\n", inputFile)
		logf("Parsing imports\n")
		_paths := getImportPathsFromFile(inputFile)
		for _, p := range _paths {
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
	var packagesToBuild = []*PkgContainer{pkgRuntime}
	fmt.Printf("# === sorted stdPackagesUsed ===\n")
	for _, _path := range stdPackagesUsed {
		fmt.Printf("#  %s\n", _path)
		packagesToBuild = append(packagesToBuild, &PkgContainer{
			path: _path,
		})
	}

	fmt.Printf("# === sorted extPackagesUsed ===\n")
	for _, _path := range extPackagesUsed {
		fmt.Printf("#  %s\n", _path)
		packagesToBuild = append(packagesToBuild, &PkgContainer{
			path: _path,
		})
	}
	mainPkg := &PkgContainer{
		name:  "main",
		files: inputFiles,
	}
	packagesToBuild = append(packagesToBuild, mainPkg)
	//[]string{"runtime.go"}
	for _, _pkg := range packagesToBuild {
		if len(_pkg.files) == 0 {
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
					obj = universe.Lookup(ident.Name)
					if obj != nil {
						logf(" matched\n")
						ident.Obj = obj
					} else {
						// we should allow unresolved for now.
						// e.g foo in X{foo:bar,}
						logf("Unresolved (maybe struct field name in composite literal): " + ident.Name)
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
	fmt.Printf("# ------- Dynamic Types ------\n")
	fmt.Printf(".data\n")
	for _, te := range typeMap {
		id := te.id
		name := te.serialized
		symbol := typeIdToSymbol(id)
		fmt.Printf("%s: # %s\n", symbol, name)
		fmt.Printf("  .quad %d\n", id)
		fmt.Printf("  .quad .S.dtype.%d\n", id)
		fmt.Printf("  .quad %d\n", len(name))
		fmt.Printf(".S.dtype.%d:\n", id)
		fmt.Printf("  .string \"%s\"\n", name)
	}
	fmt.Printf("\n")

}

type depEntry struct {
	path     string
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
	if !ok {
		panic(fmt.Sprintf("Not *astIdent but got: %T", e))
	}
	return r
}

func expr2BinaryExpr(e astExpr) *astBinaryExpr {
	var r *astBinaryExpr
	var ok bool
	r, ok = e.(*astBinaryExpr)
	if !ok {
		panic("Not *astBinaryExpr")
	}
	return r
}

func expr2UnaryExpr(e astExpr) *astUnaryExpr {
	var r *astUnaryExpr
	var ok bool
	r, ok = e.(*astUnaryExpr)
	if !ok {
		panic("Not *astUnaryExpr")
	}
	return r
}

func expr2Ellipsis(e astExpr) *astEllipsis {
	var r *astEllipsis
	var ok bool
	r, ok = e.(*astEllipsis)
	if !ok {
		panic("Not *astEllipsis")
	}
	return r
}

func expr2TypeAssertExpr(e astExpr) *astTypeAssertExpr {
	var r *astTypeAssertExpr
	var ok bool
	r, ok = e.(*astTypeAssertExpr)
	if !ok {
		panic("Not *astTypeAssertExpr")
	}
	return r
}

func expr2ArrayType(e astExpr) *astArrayType {
	var r *astArrayType
	var ok bool
	r, ok = e.(*astArrayType)
	if !ok {
		panic("Not *astArrayType")
	}
	return r
}

func expr2BasicLit(e astExpr) *astBasicLit {
	var r *astBasicLit
	var ok bool
	r, ok = e.(*astBasicLit)
	if !ok {
		panic("Not *astBasicLit")
	}
	return r
}

func expr2StarExpr(e astExpr) *astStarExpr {
	var r *astStarExpr
	var ok bool
	r, ok = e.(*astStarExpr)
	if !ok {
		panic("Not *astStarExpr")
	}
	return r
}

func expr2KeyValueExpr(e astExpr) *astKeyValueExpr {
	var r *astKeyValueExpr
	var ok bool
	r, ok = e.(*astKeyValueExpr)
	if !ok {
		panic("Not *astKeyValueExpr")
	}
	return r
}

func expr2StructType(e astExpr) *astStructType {
	var r *astStructType
	var ok bool
	r, ok = e.(*astStructType)
	if !ok {
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
	return fmt.Sprintf("%T", x)
}
