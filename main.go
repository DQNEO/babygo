package main

import (
	"os"
	"unsafe"

	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/token"

	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/path"
	"github.com/DQNEO/babygo/lib/strconv"
	"github.com/DQNEO/babygo/lib/strings"

	"github.com/DQNEO/babygo/lib/fmt"
	//"fmt"
)

var ProgName string = "babygo"

var __func__ = "__func__"

func assert(bol bool, msg string, caller string) {
	if !bol {
		panic(currentPkg.name + ":" + caller + ": " + msg)
	}
}

func unexpectedKind(knd TypeKind) {
	panic("Unexpected Kind: " + string(knd))
}

var fout *os.File

func printf(format string, a ...interface{}) {
	fmt.Fprintf(fout, format, a...)
}

var debugFrontEnd bool

func logf(format string, a ...interface{}) {
	if !debugFrontEnd {
		return
	}
	f := "# " + format
	fmt.Fprintf(fout, f, a...)
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
	printf(format2, a...)
}

func evalInt(expr ast.Expr) int {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return strconv.Atoi(e.Value)
	}
	panic("Unknown type")
}

func emitPopPrimitive(comment string) {
	printf("  popq %%rax # result of %s\n", comment)
}

func emitPopBool(comment string) {
	printf("  popq %%rax # result of %s\n", comment)
}

func emitPopAddress(comment string) {
	printf("  popq %%rax # address of %s\n", comment)
}

func emitPopString() {
	printf("  popq %%rax # string.ptr\n")
	printf("  popq %%rcx # string.len\n")
}

func emitPopInterFace() {
	printf("  popq %%rax # eface.dtype\n")
	printf("  popq %%rcx # eface.data\n")
}

func emitPopSlice() {
	printf("  popq %%rax # slice.ptr\n")
	printf("  popq %%rcx # slice.len\n")
	printf("  popq %%rdx # slice.cap\n")
}

func emitPushStackTop(condType *Type, offset int, comment string) {
	switch kind(condType) {
	case T_STRING:
		printf("  movq %d+8(%%rsp), %%rcx # copy str.len from stack top (%s)\n", offset, comment)
		printf("  movq %d+0(%%rsp), %%rax # copy str.ptr from stack top (%s)\n", offset, comment)
		printf("  pushq %%rcx # str.len\n")
		printf("  pushq %%rax # str.ptr\n")
	case T_POINTER, T_UINTPTR, T_BOOL, T_INT, T_UINT8, T_UINT16:
		printf("  movq %d(%%rsp), %%rax # copy stack top value (%s) \n", offset, comment)
		printf("  pushq %%rax\n")
	default:
		unexpectedKind(kind(condType))
	}
}

func emitAllocReturnVarsArea(size int) {
	if size == 0 {
		return
	}
	printf("  subq $%d, %%rsp # alloc return vars area\n", size)
}

func emitFreeParametersArea(size int) {
	if size == 0 {
		return
	}
	printf("  addq $%d, %%rsp # free parameters area\n", size)
}

func emitAddConst(addValue int, comment string) {
	emitComment(2, "Add const: %s\n", comment)
	printf("  popq %%rax\n")
	printf("  addq $%d, %%rax\n", addValue)
	printf("  pushq %%rax\n")
}

// "Load" means copy data from memory to registers
func emitLoadAndPush(t *Type) {
	assert(t != nil, "type should not be nil", __func__)
	emitPopAddress(string(kind(t)))
	switch kind(t) {
	case T_SLICE:
		printf("  movq %d(%%rax), %%rdx\n", 16)
		printf("  movq %d(%%rax), %%rcx\n", 8)
		printf("  movq %d(%%rax), %%rax\n", 0)
		printf("  pushq %%rdx # cap\n")
		printf("  pushq %%rcx # len\n")
		printf("  pushq %%rax # ptr\n")
	case T_STRING:
		printf("  movq %d(%%rax), %%rdx # len\n", 8)
		printf("  movq %d(%%rax), %%rax # ptr\n", 0)
		printf("  pushq %%rdx # len\n")
		printf("  pushq %%rax # ptr\n")
	case T_INTERFACE:
		printf("  movq %d(%%rax), %%rdx # data\n", 8)
		printf("  movq %d(%%rax), %%rax # dtype\n", 0)
		printf("  pushq %%rdx # data\n")
		printf("  pushq %%rax # dtype\n")
	case T_UINT8:
		printf("  movzbq %d(%%rax), %%rax # load uint8\n", 0)
		printf("  pushq %%rax\n")
	case T_UINT16:
		printf("  movzwq %d(%%rax), %%rax # load uint16\n", 0)
		printf("  pushq %%rax\n")
	case T_INT, T_BOOL:
		printf("  movq %d(%%rax), %%rax # load 64 bit\n", 0)
		printf("  pushq %%rax\n")
	case T_UINTPTR, T_POINTER, T_MAP, T_FUNC:
		printf("  movq %d(%%rax), %%rax # load 64 bit pointer\n", 0)
		printf("  pushq %%rax\n")
	case T_ARRAY, T_STRUCT:
		// pure proxy
		printf("  pushq %%rax\n")
	default:
		unexpectedKind(kind(t))
	}
}

func emitVariable(variable *Variable) {
	emitVariableAddr(variable)
	emitLoadAndPush(variable.Typ)
}

func emitFuncAddr(funcQI QualifiedIdent) {
	printf("  leaq %s(%%rip), %%rax # func addr\n", string(funcQI))
	printf("  pushq %%rax # func addr\n")
}

func emitVariableAddr(variable *Variable) {
	emitComment(2, "emit Addr of variable \"%s\" \n", variable.Name)

	if variable.IsGlobal {
		printf("  leaq %s(%%rip), %%rax # global variable \"%s\"\n", variable.GlobalSymbol, variable.Name)
	} else {
		printf("  leaq %d(%%rbp), %%rax # local variable \"%s\"\n", variable.LocalOffset, variable.Name)
	}

	printf("  pushq %%rax # variable address\n")
}

func emitListHeadAddr(list ast.Expr) {
	t := getTypeOfExpr(list)
	switch kind(t) {
	case T_ARRAY:
		emitAddr(list) // array head
	case T_SLICE:
		emitExpr(list, nil)
		emitPopSlice()
		printf("  pushq %%rax # slice.ptr\n")
	case T_STRING:
		emitExpr(list, nil)
		emitPopString()
		printf("  pushq %%rax # string.ptr\n")
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
		switch e.Obj.Kind {
		case ast.Var:
			if e.Obj.Data == nil {
				panic(fmt.Sprintf("%#v", e.Obj))
			}
			assert(e.Obj.Data != nil, "Obj.Data is nil: " + e.Name, __func__)
			vr := e.Obj.Data.(*Variable)
			emitVariableAddr(vr)
		case ast.Fun:
			qi := newQI(currentPkg.name, e.Obj.Name)
			emitFuncAddr(qi)
		default:
			panic("Unexpected kind")
		}
	case *ast.IndexExpr:
		list := e.X
		if kind(getTypeOfExpr(list)) == T_MAP {
			emitAddrForMapSet(e)
		} else {
			elmType := getTypeOfExpr(e)
			emitExpr(e.Index, nil) // index number
			emitListElementAddr(list, elmType)
		}
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
		assert(e.Obj != nil, "e.Obj should not be nil: "+e.Name, __func__)
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
				printf("  pushq %%rcx # str len\n")
				printf("  pushq %%rax # str ptr\n")
			case T_STRING: // string(string)
				emitExpr(arg0, nil)
			default:
				unexpectedKind(kind(getTypeOfExpr(arg0)))
			}
		case gInt, gUint8, gUint16, gUintptr: // int(e)
			emitExpr(arg0, nil)
		default:
			if to.Obj.Kind == ast.Typ {
				ctx := &evalContext{_type: toType}
				emitExpr(arg0, ctx)
			} else {
				throw(to.Obj)
			}
		}
	case *ast.SelectorExpr:
		// pkg.Type(arg0)
		qi := selector2QI(to)
		ff := lookupForeignIdent(qi)
		assert(ff.Obj.Kind == ast.Typ, "should be ast.Typ", __func__)
		emitConversion(e2t(ff), arg0)
	case *ast.ArrayType: // Conversion to slice
		arrayType := to
		if arrayType.Len != nil {
			throw(to)
		}
		assert(kind(getTypeOfExpr(arg0)) == T_STRING, "source type should be slice", __func__)
		emitComment(2, "Conversion of string => slice \n")
		emitExpr(arg0, nil)
		emitPopString()
		printf("  pushq %%rcx # cap\n")
		printf("  pushq %%rcx # len\n")
		printf("  pushq %%rax # ptr\n")
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
		printf("  pushq $0 # slice cap\n")
		printf("  pushq $0 # slice len\n")
		printf("  pushq $0 # slice ptr\n")
	case T_STRING:
		printf("  pushq $0 # string len\n")
		printf("  pushq $0 # string ptr\n")
	case T_INTERFACE:
		printf("  pushq $0 # interface data\n")
		printf("  pushq $0 # interface dtype\n")
	case T_INT, T_UINT8, T_BOOL:
		printf("  pushq $0 # %s zero value (number)\n", string(kind(t)))
	case T_UINTPTR, T_POINTER, T_MAP, T_FUNC:
		printf("  pushq $0 # %s zero value (nil pointer)\n", string(kind(t)))
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
		printf("  pushq %%rcx # len\n")
	case T_STRING:
		emitExpr(arg, nil)
		emitPopString()
		printf("  pushq %%rcx # len\n")
	case T_MAP:
		args := []*Arg{
			// len
			&Arg{
				e:         arg,
				paramType: getTypeOfExpr(arg),
			},
		}
		resultList := &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Type: tInt.E,
				},
			},
		}
		emitCall("runtime.lenMap", args, resultList)

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
		printf("  pushq %%rdx # cap\n")
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
	printf("  pushq $%d\n", size)
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
	printf("  xor $1, %%rax\n")
	printf("  pushq %%rax\n")
}

func emitTrue() {
	printf("  pushq $1 # true\n")
}

func emitFalse() {
	printf("  pushq $0 # false\n")
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
func emitCall(fn interface{}, args []*Arg, resultList *ast.FieldList) {
	emitComment(2, "emitCall len(args)=%d\n", len(args))

	var totalParamSize int
	for _, arg := range args {
		arg.offset = totalParamSize
		totalParamSize += getSizeOfType(arg.paramType)
	}

	emitAllocReturnVarsArea(getTotalFieldsSize(resultList))
	printf("  subq $%d, %%rsp # alloc parameters area\n", totalParamSize)
	for _, arg := range args {
		paramType := arg.paramType
		ctx := &evalContext{
			_type: paramType,
		}
		emitExprIfc(arg.e, ctx)
		emitPop(kind(paramType))
		printf("  leaq %d(%%rsp), %%rsi # place to save\n", arg.offset)
		printf("  pushq %%rsi # place to save\n")
		emitRegiToMem(paramType)
	}

	emitCallQ(fn, totalParamSize, resultList)
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

func emitCallQ(fn interface{}, totalParamSize int, resultList *ast.FieldList) {
	switch f := fn.(type){
	case string:
		symbol := f
		printf("  callq %s\n", symbol)
	case *ast.Ident:
		emitExpr(f, nil)
		printf("  popq %%rax\n")
		printf("  callq *%%rax\n")
	default:
		throw(fn)
	}
	emitFreeParametersArea(totalParamSize)
	printf("#  totalReturnSize=%d\n", getTotalFieldsSize(resultList))
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
	printf("  leave\n")
	printf("  ret\n")
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
			printf("  movzbq (%%rsp), %%rax # load uint8\n")
			printf("  addq $%d, %%rsp # free returnvars area\n", 1)
			printf("  pushq %%rax\n")
		case T_BOOL, T_INT, T_UINTPTR, T_POINTER, T_MAP:
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
	emitComment(2, "[emitFuncall] %T(...)\n", fun)
	var funcType *ast.FuncType
	var funcVal interface{}
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
			case T_MAP:
				mapType := getUnderlyingType(typeArg).E.(*ast.MapType)
				valueSize := newNumberLiteral(getSizeOfType(e2t(mapType.Value)))
				// A new, empty map value is made using the built-in function make,
				// which takes the map type and an optional capacity hint as arguments:
				length := newNumberLiteral(0)
				args := []*Arg{
					&Arg{
						e:         length,
						paramType: tUintptr,
					},
					&Arg{
						e:         valueSize,
						paramType: tUintptr,
					},
				}
				resultList := &ast.FieldList{
					List: []*ast.Field{
						&ast.Field{
							Type: tUintptr.E,
						},
					},
				}
				emitCall("runtime.makeMap", args, resultList)
				return
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
			elmType := getElementTypeOfCollectionType(getTypeOfExpr(sliceArg))
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
			funcVal = "runtime.panic"
			_args := []*Arg{&Arg{
				e:         eArgs[0],
				paramType: tEface,
			}}
			emitCall(funcVal, _args, nil)
			return
		case gDelete:
			funcVal = "runtime.deleteMap"
			_args := []*Arg{
				&Arg{
					e:         eArgs[0],
					paramType: getTypeOfExpr(eArgs[0]),
				},
				&Arg{
					e:         eArgs[1],
					paramType: tEface,
				},
			}
			emitCall(funcVal, _args, nil)
			return
		}

		if fn.Name == "makeSlice1" || fn.Name == "makeSlice8" || fn.Name == "makeSlice16" || fn.Name == "makeSlice24" {
			fn.Name = "makeSlice"
		}
		// general function call
		funcVal = getPackageSymbol(currentPkg.name, fn.Name)
		if currentPkg.name == "os" && fn.Name == "runtime_args" {
			funcVal = "runtime.runtime_args"
		} else if currentPkg.name == "os" && fn.Name == "runtime_getenv" {
			funcVal = "runtime.runtime_getenv"
		}

		switch dcl := fn.Obj.Decl.(type) {
		case *ast.FuncDecl:
			funcType =  dcl.Type
		case *ast.ValueSpec: // var f func()
			funcType =  dcl.Type.(*ast.FuncType)
			funcVal = fun
		case *ast.AssignStmt: //  f := anotherF
			rhs := dcl.Rhs[0]
			switch r := rhs.(type) {
			case *ast.SelectorExpr:
				assert(isQI(r), "expect QI", __func__)
				ff := lookupForeignFunc(selector2QI(r))
				funcType = ff.decl.Type
				funcVal = fun
			default:
				throw(r)
			}
		default:
			throw(dcl)
		}
	case *ast.SelectorExpr:
		if isQI(fn) {
			// pkg.Sel()
			qi := selector2QI(fn)
			funcVal = string(qi)
			ff := lookupForeignFunc(qi)
			funcType = ff.decl.Type
		} else {
			// method call
			receiver = fn.X
			receiverType := getTypeOfExpr(receiver)
			method := lookupMethod(receiverType, fn.Sel)
			funcType = method.FuncType
			funcVal = getMethodSymbol(method)

			if kind(receiverType) == T_POINTER {
				if method.IsPtrMethod {
					// p.mp() => as it is
				} else {
					// p.mv()
					panic("TBI")
				}
			} else {
				if method.IsPtrMethod {
					// v.mp() => (&v).mp()
					// @TODO we should check addressable
					receiver = &ast.UnaryExpr{
						Op: token.AND,
						X:  receiver,
					}
				} else {
					// v.mv() => as it is
				}
			}
		}
	case *ast.ParenExpr:
		panic("[astParenExpr] TBI ")
	default:
		throw(fun)
	}

	args := prepareArgs(funcType, receiver, eArgs, hasEllissis)
	emitCall(funcVal, args, funcType.Results)
}

func emitNil(targetType *Type) {
	if targetType == nil {
		panic("Type is required to emit nil")
	}
	switch kind(targetType) {
	case T_SLICE, T_POINTER, T_INTERFACE, T_MAP:
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

type evalContext struct {
	okContext bool
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
		case ast.Fun:
			emitAddr(e)
		default:
			panic("Unexpected ident kind:" + e.Obj.Kind.String())
		}
	}
	return false
}

// 1 or 2 values
func emitIndexExpr(e *ast.IndexExpr, ctx *evalContext) {
	if kind(getTypeOfExpr(e.X)) == T_MAP {
		emitMapGet(e, ctx)
	} else {
		emitAddr(e)
		emitLoadAndPush(getTypeOfExpr(e))
	}
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
		qi := selector2QI(e)
		ident := lookupForeignIdent(qi)
		if ident.Obj.Kind == ast.Fun {
			emitFuncAddr(qi)
		} else {
			emitExpr(ident, ctx)
		}
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
		emitComment(2, "[emitCallExpr] Conversion\n")
		emitConversion(e2t(fun), e.Args[0])
	} else {
		emitComment(2, "[emitCallExpr] Funcall\n")
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
		printf("  pushq $%d # convert char literal to int\n", int(char))
	case "INT":
		ival := strconv.Atoi(e.Value)
		printf("  pushq $%d # number literal\n", ival)
	case "STRING":
		sl := getStringLiteral(e)
		if sl.strlen == 0 {
			// zero value
			emitZeroValue(tString)
		} else {
			printf("  pushq $%d # str len\n", sl.strlen)
			printf("  leaq %s(%%rip), %%rax # str ptr\n", sl.label)
			printf("  pushq %%rax # str ptr\n")
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
		printf("  popq %%rax # e.X\n")
		printf("  imulq $-1, %%rax\n")
		printf("  pushq %%rax\n")
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
		printf("  cmpq $1, %%rax\n")
		// exit with false if left is false
		printf("  jne %s\n", labelExitWithFalse)

		// if left is true, then eval right and exit
		emitExpr(e.Y, nil) // right
		printf("  jmp %s\n", labelExit)

		printf("  %s:\n", labelExitWithFalse)
		emitFalse()
		printf("  %s:\n", labelExit)
	case "||":
		labelid++
		labelExitWithTrue := fmt.Sprintf(".L.%d.true", labelid)
		labelExit := fmt.Sprintf(".L.%d.exit", labelid)
		emitExpr(e.X, nil) // left
		emitPopBool("left")
		printf("  cmpq $1, %%rax\n")
		// exit with true if left is true
		printf("  je %s\n", labelExitWithTrue)

		// if left is false, then eval right and exit
		emitExpr(e.Y, nil) // right
		printf("  jmp %s\n", labelExit)

		printf("  %s:\n", labelExitWithTrue)
		emitTrue()
		printf("  %s:\n", labelExit)
	case "+":
		if kind(getTypeOfExpr(e.X)) == T_STRING {
			emitCatStrings(e.X, e.Y)
		} else {
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			printf("  popq %%rcx # right\n")
			printf("  popq %%rax # left\n")
			printf("  addq %%rcx, %%rax\n")
			printf("  pushq %%rax\n")
		}
	case "-":
		emitExpr(e.X, nil) // left
		emitExpr(e.Y, nil) // right
		printf("  popq %%rcx # right\n")
		printf("  popq %%rax # left\n")
		printf("  subq %%rcx, %%rax\n")
		printf("  pushq %%rax\n")
	case "*":
		emitExpr(e.X, nil) // left
		emitExpr(e.Y, nil) // right
		printf("  popq %%rcx # right\n")
		printf("  popq %%rax # left\n")
		printf("  imulq %%rcx, %%rax\n")
		printf("  pushq %%rax\n")
	case "%":
		emitExpr(e.X, nil) // left
		emitExpr(e.Y, nil) // right
		printf("  popq %%rcx # right\n")
		printf("  popq %%rax # left\n")
		printf("  movq $0, %%rdx # init %%rdx\n")
		printf("  divq %%rcx\n")
		printf("  movq %%rdx, %%rax\n")
		printf("  pushq %%rax\n")
	case "/":
		emitExpr(e.X, nil) // left
		emitExpr(e.Y, nil) // right
		printf("  popq %%rcx # right\n")
		printf("  popq %%rax # left\n")
		printf("  movq $0, %%rdx # init %%rdx\n")
		printf("  divq %%rcx\n")
		printf("  pushq %%rax\n")
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
		printf("  pushq $%d # slice.cap\n", length)
		printf("  pushq $%d # slice.len\n", length)
		printf("  pushq %%rax # slice.ptr\n")
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
			printf("  popq %%rcx # low\n")
			printf("  popq %%rax # orig_cap\n")
			printf("  subq %%rcx, %%rax # orig_cap - low\n")
			printf("  pushq %%rax # new cap\n")

			// new len = high - low
			if e.High != nil {
				emitExpr(e.High, nil)
			} else {
				// high = len(orig)
				emitLen(e.X)
			}
			emitExpr(low, nil)
			printf("  popq %%rcx # low\n")
			printf("  popq %%rax # high\n")
			printf("  subq %%rcx, %%rax # high - low\n")
			printf("  pushq %%rax # new len\n")
		} else {
			// new cap = max - low
			emitExpr(e.Max, nil)
			emitExpr(low, nil)
			printf("  popq %%rcx # low\n")
			printf("  popq %%rax # max\n")
			printf("  subq %%rcx, %%rax # new cap = max - low\n")
			printf("  pushq %%rax # new cap\n")
			// new len = high - low
			emitExpr(e.High, nil)
			emitExpr(low, nil)
			printf("  popq %%rcx # low\n")
			printf("  popq %%rax # high\n")
			printf("  subq %%rcx, %%rax # new len = high - low\n")
			printf("  pushq %%rax # new len\n")
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
		printf("  popq %%rcx # low\n")
		printf("  popq %%rax # high\n")
		printf("  subq %%rcx, %%rax # high - low\n")
		printf("  pushq %%rax # len\n")
		// no cap
	default:
		unexpectedKind(kind(listType))
	}

	emitExpr(low, nil) // index number
	elmType := getElementTypeOfCollectionType(listType)
	emitListElementAddr(list, elmType)
}

// 1 or 2 values
func emitMapGet(e *ast.IndexExpr, ctx *evalContext) {
	// MAP GET
	valueType := getTypeOfExpr(e)
	emitComment(2, "MAP GET for map[string]string\n")
	// emit addr of map element
	mp := e.X
	key := e.Index

	args := []*Arg{
		&Arg{
			e:         mp,
			paramType: tUintptr,
		},
		&Arg{
			e:         key,
			paramType: tEface,
		},
	}
	resultList := &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type: tBool.E,
			},
			&ast.Field{
				Type: tUintptr.E,
			},
		},
	}
	emitCall("runtime.getAddrForMapGet", args, resultList)
	// return values = [ptr, bool(stack top)]
	emitPopBool("map get:  ok value")
	printf("  cmpq $1, %%rax\n")
	labelid++
	labelEnd := fmt.Sprintf(".L.end_map_get.%d", labelid)
	labelElse := fmt.Sprintf(".L.not_found.%d", labelid)
	printf("  jne %s # jmp if false\n", labelElse)

	okContext := ctx != nil && ctx.okContext

	// if matched
	emitLoadAndPush(valueType)
	if okContext {
		printf("  pushq $1 # ok = true\n")
	}
	// exit
	printf("  jmp %s\n", labelEnd)

	// if not matched
	printf("  %s:\n", labelElse)
	emitPop(T_POINTER) // destroy nil
	emitZeroValue(valueType)
	if okContext {
		printf("  pushq $0 # ok = false\n")
	}

	printf("  %s:\n", labelEnd)
}

// 1 or 2 values
func emitTypeAssertExpr(e *ast.TypeAssertExpr, ctx *evalContext) {
	emitExpr(e.X, nil)
	emitDtypeLabelAddr(e2t(e.Type))
	emitCompareDtypes()

	emitPopBool("type assertion ok value")
	printf("  cmpq $1, %%rax\n")

	labelid++
	labelEnd := fmt.Sprintf(".L.end_type_assertion.%d", labelid)
	labelElse := fmt.Sprintf(".L.unmatch.%d", labelid)
	printf("  jne %s # jmp if false\n", labelElse)

	okContext := ctx != nil && ctx.okContext
	// if matched
	emitLoadAndPush(e2t(e.Type)) // load dynamic data
	if okContext {
		printf("  pushq $1 # ok = true\n")
	}
	// exit
	printf("  jmp %s\n", labelEnd)
	// if not matched
	printf("  %s:\n", labelElse)
	printf("  popq %%rax # drop ifc.data\n")
	emitZeroValue(e2t(e.Type))
	if okContext {
		printf("  pushq $0 # ok = false\n")
	}

	printf("  %s:\n", labelEnd)
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
	// push dtype label's address
	emitDtypeLabelAddr(fromType)
}

func emitExprIfc(expr ast.Expr, ctx *evalContext) {
	isNilObj := emitExpr(expr, ctx)
	if !isNilObj && ctx != nil && ctx._type != nil && isInterface(ctx._type) && !isInterface(getTypeOfExpr(expr)) {
		emitConvertToInterface(getTypeOfExpr(expr))
	}
}

type dtypeEntry struct {
	id         int
	pkgname    string
	serialized string
	label      string
}

var typeId int
var typesMap map[string]*dtypeEntry

// "**[1][]*int" => "dtype.8"
func getDtypeLabel(pkg *PkgContainer, serializedType string) string {
	s := pkg.name + ":" + serializedType
	ent, ok := typesMap[s]
	if ok {
		return ent.label
	} else {
		id := typeId
		ent = &dtypeEntry{
			id:         id,
			pkgname:    pkg.name,
			serialized: serializedType,
			label:      pkg.name + "." + "dtype." + strconv.Itoa(id),
		}
		typesMap[s] = ent
		typeId++
	}

	return ent.label
}

// Check type identity by comparing its serialization, not id or address of dtype label.
// pop pop, compare and push 1(match) or 0(not match)
func emitCompareDtypes() {
	labelid++
	labelTrue := fmt.Sprintf(".L.cmpdtypes.%d.true", labelid)
	labelFalse := fmt.Sprintf(".L.cmpdtypes.%d.false", labelid)
	labelEnd := fmt.Sprintf(".L.cmpdtypes.%d.end", labelid)
	labelCmp := fmt.Sprintf(".L.cmpdtypes.%d.cmp", labelid)
	printf("  popq %%rdx           # dtype label address A\n")
	printf("  popq %%rcx           # dtype label address B\n")

	printf("  cmpq %%rcx, %%rdx\n")
	printf("  je %s # jump if match\n", labelTrue)

	printf("  cmpq $0, %%rdx # check if A is nil\n")
	printf("  je %s # jump if nil\n", labelFalse)

	printf("  cmpq $0, %%rcx # check if B is nil\n")
	printf("  je %s # jump if nil\n", labelFalse)

	printf("  jmp %s # jump to end\n", labelCmp)

	printf("%s:\n", labelTrue)
	printf("  pushq $1\n")
	printf("  jmp %s # jump to end\n", labelEnd)

	printf("%s:\n", labelFalse)
	printf("  pushq $0\n")
	printf("  jmp %s # jump to end\n", labelEnd)

	printf("%s:\n", labelCmp)
	emitAllocReturnVarsArea(SizeOfInt) // for bool

	// push len, push ptr
	printf("  movq 16(%%rax), %%rdx           # str.len of dtype A\n")
	printf("  pushq %%rdx\n")
	printf("  movq 8(%%rax), %%rdx           # str.ptr of dtype A\n")
	printf("  pushq %%rdx\n")

	// push len, push ptr
	printf("  movq 16(%%rcx), %%rdx           # str.len of dtype B\n")
	printf("  pushq %%rdx\n")
	printf("  movq 8(%%rcx), %%rdx           # str.ptr of dtype B\n")
	printf("  pushq %%rdx\n")

	printf("  callq %s\n", "runtime.cmpstrings")
	emitFreeParametersArea(16 * 2)
	printf("%s:\n", labelEnd)
}

func emitDtypeLabelAddr(t *Type) {
	serializedType := serializeType(t)
	dtypeLabel := getDtypeLabel(currentPkg, serializedType)
	printf("  leaq %s(%%rip), %%rax # dtype label address \"%s\"\n", dtypeLabel, serializedType)
	printf("  pushq %%rax           # dtype label address\n")
}

func newNumberLiteral(x int) *ast.BasicLit {
	e := &ast.BasicLit{
		Kind:  token.INT,
		Value: strconv.Itoa(x),
	}
	return e
}

func emitAddrForMapSet(indexExpr *ast.IndexExpr) {
	// alloc heap for map value
	//size := getSizeOfType(elmType)
	emitComment(2, "[emitAddrForMapSet]\n")
	mp := indexExpr.X
	key := indexExpr.Index

	args := []*Arg{
		&Arg{
			e:         mp,
			paramType: tUintptr,
		},
		&Arg{
			e:         key,
			paramType: tEface,
		},
	}
	resultList := &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type: tUintptr.E,
			},
		},
	}
	emitCall("runtime.getAddrForMapSet", args, resultList)
}

func emitListElementAddr(list ast.Expr, elmType *Type) {
	emitListHeadAddr(list)
	emitPopAddress("list head")
	printf("  popq %%rcx # index id\n")
	printf("  movq $%d, %%rdx # elm size\n", getSizeOfType(elmType))
	printf("  imulq %%rdx, %%rcx\n")
	printf("  addq %%rcx, %%rax\n")
	printf("  pushq %%rax # addr of element\n")
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
	printf("  popq %%rcx # right\n")
	printf("  popq %%rax # left\n")
	printf("  cmpq %%rcx, %%rax\n")
	printf("  %s %%al\n", inst)
	printf("  movzbq %%al, %%rax\n") // true:1, false:0
	printf("  pushq %%rax\n")
}

func emitPop(knd TypeKind) {
	switch knd {
	case T_SLICE:
		emitPopSlice()
	case T_STRING:
		emitPopString()
	case T_INTERFACE:
		emitPopInterFace()
	case T_INT, T_BOOL:
		emitPopPrimitive(string(knd))
	case T_UINTPTR, T_POINTER, T_MAP, T_FUNC:
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
		printf("  popq %%rsi # lhs addr\n")
	} else {
		printf("  popq %%rsi # lhs addr\n")
		emitPop(knd) // rhs
	}
	if pushLhs {
		printf("  pushq %%rsi # lhs addr\n")
	}

	printf("  pushq %%rsi # place to save\n")
	emitRegiToMem(t)
}

func emitRegiToMem(t *Type) {
	printf("  popq %%rsi # place to save\n")
	k := kind(t)
	switch k {
	case T_SLICE:
		printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", 0)
		printf("  movq %%rcx, %d(%%rsi) # len to len\n", 8)
		printf("  movq %%rdx, %d(%%rsi) # cap to cap\n", 16)
	case T_STRING:
		printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", 0)
		printf("  movq %%rcx, %d(%%rsi) # len to len\n", 8)
	case T_INTERFACE:
		printf("  movq %%rax, %d(%%rsi) # store dtype\n", 0)
		printf("  movq %%rcx, %d(%%rsi) # store data\n", 8)
	case T_INT, T_BOOL:
		printf("  movq %%rax, %d(%%rsi) # assign quad\n", 0)
	case T_UINTPTR, T_POINTER, T_MAP, T_FUNC:
		printf("  movq %%rax, %d(%%rsi) # assign ptr\n", 0)
	case T_UINT16:
		printf("  movw %%ax, %d(%%rsi) # assign word\n", 0)
	case T_UINT8:
		printf("  movb %%al, %d(%%rsi) # assign byte\n", 0)
	case T_STRUCT, T_ARRAY:
		printf("  pushq $%d # size\n", getSizeOfType(t))
		printf("  pushq %%rsi # dst lhs\n")
		printf("  pushq %%rax # src rhs\n")
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
		indexExpr, isIndexExpr := rhs0.(*ast.IndexExpr)
		var isOKSytax bool
		if len(s.Lhs) == 2 && isTypeAssertion {
			isOKSytax = true
		}
		if len(s.Lhs) == 2 && isIndexExpr && kind(getTypeOfExpr(indexExpr.X)) == T_MAP {
			isOKSytax = true
		}
		if isOKSytax {
			emitComment(2, "Assignment: emitAssignWithOK rhs\n")
			ctx := &evalContext{
				okContext: true,
			}
			emitExprIfc(rhs0, ctx) // {push data}, {push bool}
			lhsOK := s.Lhs[1]
			if isBlankIdentifier(lhsOK) {
				emitPop(T_BOOL)
			} else {
				emitComment(2, "Assignment: ok variable\n")
				emitAddr(lhsOK)
				emitStore(getTypeOfExpr(lhsOK), false, false)
			}

			lhsMain := s.Lhs[0]
			if isBlankIdentifier(lhsMain) {
				emitPop(kind(getTypeOfExpr(rhs0)))
			} else {
				emitAddr(lhsMain)
				emitComment(2, "Assignment: emitStore(getTypeOfExpr(lhs))\n")
				emitStore(getTypeOfExpr(lhsMain), false, false)
			}
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
				printf("# len lhs=%d\n", len(s.Lhs))
				printf("# returnTypes=%d\n", len(returnTypes))
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
							printf("  movzbq (%%rsp), %%rax # load uint8\n")
							printf("  addq $%d, %%rsp # free returnvars area\n", 1)
							printf("  pushq %%rax\n")
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
	printf("  cmpq $1, %%rax\n")
	if s.Else != nil {
		printf("  jne %s # jmp if false\n", labelElse)
		emitStmt(s.Body) // then
		printf("  jmp %s\n", labelEndif)
		printf("  %s:\n", labelElse)
		emitStmt(s.Else) // then
	} else {
		printf("  jne %s # jmp if false\n", labelEndif)
		emitStmt(s.Body) // then
	}
	printf("  %s:\n", labelEndif)
	emitComment(2, "end if\n")
}

func emitForStmt(s *ast.ForStmt) {
	meta := getMetaForStmt(s)
	labelid++
	labelCond := fmt.Sprintf(".L.for.cond.%d", labelid)
	labelPost := fmt.Sprintf(".L.for.post.%d", labelid)
	labelExit := fmt.Sprintf(".L.for.exit.%d", labelid)

	meta.LabelPost = labelPost
	meta.LabelExit = labelExit

	if s.Init != nil {
		emitStmt(s.Init)
	}

	printf("  %s:\n", labelCond)
	if s.Cond != nil {
		emitExpr(s.Cond, nil)
		emitPopBool("for condition")
		printf("  cmpq $1, %%rax\n")
		printf("  jne %s # jmp if false\n", labelExit)
	}
	emitStmt(s.Body)
	printf("  %s:\n", labelPost) // used for "continue"
	if s.Post != nil {
		emitStmt(s.Post)
	}
	printf("  jmp %s\n", labelCond)
	printf("  %s:\n", labelExit)
}

func emitRangeMap(s *ast.RangeStmt, meta *MetaForStmt) {
	labelid++
	labelCond := fmt.Sprintf(".L.range.cond.%d", labelid)
	labelPost := fmt.Sprintf(".L.range.post.%d", labelid)
	labelExit := fmt.Sprintf(".L.range.exit.%d", labelid)

	meta.LabelPost = labelPost
	meta.LabelExit = labelExit

	// Overall design:
	//  _mp := EXPR
	//  if _mp == nil then exit
	// 	for _item = _mp.first; _item != nil; item = item.next {
	//    ...
	//  }

	emitComment(2, "ForRange map Initialization\n")

	// _mp = EXPR
	emitAssignToVar(meta.ForRange.MapVar, s.X)

	//  if _mp == nil then exit
	emitVariable(meta.ForRange.MapVar) // value of _mp
	printf("  popq %%rax\n")
	printf("  cmpq $0, %%rax\n")
	printf("  je %s # exit if nil\n", labelExit)

	// item = mp.first
	emitVariableAddr(meta.ForRange.ItemVar)
	emitVariable(meta.ForRange.MapVar) // value of _mp
	emitLoadAndPush(tUintptr)          // value of _mp.first
	emitStore(tUintptr, true, false)   // assign

	// Condition
	// if item != nil; then
	//   execute body
	// else
	//   exit
	emitComment(2, "ForRange Condition\n")
	printf("  %s:\n", labelCond)

	emitVariable(meta.ForRange.ItemVar)
	printf("  popq %%rax\n")
	printf("  cmpq $0, %%rax\n")
	printf("  je %s # exit if nil\n", labelExit)

	emitComment(2, "assign key value to variables\n")

	// assign key
	if s.Key != nil {
		keyIdent := s.Key.(*ast.Ident)
		if keyIdent.Name != "_" {
			emitAddr(s.Key) // lhs
			// emit value of item.key
			//type item struct {
			//	next  *item
			//	key_dtype uintptr
			//  key_data uintptr <-- this
			//	value uintptr
			//}
			emitVariable(meta.ForRange.ItemVar)
			printf("  popq %%rax\n")            // &item{....}
			printf("  movq 16(%%rax), %%rcx\n") // item.key_data
			printf("  pushq %%rcx\n")
			emitLoadAndPush(getTypeOfExpr(s.Key)) // load dynamic data
			emitStore(getTypeOfExpr(s.Key), true, false)
		}
	}

	// assign value
	if s.Value != nil {
		valueIdent := s.Value.(*ast.Ident)
		if valueIdent.Name != "_" {
			emitAddr(s.Value) // lhs
			// emit value of item
			//type item struct {
			//	next  *item
			//	key_dtype uintptr
			//  key_data uintptr
			//	value uintptr  <-- this
			//}
			emitVariable(meta.ForRange.ItemVar)
			printf("  popq %%rax\n")            // &item{....}
			printf("  movq 24(%%rax), %%rcx\n") // item.key_data
			printf("  pushq %%rcx\n")
			emitLoadAndPush(getTypeOfExpr(s.Value)) // load dynamic data
			emitStore(getTypeOfExpr(s.Value), true, false)
		}
	}

	// Body
	emitComment(2, "ForRange Body\n")
	emitStmt(s.Body)

	// Post statement
	// item = item.next
	emitComment(2, "ForRange Post statement\n")
	printf("  %s:\n", labelPost)            // used for "continue"
	emitVariableAddr(meta.ForRange.ItemVar) // lhs
	emitVariable(meta.ForRange.ItemVar)     // item
	emitLoadAndPush(tUintptr)               // item.next
	emitStore(tUintptr, true, false)

	printf("  jmp %s\n", labelCond)

	printf("  %s:\n", labelExit)
}

// only for array and slice for now
func emitRangeStmt(s *ast.RangeStmt) {
	meta := getMetaForStmt(s)
	if meta.ForRange.IsMap {
		emitRangeMap(s, meta)
		return
	}
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
	emitVariableAddr(meta.ForRange.LenVar)
	emitLen(s.X)
	emitStore(tInt, true, false)

	emitComment(2, "  assign 0 to indexvar\n")
	// indexvar = 0
	emitVariableAddr(meta.ForRange.Indexvar)
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
	printf("  %s:\n", labelCond)

	emitVariableAddr(meta.ForRange.Indexvar)
	emitLoadAndPush(tInt)
	emitVariableAddr(meta.ForRange.LenVar)
	emitLoadAndPush(tInt)
	emitCompExpr("setl")
	emitPopBool(" indexvar < lenvar")
	printf("  cmpq $1, %%rax\n")
	printf("  jne %s # jmp if false\n", labelExit)

	emitComment(2, "assign list[indexvar] value variables\n")
	elemType := getTypeOfExpr(s.Value)
	emitAddr(s.Value) // lhs

	emitVariableAddr(meta.ForRange.Indexvar)
	emitLoadAndPush(tInt) // index value
	emitListElementAddr(s.X, elemType)

	emitLoadAndPush(elemType)
	emitStore(elemType, true, false)

	// Body
	emitComment(2, "ForRange Body\n")
	emitStmt(s.Body)

	// Post statement: Increment indexvar and go next
	emitComment(2, "ForRange Post statement\n")
	printf("  %s:\n", labelPost)             // used for "continue"
	emitVariableAddr(meta.ForRange.Indexvar) // lhs
	emitVariableAddr(meta.ForRange.Indexvar) // rhs
	emitLoadAndPush(tInt)
	emitAddConst(1, "indexvar value ++")
	emitStore(tInt, true, false)

	// incr key variable
	if s.Key != nil {
		keyIdent := s.Key.(*ast.Ident)
		if keyIdent.Name != "_" {
			emitAddr(s.Key)                          // lhs
			emitVariableAddr(meta.ForRange.Indexvar) // rhs
			emitLoadAndPush(tInt)
			emitStore(tInt, true, false)
		}
	}

	printf("  jmp %s\n", labelCond)

	printf("  %s:\n", labelExit)
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
		if len(cc.List) == 0 {
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
			printf("  cmpq $1, %%rax\n")
			printf("  je %s # jump if match\n", labelCase)
		}
	}
	emitComment(2, "End comparison with cases\n")

	// if no case matches, then jump to
	if defaultLabel != "" {
		// default
		printf("  jmp %s\n", defaultLabel)
	} else {
		// exit
		printf("  jmp %s\n", labelEnd)
	}

	emitRevertStackTop(condType)
	for i, c := range cases {
		cc := c.(*ast.CaseClause)
		printf("%s:\n", labels[i])
		for _, _s := range cc.Body {
			emitStmt(_s)
		}
		printf("  jmp %s\n", labelEnd)
	}
	printf("%s:\n", labelEnd)
}
func emitTypeSwitchStmt(s *ast.TypeSwitchStmt) {
	meta := getMetaTypeSwitchStmt(s)
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
		if len(cc.List) == 0 {
			defaultLabel = labelCase
			continue
		}
		for _, e := range cc.List {
			emitVariableAddr(meta.SubjectVariable)
			emitPopAddress("type switch subject")
			printf("  movq (%%rax), %%rax # dtype label addr\n")
			printf("  pushq %%rax # dtype label addr\n")

			if isNil(cc.List[0]) { // case nil:
				printf("  pushq $0 # nil\n")
			} else { // case T:
				emitDtypeLabelAddr(e2t(e))
			}
			emitCompareDtypes()
			emitPopBool(" of switch-case comparison")

			printf("  cmpq $1, %%rax\n")
			printf("  je %s # jump if match\n", labelCase)
		}
	}
	emitComment(2, "End comparison with cases\n")

	// if no case matches, then jump to
	if defaultLabel != "" {
		// default
		printf("  jmp %s\n", defaultLabel)
	} else {
		// exit
		printf("  jmp %s\n", labelEnd)
	}

	for i, typeSwitchCaseClose := range meta.Cases {
		// Injecting variable and type to the subject
		if typeSwitchCaseClose.Variable != nil {
			setVariable(meta.AssignIdent.Obj, typeSwitchCaseClose.Variable)
		}
		printf("%s:\n", labels[i])

		cc := typeSwitchCaseClose.Orig
		var _isNil bool
		for _, typ := range cc.List {
			if isNil(typ) {
				_isNil = true
			}
		}
		for _, _s := range cc.Body {
			if typeSwitchCaseClose.Variable != nil {
				// do assignment
				if _isNil {
					// @TODO: assign nil to the AssignIdent of interface type
				} else {
					emitAddr(meta.AssignIdent) // push lhs

					// push rhs
					emitVariableAddr(meta.SubjectVariable)
					emitLoadAndPush(tEface)
					printf("  popq %%rax # ifc.dtype\n")
					printf("  popq %%rcx # ifc.data\n")
					printf("  pushq %%rcx # ifc.data\n")
					emitLoadAndPush(typeSwitchCaseClose.VariableType)

					// assign
					emitStore(typeSwitchCaseClose.VariableType, true, false)
				}
			}

			emitStmt(_s)
		}
		printf("  jmp %s\n", labelEnd)
	}
	printf("%s:\n", labelEnd)
}
func emitBranchStmt(s *ast.BranchStmt) {
	meta := getMetaBranchStmt(s)
	containerFor := meta.containerForStmt
	switch s.Tok.String() {
	case "continue":
		printf("jmp %s # continue\n", containerFor.LabelPost)
	case "break":
		printf("jmp %s # break\n", containerFor.LabelExit)
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
	printf("  addq $%d, %%rsp # revert stack top\n", getSizeOfType(t))
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
	printf("# emitFuncDecl\n")
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
	printf(".global %s\n", symbol)
	printf("%s: # args %d, locals %d\n", symbol, fnc.Argsarea, fnc.Localarea)
	printf("  pushq %%rbp\n")
	printf("  movq %%rsp, %%rbp\n")
	if len(fnc.LocalVars) > 0 {
		for i := len(fnc.LocalVars) - 1; i >= 0; i-- {
			v := fnc.LocalVars[i]
			logf("  # -%d(%%rbp) local variable %d \"%s\"\n", -v.LocalOffset, getSizeOfType(v.Typ), v.Name)
		}
	}
	logf("  #  0(%%rbp) previous rbp\n")
	logf("  #  8(%%rbp) return address\n")

	if fnc.Localarea != 0 {
		printf("  subq $%d, %%rsp # local area\n", -fnc.Localarea)
	}
	for _, stmt := range fnc.Stmts {
		emitStmt(stmt)
	}
	printf("  leave\n")
	printf("  ret\n")
}

func emitGlobalVariableComplex(name *ast.Ident, t *Type, val ast.Expr) {
	typeKind := kind(t)
	switch typeKind {
	case T_POINTER:
		printf("# init global %s:\n", name.Name)
		emitAssign(name, val)
	case T_MAP:
		emitAssign(name, val)
	case T_INTERFACE:
		emitAssign(name, val)
	}
}

func emitGlobalVariable(pkg *PkgContainer, name *ast.Ident, t *Type, val ast.Expr) {
	typeKind := kind(t)
	printf(".global %s.%s\n", pkg.name, name.Name)
	printf("%s.%s: # T %s\n", pkg.name, name.Name, string(typeKind))
	switch typeKind {
	case T_STRING:
		switch vl := val.(type) {
		case nil:
			printf("  .quad 0\n")
			printf("  .quad 0\n")
		case *ast.BasicLit:
			sl := getStringLiteral(vl)
			printf("  .quad %s\n", sl.label)
			printf("  .quad %d\n", sl.strlen)
		default:
			panic("Unsupported global string value")
		}
	case T_BOOL:
		switch vl := val.(type) {
		case nil:
			printf("  .quad 0 # bool zero value\n")
		case *ast.Ident:
			switch vl.Obj {
			case gTrue:
				printf("  .quad 1 # bool true\n")
			case gFalse:
				printf("  .quad 0 # bool false\n")
			default:
				throw(val)
			}
		default:
			throw(val)
		}
	case T_INT:
		switch vl := val.(type) {
		case nil:
			printf("  .quad 0\n")
		case *ast.BasicLit:
			printf("  .quad %s\n", vl.Value)
		default:
			throw(val)
		}
	case T_UINT8:
		switch vl := val.(type) {
		case nil:
			printf("  .byte 0\n")
		case *ast.BasicLit:
			printf("  .byte %s\n", vl.Value)
		default:
			throw(val)
		}
	case T_UINT16:
		switch vl := val.(type) {
		case nil:
			printf("  .word 0\n")
		case *ast.BasicLit:
			printf("  .word %s\n", vl.Value)
		default:
			throw(val)
		}
	case T_UINTPTR:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		printf("  .quad 0\n")
	case T_SLICE:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		printf("  .quad 0 # ptr\n")
		printf("  .quad 0 # len\n")
		printf("  .quad 0 # cap\n")
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
			printf(zeroValue)
		}
	case T_POINTER:
		// will be set in the initGlobal func
		printf("  .quad 0\n")
	case T_MAP:
		// will be set in the initGlobal func
		printf("  .quad 0\n")
	case T_INTERFACE:
		// will be set in the initGlobal func
		printf("  .quad 0\n")
		printf("  .quad 0\n")
	default:
		unexpectedKind(typeKind)
	}
}

func generateCode(pkg *PkgContainer) {
	printf("#===================== generateCode %s =====================\n", pkg.name)
	printf(".data\n")
	for _, con := range pkg.stringLiterals {
		emitComment(0, "string literals\n")
		printf("%s:\n", con.sl.label)
		printf("  .string %s\n", con.sl.value)
	}

	for _, spec := range pkg.vars {
		var val ast.Expr
		if len(spec.Values) > 0 {
			val = spec.Values[0]
		}
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		} else {
			t = getTypeOfExpr(val)
		}
		if t == nil {
			panic("type cannot be nil for global variable: " + spec.Names[0].Name)
		}
		emitGlobalVariable(pkg, spec.Names[0], t, val)
	}
	printf("\n")
	printf(".text\n")
	printf(".global %s.__initGlobals\n", pkg.name)
	printf("%s.__initGlobals:\n", pkg.name)
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
	printf("  ret\n")

	for _, fnc := range pkg.funcs {
		emitFuncDecl(pkg.name, fnc)
	}

	emitDynamicTypes(typesMap)
	printf("\n")

	for _, file := range pkg.files {
		if strings.HasSuffix(file, ".s") {
			printf("# === static assembly %s ====\n", file)
			asmContents, err := os.ReadFile(file)
			if err != nil {
				panic(err)
			}
			printf("%s", string(asmContents))
		}
	}
}

func emitDynamicTypes(mapDtypes map[string]*dtypeEntry) {
	printf("# ------- Dynamic Types ------\n")
	printf(".data\n")

	sliceTypeMap := make([]string, len(mapDtypes)+1, len(mapDtypes)+1)

	// sort map in order to assure the deterministic results
	for key, ent := range mapDtypes {
		sliceTypeMap[ent.id] = key
	}

	// skip id=0
	for id := 1; id < len(sliceTypeMap); id++ {
		key := sliceTypeMap[id]
		ent := mapDtypes[key]

		printf("%s: # %s\n", ent.label, key)
		printf("  .quad %d\n", id)
		printf("  .quad .%s.S.dtype.%d\n", ent.pkgname, id)
		printf("  .quad %d\n", len(ent.serialized))
		printf(".%s.S.dtype.%d:\n", ent.pkgname, id)
		printf("  .string \"%s\"\n", ent.serialized)
	}
	printf("\n")
}

// --- type ---
type Type struct {
	E ast.Expr // original
}

type TypeKind string

const T_STRING TypeKind = "T_STRING"
const T_INTERFACE TypeKind = "T_INTERFACE"
const T_FUNC TypeKind = "T_FUNC"
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
const T_MAP TypeKind = "T_MAP"

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
			variable, isVariable := e.Obj.Data.(*Variable)
			if isVariable {
				return variable.Typ
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
		case ast.Fun:
			return e2t(e.Obj.Decl.(*ast.FuncDecl).Type)
		default:
			panic("Obj=" + e.Obj.Name + ", Kind=" +  e.Obj.Kind.String())
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
			elmType := getElementTypeOfCollectionType(listType)
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
		return getElementTypeOfCollectionType(getTypeOfExpr(list))
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
	}
	panic("bad type\n")
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
		case ast.Var:
			switch dcl := fn.Obj.Decl.(type) {
			case *ast.AssignStmt: // short var decl
				rhs := dcl.Rhs[0]
				switch r := rhs.(type) {
				case *ast.SelectorExpr:
					assert(isQI(r), "expect QI", __func__)
					ff := lookupForeignFunc(selector2QI(r))
					return fieldList2Types(ff.decl.Type.Results)
				}
			case *ast.ValueSpec: // var v func(T1)T2
				return fieldList2Types(dcl.Type.(*ast.FuncType).Results)
			default:
				throw(dcl)
			}
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
			decl := fn.Obj.Decl
			if decl == nil {
				panic("decl of function " + fn.Name + " should not nil")
			}
			switch dcl := decl.(type) {
			case *ast.FuncDecl:
				return fieldList2Types(dcl.Type.Results)
			default:
				throw(decl)
			}
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
			case gError:
				return "error"
			default:
				// named type
				decl := e.Obj.Decl
				typeSpec := decl.(*ast.TypeSpec)
				pkgName := typeSpec.Name.Obj.Data.(string)
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
	case *ast.MapType:
		return "map[" + serializeType(e2t(e.Key)) + "]" + serializeType(e2t(e.Value))
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
	case *ast.StructType, *ast.ArrayType, *ast.StarExpr, *ast.Ellipsis, *ast.MapType, *ast.InterfaceType:
		// type literal
		return t
	case *ast.Ident:
		assert(e.Obj.Kind == ast.Typ, "should be ast.Typ : "+e.Obj.Name, __func__)
		switch e.Obj {
		case gUintptr, gInt, gInt32, gString, gUint8, gUint16, gBool:
			return t
		case gError:
			return e2t(&ast.InterfaceType{
				Methods: nil, //  @FIXME
			})
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
	case *ast.FuncType:
		return t
	}
	throw(t.E)
	return nil
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
		case gError:
			return T_INTERFACE
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
	case *ast.MapType:
		return T_MAP
	case *ast.InterfaceType:
		return T_INTERFACE
	case *ast.FuncType:
		return T_FUNC
	}
	panic("should not reach here")
}

func isInterface(t *Type) bool {
	return kind(t) == T_INTERFACE
}

func getElementTypeOfCollectionType(t *Type) *Type {
	ut := getUnderlyingType(t)
	switch kind(ut) {
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
	case T_MAP:
		mapType := ut.E.(*ast.MapType)
		return e2t(mapType.Value)
	default:
		unexpectedKind(kind(t))
	}
	return nil
}

func getKeyTypeOfCollectionType(t *Type) *Type {
	ut := getUnderlyingType(t)
	switch kind(ut) {
	case T_SLICE, T_ARRAY, T_STRING:
		return tInt
	case T_MAP:
		mapType := ut.E.(*ast.MapType)
		return e2t(mapType.Key)
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
	case T_UINTPTR, T_POINTER, T_MAP:
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
	case T_FUNC:
		return SizeOfPtr
	default:
		unexpectedKind(kind(t))
	}
	return 0
}

func lookupStructField(structType *ast.StructType, selName string) *ast.Field {
	for _, field := range structType.Fields.List {
		if field.Names[0].Name == selName {
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

var currentFor *MetaForStmt
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
	method := &Method{
		PkgName:      pkgName,
		RcvNamedType: rcvNamedType,
		IsPtrMethod:  isPtr,
		Name:         funcDecl.Name.Name,
		FuncType:     funcDecl.Type,
	}
	return method
}

// https://golang.org/ref/spec#Method_sets
// @TODO map key should be a QI ?
var MethodSets = make(map[unsafe.Pointer]*NamedType)

type NamedType struct {
	methodSet map[string]*Method
}

func registerMethod(method *Method) {
	key := unsafe.Pointer(method.RcvNamedType.Obj)
	namedType, ok := MethodSets[key]
	if !ok {
		namedType = &NamedType{
			methodSet: make(map[string]*Method),
		}
		MethodSets[key] = namedType
	}
	namedType.methodSet[method.Name] = method
}

func lookupMethod(rcvT *Type, methodName *ast.Ident) *Method {
	rcvType := rcvT.E
	rcvPointerType, isPtr := rcvType.(*ast.StarExpr)
	if isPtr {
		rcvType = rcvPointerType.X
	}
	var typeObj *ast.Object
	switch typ := rcvType.(type) {
	case *ast.Ident:
		typeObj = typ.Obj
	case *ast.SelectorExpr:
		t := lookupForeignIdent(selector2QI(typ))
		typeObj = t.Obj
	default:
		panic(rcvType)
	}

	namedType, ok := MethodSets[unsafe.Pointer(typeObj)]
	if !ok {
		panic(typeObj.Name + " has no methodSet")
	}
	method, ok := namedType.methodSet[methodName.Name]
	if !ok {
		panic("method not found")
	}
	return method
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
		for _, v := range spec.Values {
			walkExpr(v)
		}
	}
}
func walkAssignStmt(s *ast.AssignStmt) {
	walkExpr(s.Lhs[0])
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
			case *ast.TypeAssertExpr: // v, ok := x.(T)
				typ0 := getTypeOfExpr(rhs0)
				types = []*Type{typ0, tBool}
			case *ast.IndexExpr: // v, ok := m[k]
				typ0 := getTypeOfExpr(rhs0)
				types = []*Type{typ0, tBool}
			default:
				throw(rhs0)
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
	setMetaReturnStmt(s, &MetaReturnStmt{
		Fnc: currentFunc,
	})
	for _, r := range s.Results {
		walkExpr(r)
	}
}
func walkIfStmt(s *ast.IfStmt) {
	if s.Init != nil {
		walkStmt(s.Init)
	}
	walkExpr(s.Cond)
	walkStmt(s.Body)
	if s.Else != nil {
		walkStmt(s.Else)
	}
}
func walkBlockStmt(s *ast.BlockStmt) {
	for _, stmt := range s.List {
		walkStmt(stmt)
	}
}

func walkForStmt(s *ast.ForStmt) {
	meta := &MetaForStmt{
		Outer: currentFor,
	}
	currentFor = meta
	setMetaForStmt(s, meta)
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
	meta := &MetaForStmt{
		Outer: currentFor,
	}
	currentFor = meta
	setMetaForStmt(s, meta)
	walkExpr(s.X)
	walkStmt(s.Body)

	collectionType := getUnderlyingType(getTypeOfExpr(s.X))
	keyType := getKeyTypeOfCollectionType(collectionType)
	elmType := getElementTypeOfCollectionType(collectionType)
	switch kind(collectionType) {
	case T_SLICE, T_ARRAY:
		meta.ForRange = &MetaForRange{
			IsMap:    false,
			LenVar:   registerLocalVariable(currentFunc, ".range.len", tInt),
			Indexvar: registerLocalVariable(currentFunc, ".range.index", tInt),
		}
	case T_MAP:
		meta.ForRange = &MetaForRange{
			IsMap:   true,
			MapVar:  registerLocalVariable(currentFunc, ".range.map", tUintptr),
			ItemVar: registerLocalVariable(currentFunc, ".range.item", tUintptr),
		}
	default:
		throw(collectionType)
	}
	if s.Tok.String() == ":=" {
		// declare local variables
		keyIdent := s.Key.(*ast.Ident)
		setVariable(keyIdent.Obj, registerLocalVariable(currentFunc, keyIdent.Name, keyType))

		valueIdent := s.Value.(*ast.Ident)
		setVariable(valueIdent.Obj, registerLocalVariable(currentFunc, valueIdent.Name, elmType))
	}
	currentFor = meta.Outer
}
func walkIncDecStmt(s *ast.IncDecStmt) {
	walkExpr(s.X)
}
func walkSwitchStmt(s *ast.SwitchStmt) {
	if s.Init != nil {
		walkStmt(s.Init)
	}
	if s.Tag != nil {
		walkExpr(s.Tag)
	}
	walkStmt(s.Body)
}
func walkTypeSwitchStmt(s *ast.TypeSwitchStmt) {
	typeSwitch := &MetaTypeSwitchStmt{}
	setMetaTypeSwitchStmt(s, typeSwitch)
	var assignIdent *ast.Ident
	switch assign := s.Assign.(type) {
	case *ast.ExprStmt:
		typeAssertExpr := assign.X.(*ast.TypeAssertExpr)
		typeSwitch.Subject = typeAssertExpr.X
		walkExpr(typeAssertExpr.X)
	case *ast.AssignStmt:
		lhs := assign.Lhs[0]
		assignIdent = lhs.(*ast.Ident)
		typeSwitch.AssignIdent = assignIdent
		// ident will be a new local variable in each case clause
		typeAssertExpr := assign.Rhs[0].(*ast.TypeAssertExpr)
		typeSwitch.Subject = typeAssertExpr.X
		walkExpr(typeAssertExpr.X)
	default:
		throw(s.Assign)
	}

	typeSwitch.SubjectVariable = registerLocalVariable(currentFunc, ".switch_expr", tEface)
	for _, _case := range s.Body.List {
		cc := _case.(*ast.CaseClause)
		tscc := &MetaTypeSwitchCaseClose{
			Orig: cc,
		}
		typeSwitch.Cases = append(typeSwitch.Cases, tscc)
		if assignIdent != nil && len(cc.List) > 0 {
			var varType *Type
			if isNil(cc.List[0]) {
				varType = getTypeOfExpr(typeSwitch.Subject)
			} else {
				varType = e2t(cc.List[0])
			}
			// inject a variable of that type
			vr := registerLocalVariable(currentFunc, assignIdent.Name, varType)
			tscc.Variable = vr
			tscc.VariableType = varType
			setVariable(assignIdent.Obj, vr)
		}

		for _, stmt := range cc.Body {
			walkStmt(stmt)
		}
		if assignIdent != nil {
			setVariable(assignIdent.Obj, nil)
		}
	}
}
func isNil(e ast.Expr) bool {
	ident, ok := e.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Obj == gNil
}

func walkCaseClause(s *ast.CaseClause) {
	for _, e := range s.List {
		walkExpr(e)
	}
	for _, stmt := range s.Body {
		walkStmt(stmt)
	}
}
func walkBranchStmt(s *ast.BranchStmt) {
	assert(currentFor != nil, "break or continue should be in for body", __func__)
	setMetaBranchStmt(s, &MetaBranchStmt{
		containerForStmt: currentFor,
	})
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

func walkIdent(e *ast.Ident) {
	// what to do ?
}
func walkSelectorExpr(e *ast.SelectorExpr) {
	walkExpr(e.X)
}
func walkCallExpr(e *ast.CallExpr) {
	walkExpr(e.Fun)
	// Replace __func__ ident by a string literal
	for i, arg := range e.Args {
		ident, ok := arg.(*ast.Ident)
		if ok {
			if ident.Name == "__func__" && ident.Obj.Kind == ast.Var {
				basicLit := &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"" + currentFunc.Name + "\"",
				}
				arg = basicLit
				e.Args[i] = arg
			}
		}
		walkExpr(arg)
	}
}
func walkParenExpr(e *ast.ParenExpr) {
	walkExpr(e.X)
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
	// first argument of builtin func
	// do nothing
}
func walkMapType(e *ast.MapType) {
	// first argument of builtin func
	// do nothing
}
func walkStarExpr(e *ast.StarExpr) {
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
	case *ast.MapType:
		walkMapType(e)
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

var ExportedQualifiedIdents = make(map[string]*ast.Ident)

func lookupForeignIdent(qi QualifiedIdent) *ast.Ident {
	ident, ok := ExportedQualifiedIdents[string(qi)]
	if !ok {
		panic(qi + " Not found in ExportedQualifiedIdents")
	}
	return ident
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

	// grouping declarations by type
	for _, decl := range pkg.Decls {
		switch dcl := decl.(type) {
		case *ast.GenDecl:
			specInterface := dcl.Specs[0]
			switch spec := specInterface.(type) {
			case *ast.TypeSpec:
				typeSpecs = append(typeSpecs, spec)
			case *ast.ValueSpec:
				nameIdent := spec.Names[0]
				switch nameIdent.Obj.Kind {
				case ast.Var:
					varSpecs = append(varSpecs, spec)
				case ast.Con:
					constSpecs = append(constSpecs, spec)
				default:
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
		typeSpec.Name.Obj.Data = pkg.name // package to which the type belongs to
		t := e2t(typeSpec.Type)
		switch kind(t) {
		case T_STRUCT:
			structType := getUnderlyingType(t)
			calcStructSizeAndSetFieldOffset(structType.E.(*ast.StructType))
		}
		ExportedQualifiedIdents[string(newQI(pkg.name, typeSpec.Name.Name))] = typeSpec.Name
	}

	// collect methods in advance
	for _, funcDecl := range funcDecls {
		if funcDecl.Recv == nil { // non-method function
			qi := newQI(pkg.name, funcDecl.Name.Name)
			ExportedQualifiedIdents[string(qi)] = funcDecl.Name
		} else { // is method
			if funcDecl.Body != nil {
				method := newMethod(pkg.name, funcDecl)
				registerMethod(method)
			}
		}
	}

	for _, constSpec := range constSpecs {
		for _, v := range constSpec.Values {
			walkExpr(v)
		}
	}

	for _, varSpec := range varSpecs {
		nameIdent := varSpec.Names[0]
		assert(nameIdent.Obj.Kind == ast.Var, "should be Var", __func__)
		if varSpec.Type == nil {
			// Infer type
			val := varSpec.Values[0]
			t := getTypeOfExpr(val)
			if t == nil {
				panic("variable type is not determined : " + nameIdent.Name)
			}
			varSpec.Type = t.E
		}
		variable := newGlobalVariable(pkg.name, nameIdent.Obj.Name, e2t(varSpec.Type))
		setVariable(nameIdent.Obj, variable)
		pkg.vars = append(pkg.vars, varSpec)
		ExportedQualifiedIdents[string(newQI(pkg.name, nameIdent.Name))] = nameIdent
		for _, v := range varSpec.Values {
			// mainly to collect string literals
			walkExpr(v)
		}
	}

	for _, funcDecl := range funcDecls {
		fnc := &Func{
			Name:      funcDecl.Name.Name,
			FuncType:  funcDecl.Type,
			Localarea: 0,
			Argsarea:  16, // return address + previous rbp
		}
		currentFunc = fnc
		logf("funcdef %s\n", funcDecl.Name.Name)

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
			obj := field.Names[0].Obj
			setVariable(obj, registerParamVariable(fnc, obj.Name, e2t(field.Type)))
		}

		for i, field := range resultFields {
			if len(field.Names) == 0 {
				// unnamed retval
				registerReturnVariable(fnc, ".r"+strconv.Itoa(i), e2t(field.Type))
			} else {
				panic("TBI: named return variable is not supported")
			}
		}

		if funcDecl.Body != nil {
			fnc.Stmts = funcDecl.Body.List
			for _, stmt := range fnc.Stmts {
				walkStmt(stmt)
			}

			if funcDecl.Recv != nil { // is Method
				fnc.Method = newMethod(pkg.name, funcDecl)
			}
			pkg.funcs = append(pkg.funcs, fnc)
		}
	}
}

// --- universe ---
var gNil = &ast.Object{
	Kind: ast.Con, // is nil a constant ?
	Name: "nil",
}

var eNil = &ast.Ident{
	Obj:  gNil,
	Name: "nil",
}

var eZeroInt = &ast.BasicLit{
	Value: "0",
	Kind:  token.INT,
}

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

var gUintptr = &ast.Object{
	Kind: ast.Typ,
	Name: "uintptr",
}
var gBool = &ast.Object{
	Kind: ast.Typ,
	Name: "bool",
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

var gError = &ast.Object{
	Kind: ast.Typ,
	Name: "error",
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
var gDelete = &ast.Object{
	Kind: ast.Fun,
	Name: "delete",
}

var tBool *Type = &Type{
	E: &ast.Ident{
		Name: "bool",
		Obj:  gBool,
	},
}

var tInt *Type = &Type{
	E: &ast.Ident{
		Name: "int",
		Obj:  gInt,
	},
}

// Rune
var tInt32 *Type = &Type{
	E: &ast.Ident{
		Name: "int32",
		Obj:  gInt32,
	},
}

var tUintptr *Type = &Type{
	E: &ast.Ident{
		Name: "uintptr",
		Obj:  gUintptr,
	},
}
var tUint8 *Type = &Type{
	E: &ast.Ident{
		Name: "uint8",
		Obj:  gUint8,
	},
}

var tUint16 *Type = &Type{
	E: &ast.Ident{
		Name: "uint16",
		Obj:  gUint16,
	},
}
var tString *Type = &Type{
	E: &ast.Ident{
		Name: "string",
		Obj:  gString,
	},
}

var tEface *Type = &Type{
	E: &ast.InterfaceType{},
}

var generalSlice ast.Expr = &ast.Ident{}

func createUniverse() *ast.Scope {
	universe := ast.NewScope(nil)
	objects := []*ast.Object{
		gNil,
		// constants
		gTrue, gFalse,
		// types
		gString, gUintptr, gBool, gInt, gUint8, gUint16, gError,
		// funcs
		gNew, gMake, gAppend, gLen, gCap, gPanic, gDelete,
	}
	for _, obj := range objects {
		universe.Insert(obj)
	}

	// setting aliases
	universe.Objects["byte"] = gUint8

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
	mapImports := make(map[string]bool)
	for _, imprt := range file.Imports {
		// unwrap double quote "..."
		rawValue := imprt.Path.Value
		pth := rawValue[1 : len(rawValue)-1]
		base := path.Base(pth)
		mapImports[base] = true
	}
	for _, ident := range file.Unresolved {
		// lookup imported package name
		_, ok := mapImports[ident.Name]
		if ok {
			ident.Obj = &ast.Object{
				Kind: ast.Pkg,
				Name: ident.Name,
			}
			logf("# resolved: %s\n", ident.Name)
		}
	}
}

// "some/dir" => []string{"a.go", "b.go"}
func findFilesInDir(dir string) []string {
	dirents := mylib.GetDirents(dir)
	var r []string
	for _, dirent := range dirents {
		if dirent == "." || dirent == ".." {
			continue
		}
		if dirent == "_.s" {
			continue
		}
		if strings.HasSuffix(dirent, ".go") || strings.HasSuffix(dirent, ".s") {
			r = append(r, dirent)
		}
	}
	return r
}

func isStdLib(pth string) bool {
	return !strings.Contains(pth, "/")
}

func getImportPathsFromFile(file string) []string {
	fset := &token.FileSet{}
	astFile0 := parseImports(fset, file)
	var paths []string
	for _, importSpec := range astFile0.Imports {
		rawValue := importSpec.Path.Value
		logf("import %s\n", rawValue)
		pth := rawValue[1 : len(rawValue)-1]
		paths = append(paths, pth)
	}
	return paths
}

func removeNode(tree DependencyTree, node string) {
	for _, paths := range tree {
		delete(paths, node)
	}

	delete(tree, node)
}

func getKeys(tree DependencyTree) []string {
	var keys []string
	for k, _ := range tree {
		keys = append(keys, k)
	}
	return keys
}

type DependencyTree map[string]map[string]bool

// Do topological sort
// In the result list, the independent (lowest level) packages come first.
func sortTopologically(tree DependencyTree) []string {
	var sorted []string
	for len(tree) > 0 {
		keys := getKeys(tree)
		mylib.SortStrings(keys)
		for _, _path := range keys {
			children, ok := tree[_path]
			if !ok {
				panic("not found in tree")
			}
			if len(children) == 0 {
				// collect leaf node
				sorted = append(sorted, _path)
				removeNode(tree, _path)
			}
		}
	}
	return sorted
}

func getPackageDir(importPath string) string {
	if isStdLib(importPath) {
		return prjSrcPath + "/" + importPath
	} else {
		return srcPath + "/" + importPath
	}
}

func collectDependency(tree DependencyTree, paths map[string]bool) {
	for pkgPath, _ := range paths {
		if pkgPath == "unsafe" || pkgPath == "runtime" {
			continue
		}
		packageDir := getPackageDir(pkgPath)
		fnames := findFilesInDir(packageDir)
		children := make(map[string]bool)
		for _, fname := range fnames {
			if !strings.HasSuffix(fname, ".go") {
				// skip ".s"
				continue
			}
			_paths := getImportPathsFromFile(packageDir + "/" + fname)
			for _, pth := range _paths {
				if pth == "unsafe" || pth == "runtime" {
					continue
				}
				children[pth] = true
			}
		}
		tree[pkgPath] = children
		collectDependency(tree, children)
	}
}

var srcPath string
var prjSrcPath string

func collectAllPackages(inputFiles []string) []string {
	directChildren := collectDirectDependents(inputFiles)
	tree := make(DependencyTree)
	collectDependency(tree, directChildren)
	sortedPaths := sortTopologically(tree)

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

func collectDirectDependents(inputFiles []string) map[string]bool {
	importPaths := make(map[string]bool)
	for _, inputFile := range inputFiles {
		logf("input file: \"%s\"\n", inputFile)
		logf("Parsing imports\n")
		paths := getImportPathsFromFile(inputFile)
		for _, pth := range paths {
			importPaths[pth] = true
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

func parseImports(fset *token.FileSet, filename string) *ast.File {
	f, err := parserParseFile(fset, filename, nil, parserImportsOnly)
	if err != nil {
		panic(filename + ":" + err.Error())
	}
	return f
}

func parseFile(fset *token.FileSet, filename string) *ast.File {
	f, err := parserParseFile(fset, filename, nil, 0)
	if err != nil {
		panic(err.Error())
	}
	return f
}

func buildPackage(_pkg *PkgContainer, universe *ast.Scope) {
	typesMap = make(map[string]*dtypeEntry)
	typeId = 1

	logf("Building package : %s\n", _pkg.path)
	fset := &token.FileSet{}
	pkgScope := ast.NewScope(universe)
	for _, file := range _pkg.files {
		if strings.HasSuffix(file, ".s") {
			continue
		}
		logf("Parsing file: %s\n", file)
		astFile := parseFile(fset, file)
		_pkg.name = astFile.Name.Name
		_pkg.astFiles = append(_pkg.astFiles, astFile)
		for name, obj := range astFile.Scope.Objects {
			pkgScope.Objects[name] = obj
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
	fmt.Printf("    %s version:  show version\n", ProgName)
	fmt.Printf("    %s [-DF] [-DG] filename\n", ProgName)
}

func main() {
	srcPath = os.Getenv("GOPATH") + "/src"
	prjSrcPath = srcPath + "/github.com/DQNEO/babygo/src"

	if len(os.Args) == 1 {
		showHelp()
		return
	}

	if os.Args[1] == "version" {
		fmt.Printf("babygo version 0.0.2  linux/amd64\n")
		return
	} else if os.Args[1] == "help" {
		showHelp()
		return
	} else if os.Args[1] == "panic" {
		panicVersion := strconv.Itoa(mylib.Sum(1, 1))
		panic("I am panic version " + panicVersion)
	}

	workdir := os.Getenv("WORKDIR")
	if workdir == "" {
		workdir = "/tmp"
	}
	initAsm, err := os.Create(workdir + "/a.s")
	if err != nil {
		panic(err)
	}
	fout = initAsm
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
			name:  path.Base(_path),
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
		if _pkg.name == "" {
			panic("empty pkg name")
		}
		pgkAsm, err := os.Create(fmt.Sprintf("%s/%s.s", workdir, _pkg.name))
		if err != nil {
			panic(err)
		}
		fout = pgkAsm
		buildPackage(_pkg, universe)
		pgkAsm.Close()
	}
	initAsm.Close()
}

func setVariable(obj *ast.Object, vr *Variable) {
	if vr == nil {
		obj.Data = nil
	} else {
		obj.Data = vr
	}
}

// --- AST meta data ---
var mapMeta = make(map[unsafe.Pointer]interface{})

type MetaReturnStmt struct {
	Fnc *Func
}

type MetaForRange struct {
	IsMap    bool
	LenVar   *Variable
	Indexvar *Variable
	MapVar   *Variable // map
	ItemVar  *Variable // map element
}

type MetaForStmt struct {
	LabelPost string // for continue
	LabelExit string // for break
	Outer     *MetaForStmt
	ForRange  *MetaForRange // for array or slice
}

type MetaBranchStmt struct {
	containerForStmt *MetaForStmt
}

type MetaTypeSwitchStmt struct {
	Subject         ast.Expr
	SubjectVariable *Variable
	AssignIdent     *ast.Ident
	Cases           []*MetaTypeSwitchCaseClose
}

type MetaTypeSwitchCaseClose struct {
	Variable     *Variable
	VariableType *Type
	Orig         *ast.CaseClause
}
type Func struct {
	Name      string
	Stmts     []ast.Stmt
	Localarea int
	Argsarea  int
	LocalVars []*Variable
	Params    []*Variable
	Retvars   []*Variable
	FuncType  *ast.FuncType
	Method    *Method
}
type Method struct {
	PkgName      string
	RcvNamedType *ast.Ident
	IsPtrMethod  bool
	Name         string
	FuncType     *ast.FuncType
}
type Variable struct {
	Name         string
	IsGlobal     bool
	GlobalSymbol string
	LocalOffset  int
	Typ          *Type
}

func getStructFieldOffset(field *ast.Field) int {
	return mapMeta[unsafe.Pointer(field)].(int)
}

func setStructFieldOffset(field *ast.Field, offset int) {
	mapMeta[unsafe.Pointer(field)] = offset
}

func getMetaReturnStmt(s *ast.ReturnStmt) *MetaReturnStmt {
	return mapMeta[unsafe.Pointer(s)].(*MetaReturnStmt)
}

func setMetaReturnStmt(s *ast.ReturnStmt, meta *MetaReturnStmt) {
	mapMeta[unsafe.Pointer(s)] = meta
}

func getMetaForStmt(stmt ast.Stmt) *MetaForStmt {
	switch s := stmt.(type) {
	case *ast.ForStmt:
		return mapMeta[unsafe.Pointer(s)].(*MetaForStmt)
	case *ast.RangeStmt:
		return mapMeta[unsafe.Pointer(s)].(*MetaForStmt)
	default:
		panic(stmt)
	}
}

func setMetaForStmt(stmt ast.Stmt, meta *MetaForStmt) {
	switch s := stmt.(type) {
	case *ast.ForStmt:
		mapMeta[unsafe.Pointer(s)] = meta
	case *ast.RangeStmt:
		mapMeta[unsafe.Pointer(s)] = meta
	default:
		panic(stmt)
	}
}

func getMetaBranchStmt(s *ast.BranchStmt) *MetaBranchStmt {
	return mapMeta[unsafe.Pointer(s)].(*MetaBranchStmt)
}

func setMetaBranchStmt(s *ast.BranchStmt, meta *MetaBranchStmt) {
	mapMeta[unsafe.Pointer(s)] = meta
}

func getMetaTypeSwitchStmt(s *ast.TypeSwitchStmt) *MetaTypeSwitchStmt {
	return mapMeta[unsafe.Pointer(s)].(*MetaTypeSwitchStmt)
}

func setMetaTypeSwitchStmt(s *ast.TypeSwitchStmt, meta *MetaTypeSwitchStmt) {
	mapMeta[unsafe.Pointer(s)] = meta
}

func throw(x interface{}) {
	panic(fmt.Sprintf("%T", x))
}
