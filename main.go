package main

import (
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/token"

	"os"
	"syscall"

	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/path"
	"github.com/DQNEO/babygo/lib/strconv"
	"github.com/DQNEO/babygo/lib/strings"
)

var __func__ = "__func__"

func assert(bol bool, msg string, caller string) {
	if !bol {
		panic(caller + ": " + msg)
	}
}

func unexpectedKind(knd TypeKind) {
	panic("Unexpected Kind: " + string(knd))
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
	for i := 0; i < indent; i++ {
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
func emitLoadAndPush(t *Type) {
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

func emitVariableAddr(variable *Variable) {
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
		if e.Obj == nil {
			panic("ident.Obj is nil: " + e.Name)
		}
		assert(e.Obj.Kind == ast.Var, "should be ast.Var", __func__)
		vr := obj2var(e.Obj)
		emitVariableAddr(vr)
	case *ast.IndexExpr:
		emitExpr(e.Index, nil) // index number
		list := e.X
		elmType := getTypeOfExpr(e)
		emitListElementAddr(list, elmType)
	case *ast.StarExpr:
		emitExpr(e.X, nil)
	case *ast.SelectorExpr:
		if isQI(e) { // pkg.SomeType
			ident := lookupForeignIdent(selector2QI(e))
			emitAddr(ident)
		} else { // (e).field
			typeOfX := getUnderlyingType(getTypeOfExpr(e.X))
			var structTypeLiteral *ast.StructType
			switch typ := typeOfX.E.(type) {
			case *ast.StructType: // strct.field
				structTypeLiteral = typ
				emitAddr(e.X)
			case *ast.StarExpr: // ptr.field
				structTypeLiteral = getUnderlyingStructType(e2t(typ.X))
				emitExpr(e.X, nil)
			default:
				unexpectedKind(kind(typeOfX))
			}

			field := lookupStructField(structTypeLiteral, e.Sel.Name)
			offset := getStructFieldOffset(field)
			emitAddConst(offset, "struct head address + struct.field offset")
		}
	case *ast.CompositeLit:
		knd := kind(getTypeOfExpr(e))
		switch knd {
		case T_STRUCT:
			// result of evaluation of a struct literal is its address
			emitExpr(e, nil)
		default:
			unexpectedKind(knd)
		}
	default:
		throw(expr)
	}
}

func isType(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.ArrayType:
		return true
	case *ast.Ident:
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
	}
	return false
}

// explicit conversion T(e)
func emitConversion(toType *Type, arg0 ast.Expr) {
	emitComment(2, "[emitConversion]\n")
	switch to := toType.E.(type) {
	case *ast.Ident:
		switch to.Obj {
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
			emitExpr(arg0, nil)
		default:
			if to.Obj.Kind == ast.Typ {
				emitExpr(arg0, nil)
			} else {
				throw(to.Obj)
			}
		}
	case *ast.SelectorExpr:
		// pkg.Type(arg0)
		qi := selector2QI(to)
		if string(qi) == "unsafe.Pointer" {
			emitExpr(arg0, nil)
		} else {
			ff := lookupForeignIdent(qi)
			assert(ff.Obj.Kind == ast.Typ, "should be ast.Typ", __func__)
			emitConversion(e2t(ff), arg0)
		}
	case *ast.ArrayType: // Conversion to slice
		arrayType := to
		if arrayType.Len != nil {
			throw(to)
		}
		assert(kind(getTypeOfExpr(arg0)) == T_STRING, "source type should be slice", __func__)
		emitComment(2, "Conversion of string => slice \n")
		emitExpr(arg0, nil)
		emitPopString()
		fmt.Printf("  pushq %%rcx # cap\n")
		fmt.Printf("  pushq %%rcx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case *ast.ParenExpr: // (T)(arg0)
		emitConversion(e2t(to.X), arg0)
	case *ast.StarExpr: // (*T)(arg0)
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
		throw(to)
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
		fmt.Printf("  pushq $0 # %s zero value\n", string(kind(t)))
	case T_STRUCT:
		structSize := getSizeOfType(t)
		emitComment(2, "zero value of a struct. size=%d (allocating on heap)\n", structSize)
		emitCallMalloc(structSize)
	default:
		unexpectedKind(kind(t))
	}
}

func emitLen(arg ast.Expr) {
	switch kind(getTypeOfExpr(arg)) {
	case T_ARRAY:
		arrayType := getTypeOfExpr(arg).E.(*ast.ArrayType)
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
}

func emitCap(arg ast.Expr) {
	switch kind(getTypeOfExpr(arg)) {
	case T_ARRAY:
		arrayType := getTypeOfExpr(arg).E.(*ast.ArrayType)
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
	// call malloc and return pointer
	ff := lookupForeignFunc(newQI("runtime", "malloc"))
	emitAllocReturnVarsAreaFF(ff)
	fmt.Printf("  pushq $%d\n", size)
	emitCallFF(ff)
}

func emitStructLiteral(e *ast.CompositeLit) {
	// allocate heap area with zero value
	emitComment(2, "emitStructLiteral\n")
	structType := e2t(e.Type)
	emitZeroValue(structType) // push address of the new storage
	for _, elm := range e.Elts {
		kvExpr := elm.(*ast.KeyValueExpr)
		fieldName := kvExpr.Key.(*ast.Ident)
		field := lookupStructField(getUnderlyingStructType(structType), fieldName.Name)
		fieldType := e2t(field.Type)
		fieldOffset := getStructFieldOffset(field)
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
	elmType := e2t(arrayType.Elt)
	elmSize := getSizeOfType(elmType)
	memSize := elmSize * arrayLen
	emitCallMalloc(memSize) // push
	for i, elm := range elts {
		// push lhs address
		emitPushStackTop(tUintptr, 0, "malloced address")
		emitAddConst(elmSize*i, "malloced address + elmSize * index")
		// push rhs value
		ctx := &evalContext{
			_type: elmType,
		}
		emitExprIfc(elm, ctx)
		// assign
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
	paramType *Type // expected type
	offset    int
}

func prepareArgs(funcType *ast.FuncType, receiver ast.Expr, eArgs []ast.Expr, expandElipsis bool) []*Arg {
	if funcType == nil {
		panic("no funcType")
	}
	var args []*Arg
	params := funcType.Params.List
	var variadicArgs []ast.Expr // nil means there is no variadic in func params
	var variadicElmType ast.Expr
	var param *ast.Field
	lenParams := len(params)
	for argIndex, eArg := range eArgs {
		if argIndex < lenParams {
			param = params[argIndex]
			elp, isEllpsis := param.Type.(*ast.Ellipsis)
			if isEllpsis {
				variadicElmType = elp.Elt
				variadicArgs = make([]ast.Expr, 0, 20)
			}
		}
		if variadicElmType != nil && !expandElipsis {
			variadicArgs = append(variadicArgs, eArg)
			continue
		}

		paramType := e2t(param.Type)
		arg := &Arg{
			e:         eArg,
			paramType: paramType,
		}
		args = append(args, arg)
	}

	if variadicElmType != nil && !expandElipsis {
		// collect args as a slice
		sliceType := &ast.ArrayType{
			Elt: variadicElmType,
		}
		vargsSliceWrapper := &ast.CompositeLit{
			Type: sliceType,
			Elts: variadicArgs,
		}
		args = append(args, &Arg{
			e:         vargsSliceWrapper,
			paramType: e2t(sliceType),
		})
	} else if len(args) < len(params) {
		// Add nil as a variadic arg
		param := params[len(args)]
		elp := param.Type.(*ast.Ellipsis)
		args = append(args, &Arg{
			e:         eNil,
			paramType: e2t(elp),
		})
	}

	if receiver != nil { // method call
		var receiverAndArgs []*Arg = []*Arg{
			&Arg{
				e:         receiver,
				paramType: getTypeOfExpr(receiver),
			},
		}
		for _, arg := range args {
			receiverAndArgs = append(receiverAndArgs, arg)
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
		totalParamSize += getSizeOfType(arg.paramType)
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
	emitAllocReturnVarsArea(getTotalFieldsSize(ff.decl.Type.Results))
}

func getTotalFieldsSize(flist *ast.FieldList) int {
	if flist == nil {
		return 0
	}
	var r int
	for _, fld := range flist.List {
		r += getSizeOfType(e2t(fld.Type))
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
	meta := getMetaReturnStmt(s)
	fnc := meta.Fnc
	if len(fnc.Retvars) != len(s.Results) {
		panic("length of return and func type do not match")
	}

	_len := len(s.Results)
	for i := 0; i < _len; i++ {
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
		retval0 := resultList.List[0]
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
	var funcType *ast.FuncType
	var symbol string
	var receiver ast.Expr
	switch fn := fun.(type) {
	case *ast.Ident:
		// check if it's a builtin func
		switch fn.Obj {
		case gLen:
			emitLen(eArgs[0])
			return
		case gCap:
			emitCap(eArgs[0])
			return
		case gNew:
			typeArg := e2t(eArgs[0])
			// size to malloc
			size := getSizeOfType(typeArg)
			emitCallMalloc(size)
			return
		case gMake:
			typeArg := e2t(eArgs[0])
			switch kind(typeArg) {
			case T_SLICE:
				// make([]T, ...)
				arrayType := getUnderlyingType(typeArg).E.(*ast.ArrayType)
				elmSize := getSizeOfType(e2t(arrayType.Elt))
				numlit := newNumberLiteral(elmSize)
				args := []*Arg{
					// elmSize
					&Arg{
						e:         numlit,
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

				resultList := &ast.FieldList{
					List: []*ast.Field{
						&ast.Field{
							Type: generalSlice,
						},
					},
				}
				emitCall("runtime.makeSlice", args, resultList)
				return
			default:
				throw(typeArg)
			}
		case gAppend:
			sliceArg := eArgs[0]
			elemArg := eArgs[1]
			elmType := getElementTypeOfListType(getTypeOfExpr(sliceArg))
			elmSize := getSizeOfType(elmType)
			args := []*Arg{
				// slice
				&Arg{
					e:         sliceArg,
					paramType: e2t(generalSlice),
				},
				// elm
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
				throw(elmSize)
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

		if fn.Name == "makeSlice1" || fn.Name == "makeSlice8" || fn.Name == "makeSlice16" || fn.Name == "makeSlice24" {
			fn.Name = "makeSlice"
		}
		// general function call
		symbol = getPackageSymbol(currentPkg.name, fn.Name)
		if currentPkg.name == "os" && fn.Name == "runtime_args" {
			symbol = "runtime.runtime_args"
		} else if currentPkg.name == "os" && fn.Name == "runtime_getenv" {
			symbol = "runtime.runtime_getenv"
		}

		fndecl := fn.Obj.Decl.(*ast.FuncDecl)
		funcType = fndecl.Type
	case *ast.SelectorExpr:
		if isQI(fn) {
			// pkg.Sel()
			qi := selector2QI(fn)
			symbol = string(qi)
			ff := lookupForeignFunc(qi)
			funcType = ff.decl.Type
		} else {
			// method call
			receiver = fn.X
			receiverType := getTypeOfExpr(receiver)
			method := lookupMethod(receiverType, fn.Sel)
			funcType = method.FuncType
			symbol = getMethodSymbol(method)
		}
	case *ast.ParenExpr:
		panic("[astParenExpr] TBI ")
	default:
		throw(fun)
	}

	args := prepareArgs(funcType, receiver, eArgs, hasEllissis)
	emitCall(symbol, args, funcType.Results)
}

func emitNil(targetType *Type) {
	if targetType == nil {
		panic("Type is required to emit nil")
	}
	switch kind(targetType) {
	case T_SLICE, T_POINTER, T_INTERFACE:
		emitZeroValue(targetType)
	default:
		unexpectedKind(kind(targetType))
	}
}

func emitNamedConst(ident *ast.Ident, ctx *evalContext) {
	valSpec := ident.Obj.Decl.(*ast.ValueSpec)
	lit := valSpec.Values[0].(*ast.BasicLit)
	emitExprIfc(lit, ctx)
}

type okContext struct {
	needMain bool
	needOk   bool
}

type evalContext struct {
	okContext *okContext
	_type     *Type
}

// 1 value
func emitIdent(e *ast.Ident, ctx *evalContext) bool {
	switch e.Obj {
	case gTrue: // true constant
		emitTrue()
	case gFalse: // false constant
		emitFalse()
	case gNil:
		assert(ctx._type != nil, "context of nil is not passed", __func__)
		emitNil(ctx._type)
		return true
	default:
		assert(e.Obj != nil, "should not be nil", __func__)
		switch e.Obj.Kind {
		case ast.Var:
			emitAddr(e)
			emitLoadAndPush(getTypeOfExpr(e))
		case ast.Con:
			emitNamedConst(e, ctx)
		default:
			panic("Unexpected ident kind:")
		}
	}
	return false
}

// 1 or 2 values
func emitIndexExpr(e *ast.IndexExpr, ctx *evalContext) {
	emitAddr(e)
	emitLoadAndPush(getTypeOfExpr(e))
}

// 1 value
func emitStarExpr(e *ast.StarExpr, ctx *evalContext) {
	emitAddr(e)
	emitLoadAndPush(getTypeOfExpr(e))
}

// 1 value X.Sel
func emitSelectorExpr(e *ast.SelectorExpr, ctx *evalContext) {
	// pkg.Ident or strct.field
	if isQI(e) {
		ident := lookupForeignIdent(selector2QI(e))
		emitExpr(ident, ctx)
	} else {
		// strct.field
		emitAddr(e)
		emitLoadAndPush(getTypeOfExpr(e))
	}
}

// multi values Fun(Args)
func emitCallExpr(e *ast.CallExpr, ctx *evalContext) {
	var fun = e.Fun
	// check if it's a conversion
	if isType(fun) {
		emitConversion(e2t(fun), e.Args[0])
	} else {
		emitFuncall(fun, e.Args, e.Ellipsis != token.NoPos)
	}
}

// multi values (e)
func emitParenExpr(e *ast.ParenExpr, ctx *evalContext) {
	emitExpr(e.X, ctx)
}

// 1 value
func emitBasicLit(e *ast.BasicLit, ctx *evalContext) {
	switch e.Kind.String() {
	case "CHAR":
		var val = e.Value
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
	case "INT":
		ival := strconv.Atoi(e.Value)
		fmt.Printf("  pushq $%d # number literal\n", ival)
	case "STRING":
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
}

// 1 value
func emitUnaryExpr(e *ast.UnaryExpr, ctx *evalContext) {
	switch e.Op.String() {
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
		throw(e.Op)
	}
}

// 1 value
func emitBinaryExpr(e *ast.BinaryExpr, ctx *evalContext) {
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
		panic(e.Op.String())
	}
}

// 1 value
func emitCompositeLit(e *ast.CompositeLit, ctx *evalContext) {
	// slice , array, map or struct
	ut := getUnderlyingType(getTypeOfExpr(e))
	switch kind(ut) {
	case T_STRUCT:
		emitStructLiteral(e)
	case T_ARRAY:
		arrayType := ut.E.(*ast.ArrayType)
		arrayLen := evalInt(arrayType.Len)
		emitArrayLiteral(arrayType, arrayLen, e.Elts)
	case T_SLICE:
		arrayType := ut.E.(*ast.ArrayType)
		length := len(e.Elts)
		emitArrayLiteral(arrayType, length, e.Elts)
		emitPopAddress("malloc")
		fmt.Printf("  pushq $%d # slice.cap\n", length)
		fmt.Printf("  pushq $%d # slice.len\n", length)
		fmt.Printf("  pushq %%rax # slice.ptr\n")
	default:
		unexpectedKind(kind(e2t(e.Type)))
	}
}

// 1 value list[low:high]
func emitSliceExpr(e *ast.SliceExpr, ctx *evalContext) {
	list := e.X
	listType := getTypeOfExpr(list)

	// For convenience, any of the indices may be omitted.
	// A missing low index defaults to zero;
	var low ast.Expr
	if e.Low != nil {
		low = e.Low
	} else {
		low = eZeroInt
	}

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
			// high = len(orig)
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

	emitExpr(low, nil) // index number
	elmType := getElementTypeOfListType(listType)
	emitListElementAddr(list, elmType)
}

// 1 or 2 values
func emitTypeAssertExpr(e *ast.TypeAssertExpr, ctx *evalContext) {
	emitExpr(e.X, nil)
	fmt.Printf("  popq  %%rax # ifc.dtype\n")
	fmt.Printf("  popq  %%rcx # ifc.data\n")
	fmt.Printf("  pushq %%rax # ifc.data\n")
	typ := e2t(e.Type)
	sType := serializeType(typ)
	tid := getTypeId(sType)
	typeSymbol := typeIdToSymbol(tid)
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
			emitExpr(e.X, nil)
			fmt.Printf("  popq %%rax # garbage\n")
			emitLoadAndPush(e2t(e.Type)) // load dynamic data
		}
		if ctx.okContext.needOk {
			fmt.Printf("  pushq $1 # ok = true\n")
		}
	} else {
		// default context is single value context
		emitComment(2, " single value context\n")
		emitExpr(e.X, nil)
		fmt.Printf("  popq %%rax # garbage\n")
		emitLoadAndPush(e2t(e.Type)) // load dynamic data
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
	emitComment(2, "[emitExpr] dtype=%T\n", expr)
	switch e := expr.(type) {
	case *ast.Ident:
		return emitIdent(e, ctx) // 1 value
	case *ast.IndexExpr:
		emitIndexExpr(e, ctx) // 1 or 2 values
	case *ast.StarExpr:
		emitStarExpr(e, ctx) // 1 value
	case *ast.SelectorExpr:
		emitSelectorExpr(e, ctx) // 1 value X.Sel
	case *ast.CallExpr:
		emitCallExpr(e, ctx) // multi values Fun(Args)
	case *ast.ParenExpr:
		emitParenExpr(e, ctx) // multi values (e)
	case *ast.BasicLit:
		emitBasicLit(e, ctx) // 1 value
	case *ast.UnaryExpr:
		emitUnaryExpr(e, ctx) // 1 value
	case *ast.BinaryExpr:
		emitBinaryExpr(e, ctx) // 1 value
	case *ast.CompositeLit:
		emitCompositeLit(e, ctx) // 1 value
	case *ast.SliceExpr:
		emitSliceExpr(e, ctx) // 1 value list[low:high]
	case *ast.TypeAssertExpr:
		emitTypeAssertExpr(e, ctx) // 1 or 2 values
	default:
		throw(expr)
	}
	return false
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

func emitExprIfc(expr ast.Expr, ctx *evalContext) {
	isNilObj := emitExpr(expr, ctx)
	if !isNilObj && ctx != nil && ctx._type != nil && isInterface(ctx._type) && !isInterface(getTypeOfExpr(expr)) {
		emitConvertToInterface(getTypeOfExpr(expr))
	}
}

var typeId int = 1

func typeIdToSymbol(id int) string {
	return "dtype." + strconv.Itoa(id)
}

var typesMap mapStringInt

func getTypeId(serialized string) int {
	pTypeMap := &typesMap
	id , ok := pTypeMap.get(serialized)
	if ok {
		return id
	}
	pTypeMap.set(serialized, typeId)
	r := typeId
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

func newNumberLiteral(x int) *ast.BasicLit {
	e := &ast.BasicLit{
		Kind:  token.INT,
		Value: strconv.Itoa(x),
	}
	return e
}

func emitListElementAddr(list ast.Expr, elmType *Type) {
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
		},
		&Arg{
			e:         right,
			paramType: tString,
		},
	}
	resultList := &ast.FieldList{
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
	resultList := &ast.FieldList{
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
		ff := lookupForeignFunc(newQI("runtime", "memcopy"))
		emitCallFF(ff)
	default:
		unexpectedKind(k)
	}
}

func isBlankIdentifier(e ast.Expr) bool {
	ident, isIdent := e.(*ast.Ident)
	if !isIdent {
		return false
	}
	return ident.Name == "_"
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

func emitAssignToVar(vr *Variable, rhs ast.Expr) {
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
	emitComment(2, "Assignment: emitAddr(lhs)\n")
	emitAddr(lhs)
	emitComment(2, "Assignment: emitExpr(rhs)\n")
	ctx := &evalContext{
		_type: getTypeOfExpr(lhs),
	}
	emitExprIfc(rhs, ctx)
	emitStore(getTypeOfExpr(lhs), true, false)
}

func emitBlockStmt(s *ast.BlockStmt) {
	for _, s := range s.List {
		emitStmt(s)
	}
}
func emitExprStmt(s *ast.ExprStmt) {
	emitExpr(s.X, nil)
}
func emitDeclStmt(s *ast.DeclStmt) {
	genDecl := s.Decl.(*ast.GenDecl)
	declSpec := genDecl.Specs[0]
	switch spec := declSpec.(type) {
	case *ast.ValueSpec:
		valSpec := spec
		t := e2t(valSpec.Type)
		lhs := valSpec.Names[0]
		if len(valSpec.Values) == 0 {
			emitComment(2, "lhs addresss\n")
			emitAddr(lhs)
			emitComment(2, "emitZeroValue\n")
			emitZeroValue(t)
			emitComment(2, "Assignment: zero value\n")
			emitStore(t, true, false)
		} else if len(valSpec.Values) == 1 {
			// assignment
			rhs := valSpec.Values[0]
			emitAssign(lhs, rhs)
		} else {
			panic("TBI")
		}
	default:
		throw(declSpec)
	}
}
func emitAssignStmt(s *ast.AssignStmt) {
	switch s.Tok.String() {
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
				ident, isIdent := lhs0.(*ast.Ident)
				if isIdent && ident.Name == "_" {
					panic(" _ is not supported yet")
				}
				emitAssign(lhs0, rhs0)
			} else if len(s.Lhs) >= 1 && len(s.Rhs) == 1 {
				// multi-values expr
				// a, b, c = f()
				emitExpr(rhs0, nil) // @TODO interface conversion
				callExpr := rhs0.(*ast.CallExpr)
				returnTypes := getCallResultTypes(callExpr)
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
	case "+=":
		binaryExpr := &ast.BinaryExpr{
			X:  s.Lhs[0],
			Op: token.ADD,
			Y:  s.Rhs[0],
		}
		emitAssign(s.Lhs[0], binaryExpr)
	case "-=":
		binaryExpr := &ast.BinaryExpr{
			X:  s.Lhs[0],
			Op: token.SUB,
			Y:  s.Rhs[0],
		}
		emitAssign(s.Lhs[0], binaryExpr)
	default:
		panic("TBI: assignment of " + s.Tok.String())
	}
}
func emitIfStmt(s *ast.IfStmt) {
	emitComment(2, "if\n")

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
	emitComment(2, "end if\n")
}

func emitForStmt(s *ast.ForStmt) {
	meta := s.Meta
	labelid++
	labelCond := fmt.Sprintf(".L.for.cond.%d", labelid)
	labelPost := fmt.Sprintf(".L.for.post.%d", labelid)
	labelExit := fmt.Sprintf(".L.for.exit.%d", labelid)

	meta.LabelPost = labelPost
	meta.LabelExit = labelExit

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
}

// only for array and slice for now
func emitRangeStmt(s *ast.RangeStmt) {
	meta := s.Meta
	labelid++
	labelCond := fmt.Sprintf(".L.range.cond.%d", labelid)
	labelPost := fmt.Sprintf(".L.range.post.%d", labelid)
	labelExit := fmt.Sprintf(".L.range.exit.%d", labelid)

	meta.LabelPost = labelPost
	meta.LabelExit = labelExit
	// initialization: store len(rangeexpr)
	emitComment(2, "ForRange Initialization\n")
	emitComment(2, "  assign length to lenvar\n")
	// lenvar = len(s.X)
	emitVariableAddr(meta.RngLenvar)
	emitLen(s.X)
	emitStore(tInt, true, false)

	emitComment(2, "  assign 0 to indexvar\n")
	// indexvar = 0
	emitVariableAddr(meta.RngIndexvar)
	emitZeroValue(tInt)
	emitStore(tInt, true, false)

	// init key variable with 0
	if s.Key != nil {
		keyIdent := s.Key.(*ast.Ident)
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

	emitVariableAddr(meta.RngIndexvar)
	emitLoadAndPush(tInt)
	emitVariableAddr(meta.RngLenvar)
	emitLoadAndPush(tInt)
	emitCompExpr("setl")
	emitPopBool(" indexvar < lenvar")
	fmt.Printf("  cmpq $1, %%rax\n")
	fmt.Printf("  jne %s # jmp if false\n", labelExit)

	emitComment(2, "assign list[indexvar] value variables\n")
	elemType := getTypeOfExpr(s.Value)
	emitAddr(s.Value) // lhs

	emitVariableAddr(meta.RngIndexvar)
	emitLoadAndPush(tInt) // index value
	emitListElementAddr(s.X, elemType)

	emitLoadAndPush(elemType)
	emitStore(elemType, true, false)

	// Body
	emitComment(2, "ForRange Body\n")
	emitStmt(s.Body)

	// Post statement: Increment indexvar and go next
	emitComment(2, "ForRange Post statement\n")
	fmt.Printf("  %s:\n", labelPost)   // used for "continue"
	emitVariableAddr(meta.RngIndexvar) // lhs
	emitVariableAddr(meta.RngIndexvar) // rhs
	emitLoadAndPush(tInt)
	emitAddConst(1, "indexvar value ++")
	emitStore(tInt, true, false)

	// incr key variable
	if s.Key != nil {
		keyIdent := s.Key.(*ast.Ident)
		if keyIdent.Name != "_" {
			emitAddr(s.Key)                    // lhs
			emitVariableAddr(meta.RngIndexvar) // rhs
			emitLoadAndPush(tInt)
			emitStore(tInt, true, false)
		}
	}

	fmt.Printf("  jmp %s\n", labelCond)

	fmt.Printf("  %s:\n", labelExit)
}
func emitIncDecStmt(s *ast.IncDecStmt) {
	var addValue int
	switch s.Tok.String() {
	case "++":
		addValue = 1
	case "--":
		addValue = -1
	default:
		panic("Unexpected Tok=" + s.Tok.String())
	}
	emitAddr(s.X)
	emitExpr(s.X, nil)
	emitAddConst(addValue, "rhs ++ or --")
	emitStore(getTypeOfExpr(s.X), true, false)
}
func emitSwitchStmt(s *ast.SwitchStmt) {
	labelid++
	labelEnd := fmt.Sprintf(".L.switch.%d.exit", labelid)
	if s.Init != nil {
		panic("TBI")
	}
	if s.Tag == nil {
		panic("Omitted tag is not supported yet")
	}
	emitExpr(s.Tag, nil)
	condType := getTypeOfExpr(s.Tag)
	cases := s.Body.List
	var labels = make([]string, len(cases), len(cases))
	var defaultLabel string
	emitComment(2, "Start comparison with cases\n")
	for i, c := range cases {
		cc := c.(*ast.CaseClause)
		labelid++
		labelCase := fmt.Sprintf(".L.case.%d", labelid)
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
		cc := c.(*ast.CaseClause)
		fmt.Printf("%s:\n", labels[i])
		for _, _s := range cc.Body {
			emitStmt(_s)
		}
		fmt.Printf("  jmp %s\n", labelEnd)
	}
	fmt.Printf("%s:\n", labelEnd)
}
func emitTypeSwitchStmt(s *ast.TypeSwitchStmt) {
	meta := s.Node
	labelid++
	labelEnd := fmt.Sprintf(".L.typeswitch.%d.exit", labelid)

	// subjectVariable = subject
	emitVariableAddr(meta.SubjectVariable)
	emitExpr(meta.Subject, nil)
	emitStore(tEface, true, false)

	cases := s.Body.List
	var labels = make([]string, len(cases), len(cases))
	var defaultLabel string
	emitComment(2, "Start comparison with cases\n")
	for i, c := range cases {
		cc := c.(*ast.CaseClause)
		labelid++
		labelCase := ".L.case." + strconv.Itoa(labelid)
		labels[i] = labelCase
		if len(cc.List) == 0 { // @TODO implement slice nil comparison
			defaultLabel = labelCase
			continue
		}
		for _, e := range cc.List {
			emitVariableAddr(meta.SubjectVariable)
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

	for i, typeSwitchCaseClose := range meta.Cases {
		// Injecting variable and type to the subject
		if typeSwitchCaseClose.Variable != nil {
			setVariable(meta.AssignIdent.Obj, typeSwitchCaseClose.Variable)
		}
		fmt.Printf("%s:\n", labels[i])

		for _, _s := range typeSwitchCaseClose.Orig.Body {
			if typeSwitchCaseClose.Variable != nil {
				// do assignment
				emitAddr(meta.AssignIdent)
				emitVariableAddr(meta.SubjectVariable)
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
}
func emitBranchStmt(s *ast.BranchStmt) {
	containerFor := s.CurrentFor
	switch s.Tok.String() {
	case "continue":
		fmt.Printf("jmp %s # continue\n", containerFor.LabelPost)
	case "break":
		fmt.Printf("jmp %s # break\n", containerFor.LabelExit)
	default:
		throw(s.Tok)
	}
}

func emitStmt(stmt ast.Stmt) {
	emitComment(2, "== Statement %T ==\n", stmt)
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		emitBlockStmt(s)
	case *ast.ExprStmt:
		emitExprStmt(s)
	case *ast.DeclStmt:
		emitDeclStmt(s)
	case *ast.AssignStmt:
		emitAssignStmt(s)
	case *ast.ReturnStmt:
		emitReturnStmt(s)
	case *ast.IfStmt:
		emitIfStmt(s)
	case *ast.ForStmt:
		emitForStmt(s)
	case *ast.RangeStmt:
		emitRangeStmt(s) // only for array and slice
	case *ast.IncDecStmt:
		emitIncDecStmt(s)
	case *ast.SwitchStmt:
		emitSwitchStmt(s)
	case *ast.TypeSwitchStmt:
		emitTypeSwitchStmt(s)
	case *ast.BranchStmt:
		emitBranchStmt(s)
	default:
		throw(stmt)
	}
}

func emitRevertStackTop(t *Type) {
	fmt.Printf("  addq $%d, %%rsp # revert stack top\n", getSizeOfType(t))
}

var labelid int

func getMethodSymbol(method *Method) string {
	rcvTypeName := method.RcvNamedType
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

func emitFuncDecl(pkgName string, fnc *Func) {
	fmt.Printf("# emitFuncDecl\n")
	if len(fnc.Params) > 0 {
		for i := 0; i < len(fnc.Params); i++ {
			v := fnc.Params[i]
			logf("  #       params %d %d \"%s\" %s\n", v.LocalOffset, getSizeOfType(v.Typ), v.Name, string(kind(v.Typ)))
		}
	}
	if len(fnc.Retvars) > 0 {
		for i := 0; i < len(fnc.Retvars); i++ {
			v := fnc.Retvars[i]
			logf("  #       retvars %d %d \"%s\" %s\n", v.LocalOffset, getSizeOfType(v.Typ), v.Name, string(kind(v.Typ)))
		}
	}

	var symbol string
	if fnc.Method != nil {
		symbol = getMethodSymbol(fnc.Method)
	} else {
		symbol = getPackageSymbol(pkgName, fnc.Name)
	}
	fmt.Printf("%s: # args %d, locals %d\n", symbol, fnc.Argsarea, fnc.Localarea)
	fmt.Printf("  pushq %%rbp\n")
	fmt.Printf("  movq %%rsp, %%rbp\n")
	if len(fnc.LocalVars) > 0 {
		for i := len(fnc.LocalVars) - 1; i >= 0; i-- {
			v := fnc.LocalVars[i]
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

func emitGlobalVariableComplex(name *ast.Ident, t *Type, val ast.Expr) {
	typeKind := kind(t)
	switch typeKind {
	case T_POINTER:
		fmt.Printf("# init global %s:\n", name.Name)
		emitAssign(name, val)
	}
}

func emitGlobalVariable(pkg *PkgContainer, name *ast.Ident, t *Type, val ast.Expr) {
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
				throw(val)
			}
		default:
			throw(val)
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
			throw(val)
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
			throw(val)
		}
	case T_UINT16:
		if val == nil {
			fmt.Printf("  .word 0\n")
			return
		}
		switch vl := val.(type) {
		case *ast.BasicLit:
			fmt.Printf("  .word %s\n", vl.Value)
		default:
			throw(val)
		}
	case T_POINTER:
		// will be set in the initGlobal func
		fmt.Printf("  .quad 0\n")
	case T_UINTPTR:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		fmt.Printf("  .quad 0\n")
	case T_SLICE:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		fmt.Printf("  .quad 0 # ptr\n")
		fmt.Printf("  .quad 0 # len\n")
		fmt.Printf("  .quad 0 # cap\n")
	case T_ARRAY:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		arrayType := t.E.(*ast.ArrayType)
		assert(arrayType.Len != nil, "slice type is not expected", __func__)
		length := evalInt(arrayType.Len)
		var zeroValue string
		knd := kind(e2t(arrayType.Elt))
		switch knd {
		case T_INT:
			zeroValue = "  .quad 0 # int zero value\n"
		case T_UINT8:
			zeroValue = "  .byte 0 # uint8 zero value\n"
		case T_STRING:
			zeroValue = "  .quad 0 # string zero value (ptr)\n"
			zeroValue += "  .quad 0 # string zero value (len)\n"
		case T_INTERFACE:
			zeroValue = "  .quad 0 # eface zero value (dtype)\n"
			zeroValue += "  .quad 0 # eface zero value (data)\n"
		default:
			unexpectedKind(knd)
		}

		for i := 0; i < length; i++ {
			fmt.Printf(zeroValue)
		}
	default:
		unexpectedKind(typeKind)
	}
}

func generateCode(pkg *PkgContainer) {
	fmt.Printf("#===================== generateCode %s =====================\n", pkg.name)
	fmt.Printf(".data\n")
	for _, con := range pkg.stringLiterals {
		emitComment(0, "string literals\n")
		fmt.Printf("%s:\n", con.sl.label)
		fmt.Printf("  .string %s\n", con.sl.value)
	}

	for _, spec := range pkg.vars {
		var val ast.Expr
		if len(spec.Values) > 0 {
			val = spec.Values[0]
		}
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}

		emitGlobalVariable(pkg, spec.Names[0], t, val)

	}

	fmt.Printf("\n")
	fmt.Printf(".text\n")
	fmt.Printf("%s.__initGlobals:\n", pkg.name)
	for _, spec := range pkg.vars {
		if len(spec.Values) == 0 {
			continue
		}
		val := spec.Values[0]
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariableComplex(spec.Names[0], t, val)
	}
	fmt.Printf("  ret\n")

	for _, fnc := range pkg.funcs {
		emitFuncDecl(pkg.name, fnc)
	}

	fmt.Printf("\n")
}

func emitDynamicTypes(typeMap []*mapStringIntEntry) {
	// emitting dynamic types
	fmt.Printf("# ------- Dynamic Types ------\n")
	fmt.Printf(".data\n")
	for _, te := range typeMap {
		id := te.value
		name := te.key
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

// types of an expr in single value context
func getTypeOfExpr(expr ast.Expr) *Type {
	switch e := expr.(type) {
	case *ast.Ident:
		assert(e.Obj != nil, "Obj is nil in ident '"+e.Name+"'", __func__)
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
			switch dcl := e.Obj.Decl.(type) {
			case *ast.ValueSpec:
				return e2t(dcl.Type)
			case *ast.Field:
				return e2t(dcl.Type)
			case *ast.AssignStmt: // var lhs = rhs | lhs := rhs
				return getTypeOfExpr(dcl.Rhs[0])
			default:
				panic("Unknown type")
			}
		case ast.Con:
			switch e.Obj {
			case gTrue, gFalse:
				return tBool
			default:
				switch decl2 := e.Obj.Decl.(type) {
				case *ast.ValueSpec:
					return e2t(decl2.Type)
				default:
					panic("cannot decide type of cont =" + e.Obj.Name)
				}
			}
		default:
			panic("Obj=" + e.Obj.Name + e.Obj.Kind)
		}
	case *ast.BasicLit:
		// The default type of an untyped constant is bool, rune, int, float64, complex128 or string respectively,
		// depending on whether it is a boolean, rune, integer, floating-point, complex, or string constant.
		switch e.Kind.String() {
		case "STRING":
			return tString
		case "INT":
			return tInt
		case "CHAR":
			return tInt32
		default:
			panic(e.Kind.String())
		}
	case *ast.UnaryExpr:
		switch e.Op.String() {
		case "+":
			return getTypeOfExpr(e.X)
		case "-":
			return getTypeOfExpr(e.X)
		case "!":
			return tBool
		case "&":
			t := getTypeOfExpr(e.X)
			starExpr := &ast.StarExpr{
				X: t.E,
			}
			return e2t(starExpr)
		case "range":
			listType := getTypeOfExpr(e.X)
			elmType := getElementTypeOfListType(listType)
			return elmType
		default:
			panic(e.Op.String())
		}
	case *ast.BinaryExpr:
		switch e.Op.String() {
		case "==", "!=", "<", ">", "<=", ">=":
			return tBool
		default:
			return getTypeOfExpr(e.X)
		}
	case *ast.IndexExpr:
		list := e.X
		return getElementTypeOfListType(getTypeOfExpr(list))
	case *ast.CallExpr: // funcall or conversion
		types := getCallResultTypes(e)
		assert(len(types) == 1, "single value is expected", __func__)
		return types[0]
	case *ast.SliceExpr:
		underlyingCollectionType := getTypeOfExpr(e.X)
		if kind(underlyingCollectionType) == T_STRING {
			// str2 = str1[n:m]
			return tString
		}
		var elementTyp ast.Expr
		switch colType := underlyingCollectionType.E.(type) {
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
		ptrType := t.E.(*ast.StarExpr)
		return e2t(ptrType.X)
	case *ast.SelectorExpr:
		if isQI(e) { // pkg.SomeType
			ident := lookupForeignIdent(selector2QI(e))
			return getTypeOfExpr(ident)
		} else { // (e).field
			ut := getUnderlyingType(getTypeOfExpr(e.X))
			var structTypeLiteral *ast.StructType
			switch typ := ut.E.(type) {
			case *ast.StructType: // strct.field
				structTypeLiteral = typ
			case *ast.StarExpr: // ptr.field
				structType := e2t(typ.X)
				structTypeLiteral = getUnderlyingStructType(structType)
			}
			field := lookupStructField(structTypeLiteral, e.Sel.Name)
			return e2t(field.Type)
		}
	case *ast.CompositeLit:
		return e2t(e.Type)
	case *ast.ParenExpr:
		return getTypeOfExpr(e.X)
	case *ast.TypeAssertExpr:
		return e2t(e.Type)
	default:
		panic(expr)
	}
	panic("nil type is not allowed\n")
}

func fieldList2Types(fldlist *ast.FieldList) []*Type {
	var r []*Type
	for _, e2 := range fldlist.List {
		t := e2t(e2.Type)
		r = append(r, t)
	}
	return r
}

func getCallResultTypes(e *ast.CallExpr) []*Type {
	switch fn := e.Fun.(type) {
	case *ast.Ident:
		if fn.Obj == nil {
			throw(fn)
		}
		switch fn.Obj.Kind {
		case ast.Typ: // conversion
			return []*Type{e2t(fn)}
		case ast.Fun:
			switch fn.Obj {
			case gLen, gCap:
				return []*Type{tInt}
			case gNew:
				starExpr := &ast.StarExpr{
					X: e.Args[0],
				}
				return []*Type{e2t(starExpr)}
			case gMake:
				return []*Type{e2t(e.Args[0])}
			case gAppend:
				return []*Type{e2t(e.Args[0])}
			}
			var decl = fn.Obj.Decl
			if decl == nil {
				panic("decl of function " + fn.Name + " is  nil")
			}
			switch dcl := decl.(type) {
			case *ast.FuncDecl:
				return fieldList2Types(dcl.Type.Results)
			default:
				panic("[astCallExpr] unknown dtype")
			}
			panic("[astCallExpr] Fun ident " + fn.Name)
		}
	case *ast.ParenExpr: // (X)(e) funcall or conversion
		if isType(fn.X) {
			return []*Type{e2t(fn.X)}
		} else {
			panic("TBI: what should we do ?")
		}
	case *ast.ArrayType: // conversion [n]T(e) or []T(e)
		return []*Type{e2t(fn)}
	case *ast.SelectorExpr:
		if isType(fn) {
			return []*Type{e2t(fn)}
		}
		if isQI(fn) { // pkg.Sel()
			ff := lookupForeignFunc(selector2QI(fn))
			return fieldList2Types(ff.decl.Type.Results)
		} else { // obj.method()
			rcvType := getTypeOfExpr(fn.X)
			method := lookupMethod(rcvType, fn.Sel)
			return fieldList2Types(method.FuncType.Results)
		}
	case *ast.InterfaceType:
		return []*Type{tEface}
	}

	throw(e)
	return nil
}

func e2t(typeExpr ast.Expr) *Type {
	if typeExpr == nil {
		panic("nil is not allowed")
	}
	return &Type{
		E: typeExpr,
	}
}

func serializeType(t *Type) string {
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
			throw(e.Obj)
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
				typeSpec := decl.(*ast.TypeSpec)
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
		throw(t)
	}
	return ""
}

func getUnderlyingStructType(t *Type) *ast.StructType {
	ut := getUnderlyingType(t)
	return ut.E.(*ast.StructType)
}

func getUnderlyingType(t *Type) *Type {
	if t == nil {
		panic("nil type is not expected")
	}
	if t.E == generalSlice {
		return t
	}

	switch e := t.E.(type) {
	case *ast.StructType, *ast.ArrayType, *ast.StarExpr, *ast.Ellipsis, *ast.InterfaceType:
		// type literal
		return t
	case *ast.Ident:
		assert(e.Obj.Kind == ast.Typ, "should be ast.Typ", __func__)
		if isPredeclaredType(e.Obj) {
			return t
		}
		// defined type or alias
		typeSpec := e.Obj.Decl.(*ast.TypeSpec)
		// get RHS in its type definition recursively
		return getUnderlyingType(e2t(typeSpec.Type))
	case *ast.SelectorExpr:
		ident := lookupForeignIdent(selector2QI(e))
		return getUnderlyingType(e2t(ident))
	case *ast.ParenExpr:
		return getUnderlyingType(e2t(e.X))
	}
	panic("should not reach here")
}

func kind(t *Type) TypeKind {
	if t == nil {
		panic("nil type is not expected")
	}

	ut := getUnderlyingType(t)
	if ut.E == generalSlice {
		return T_SLICE
	}

	switch e := ut.E.(type) {
	case *ast.Ident:
		assert(e.Obj.Kind == ast.Typ, "should be ast.Typ", __func__)
		switch e.Obj {
		case gUintptr:
			return T_UINTPTR
		case gInt:
			return T_INT
		case gInt32:
			return T_INT32
		case gString:
			return T_STRING
		case gUint8:
			return T_UINT8
		case gUint16:
			return T_UINT16
		case gBool:
			return T_BOOL
		default:
			panic("Unexpected type")
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
	}
	panic("should not reach here")
}

func isInterface(t *Type) bool {
	return kind(t) == T_INTERFACE
}

func getElementTypeOfListType(t *Type) *Type {
	ut := getUnderlyingType(t)
	switch kind(t) {
	case T_SLICE, T_ARRAY:
		switch e := ut.E.(type) {
		case *ast.ArrayType:
			return e2t(e.Elt)
		case *ast.Ellipsis:
			return e2t(e.Elt)
		default:
			throw(t.E)
		}
	case T_STRING:
		return tUint8
	default:
		unexpectedKind(kind(t))
	}
	return nil
}

const SizeOfSlice int = 24
const SizeOfString int = 16
const SizeOfInt int = 8
const SizeOfUint8 int = 1
const SizeOfUint16 int = 2
const SizeOfPtr int = 8
const SizeOfInterface int = 16

func getSizeOfType(t *Type) int {
	ut := getUnderlyingType(t)
	switch kind(ut) {
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
		arrayType := ut.E.(*ast.ArrayType)
		elmSize := getSizeOfType(e2t(arrayType.Elt))
		return elmSize * evalInt(arrayType.Len)
	case T_STRUCT:
		return calcStructSizeAndSetFieldOffset(ut.E.(*ast.StructType))
	default:
		unexpectedKind(kind(t))
	}
	return 0
}

func getStructFieldOffset(field *ast.Field) int {
	offset := field.Offset
	return offset
}

func setStructFieldOffset(field *ast.Field, offset int) {
	field.Offset = offset
}

func lookupStructField(structType *ast.StructType, selName string) *ast.Field {
	for _, field := range structType.Fields.List {
		if field.Name.Name == selName {
			return field
		}
	}
	panic("Unexpected flow: struct field not found:" + selName)
}

func calcStructSizeAndSetFieldOffset(structType *ast.StructType) int {
	var offset int = 0
	for _, field := range structType.Fields.List {
		setStructFieldOffset(field, offset)
		size := getSizeOfType(e2t(field.Type))
		offset += size
	}
	return offset
}

// --- walk ---
type sliteral struct {
	label  string
	strlen int
	value  string // raw value
}

type stringLiteralsContainer struct {
	lit *ast.BasicLit
	sl  *sliteral
}

func registerParamVariable(fnc *Func, name string, t *Type) *Variable {
	vr := newLocalVariable(name, fnc.Argsarea, t)
	size := getSizeOfType(t)
	fnc.Argsarea += size
	fnc.Params = append(fnc.Params, vr)
	return vr
}

func registerReturnVariable(fnc *Func, name string, t *Type) *Variable {
	vr := newLocalVariable(name, fnc.Argsarea, t)
	size := getSizeOfType(t)
	fnc.Argsarea += size
	fnc.Retvars = append(fnc.Retvars, vr)
	return vr
}

func registerLocalVariable(fnc *Func, name string, t *Type) *Variable {
	assert(t != nil && t.E != nil, "type of local var should not be nil", __func__)
	fnc.Localarea -= getSizeOfType(t)
	vr := newLocalVariable(name, currentFunc.Localarea, t)
	fnc.LocalVars = append(fnc.LocalVars, vr)
	return vr
}

var currentFunc *Func

func getStringLiteral(lit *ast.BasicLit) *sliteral {
	for _, container := range currentPkg.stringLiterals {
		if container.lit == lit {
			return container.sl
		}
	}

	panic("string literal not found:" + lit.Value)
}

func registerStringLiteral(lit *ast.BasicLit) {
	if currentPkg.name == "" {
		panic("no pkgName")
	}

	var strlen int
	for _, c := range []uint8(lit.Value) {
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
	cont := &stringLiteralsContainer{
		sl:  sl,
		lit: lit,
	}
	currentPkg.stringLiterals = append(currentPkg.stringLiterals, cont)
}

func newGlobalVariable(pkgName string, name string, t *Type) *Variable {
	return &Variable{
		Name:         name,
		IsGlobal:     true,
		GlobalSymbol: pkgName + "." + name,
		Typ:          t,
	}
}

func newLocalVariable(name string, localoffset int, t *Type) *Variable {
	return &Variable{
		Name:        name,
		IsGlobal:    false,
		LocalOffset: localoffset,
		Typ:         t,
	}
}

type methodEntry struct {
	name   string
	method *Method
}

type namedTypeEntry struct {
	//name    string
	obj     *ast.Object
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
	ident, isIdent := e.X.(*ast.Ident)
	if !isIdent {
		return false
	}
	return ident.Obj.Kind == ast.Pkg
}

func selector2QI(e *ast.SelectorExpr) QualifiedIdent {
	pkgName := e.X.(*ast.Ident)
	assert(pkgName.Obj.Kind == ast.Pkg, "should be ast.Pkg", __func__)
	return newQI(pkgName.Name, e.Sel.Name)
}

func newMethod(pkgName string, funcDecl *ast.FuncDecl) *Method {
	rcvType := funcDecl.Recv.List[0].Type
	rcvPointerType, isPtr := rcvType.(*ast.StarExpr)
	if isPtr {
		rcvType = rcvPointerType.X
	}
	rcvNamedType := rcvType.(*ast.Ident)
	var method = &Method{
		PkgName:      pkgName,
		RcvNamedType: rcvNamedType,
		IsPtrMethod:  isPtr,
		Name:         funcDecl.Name.Name,
		FuncType:     funcDecl.Type,
	}
	return method
}

func registerMethod(method *Method) {
	var nt = findNamedType(method.RcvNamedType.Obj)
	if nt == nil {
		nt = &namedTypeEntry{
			obj:     method.RcvNamedType.Obj,
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

func lookupMethod(rcvT *Type, methodName *ast.Ident) *Method {
	rcvType := rcvT.E
	rcvPointerType, isPtr := rcvType.(*ast.StarExpr)
	if isPtr {
		rcvType = rcvPointerType.X
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
}

func walkExprStmt(s *ast.ExprStmt) {
	walkExpr(s.X)
}
func walkDeclStmt(s *ast.DeclStmt) {
	genDecl := s.Decl.(*ast.GenDecl)
	declSpec := genDecl.Specs[0]
	switch spec := declSpec.(type) {
	case *ast.ValueSpec:
		if spec.Type == nil { // var x = e
			if len(spec.Values) == 0 {
				panic("invalid syntax")
			}
			// infer type from rhs
			val := spec.Values[0]
			logf("inferring type of variable %s\n", spec.Names[0].Name)
			typ := getTypeOfExpr(val)
			if typ == nil || typ.E == nil {
				panic("rhs should have a type")
			}
			spec.Type = typ.E
		}
		t := e2t(spec.Type)
		obj := spec.Names[0].Obj
		setVariable(obj, registerLocalVariable(currentFunc, obj.Name, t))
		if len(spec.Values) > 0 {
			walkExpr(spec.Values[0])
		}
	}
}
func walkAssignStmt(s *ast.AssignStmt) {
	if s.Tok.String() == ":=" {
		rhs0 := s.Rhs[0]
		walkExpr(rhs0)
		var isMultiValuedContext bool
		if len(s.Lhs) > 1 && len(s.Rhs) == 1 {
			isMultiValuedContext = true
		}

		if isMultiValuedContext {
			// a, b, c := rhs0
			// infer type
			var types []*Type
			switch rhs := rhs0.(type) {
			case *ast.CallExpr:
				types = getCallResultTypes(rhs)
			case *ast.TypeAssertExpr:
				typ0 := getTypeOfExpr(rhs0)
				types = []*Type{typ0, tBool}
			default:
				panic("TBI")
			}
			for i, lhs := range s.Lhs {
				obj := lhs.(*ast.Ident).Obj
				setVariable(obj, registerLocalVariable(currentFunc, obj.Name, types[i]))
			}
		} else {
			for i, lhs := range s.Lhs {
				obj := lhs.(*ast.Ident).Obj
				rhs := s.Rhs[i]
				walkExpr(rhs)
				rhsType := getTypeOfExpr(s.Rhs[i])
				setVariable(obj, registerLocalVariable(currentFunc, obj.Name, rhsType))
			}
		}
	} else {
		for _, rhs := range s.Rhs {
			walkExpr(rhs)
		}
	}
}
func walkReturnStmt(s *ast.ReturnStmt) {
	s.Meta = &ast.MetaReturnStmt{
		Fnc: currentFunc,
	}
	for _, rt := range s.Results {
		walkExpr(rt)
	}
}
func walkIfStmt(s *ast.IfStmt) {
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
}
func walkForStmt(s *ast.ForStmt) {
	meta := &ast.MetaForStmt{
		Outer: currentFor,
	}
	currentFor = meta
	s.Meta = meta
	if s.Init != nil {
		walkStmt(s.Init)
	}
	if s.Cond != nil {
		walkExpr(s.Cond)
	}
	if s.Post != nil {
		walkStmt(s.Post)
	}
	walkStmt(s.Body)
	currentFor = meta.Outer
}
func walkRangeStmt(s *ast.RangeStmt) {
	meta := &ast.MetaForStmt{
		Outer: currentFor,
	}
	currentFor = meta
	s.Meta = meta
	walkExpr(s.X)
	walkStmt(s.Body)
	meta.RngLenvar = registerLocalVariable(currentFunc, ".range.len", tInt)
	meta.RngIndexvar = registerLocalVariable(currentFunc, ".range.index", tInt)
	if s.Tok.String() == ":=" {
		// short var decl
		listType := getTypeOfExpr(s.X)

		keyIdent := s.Key.(*ast.Ident)
		//@TODO map key can be any type
		//keyType := getKeyTypeOfListType(listType)
		keyType := tInt
		setVariable(keyIdent.Obj, registerLocalVariable(currentFunc, keyIdent.Name, keyType))

		// determine type of Value
		elmType := getElementTypeOfListType(listType)
		valueIdent := s.Value.(*ast.Ident)
		setVariable(valueIdent.Obj, registerLocalVariable(currentFunc, valueIdent.Name, elmType))
	}
	currentFor = meta.Outer
}
func walkIncDecStmt(s *ast.IncDecStmt) {
	walkExpr(s.X)
}
func walkBlockStmt(s *ast.BlockStmt) {
	for _, _s := range s.List {
		walkStmt(_s)
	}
}
func walkBranchStmt(s *ast.BranchStmt) {
	s.CurrentFor = currentFor
}
func walkSwitchStmt(s *ast.SwitchStmt) {
	if s.Tag != nil {
		walkExpr(s.Tag)
	}
	walkStmt(s.Body)
}
func walkTypeSwitchStmt(s *ast.TypeSwitchStmt) {
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
}
func walkCaseClause(s *ast.CaseClause) {
	for _, e := range s.List {
		walkExpr(e)
	}
	for _, stmt := range s.Body {
		walkStmt(stmt)
	}
}
func walkStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		walkExprStmt(s)
	case *ast.DeclStmt:
		walkDeclStmt(s)
	case *ast.AssignStmt:
		walkAssignStmt(s)
	case *ast.ReturnStmt:
		walkReturnStmt(s)
	case *ast.IfStmt:
		walkIfStmt(s)
	case *ast.BlockStmt:
		walkBlockStmt(s)
	case *ast.ForStmt:
		walkForStmt(s)
	case *ast.RangeStmt:
		walkRangeStmt(s)
	case *ast.IncDecStmt:
		walkIncDecStmt(s)
	case *ast.SwitchStmt:
		walkSwitchStmt(s)
	case *ast.TypeSwitchStmt:
		walkTypeSwitchStmt(s)
	case *ast.CaseClause:
		walkCaseClause(s)
	case *ast.BranchStmt:
		walkBranchStmt(s)
	default:
		throw(stmt)
	}
}

var currentFor *ast.MetaForStmt

func walkIdent(e *ast.Ident) {
}
func walkCallExpr(e *ast.CallExpr) {
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
}
func walkBasicLit(e *ast.BasicLit) {
	switch e.Kind.String() {
	case "INT":
	case "CHAR":
	case "STRING":
		registerStringLiteral(e)
	default:
		panic("Unexpected literal kind:" + e.Kind.String())
	}
}
func walkCompositeLit(e *ast.CompositeLit) {
	for _, v := range e.Elts {
		walkExpr(v)
	}
}
func walkUnaryExpr(e *ast.UnaryExpr) {
	walkExpr(e.X)
}
func walkBinaryExpr(e *ast.BinaryExpr) {
	walkExpr(e.X) // left
	walkExpr(e.Y) // right
}
func walkIndexExpr(e *ast.IndexExpr) {
	walkExpr(e.Index)
	walkExpr(e.X)
}
func walkSliceExpr(e *ast.SliceExpr) {
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
}
// []T(e)
func walkArrayType(e *ast.ArrayType) {
	// first argument of builtin func like make()
	// do nothing
}
func walkStarExpr(e *ast.StarExpr) {
	walkExpr(e.X)
}
func walkSelectorExpr(e *ast.SelectorExpr) {
	walkExpr(e.X)
}

func walkParenExpr(e *ast.ParenExpr) {
	walkExpr(e.X)
}
func walkKeyValueExpr(e *ast.KeyValueExpr) {
	walkExpr(e.Key)
	walkExpr(e.Value)
}
func walkInterfaceType(e *ast.InterfaceType) {
	// interface{}(e)  conversion. Nothing to do.
}
func walkTypeAssertExpr(e *ast.TypeAssertExpr) {
	walkExpr(e.X)
}
func walkExpr(expr ast.Expr) {
	logf(" [walkExpr] dtype=%s\n", dtypeOf(expr))
	switch e := expr.(type) {
	case *ast.Ident:
		walkIdent(e)
	case *ast.SelectorExpr:
		walkSelectorExpr(e)
	case *ast.CallExpr:
		walkCallExpr(e)
	case *ast.ParenExpr:
		walkParenExpr(e)
	case *ast.BasicLit:
		walkBasicLit(e)
	case *ast.CompositeLit:
		walkCompositeLit(e)
	case *ast.UnaryExpr:
		walkUnaryExpr(e)
	case *ast.BinaryExpr:
		walkBinaryExpr(e)
	case *ast.IndexExpr:
		walkIndexExpr(e)
	case *ast.ArrayType:
		walkArrayType(e) // []T(e)
	case *ast.SliceExpr:
		walkSliceExpr(e)
	case *ast.StarExpr:
		walkStarExpr(e)
	case *ast.KeyValueExpr:
		walkKeyValueExpr(e)
	case *ast.InterfaceType:
		walkInterfaceType(e)
	case *ast.TypeAssertExpr:
		walkTypeAssertExpr(e)
	default:
		throw(expr)
	}
}

var ExportedQualifiedIdents []*exportEntry

type exportEntry struct {
	qi  QualifiedIdent
	any *ast.Ident
}

func lookupForeignIdent(qi QualifiedIdent) *ast.Ident {
	for _, entry := range ExportedQualifiedIdents {
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
	ident := lookupForeignIdent(qi)
	assert(ident.Obj.Kind == ast.Fun, "should be Fun", __func__)
	decl := ident.Obj.Decl.(*ast.FuncDecl)
	return &ForeignFunc{
		symbol: string(qi),
		decl:   decl,
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
			switch spec := dcl.Specs[0].(type) {
			case *ast.TypeSpec:
				typeSpecs = append(typeSpecs, spec)
			case *ast.ValueSpec:
				if spec.Names[0].Obj.Kind == ast.Var {
					varSpecs = append(varSpecs, spec)
				} else if spec.Names[0].Obj.Kind == ast.Con {
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
		t := e2t(typeSpec.Type)
		switch kind(t) {
		case T_STRUCT:
			structType := getUnderlyingType(t)
			calcStructSizeAndSetFieldOffset(structType.E.(*ast.StructType))
		}
		exportEntry := &exportEntry{
			qi:  newQI(pkg.name, typeSpec.Name.Name),
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
			fdcl, ok = funcDecl.Name.Obj.Decl.(*ast.FuncDecl)
			if !ok || funcDecl != fdcl {
				panic("Bad func decl reference:" + funcDecl.Name.Name)
			}
			exportEntry := &exportEntry{
				qi:  newQI(pkg.name, funcDecl.Name.Name),
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
		walkExpr(constSpec.Values[0])
	}

	for _, valSpec := range varSpecs {
		var nameIdent = valSpec.Names[0]
		assert(nameIdent.Obj.Kind == ast.Var, "should be Var", __func__)
		if valSpec.Type == nil {
			var val = valSpec.Values[0]
			var t = getTypeOfExpr(val)
			valSpec.Type = t.E
		}
		setVariable(nameIdent.Obj, newGlobalVariable(pkg.name, nameIdent.Obj.Name, e2t(valSpec.Type)))
		pkg.vars = append(pkg.vars, valSpec)
		exportEntry := &exportEntry{
			qi:  newQI(pkg.name, nameIdent.Name),
			any: nameIdent,
		}
		ExportedQualifiedIdents = append(ExportedQualifiedIdents, exportEntry)
		if len(valSpec.Values) > 0 {
			walkExpr(valSpec.Values[0])
		}
	}

	for _, funcDecl := range funcDecls {
		fnc := &Func{
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

var tInt *Type
var tInt32 *Type // Rune
var tUint8 *Type
var tUint16 *Type
var tUintptr *Type
var tString *Type
var tEface *Type
var tBool *Type
var generalSlice ast.Expr

func isPredeclaredType(obj *ast.Object) bool {
	switch obj {
	case gUintptr, gInt, gInt32, gString, gUint8, gUint16, gBool:
		return true
	}
	return false
}

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
	funcs          []*Func
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
	for _, astFile := range _pkg.astFiles {
		resolveImports(astFile)
		var unresolved []*ast.Ident
		for _, ident := range astFile.Unresolved {
			obj := pkgScope.Lookup(ident.Name)
			if obj != nil {
				ident.Obj = obj
			} else {
				logf("# unresolved: %s\n", ident.Name)
				obj := universe.Lookup(ident.Name)
				if obj != nil {
					ident.Obj = obj
				} else {
					// we should allow unresolved for now.
					// e.g foo in X{foo:bar,}
					logf("Unresolved (maybe struct field name in composite literal): " + ident.Name)
					unresolved = append(unresolved, ident)
				}
			}
		}
		for _, dcl := range astFile.Decls {
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

func initGlobals() {
	eNil = identNil
	eZeroInt = &ast.BasicLit{
		Value: "0",
		Kind:  "INT",
	}
	generalSlice = &ast.Ident{}
	tInt = &Type{
		E: &ast.Ident{
			Name: "int",
			Obj:  gInt,
		},
	}
	tInt32 = &Type{
		E: &ast.Ident{
			Name: "int32",
			Obj:  gInt32,
		},
	}
	tUint8 = &Type{
		E: &ast.Ident{
			Name: "uint8",
			Obj:  gUint8,
		},
	}

	tUint16 = &Type{
		E: &ast.Ident{
			Name: "uint16",
			Obj:  gUint16,
		},
	}
	tUintptr = &Type{
		E: &ast.Ident{
			Name: "uintptr",
			Obj:  gUintptr,
		},
	}

	tString = &Type{
		E: &ast.Ident{
			Name: "string",
			Obj:  gString,
		},
	}

	tEface = &Type{
		E: &ast.InterfaceType{},
	}

	tBool = &Type{
		E: &ast.Ident{
			Name: "bool",
			Obj:  gBool,
		},
	}
}

func main() {
	initGlobals()
	srcPath = os.Getenv("GOPATH") + "/src"
	prjSrcPath = srcPath + "/github.com/DQNEO/babygo/src"

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

	var inputFiles []string
	for _, arg := range os.Args[1:] {
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

	emitDynamicTypes(typesMap)
}

// --- util ---
func obj2var(obj *ast.Object) *Variable {
	assert(obj.Kind == ast.Var, "should be ast.Var", __func__)
	assert(obj.Variable != nil, "should not be nil", __func__)
	return obj.Variable
}

func setVariable(obj *ast.Object, vr *Variable) {
	obj.Variable = vr
}

func throw(x interface{}) {
	panic(x)
}

func panic2(caller string, x string) {
	panic(caller + ": " + x)
}

func getMetaReturnStmt(s *ast.ReturnStmt) *ast.MetaReturnStmt {
	return s.Meta
}

