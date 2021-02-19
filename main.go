package main

import (
	"os"
	"syscall"

	"github.com/DQNEO/babygo/lib/ast"
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

func unexpectedKind(knd TypeKind) {
	panic("Unexpected Kind: " + string(knd))
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
	f := "# " + format
	s := fmt.Sprintf(f, a...)
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
	format2 := string(spaces) + "# " + format
	fmt.Printf(format2, a...)
}

func evalInt(expr ast.Expr) int {
	switch e := expr.(type) {
	case *ast.BasicLit:
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

func emitPushStackTop(condType *ast.Type, offset int, comment string) {
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
		unexpectedKind(kind(condType))
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
func emitLoadAndPush(t *ast.Type) {
	assert(t != nil, "type should not be nil", __func__)
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
		unexpectedKind(kind(t))
	}
}

func emitVariableAddr(variable *ast.Variable) {
	emitComment(2, "emit Addr of variable \"%s\" \n", variable.Name)

	if variable.IsGlobal {
		fmt.Printf("  leaq %s(%%rip), %%rax # global variable \"%s\"\n", variable.GlobalSymbol, variable.Name)
	} else {
		fmt.Printf("  leaq %d(%%rbp), %%rax # local variable \"%s\"\n", variable.LocalOffset, variable.Name)
	}

	fmt.Printf("  pushq %%rax # variable address\n")
}

func emitListHeadAddr(list ast.Expr) {
	t := getTypeOfExpr(list)
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
		unexpectedKind(kind(t))
	}
}

func emitAddr(expr ast.Expr) {
	emitComment(2, "[emitAddr] %T\n", expr)
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Name == "_" {
			panic(" \"_\" has no address")
		}
		if e.Obj == nil {
			panic("ident.Obj is nil: " + e.Name)
		}
		if e.Obj.Kind == ast.Var {
			vr := obj2var(e.Obj)
			emitVariableAddr(vr)
		} else {
			panic2(__func__, "Unexpected Kind "+e.Obj.Kind)
		}
	case *ast.IndexExpr:
		emitExpr(e.Index, nil) // index number
		list := e.X
		elmType := getTypeOfExpr(expr)
		emitListElementAddr(list, elmType)
	case *ast.StarExpr:
		emitExpr(e.X, nil)
	case *ast.SelectorExpr: // (X).Sel
		typeOfX := getTypeOfExpr(e.X)
		var structType *ast.Type
		switch kind(typeOfX) {
		case T_STRUCT:
			// strct.field
			structType = typeOfX
			emitAddr(e.X)
		case T_POINTER:
			// ptr.field
			ptrType := typeOfX.E.(*ast.StarExpr)
			structType = e2t(ptrType.X)
			emitExpr(e.X, nil)
		default:
			unexpectedKind(kind(typeOfX))
		}
		var field = lookupStructField(getStructTypeSpec(structType), e.Sel.Name)
		var offset = getStructFieldOffset(field)
		emitAddConst(offset, "struct head address + struct.field offset")
	case *ast.CompositeLit:
		var knd = kind(getTypeOfExpr(expr))
		switch knd {
		case T_STRUCT:
			// result of evaluation of a struct literal is its address
			emitExpr(expr, nil)
		default:
			unexpectedKind(knd)
		}
	default:
		panic2(__func__, "TBI "+dtypeOf(expr))
	}
}

func isType(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.ArrayType:
		return true
	case *ast.Ident:
		if e == nil {
			panic2(__func__, "ident should not be nil")
		}
		if e.Obj == nil {
			panic2(__func__, " unresolved ident:"+e.Name)
		}
		emitComment(2, "[isType][DEBUG] e.Name = %s\n", e.Name)
		emitComment(2, "[isType][DEBUG] e.Obj = %s,%s\n",
			e.Obj.Name, e.Obj.Kind)
		return e.Obj.Kind == ast.Typ
	case *ast.SelectorExpr:
		if isQI(e) {
			qi := selector2QI(e)
			ident := lookupForeignIdent(qi)
			if ident.Obj.Kind == ast.Typ {
				return true
			}
		}
	case *ast.ParenExpr:
		return isType(e.X)
	case *ast.StarExpr:
		return isType(e.X)
	case *ast.InterfaceType:
		return true
	default:
		emitComment(2, "[isType][%s] is not considered a type\n", dtypeOf(expr))
	}

	return false

}

// explicit conversion T(e)
func emitConversion(toType *ast.Type, arg0 ast.Expr) {
	emitComment(2, "[emitConversion]\n")
	var to = toType.E
	switch tt := to.(type) {
	case *ast.Ident:
		switch tt.Obj {
		case gString: // string(e)
			switch kind(getTypeOfExpr(arg0)) {
			case T_SLICE: // string(slice)
				emitExpr(arg0, nil) // slice
				emitPopSlice()
				fmt.Printf("  pushq %%rcx # str len\n")
				fmt.Printf("  pushq %%rax # str ptr\n")
			case T_STRING: // string(string)
				emitExpr(arg0, nil)
			default:
				unexpectedKind(kind(getTypeOfExpr(arg0)))
			}
		case gInt, gUint8, gUint16, gUintptr: // int(e)
			emitComment(2, "[emitConversion] to int \n")
			emitExpr(arg0, nil)
		default:
			if tt.Obj.Kind == ast.Typ {
				var isTypeSpec bool
				_, isTypeSpec = tt.Obj.Decl.(*ast.TypeSpec)
				if !isTypeSpec {
					panic2(__func__, "Something is wrong")
				}
				//e2t(tt.Obj.Decl.typeSpec.Type))
				emitExpr(arg0, nil)
			} else {
				panic2(__func__, "[*ast.Ident] TBI : "+tt.Obj.Name)
			}
		}
	case *ast.SelectorExpr:
		// pkg.Type(arg0)
		qi := selector2QI(tt)
		if string(qi) == "unsafe.Pointer" {
			emitExpr(arg0, nil)
		} else {
			panic("TBI")
		}
	case *ast.ArrayType: // Conversion to slice
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
	case *ast.ParenExpr:
		emitConversion(e2t(tt.X), arg0)
	case *ast.StarExpr: // (*T)(e)
		emitComment(2, "[emitConversion] to pointer \n")
		emitExpr(arg0, nil)
	case *ast.InterfaceType:
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

func emitZeroValue(t *ast.Type) {
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
		fmt.Printf("  pushq $0 # %s zero value\n", string(kind(t)))
	case T_STRUCT:
		var structSize = getSizeOfType(t)
		emitComment(2, "zero value of a struct. size=%d (allocating on heap)\n", structSize)
		emitCallMalloc(structSize)
	default:
		unexpectedKind(kind(t))
	}
}

func emitLen(arg ast.Expr) {
	emitComment(2, "[%s] begin\n", __func__)
	switch kind(getTypeOfExpr(arg)) {
	case T_ARRAY:
		var typ = getTypeOfExpr(arg)
		var arrayType = expr2ArrayType(typ.E)
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
		unexpectedKind(kind(getTypeOfExpr(arg)))
	}
	emitComment(2, "[%s] end\n", __func__)
}

func emitCap(arg ast.Expr) {
	switch kind(getTypeOfExpr(arg)) {
	case T_ARRAY:
		var typ = getTypeOfExpr(arg)
		var arrayType = expr2ArrayType(typ.E)
		emitExpr(arrayType.Len, nil)
	case T_SLICE:
		emitExpr(arg, nil)
		emitPopSlice()
		fmt.Printf("  pushq %%rdx # cap\n")
	case T_STRING:
		panic("cap() cannot accept string type")
	default:
		unexpectedKind(kind(getTypeOfExpr(arg)))
	}
}

func emitCallMalloc(size int) {
	emitComment(2, "emitCallMalloc\n")
	// call malloc and return pointer
	qi := newQI("runtime", "malloc")
	ff := lookupForeignFunc(qi)
	emitAllocReturnVarsAreaFF(ff)
	fmt.Printf("  pushq $%d\n", size)
	emitCallFF(ff)
}

func emitStructLiteral(e *ast.CompositeLit) {
	// allocate heap area with zero value
	emitComment(2,"emitStructLiteral\n")
	var structType = e2t(e.Type)
	emitZeroValue(structType) // push address of the new storage
	var kvExpr *ast.KeyValueExpr
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

func emitArrayLiteral(arrayType *ast.ArrayType, arrayLen int, elts []ast.Expr) {
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
	e         ast.Expr
	paramType *ast.Type // expected type
	offset    int
}

func prepareArgs(funcType *ast.FuncType, receiver ast.Expr, eArgs []ast.Expr, expandElipsis bool) []*Arg {
	if funcType == nil {
		panic("no funcType")
	}
	var params = funcType.Params.List
	var variadicArgs []ast.Expr
	var variadicElmType ast.Expr
	var args []*Arg
	var param *ast.Field
	var arg *Arg
	lenParams := len(params)
	for argIndex, eArg := range eArgs {
		emitComment(2, "[%s][*ast.Ident][default] loop idx %d, len params %d\n", __func__, argIndex, lenParams)
		if argIndex < lenParams {
			param = params[argIndex]
			if isExprEllipsis(param.Type) {
				ellipsis := expr2Ellipsis(param.Type)
				variadicElmType = ellipsis.Elt
				variadicArgs = make([]ast.Expr, 0, 20)
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
		var sliceType = &ast.ArrayType{}
		sliceType.Elt = variadicElmType
		var eSliceType = sliceType
		var sliceLiteral = &ast.CompositeLit{}
		sliceLiteral.Type = eSliceType
		sliceLiteral.Elts = variadicArgs
		var _arg = &Arg{
			e:         sliceLiteral,
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
func emitCall(symbol string, args []*Arg, resultList *ast.FieldList) {
	emitComment(2, "emitArgs len=%d\n", len(args))

	var totalParamSize int
	for _, arg := range args {
		arg.offset = totalParamSize
		totalParamSize = totalParamSize + getSizeOfType(arg.paramType)
	}

	emitAllocReturnVarsArea(getTotalFieldsSize(resultList))
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

	emitCallQ(symbol, totalParamSize, resultList)
}

func emitAllocReturnVarsAreaFF(ff *ForeignFunc) {
	r := ff.decl.Type.Results
	size := getTotalFieldsSize(r)
	emitAllocReturnVarsArea(size)
}

func getTotalFieldsSize(flist *ast.FieldList) int {
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
	totalParamSize := getTotalFieldsSize(ff.decl.Type.Params)
	emitCallQ(ff.symbol, totalParamSize, ff.decl.Type.Results)
}

func emitCallQ(symbol string, totalParamSize int, resultList *ast.FieldList) {
	fmt.Printf("  callq %s\n", symbol)
	emitFreeParametersArea(totalParamSize)
	fmt.Printf("#  totalReturnSize=%d\n", getTotalFieldsSize(resultList))
	emitFreeAndPushReturnedValue(resultList)
}

// callee
func emitReturnStmt(s *ast.ReturnStmt) {
	node := s.Node
	fnc := node.Fnc
	if len(fnc.Retvars) != len(s.Results) {
		panic("length of return and func type do not match")
	}

	var i int
	_len := len(s.Results)
	for i = 0; i < _len; i++ {
		emitAssignToVar(fnc.Retvars[i], s.Results[i])
	}
	fmt.Printf("  leave\n")
	fmt.Printf("  ret\n")
}

// caller
func emitFreeAndPushReturnedValue(resultList *ast.FieldList) {
	if resultList == nil {
		return
	}
	switch len(resultList.List) {
	case 0:
		// do nothing
	case 1:
		var retval0 = resultList.List[0]
		knd := kind(e2t(retval0.Type))
		switch knd {
		case T_STRING, T_INTERFACE:
		case T_UINT8:
			fmt.Printf("  movzbq (%%rsp), %%rax # load uint8\n")
			fmt.Printf("  addq $%d, %%rsp # free returnvars area\n", 1)
			fmt.Printf("  pushq %%rax\n")
		case T_BOOL, T_INT, T_UINTPTR, T_POINTER:
		case T_SLICE:
		default:
			unexpectedKind(knd)
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
func emitFuncall(fun ast.Expr, eArgs []ast.Expr, hasEllissis bool) {
	var symbol string
	var receiver ast.Expr
	var funcType *ast.FuncType
	switch fn := fun.(type) {
	case *ast.Ident:
		emitComment(2, "[%s][*ast.Ident]\n", __func__)
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
				var arrayType = expr2ArrayType(typeArg.E)
				//assert(ok, "should be *ast.ArrayType")
				var elmSize = getSizeOfType(e2t(arrayType.Elt))
				var numlit = newNumberLiteral(elmSize)
				var eNumLit = numlit

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

				var resultList = &ast.FieldList{
					List: []*ast.Field{
						&ast.Field{
							Type: generalSlice,
						},
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
			resultList := &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: generalSlice,
					},
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
		symbol = getPackageSymbol(currentPkg.name, fn.Name)
		emitComment(2, "[%s][*ast.Ident][default] start\n", __func__)
		if currentPkg.name == "os" && fn.Name == "runtime_args" {
			symbol = "runtime.runtime_args"
		} else if currentPkg.name == "os" && fn.Name == "runtime_getenv" {
			symbol = "runtime.runtime_getenv"
		}

		var obj = fn.Obj
		if obj.Decl == nil {
			panic2(__func__, "[*ast.CallExpr] decl is nil")
		}

		var fndecl *ast.FuncDecl // = decl.funcDecl
		var isFuncDecl bool
		fndecl, isFuncDecl = obj.Decl.(*ast.FuncDecl)
		if !isFuncDecl || fndecl == nil {
			panic2(__func__, "[*ast.CallExpr] fndecl is nil")
		}
		if fndecl.Type == nil {
			panic2(__func__, "[*ast.CallExpr] fndecl.Type is nil")
		}
		funcType = fndecl.Type
	case *ast.SelectorExpr:
		if isQI(fn) {
			// pkg.Sel()
			qi := selector2QI(fn)
			symbol = string(qi)
			ff := lookupForeignFunc(qi)
			funcType = ff.decl.Type
		} else {
			receiver = fn.X
			receiverType := getTypeOfExpr(receiver)
			method := lookupMethod(receiverType, fn.Sel)
			funcType = method.FuncType
			symbol = getMethodSymbol(method)
		}
	case *ast.ParenExpr:
		panic2(__func__, "[astParenExpr] TBI ")
	default:
		panic2(__func__, "TBI fun.dtype="+dtypeOf(fun))
	}

	args := prepareArgs(funcType, receiver, eArgs, hasEllissis)
	emitCall(symbol, args, funcType.Results)
}

func emitNil(targetType *ast.Type) {
	if targetType == nil {
		panic2(__func__, "Type is required to emit nil")
	}
	switch kind(targetType) {
	case T_SLICE, T_POINTER, T_INTERFACE:
		emitZeroValue(targetType)
	default:
		unexpectedKind(kind(targetType))
	}
}

func emitNamedConst(ident *ast.Ident, ctx *evalContext) {
	var valSpec *ast.ValueSpec
	var ok bool
	valSpec, ok = ident.Obj.Decl.(*ast.ValueSpec)
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
	_type     *ast.Type
}

// targetType is the type of someone who receives the expr value.
// There are various forms:
//   Assignment:       x = expr
//   Function call:    x(expr)
//   Return:           return expr
//   CompositeLiteral: T{key:expr}
// targetType is used when:
//   - the expr is nil
//   - the target type is interface and expr is not.
func emitExpr(expr ast.Expr, ctx *evalContext) bool {
	var isNilObject bool
	emitComment(2, "[emitExpr] dtype=%s\n", dtypeOf(expr))
	switch e := expr.(type) {
	case *ast.Ident:
		switch e.Obj {
		case gTrue:
			emitTrue()
		case gFalse:
			emitFalse()
		case gNil:
			assert(ctx._type != nil, "context of nil is not passed", __func__)
			emitNil(ctx._type)
			isNilObject = true
		default:
			assert(e.Obj != nil, "should not be nil", __func__)
			switch e.Obj.Kind {
			case ast.Var:
				emitAddr(expr)
				emitLoadAndPush(getTypeOfExpr(expr))
			case ast.Con:
				emitNamedConst(e, ctx)
			default:
				panic("Unexpected ident kind:" + e.Obj.Kind)
			}
		}
	case *ast.IndexExpr: // 1 or 2 values
		emitAddr(e)
		emitLoadAndPush(getTypeOfExpr(e))
	case *ast.StarExpr: // 1 value
		emitAddr(e)
		emitLoadAndPush(getTypeOfExpr(e))
	case *ast.SelectorExpr: // 1 value X.Sel
		// pkg.Ident or strct.field
		if isQI(e) {
			ident := lookupForeignIdent(selector2QI(e))
			emitExpr(ident, ctx)
		} else {
			// strct.field
			emitAddr(expr)
			emitLoadAndPush(getTypeOfExpr(expr))
		}
	case *ast.CallExpr:
		var fun = e.Fun
		emitComment(2, "[%s][*ast.CallExpr]\n", __func__)
		if isType(fun) {
			emitConversion(e2t(fun), e.Args[0])
		} else {
			emitFuncall(fun, e.Args, e.Ellipsis)
		}
	case *ast.ParenExpr:
		emitExpr(e.X, ctx)
	case *ast.BasicLit:
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
			panic2(__func__, "[*ast.BasicLit] TBI : "+basicLit.Kind)
		}
	case *ast.UnaryExpr:
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
	case *ast.BinaryExpr:
		switch e.Op {
		case "&&":
			labelid++
			var labelExitWithFalse = fmt.Sprintf(".L.%d.false", labelid)
			var labelExit = fmt.Sprintf(".L.%d.exit", labelid)
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
		case "||":
			labelid++
			var labelExitWithTrue = fmt.Sprintf(".L.%d.true", labelid)
			var labelExit = fmt.Sprintf(".L.%d.exit", labelid)
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
		case "+":
			if kind(getTypeOfExpr(e.X)) == T_STRING {
				emitCatStrings(e.X, e.Y)
			} else {
				emitExpr(e.X, nil) // left
				emitExpr(e.Y, nil) // right
				fmt.Printf("  popq %%rcx # right\n")
				fmt.Printf("  popq %%rax # left\n")
				fmt.Printf("  addq %%rcx, %%rax\n")
				fmt.Printf("  pushq %%rax\n")
			}
		case "-":
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			fmt.Printf("  popq %%rcx # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  subq %%rcx, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "*":
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			fmt.Printf("  popq %%rcx # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  imulq %%rcx, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "%":
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			fmt.Printf("  popq %%rcx # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  movq $0, %%rdx # init %%rdx\n")
			fmt.Printf("  divq %%rcx\n")
			fmt.Printf("  movq %%rdx, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "/":
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
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
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			emitCompExpr("setl")
		case "<=":
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			emitCompExpr("setle")
		case ">":
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			emitCompExpr("setg")
		case ">=":
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			emitCompExpr("setge")
		default:
			panic2(__func__, "# TBI: binary operation for "+e.Op)
		}
	case *ast.CompositeLit:
		// slice , array, map or struct

		switch kind(e2t(e.Type)) {
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
			unexpectedKind(kind(e2t(e.Type)))
		}
	case *ast.SliceExpr:
		var list = e.X
		var listType = getTypeOfExpr(list)

		// For convenience, any of the indices may be omitted.
		// A missing low index defaults to zero;
		var low ast.Expr
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
			unexpectedKind(kind(listType))
		}

		emitExpr(low, nil)
		var elmType = getElementTypeOfListType(listType)
		emitListElementAddr(list, elmType)
	case *ast.TypeAssertExpr:
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
		if ctx != nil && ctx.okContext != nil {
			// ok context
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
			// default context is single value context
			emitComment(2, " single value context\n")
			emitExpr(expr2TypeAssertExpr(expr).X, nil)
			fmt.Printf("  popq %%rax # garbage\n")
			emitLoadAndPush(e2t(expr2TypeAssertExpr(expr).Type)) // load dynamic data
		}

		// exit
		fmt.Printf("  jmp %s\n", labelTypeAssertionEnd)

		// if not matched
		fmt.Printf("  %s:\n", labelElse)
		if ctx != nil && ctx.okContext != nil {
			// ok context
			emitComment(2, " double value context\n")
			if ctx.okContext.needMain {
				emitZeroValue(typ)
			}
			if ctx.okContext.needOk {
				fmt.Printf("  pushq $0 # ok = false\n")
			}
		} else {
			// default context is single value context
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
func emitConvertToInterface(fromType *ast.Type) {
	emitComment(2, "ConversionToInterface\n")
	memSize := getSizeOfType(fromType)
	// copy data to heap
	emitCallMalloc(memSize)
	emitStore(fromType, false, true) // heap addr pushed
	// push type id
	emitDtypeSymbol(fromType)
}

func emitExprIfc(expr ast.Expr, ctx *evalContext) {
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

func emitDtypeSymbol(t *ast.Type) {
	str := serializeType(t)
	typeId := getTypeId(str)
	typeSymbol := typeIdToSymbol(typeId)
	fmt.Printf("  leaq %s(%%rip), %%rax # type symbol \"%s\"\n", typeSymbol, str)
	fmt.Printf("  pushq %%rax           # type symbol\n")
}

func newNumberLiteral(x int) *ast.BasicLit {
	var r = &ast.BasicLit{}
	r.Kind = "INT"
	r.Value = strconv.Itoa(x)
	return r
}

func emitListElementAddr(list ast.Expr, elmType *ast.Type) {
	emitListHeadAddr(list)
	emitPopAddress("list head")
	fmt.Printf("  popq %%rcx # index id\n")
	fmt.Printf("  movq $%d, %%rdx # elm size\n", getSizeOfType(elmType))
	fmt.Printf("  imulq %%rdx, %%rcx\n")
	fmt.Printf("  addq %%rcx, %%rax\n")
	fmt.Printf("  pushq %%rax # addr of element\n")
}

func emitCatStrings(left ast.Expr, right ast.Expr) {
	args := []*Arg{
		&Arg{
			e:         left,
			paramType: tString,
		}, &Arg{
			e:         right,
			paramType: tString,
		},
	}
	var resultList = &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type: tString.E,
			},
		},
	}

	emitCall("runtime.catstrings", args, resultList)
}

func emitCompStrings(left ast.Expr, right ast.Expr) {
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
	var resultList = &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type: tBool.E,
			},
		},
	}
	emitCall("runtime.cmpstrings", args, resultList)
}

func emitBinaryExprComparison(left ast.Expr, right ast.Expr) {
	if kind(getTypeOfExpr(left)) == T_STRING {
		emitCompStrings(left, right)
	} else if kind(getTypeOfExpr(left)) == T_INTERFACE {
		var t = getTypeOfExpr(left)
		ff := lookupForeignFunc(newQI("runtime", "cmpinterface"))
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

func emitPop(knd TypeKind) {
	switch knd {
	case T_SLICE:
		emitPopSlice()
	case T_STRING:
		emitPopString()
	case T_INTERFACE:
		emitPopInterFace()
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		emitPopPrimitive(string(knd))
	case T_UINT16:
		emitPopPrimitive(string(knd))
	case T_UINT8:
		emitPopPrimitive(string(knd))
	case T_STRUCT, T_ARRAY:
		emitPopPrimitive(string(knd))
	default:
		unexpectedKind(knd)
	}
}

func emitStore(t *ast.Type, rhsTop bool, pushLhs bool) {
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

func emitRegiToMem(t *ast.Type) {
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
		ff := lookupForeignFunc(newQI("runtime", "memcopy"))
		emitCallFF(ff)
	default:
		unexpectedKind(k)
	}
}

func isBlankIdentifier(e ast.Expr) bool {
	if !isExprIdent(e) {
		return false
	}
	return expr2Ident(e).Name == "_"
}

// support assignment of ok syntax. Blank ident is considered.
func emitAssignWithOK(lhss []ast.Expr, rhs ast.Expr) {
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

func emitAssignToVar(vr *ast.Variable, rhs ast.Expr) {
	emitComment(2, "Assignment: emitAddr(lhs)\n")
	emitVariableAddr(vr)
	emitComment(2, "Assignment: emitExpr(rhs)\n")
	ctx := &evalContext{
		_type: vr.Typ,
	}
	emitExprIfc(rhs, ctx)
	emitComment(2, "Assignment: emitStore(getTypeOfExpr(lhs))\n")
	emitStore(vr.Typ, true, false)
}

func emitAssign(lhs ast.Expr, rhs ast.Expr) {
	emitComment(2, "Assignment: emitAddr(lhs:%s)\n", dtypeOf(lhs))
	emitAddr(lhs)
	emitComment(2, "Assignment: emitExpr(rhs)\n")
	ctx := &evalContext{
		_type: getTypeOfExpr(lhs),
	}
	emitExprIfc(rhs, ctx)
	emitStore(getTypeOfExpr(lhs), true, false)
}


func emitStmt(stmt ast.Stmt) {
	emitComment(2, "== Statement %s ==\n", dtypeOf(stmt))
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		for _, s := range s.List {
			emitStmt(s)
		}
	case *ast.ExprStmt:
		emitExpr(s.X, nil)
	case *ast.DeclStmt:
		genDecl := s.Decl.(*ast.GenDecl)
		declSpec := genDecl.Spec
		switch spec := declSpec.(type) {
		case *ast.ValueSpec:
			valSpec := spec
			t := e2t(valSpec.Type)
			lhs := valSpec.Name
			if valSpec.Value == nil {
				emitComment(2, "lhs addresss\n")
				emitAddr(lhs)
				emitComment(2, "emitZeroValue for %s\n", dtypeOf(t.E))
				emitZeroValue(t)
				emitComment(2, "Assignment: zero value\n")
				emitStore(t, true, false)
			} else {
				rhs := valSpec.Value
				emitAssign(lhs, rhs)
			}
		default:
			panic(declSpec)
		}
	case *ast.AssignStmt:
		switch s.Tok {
		case "=", ":=":
			rhs0 := s.Rhs[0]
			_, isTypeAssertion := rhs0.(*ast.TypeAssertExpr)
			if len(s.Lhs) == 2 && isTypeAssertion {
				emitAssignWithOK(s.Lhs, rhs0)
			} else {
				if len(s.Lhs) == 1 && len(s.Rhs) == 1 {
					// 1 to 1 assignment
					// x = e
					lhs0 := s.Lhs[0]
					var ident *ast.Ident
					var isIdent bool
					ident, isIdent = lhs0.(*ast.Ident)
					if isIdent && ident.Name == "_" {
						panic(" _ is not supported yet")
					}
					emitAssign(lhs0, rhs0)
				} else if len(s.Lhs) >= 1 && len(s.Rhs) == 1 {
					// multi-values expr
					// a, b, c = f()
					emitExpr(rhs0, nil) // @TODO interface conversion
					var _callExpr *ast.CallExpr
					var ok bool
					_callExpr, ok = rhs0.(*ast.CallExpr)
					assert(ok, "should be a CallExpr", __func__)
					returnTypes := getCallResultTypes(_callExpr)
					fmt.Printf("# len lhs=%d\n", len(s.Lhs))
					fmt.Printf("# returnTypes=%d\n", len(returnTypes))
					assert(len(returnTypes) == len(s.Lhs), fmt.Sprintf("length unmatches %d <=> %d", len(s.Lhs), len(returnTypes)), __func__)
					length := len(returnTypes)
					for i := 0; i < length; i++ {
						lhs := s.Lhs[i]
						rhsType := returnTypes[i]
						if isBlankIdentifier(lhs) {
							emitPop(kind(rhsType))
						} else {
							switch kind(rhsType) {
							case T_UINT8:
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
		default:
			panic("TBI: assignment of " + s.Tok)
		}
	case *ast.ReturnStmt:
		emitReturnStmt(s)
	case *ast.IfStmt:
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
	case *ast.ForStmt:
		labelid++
		var labelCond = ".L.for.cond." + strconv.Itoa(labelid)
		var labelPost = ".L.for.post." + strconv.Itoa(labelid)
		var labelExit = ".L.for.exit." + strconv.Itoa(labelid)
		//forStmt, ok := mapForNodeToFor[s]
		//assert(ok, "map value should exist")
		s.LabelPost = labelPost
		s.LabelExit = labelExit

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
		emitStmt(s.Body)
		fmt.Printf("  %s:\n", labelPost) // used for "continue"
		if s.Post != nil {
			emitStmt(s.Post)
		}
		fmt.Printf("  jmp %s\n", labelCond)
		fmt.Printf("  %s:\n", labelExit)
	case *ast.RangeStmt: // only for array and slice
		labelid++
		var labelCond = ".L.range.cond." + strconv.Itoa(labelid)
		var labelPost = ".L.range.post." + strconv.Itoa(labelid)
		var labelExit = ".L.range.exit." + strconv.Itoa(labelid)

		s.LabelPost = labelPost
		s.LabelExit = labelExit
		// initialization: store len(rangeexpr)
		emitComment(2, "ForRange Initialization\n")

		emitComment(2, "  assign length to lenvar\n")
		// lenvar = len(s.X)
		emitVariableAddr(s.Lenvar)
		emitLen(s.X)
		emitStore(tInt, true, false)

		emitComment(2, "  assign 0 to indexvar\n")
		// indexvar = 0
		emitVariableAddr(s.Indexvar)
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

		emitVariableAddr(s.Indexvar)
		emitLoadAndPush(tInt)
		emitVariableAddr(s.Lenvar)
		emitLoadAndPush(tInt)
		emitCompExpr("setl")
		emitPopBool(" indexvar < lenvar")
		fmt.Printf("  cmpq $1, %%rax\n")
		fmt.Printf("  jne %s # jmp if false\n", labelExit)

		emitComment(2, "assign list[indexvar] value variables\n")
		var elemType = getTypeOfExpr(s.Value)
		emitAddr(s.Value) // lhs

		emitVariableAddr(s.Indexvar)
		emitLoadAndPush(tInt) // index value
		emitListElementAddr(s.X, elemType)

		emitLoadAndPush(elemType)
		emitStore(elemType, true, false)

		// Body
		emitComment(2, "ForRange Body\n")
		emitStmt(s.Body)

		// Post statement: Increment indexvar and go next
		emitComment(2, "ForRange Post statement\n")
		fmt.Printf("  %s:\n", labelPost) // used for "continue"
		emitVariableAddr(s.Indexvar)     // lhs
		emitVariableAddr(s.Indexvar)     // rhs
		emitLoadAndPush(tInt)
		emitAddConst(1, "indexvar value ++")
		emitStore(tInt, true, false)

		if s.Key != nil {
			keyIdent := expr2Ident(s.Key)
			if keyIdent.Name != "_" {
				emitAddr(s.Key)              // lhs
				emitVariableAddr(s.Indexvar) // rhs
				emitLoadAndPush(tInt)
				emitStore(tInt, true, false)
			}
		}

		fmt.Printf("  jmp %s\n", labelCond)

		fmt.Printf("  %s:\n", labelExit)

	case *ast.IncDecStmt:
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
	case *ast.SwitchStmt:
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
			assert(isStmtCaseClause(c), "should be *ast.CaseClause", __func__)
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
					ff := lookupForeignFunc(newQI("runtime", "cmpstrings"))
					emitAllocReturnVarsAreaFF(ff)
					emitPushStackTop(condType, SizeOfInt, "switch expr")
					emitExpr(e, nil)

					emitCallFF(ff)
				case T_INTERFACE:
					ff := lookupForeignFunc(newQI("runtime", "cmpinterface"))
					emitAllocReturnVarsAreaFF(ff)
					emitPushStackTop(condType, SizeOfInt, "switch expr")
					emitExpr(e, nil)
					emitCallFF(ff)
				case T_INT, T_UINT8, T_UINT16, T_UINTPTR, T_POINTER:
					emitPushStackTop(condType, 0, "switch expr")
					emitExpr(e, nil)
					emitCompExpr("sete")
				default:
					unexpectedKind(kind(condType))
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
	case *ast.TypeSwitchStmt:
		typeSwitch := s.Node
		//		assert(ok, "should exist")
		labelid++
		labelEnd := fmt.Sprintf(".L.typeswitch.%d.exit", labelid)

		// subjectVariable = subject
		emitVariableAddr(typeSwitch.SubjectVariable)
		emitExpr(typeSwitch.Subject, nil)
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
				emitVariableAddr(typeSwitch.SubjectVariable)
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

		for i, typeSwitchCaseClose := range typeSwitch.Cases {
			if typeSwitchCaseClose.Variable != nil {
				setVariable(typeSwitch.AssignIdent.Obj, typeSwitchCaseClose.Variable)
			}
			fmt.Printf("%s:\n", labels[i])

			for _, _s := range typeSwitchCaseClose.Orig.Body {
				if typeSwitchCaseClose.Variable != nil {
					// do assignment
					expr := typeSwitch.AssignIdent
					emitAddr(expr)
					emitVariableAddr(typeSwitch.SubjectVariable)
					emitLoadAndPush(tEface)
					fmt.Printf("  popq %%rax # ifc.dtype\n")
					fmt.Printf("  popq %%rcx # ifc.data\n")
					fmt.Printf("  push %%rcx # ifc.data\n")
					emitLoadAndPush(typeSwitchCaseClose.VariableType)

					emitStore(typeSwitchCaseClose.VariableType, true, false)
				}

				emitStmt(_s)
			}
			fmt.Printf("  jmp %s\n", labelEnd)
		}
		fmt.Printf("%s:\n", labelEnd)

	case *ast.BranchStmt:
		var containerFor = s.CurrentFor
		var labelToGo string
		switch s.Tok {
		case "continue":
			switch s := containerFor.(type) {
			case *ast.ForStmt:
				labelToGo = s.LabelPost
			case *ast.RangeStmt:
				labelToGo = s.LabelPost
			default:
				panic2(__func__, "unexpected container dtype="+dtypeOf(containerFor))
			}
			fmt.Printf("jmp %s # continue\n", labelToGo)
		case "break":
			switch s := containerFor.(type) {
			case *ast.ForStmt:
				labelToGo = s.LabelExit
			case *ast.RangeStmt:
				labelToGo = s.LabelExit
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

func emitRevertStackTop(t *ast.Type) {
	fmt.Printf("  addq $%d, %%rsp # revert stack top\n", getSizeOfType(t))
}

var labelid int

func getMethodSymbol(method *ast.Method) string {
	var rcvTypeName = method.RcvNamedType
	var subsymbol string
	if method.IsPtrMethod {
		subsymbol = "$" + rcvTypeName.Name + "." + method.Name // pointer
	} else {
		subsymbol = rcvTypeName.Name + "." + method.Name // value
	}

	return getPackageSymbol(method.PkgName, subsymbol)
}

func getPackageSymbol(pkgName string, subsymbol string) string {
	return pkgName + "." + subsymbol
}

func emitFuncDecl(pkgName string, fnc *ast.Func) {
	fmt.Printf("# emitFuncDecl\n")
	var i int
	if len(fnc.Params) > 0 {
		for i = 0; i < len(fnc.Params); i++ {
			v := fnc.Params[i]
			logf("  #       params %d %d \"%s\" %s\n", (v.LocalOffset), getSizeOfType(v.Typ), v.Name, (kind(v.Typ)))
		}
	}
	if len(fnc.Retvars) > 0 {
		for i := 0; i < len(fnc.Retvars); i++ {
			v := fnc.Retvars[i]
			logf("  #       retvars %d %d \"%s\" %s\n", (v.LocalOffset), getSizeOfType(v.Typ), v.Name, (kind(v.Typ)))
		}
	}

	var symbol string
	if fnc.Method != nil {
		symbol = getMethodSymbol(fnc.Method)
	} else {
		symbol = getPackageSymbol(pkgName, fnc.Name)
	}
	fmt.Printf("%s: # args %d, locals %d\n", symbol, int(fnc.Argsarea), int(fnc.Localarea))
	fmt.Printf("  pushq %%rbp\n")
	fmt.Printf("  movq %%rsp, %%rbp\n")
	if len(fnc.Localvars) > 0 {
		for i = len(fnc.Vars) - 1; i >= 0; i-- {
			v := fnc.Vars[i]
			logf("  # -%d(%%rbp) local variable %d \"%s\"\n", -v.LocalOffset, getSizeOfType(v.Typ), v.Name)
		}
	}
	logf("  #  0(%%rbp) previous rbp\n")
	logf("  #  8(%%rbp) return address\n")

	if fnc.Localarea != 0 {
		fmt.Printf("  subq $%d, %%rsp # local area\n", -fnc.Localarea)
	}

	if fnc.Body != nil {
		emitStmt(fnc.Body)
	}

	fmt.Printf("  leave\n")
	fmt.Printf("  ret\n")
}

func emitGlobalVariableComplex(name *ast.Ident, t *ast.Type, val ast.Expr) {
	typeKind := kind(t)
	switch typeKind {
	case T_POINTER:
		fmt.Printf("# init global %s:\n", name.Name)
		lhs := name
		emitAssign(lhs, val)
	}
}

func emitGlobalVariable(pkg *PkgContainer, name *ast.Ident, t *ast.Type, val ast.Expr) {
	typeKind := kind(t)
	fmt.Printf("%s.%s: # T %s\n", pkg.name, name.Name, string(typeKind))
	switch typeKind {
	case T_STRING:
		if val == nil {
			fmt.Printf("  .quad 0\n")
			fmt.Printf("  .quad 0\n")
			return
		}
		switch vl := val.(type) {
		case *ast.BasicLit:
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
		case *ast.Ident:
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
		case *ast.BasicLit:
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
		case *ast.BasicLit:
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
		case *ast.BasicLit:
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
		var arrayType = expr2ArrayType(t.E)
		if arrayType.Len == nil {
			panic2(__func__, "global slice is not supported")
		}
		bl := expr2BasicLit(arrayType.Len)
		var length = evalInt(bl)
		emitComment(0, "[emitGlobalVariable] array length uint8=%d\n", length)
		var zeroValue string
		knd := kind(e2t(arrayType.Elt))
		switch knd {
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
			unexpectedKind(knd)
		}

		var i int
		for i = 0; i < length; i++ {
			fmt.Printf(zeroValue)
		}
	default:
		unexpectedKind(typeKind)
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
		var t *ast.Type
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
		var t *ast.Type
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

func emitDynamicTypes(typeMap []*typeEntry) {
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

// --- type ---
type TypeKind string

const T_STRING TypeKind = "T_STRING"
const T_INTERFACE TypeKind = "T_INTERFACE"
const T_SLICE TypeKind = "T_SLICE"
const T_BOOL TypeKind = "T_BOOL"
const T_INT TypeKind = "T_INT"
const T_INT32 TypeKind = "T_INT32"
const T_UINT8 TypeKind = "T_UINT8"
const T_UINT16 TypeKind = "T_UINT16"
const T_UINTPTR TypeKind = "T_UINTPTR"
const T_ARRAY TypeKind = "T_ARRAY"
const T_STRUCT TypeKind = "T_STRUCT"
const T_POINTER TypeKind = "T_POINTER"

func getTypeOfExpr(expr ast.Expr) *ast.Type {
	//emitComment(0, "[%s] start\n", __func__)
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Obj == nil {
			panic(e.Name)
		}
		switch e.Obj.Kind {
		case ast.Var:
			// injected type is the 1st priority
			// this use case happens in type switch with short decl var
			// switch ident := x.(type) {
			// case T:
			//    y := ident // <= type of ident cannot be associated directly with ident
			//
			if e.Obj.Variable != nil {
				return e.Obj.Variable.Typ
			}
			switch decl := e.Obj.Decl.(type) {
			case *ast.ValueSpec:
				var t = &ast.Type{}
				t.E = decl.Type
				return t
			case *ast.Field:
				var t = &ast.Type{}
				t.E = decl.Type
				return t
			case *ast.AssignStmt: // lhs := rhs
				return getTypeOfExpr(decl.Rhs[0])
			default:
				panic2(__func__, "unkown dtype ")
			}
		case ast.Con:
			switch e.Obj {
			case gTrue, gFalse:
				return tBool
			}
			switch decl2 := e.Obj.Decl.(type) {
			case *ast.ValueSpec:
				return e2t(decl2.Type)
			default:
				panic2(__func__, "cannot decide type of cont ="+e.Obj.Name)
			}
		default:
			panic2(__func__, "2:Obj="+e.Obj.Name+e.Obj.Kind)
		}
	case *ast.BasicLit:
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
	case *ast.IndexExpr:
		var list = e.X
		return getElementTypeOfListType(getTypeOfExpr(list))
	case *ast.UnaryExpr:
		switch e.Op {
		case "+":
			return getTypeOfExpr(e.X)
		case "-":
			return getTypeOfExpr(e.X)
		case "!":
			return tBool
		case "&":
			var starExpr = &ast.StarExpr{}
			var t = getTypeOfExpr(e.X)
			starExpr.X = t.E
			return e2t(starExpr)
		case "range":
			listType := getTypeOfExpr(e.X)
			elmType := getElementTypeOfListType(listType)
			return elmType
		default:
			panic2(__func__, "TBI: Op="+e.Op)
		}
	case *ast.CallExpr:
		types := getCallResultTypes(e)
		assert(len(types) == 1, "single value is expected", __func__)
		return types[0]
	case *ast.SliceExpr:
		var underlyingCollectionType = getTypeOfExpr(e.X)
		if kind(underlyingCollectionType) == T_STRING {
			// str2 = str1[n:m]
			return tString
		}
		var elementTyp ast.Expr
		switch underlyingCollectionType.E.(type) {
		case *ast.ArrayType:
			elementTyp = expr2ArrayType(underlyingCollectionType.E).Elt
		}
		var t = &ast.ArrayType{}
		t.Len = nil
		t.Elt = elementTyp
		return e2t(t)
	case *ast.StarExpr:
		var t = getTypeOfExpr(e.X)
		var ptrType = expr2StarExpr(t.E)
		if ptrType == nil {
			panic2(__func__, "starExpr shoud not be nil")
		}
		return e2t(ptrType.X)
	case *ast.BinaryExpr:
		binaryExpr := e
		switch binaryExpr.Op {
		case "==", "!=", "<", ">", "<=", ">=":
			return tBool
		default:
			return getTypeOfExpr(binaryExpr.X)
		}
	case *ast.SelectorExpr:
		if isQI(e) {
			ident := lookupForeignIdent(selector2QI(e))
			return getTypeOfExpr(ident)
		} else {
			structType := getStructTypeOfX(e)
			field := lookupStructField(getStructTypeSpec(structType), e.Sel.Name)
			return e2t(field.Type)
		}
	case *ast.CompositeLit:
		return e2t(e.Type)
	case *ast.ParenExpr:
		return getTypeOfExpr(e.X)
	case *ast.TypeAssertExpr:
		return e2t(expr2TypeAssertExpr(expr).Type)
	case *ast.InterfaceType:
		return tEface
	default:
		panic2(__func__, "TBI:dtype="+dtypeOf(expr))
	}

	panic2(__func__, "nil type is not allowed\n")
	var r *ast.Type
	return r
}

func fieldList2Types(fldlist *ast.FieldList) []*ast.Type {
	var r []*ast.Type
	for _, e2 := range fldlist.List {
		t := e2t(e2.Type)
		r = append(r, t)
	}
	return r
}

func getCallResultTypes(e *ast.CallExpr) []*ast.Type {
	emitComment(2, "[%s] *ast.CallExpr\n", __func__)
	var fun = e.Fun
	switch fn := fun.(type) {
	case *ast.Ident:
		if fn.Obj == nil {
			panic2(__func__, "[astCallExpr] nil Obj is not allowed")
		}
		switch fn.Obj.Kind {
		case ast.Typ:
			return []*ast.Type{e2t(fun)}
		case ast.Fun:
			switch fn.Obj {
			case gLen, gCap:
				return []*ast.Type{tInt}
			case gNew:
				var starExpr = &ast.StarExpr{}
				starExpr.X = e.Args[0]
				return []*ast.Type{e2t(starExpr)}
			case gMake:
				return []*ast.Type{e2t(e.Args[0])}
			case gAppend:
				return []*ast.Type{e2t(e.Args[0])}
			}
			var decl = fn.Obj.Decl
			if decl == nil {
				panic2(__func__, "decl of function "+fn.Name+" is  nil")
			}
			switch dcl := decl.(type) {
			case *ast.FuncDecl:
				return fieldList2Types(dcl.Type.Results)
			default:
				panic2(__func__, "[astCallExpr] unknown dtype")
			}
			panic2(__func__, "[astCallExpr] Fun ident "+fn.Name)
		}
	case *ast.ParenExpr: // (X)(e) funcall or conversion
		if isType(fn.X) {
			return []*ast.Type{e2t(fn.X)}
		} else {
			panic("TBI: what should we do ?")
		}
	case *ast.ArrayType:
		return []*ast.Type{e2t(fun)}
	case *ast.SelectorExpr:
		if isType(fn) {
			return []*ast.Type{e2t(fn)}
		}
		if isQI(fn) {  // pkg.Sel()
			ff := lookupForeignFunc(selector2QI(fn))
			return fieldList2Types(ff.decl.Type.Results)
		} else {  // obj.method()
			var xType = getTypeOfExpr(fn.X)
			var method = lookupMethod(xType, fn.Sel)
			return fieldList2Types(method.FuncType.Results)
		}
	case *ast.InterfaceType:
		return []*ast.Type{tEface}
	default:
		panic2(__func__, "[astCallExpr] dtype="+dtypeOf(e.Fun))
	}
	return nil
}

func e2t(typeExpr ast.Expr) *ast.Type {
	if typeExpr == nil {
		panic2(__func__, "nil is not allowed")
	}
	var r = &ast.Type{}
	r.E = typeExpr
	return r
}

func serializeType(t *ast.Type) string {
	if t == nil {
		panic("nil type is not expected")
	}
	if t.E == generalSlice {
		panic("TBD: generalSlice")
	}

	switch e := t.E.(type) {
	case *ast.Ident:
		if e.Obj == nil {
			panic("Unresolved identifier:" + e.Name)
		}
		if e.Obj.Kind == ast.Var {
			throw("bug?")
		} else if e.Obj.Kind == ast.Typ {
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
				var typeSpec *ast.TypeSpec
				var ok bool
				typeSpec, ok = decl.(*ast.TypeSpec)
				if !ok {
					panic("unexpected dtype")
				}
				pkgName := typeSpec.Name.Obj.PkgName
				return pkgName + "." + typeSpec.Name.Name
			}
		}
	case *ast.StructType:
		return "struct"
	case *ast.ArrayType:
		if e.Len == nil {
			if e.Elt == nil {
				panic(e)
			}
			return "[]" + serializeType(e2t(e.Elt))
		} else {
			return "[" + strconv.Itoa(evalInt(e.Len)) + "]" + serializeType(e2t(e.Elt))
		}
	case *ast.StarExpr:
		return "*" + serializeType(e2t(e.X))
	case *ast.Ellipsis: // x ...T
		panic("TBD: Ellipsis")
	case *ast.InterfaceType:
		return "interface"
	case *ast.SelectorExpr:
		qi := selector2QI(e)
		return string(qi)
	default:
		throw(dtypeOf(t.E))
	}
	return ""
}

func kind(t *ast.Type) TypeKind {
	if t == nil {
		panic2(__func__, "nil type is not expected\n")
	}
	if t.E == generalSlice {
		return T_SLICE
	}

	switch e := t.E.(type) {
	case *ast.Ident:
		var ident = e
		switch ident.Obj.Name {
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
			var typeSpec *ast.TypeSpec
			var ok bool
			typeSpec, ok = decl.(*ast.TypeSpec)
			if !ok {
				panic2(__func__, "unsupported decl :")
			}
			return kind(e2t(typeSpec.Type))
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
	case *ast.InterfaceType:
		return T_INTERFACE
	case *ast.ParenExpr:
		panic(dtypeOf(e))
	case *ast.SelectorExpr:
		ident := lookupForeignIdent(selector2QI(e))
		return kind(e2t(ident))
	default:
		panic2(__func__, "Unkown dtype:"+dtypeOf(t.E))
	}
	panic2(__func__, "error")
	return ""
}

func isInterface(t *ast.Type) bool {
	return kind(t) == T_INTERFACE
}

func getStructTypeOfX(e *ast.SelectorExpr) *ast.Type {
	var typeOfX = getTypeOfExpr(e.X)
	var structType *ast.Type
	switch kind(typeOfX) {
	case T_STRUCT:
		// strct.field => e.X . e.Sel
		structType = typeOfX
	case T_POINTER:
		// ptr.field => e.X . e.Sel
		ptrType := expr2StarExpr(typeOfX.E)
		structType = e2t(ptrType.X)
	default:
		unexpectedKind(kind(typeOfX) )
	}
	return structType
}

func getElementTypeOfListType(t *ast.Type) *ast.Type {
	switch kind(t) {
	case T_SLICE, T_ARRAY:
		switch e := t.E.(type) {
		case *ast.ArrayType:
			return e2t(e.Elt)
		case *ast.Ellipsis:
			return e2t(e.Elt)
		default:
			throw(dtypeOf(t.E))
		}
	case T_STRING:
		return tUint8
	default:
		unexpectedKind(kind(t))
	}
	var r *ast.Type
	return r
}

const SizeOfSlice int = 24
const SizeOfString int = 16
const SizeOfInt int = 8
const SizeOfUint8 int = 1
const SizeOfUint16 int = 2
const SizeOfPtr int = 8
const SizeOfInterface int = 16

func getSizeOfType(t *ast.Type) int {
	switch kind(t) {
	case T_SLICE:
		return SizeOfSlice
	case T_STRING:
		return SizeOfString
	case T_INT:
		return SizeOfInt
	case T_UINTPTR, T_POINTER:
		return SizeOfPtr
	case T_UINT8:
		return SizeOfUint8
	case T_UINT16:
		return SizeOfUint16
	case T_BOOL:
		return SizeOfInt
	case T_INTERFACE:
		return SizeOfInterface
	case T_ARRAY:
		arrayType := t.E.(*ast.ArrayType)
		elemSize := getSizeOfType(e2t(arrayType.Elt))
		return elemSize * evalInt(arrayType.Len)
	case T_STRUCT:
		return calcStructSizeAndSetFieldOffset(getStructTypeSpec(t))
	default:
		unexpectedKind(kind(t))
	}
	return 0
}

func getStructFieldOffset(field *ast.Field) int {
	var offset = field.Offset
	return offset
}

func setStructFieldOffset(field *ast.Field, offset int) {
	field.Offset = offset
}

func getStructFields(structTypeSpec *ast.TypeSpec) []*ast.Field {
	var structType = expr2StructType(structTypeSpec.Type)
	return structType.Fields.List
}

func getStructTypeSpec(typ *ast.Type) *ast.TypeSpec {
	if kind(typ) != T_STRUCT {
		panic2(__func__, "not T_STRUCT")
	}
	var typeName *ast.Ident
	switch t := typ.E.(type) {
	case *ast.Ident:
		typeName = t
	case *ast.SelectorExpr:
		typeName = lookupForeignIdent(selector2QI(t))
	default:
		panic(typ.E)
	}

	var typeSpec *ast.TypeSpec
	var ok bool
	typeSpec, ok = typeName.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		panic2(__func__, "not *ast.TypeSpec")
	}

	return typeSpec
}

func lookupStructField(structTypeSpec *ast.TypeSpec, selName string) *ast.Field {
	var field *ast.Field
	for _, field := range getStructFields(structTypeSpec) {
		if field.Name.Name == selName {
			return field
		}
	}
	panic("Unexpected flow: struct field not found:" + selName)
	return field
}

func calcStructSizeAndSetFieldOffset(structTypeSpec *ast.TypeSpec) int {
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
	lit *ast.BasicLit
	sl  *sliteral
}


//type localoffsetint int //@TODO

func registerParamVariable(fnc *ast.Func, name string, t *ast.Type) *ast.Variable {
	vr := newLocalVariable(name, fnc.Argsarea, t)
	fnc.Argsarea = fnc.Argsarea + getSizeOfType(t)
	fnc.Params = append(fnc.Params, vr)
	return vr
}

func registerReturnVariable(fnc *ast.Func, name string, t *ast.Type) *ast.Variable {
	vr := newLocalVariable(name, fnc.Argsarea, t)
	size := getSizeOfType(t)
	fnc.Argsarea = fnc.Argsarea + size
	fnc.Retvars = append(fnc.Retvars, vr)
	return vr
}

func registerLocalVariable(fnc *ast.Func, name string, t *ast.Type) *ast.Variable {
	assert(t != nil && t.E != nil, "type of local var should not be nil", __func__)
	fnc.Localarea = fnc.Localarea - getSizeOfType(t)
	vr := newLocalVariable(name, currentFunc.Localarea, t)
	fnc.Vars = append(fnc.Vars, vr)
	return vr
}

var currentFunc *ast.Func

func getStringLiteral(lit *ast.BasicLit) *sliteral {
	for _, container := range currentPkg.stringLiterals {
		if container.lit == lit {
			return container.sl
		}
	}

	panic2(__func__, "string literal not found:"+lit.Value)
	var r *sliteral
	return r
}

func registerStringLiteral(lit *ast.BasicLit) {
	logf(" [registerStringLiteral] begin\n")

	if currentPkg.name == "" {
		panic2(__func__, "no pkgName")
	}

	var strlen int
	var vl = []uint8(lit.Value)
	for _, c := range vl {
		if c != '\\' {
			strlen++
		}
	}

	label := fmt.Sprintf(".%s.S%d", currentPkg.name, currentPkg.stringIndex)
	currentPkg.stringIndex++

	sl := &sliteral{
		label:  label,
		strlen: strlen - 2,
		value:  lit.Value,
	}
	logf(" [registerStringLiteral] label=%s, strlen=%d %s\n", sl.label, sl.strlen, sl.value)
	cont := &stringLiteralsContainer{}
	cont.sl = sl
	cont.lit = lit
	currentPkg.stringLiterals = append(currentPkg.stringLiterals, cont)
}

func newGlobalVariable(pkgName string, name string, t *ast.Type) *ast.Variable {
	vr := &ast.Variable{
		Name:         name,
		IsGlobal:     true,
		GlobalSymbol: pkgName + "." + name,
		Typ:          t,
	}
	return vr
}

func newLocalVariable(name string, localoffset int, t *ast.Type) *ast.Variable {
	vr := &ast.Variable{
		Name:        name,
		IsGlobal:    false,
		LocalOffset: localoffset,
		Typ:         t,
	}
	return vr
}

type methodEntry struct {
	name   string
	method *ast.Method
}

type namedTypeEntry struct {
	//name    string
	obj *ast.Object
	methods []*methodEntry
}

var typesWithMethods []*namedTypeEntry

func findNamedType(obj *ast.Object) *namedTypeEntry {
	for _, t := range typesWithMethods {
		if t.obj == obj {
			return t
		}
	}
	return nil
}

type QualifiedIdent string

func newQI(pkg string, ident string) QualifiedIdent {
	return QualifiedIdent(pkg + "." + ident)
}

func isQI(e *ast.SelectorExpr) bool {
	var ident *ast.Ident
	var isIdent bool
	ident, isIdent = e.X.(*ast.Ident)
	if !isIdent {
		return false
	}
	return ident.Obj.Kind == ast.Pkg
}

func selector2QI(e *ast.SelectorExpr) QualifiedIdent {
	var pkgName *ast.Ident
	var isIdent bool
	pkgName, isIdent = e.X.(*ast.Ident)
	if !isIdent {
		panic(e)
	}
	assert(pkgName.Obj != nil, "Obj should not be nil: " + pkgName.Name + "." + e.Sel.Name,  __func__)
	assert(pkgName.Obj.Kind == ast.Pkg, "should be ast.Pkg", __func__)
	return newQI(pkgName.Name, e.Sel.Name)
}



func newMethod(pkgName string, funcDecl *ast.FuncDecl) *ast.Method {
	var rcvType = funcDecl.Recv.List[0].Type
	var isPtr bool
	if isExprStarExpr(rcvType) {
		isPtr = true
		rcvType = expr2StarExpr(rcvType).X
	}

	rcvNamedType := expr2Ident(rcvType)
	var method = &ast.Method{
		PkgName:      pkgName,
		RcvNamedType: rcvNamedType,
		IsPtrMethod:  isPtr,
		Name:         funcDecl.Name.Name,
		FuncType:     funcDecl.Type,
	}
	return method
}

func registerMethod(method *ast.Method) {
	var nt = findNamedType(method.RcvNamedType.Obj)
	if nt == nil {
		nt = &namedTypeEntry{
			obj:    method.RcvNamedType.Obj,
			methods: nil,
		}
		typesWithMethods = append(typesWithMethods, nt)
	}

	var me *methodEntry = &methodEntry{
		name:   method.Name,
		method: method,
	}
	nt.methods = append(nt.methods, me)
}

func lookupMethod(rcvT *ast.Type, methodName *ast.Ident) *ast.Method {
	var rcvType = rcvT.E
	if isExprStarExpr(rcvType) {
		rcvType = expr2StarExpr(rcvType).X
	}

	var nt *namedTypeEntry
	switch typ := rcvType.(type) {
	case *ast.Ident:
		 nt = findNamedType(typ.Obj)
		if nt == nil {
			panic(typ.Name + " has no moethodeiverTypeName:")
		}
	case *ast.SelectorExpr:
		qi := selector2QI(typ)
		t := lookupForeignIdent(qi)
		nt = findNamedType(t.Obj)
		if nt == nil {
			panic(string(qi) + " has no moethodeiverTypeName:")
		}
	}


	for _, me := range nt.methods {
		if me.name == methodName.Name {
			return me.method
		}
	}

	panic("method not found: " + methodName.Name)
	return nil
}

func walkStmt(stmt ast.Stmt) {
	logf(" [%s] begin dtype=%s\n", __func__, dtypeOf(stmt))
	switch s := stmt.(type) {
	case *ast.DeclStmt:
		logf(" [%s] *ast.DeclStmt\n", __func__)
		var declStmt = s
		if declStmt.Decl == nil {
			panic2(__func__, "ERROR\n")
		}
		var dcl = declStmt.Decl
		var genDecl *ast.GenDecl
		var ok bool
		genDecl, ok = dcl.(*ast.GenDecl)
		if !ok {
			panic2(__func__, "[dcl.dtype] internal error")
		}
		var valSpec *ast.ValueSpec
		valSpec, ok = genDecl.Spec.(*ast.ValueSpec)
		if valSpec.Type == nil {
			if valSpec.Value == nil {
				panic2(__func__, "type inference requires a value")
			}
			var _typ = getTypeOfExpr(valSpec.Value)
			if _typ != nil && _typ.E != nil {
				valSpec.Type = _typ.E
			} else {
				panic2(__func__, "type inference failed")
			}
		}
		var typ = valSpec.Type // Type can be nil
		logf(" [walkStmt] valSpec Name=%s, Type=%s\n",
			valSpec.Name.Name, dtypeOf(typ))

		t := e2t(typ)
		setVariable(valSpec.Name.Obj, registerLocalVariable(currentFunc, valSpec.Name.Name, t))
		logf(" var %s offset = %d\n", valSpec.Name.Obj.Name,
			valSpec.Name.Obj.Variable.LocalOffset)
		if valSpec.Value != nil {
			walkExpr(valSpec.Value)
		}
	case *ast.AssignStmt:
		if s.Tok == ":=" {
			// short var decl
			rhs0 := s.Rhs[0]
			walkExpr(rhs0)
			// infer type
			var typ0 *ast.Type
			switch rhs := rhs0.(type) {
			case *ast.CallExpr:
				types := getCallResultTypes(rhs)
				typ0 = types[0]
			case *ast.TypeAssertExpr:
				typ0 = getTypeOfExpr(rhs0)
				if len(s.Lhs) == 2 { // lhs0, lhs1 := x.(T)
					// declare lhs1 as an ok variable
					okObj := s.Lhs[1].(*ast.Ident).Obj
					//throw(okObj)
					setVariable(okObj, registerLocalVariable(currentFunc, okObj.Name, tBool))
				}
			default:
				typ0 = getTypeOfExpr(rhs0)
			}

			if typ0 != nil && typ0.E != nil {
			} else {
				panic("type inference is not supported")
			}
			obj0 := s.Lhs[0].(*ast.Ident).Obj
			setVariable(obj0 , registerLocalVariable(currentFunc, obj0.Name, typ0))
		} else {
			walkExpr(s.Rhs[0])
		}
	case *ast.ExprStmt:
		walkExpr(s.X)
	case *ast.ReturnStmt:
		s.Node = &ast.NodeReturnStmt{
			Fnc: currentFunc,
		}
		for _, rt := range s.Results {
			walkExpr(rt)
		}
	case *ast.IfStmt:
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
	case *ast.ForStmt:
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
	case *ast.RangeStmt:
		walkExpr(s.X)
		s.Outer = currentFor
		currentFor = stmt
		var _s = s.Body
		walkStmt(_s)
		var lenvar = registerLocalVariable(currentFunc, ".range.len", tInt)
		var indexvar = registerLocalVariable(currentFunc,".range.index", tInt)

		if s.Tok == ":=" {
			listType := getTypeOfExpr(s.X)

			keyIdent := expr2Ident(s.Key)
			//@TODO map key can be any type
			//keyType := getKeyTypeOfListType(listType)
			var keyType *ast.Type = tInt
			setVariable(keyIdent.Obj, registerLocalVariable(currentFunc, keyIdent.Name, keyType))

			// determine type of Value
			elmType := getElementTypeOfListType(listType)
			valueIdent := expr2Ident(s.Value)
			setVariable(valueIdent.Obj, registerLocalVariable(currentFunc, valueIdent.Name, elmType))
		}
		s.Lenvar = lenvar
		s.Indexvar = indexvar
		currentFor = s.Outer
	case *ast.IncDecStmt:
		walkExpr(s.X)
	case *ast.BlockStmt:
		for _, _s := range s.List {
			walkStmt(_s)
		}
	case *ast.BranchStmt:
		s.CurrentFor = currentFor
	case *ast.SwitchStmt:
		if s.Tag != nil {
			walkExpr(s.Tag)
		}
		walkStmt(s.Body)
	case *ast.TypeSwitchStmt:
		typeSwitch := &ast.NodeTypeSwitchStmt{}
		s.Node = typeSwitch
		var assignIdent *ast.Ident
		switch s2 := s.Assign.(type) {
		case *ast.ExprStmt:
			typeAssertExpr := expr2TypeAssertExpr(s2.X)
			//assert(ok, "should be *ast.TypeAssertExpr")
			typeSwitch.Subject = typeAssertExpr.X
			walkExpr(typeAssertExpr.X)
		case *ast.AssignStmt:
			lhs := s2.Lhs[0]
			//var ok bool
			assignIdent = expr2Ident(lhs)
			//assert(ok, "lhs should be ident")
			typeSwitch.AssignIdent = assignIdent
			// ident will be a new local variable in each case clause
			typeAssertExpr := expr2TypeAssertExpr(s2.Rhs[0])
			//assert(ok, "should be *ast.TypeAssertExpr")
			typeSwitch.Subject = typeAssertExpr.X
			walkExpr(typeAssertExpr.X)
		default:
			throw(dtypeOf(s.Assign))
		}

		typeSwitch.SubjectVariable = registerLocalVariable(currentFunc, ".switch_expr", tEface)
		for _, _case := range s.Body.List {
			cc := stmt2CaseClause(_case)
			tscc := &ast.TypeSwitchCaseClose{
				Orig: cc,
			}
			typeSwitch.Cases = append(typeSwitch.Cases, tscc)
			if assignIdent != nil && len(cc.List) > 0 {
				// inject a variable of that type
				varType := e2t(cc.List[0])
				vr := registerLocalVariable(currentFunc, assignIdent.Name, varType)
				tscc.Variable = vr
				tscc.VariableType = varType
				setVariable(assignIdent.Obj, vr)
			}

			for _, s_ := range cc.Body {
				walkStmt(s_)
			}
			if assignIdent != nil {
				assignIdent.Obj.Variable = nil
			}
		}
	case *ast.CaseClause:
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

var currentFor ast.Stmt

func walkExpr(expr ast.Expr) {
	logf(" [walkExpr] dtype=%s\n", dtypeOf(expr))
	switch e := expr.(type) {
	case *ast.Ident:
		// what to do ?
	case *ast.CallExpr:
		walkExpr(e.Fun)
		// Replace __func__ ident by a string literal
		var basicLit *ast.BasicLit
		var newArg ast.Expr
		for i, arg := range e.Args {
			if isExprIdent(arg) {
				ident := expr2Ident(arg)
				if ident.Name == "__func__" && ident.Obj.Kind == ast.Var {
					basicLit = &ast.BasicLit{}
					basicLit.Kind = "STRING"
					basicLit.Value = "\"" + currentFunc.Name + "\""
					newArg = basicLit
					e.Args[i] = newArg
					arg = newArg
				}
			}
			walkExpr(arg)
		}
	case *ast.BasicLit:
		basicLit := e
		switch basicLit.Kind {
		case "STRING":
			registerStringLiteral(basicLit)
		}
	case *ast.CompositeLit:
		for _, v := range e.Elts {
			walkExpr(v)
		}
	case *ast.UnaryExpr:
		walkExpr(e.X)
	case *ast.BinaryExpr:
		binaryExpr := e
		walkExpr(binaryExpr.X)
		walkExpr(binaryExpr.Y)
	case *ast.IndexExpr:
		walkExpr(e.Index)
		walkExpr(e.X)
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
	case *ast.StarExpr:
		walkExpr(e.X)
	case *ast.SelectorExpr:
		walkExpr(e.X)
	case *ast.ArrayType: // []T(e)
		// do nothing ?
	case *ast.ParenExpr:
		walkExpr(e.X)
	case *ast.KeyValueExpr:
		walkExpr(e.Key)
		walkExpr(e.Value)
	case *ast.InterfaceType:
		// interface{}(e)  conversion. Nothing to do.
	case *ast.TypeAssertExpr:
		walkExpr(e.X)
	default:
		panic2(__func__, "TBI:"+dtypeOf(expr))
	}
}

var ExportedQualifiedIdents []*exportEntry

type exportEntry struct {
	qi  QualifiedIdent
	any *ast.Ident
}

func lookupForeignIdent(qi QualifiedIdent) *ast.Ident {
	logf("lookupForeignIdent... %s\n", qi)
	for _, entry := range ExportedQualifiedIdents {
		logf("  looking into %s\n", entry.qi)
		if entry.qi == qi {
			return entry.any
		}
	}
	panic("QI not found: " + string(qi))
}

type ForeignFunc struct {
	symbol string
	decl   *ast.FuncDecl
}

func lookupForeignFunc(qi QualifiedIdent) *ForeignFunc {
	logf("lookupForeignFunc... \n")
	ident := lookupForeignIdent(qi)
	decl := ident.Obj.Decl
	var fdecl *ast.FuncDecl
	var ok bool
	fdecl, ok = decl.(*ast.FuncDecl)
	if !ok {
		panic("not fdecl")
	}
	return &ForeignFunc{
		symbol: string(qi),
		decl:   fdecl,
	}
}

// Purpose of walk:
// Global:
// - collect methods
// - collect string literals
// - collect global variables
// - determine struct size and field offset
// Local:
// - collect string literals
// - collect local variables and set offset
// - determine types of variable declarations
func walk(pkg *PkgContainer) {
	var typeSpecs []*ast.TypeSpec
	var funcDecls []*ast.FuncDecl
	var varSpecs []*ast.ValueSpec
	var constSpecs []*ast.ValueSpec

	for _, decl := range pkg.Decls {
		switch dcl := decl.(type) {
		case *ast.GenDecl:
			switch spec := dcl.Spec.(type) {
			case *ast.TypeSpec:
				typeSpecs = append(typeSpecs, spec)
			case *ast.ValueSpec:
				if spec.Name.Obj.Kind == ast.Var {
					varSpecs = append(varSpecs, spec)
				} else if spec.Name.Obj.Kind == ast.Con {
					constSpecs = append(constSpecs, spec)
				} else {
					panic("Unexpected")
				}
			}
		case *ast.FuncDecl:
			funcDecls = append(funcDecls, dcl)
		default:
			panic("Unexpected")
		}
	}

	for _, typeSpec := range typeSpecs {
		typeSpec.Name.Obj.PkgName = pkg.name // package to which the type belongs to
		switch kind(e2t(typeSpec.Type)) {
		case T_STRUCT:
			logf("calcStructSizeAndSetFieldOffset of %s\n", typeSpec.Name.Name)
			calcStructSizeAndSetFieldOffset(typeSpec)
		}
		exportEntry := &exportEntry{
			qi:  newQI(pkg.name , typeSpec.Name.Name),
			any: typeSpec.Name,
		}
		ExportedQualifiedIdents = append(ExportedQualifiedIdents, exportEntry)
	}

	// collect methods in advance
	for _, funcDecl := range funcDecls {
		if funcDecl.Recv == nil { // non-method function
			if funcDecl.Name.Obj == nil {
				panic("funcDecl.Name.Obj is nil:" + funcDecl.Name.Name)
			}
			var fdcl *ast.FuncDecl
			var ok bool
			fdcl , ok = funcDecl.Name.Obj.Decl.(*ast.FuncDecl)
			if !ok || funcDecl != fdcl {
				panic("Bad func decl reference:" +  funcDecl.Name.Name)
			}
			exportEntry := &exportEntry{
				qi:  newQI(pkg.name , funcDecl.Name.Name),
				any: funcDecl.Name,
			}
			ExportedQualifiedIdents = append(ExportedQualifiedIdents, exportEntry)
		} else { // method
			if funcDecl.Body != nil {
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
		assert(nameIdent.Obj.Kind == ast.Var, "should be Var", __func__)
		if valSpec.Type == nil {
			var val = valSpec.Value
			var t = getTypeOfExpr(val)
			valSpec.Type = t.E
		}
		setVariable(nameIdent.Obj,  newGlobalVariable(pkg.name, nameIdent.Obj.Name, e2t(valSpec.Type)))
		pkg.vars = append(pkg.vars, valSpec)
		exportEntry := &exportEntry{
			qi:  newQI(pkg.name, nameIdent.Name),
			any: nameIdent,
		}
		ExportedQualifiedIdents = append(ExportedQualifiedIdents, exportEntry)
		if valSpec.Value != nil {
			walkExpr(valSpec.Value)
		}
	}

	for _, funcDecl := range funcDecls {
		fnc := &ast.Func{
			Name:      funcDecl.Name.Name,
			FuncType:  funcDecl.Type,
			Localarea: 0,
			Argsarea:  16,
		}
		currentFunc = fnc
		logf(" [sema] == ast.FuncDecl %s ==\n", funcDecl.Name.Name)
		//var paramoffset = 16
		var paramFields []*ast.Field
		var resultFields []*ast.Field

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
			setVariable(obj, registerParamVariable(fnc, obj.Name, e2t(field.Type)))
		}

		for i, field := range resultFields {
			if field.Name == nil {
				// unnamed retval
				registerReturnVariable(fnc, ".r"+strconv.Itoa(i), e2t(field.Type))
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
				fnc.Method = newMethod(pkg.name, funcDecl)
			}
			pkg.funcs = append(pkg.funcs, fnc)
		}
	}
}

// --- universe ---
var gNil = &ast.Object{
	Kind: ast.Con, // is it Con ?
	Name: "nil",
}

var identNil = &ast.Ident{
	Obj:  gNil,
	Name: "nil",
}

var eNil ast.Expr
var eZeroInt ast.Expr

var gTrue = &ast.Object{
	Kind: ast.Con,
	Name: "true",
}
var gFalse = &ast.Object{
	Kind: ast.Con,
	Name: "false",
}

var gString = &ast.Object{
	Kind: ast.Typ,
	Name: "string",
}

var gInt = &ast.Object{
	Kind: ast.Typ,
	Name: "int",
}

var gInt32 = &ast.Object{
	Kind: ast.Typ,
	Name: "int32",
}

var gUint8 = &ast.Object{
	Kind: ast.Typ,
	Name: "uint8",
}

var gUint16 = &ast.Object{
	Kind: ast.Typ,
	Name: "uint16",
}
var gUintptr = &ast.Object{
	Kind: ast.Typ,
	Name: "uintptr",
}
var gBool = &ast.Object{
	Kind: ast.Typ,
	Name: "bool",
}

var gNew = &ast.Object{
	Kind: ast.Fun,
	Name: "new",
}

var gMake = &ast.Object{
	Kind: ast.Fun,
	Name: "make",
}
var gAppend = &ast.Object{
	Kind: ast.Fun,
	Name: "append",
}

var gLen = &ast.Object{
	Kind: ast.Fun,
	Name: "len",
}

var gCap = &ast.Object{
	Kind: ast.Fun,
	Name: "cap",
}
var gPanic = &ast.Object{
	Kind: ast.Fun,
	Name: "panic",
}

var tInt *ast.Type
var tInt32 *ast.Type // Rune
var tUint8 *ast.Type
var tUint16 *ast.Type
var tUintptr *ast.Type
var tString *ast.Type
var tEface *ast.Type
var tBool *ast.Type
var generalSlice ast.Expr

func createUniverse() *ast.Scope {
	var universe = new(ast.Scope)

	objects := []*ast.Object{
		gNil,
		// constants
		gTrue, gFalse,
		// types
		gString, gUintptr, gBool, gInt, gUint8, gUint16,
		// funcs
		gNew, gMake, gAppend, gLen, gCap, gPanic,
	}
	for _, obj := range objects {
		universe.Insert(obj)
	}


	// setting aliases
	universe.Objects = append(universe.Objects, &ast.ObjectEntry{
		Name: "byte",
		Obj:  gUint8,
	})

	return universe
}

// --- builder ---
var currentPkg *PkgContainer

type PkgContainer struct {
	path           string
	name           string
	files          []string
	astFiles       []*ast.File
	vars           []*ast.ValueSpec
	funcs          []*ast.Func
	stringLiterals []*stringLiteralsContainer
	stringIndex    int
	Decls          []ast.Decl
}

func resolveImports(file *ast.File) {
	var mapImports []string
	for _, imprt := range file.Imports {
		// unwrap double quote "..."
		rawPath := imprt.Path[1:(len(imprt.Path) - 1)]
		base := path.Base(rawPath)
		mapImports = append(mapImports, base)
	}
	for _, ident := range file.Unresolved {
		if mylib.InArray(ident.Name, mapImports) {
			ident.Obj = &ast.Object{
				Kind: ast.Pkg,
				Name: ident.Name,
			}
			logf("# resolved: %s\n", ident.Name)
		}
	}
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
		if dirent == "." || dirent == ".." || !strings.HasSuffix(dirent, ".go") {
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

func getPackageDir(importPath string) string {
	if isStdLib(importPath) {
		return prjSrcPath + "/" + importPath
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
		if pkgPath == "unsafe" || pkgPath == "runtime" {
			continue
		}
		logf("   in pkgPath=%s\n", pkgPath)
		packageDir := getPackageDir(pkgPath)
		fnames := findFilesInDir(packageDir)
		var children []string
		for _, fname := range fnames {
			_paths := getImportPathsFromFile(packageDir + "/" + fname)
			for _, p := range _paths {
				if p == "unsafe" || p == "runtime" {
					continue
				}
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

type depEntry struct {
	path     string
	children []string
}

var srcPath string
var prjSrcPath string

func collectAllPackages(inputFiles []string) []string {
	var tree []*depEntry
	directChildren := collectDirectDependents(inputFiles)
	tree = collectDependency(tree, directChildren)
	sortedPaths := sortDepTree(tree)

	// sort packages by this order
	// 1: pseudo
	// 2: stdlib
	// 3: external
	paths := []string{"unsafe", "runtime"}
	for _, pth := range sortedPaths {
		if isStdLib(pth) {
			paths = append(paths, pth)
		}
	}

	for _, pth := range sortedPaths {
		if !isStdLib(pth) {
			paths = append(paths, pth)
		}
	}
	return paths
}

func collectDirectDependents(inputFiles []string) []string {
	var importPaths []string

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
	return importPaths
}

func collectSourceFiles(pkgDir string) []string {
	fnames := findFilesInDir(pkgDir)
	var files []string
	for _, fname := range fnames {
		logf("fname: %s\n", fname)
		srcFile := pkgDir + "/" + fname
		files = append(files, srcFile)
	}
	return files
}

func buildPackage(_pkg *PkgContainer, universe *ast.Scope) {
	logf("Building package : %s\n", _pkg.path)
	pkgScope := ast.NewScope(universe)
	for _, file := range _pkg.files {
		logf("Parsing file: %s\n", file)
		af := parseFile(file, false)
		_pkg.name = af.Name
		_pkg.astFiles = append(_pkg.astFiles, af)
		for _, oe := range af.Scope.Objects {
			pkgScope.Objects = append(pkgScope.Objects, oe)
		}
	}
	for _, af := range _pkg.astFiles {
		resolveImports(af)
		logf("[%s] start\n", __func__)
		// inject predeclared identifers
		var unresolved []*ast.Ident
		logf(" [SEMA] resolving af.Unresolved (n=%d)\n", len(af.Unresolved))
		for _, ident := range af.Unresolved {
			logf(" [SEMA] resolving ident %s ... \n", ident.Name)
			var obj *ast.Object = pkgScope.Lookup(ident.Name)
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
	logf("Walking package: %s\n", _pkg.name)
	walk(_pkg)
	generateCode(_pkg)
}

// --- main ---
func showHelp() {
	fmt.Printf("Usage:\n")
	fmt.Printf("    babygo version:  show version\n")
	fmt.Printf("    babygo [-DF] [-DG] filename\n")
}

func main() {
	srcPath = os.Getenv("GOPATH") + "/src"
	prjSrcPath = srcPath + "/github.com/DQNEO/babygo/src/"

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

	eNil = identNil
	eZeroInt = &ast.BasicLit{
		Value: "0",
		Kind:  "INT",
	}
	generalSlice = &ast.Ident{}
	tInt = &ast.Type{
		E: &ast.Ident{
			Name: "int",
			Obj:  gInt,
		},
	}
	tInt32 = &ast.Type{
		E: &ast.Ident{
			Name: "int32",
			Obj:  gInt32,
		},
	}
	tUint8 = &ast.Type{
		E: &ast.Ident{
			Name: "uint8",
			Obj:  gUint8,
		},
	}

	tUint16 = &ast.Type{
		E: &ast.Ident{
			Name: "uint16",
			Obj:  gUint16,
		},
	}
	tUintptr = &ast.Type{
		E: &ast.Ident{
			Name: "uintptr",
			Obj:  gUintptr,
		},
	}

	tString = &ast.Type{
		E: &ast.Ident{
			Name: "string",
			Obj:  gString,
		},
	}

	tEface = &ast.Type{
		E: &ast.InterfaceType{},
	}

	tBool = &ast.Type{
		E: &ast.Ident{
			Name: "bool",
			Obj:  gBool,
		},
	}

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

	paths := collectAllPackages(inputFiles)
	var packagesToBuild []*PkgContainer
	for _, _path := range paths {
		files := collectSourceFiles(getPackageDir(_path))
		packagesToBuild = append(packagesToBuild, &PkgContainer{
			path:  _path,
			files: files,
		})
	}

	packagesToBuild = append(packagesToBuild, &PkgContainer{
		name:  "main",
		files: inputFiles,
	})

	var universe = createUniverse()
	for _, _pkg := range packagesToBuild {
		currentPkg = _pkg
		buildPackage(_pkg, universe)
	}

	emitDynamicTypes(typeMap)
}

func obj2var(obj *ast.Object) *ast.Variable {
	assert(obj.Kind == ast.Var, "should be ast.Var", __func__)
	assert(obj.Variable != nil, "should not be nil", __func__)
	return obj.Variable
}

func setVariable(obj *ast.Object, vr *ast.Variable) {
	obj.Variable = vr
}
