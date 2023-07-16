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
	//gofmt "fmt"
)

const ThrowFormat string = "%T"

const Version string = "0.0.7"

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

func logf2(format string, a ...interface{}) {
	f := "# " + format
	fmt.Fprintf(os.Stderr, f, a...)
}

var debugFrontEnd bool

func logf(format string, a ...interface{}) {
	if !debugFrontEnd {
		return
	}
	f := "# " + format
	fmt.Fprintf(os.Stderr, f, a...)
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

func emitConstInt(expr ast.Expr) {
	i := evalInt(expr)
	printf("  pushq $%d # const number literal\n", i)
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

func emitListHeadAddr(list MetaExpr) {
	t := getTypeOfExprMeta(list)
	switch kind(t) {
	case T_ARRAY:
		emitAddr(list) // array head
	case T_SLICE:
		emitExpr(list)
		emitPopSlice()
		printf("  pushq %%rax # slice.ptr\n")
	case T_STRING:
		emitExpr(list)
		emitPopString()
		printf("  pushq %%rax # string.ptr\n")
	default:
		unexpectedKind(kind(t))
	}
}

func emitAddr(expr MetaExpr) {
	emitComment(2, "[emitAddr] %T\n", expr)
	switch m := expr.(type) {
	case *MetaIdent:
		switch m.kind {
		case "var":
			vr := m.e.Obj.Data.(*Variable)
			emitVariableAddr(vr)
		case "fun":
			qi := newQI(currentPkg.name, m.e.Obj.Name)
			emitFuncAddr(qi)
		default:
			panic("Unexpected kind")
		}
	case *MetaIndexExpr:
		if kind(getTypeOfExprMeta(m.X)) == T_MAP {
			emitAddrForMapSet(m)
		} else {
			elmType := getTypeOfExprMeta(m)
			emitExpr(m.Index) // index number
			emitListElementAddr(m.X, elmType)
		}
	case *MetaStarExpr:
		emitExpr(m.X)
	case *MetaSelectorExpr:
		if isQI(m.e) { // pkg.Var|pkg.Const
			qi := selector2QI(m.e)
			ident := lookupForeignIdent(qi)
			switch ident.Obj.Kind {
			case ast.Var:
				printf("  leaq %s(%%rip), %%rax # external global variable \n", string(qi))
				printf("  pushq %%rax # variable address\n")
			case ast.Fun:
				emitFuncAddr(qi)
			case ast.Con:
				panic("TBI")
			default:
				panic("Unexpected foreign ident kind:" + ident.Obj.Kind.String())
			}
		} else { // (e).field
			typeOfX := getUnderlyingType(getTypeOfExprMeta(m.X))
			var structTypeLiteral *ast.StructType
			switch typ := typeOfX.E.(type) {
			case *ast.StructType: // strct.field
				structTypeLiteral = typ
				emitAddr(m.X)
			case *ast.StarExpr: // ptr.field
				structTypeLiteral = getUnderlyingStructType(e2t(typ.X))
				emitExpr(m.X)
			default:
				unexpectedKind(kind(typeOfX))
			}

			field := lookupStructField(structTypeLiteral, m.e.Sel.Name)
			offset := getStructFieldOffset(field)
			emitAddConst(offset, "struct head address + struct.field offset")
		}
	case *MetaCompositLiteral:
		knd := kind(getTypeOfExprMeta(m))
		switch knd {
		case T_STRUCT:
			// result of evaluation of a struct literal is its address
			emitExpr(m)
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
func emitConversion(toType *Type, arg0 MetaExpr) {
	emitComment(2, "[emitConversion]\n")
	switch to := toType.E.(type) {
	case *ast.Ident:
		switch to.Obj {
		case gString: // string(e)
			switch kind(getTypeOfExprMeta(arg0)) {
			case T_SLICE: // string(slice)
				emitExpr(arg0) // slice
				emitPopSlice()
				printf("  pushq %%rcx # str len\n")
				printf("  pushq %%rax # str ptr\n")
			case T_STRING: // string(string)
				emitExpr(arg0)
			default:
				unexpectedKind(kind(getTypeOfExprMeta(arg0)))
			}
		case gInt, gUint8, gUint16, gUintptr: // int(e)
			emitExpr(arg0)
		default:
			if to.Obj.Kind == ast.Typ {
				emitExpr(arg0)
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
		assert(kind(getTypeOfExprMeta(arg0)) == T_STRING, "source type should be slice", __func__)
		emitComment(2, "Conversion of string => slice \n")
		emitExpr(arg0)
		emitPopString()
		printf("  pushq %%rcx # cap\n")
		printf("  pushq %%rcx # len\n")
		printf("  pushq %%rax # ptr\n")
	case *ast.ParenExpr: // (T)(arg0)
		emitConversion(e2t(to.X), arg0)
	case *ast.StarExpr: // (*T)(arg0)
		emitExpr(arg0)
	case *ast.InterfaceType:
		emitExpr(arg0)
		if isInterface(getTypeOfExprMeta(arg0)) {
			// do nothing
		} else {
			// Convert dynamic value to interface
			emitConvertToInterface(getTypeOfExprMeta(arg0))
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
	case T_ARRAY:
		size := getSizeOfType(t)
		emitComment(2, "zero value of an array. size=%d (allocating on heap)\n", size)
		emitCallMalloc(size)
	case T_STRUCT:
		structSize := getSizeOfType(t)
		emitComment(2, "zero value of a struct. size=%d (allocating on heap)\n", structSize)
		emitCallMalloc(structSize)
	default:
		unexpectedKind(kind(t))
	}
}

func emitLen(arg MetaExpr) {
	switch kind(getTypeOfExprMeta(arg)) {
	case T_ARRAY:
		arrayType := getTypeOfExprMeta(arg).E.(*ast.ArrayType)
		emitConstInt(arrayType.Len)
	case T_SLICE:
		emitExpr(arg)
		emitPopSlice()
		printf("  pushq %%rcx # len\n")
	case T_STRING:
		emitExpr(arg)
		emitPopString()
		printf("  pushq %%rcx # len\n")
	case T_MAP:
		args := []*Arg{
			// len
			&Arg{
				meta:      arg,
				paramType: getTypeOfExprMeta(arg),
			},
		}
		resultList := &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Type: tInt.E,
				},
			},
		}
		emitCallDirect("runtime.lenMap", args, resultList)

	default:
		unexpectedKind(kind(getTypeOfExprMeta(arg)))
	}
}

func emitCap(arg MetaExpr) {
	switch kind(getTypeOfExprMeta(arg)) {
	case T_ARRAY:
		arrayType := getTypeOfExprMeta(arg).E.(*ast.ArrayType)
		emitConstInt(arrayType.Len)
	case T_SLICE:
		emitExpr(arg)
		emitPopSlice()
		printf("  pushq %%rdx # cap\n")
	case T_STRING:
		panic("cap() cannot accept string type")
	default:
		unexpectedKind(kind(getTypeOfExprMeta(arg)))
	}
}

func emitCallMalloc(size int) {
	// call malloc and return pointer
	ff := lookupForeignFunc(newQI("runtime", "malloc"))
	emitAllocReturnVarsAreaFF(ff)
	printf("  pushq $%d\n", size)
	emitCallFF(ff)
}

type MetaStructLiteralElement struct {
	field     *ast.Field
	fieldType *Type
	ValueMeta MetaExpr
}

func emitStructLiteral(meta *MetaCompositLiteral) {
	// allocate heap area with zero value
	emitComment(2, "emitStructLiteral\n")
	structType := meta.typ
	emitZeroValue(structType) // push address of the new storage
	metaElms := meta.strctEements
	for _, metaElm := range metaElms {
		// push lhs address
		emitPushStackTop(tUintptr, 0, "address of struct heaad")

		fieldOffset := getStructFieldOffset(metaElm.field)
		emitAddConst(fieldOffset, "address of struct field")

		// push rhs value
		emitExpr(metaElm.ValueMeta)
		mayEmitConvertTooIfc(metaElm.ValueMeta, metaElm.fieldType)

		// assign
		emitStore(metaElm.fieldType, true, false)
	}
}

func emitArrayLiteral(meta *MetaCompositLiteral) {
	elmType := meta.elmType
	elmSize := getSizeOfType(elmType)
	memSize := elmSize * meta.len

	emitCallMalloc(memSize) // push
	for i, elm := range meta.metaElms {
		// push lhs address
		emitPushStackTop(tUintptr, 0, "malloced address")
		emitAddConst(elmSize*i, "malloced address + elmSize * index")
		// push rhs value
		emitExpr(elm)
		mayEmitConvertTooIfc(elm, elmType)

		// assign
		emitStore(elmType, true, false)
	}
}

func emitInvertBoolValue() {
	emitPopBool("")
	printf("  xorq $1, %%rax\n")
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
	meta      MetaExpr
	paramType *Type // expected type
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
			// walk of eArg will be done later in walkCompositeLit
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
		paramType := e2t(elp)
		iNil := &ast.Ident{
			Obj:  gNil,
			Name: "nil",
		}
		//		exprTypeMeta[unsafe.Pointer(iNil)] = e2t(elp)
		args = append(args, &Arg{
			e:         iNil,
			paramType: paramType,
		})
	}

	for _, arg := range args {
		ctx := &evalContext{_type: arg.paramType}
		arg.meta = walkExpr(arg.e, ctx)
	}

	if receiver != nil { // method call
		meta := walkExpr(receiver, nil)
		var receiverAndArgs []*Arg = []*Arg{
			&Arg{
				e:         receiver,
				paramType: getTypeOfExprMeta(meta),
				meta:      meta,
			},
		}
		for _, arg := range args {
			receiverAndArgs = append(receiverAndArgs, arg)
		}
		return receiverAndArgs
	}

	return args
}

func emitCallDirect(symbol string, args []*Arg, resultList *ast.FieldList) {
	emitCall(NewFuncValueFromSymbol(symbol), args, resultList)
}

// see "ABI of stack layout" in the emitFuncall comment
func emitCall(fv *FuncValue, args []*Arg, resultList *ast.FieldList) {
	emitComment(2, "emitCall len(args)=%d\n", len(args))

	var totalParamSize int
	var offsets []int
	for _, arg := range args {
		offsets = append(offsets, totalParamSize)
		totalParamSize += getSizeOfType(arg.paramType)
	}

	emitAllocReturnVarsArea(getTotalFieldsSize(resultList))
	printf("  subq $%d, %%rsp # alloc parameters area\n", totalParamSize)
	for i, arg := range args {
		paramType := arg.paramType
		if arg.meta == nil {
			panic("arg.meta should not be nil")
		}
		emitExpr(arg.meta)
		mayEmitConvertTooIfc(arg.meta, paramType)
		emitPop(kind(paramType))
		printf("  leaq %d(%%rsp), %%rsi # place to save\n", offsets[i])
		printf("  pushq %%rsi # place to save\n")
		emitRegiToMem(paramType)
	}

	emitCallQ(fv, totalParamSize, resultList)
}

func emitAllocReturnVarsAreaFF(ff *ForeignFunc) {
	emitAllocReturnVarsArea(getTotalFieldsSize(ff.funcType.Results))
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
	totalParamSize := getTotalFieldsSize(ff.funcType.Params)
	emitCallQ(NewFuncValueFromSymbol(ff.symbol), totalParamSize, ff.funcType.Results)
}

func NewFuncValueFromSymbol(symbol string) *FuncValue {
	return &FuncValue{
		isDirect: true,
		symbol:   symbol,
	}
}

type FuncValue struct {
	isDirect bool     // direct or indirect
	symbol   string   // for direct call
	expr     MetaExpr // for indirect call
}

func emitCallQ(fv *FuncValue, totalParamSize int, resultList *ast.FieldList) {
	if fv.isDirect {
		if fv.symbol == "" {
			panic("callq target must not be empty")
		}
		printf("  callq %s\n", fv.symbol)
	} else {
		emitExpr(fv.expr)
		printf("  popq %%rax\n")
		printf("  callq *%%rax\n")
	}

	emitFreeParametersArea(totalParamSize)
	printf("  #  totalReturnSize=%d\n", getTotalFieldsSize(resultList))
	emitFreeAndPushReturnedValue(resultList)
}

// callee
func emitReturnStmt(meta *MetaReturnStmt) {
	funcDef := meta.Fnc
	if len(funcDef.Retvars) != len(meta.MetaResults) {
		panic("length of return and func type do not match")
	}

	_len := len(meta.MetaResults)
	for i := 0; i < _len; i++ {
		emitAssignToVar(funcDef.Retvars[i], meta.MetaResults[i])
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

func emitBuiltinFunCall(obj *ast.Object, typeArg0 *Type, arg0 MetaExpr, arg1 MetaExpr, arg2 MetaExpr) {
	switch obj {
	case gLen:
		emitLen(arg0)
		return
	case gCap:
		emitCap(arg0)
		return
	case gNew:
		// size to malloc
		size := getSizeOfType(typeArg0)
		emitCallMalloc(size)
		return
	case gMake:
		typeArg := typeArg0
		switch kind(typeArg) {
		case T_MAP:
			mapType := getUnderlyingType(typeArg).E.(*ast.MapType)
			valueSize := newNumberLiteral(getSizeOfType(e2t(mapType.Value)))
			// A new, empty map value is made using the built-in function make,
			// which takes the map type and an optional capacity hint as arguments:
			length := newNumberLiteral(0)
			args := []*Arg{
				&Arg{
					meta:      length,
					paramType: tUintptr,
				},
				&Arg{
					meta:      valueSize,
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
			emitCallDirect("runtime.makeMap", args, resultList)
			return
		case T_SLICE:
			// make([]T, ...)
			arrayType := getUnderlyingType(typeArg).E.(*ast.ArrayType)
			elmSize := getSizeOfType(e2t(arrayType.Elt))
			numlit := newNumberLiteral(elmSize)
			args := []*Arg{
				// elmSize
				&Arg{
					meta:      numlit,
					paramType: tInt,
				},
				// len
				&Arg{
					meta:      arg1,
					paramType: tInt,
				},
				// cap
				&Arg{
					meta:      arg2,
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
			emitCallDirect("runtime.makeSlice", args, resultList)
			return
		default:
			throw(typeArg)
		}
	case gAppend:
		sliceArg := arg0
		elemArg := arg1
		elmType := getElementTypeOfCollectionType(getTypeOfExprMeta(sliceArg))
		elmSize := getSizeOfType(elmType)
		args := []*Arg{
			// slice
			&Arg{
				meta:      sliceArg,
				paramType: e2t(generalSlice),
			},
			// elm
			&Arg{
				meta:      elemArg,
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
		emitCallDirect(symbol, args, resultList)
		return
	case gPanic:
		funcVal := "runtime.panic"
		_args := []*Arg{&Arg{
			meta:      arg0,
			paramType: tEface,
		}}
		emitCallDirect(funcVal, _args, nil)
		return
	case gDelete:
		funcVal := "runtime.deleteMap"
		_args := []*Arg{
			&Arg{
				meta:      arg0,
				paramType: getTypeOfExprMeta(arg0),
			},
			&Arg{
				meta:      arg1,
				paramType: tEface,
			},
		}
		emitCallDirect(funcVal, _args, nil)
		return
	}

}

// ABI of stack layout in function call
//
// string:
//
//	str.ptr
//	str.len
//
// slice:
//
//	slc.ptr
//	slc.len
//	slc.cap
//
// # ABI of function call
//
// call f(i1 int, i2 int) (r1 int, r2 int)
//
//	-- stack top
//	i1
//	i2
//	r1
//	r2
//
// call f(i int, s string, slc []T) int
//
//	-- stack top
//	i
//	s.ptr
//	s.len
//	slc.ptr
//	slc.len
//	slc.cap
//	r
//	--
func emitFuncall(meta *MetaCallExpr) {

	emitComment(2, "[emitFuncall] %T(...)\n", meta.fun)
	// check if it's a builtin func
	if meta.builtin != nil {
		emitBuiltinFunCall(meta.builtin, meta.typeArg0, meta.arg0, meta.arg1, meta.arg2)
		return
	}

	//	args := prepareArgs(meta.funcType, meta.receiver, meta.args, meta.hasEllipsis)
	emitCall(meta.funcVal, meta.metaArgs, meta.funcType.Results)
}

type evalContext struct {
	okContext bool
	_type     *Type
}

// 1 value
func emitIdent(meta *MetaIdent) {
	e := meta.e
	//logf2("emitIdent ident=%s\n", e.Name)
	assert(e.Obj != nil, " ident.Obj should not be nil:"+e.Name, __func__)
	switch e.Obj {
	case gTrue: // true constant
		emitTrue()
	case gFalse: // false constant
		emitFalse()
	case gNil: // zero value
		metaType := meta.typ
		if metaType == nil {
			//gofmt.Fprintf(os.Stderr, "exprTypeMeta=%v\n", exprTypeMeta)
			panic("untyped nil is not allowed. Probably the type is not set in walk phase. pkg=" + currentPkg.name)
		}
		// emit zero value of the type
		switch kind(metaType) {
		case T_SLICE, T_POINTER, T_INTERFACE, T_MAP:
			emitZeroValue(metaType)
		default:
			unexpectedKind(kind(metaType))
		}
	default:
		switch e.Obj.Kind {
		case ast.Var:
			emitAddr(meta)
			emitLoadAndPush(getTypeOfExprMeta(meta))
		case ast.Con:
			emitExpr(meta.conLiteral)
		case ast.Fun:
			emitAddr(meta)
		default:
			panic("Unexpected ident kind:" + e.Obj.Kind.String() + " name=" + e.Name)
		}
	}
}

// 1 or 2 values
func emitIndexExpr(meta *MetaIndexExpr) {
	if meta.IsMap {
		emitMapGet(meta, meta.NeedsOK)
	} else {
		emitAddr(meta)
		emitLoadAndPush(getTypeOfExprMeta(meta))
	}
}

// 1 value
func emitStarExpr(meta *MetaStarExpr) {
	emitAddr(meta)
	emitLoadAndPush(getTypeOfExprMeta(meta))
}

// 1 value X.Sel
func emitSelectorExpr(meta *MetaSelectorExpr) {
	e := meta.e
	// pkg.Ident or strct.field
	if isQI(e) {
		qi := selector2QI(e)
		ident := lookupForeignIdent(qi)
		switch ident.Obj.Kind {
		case ast.Fun:
			emitFuncAddr(qi)
		case ast.Var:
			m := walkIdent(ident, nil)
			emitExpr(m)
		case ast.Con:
			m := walkIdent(ident, nil)
			emitExpr(m)
		}
	} else {
		// strct.field
		emitAddr(meta)
		emitLoadAndPush(getTypeOfExprMeta(meta))
	}
}

// multi values Fun(Args)
func emitCallExpr(meta *MetaCallExpr) {
	// check if it's a conversion
	if meta.isConversion {
		emitComment(2, "[emitCallExpr] Conversion\n")
		emitConversion(meta.toType, meta.arg0)
	} else {
		emitComment(2, "[emitCallExpr] Funcall\n")
		emitFuncall(meta)
	}
}

// 1 value
func emitBasicLit(mt *MetaBasicLit) {
	switch mt.Kind {
	case "CHAR":
		printf("  pushq $%d # convert char literal to int\n", mt.charValue)
	case "INT":
		printf("  pushq $%d # number literal\n", mt.intValue)
	case "STRING":
		sl := mt.stringLiteral
		if sl.strlen == 0 {
			// zero value
			emitZeroValue(tString)
		} else {
			printf("  pushq $%d # str len\n", sl.strlen)
			printf("  leaq %s(%%rip), %%rax # str ptr\n", sl.label)
			printf("  pushq %%rax # str ptr\n")
		}
	default:
		panic("Unexpected literal kind:" + mt.Kind)
	}
}

// 1 value
func emitUnaryExpr(meta *MetaUnaryExpr) {
	e := meta.e
	switch e.Op.String() {
	case "+":
		emitExpr(meta.X)
	case "-":
		emitExpr(meta.X)
		printf("  popq %%rax # e.X\n")
		printf("  imulq $-1, %%rax\n")
		printf("  pushq %%rax\n")
	case "&":
		emitAddr(meta.X)
	case "!":
		emitExpr(meta.X)
		emitInvertBoolValue()
	default:
		throw(e.Op)
	}
}

// 1 value
func emitBinaryExpr(meta *MetaBinaryExpr) {
	e := meta.e
	switch e.Op.String() {
	case "&&":
		labelid++
		labelExitWithFalse := fmt.Sprintf(".L.%d.false", labelid)
		labelExit := fmt.Sprintf(".L.%d.exit", labelid)
		emitExpr(meta.X) // left
		emitPopBool("left")
		printf("  cmpq $1, %%rax\n")
		// exit with false if left is false
		printf("  jne %s\n", labelExitWithFalse)

		// if left is true, then eval right and exit
		emitExpr(meta.Y) // right
		printf("  jmp %s\n", labelExit)

		printf("  %s:\n", labelExitWithFalse)
		emitFalse()
		printf("  %s:\n", labelExit)
	case "||":
		labelid++
		labelExitWithTrue := fmt.Sprintf(".L.%d.true", labelid)
		labelExit := fmt.Sprintf(".L.%d.exit", labelid)
		emitExpr(meta.X) // left
		emitPopBool("left")
		printf("  cmpq $1, %%rax\n")
		// exit with true if left is true
		printf("  je %s\n", labelExitWithTrue)

		// if left is false, then eval right and exit
		emitExpr(meta.Y) // right
		printf("  jmp %s\n", labelExit)

		printf("  %s:\n", labelExitWithTrue)
		emitTrue()
		printf("  %s:\n", labelExit)
	case "+":
		if kind(getTypeOfExprMeta(meta.X)) == T_STRING {
			emitCatStrings(meta.X, meta.Y)
		} else {
			emitExpr(meta.X) // left
			emitExpr(meta.Y) // right
			printf("  popq %%rcx # right\n")
			printf("  popq %%rax # left\n")
			printf("  addq %%rcx, %%rax\n")
			printf("  pushq %%rax\n")
		}
	case "-":
		emitExpr(meta.X) // left
		emitExpr(meta.Y) // right
		printf("  popq %%rcx # right\n")
		printf("  popq %%rax # left\n")
		printf("  subq %%rcx, %%rax\n")
		printf("  pushq %%rax\n")
	case "*":
		emitExpr(meta.X) // left
		emitExpr(meta.Y) // right
		printf("  popq %%rcx # right\n")
		printf("  popq %%rax # left\n")
		printf("  imulq %%rcx, %%rax\n")
		printf("  pushq %%rax\n")
	case "%":
		emitExpr(meta.X) // left
		emitExpr(meta.Y) // right
		printf("  popq %%rcx # right\n")
		printf("  popq %%rax # left\n")
		printf("  movq $0, %%rdx # init %%rdx\n")
		printf("  divq %%rcx\n")
		printf("  movq %%rdx, %%rax\n")
		printf("  pushq %%rax\n")
	case "/":
		emitExpr(meta.X) // left
		emitExpr(meta.Y) // right
		printf("  popq %%rcx # right\n")
		printf("  popq %%rax # left\n")
		printf("  movq $0, %%rdx # init %%rdx\n")
		printf("  divq %%rcx\n")
		printf("  pushq %%rax\n")
	case "==":
		emitBinaryExprComparison(meta.X, meta.Y)
	case "!=":
		emitBinaryExprComparison(meta.X, meta.Y)
		emitInvertBoolValue()
	case "<":
		emitExpr(meta.X) // left
		emitExpr(meta.Y) // right
		emitCompExpr("setl")
	case "<=":
		emitExpr(meta.X) // left
		emitExpr(meta.Y) // right
		emitCompExpr("setle")
	case ">":
		emitExpr(meta.X) // left
		emitExpr(meta.Y) // right
		emitCompExpr("setg")
	case ">=":
		emitExpr(meta.X) // left
		emitExpr(meta.Y) // right
		emitCompExpr("setge")
	case "|":
		emitExpr(meta.X) // left
		emitExpr(meta.Y) // right
		emitBitWiseOr()
	case "&":
		emitExpr(meta.X) // left
		emitExpr(meta.Y) // right
		emitBitWiseAnd()
	default:
		panic(e.Op.String())
	}
}

// 1 value
func emitCompositeLit(meta *MetaCompositLiteral) {
	// slice , array, map or struct
	switch meta.kind {
	case "struct":
		emitStructLiteral(meta)
	case "array":
		emitArrayLiteral(meta)
	case "slice":
		emitArrayLiteral(meta)
		emitPopAddress("malloc")
		printf("  pushq $%d # slice.cap\n", meta.len)
		printf("  pushq $%d # slice.len\n", meta.len)
		printf("  pushq %%rax # slice.ptr\n")
	default:
		panic(meta.kind)
	}
}

// 1 value list[low:high]
func emitSliceExpr(meta *MetaSliceExpr) {
	list := meta.X
	listType := getTypeOfExprMeta(list)

	switch kind(listType) {
	case T_SLICE, T_ARRAY:
		if meta.Max == nil {
			// new cap = cap(operand) - low
			emitCap(meta.X)
			emitExpr(meta.Low)
			printf("  popq %%rcx # low\n")
			printf("  popq %%rax # orig_cap\n")
			printf("  subq %%rcx, %%rax # orig_cap - low\n")
			printf("  pushq %%rax # new cap\n")

			// new len = high - low
			if meta.High != nil {
				emitExpr(meta.High)
			} else {
				// high = len(orig)
				emitLen(meta.X)
			}
			emitExpr(meta.Low)
			printf("  popq %%rcx # low\n")
			printf("  popq %%rax # high\n")
			printf("  subq %%rcx, %%rax # high - low\n")
			printf("  pushq %%rax # new len\n")
		} else {
			// new cap = max - low
			emitExpr(meta.Max)
			emitExpr(meta.Low)
			printf("  popq %%rcx # low\n")
			printf("  popq %%rax # max\n")
			printf("  subq %%rcx, %%rax # new cap = max - low\n")
			printf("  pushq %%rax # new cap\n")
			// new len = high - low
			emitExpr(meta.High)
			emitExpr(meta.Low)
			printf("  popq %%rcx # low\n")
			printf("  popq %%rax # high\n")
			printf("  subq %%rcx, %%rax # new len = high - low\n")
			printf("  pushq %%rax # new len\n")
		}
	case T_STRING:
		// new len = high - low
		if meta.High != nil {
			emitExpr(meta.High)
		} else {
			// high = len(orig)
			emitLen(meta.X)
		}
		emitExpr(meta.Low)
		printf("  popq %%rcx # low\n")
		printf("  popq %%rax # high\n")
		printf("  subq %%rcx, %%rax # high - low\n")
		printf("  pushq %%rax # len\n")
		// no cap
	default:
		unexpectedKind(kind(listType))
	}

	emitExpr(meta.Low) // index number
	elmType := getElementTypeOfCollectionType(listType)
	emitListElementAddr(list, elmType)
}

// 1 or 2 values
func emitMapGet(m *MetaIndexExpr, okContext bool) {
	valueType := getTypeOfExprMeta(m)

	emitComment(2, "MAP GET for map[string]string\n")
	// emit addr of map element
	mp := m.X
	key := m.Index

	args := []*Arg{
		&Arg{
			meta:      mp,
			paramType: tUintptr,
		},
		&Arg{
			meta:      key,
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
	emitCallDirect("runtime.getAddrForMapGet", args, resultList)
	// return values = [ptr, bool(stack top)]
	emitPopBool("map get:  ok value")
	printf("  cmpq $1, %%rax\n")
	labelid++
	labelEnd := fmt.Sprintf(".L.end_map_get.%d", labelid)
	labelElse := fmt.Sprintf(".L.not_found.%d", labelid)
	printf("  jne %s # jmp if false\n", labelElse)

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
func emitTypeAssertExpr(meta *MetaTypeAssertExpr) {
	okContext := meta.NeedsOK
	e := meta.e
	emitExpr(meta.X)
	emitDtypeLabelAddr(e2t(e.Type))
	emitCompareDtypes()

	emitPopBool("type assertion ok value")
	printf("  cmpq $1, %%rax\n")

	labelid++
	labelEnd := fmt.Sprintf(".L.end_type_assertion.%d", labelid)
	labelElse := fmt.Sprintf(".L.unmatch.%d", labelid)
	printf("  jne %s # jmp if false\n", labelElse)

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

func isUniverseNil(meta MetaExpr) bool {
	m, ok := meta.(*MetaIdent)
	if !ok {
		return false
	}
	return m.isUniverseNil()
}

func emitExpr(meta MetaExpr) {
	emitComment(2, "[emitExpr] meta=%T\n", meta)
	switch m := meta.(type) {
	case *MetaBasicLit:
		emitBasicLit(m)
	case *MetaCompositLiteral:
		emitCompositeLit(m)
	case *MetaIdent:
		emitIdent(m)
	case *MetaSelectorExpr:
		emitSelectorExpr(m)
	case *MetaCallExpr:
		emitCallExpr(m) // can be tuple
	case *MetaIndexExpr:
		emitIndexExpr(m) // can be tuple
	case *MetaSliceExpr:
		emitSliceExpr(m)
	case *MetaStarExpr:
		emitStarExpr(m)
	case *MetaUnaryExpr:
		emitUnaryExpr(m)
	case *MetaBinaryExpr:
		emitBinaryExpr(m)
	case *MetaTypeAssertExpr:
		emitTypeAssertExpr(m) // can be tuple
	default:
		panic(fmt.Sprintf("meta type:%T", meta))
	}
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

func mayEmitConvertTooIfc(meta MetaExpr, ctxType *Type) {
	if !isUniverseNil(meta) && ctxType != nil && isInterface(ctxType) && !isInterface(getTypeOfExprMeta(meta)) {
		emitConvertToInterface(getTypeOfExprMeta(meta))
	}
}

type dtypeEntry struct {
	id         int
	serialized string
	label      string
}

var typeId int
var typesMap map[string]*dtypeEntry

// "**[1][]*int" => "dtype.8"
func getDtypeLabel(serializedType string) string {
	s := serializedType
	ent, ok := typesMap[s]
	if ok {
		return ent.label
	} else {
		id := typeId
		ent = &dtypeEntry{
			id:         id,
			serialized: serializedType,
			label:      "." + "dtype." + strconv.Itoa(id),
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

	printf("  %s:\n", labelTrue)
	printf("  pushq $1\n")
	printf("  jmp %s # jump to end\n", labelEnd)

	printf("  %s:\n", labelFalse)
	printf("  pushq $0\n")
	printf("  jmp %s # jump to end\n", labelEnd)

	printf("  %s:\n", labelCmp)
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
	printf("  %s:\n", labelEnd)
}

func emitDtypeLabelAddr(t *Type) {
	serializedType := serializeType(t)
	dtypeLabel := getDtypeLabel(serializedType)
	printf("  leaq %s(%%rip), %%rax # dtype label address \"%s\"\n", dtypeLabel, serializedType)
	printf("  pushq %%rax           # dtype label address\n")
}

func newNumberLiteral(x int) *MetaBasicLit {
	e := &ast.BasicLit{
		Kind:  token.INT,
		Value: strconv.Itoa(x),
	}
	return walkBasicLit(e, nil)
}

func emitAddrForMapSet(indexExpr *MetaIndexExpr) {
	// alloc heap for map value
	//size := getSizeOfType(elmType)
	emitComment(2, "[emitAddrForMapSet]\n")
	mp := indexExpr.X
	key := indexExpr.Index

	args := []*Arg{
		&Arg{
			meta:      mp,
			paramType: tUintptr,
		},
		&Arg{
			meta:      key,
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
	emitCallDirect("runtime.getAddrForMapSet", args, resultList)
}

func emitListElementAddr(list MetaExpr, elmType *Type) {
	emitListHeadAddr(list)
	emitPopAddress("list head")
	printf("  popq %%rcx # index id\n")
	printf("  movq $%d, %%rdx # elm size\n", getSizeOfType(elmType))
	printf("  imulq %%rdx, %%rcx\n")
	printf("  addq %%rcx, %%rax\n")
	printf("  pushq %%rax # addr of element\n")
}

func emitCatStrings(left MetaExpr, right MetaExpr) {
	args := []*Arg{
		&Arg{
			meta:      left,
			paramType: tString,
		},
		&Arg{
			meta:      right,
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
	emitCallDirect("runtime.catstrings", args, resultList)
}

func emitCompStrings(left MetaExpr, right MetaExpr) {
	args := []*Arg{
		&Arg{
			meta:      left,
			paramType: tString,
		},
		&Arg{
			meta:      right,
			paramType: tString,
		},
	}
	resultList := &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type: tBool.E,
			},
		},
	}
	emitCallDirect("runtime.cmpstrings", args, resultList)
}

func emitBinaryExprComparison(left MetaExpr, right MetaExpr) {
	if kind(getTypeOfExprMeta(left)) == T_STRING {
		emitCompStrings(left, right)
	} else if kind(getTypeOfExprMeta(left)) == T_INTERFACE {
		//var t = getTypeOfExpr(left)
		ff := lookupForeignFunc(newQI("runtime", "cmpinterface"))
		emitAllocReturnVarsAreaFF(ff)
		//@TODO: confirm nil comparison with interfaces
		emitExpr(left)  // left
		emitExpr(right) // right
		emitCallFF(ff)
	} else {
		// Assuming pointer-like types (pointer, map)
		//var t = getTypeOfExpr(left)
		emitExpr(left)  // left
		emitExpr(right) // right
		emitCompExpr("sete")
	}

	//@TODO: implement nil comparison with slices

}

// @TODO handle larger types than int
func emitCompExpr(inst string) {
	printf("  popq %%rcx # right\n")
	printf("  popq %%rax # left\n")
	printf("  cmpq %%rcx, %%rax\n")
	printf("  %s %%al\n", inst)
	printf("  movzbq %%al, %%rax\n") // true:1, false:0
	printf("  pushq %%rax\n")
}

func emitBitWiseOr() {
	printf("  popq %%rcx # right\n")
	printf("  popq %%rax # left\n")
	printf("  orq %%rcx, %%rax # bitwise or\n")
	printf("  pushq %%rax\n")
}

func emitBitWiseAnd() {
	printf("  popq %%rcx # right\n")
	printf("  popq %%rax # left\n")
	printf("  andq %%rcx, %%rax # bitwise and\n")
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

func isBlankIdentifierMeta(m MetaExpr) bool {
	ident, isIdent := m.(*MetaIdent)
	if !isIdent {
		return false
	}
	return ident.kind == "blank"
}

func isBlankIdentifier(e ast.Expr) bool {
	ident, isIdent := e.(*ast.Ident)
	if !isIdent {
		return false
	}
	return ident.Name == "_"
}

func emitAssignToVar(vr *Variable, rhs MetaExpr) {
	emitComment(2, "Assignment: emitVariableAddr(lhs)\n")
	emitVariableAddr(vr)
	emitComment(2, "Assignment: emitExpr(rhs)\n")

	emitExpr(rhs)
	mayEmitConvertTooIfc(rhs, vr.Typ)
	emitComment(2, "Assignment: emitStore(getTypeOfExpr(lhs))\n")
	emitStore(vr.Typ, true, false)
}

func emitAssignZeroValue(lhs MetaExpr, lhsType *Type) {
	emitComment(2, "emitAssignZeroValue\n")
	emitComment(2, "lhs addresss\n")
	emitAddr(lhs)
	emitComment(2, "emitZeroValue\n")
	emitZeroValue(lhsType)
	emitStore(lhsType, true, false)

}

func emitSingleAssign(lhs MetaExpr, rhs MetaExpr) {
	//	lhs := metaSingle.lhs
	//	rhs := metaSingle.rhs
	if isBlankIdentifierMeta(lhs) {
		emitExpr(rhs)
		emitPop(kind(getTypeOfExprMeta(rhs)))
		return
	}
	emitComment(2, "Assignment: emitAddr(lhs)\n")
	emitAddr(lhs)
	emitComment(2, "Assignment: emitExpr(rhs)\n")
	emitExpr(rhs)
	mayEmitConvertTooIfc(rhs, getTypeOfExprMeta(lhs))
	emitStore(getTypeOfExprMeta(lhs), true, false)
}

func emitBlockStmt(s *MetaBlockStmt) {
	for _, s := range s.List {
		emitStmt(s)
	}
}

func emitExprStmt(s *MetaExprStmt) {
	emitExpr(s.X)
}

// local decl stmt
func emitDeclStmt(meta *MetaVarDecl) {
	if meta.single.rhsMeta == nil {
		// Assign zero value to LHS
		emitAssignZeroValue(meta.single.lhsMeta, meta.lhsType)
	} else {
		emitSingleAssign(meta.single.lhsMeta, meta.single.rhsMeta)
	}
}

// a, b = OkExpr
func emitOkAssignment(meta *MetaTupleAssign) {
	rhs0 := meta.rhsMeta
	rhsTypes := meta.rhsTypes

	// ABI of stack layout after evaluating rhs0
	//	-- stack top
	//	bool
	//	data
	emitExpr(rhs0)
	for i := 1; i >= 0; i-- {
		lhsMeta := meta.lhsMetas[i]
		rhsType := rhsTypes[i]
		if isBlankIdentifierMeta(lhsMeta) {
			emitPop(kind(rhsType))
		} else {
			// @TODO interface conversion
			emitAddr(lhsMeta)
			emitStore(getTypeOfExprMeta(lhsMeta), false, false)
		}

	}
}

// assigns multi-values of a funcall
// a, b (, c...) = f()
func emitFuncallAssignment(meta *MetaTupleAssign) {
	rhs0 := meta.rhsMeta
	rhsTypes := meta.rhsTypes

	emitExpr(rhs0)
	for i := 0; i < len(rhsTypes); i++ {
		lhsMeta := meta.lhsMetas[i]
		rhsType := rhsTypes[i]
		if isBlankIdentifierMeta(lhsMeta) {
			emitPop(kind(rhsType))
		} else {
			switch kind(rhsType) {
			case T_UINT8:
				// repush stack top
				printf("  movzbq (%%rsp), %%rax # load uint8\n")
				printf("  addq $%d, %%rsp # free returnvars area\n", 1)
				printf("  pushq %%rax\n")
			}
			// @TODO interface conversion
			emitAddr(lhsMeta)
			emitStore(getTypeOfExprMeta(lhsMeta), false, false)
		}
	}
}

func emitIfStmt(meta *MetaIfStmt) {
	emitComment(2, "if\n")

	labelid++
	labelEndif := fmt.Sprintf(".L.endif.%d", labelid)
	labelElse := fmt.Sprintf(".L.else.%d", labelid)

	emitExpr(meta.CondMeta)
	emitPopBool("if condition")
	printf("  cmpq $1, %%rax\n")
	if meta.Else != nil {
		printf("  jne %s # jmp if false\n", labelElse)
		emitBlockStmt(meta.Body) // then
		printf("  jmp %s\n", labelEndif)
		printf("  %s:\n", labelElse)
		emitStmt(meta.Else) // then
	} else {
		printf("  jne %s # jmp if false\n", labelEndif)
		emitBlockStmt(meta.Body) // then
	}
	printf("  %s:\n", labelEndif)
	emitComment(2, "end if\n")
}

func emitForStmt(meta *MetaForStmt) {
	labelid++
	labelCond := fmt.Sprintf(".L.for.cond.%d", labelid)
	labelPost := fmt.Sprintf(".L.for.post.%d", labelid)
	labelExit := fmt.Sprintf(".L.for.exit.%d", labelid)

	meta.LabelPost = labelPost
	meta.LabelExit = labelExit

	if meta.Init != nil {
		emitStmt(meta.Init)
	}

	printf("  %s:\n", labelCond)
	if meta.CondMeta != nil {
		emitExpr(meta.CondMeta)
		emitPopBool("for condition")
		printf("  cmpq $1, %%rax\n")
		printf("  jne %s # jmp if false\n", labelExit)
	}
	emitBlockStmt(meta.Body)
	printf("  %s:\n", labelPost) // used for "continue"
	if meta.Post != nil {
		emitStmt(meta.Post)
	}
	printf("  jmp %s\n", labelCond)
	printf("  %s:\n", labelExit)
}

func emitRangeMap(meta *MetaForStmt) {
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
	emitAssignToVar(meta.ForRange.MapVar, meta.ForRange.X)

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
	keyMeta := meta.ForRange.Key
	if keyMeta != nil {
		if !isBlankIdentifierMeta(keyMeta) {
			emitAddr(keyMeta) // lhs
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
			emitLoadAndPush(getTypeOfExprMeta(keyMeta)) // load dynamic data
			emitStore(getTypeOfExprMeta(keyMeta), true, false)
		}
	}

	// assign value
	valueMeta := meta.ForRange.Value
	if valueMeta != nil {
		if !isBlankIdentifierMeta(valueMeta) {
			emitAddr(valueMeta) // lhs
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
			emitLoadAndPush(getTypeOfExprMeta(valueMeta)) // load dynamic data
			emitStore(getTypeOfExprMeta(valueMeta), true, false)
		}
	}

	// Body
	emitComment(2, "ForRange Body\n")
	emitBlockStmt(meta.Body)

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

func emitRangeStmt(meta *MetaForStmt) {
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
	emitLen(meta.ForRange.X)
	emitStore(tInt, true, false)

	emitComment(2, "  assign 0 to indexvar\n")
	// indexvar = 0
	emitVariableAddr(meta.ForRange.Indexvar)
	emitZeroValue(tInt)
	emitStore(tInt, true, false)

	// init key variable with 0
	keyMeta := meta.ForRange.Key
	if keyMeta != nil {
		if !isBlankIdentifierMeta(keyMeta) {
			emitAddr(keyMeta) // lhs
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
	elemType := getTypeOfExprMeta(meta.ForRange.Value)
	emitAddr(meta.ForRange.Value) // lhs

	emitVariableAddr(meta.ForRange.Indexvar)
	emitLoadAndPush(tInt) // index value
	emitListElementAddr(meta.ForRange.X, elemType)

	emitLoadAndPush(elemType)
	emitStore(elemType, true, false)

	// Body
	emitComment(2, "ForRange Body\n")
	emitBlockStmt(meta.Body)

	// Post statement: Increment indexvar and go next
	emitComment(2, "ForRange Post statement\n")
	printf("  %s:\n", labelPost)             // used for "continue"
	emitVariableAddr(meta.ForRange.Indexvar) // lhs
	emitVariableAddr(meta.ForRange.Indexvar) // rhs
	emitLoadAndPush(tInt)
	emitAddConst(1, "indexvar value ++")
	emitStore(tInt, true, false)

	// incr key variable
	if keyMeta != nil {
		if !isBlankIdentifierMeta(keyMeta) {
			emitAddr(keyMeta)                        // lhs
			emitVariableAddr(meta.ForRange.Indexvar) // rhs
			emitLoadAndPush(tInt)
			emitStore(tInt, true, false)
		}
	}

	printf("  jmp %s\n", labelCond)

	printf("  %s:\n", labelExit)
}

func emitSwitchStmt(s *MetaSwitchStmt) {
	labelid++
	labelEnd := fmt.Sprintf(".L.switch.%d.exit", labelid)
	if s.Init != nil {
		panic("TBI")
	}
	if s.Tag == nil {
		panic("Omitted tag is not supported yet")
	}
	emitExpr(s.Tag)
	condType := getTypeOfExprMeta(s.Tag)
	cases := s.cases
	var labels = make([]string, len(cases), len(cases))
	var defaultLabel string
	emitComment(2, "Start comparison with cases\n")
	for i, cc := range cases {
		labelid++
		labelCase := fmt.Sprintf(".L.case.%d", labelid)
		labels[i] = labelCase
		if len(cc.ListMeta) == 0 {
			defaultLabel = labelCase
			continue
		}
		for _, m := range cc.ListMeta {
			assert(getSizeOfType(condType) <= 8 || kind(condType) == T_STRING, "should be one register size or string", __func__)
			switch kind(condType) {
			case T_STRING:
				ff := lookupForeignFunc(newQI("runtime", "cmpstrings"))
				emitAllocReturnVarsAreaFF(ff)

				emitPushStackTop(condType, SizeOfInt, "switch expr")
				emitExpr(m)

				emitCallFF(ff)
			case T_INTERFACE:
				ff := lookupForeignFunc(newQI("runtime", "cmpinterface"))

				emitAllocReturnVarsAreaFF(ff)

				emitPushStackTop(condType, SizeOfInt, "switch expr")
				emitExpr(m)

				emitCallFF(ff)
			case T_INT, T_UINT8, T_UINT16, T_UINTPTR, T_POINTER:
				emitPushStackTop(condType, 0, "switch expr")
				emitExpr(m)
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
	for i, cc := range s.cases {
		printf("  %s:\n", labels[i])
		for _, _s := range cc.Body {
			emitStmt(_s)
		}
		printf("  jmp %s\n", labelEnd)
	}
	printf("  %s:\n", labelEnd)
}
func emitTypeSwitchStmt(meta *MetaTypeSwitchStmt) {
	labelid++
	labelEnd := fmt.Sprintf(".L.typeswitch.%d.exit", labelid)

	// subjectVariable = subject
	emitVariableAddr(meta.SubjectVariable)
	emitExpr(meta.Subject)
	emitStore(tEface, true, false)

	cases := meta.Cases
	var labels = make([]string, len(cases), len(cases))
	var defaultLabel string
	emitComment(2, "Start comparison with cases\n")
	for i, c := range meta.Cases {
		labelid++
		labelCase := ".L.case." + strconv.Itoa(labelid)
		labels[i] = labelCase
		if len(c.types) == 0 {
			defaultLabel = labelCase
			continue
		}
		for _, t := range c.types {
			emitVariableAddr(meta.SubjectVariable)
			emitPopAddress("type switch subject")
			printf("  movq (%%rax), %%rax # dtype label addr\n")
			printf("  pushq %%rax # dtype label addr\n")

			if t == nil { // case nil:
				printf("  pushq $0 # nil\n")
			} else { // case T:s
				emitDtypeLabelAddr(t)
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

	for i, c := range meta.Cases {
		// Injecting variable and type to the subject
		if c.Variable != nil {
			setVariable(meta.AssignIdent, c.Variable)
		}
		printf("  %s:\n", labels[i])

		var _isNil bool
		for _, typ := range c.types {
			if typ == nil {
				_isNil = true
			}
		}

		if c.Variable != nil {
			// do assignment
			if _isNil {
				// @TODO: assign nil to the AssignIdent of interface type
			} else {
				emitVariableAddr(c.Variable) // push lhs

				// push rhs
				emitVariableAddr(meta.SubjectVariable)
				emitLoadAndPush(tEface)
				printf("  popq %%rax # ifc.dtype\n")
				printf("  popq %%rcx # ifc.data\n")
				printf("  pushq %%rcx # ifc.data\n")
				emitLoadAndPush(c.VariableType)

				// assign
				emitStore(c.VariableType, true, false)
			}
		}

		for _, _s := range c.Body {
			emitStmt(_s)
		}
		printf("  jmp %s\n", labelEnd)
	}
	printf("  %s:\n", labelEnd)
}

func emitBranchStmt(meta *MetaBranchStmt) {
	containerFor := meta.containerForStmt
	switch meta.ContinueOrBreak {
	case 1: // continue
		printf("  jmp %s # continue\n", containerFor.LabelPost)
	case 2: // break
		printf("  jmp %s # break\n", containerFor.LabelExit)
	default:
		throw(meta.ContinueOrBreak)
	}
}

func emitGoStmt(m *MetaGoStmt) {
	emitCallMalloc(SizeOfPtr) // area := new(func())
	emitExpr(m.fun)
	printf("  popq %%rax # func addr\n")
	printf("  popq %%rcx # malloced area\n")
	printf("  movq %%rax, (%%rcx) # malloced area\n") // *area = fn
	printf("  pushq %%rcx # malloced area\n")
	printf("  pushq $0 # arg size\n")
	printf("  callq runtime.newproc\n") // runtime.newproc(0, area)
	printf("  popq %%rax\n")
	printf("  popq %%rax\n")
	printf("  callq runtime.mstart0\n")
}

func emitStmt(mtstmt MetaStmt) {
	switch meta := mtstmt.(type) {
	case *MetaBlockStmt:
		emitBlockStmt(meta)
	case *MetaExprStmt:
		emitExprStmt(meta)
	case *MetaVarDecl:
		emitDeclStmt(meta)
	case *MetaSingleAssign:
		emitSingleAssign(meta.lhsMeta, meta.rhsMeta)
	case *MetaOkAssignStmt:
		emitOkAssignment(meta.tuple)
	case *MetaFuncallAssignStmt:
		emitFuncallAssignment(meta.tuple)
	case *MetaReturnStmt:
		emitReturnStmt(meta)
	case *MetaIfStmt:
		emitIfStmt(meta)
	case *MetaForStmt:
		if meta.ForRange != nil {
			if meta.ForRange.IsMap {
				emitRangeMap(meta)
			} else {
				emitRangeStmt(meta)
			}
		} else {
			emitForStmt(meta)
		}
	case *MetaSwitchStmt:
		emitSwitchStmt(meta)
	case *MetaTypeSwitchStmt:
		emitTypeSwitchStmt(meta)
	case *MetaBranchStmt:
		emitBranchStmt(meta)
	case *MetaGoStmt:
		emitGoStmt(meta)
	default:
		panic(fmt.Sprintf("unknown type:%T", mtstmt))
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
	printf("\n")
	logf2("# emitFuncDecl pkg=%s, fnc.name=%s\n", pkgName, fnc.Name)
	var symbol string
	if fnc.Method != nil {
		symbol = getMethodSymbol(fnc.Method)
		printf("# Method %s\n", symbol)
	} else {
		symbol = getPackageSymbol(pkgName, fnc.Name)
		printf("# Function %s\n", symbol)
	}
	printf(".global %s\n", symbol)
	printf("%s: # args %d, locals %d\n", symbol, fnc.Argsarea, fnc.Localarea)
	printf("  pushq %%rbp\n")
	printf("  movq %%rsp, %%rbp\n")

	if fnc.Localarea != 0 {
		printf("  subq $%d, %%rsp # local area\n", -fnc.Localarea)
	}
	for _, m := range fnc.Stmts {
		emitStmt(m)
	}
	printf("  leave\n")
	printf("  ret\n")
}

func emitGlobalVariable(pkg *PkgContainer, vr *packageVar) {
	name := vr.name.Name
	t := vr.typ
	typeKind := kind(vr.typ)
	val := vr.val
	printf(".global %s.%s\n", pkg.name, name)
	printf("%s.%s: # T %s\n", pkg.name, name, string(typeKind))

	metaVal := vr.metaVal
	_ = metaVal
	switch typeKind {
	case T_STRING:
		if metaVal == nil {
			// no value
			printf("  .quad 0\n")
			printf("  .quad 0\n")
		} else {
			lit, ok := metaVal.(*MetaBasicLit)
			if !ok {
				panic("only BasicLit is supported")
			}
			sl := lit.stringLiteral
			printf("  .quad %s\n", sl.label)
			printf("  .quad %d\n", sl.strlen)
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
	case T_INT32:
		switch vl := val.(type) {
		case nil:
			printf("  .long 0\n")
		case *ast.BasicLit:
			printf("  .long %s\n", vl.Value)
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
	case T_STRUCT:
		if val != nil {
			panic("Unsupported global value")
		}
		for i := 0; i < getSizeOfType(t); i++ {
			printf("  .byte 0 # struct zero value\n")
		}
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
	case T_POINTER, T_FUNC:
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
	printf("#--- string literals\n")
	printf(".data\n")
	for _, sl := range pkg.stringLiterals {
		printf("%s:\n", sl.label)
		printf("  .string %s\n", sl.value)
	}

	printf("#--- global vars (static values)\n")
	for _, vr := range pkg.vars {
		if vr.typ == nil {
			panic("type cannot be nil for global variable: " + vr.name.Name)
		}
		emitGlobalVariable(pkg, vr)
	}

	printf("\n")
	printf("#--- global vars (dynamic value setting)\n")
	printf(".text\n")
	printf(".global %s.__initGlobals\n", pkg.name)
	printf("%s.__initGlobals:\n", pkg.name)
	for _, vr := range pkg.vars {
		if vr.val == nil {
			continue
		}
		typeKind := kind(vr.typ)
		switch typeKind {
		case T_POINTER, T_MAP, T_INTERFACE:
			printf("# init global %s:\n", vr.name.Name)
			emitSingleAssign(vr.metaVar, vr.metaVal)
		}
	}
	printf("  ret\n")

	for _, fnc := range pkg.funcs {
		emitFuncDecl(pkg.name, fnc)
	}

	emitDynamicTypes(typesMap)
	printf("\n")
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
		printf("  .quad .string.dtype.%d\n", id)
		printf("  .quad %d\n", len(ent.serialized))
		printf(".string.dtype.%d:\n", id)
		printf("  .string \"%s\"\n", ent.serialized)
	}
	printf("\n")
}

// --- type ---
type Type struct {
	E       ast.Expr // original
	PkgName string
	Name    string
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
func getTypeOfExprMeta(meta MetaExpr) *Type {
	switch m := meta.(type) {
	case *MetaIdent:
		return getTypeOfExpr(m.e)
	case *MetaBasicLit:
		// The default type of an untyped constant is bool, rune, int, float64, complex128 or string respectively,
		// depending on whether it is a boolean, rune, integer, floating-point, complex, or string constant.
		switch m.Kind {
		case "STRING":
			return tString
		case "INT":
			return tInt
		case "CHAR":
			return tInt32
		default:
			panic(m.Kind)
		}
	case *MetaUnaryExpr:
		return getTypeOfExpr(m.e)
	case *MetaBinaryExpr:
		switch m.Op {
		case "==", "!=", "<", ">", "<=", ">=":
			return tBool
		default:
			return getTypeOfExpr(m.e)
		}
	case *MetaIndexExpr:
		return getTypeOfExpr(m.e)
	case *MetaCallExpr: // funcall or conversion
		return getTypeOfExpr(m.e)
	case *MetaSliceExpr:
		return getTypeOfExpr(m.e)
	case *MetaStarExpr:
		return getTypeOfExpr(m.e)
	case *MetaSelectorExpr:
		return getTypeOfExpr(m.e)
	case *MetaCompositLiteral:
		return getTypeOfExpr(m.e)
	case *MetaTypeAssertExpr:
		return getTypeOfExpr(m.e)
	}
	panic("bad type\n")
}

func getTypeOfExpr(expr ast.Expr) *Type {
	switch e := expr.(type) {
	case *ast.Ident:
		assert(e.Obj != nil, "Obj is nil in ident '"+e.Name+"'", __func__)
		switch e.Obj.Kind {
		case ast.Var:
			variable, isVariable := e.Obj.Data.(*Variable)
			if isVariable {
				return variable.Typ
			}
			panic("Variable is not set")
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
			panic("Obj=" + e.Obj.Name + ", Kind=" + e.Obj.Kind.String())
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
					return fieldList2Types(ff.funcType.Results)
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
				walkExpr(starExpr, nil)
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
			return fieldList2Types(ff.funcType.Results)
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
	if t.Name != "" {
		return t.PkgName + "." + t.Name
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
		r := "struct{"
		if e.Fields != nil {
			for _, field := range e.Fields.List {
				name := field.Names[0].Name
				typ := e2t(field.Type)
				r += fmt.Sprintf("%s %s;", name, serializeType(typ))
			}
		}
		return r + "}"
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
		return "interface{}" // @TODO list methods
	case *ast.MapType:
		return "map[" + serializeType(e2t(e.Key)) + "]" + serializeType(e2t(e.Value))
	case *ast.SelectorExpr:
		qi := selector2QI(e)
		return string(qi)
	case *ast.FuncType:
		return "func"
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
		assert(e.Obj != nil, "should not be nil : "+e.Name, __func__)
		assert(e.Obj.Kind == ast.Typ, "should be ast.Typ : "+e.Name, __func__)
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

func registerStringLiteral(lit *ast.BasicLit) *sliteral {
	if currentPkg.name == "" {
		panic("no pkgName")
	}

	var strlen int
	for _, c := range []uint8(lit.Value) {
		if c != '\\' {
			strlen++
		}
	}

	label := fmt.Sprintf(".string_%d", currentPkg.stringIndex)
	currentPkg.stringIndex++

	sl := &sliteral{
		label:  label,
		strlen: strlen - 2,
		value:  lit.Value,
	}
	currentPkg.stringLiterals = append(currentPkg.stringLiterals, sl)
	return sl
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
	assert(ident.Obj != nil, "ident.Obj is nil:"+ident.Name, __func__)
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
		panic("method not found: " + methodName.Name)
	}
	return method
}

func walkExprStmt(s *ast.ExprStmt) *MetaExprStmt {
	m := walkExpr(s.X, nil)
	return &MetaExprStmt{X: m}
}

func walkDeclStmt(s *ast.DeclStmt) *MetaVarDecl {
	genDecl := s.Decl.(*ast.GenDecl)
	declSpec := genDecl.Specs[0]
	switch spec := declSpec.(type) {
	case *ast.ValueSpec:
		if spec.Type == nil { // var x = e
			// infer lhs type from rhs
			if len(spec.Values) == 0 {
				panic("invalid syntax")
			}
			val := spec.Values[0]
			typ := getTypeOfExpr(val)
			if typ == nil || typ.E == nil {
				panic("rhs should have a type")
			}
			spec.Type = typ.E // set lhs type
		} else {
			walkExpr(spec.Type, nil)
		}
		t := e2t(spec.Type)
		lhsIdent := spec.Names[0]
		obj := lhsIdent.Obj
		setVariable(obj, registerLocalVariable(currentFunc, obj.Name, t))
		lhsMeta := walkIdent(lhsIdent, nil)
		var rhsMeta MetaExpr
		if len(spec.Values) > 0 {
			rhs := spec.Values[0]
			ctx := &evalContext{_type: t}
			rhsMeta = walkExpr(rhs, ctx)
		}
		single := &MetaSingleAssign{
			lhsMeta: lhsMeta,
			rhsMeta: rhsMeta,
		}
		return &MetaVarDecl{
			single:  single,
			lhsType: t,
		}
	default:
		// @TODO type, const, etc
	}

	panic("TBI")
}

func IsOkSyntax(s *ast.AssignStmt) bool {
	rhs0 := s.Rhs[0]
	_, isTypeAssertion := rhs0.(*ast.TypeAssertExpr)
	if len(s.Lhs) == 2 && isTypeAssertion {
		return true
	}
	indexExpr, isIndexExpr := rhs0.(*ast.IndexExpr)
	if len(s.Lhs) == 2 && isIndexExpr && kind(getTypeOfExpr(indexExpr.X)) == T_MAP {
		return true
	}
	return false
}

func walkAssignStmt(s *ast.AssignStmt) MetaStmt {
	var mt MetaStmt
	stok := s.Tok.String()
	switch stok {
	case "=":
		if IsOkSyntax(s) {
			rhsMeta := walkExpr(s.Rhs[0], &evalContext{okContext: true})
			rhsTypes := []*Type{getTypeOfExpr(s.Rhs[0]), tBool}
			var lhsMetas []MetaExpr
			for _, lhs := range s.Lhs {
				lm := walkExpr(lhs, nil)
				lhsMetas = append(lhsMetas, lm)
			}
			mt = &MetaOkAssignStmt{
				tuple: &MetaTupleAssign{
					lhsMetas: lhsMetas,
					rhsMeta:  rhsMeta,
					rhsTypes: rhsTypes,
				},
			}
		} else {
			if len(s.Lhs) == 1 && len(s.Rhs) == 1 {
				var ctx *evalContext
				if !isBlankIdentifier(s.Lhs[0]) {
					ctx = &evalContext{
						_type: getTypeOfExpr(s.Lhs[0]),
					}
				}
				rhsMeta := walkExpr(s.Rhs[0], ctx)
				var lhsMetas []MetaExpr
				for _, lhs := range s.Lhs {
					lm := walkExpr(lhs, nil)
					lhsMetas = append(lhsMetas, lm)
				}
				mt = &MetaSingleAssign{
					lhsMeta: lhsMetas[0],
					rhsMeta: rhsMeta,
				}
			} else if len(s.Lhs) >= 1 && len(s.Rhs) == 1 {
				rhsMeta := walkExpr(s.Rhs[0], nil)
				callExpr := s.Rhs[0].(*ast.CallExpr)
				returnTypes := getCallResultTypes(callExpr)
				assert(len(returnTypes) == len(s.Lhs), fmt.Sprintf("length unmatches %d <=> %d", len(s.Lhs), len(returnTypes)), __func__)
				var lhsMetas []MetaExpr
				for _, lhs := range s.Lhs {
					lm := walkExpr(lhs, nil)
					lhsMetas = append(lhsMetas, lm)
				}

				mt = &MetaFuncallAssignStmt{
					tuple: &MetaTupleAssign{
						lhsMetas: lhsMetas,
						rhsMeta:  rhsMeta,
						rhsTypes: returnTypes,
					},
				}
			} else {
				panic("TBI")
			}
		}
		return mt
	case ":=":

		if IsOkSyntax(s) {
			rhsMeta := walkExpr(s.Rhs[0], &evalContext{okContext: true})
			rhsTypes := []*Type{getTypeOfExpr(s.Rhs[0]), tBool}
			assert(len(s.Lhs) == len(rhsTypes), fmt.Sprintf("length unmatches %d <=> %d", len(s.Lhs), len(rhsTypes)), __func__)

			lhsTypes := rhsTypes
			var lhsMetas []MetaExpr
			for i, lhs := range s.Lhs {
				typ := lhsTypes[i]
				obj := lhs.(*ast.Ident).Obj
				setVariable(obj, registerLocalVariable(currentFunc, obj.Name, typ))
				lm := walkExpr(lhs, nil)
				lhsMetas = append(lhsMetas, lm)
			}

			mt = &MetaOkAssignStmt{
				tuple: &MetaTupleAssign{
					lhsMetas: lhsMetas,
					rhsMeta:  rhsMeta,
					rhsTypes: rhsTypes,
				},
			}
			return mt
		} else if len(s.Lhs) > 1 && len(s.Rhs) == 1 {
			rhsMeta := walkExpr(s.Rhs[0], nil)
			callExpr := s.Rhs[0].(*ast.CallExpr)
			rhsTypes := getCallResultTypes(callExpr)
			assert(len(s.Lhs) == len(rhsTypes), fmt.Sprintf("length unmatches %d <=> %d", len(s.Lhs), len(rhsTypes)), __func__)

			lhsTypes := rhsTypes
			var lhsMetas []MetaExpr
			for i, lhs := range s.Lhs {
				typ := lhsTypes[i]
				obj := lhs.(*ast.Ident).Obj
				setVariable(obj, registerLocalVariable(currentFunc, obj.Name, typ))
				lm := walkExpr(lhs, nil)
				lhsMetas = append(lhsMetas, lm)
			}
			_ = lhsMetas

			mt = &MetaFuncallAssignStmt{
				tuple: &MetaTupleAssign{
					lhsMetas: lhsMetas,
					rhsMeta:  rhsMeta,
					rhsTypes: rhsTypes,
				},
			}
			return mt
		} else if len(s.Lhs) == 1 && len(s.Rhs) == 1 {
			rhsMeta := walkExpr(s.Rhs[0], nil) // FIXME
			rhsType := getTypeOfExpr(s.Rhs[0])
			lhsTypes := []*Type{rhsType}
			var lhsMetas []MetaExpr
			for i, lhs := range s.Lhs {
				typ := lhsTypes[i]
				obj := lhs.(*ast.Ident).Obj
				setVariable(obj, registerLocalVariable(currentFunc, obj.Name, typ))
				lm := walkExpr(lhs, nil)
				lhsMetas = append(lhsMetas, lm)
			}

			mt = &MetaSingleAssign{
				lhsMeta: lhsMetas[0],
				rhsMeta: rhsMeta,
			}
			return mt
		} else {
			panic("TBI")
		}
		return mt

	case "+=", "-=":
		var op token.Token
		switch stok {
		case "+=":
			op = token.ADD
		case "-=":
			op = token.SUB
		}
		binaryExpr := &ast.BinaryExpr{
			X:  s.Lhs[0],
			Op: op,
			Y:  s.Rhs[0],
		}
		rhsMeta := walkExpr(binaryExpr, nil)
		lhsMeta := walkExpr(s.Lhs[0], nil)
		return &MetaSingleAssign{
			lhsMeta: lhsMeta,
			rhsMeta: rhsMeta,
		}
	default:
		panic("TBI")
	}
	return nil
}

type MetaSingleAssign struct {
	lhsMeta MetaExpr
	//rhs ast.Expr // rhs can be nil
	rhsMeta MetaExpr // can be nil
}

type MetaTupleAssign struct {
	lhsMetas []MetaExpr
	rhsMeta  MetaExpr
	rhsTypes []*Type
}

type MetaVarDecl struct {
	single  *MetaSingleAssign
	lhsType *Type
}

type MetaOkAssignStmt struct {
	tuple *MetaTupleAssign
}
type MetaFuncallAssignStmt struct {
	tuple *MetaTupleAssign
}

func walkReturnStmt(s *ast.ReturnStmt) *MetaReturnStmt {
	funcDef := currentFunc
	if len(funcDef.Retvars) != len(s.Results) {
		panic("length of return and func type do not match")
	}

	_len := len(funcDef.Retvars)
	var results []MetaExpr
	for i := 0; i < _len; i++ {
		expr := s.Results[i]
		retTyp := funcDef.Retvars[i].Typ
		ctx := &evalContext{
			_type: retTyp,
		}
		m := walkExpr(expr, ctx)
		results = append(results, m)
	}
	return &MetaReturnStmt{
		Fnc:         funcDef,
		MetaResults: results,
	}
}

type MetaIfStmt struct {
	Init     MetaStmt
	CondMeta MetaExpr
	Body     *MetaBlockStmt
	Else     MetaStmt
}

func walkIfStmt(s *ast.IfStmt) *MetaIfStmt {
	var mInit MetaStmt
	var mElse MetaStmt
	var condMeta MetaExpr
	if s.Init != nil {
		mInit = walkStmt(s.Init)
	}
	if s.Cond != nil {
		condMeta = walkExpr(s.Cond, nil)
	}
	mtBlock := walkBlockStmt(s.Body)
	if s.Else != nil {
		mElse = walkStmt(s.Else)
	}
	return &MetaIfStmt{
		Init:     mInit,
		CondMeta: condMeta,
		Body:     mtBlock,
		Else:     mElse,
	}
}

func walkBlockStmt(s *ast.BlockStmt) *MetaBlockStmt {
	mt := &MetaBlockStmt{}
	for _, stmt := range s.List {
		meta := walkStmt(stmt)
		mt.List = append(mt.List, meta)
	}
	return mt
}

func walkForStmt(s *ast.ForStmt) *MetaForStmt {
	meta := &MetaForStmt{
		Outer: currentFor,
	}
	currentFor = meta

	if s.Init != nil {
		meta.Init = walkStmt(s.Init)
	}
	if s.Cond != nil {
		meta.CondMeta = walkExpr(s.Cond, nil)
	}
	if s.Post != nil {
		meta.Post = walkStmt(s.Post)
	}
	mtBlock := walkBlockStmt(s.Body)
	meta.Body = mtBlock
	currentFor = meta.Outer
	return meta
}
func walkRangeStmt(s *ast.RangeStmt) *MetaForStmt {
	meta := &MetaForStmt{
		Outer: currentFor,
	}
	currentFor = meta
	metaX := walkExpr(s.X, nil)

	collectionType := getUnderlyingType(getTypeOfExpr(s.X))
	keyType := getKeyTypeOfCollectionType(collectionType)
	elmType := getElementTypeOfCollectionType(collectionType)
	walkExpr(tInt.E, nil)
	switch kind(collectionType) {
	case T_SLICE, T_ARRAY:
		meta.ForRange = &MetaForRange{
			IsMap:    false,
			LenVar:   registerLocalVariable(currentFunc, ".range.len", tInt),
			Indexvar: registerLocalVariable(currentFunc, ".range.index", tInt),
			X:        metaX,
		}
	case T_MAP:
		meta.ForRange = &MetaForRange{
			IsMap:   true,
			MapVar:  registerLocalVariable(currentFunc, ".range.map", tUintptr),
			ItemVar: registerLocalVariable(currentFunc, ".range.item", tUintptr),
			X:       metaX,
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
	if s.Key != nil {
		meta.ForRange.Key = walkExpr(s.Key, nil)
	}
	if s.Value != nil {
		meta.ForRange.Value = walkExpr(s.Value, nil)
	}

	mtBlock := walkBlockStmt(s.Body)
	meta.Body = mtBlock
	currentFor = meta.Outer
	return meta
}

func walkIncDecStmt(s *ast.IncDecStmt) *MetaSingleAssign {
	var binop token.Token
	switch s.Tok.String() {
	case "++":
		binop = token.ADD
	case "--":
		binop = token.SUB
	default:
		panic("Unexpected Tok=" + s.Tok.String())
	}
	exprOne := &ast.BasicLit{Kind: token.INT, Value: "1"}
	newRhs := &ast.BinaryExpr{
		X:  s.X,
		Y:  exprOne,
		Op: binop,
	}
	rhsMeta := walkExpr(newRhs, nil)
	lhsMeta := walkExpr(s.X, nil)
	return &MetaSingleAssign{
		lhsMeta: lhsMeta,
		rhsMeta: rhsMeta,
	}
}

func walkSwitchStmt(s *ast.SwitchStmt) *MetaSwitchStmt {
	meta := &MetaSwitchStmt{}
	if s.Init != nil {
		meta.Init = walkStmt(s.Init)
	}
	if s.Tag != nil {
		meta.Tag = walkExpr(s.Tag, nil)
	}
	var cases []*MetaCaseClause
	for _, _case := range s.Body.List {
		cc := _case.(*ast.CaseClause)
		_cc := walkCaseClause(cc)
		cases = append(cases, _cc)
	}
	meta.cases = cases

	return meta
}

func walkTypeSwitchStmt(meta *ast.TypeSwitchStmt) *MetaTypeSwitchStmt {
	typeSwitch := &MetaTypeSwitchStmt{}
	var assignIdent *ast.Ident

	switch assign := meta.Assign.(type) {
	case *ast.ExprStmt:
		typeAssertExpr := assign.X.(*ast.TypeAssertExpr)
		typeSwitch.Subject = walkExpr(typeAssertExpr.X, nil)
	case *ast.AssignStmt:
		lhs := assign.Lhs[0]
		assignIdent = lhs.(*ast.Ident)
		typeSwitch.AssignIdent = assignIdent.Obj
		// ident will be a new local variable in each case clause
		typeAssertExpr := assign.Rhs[0].(*ast.TypeAssertExpr)
		typeSwitch.Subject = walkExpr(typeAssertExpr.X, nil)
	default:
		throw(meta.Assign)
	}

	typeSwitch.SubjectVariable = registerLocalVariable(currentFunc, ".switch_expr", tEface)

	var cases []*MetaTypeSwitchCaseClose
	for _, _case := range meta.Body.List {
		cc := _case.(*ast.CaseClause)
		tscc := &MetaTypeSwitchCaseClose{}
		cases = append(cases, tscc)

		if assignIdent != nil && len(cc.List) > 0 {
			var varType *Type
			if isNil(cc.List[0]) {
				varType = getTypeOfExprMeta(typeSwitch.Subject)
			} else {
				varType = e2t(cc.List[0])
			}
			// inject a variable of that type
			vr := registerLocalVariable(currentFunc, assignIdent.Name, varType)
			tscc.Variable = vr
			tscc.VariableType = varType
			setVariable(assignIdent.Obj, vr)
		}
		var body []MetaStmt
		for _, stmt := range cc.Body {
			m := walkStmt(stmt)
			body = append(body, m)
		}
		tscc.Body = body
		var types []*Type
		for _, e := range cc.List {
			var typ *Type
			if !isNil(e) {
				typ = e2t(e)
			}
			types = append(types, typ) // universe nil can be appended
		}
		tscc.types = types
		if assignIdent != nil {
			setVariable(assignIdent.Obj, nil)
		}
	}
	typeSwitch.Cases = cases

	return typeSwitch
}
func isNil(e ast.Expr) bool {
	ident, ok := e.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Obj == gNil
}

func walkCaseClause(s *ast.CaseClause) *MetaCaseClause {
	var listMeta []MetaExpr
	for _, e := range s.List {
		m := walkExpr(e, nil)
		listMeta = append(listMeta, m)
	}
	var body []MetaStmt
	for _, stmt := range s.Body {
		metaStmt := walkStmt(stmt)
		body = append(body, metaStmt)
	}
	return &MetaCaseClause{
		ListMeta: listMeta,
		Body:     body,
	}
}

func walkBranchStmt(s *ast.BranchStmt) *MetaBranchStmt {
	assert(currentFor != nil, "break or continue should be in for body", __func__)
	var continueOrBreak int
	switch s.Tok.String() {
	case "continue":
		continueOrBreak = 1
	case "break":
		continueOrBreak = 2
	default:
		panic("Unexpected token")
	}

	return &MetaBranchStmt{
		containerForStmt: currentFor,
		ContinueOrBreak:  continueOrBreak,
	}
}

func walkGoStmt(s *ast.GoStmt) *MetaGoStmt {
	fun := walkExpr(s.Call.Fun, nil)
	return &MetaGoStmt{
		fun: fun,
	}
}

type MetaStmt interface{}

type MetaBlockStmt struct {
	List []MetaStmt
}

type MetaExprStmt struct {
	X MetaExpr
}

type MetaGoStmt struct {
	fun MetaExpr
}

type MetaReturnStmt struct {
	Fnc         *Func
	MetaResults []MetaExpr
}

type MetaForRange struct {
	IsMap    bool
	LenVar   *Variable
	Indexvar *Variable
	MapVar   *Variable // map
	ItemVar  *Variable // map element
	X        MetaExpr
	Key      MetaExpr
	Value    MetaExpr
}

type MetaForStmt struct {
	LabelPost string // for continue
	LabelExit string // for break
	Outer     *MetaForStmt
	ForRange  *MetaForRange
	Init      MetaStmt
	CondMeta  MetaExpr
	Body      *MetaBlockStmt
	Post      MetaStmt
}

type MetaBranchStmt struct {
	containerForStmt *MetaForStmt
	ContinueOrBreak  int // 1: continue, 2:break
}

type MetaCaseClause struct {
	ListMeta []MetaExpr
	Body     []MetaStmt
}

type MetaSwitchStmt struct {
	Init  MetaStmt
	cases []*MetaCaseClause
	Tag   MetaExpr
}

type MetaTypeSwitchStmt struct {
	Subject         MetaExpr
	SubjectVariable *Variable
	AssignIdent     *ast.Object
	Cases           []*MetaTypeSwitchCaseClose
	cases           []*MetaCaseClause
}

type MetaTypeSwitchCaseClose struct {
	Variable     *Variable
	VariableType *Type
	types        []*Type
	Body         []MetaStmt
}

func walkStmt(stmt ast.Stmt) MetaStmt {
	var mt MetaStmt
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		mt = walkBlockStmt(s)
	case *ast.ExprStmt:
		mt = walkExprStmt(s)
	case *ast.DeclStmt:
		mt = walkDeclStmt(s)
	case *ast.AssignStmt:
		mt = walkAssignStmt(s)
	case *ast.ReturnStmt:
		mt = walkReturnStmt(s)
	case *ast.IfStmt:
		mt = walkIfStmt(s)
	case *ast.ForStmt:
		mt = walkForStmt(s)
	case *ast.RangeStmt:
		mt = walkRangeStmt(s)
	case *ast.IncDecStmt:
		mt = walkIncDecStmt(s)
	case *ast.SwitchStmt:
		mt = walkSwitchStmt(s)
	case *ast.TypeSwitchStmt:
		mt = walkTypeSwitchStmt(s)
	case *ast.BranchStmt:
		mt = walkBranchStmt(s)
	case *ast.GoStmt:
		mt = walkGoStmt(s)
	default:
		throw(stmt)
	}

	assert(mt != nil, "meta should not be nil", __func__)
	return mt
}

func (m *MetaIdent) isUniverseNil() bool {
	return m.kind == "nil"
}

func walkIdent(e *ast.Ident, ctx *evalContext) *MetaIdent {
	meta := &MetaIdent{
		e:    e,
		Name: e.Name,
	}
	logfncname := "(toplevel)"
	if currentFunc != nil {
		logfncname = currentFunc.Name
	}
	//logf2("walkIdent: pkg=%s func=%s, ident=%s\n", currentPkg.name, logfncname, e.Name)
	_ = logfncname
	// what to do ?
	if e.Name == "_" {
		// blank identifier
		// e.Obj is nil in this case.
		// @TODO do something
		meta.kind = "blank"
		return meta
	}
	assert(e.Obj != nil, currentPkg.name+" ident.Obj should not be nil:"+e.Name, __func__)
	switch e.Obj {
	case gNil:
		assert(ctx != nil, "ctx of nil is not passed", __func__)
		assert(ctx._type != nil, "ctx._type of nil is not passed", __func__)
		meta.typ = ctx._type
		meta.kind = "nil"
	case gTrue:
		meta.kind = "true"
	case gFalse:
		meta.kind = "false"
	default:
		switch e.Obj.Kind {
		case ast.Var:
			meta.kind = "var"
		case ast.Con:
			meta.kind = "con"
			// TODO: attach type
			valSpec := e.Obj.Decl.(*ast.ValueSpec)
			lit := valSpec.Values[0].(*ast.BasicLit)
			meta.conLiteral = walkBasicLit(lit, nil)
		case ast.Fun:
			meta.kind = "fun"
		case ast.Typ:
			meta.kind = "typ"
			// int(expr)
			// TODO: this should be avoided in advance
		default: // ast.Pkg
			panic("Unexpected ident kind:" + e.Obj.Kind.String() + " name:" + e.Name)
		}

	}
	return meta
}

func walkSelectorExpr(e *ast.SelectorExpr, ctx *evalContext) *MetaSelectorExpr {
	meta := &MetaSelectorExpr{e: e}
	if isQI(e) {
		// pkg.ident
		qi := selector2QI(e)
		ident := lookupForeignIdent(qi)
		if ident.Obj.Kind == ast.Fun {
			// what to do ?
		} else {
			walkExpr(ident, ctx)
		}
	} else {
		// expr.field
		meta.X = walkExpr(e.X, ctx)
	}
	return meta
}

func walkCallExpr(e *ast.CallExpr, ctx *evalContext) *MetaCallExpr {
	meta := &MetaCallExpr{
		e: e,
	}
	if isType(e.Fun) {
		meta.isConversion = true
		meta.toType = e2t(e.Fun)
		assert(len(e.Args) == 1, "convert must take only 1 argument", __func__)
		//logf2("walkCallExpr: is Conversion\n")
		ctx := &evalContext{
			_type: e2t(e.Fun),
		}
		meta.arg0 = walkExpr(e.Args[0], ctx)
		return meta
	}

	meta.isConversion = false
	meta.fun = e.Fun
	meta.hasEllipsis = e.Ellipsis != token.NoPos
	meta.args = e.Args

	// function call
	metaFun := walkExpr(e.Fun, nil)

	// Replace __func__ ident by a string literal
	//for i, arg := range meta.args {
	//	ident, ok := arg.(*ast.Ident)
	//	if ok {
	//		if ident.Name == "__func__" && ident.Obj.Kind == ast.Var {
	//			basicLit := &ast.BasicLit{
	//				Kind:  token.STRING,
	//				Value: "\"" + currentFunc.Name + "\"",
	//			}
	//			arg = basicLit
	//			e.Args[i] = arg
	//		}
	//	}
	//
	//}

	identFun, isIdent := meta.fun.(*ast.Ident)
	if isIdent {
		switch identFun.Obj {
		case gLen, gCap:
			meta.builtin = identFun.Obj
			meta.arg0 = walkExpr(meta.args[0], nil)
			return meta
		case gNew:
			meta.builtin = identFun.Obj
			walkExpr(meta.args[0], nil)
			meta.typeArg0 = e2t(meta.args[0])
			return meta
		case gMake:
			meta.builtin = identFun.Obj
			walkExpr(meta.args[0], nil)
			meta.typeArg0 = e2t(meta.args[0])

			ctx := &evalContext{_type: tInt}
			if len(meta.args) > 1 {
				meta.arg1 = walkExpr(meta.args[1], ctx)
			}

			if len(meta.args) > 2 {
				meta.arg2 = walkExpr(meta.args[2], ctx)
			}
			return meta
		case gAppend:
			meta.builtin = identFun.Obj
			meta.arg0 = walkExpr(meta.args[0], nil)
			meta.arg1 = walkExpr(meta.args[1], nil)
			return meta
		case gPanic:
			meta.builtin = identFun.Obj
			meta.arg0 = walkExpr(meta.args[0], nil)
			return meta
		case gDelete:
			meta.builtin = identFun.Obj
			meta.arg0 = walkExpr(meta.args[0], nil)
			meta.arg1 = walkExpr(meta.args[1], nil)
			return meta
		}
	}

	var funcType *ast.FuncType
	var funcVal *FuncValue
	var receiver ast.Expr

	switch fn := meta.fun.(type) {
	case *ast.Ident:
		// general function call
		symbol := getPackageSymbol(currentPkg.name, fn.Name)
		switch currentPkg.name {
		case "os":
			switch fn.Name {
			case "runtime_args":
				symbol = getPackageSymbol("runtime", "runtime_args")
			case "runtime_getenv":
				symbol = getPackageSymbol("runtime", "runtime_getenv")
			}
		case "runtime":
			if fn.Name == "makeSlice1" || fn.Name == "makeSlice8" || fn.Name == "makeSlice16" || fn.Name == "makeSlice24" {
				fn.Name = "makeSlice"
				symbol = getPackageSymbol("runtime", fn.Name)
			}
		}
		funcVal = NewFuncValueFromSymbol(symbol)
		switch dcl := fn.Obj.Decl.(type) {
		case *ast.FuncDecl:
			funcType = dcl.Type
		case *ast.ValueSpec: // var f func()
			funcType = dcl.Type.(*ast.FuncType)
			funcVal = &FuncValue{
				expr: metaFun,
			}
		case *ast.AssignStmt: // f := staticF
			assert(fn.Obj.Data != nil, "funcvalue should be a variable:"+fn.Name, __func__)
			rhs := dcl.Rhs[0]
			switch r := rhs.(type) {
			case *ast.SelectorExpr:
				assert(isQI(r), "expect QI", __func__)
				qi := selector2QI(r)
				ff := lookupForeignFunc(qi)
				funcType = ff.funcType
				funcVal = NewFuncValueFromSymbol(string(qi))
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
			funcVal = NewFuncValueFromSymbol(string(qi))
			ff := lookupForeignFunc(qi)
			funcType = ff.funcType
		} else {
			// method call
			receiver = fn.X
			receiverType := getTypeOfExpr(receiver)
			method := lookupMethod(receiverType, fn.Sel)
			funcType = method.FuncType
			funcVal = NewFuncValueFromSymbol(getMethodSymbol(method))

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
					//walkExpr(receiver, nil)
				} else {
					// v.mv() => as it is
				}
			}
		}
	default:
		throw(meta.fun)
	}

	meta.funcType = funcType
	meta.receiver = receiver
	meta.funcVal = funcVal
	meta.metaArgs = prepareArgs(meta.funcType, meta.receiver, meta.args, meta.hasEllipsis)
	return meta
}

func walkBasicLit(e *ast.BasicLit, ctx *evalContext) *MetaBasicLit {
	mt := &MetaBasicLit{
		e:     e,
		Kind:  e.Kind.String(),
		Value: e.Value,
	}

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
		mt.charValue = int(char)
	case "INT":
		mt.intValue = strconv.Atoi(mt.Value)
	case "STRING":
		mt.stringLiteral = registerStringLiteral(e)
	default:
		panic("Unexpected literal kind:" + e.Kind.String())
	}
	return mt
}

func walkCompositeLit(e *ast.CompositeLit, ctx *evalContext) *MetaCompositLiteral {
	walkExpr(e.Type, nil) // a[len("foo")]{...} // "foo" should be walked
	ut := getUnderlyingType(getTypeOfExpr(e))
	var knd string
	switch kind(ut) {
	case T_STRUCT:
		knd = "struct"
	case T_ARRAY:
		knd = "array"
	case T_SLICE:
		knd = "slice"
	default:
		unexpectedKind(kind(e2t(e.Type)))
	}
	meta := &MetaCompositLiteral{
		e:    e,
		kind: knd,
		typ:  getTypeOfExpr(e),
	}

	switch kind(ut) {
	case T_STRUCT:
		structType := meta.typ
		var metaElms []*MetaStructLiteralElement
		for _, elm := range e.Elts {
			kvExpr := elm.(*ast.KeyValueExpr)
			fieldName := kvExpr.Key.(*ast.Ident)
			field := lookupStructField(getUnderlyingStructType(structType), fieldName.Name)
			fieldType := e2t(field.Type)
			ctx := &evalContext{_type: fieldType}
			// attach type to nil : STRUCT{Key:nil}
			valueMeta := walkExpr(kvExpr.Value, ctx)

			metaElm := &MetaStructLiteralElement{
				field:     field,
				fieldType: fieldType,
				ValueMeta: valueMeta,
			}

			metaElms = append(metaElms, metaElm)
		}
		meta.strctEements = metaElms
	case T_ARRAY:
		meta.arrayType = ut.E.(*ast.ArrayType)
		meta.len = evalInt(meta.arrayType.Len)
		meta.elmType = e2t(meta.arrayType.Elt)
		ctx := &evalContext{_type: meta.elmType}
		var ms []MetaExpr
		for _, v := range e.Elts {
			m := walkExpr(v, ctx)
			ms = append(ms, m)
		}
		meta.metaElms = ms
	case T_SLICE:
		meta.arrayType = ut.E.(*ast.ArrayType)
		meta.len = len(e.Elts)
		meta.elmType = e2t(meta.arrayType.Elt)
		ctx := &evalContext{_type: meta.elmType}
		var ms []MetaExpr
		for _, v := range e.Elts {
			m := walkExpr(v, ctx)
			ms = append(ms, m)
		}
		meta.metaElms = ms
	}
	return meta
}

func walkUnaryExpr(e *ast.UnaryExpr, ctx *evalContext) *MetaUnaryExpr {
	meta := &MetaUnaryExpr{e: e}
	meta.X = walkExpr(e.X, nil)
	return meta
}

func walkBinaryExpr(e *ast.BinaryExpr, ctx *evalContext) *MetaBinaryExpr {
	meta := &MetaBinaryExpr{
		e:  e,
		Op: e.Op.String(),
	}
	var xCtx *evalContext
	var yCtx *evalContext
	if isNil(e.X) {
		// Y should be typed
		xCtx = &evalContext{_type: getTypeOfExpr(e.Y)}
	}
	if isNil(e.Y) {
		// Y should be typed
		yCtx = &evalContext{_type: getTypeOfExpr(e.X)}
	}

	meta.X = walkExpr(e.X, xCtx) // left
	meta.Y = walkExpr(e.Y, yCtx) // right
	return meta
}

func walkIndexExpr(e *ast.IndexExpr, ctx *evalContext) *MetaIndexExpr {
	meta := &MetaIndexExpr{
		e: e,
	}
	if kind(getTypeOfExpr(e.X)) == T_MAP {
		meta.IsMap = true
		if ctx != nil && ctx.okContext {
			meta.NeedsOK = true
		}
	}
	meta.Index = walkExpr(e.Index, nil) // @TODO pass context for map,slice,array
	meta.X = walkExpr(e.X, nil)
	return meta
}

func walkSliceExpr(e *ast.SliceExpr, ctx *evalContext) *MetaSliceExpr {
	meta := &MetaSliceExpr{e: e}

	// For convenience, any of the indices may be omitted.

	// A missing low index defaults to zero;
	if e.Low != nil {
		meta.Low = walkExpr(e.Low, nil)
	} else {
		eZeroInt := &ast.BasicLit{
			Value: "0",
			Kind:  token.INT,
		}
		meta.Low = walkExpr(eZeroInt, nil)
	}

	if e.High != nil {
		meta.High = walkExpr(e.High, nil)
	}
	if e.Max != nil {
		meta.Max = walkExpr(e.Max, nil)
	}
	meta.X = walkExpr(e.X, nil)
	return meta
}

// [N]T(e)
func walkArrayType(e *ast.ArrayType) {
	// BasicLit in N should be walked
	// e.g. A[5], A[len("foo")]
	if e.Len != nil {
		walkExpr(e.Len, nil)
	}
}
func walkMapType(e *ast.MapType) {
	// first argument of builtin func
	// do nothing
}
func walkStarExpr(e *ast.StarExpr, ctx *evalContext) *MetaStarExpr {
	meta := &MetaStarExpr{e: e}
	meta.X = walkExpr(e.X, nil)
	return meta
}

func walkKeyValueExpr(e *ast.KeyValueExpr, ctx *evalContext) {
	// MYSTRUCT{key:value}
	// key is not an expression in struct literals.
	// Actually struct case is handled in walkCompositeLit().
	// In map, array or slice types, key can be an expression.
	// // map
	// expr := "hello"
	// a := map[string]int{expr: 0}
	// fmt.Println(a) // => map[hello:0]

	//const key = 1
	//s := []bool{key: true}
	//fmt.Println(s) // => map[hello:0]

	// const key = 1
	// s := []bool{key: true} // => [false true]

	//walkExpr(e.Key, nil)
	panic("TBI")
}

func walkInterfaceType(e *ast.InterfaceType) {
	// interface{}(e)  conversion. Nothing to do.
}

func walkTypeAssertExpr(e *ast.TypeAssertExpr, ctx *evalContext) *MetaTypeAssertExpr {
	meta := &MetaTypeAssertExpr{e: e}
	if ctx != nil && ctx.okContext {
		meta.NeedsOK = true
	}
	meta.X = walkExpr(e.X, nil)
	return meta
}

type MetaExpr interface{}

type MetaBasicLit struct {
	e             *ast.BasicLit
	Kind          string
	Value         string
	charValue     int
	intValue      int
	stringLiteral *sliteral
}

type MetaIdent struct {
	e          *ast.Ident
	kind       string // "blank|nil|true|false|var|con|fun|typ"
	Name       string
	conLiteral MetaExpr
	typ        *Type
}

type MetaSelectorExpr struct {
	e *ast.SelectorExpr
	X MetaExpr
}
type MetaSliceExpr struct {
	e    *ast.SliceExpr
	Low  MetaExpr
	High MetaExpr
	Max  MetaExpr
	X    MetaExpr
}
type MetaStarExpr struct {
	e *ast.StarExpr
	X MetaExpr
}
type MetaUnaryExpr struct {
	e *ast.UnaryExpr
	X MetaExpr
}
type MetaBinaryExpr struct {
	e  *ast.BinaryExpr
	Op string
	X  MetaExpr
	Y  MetaExpr
}

type MetaCompositLiteral struct {
	e    *ast.CompositeLit
	kind string // "struct", "array", "slice" // @TODO "map"

	// for struct
	typ          *Type                       // type of the composite
	strctEements []*MetaStructLiteralElement // for "struct"

	// for array or slice
	arrayType *ast.ArrayType
	len       int
	elmType   *Type
	metaElms  []MetaExpr
}

type MetaCallExpr struct {
	e            *ast.CallExpr
	isConversion bool

	// For Conversion
	toType *Type

	arg0     MetaExpr // For conversion, len, cap
	typeArg0 *Type
	arg1     MetaExpr
	arg2     MetaExpr

	// For funcall
	fun         ast.Expr
	hasEllipsis bool
	args        []ast.Expr

	builtin *ast.Object

	// general funcall
	funcType *ast.FuncType
	funcVal  *FuncValue
	receiver ast.Expr
	metaArgs []*Arg
}

type MetaIndexExpr struct {
	IsMap   bool // mp[k]
	NeedsOK bool // when map, is it ok syntax ?
	Index   MetaExpr
	X       MetaExpr
	e       *ast.IndexExpr
}

type MetaTypeAssertExpr struct {
	NeedsOK bool
	X       MetaExpr
	e       *ast.TypeAssertExpr
}

// ctx type is the type of someone who receives the expr value.
// There are various forms:
//
//	Assignment:       x = expr
//	Function call:    x(expr)
//	Return:           return expr
//	CompositeLiteral: T{key:expr}
//
// targetType is used when:
//   - the expr is nil
//   - the target type is interface and expr is not.
func walkExpr(expr ast.Expr, ctx *evalContext) MetaExpr {
	switch e := expr.(type) {
	case *ast.ParenExpr:
		return walkExpr(e.X, ctx)
	case *ast.BasicLit:
		return walkBasicLit(e, ctx)
	case *ast.KeyValueExpr:
		walkKeyValueExpr(e, ctx)
		return nil
	case *ast.CompositeLit:
		return walkCompositeLit(e, ctx)
	case *ast.Ident:
		return walkIdent(e, ctx)
	case *ast.SelectorExpr:
		return walkSelectorExpr(e, ctx)
	case *ast.CallExpr:
		return walkCallExpr(e, ctx)
	case *ast.IndexExpr:
		return walkIndexExpr(e, ctx)
	case *ast.SliceExpr:
		return walkSliceExpr(e, ctx)
	case *ast.StarExpr:
		return walkStarExpr(e, ctx)
	case *ast.UnaryExpr:
		return walkUnaryExpr(e, ctx)
	case *ast.BinaryExpr:
		return walkBinaryExpr(e, ctx)
	case *ast.TypeAssertExpr:
		return walkTypeAssertExpr(e, ctx)
	case *ast.ArrayType: // type
		walkArrayType(e) // []T(e)
		return nil
	case *ast.MapType: // type
		walkMapType(e)
		return nil
	case *ast.InterfaceType: // type
		walkInterfaceType(e)
		return nil
	case *ast.FuncType:
		return nil // @TODO walk
	default:
		panic(fmt.Sprintf("unknown type %T", expr))
	}
	panic("TBI")
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
	symbol   string
	funcType *ast.FuncType
}

func lookupForeignFunc(qi QualifiedIdent) *ForeignFunc {
	ident := lookupForeignIdent(qi)
	assert(ident.Obj.Kind == ast.Fun, "should be Fun", __func__)
	decl := ident.Obj.Decl.(*ast.FuncDecl)
	return &ForeignFunc{
		symbol:   string(qi),
		funcType: decl.Type,
	}
}

type Func struct {
	Name      string
	Stmts     []MetaStmt
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

func setVariable(obj *ast.Object, vr *Variable) {
	assert(obj.Kind == ast.Var, "obj is not  ast.Var", __func__)
	if vr == nil {
		obj.Data = nil
	} else {
		obj.Data = vr
	}
}

// Purpose of walk:
// - collect string literals
// - collect method declarations
// - collect global variables
// - collect local variables and set offset
// - determine struct size and field offset
// - determine types of variable declarations
// - attach type to universe nil
// - transmit ok sytanx context
// - (hope) attach type to untyped constants
// - (hope) transmit the need of interface conversion
func walk(pkg *PkgContainer) {
	var typeSpecs []*ast.TypeSpec
	var funcDecls []*ast.FuncDecl
	var varSpecs []*ast.ValueSpec
	var constSpecs []*ast.ValueSpec

	var exportedTpyes []*Type
	logf2("grouping declarations by type\n")
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

	logf2("checking typeSpecs...\n")
	for _, typeSpec := range typeSpecs {
		//@TODO check serializeType()'s *ast.Ident case
		typeSpec.Name.Obj.Data = pkg.name // package to which the type belongs to
		eType := &ast.Ident{
			Obj: &ast.Object{
				Kind: ast.Typ,
				Decl: typeSpec,
			},
		}
		t := e2t(eType)
		t.PkgName = pkg.name
		t.Name = typeSpec.Name.Name
		exportedTpyes = append(exportedTpyes, t)
		switch kind(t) {
		case T_STRUCT:
			structType := getUnderlyingType(t)
			calcStructSizeAndSetFieldOffset(structType.E.(*ast.StructType))
		}
		ExportedQualifiedIdents[string(newQI(pkg.name, typeSpec.Name.Name))] = typeSpec.Name
	}

	logf2("checking funcDecls...\n")

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

	logf2("walking constSpecs...\n")

	for _, constSpec := range constSpecs {
		for _, v := range constSpec.Values {
			walkExpr(v, nil) // @TODO: store meta
		}
	}

	logf2("walking varSpecs...\n")
	for _, varSpec := range varSpecs {
		nameIdent := varSpec.Names[0]
		assert(nameIdent.Obj.Kind == ast.Var, "should be Var", __func__)
		var t *Type
		if varSpec.Type == nil {
			// Infer type
			val := varSpec.Values[0]
			t = getTypeOfExpr(val)
			if t == nil {
				panic("variable type is not determined : " + nameIdent.Name)
			}
			varSpec.Type = t.E
		} else {
			walkExpr(varSpec.Type, nil)
			t = e2t(varSpec.Type)
		}
		variable := newGlobalVariable(pkg.name, nameIdent.Obj.Name, e2t(varSpec.Type))
		setVariable(nameIdent.Obj, variable)
		var val ast.Expr
		var metaVal MetaExpr
		if len(varSpec.Values) > 0 {
			val = varSpec.Values[0]
			// collect string literals
			metaVal = walkExpr(val, nil)
		}
		metaVar := walkIdent(nameIdent, nil)
		pkgVar := &packageVar{
			spec:    varSpec,
			name:    nameIdent,
			val:     val,
			metaVal: metaVal, // can be nil
			metaVar: metaVar,
			typ:     t,
		}
		pkg.vars = append(pkg.vars, pkgVar)
		ExportedQualifiedIdents[string(newQI(pkg.name, nameIdent.Name))] = nameIdent
	}

	logf2("walking funcDecls in detail ...\n")

	for _, funcDecl := range funcDecls {
		logf2("walking funcDecl \"%s\" \n", funcDecl.Name.Name)
		fnc := &Func{
			Name:      funcDecl.Name.Name,
			FuncType:  funcDecl.Type,
			Localarea: 0,
			Argsarea:  16, // return address + previous rbp
		}
		currentFunc = fnc

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
			var ms []MetaStmt
			for _, stmt := range funcDecl.Body.List {
				m := walkStmt(stmt)
				ms = append(ms, m)
			}
			fnc.Stmts = ms

			if funcDecl.Recv != nil { // is Method
				fnc.Method = newMethod(pkg.name, funcDecl)
			}
			pkg.funcs = append(pkg.funcs, fnc)
		}
		currentFunc = nil
	}

	printf("# Package types:\n")
	for _, typ := range exportedTpyes {
		printf("# type %s %s\n", serializeType(typ), serializeType(getUnderlyingType(typ)))
	}
}

// --- universe ---
var gNil = &ast.Object{
	Kind: ast.Con, // is nil a constant ?
	Name: "nil",
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
		gString, gUintptr, gBool, gInt, gUint8, gUint16, gInt32, gError,
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

type PackageToBuild struct {
	path  string
	name  string
	files []string
}

type packageVar struct {
	spec    *ast.ValueSpec
	name    *ast.Ident
	val     ast.Expr // can be nil
	metaVal MetaExpr // can be nil
	typ     *Type    // cannot be nil
	metaVar *MetaIdent
}

type PkgContainer struct {
	path           string
	name           string
	astFiles       []*ast.File
	vars           []*packageVar
	funcs          []*Func
	stringLiterals []*sliteral
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
		}
	}
}

// "some/dir" => []string{"a.go", "b.go"}
func findFilesInDir(dir string) []string {
	dirents, _ := mylib.Readdirnames(dir)
	var r []string
	for _, dirent := range dirents {
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
	return !strings.Contains(pth, ".")
}

func getImportPathsFromFile(file string) []string {
	fset := &token.FileSet{}
	astFile0 := parseImports(fset, file)
	var paths []string
	for _, importSpec := range astFile0.Imports {
		rawValue := importSpec.Path.Value
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
		srcFile := pkgDir + "/" + fname
		files = append(files, srcFile)
	}
	return files
}

const parserImportsOnly = 2 // parser.ImportsOnly

func parseImports(fset *token.FileSet, filename string) *ast.File {
	f, err := ParseFile(fset, filename, nil, parserImportsOnly)
	if err != nil {
		panic(filename + ":" + err.Error())
	}
	return f
}

func parseFile(fset *token.FileSet, filename string) *ast.File {
	f, err := ParseFile(fset, filename, nil, 0)
	if err != nil {
		panic(err.Error())
	}
	return f
}

// compile compiles go files of a package into an assembly file, and copy input assembly files into it.
func compile(universe *ast.Scope, path string, name string, gofiles []string, asmfiles []string, outFilePath string) {
	_pkg := &PkgContainer{name: name, path: path}
	currentPkg = _pkg

	outAsmFile, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	fout = outAsmFile

	typesMap = make(map[string]*dtypeEntry)
	typeId = 1

	logf("Building package : %s\n", _pkg.path)
	fset := &token.FileSet{}
	pkgScope := ast.NewScope(universe)
	for _, file := range gofiles {
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
			logf("resolving %s ...", ident.Name)
			obj := pkgScope.Lookup(ident.Name)
			if obj != nil {
				logf("  ===> obj found in pkg scope\n")
				ident.Obj = obj
			} else {
				obj := universe.Lookup(ident.Name)
				if obj != nil {
					logf("  ===> obj found in universe scope\n")
					ident.Obj = obj
				} else {
					logf("  ===> NOT FOUND\n")
					// we should allow unresolved for now.
					// e.g foo in X{foo:bar,}
					unresolved = append(unresolved, ident)
				}
			}
		}
		for _, dcl := range astFile.Decls {
			_pkg.Decls = append(_pkg.Decls, dcl)
		}
	}
	logf("Walking package: %s\n", _pkg.name)
	printf("#=== Package %s\n", _pkg.path)
	printf("#--- walk \n")
	walk(_pkg)
	generateCode(_pkg)

	// append static asm files
	for _, file := range asmfiles {
		fmt.Fprintf(outAsmFile, "# === static assembly %s ====\n", file)
		asmContents, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		outAsmFile.Write(asmContents)
	}

	outAsmFile.Close()
	fout = nil
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
		fmt.Printf("babygo version %s  linux/amd64\n", Version)
		return
	} else if os.Args[1] == "help" {
		showHelp()
		return
	} else if os.Args[1] == "panic" {
		panicVersion := strconv.Itoa(mylib.Sum(1, 1))
		panic("I am panic version " + panicVersion)
	}

	buildAll(os.Args[1:])
}

func buildAll(args []string) {
	workdir := os.Getenv("WORKDIR")
	if workdir == "" {
		workdir = "/tmp"
	}
	logf("Build start\n")

	var inputFiles []string
	for _, arg := range args {
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
	var packagesToBuild []*PackageToBuild
	for _, _path := range paths {
		files := collectSourceFiles(getPackageDir(_path))
		packagesToBuild = append(packagesToBuild, &PackageToBuild{
			name:  path.Base(_path),
			path:  _path,
			files: files,
		})
	}

	packagesToBuild = append(packagesToBuild, &PackageToBuild{
		name:  "main",
		path:  "main",
		files: inputFiles,
	})
	var universe = createUniverse()
	for _, _pkg := range packagesToBuild {
		if _pkg.name == "" {
			panic("empty pkg name")
		}
		var asmBasename []byte
		for _, ch := range []byte(_pkg.path) {
			if ch == '/' {
				ch = '@'
			}
			asmBasename = append(asmBasename, ch)
		}
		outFilePath := fmt.Sprintf("%s/%s", workdir, string(asmBasename)+".s")
		var gofiles []string
		var asmfiles []string
		for _, f := range _pkg.files {
			if strings.HasSuffix(f, ".go") {
				gofiles = append(gofiles, f)
			} else if strings.HasSuffix(f, ".s") {
				asmfiles = append(asmfiles, f)
			}

		}
		compile(universe, _pkg.path, _pkg.name, gofiles, asmfiles, outFilePath)
	}
}

// --- AST meta data ---
var mapFieldOffset = make(map[unsafe.Pointer]int)

func getStructFieldOffset(field *ast.Field) int {
	return mapFieldOffset[unsafe.Pointer(field)]
}

func setStructFieldOffset(field *ast.Field, offset int) {
	mapFieldOffset[unsafe.Pointer(field)] = offset
}

func throw(x interface{}) {
	panic(fmt.Sprintf(ThrowFormat, x))
}
