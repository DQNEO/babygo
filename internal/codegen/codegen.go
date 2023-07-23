package codegen

import (
	"os"

	"github.com/DQNEO/babygo/internal/ir"
	"github.com/DQNEO/babygo/internal/sema"
	"github.com/DQNEO/babygo/internal/types"
	"github.com/DQNEO/babygo/internal/universe"
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/strconv"
	"github.com/DQNEO/babygo/lib/token"
)

var Fout *os.File
var DebugCodeGen bool

var __func__ = "__func__"

func assert(bol bool, msg string, caller string) {
	if !bol {
		panic(sema.CurrentPkg.Name + ":" + caller + ": " + msg)
	}
}

const ThrowFormat string = "%T"

func throw(x interface{}) {
	panic(fmt.Sprintf(ThrowFormat, x))
}

func getLoc(pos token.Pos) string {
	posit := sema.Fset.Position(pos)
	fileno, ok := sema.CurrentPkg.FileNoMap[posit.Filename]
	if !ok {
		// Stmt or Expr in foreign package cannot be found in the map.
		return ""
	}
	return fmt.Sprintf("%d %d", fileno, posit.Line)
}

func printf(format string, a ...interface{}) {
	fmt.Fprintf(Fout, format, a...)
}

func unexpectedKind(knd types.TypeKind) {
	panic("Unexpected Kind: " + string(knd))
}

func emitComment(indent int, format string, a ...interface{}) {
	if !DebugCodeGen {
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
	i := sema.EvalInt(expr)
	printf("  pushq $%d # const number literal\n", i)
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

func emitPushStackTop(condType *types.Type, offset int, comment string) {
	switch sema.Kind(condType) {
	case types.T_STRING:
		printf("  movq %d+8(%%rsp), %%rcx # copy str.len from stack top (%s)\n", offset, comment)
		printf("  movq %d+0(%%rsp), %%rax # copy str.ptr from stack top (%s)\n", offset, comment)
		printf("  pushq %%rcx # str.len\n")
		printf("  pushq %%rax # str.ptr\n")
	case types.T_POINTER, types.T_UINTPTR, types.T_BOOL, types.T_INT, types.T_UINT8, types.T_UINT16:
		printf("  movq %d(%%rsp), %%rax # copy stack top value (%s) \n", offset, comment)
		printf("  pushq %%rax\n")
	default:
		unexpectedKind(sema.Kind(condType))
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
func emitLoadAndPush(t *types.Type) {
	assert(t != nil, "type should not be nil", __func__)
	emitPopAddress(string(sema.Kind(t)))
	switch sema.Kind(t) {
	case types.T_SLICE:
		printf("  movq %d(%%rax), %%rdx\n", 16)
		printf("  movq %d(%%rax), %%rcx\n", 8)
		printf("  movq %d(%%rax), %%rax\n", 0)
		printf("  pushq %%rdx # cap\n")
		printf("  pushq %%rcx # len\n")
		printf("  pushq %%rax # ptr\n")
	case types.T_STRING:
		printf("  movq %d(%%rax), %%rdx # len\n", 8)
		printf("  movq %d(%%rax), %%rax # ptr\n", 0)
		printf("  pushq %%rdx # len\n")
		printf("  pushq %%rax # ptr\n")
	case types.T_INTERFACE:
		printf("  movq %d(%%rax), %%rdx # data\n", 8)
		printf("  movq %d(%%rax), %%rax # dtype\n", 0)
		printf("  pushq %%rdx # data\n")
		printf("  pushq %%rax # dtype\n")
	case types.T_UINT8:
		printf("  movzbq %d(%%rax), %%rax # load uint8\n", 0)
		printf("  pushq %%rax\n")
	case types.T_UINT16:
		printf("  movzwq %d(%%rax), %%rax # load uint16\n", 0)
		printf("  pushq %%rax\n")
	case types.T_INT, types.T_BOOL:
		printf("  movq %d(%%rax), %%rax # load 64 bit\n", 0)
		printf("  pushq %%rax\n")
	case types.T_UINTPTR, types.T_POINTER, types.T_MAP, types.T_FUNC:
		printf("  movq %d(%%rax), %%rax # load 64 bit pointer\n", 0)
		printf("  pushq %%rax\n")
	case types.T_ARRAY, types.T_STRUCT:
		// pure proxy
		printf("  pushq %%rax\n")
	default:
		unexpectedKind(sema.Kind(t))
	}
}

func emitVariable(variable *ir.Variable) {
	emitVariableAddr(variable)
	emitLoadAndPush(variable.Typ)
}

func emitFuncAddr(funcQI ir.QualifiedIdent) {
	printf("  leaq %s(%%rip), %%rax # func addr\n", string(funcQI))
	printf("  pushq %%rax # func addr\n")
}

func emitVariableAddr(variable *ir.Variable) {
	emitComment(2, "emit Addr of variable \"%s\" \n", variable.Name)

	if variable.IsGlobal {
		printf("  leaq %s(%%rip), %%rax # global variable \"%s\"\n", variable.GlobalSymbol, variable.Name)
	} else {
		printf("  leaq %d(%%rbp), %%rax # local variable \"%s\"\n", variable.LocalOffset, variable.Name)
	}

	printf("  pushq %%rax # variable address\n")
}

func emitListHeadAddr(list ir.MetaExpr) {
	t := sema.GetTypeOfExpr(list)
	switch sema.Kind(t) {
	case types.T_ARRAY:
		emitAddr(list) // array head
	case types.T_SLICE:
		emitExpr(list)
		emitPopSlice()
		printf("  pushq %%rax # slice.ptr\n")
	case types.T_STRING:
		emitExpr(list)
		emitPopString()
		printf("  pushq %%rax # string.ptr\n")
	default:
		unexpectedKind(sema.Kind(t))
	}
}

func emitAddr(meta ir.MetaExpr) {
	emitComment(2, "[emitAddr] %T\n", meta)
	switch m := meta.(type) {
	case *ir.MetaIdent:
		switch m.Kind {
		case "var":
			emitVariableAddr(m.Variable)
		case "fun":
			qi := sema.NewQI(sema.CurrentPkg.Name, m.Name)
			emitFuncAddr(qi)
		default:
			panic("Unexpected kind")
		}
	case *ir.MetaIndexExpr:
		if sema.Kind(sema.GetTypeOfExpr(m.X)) == types.T_MAP {
			emitAddrForMapSet(m)
		} else {
			elmType := sema.GetTypeOfExpr(m)
			emitExpr(m.Index) // index number
			emitListElementAddr(m.X, elmType)
		}
	case *ir.MetaStarExpr:
		emitExpr(m.X)
	case *ir.MetaSelectorExpr:
		if m.IsQI { // pkg.Var|pkg.Const
			qi := m.QI
			ident := sema.LookupForeignIdent(qi)
			switch ident.Obj.Kind {
			case ast.Var:
				printf("  leaq %s(%%rip), %%rax # external global variable \n", string(qi))
				printf("  pushq %%rax # variable address\n")
			case ast.Fun:
				emitFuncAddr(qi)
			case ast.Con:
				panic("TBI 267")
			default:
				panic("Unexpected foreign ident kind:" + ident.Obj.Kind.String())
			}
		} else { // (e).field
			typeOfX := sema.GetUnderlyingType(sema.GetTypeOfExpr(m.X))
			var structTypeLiteral *ast.StructType
			switch typ := typeOfX.E.(type) {
			case *ast.StructType: // strct.field
				structTypeLiteral = typ
				emitAddr(m.X)
			case *ast.StarExpr: // ptr.field
				structTypeLiteral = sema.GetUnderlyingStructType(sema.E2T(typ.X))
				emitExpr(m.X)
			default:
				unexpectedKind(sema.Kind(typeOfX))
			}

			field := sema.LookupStructField(structTypeLiteral, m.SelName)
			offset := sema.GetStructFieldOffset(field)
			emitAddConst(offset, "struct head address + struct.field offset")
		}
	case *ir.MetaCompositLit:
		knd := sema.Kind(sema.GetTypeOfExpr(m))
		switch knd {
		case types.T_STRUCT:
			// result of evaluation of a struct literal is its address
			emitExpr(m)
		default:
			unexpectedKind(knd)
		}
	default:
		throw(meta)
	}
}

// explicit conversion T(e)
func emitConversion(toType *types.Type, arg0 ir.MetaExpr) {
	emitComment(2, "[emitConversion]\n")
	switch to := toType.E.(type) {
	case *ast.Ident:
		switch to.Obj {
		case universe.String: // string(e)
			switch sema.Kind(sema.GetTypeOfExpr(arg0)) {
			case types.T_SLICE: // string(slice)
				emitExpr(arg0) // slice
				emitPopSlice()
				printf("  pushq %%rcx # str len\n")
				printf("  pushq %%rax # str ptr\n")
			case types.T_STRING: // string(string)
				emitExpr(arg0)
			default:
				unexpectedKind(sema.Kind(sema.GetTypeOfExpr(arg0)))
			}
		case universe.Int, universe.Uint8, universe.Uint16, universe.Uintptr: // int(e)
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
		qi := sema.Selector2QI(to)
		ff := sema.LookupForeignIdent(qi)
		assert(ff.Obj.Kind == ast.Typ, "should be ast.Typ", __func__)
		emitConversion(sema.E2T(ff), arg0)
	case *ast.ArrayType: // Conversion to slice
		arrayType := to
		if arrayType.Len != nil {
			throw(to)
		}
		assert(sema.Kind(sema.GetTypeOfExpr(arg0)) == types.T_STRING, "source type should be slice", __func__)
		emitComment(2, "Conversion of string => slice \n")
		emitExpr(arg0)
		emitPopString()
		printf("  pushq %%rcx # cap\n")
		printf("  pushq %%rcx # len\n")
		printf("  pushq %%rax # ptr\n")
	case *ast.ParenExpr: // (T)(arg0)
		emitConversion(sema.E2T(to.X), arg0)
	case *ast.StarExpr: // (*T)(arg0)
		emitExpr(arg0)
	case *ast.InterfaceType:
		emitExpr(arg0)
		if sema.IsInterface(sema.GetTypeOfExpr(arg0)) {
			// do nothing
		} else {
			// Convert dynamic value to interface
			emitConvertToInterface(sema.GetTypeOfExpr(arg0))
		}
	default:
		throw(to)
	}
}

func emitZeroValue(t *types.Type) {
	switch sema.Kind(t) {
	case types.T_SLICE:
		printf("  pushq $0 # slice cap\n")
		printf("  pushq $0 # slice len\n")
		printf("  pushq $0 # slice ptr\n")
	case types.T_STRING:
		printf("  pushq $0 # string len\n")
		printf("  pushq $0 # string ptr\n")
	case types.T_INTERFACE:
		printf("  pushq $0 # interface data\n")
		printf("  pushq $0 # interface dtype\n")
	case types.T_INT, types.T_UINT8, types.T_BOOL:
		printf("  pushq $0 # %s zero value (number)\n", string(sema.Kind(t)))
	case types.T_UINTPTR, types.T_POINTER, types.T_MAP, types.T_FUNC:
		printf("  pushq $0 # %s zero value (nil pointer)\n", string(sema.Kind(t)))
	case types.T_ARRAY:
		size := sema.GetSizeOfType(t)
		emitComment(2, "zero value of an array. size=%d (allocating on heap)\n", size)
		emitCallMalloc(size)
	case types.T_STRUCT:
		structSize := sema.GetSizeOfType(t)
		emitComment(2, "zero value of a struct. size=%d (allocating on heap)\n", structSize)
		emitCallMalloc(structSize)
	default:
		unexpectedKind(sema.Kind(t))
	}
}

func emitLen(arg ir.MetaExpr) {
	switch sema.Kind(sema.GetTypeOfExpr(arg)) {
	case types.T_ARRAY:
		arrayType := sema.GetTypeOfExpr(arg).E.(*ast.ArrayType)
		emitConstInt(arrayType.Len)
	case types.T_SLICE:
		emitExpr(arg)
		emitPopSlice()
		printf("  pushq %%rcx # len\n")
	case types.T_STRING:
		emitExpr(arg)
		emitPopString()
		printf("  pushq %%rcx # len\n")
	case types.T_MAP:
		args := []*ir.MetaArg{
			// len
			&ir.MetaArg{
				Meta:      arg,
				ParamType: sema.GetTypeOfExpr(arg),
			},
		}
		resultList := &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Type: types.Int.E,
				},
			},
		}
		emitCallDirect("runtime.lenMap", args, resultList)

	default:
		unexpectedKind(sema.Kind(sema.GetTypeOfExpr(arg)))
	}
}

func emitCap(arg ir.MetaExpr) {
	switch sema.Kind(sema.GetTypeOfExpr(arg)) {
	case types.T_ARRAY:
		arrayType := sema.GetTypeOfExpr(arg).E.(*ast.ArrayType)
		emitConstInt(arrayType.Len)
	case types.T_SLICE:
		emitExpr(arg)
		emitPopSlice()
		printf("  pushq %%rdx # cap\n")
	case types.T_STRING:
		panic("cap() cannot accept string type")
	default:
		unexpectedKind(sema.Kind(sema.GetTypeOfExpr(arg)))
	}
}

func emitCallMalloc(size int) {
	// call malloc and return pointer
	ff := sema.LookupForeignFunc(sema.NewQI("runtime", "malloc"))
	emitAllocReturnVarsAreaFF(ff)
	printf("  pushq $%d\n", size)
	emitCallFF(ff)
}

func emitStructLiteral(meta *ir.MetaCompositLit) {
	// allocate heap area with zero value
	emitComment(2, "emitStructLiteral\n")
	structType := meta.Type
	emitZeroValue(structType) // push address of the new storage
	metaElms := meta.StructElements
	for _, metaElm := range metaElms {
		// push lhs address
		emitPushStackTop(types.Uintptr, 0, "address of struct heaad")

		fieldOffset := sema.GetStructFieldOffset(metaElm.Field)
		emitAddConst(fieldOffset, "address of struct field")

		// push rhs value
		emitExpr(metaElm.ValueMeta)
		mayEmitConvertTooIfc(metaElm.ValueMeta, metaElm.FieldType)

		// assign
		emitStore(metaElm.FieldType, true, false)
	}
}

func emitArrayLiteral(meta *ir.MetaCompositLit) {
	elmType := meta.ElmType
	elmSize := sema.GetSizeOfType(elmType)
	memSize := elmSize * meta.Len

	emitCallMalloc(memSize) // push
	for i, elm := range meta.MetaElms {
		// push lhs address
		emitPushStackTop(types.Uintptr, 0, "malloced address")
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

func emitCallDirect(symbol string, args []*ir.MetaArg, resultList *ast.FieldList) {
	returnTypes := sema.FieldList2Types(resultList)
	emitCall(sema.NewFuncValueFromSymbol(symbol), args, returnTypes)
}

// see "ABI of stack layout" in the emitFuncall comment
func emitCall(fv *ir.FuncValue, args []*ir.MetaArg, returnTypes []*types.Type) {
	emitComment(2, "emitCall len(args)=%d\n", len(args))
	var totalParamSize int
	var offsets []int
	for _, arg := range args {
		offsets = append(offsets, totalParamSize)
		if arg.ParamType == nil {
			panic("ParamType must not be nil")
		}
		totalParamSize += sema.GetSizeOfType(arg.ParamType)
	}

	emitAllocReturnVarsArea(getTotalSizeOfType(returnTypes))
	printf("  subq $%d, %%rsp # alloc parameters area\n", totalParamSize)
	for i, arg := range args {
		paramType := arg.ParamType
		if arg.Meta == nil {
			panic("arg.meta should not be nil")
		}
		emitExpr(arg.Meta)
		mayEmitConvertTooIfc(arg.Meta, paramType)
		emitPop(sema.Kind(paramType))
		printf("  leaq %d(%%rsp), %%rsi # place to save\n", offsets[i])
		printf("  pushq %%rsi # place to save\n")
		emitRegiToMem(paramType)
	}

	emitCallQ(fv, totalParamSize, returnTypes)
}

func emitAllocReturnVarsAreaFF(ff *ir.ForeignFunc) {
	emitAllocReturnVarsArea(getTotalFieldsSize(ff.FuncType.Results))
}

func getTotalSizeOfType(types []*types.Type) int {
	var r int
	for _, t := range types {
		r += sema.GetSizeOfType(t)
	}
	return r
}

func getTotalFieldsSize(flist *ast.FieldList) int {
	if flist == nil {
		return 0
	}
	var r int
	for _, fld := range flist.List {
		r += sema.GetSizeOfType(sema.E2T(fld.Type))
	}
	return r
}

func emitCallFF(ff *ir.ForeignFunc) {
	totalParamSize := getTotalFieldsSize(ff.FuncType.Params)
	returnTypes := sema.FieldList2Types(ff.FuncType.Results)
	emitCallQ(sema.NewFuncValueFromSymbol(ff.Symbol), totalParamSize, returnTypes)
}

func emitCallQ(fv *ir.FuncValue, totalParamSize int, returnTypes []*types.Type) {
	if fv.IsDirect {
		if fv.Symbol == "" {
			panic("callq target must not be empty")
		}
		printf("  callq %s\n", fv.Symbol)
	} else {
		emitExpr(fv.Expr)
		printf("  popq %%rax\n")
		printf("  callq *%%rax\n")
	}

	emitFreeParametersArea(totalParamSize)
	printf("  #  totalReturnSize=%d\n", getTotalSizeOfType(returnTypes))
	emitFreeAndPushReturnedValue(returnTypes)
}

// callee
func emitReturnStmt(meta *ir.MetaReturnStmt) {
	funcDef := meta.Fnc
	if len(funcDef.Retvars) != len(meta.Results) {
		panic("length of return and func type do not match")
	}

	_len := len(meta.Results)
	for i := 0; i < _len; i++ {
		emitAssignToVar(funcDef.Retvars[i], meta.Results[i])
	}
	printf("  leave\n")
	printf("  ret\n")
}

// caller
func emitFreeAndPushReturnedValue(returnTypes []*types.Type) {
	switch len(returnTypes) {
	case 0:
		// do nothing
	case 1:
		knd := sema.Kind(returnTypes[0])
		switch knd {
		case types.T_STRING, types.T_INTERFACE:
		case types.T_UINT8:
			printf("  movzbq (%%rsp), %%rax # load uint8\n")
			printf("  addq $%d, %%rsp # free returnvars area\n", 1)
			printf("  pushq %%rax\n")
		case types.T_BOOL, types.T_INT, types.T_UINTPTR, types.T_POINTER, types.T_MAP:
		case types.T_SLICE:
		default:
			unexpectedKind(knd)
		}
	default:
		//panic("TBI")
	}
}

func emitBuiltinFunCall(obj *ast.Object, typeArg0 *types.Type, arg0 ir.MetaExpr, arg1 ir.MetaExpr, arg2 ir.MetaExpr) {
	switch obj {
	case universe.Len:
		emitLen(arg0)
		return
	case universe.Cap:
		emitCap(arg0)
		return
	case universe.New:
		// size to malloc
		size := sema.GetSizeOfType(typeArg0)
		emitCallMalloc(size)
		return
	case universe.Make:
		typeArg := typeArg0
		switch sema.Kind(typeArg) {
		case types.T_MAP:
			mapType := sema.GetUnderlyingType(typeArg).E.(*ast.MapType)
			valueSize := sema.NewNumberLiteral(sema.GetSizeOfType(sema.E2T(mapType.Value)))
			// A new, empty map value is made using the built-in function make,
			// which takes the map type and an optional capacity hint as arguments:
			length := sema.NewNumberLiteral(0)
			args := []*ir.MetaArg{
				&ir.MetaArg{
					Meta:      length,
					ParamType: types.Uintptr,
				},
				&ir.MetaArg{
					Meta:      valueSize,
					ParamType: types.Uintptr,
				},
			}
			resultList := &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: types.Uintptr.E,
					},
				},
			}
			emitCallDirect("runtime.makeMap", args, resultList)
			return
		case types.T_SLICE:
			// make([]T, ...)
			arrayType := sema.GetUnderlyingType(typeArg).E.(*ast.ArrayType)
			elmSize := sema.GetSizeOfType(sema.E2T(arrayType.Elt))
			numlit := sema.NewNumberLiteral(elmSize)
			args := []*ir.MetaArg{
				// elmSize
				&ir.MetaArg{
					Meta:      numlit,
					ParamType: types.Int,
				},
				// len
				&ir.MetaArg{
					Meta:      arg1,
					ParamType: types.Int,
				},
				// cap
				&ir.MetaArg{
					Meta:      arg2,
					ParamType: types.Int,
				},
			}

			resultList := &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: sema.GeneralSlice,
					},
				},
			}
			emitCallDirect("runtime.makeSlice", args, resultList)
			return
		default:
			throw(typeArg)
		}
	case universe.Append:
		sliceArg := arg0
		elemArg := arg1
		elmType := sema.GetElementTypeOfCollectionType(sema.GetTypeOfExpr(sliceArg))
		elmSize := sema.GetSizeOfType(elmType)
		args := []*ir.MetaArg{
			// slice
			&ir.MetaArg{
				Meta:      sliceArg,
				ParamType: sema.E2T(sema.GeneralSlice),
			},
			// elm
			&ir.MetaArg{
				Meta:      elemArg,
				ParamType: elmType,
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
					Type: sema.GeneralSlice,
				},
			},
		}
		emitCallDirect(symbol, args, resultList)
		return
	case universe.Panic:
		funcVal := "runtime.panic"
		_args := []*ir.MetaArg{&ir.MetaArg{
			Meta:      arg0,
			ParamType: types.Eface,
		}}
		emitCallDirect(funcVal, _args, nil)
		return
	case universe.Delete:
		funcVal := "runtime.deleteMap"
		_args := []*ir.MetaArg{
			&ir.MetaArg{
				Meta:      arg0,
				ParamType: sema.GetTypeOfExpr(arg0),
			},
			&ir.MetaArg{
				Meta:      arg1,
				ParamType: types.Eface,
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
func emitFuncall(meta *ir.MetaCallExpr) {
	emitComment(2, "[emitFuncall]\n")
	// check if it's a builtin func
	if meta.Builtin != nil {
		emitBuiltinFunCall(meta.Builtin, meta.TypeArg0, meta.Arg0, meta.Arg1, meta.Arg2)
		return
	}
	emitCall(meta.FuncVal, meta.MetaArgs, meta.Types)
}

// 1 value
func emitIdent(meta *ir.MetaIdent) {
	//logf("emitIdent ident=%s\n", meta.Name)
	switch meta.Kind {
	case "true": // true constant
		emitTrue()
	case "false": // false constant
		emitFalse()
	case "nil": // zero value
		metaType := meta.Type
		if metaType == nil {
			//gofmt.Fprintf(os.Stderr, "exprTypeMeta=%v\n", exprTypeMeta)
			panic("untyped nil is not allowed. Probably the type is not set in walk phase. pkg=" + sema.CurrentPkg.Name)
		}
		// emit zero value of the type
		switch sema.Kind(metaType) {
		case types.T_SLICE, types.T_POINTER, types.T_INTERFACE, types.T_MAP:
			emitZeroValue(metaType)
		default:
			unexpectedKind(sema.Kind(metaType))
		}
	case "var":
		emitAddr(meta)
		emitLoadAndPush(sema.GetTypeOfExpr(meta))
	case "con":
		if meta.Const.IsGlobal && sema.Kind(meta.Type) == types.T_STRING {
			// Treat like a global variable.
			// emit addr
			printf("  leaq %s(%%rip), %%rax # global const \"%s\"\n", meta.Const.GlobalSymbol, meta.Const.Name)
			printf("  pushq %%rax # global const address\n")

			// load and push
			emitLoadAndPush(meta.Type)
		} else {
			// emit literal directly
			emitBasicLit(meta.Const.Literal)
		}
	case "fun":
		emitAddr(meta)
	default:
		panic("Unexpected ident kind:" + meta.Kind + " name=" + meta.Name)
	}
}

// 1 or 2 values
func emitIndexExpr(meta *ir.MetaIndexExpr) {
	if meta.IsMap {
		emitMapGet(meta, meta.NeedsOK)
	} else {
		emitAddr(meta)
		emitLoadAndPush(sema.GetTypeOfExpr(meta))
	}
}

// 1 value
func emitStarExpr(meta *ir.MetaStarExpr) {
	emitAddr(meta)
	emitLoadAndPush(sema.GetTypeOfExpr(meta))
}

// 1 value X.Sel
func emitSelectorExpr(meta *ir.MetaSelectorExpr) {
	// pkg.Ident or strct.field
	if meta.IsQI {
		qi := meta.QI
		ident := sema.LookupForeignIdent(qi)
		switch ident.Obj.Kind {
		case ast.Fun:
			emitFuncAddr(qi)
		case ast.Var:
			m := sema.WalkIdent(ident, nil)
			emitExpr(m)
		case ast.Con:
			m := sema.WalkIdent(ident, nil)
			emitExpr(m)
		}
	} else {
		// strct.field
		emitAddr(meta)
		emitLoadAndPush(sema.GetTypeOfExpr(meta))
	}
}

// multi values Fun(Args)
func emitCallExpr(meta *ir.MetaCallExpr) {
	// check if it's a conversion
	if meta.IsConversion {
		emitComment(2, "[emitCallExpr] Conversion\n")
		emitConversion(meta.ToType, meta.Arg0)
	} else {
		emitComment(2, "[emitCallExpr] Funcall\n")
		emitFuncall(meta)
	}
}

// 1 value
func emitBasicLit(mt *ir.MetaBasicLit) {
	switch mt.Kind {
	case "CHAR":
		printf("  pushq $%d # convert char literal to int\n", mt.CharVal)
	case "INT":
		printf("  pushq $%d # number literal\n", mt.IntVal)
	case "STRING":
		sl := mt.StrVal
		if sl.Strlen == 0 {
			// zero value
			emitZeroValue(types.String)
		} else {
			printf("  pushq $%d # str len\n", sl.Strlen)
			printf("  leaq %s(%%rip), %%rax # str ptr\n", sl.Label)
			printf("  pushq %%rax # str ptr\n")
		}
	default:
		panic("Unexpected literal kind:" + mt.Kind)
	}
}

// 1 value
func emitUnaryExpr(meta *ir.MetaUnaryExpr) {
	switch meta.Op {
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
		panic(meta.Op)
	}
}

// 1 value
func emitBinaryExpr(meta *ir.MetaBinaryExpr) {
	switch meta.Op {
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
		if sema.Kind(sema.GetTypeOfExpr(meta.X)) == types.T_STRING {
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
		panic(meta.Op)
	}
}

// 1 value
func emitCompositeLit(meta *ir.MetaCompositLit) {
	// slice , array, map or struct
	switch meta.Kind {
	case "struct":
		emitStructLiteral(meta)
	case "array":
		emitArrayLiteral(meta)
	case "slice":
		emitArrayLiteral(meta)
		emitPopAddress("malloc")
		printf("  pushq $%d # slice.cap\n", meta.Len)
		printf("  pushq $%d # slice.len\n", meta.Len)
		printf("  pushq %%rax # slice.ptr\n")
	default:
		panic(meta.Kind)
	}
}

// 1 value list[low:high]
func emitSliceExpr(meta *ir.MetaSliceExpr) {
	list := meta.X
	listType := sema.GetTypeOfExpr(list)

	switch sema.Kind(listType) {
	case types.T_SLICE, types.T_ARRAY:
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
	case types.T_STRING:
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
		unexpectedKind(sema.Kind(listType))
	}

	emitExpr(meta.Low) // index number
	elmType := sema.GetElementTypeOfCollectionType(listType)
	emitListElementAddr(list, elmType)
}

// 1 or 2 values
func emitMapGet(m *ir.MetaIndexExpr, okContext bool) {
	valueType := sema.GetTypeOfExpr(m)

	emitComment(2, "MAP GET for map[string]string\n")
	// emit addr of map element
	mp := m.X
	key := m.Index

	args := []*ir.MetaArg{
		&ir.MetaArg{
			Meta:      mp,
			ParamType: types.Uintptr,
		},
		&ir.MetaArg{
			Meta:      key,
			ParamType: types.Eface,
		},
	}
	resultList := &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type: types.Bool.E,
			},
			&ast.Field{
				Type: types.Uintptr.E,
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
	emitPop(types.T_POINTER) // destroy nil
	emitZeroValue(valueType)
	if okContext {
		printf("  pushq $0 # ok = false\n")
	}

	printf("  %s:\n", labelEnd)
}

// 1 or 2 values
func emitTypeAssertExpr(meta *ir.MetaTypeAssertExpr) {
	emitExpr(meta.X)
	emitDtypeLabelAddr(meta.Type)
	emitCompareDtypes()

	emitPopBool("type assertion ok value")
	printf("  cmpq $1, %%rax\n")

	labelid++
	labelEnd := fmt.Sprintf(".L.end_type_assertion.%d", labelid)
	labelElse := fmt.Sprintf(".L.unmatch.%d", labelid)
	printf("  jne %s # jmp if false\n", labelElse)

	// if matched
	emitLoadAndPush(meta.Type) // load dynamic data
	if meta.NeedsOK {
		printf("  pushq $1 # ok = true\n")
	}
	// exit
	printf("  jmp %s\n", labelEnd)
	// if not matched
	printf("  %s:\n", labelElse)
	printf("  popq %%rax # drop ifc.data\n")
	emitZeroValue(meta.Type)
	if meta.NeedsOK {
		printf("  pushq $0 # ok = false\n")
	}

	printf("  %s:\n", labelEnd)
}

func emitExpr(meta ir.MetaExpr) {
	loc := getLoc(sema.Pos(meta))
	if loc == "" {
		printf("  # noloc %T\n", meta)
	} else {
		printf("  .loc %s # %T\n", loc, meta)
	}
	switch m := meta.(type) {
	case *ir.MetaBasicLit:
		emitBasicLit(m)
	case *ir.MetaCompositLit:
		emitCompositeLit(m)
	case *ir.MetaIdent:
		emitIdent(m)
	case *ir.MetaSelectorExpr:
		emitSelectorExpr(m)
	case *ir.MetaCallExpr:
		emitCallExpr(m) // can be Tuple
	case *ir.MetaIndexExpr:
		emitIndexExpr(m) // can be Tuple
	case *ir.MetaSliceExpr:
		emitSliceExpr(m)
	case *ir.MetaStarExpr:
		emitStarExpr(m)
	case *ir.MetaUnaryExpr:
		emitUnaryExpr(m)
	case *ir.MetaBinaryExpr:
		emitBinaryExpr(m)
	case *ir.MetaTypeAssertExpr:
		emitTypeAssertExpr(m) // can be Tuple
	default:
		panic(fmt.Sprintf("meta type:%T", meta))
	}
}

// convert stack top value to interface
func emitConvertToInterface(fromType *types.Type) {
	emitComment(2, "ConversionToInterface\n")
	memSize := sema.GetSizeOfType(fromType)
	// copy data to heap
	emitCallMalloc(memSize)
	emitStore(fromType, false, true) // heap addr pushed
	// push dtype label's address
	emitDtypeLabelAddr(fromType)
}

func mayEmitConvertTooIfc(meta ir.MetaExpr, ctxType *types.Type) {
	if !sema.IsNil(meta) && ctxType != nil && sema.IsInterface(ctxType) && !sema.IsInterface(sema.GetTypeOfExpr(meta)) {
		emitConvertToInterface(sema.GetTypeOfExpr(meta))
	}
}

type DtypeEntry struct {
	id         int
	serialized string
	label      string
}

var TypeId int
var TypesMap map[string]*DtypeEntry

// "**[1][]*int" => "dtype.8"
func getDtypeLabel(serializedType string) string {
	s := serializedType
	ent, ok := TypesMap[s]
	if ok {
		return ent.label
	} else {
		id := TypeId
		ent = &DtypeEntry{
			id:         id,
			serialized: serializedType,
			label:      "." + "dtype." + strconv.Itoa(id),
		}
		TypesMap[s] = ent
		TypeId++
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
	emitAllocReturnVarsArea(sema.SizeOfInt) // for bool

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

func emitDtypeLabelAddr(t *types.Type) {
	serializedType := serializeType(t)
	dtypeLabel := getDtypeLabel(serializedType)
	printf("  leaq %s(%%rip), %%rax # dtype label address \"%s\"\n", dtypeLabel, serializedType)
	printf("  pushq %%rax           # dtype label address\n")
}

func emitAddrForMapSet(indexExpr *ir.MetaIndexExpr) {
	// alloc heap for map value
	//size := GetSizeOfType(elmType)
	emitComment(2, "[emitAddrForMapSet]\n")
	mp := indexExpr.X
	key := indexExpr.Index

	args := []*ir.MetaArg{
		&ir.MetaArg{
			Meta:      mp,
			ParamType: types.Uintptr,
		},
		&ir.MetaArg{
			Meta:      key,
			ParamType: types.Eface,
		},
	}
	resultList := &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type: types.Uintptr.E,
			},
		},
	}
	emitCallDirect("runtime.getAddrForMapSet", args, resultList)
}

func emitListElementAddr(list ir.MetaExpr, elmType *types.Type) {
	emitListHeadAddr(list)
	emitPopAddress("list head")
	printf("  popq %%rcx # index id\n")
	printf("  movq $%d, %%rdx # elm size\n", sema.GetSizeOfType(elmType))
	printf("  imulq %%rdx, %%rcx\n")
	printf("  addq %%rcx, %%rax\n")
	printf("  pushq %%rax # addr of element\n")
}

func emitCatStrings(left ir.MetaExpr, right ir.MetaExpr) {
	args := []*ir.MetaArg{
		&ir.MetaArg{
			Meta:      left,
			ParamType: types.String,
		},
		&ir.MetaArg{
			Meta:      right,
			ParamType: types.String,
		},
	}
	resultList := &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type: types.String.E,
			},
		},
	}
	emitCallDirect("runtime.catstrings", args, resultList)
}

func emitCompStrings(left ir.MetaExpr, right ir.MetaExpr) {
	args := []*ir.MetaArg{
		&ir.MetaArg{
			Meta:      left,
			ParamType: types.String,
		},
		&ir.MetaArg{
			Meta:      right,
			ParamType: types.String,
		},
	}
	resultList := &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type: types.Bool.E,
			},
		},
	}
	emitCallDirect("runtime.cmpstrings", args, resultList)
}

func emitBinaryExprComparison(left ir.MetaExpr, right ir.MetaExpr) {
	if sema.Kind(sema.GetTypeOfExpr(left)) == types.T_STRING {
		emitCompStrings(left, right)
	} else if sema.Kind(sema.GetTypeOfExpr(left)) == types.T_INTERFACE {
		//var t = GetTypeOfExpr(left)
		ff := sema.LookupForeignFunc(sema.NewQI("runtime", "cmpinterface"))
		emitAllocReturnVarsAreaFF(ff)
		//@TODO: confirm nil comparison with interfaces
		emitExpr(left)  // left
		emitExpr(right) // right
		emitCallFF(ff)
	} else {
		// Assuming 64 bit types (int, pointer, map, etc)
		//var t = GetTypeOfExpr(left)
		emitExpr(left)  // left
		emitExpr(right) // right
		emitCompExpr("sete")

		//@TODO: slice <=> nil comparison works accidentally, but stack needs to be popped 2 times more.
	}

}

// @TODO handle larger Types than int
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

func emitPop(knd types.TypeKind) {
	switch knd {
	case types.T_SLICE:
		emitPopSlice()
	case types.T_STRING:
		emitPopString()
	case types.T_INTERFACE:
		emitPopInterFace()
	case types.T_INT, types.T_BOOL:
		emitPopPrimitive(string(knd))
	case types.T_UINTPTR, types.T_POINTER, types.T_MAP, types.T_FUNC:
		emitPopPrimitive(string(knd))
	case types.T_UINT16:
		emitPopPrimitive(string(knd))
	case types.T_UINT8:
		emitPopPrimitive(string(knd))
	case types.T_STRUCT, types.T_ARRAY:
		emitPopPrimitive(string(knd))
	default:
		unexpectedKind(knd)
	}
}

func emitStore(t *types.Type, rhsTop bool, pushLhs bool) {
	knd := sema.Kind(t)
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

func emitRegiToMem(t *types.Type) {
	printf("  popq %%rsi # place to save\n")
	k := sema.Kind(t)
	switch k {
	case types.T_SLICE:
		printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", 0)
		printf("  movq %%rcx, %d(%%rsi) # len to len\n", 8)
		printf("  movq %%rdx, %d(%%rsi) # cap to cap\n", 16)
	case types.T_STRING:
		printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", 0)
		printf("  movq %%rcx, %d(%%rsi) # len to len\n", 8)
	case types.T_INTERFACE:
		printf("  movq %%rax, %d(%%rsi) # store dtype\n", 0)
		printf("  movq %%rcx, %d(%%rsi) # store data\n", 8)
	case types.T_INT, types.T_BOOL:
		printf("  movq %%rax, %d(%%rsi) # assign quad\n", 0)
	case types.T_UINTPTR, types.T_POINTER, types.T_MAP, types.T_FUNC:
		printf("  movq %%rax, %d(%%rsi) # assign ptr\n", 0)
	case types.T_UINT16:
		printf("  movw %%ax, %d(%%rsi) # assign word\n", 0)
	case types.T_UINT8:
		printf("  movb %%al, %d(%%rsi) # assign byte\n", 0)
	case types.T_STRUCT, types.T_ARRAY:
		printf("  pushq $%d # size\n", sema.GetSizeOfType(t))
		printf("  pushq %%rsi # dst lhs\n")
		printf("  pushq %%rax # src rhs\n")
		ff := sema.LookupForeignFunc(sema.NewQI("runtime", "memcopy"))
		emitCallFF(ff)
	default:
		unexpectedKind(k)
	}
}

func emitAssignToVar(vr *ir.Variable, rhs ir.MetaExpr) {
	emitComment(2, "Assignment: emitVariableAddr(lhs)\n")
	emitVariableAddr(vr)
	emitComment(2, "Assignment: emitExpr(rhs)\n")

	emitExpr(rhs)
	mayEmitConvertTooIfc(rhs, vr.Typ)
	emitComment(2, "Assignment: emitStore(typeof(lhs))\n")
	emitStore(vr.Typ, true, false)
}

func emitAssignZeroValue(lhs ir.MetaExpr, lhsType *types.Type) {
	emitComment(2, "emitAssignZeroValue\n")
	emitComment(2, "lhs addresss\n")
	emitAddr(lhs)
	emitComment(2, "emitZeroValue\n")
	emitZeroValue(lhsType)
	emitStore(lhsType, true, false)

}

func emitSingleAssign(lhs ir.MetaExpr, rhs ir.MetaExpr) {
	//	lhs := metaSingle.lhs
	//	rhs := metaSingle.rhs
	if sema.IsBlankIdentifierMeta(lhs) {
		emitExpr(rhs)
		emitPop(sema.Kind(sema.GetTypeOfExpr(rhs)))
		return
	}
	emitComment(2, "Assignment: emitAddr(lhs)\n")
	emitAddr(lhs)
	emitComment(2, "Assignment: emitExpr(rhs)\n")
	emitExpr(rhs)
	mayEmitConvertTooIfc(rhs, sema.GetTypeOfExpr(lhs))
	emitStore(sema.GetTypeOfExpr(lhs), true, false)
}

func emitBlockStmt(s *ir.MetaBlockStmt) {
	for _, s := range s.List {
		emitStmt(s)
	}
}

func emitExprStmt(s *ir.MetaExprStmt) {
	emitExpr(s.X)
}

// local decl stmt
func emitDeclStmt(meta *ir.MetaVarDecl) {
	if meta.Single.Rhs == nil {
		// Assign zero value to LHS
		emitAssignZeroValue(meta.Single.Lhs, meta.LhsType)
	} else {
		emitSingleAssign(meta.Single.Lhs, meta.Single.Rhs)
	}
}

// a, b = OkExpr
func emitOkAssignment(meta *ir.MetaTupleAssign) {
	rhs0 := meta.Rhs
	rhsTypes := meta.RhsTypes

	// ABI of stack layout after evaluating rhs0
	//	-- stack top
	//	bool
	//	data
	emitExpr(rhs0)
	for i := 1; i >= 0; i-- {
		lhsMeta := meta.Lhss[i]
		rhsType := rhsTypes[i]
		if sema.IsBlankIdentifierMeta(lhsMeta) {
			emitPop(sema.Kind(rhsType))
		} else {
			// @TODO interface conversion
			emitAddr(lhsMeta)
			emitStore(sema.GetTypeOfExpr(lhsMeta), false, false)
		}

	}
}

// assigns multi-values of a funcall
// a, b (, c...) = f()
func emitFuncallAssignment(meta *ir.MetaTupleAssign) {
	rhs0 := meta.Rhs
	rhsTypes := meta.RhsTypes

	emitExpr(rhs0)
	for i := 0; i < len(rhsTypes); i++ {
		lhsMeta := meta.Lhss[i]
		rhsType := rhsTypes[i]
		if sema.IsBlankIdentifierMeta(lhsMeta) {
			emitPop(sema.Kind(rhsType))
		} else {
			switch sema.Kind(rhsType) {
			case types.T_UINT8:
				// repush stack top
				printf("  movzbq (%%rsp), %%rax # load uint8\n")
				printf("  addq $%d, %%rsp # free returnvars area\n", 1)
				printf("  pushq %%rax\n")
			}
			// @TODO interface conversion
			emitAddr(lhsMeta)
			emitStore(sema.GetTypeOfExpr(lhsMeta), false, false)
		}
	}
}

func emitIfStmt(meta *ir.MetaIfStmt) {
	emitComment(2, "if\n")

	labelid++
	labelEndif := fmt.Sprintf(".L.endif.%d", labelid)
	labelElse := fmt.Sprintf(".L.else.%d", labelid)

	emitExpr(meta.Cond)
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

func emitForStmt(meta *ir.MetaForContainer) {
	labelid++
	labelCond := fmt.Sprintf(".L.for.cond.%d", labelid)
	labelPost := fmt.Sprintf(".L.for.post.%d", labelid)
	labelExit := fmt.Sprintf(".L.for.exit.%d", labelid)

	meta.LabelPost = labelPost
	meta.LabelExit = labelExit

	if meta.ForStmt.Init != nil {
		emitStmt(meta.ForStmt.Init)
	}

	printf("  %s:\n", labelCond)
	if meta.ForStmt.Cond != nil {
		emitExpr(meta.ForStmt.Cond)
		emitPopBool("for condition")
		printf("  cmpq $1, %%rax\n")
		printf("  jne %s # jmp if false\n", labelExit)
	}
	emitBlockStmt(meta.Body)
	printf("  %s:\n", labelPost) // used for "continue"
	if meta.ForStmt.Post != nil {
		emitStmt(meta.ForStmt.Post)
	}
	printf("  jmp %s\n", labelCond)
	printf("  %s:\n", labelExit)
}

func emitRangeMap(meta *ir.MetaForContainer) {
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

	emitComment(2, "ForRangeStmt map Initialization\n")

	// _mp = EXPR
	emitAssignToVar(meta.ForRangeStmt.MapVar, meta.ForRangeStmt.X)

	//  if _mp == nil then exit
	emitVariable(meta.ForRangeStmt.MapVar) // value of _mp
	printf("  popq %%rax\n")
	printf("  cmpq $0, %%rax\n")
	printf("  je %s # exit if nil\n", labelExit)

	// item = mp.first
	emitVariableAddr(meta.ForRangeStmt.ItemVar)
	emitVariable(meta.ForRangeStmt.MapVar) // value of _mp
	emitLoadAndPush(types.Uintptr)         // value of _mp.first
	emitStore(types.Uintptr, true, false)  // assign

	// Condition
	// if item != nil; then
	//   execute body
	// else
	//   exit
	emitComment(2, "ForRangeStmt Condition\n")
	printf("  %s:\n", labelCond)

	emitVariable(meta.ForRangeStmt.ItemVar)
	printf("  popq %%rax\n")
	printf("  cmpq $0, %%rax\n")
	printf("  je %s # exit if nil\n", labelExit)

	emitComment(2, "assign key value to variables\n")

	// assign key
	keyMeta := meta.ForRangeStmt.Key
	if keyMeta != nil {
		if !sema.IsBlankIdentifierMeta(keyMeta) {
			emitAddr(keyMeta) // lhs
			// emit value of item.key
			//type item struct {
			//	next  *item
			//	key_dtype uintptr
			//  key_data uintptr <-- this
			//	value uintptr
			//}
			emitVariable(meta.ForRangeStmt.ItemVar)
			printf("  popq %%rax\n")            // &item{....}
			printf("  movq 16(%%rax), %%rcx\n") // item.key_data
			printf("  pushq %%rcx\n")
			emitLoadAndPush(sema.GetTypeOfExpr(keyMeta)) // load dynamic data
			emitStore(sema.GetTypeOfExpr(keyMeta), true, false)
		}
	}

	// assign value
	valueMeta := meta.ForRangeStmt.Value
	if valueMeta != nil {
		if !sema.IsBlankIdentifierMeta(valueMeta) {
			emitAddr(valueMeta) // lhs
			// emit value of item
			//type item struct {
			//	next  *item
			//	key_dtype uintptr
			//  key_data uintptr
			//	value uintptr  <-- this
			//}
			emitVariable(meta.ForRangeStmt.ItemVar)
			printf("  popq %%rax\n")            // &item{....}
			printf("  movq 24(%%rax), %%rcx\n") // item.key_data
			printf("  pushq %%rcx\n")
			emitLoadAndPush(sema.GetTypeOfExpr(valueMeta)) // load dynamic data
			emitStore(sema.GetTypeOfExpr(valueMeta), true, false)
		}
	}

	// Body
	emitComment(2, "ForRangeStmt Body\n")
	emitBlockStmt(meta.Body)

	// Post statement
	// item = item.next
	emitComment(2, "ForRangeStmt Post statement\n")
	printf("  %s:\n", labelPost)                // used for "continue"
	emitVariableAddr(meta.ForRangeStmt.ItemVar) // lhs
	emitVariable(meta.ForRangeStmt.ItemVar)     // item
	emitLoadAndPush(types.Uintptr)              // item.next
	emitStore(types.Uintptr, true, false)

	printf("  jmp %s\n", labelCond)

	printf("  %s:\n", labelExit)
}

func emitRangeStmt(meta *ir.MetaForContainer) {
	labelid++
	labelCond := fmt.Sprintf(".L.range.cond.%d", labelid)
	labelPost := fmt.Sprintf(".L.range.post.%d", labelid)
	labelExit := fmt.Sprintf(".L.range.exit.%d", labelid)

	meta.LabelPost = labelPost
	meta.LabelExit = labelExit
	// initialization: store len(rangeexpr)
	emitComment(2, "ForRangeStmt Initialization\n")
	emitComment(2, "  assign length to lenvar\n")
	// lenvar = len(s.X)
	emitVariableAddr(meta.ForRangeStmt.LenVar)
	emitLen(meta.ForRangeStmt.X)
	emitStore(types.Int, true, false)

	emitComment(2, "  assign 0 to indexvar\n")
	// indexvar = 0
	emitVariableAddr(meta.ForRangeStmt.Indexvar)
	emitZeroValue(types.Int)
	emitStore(types.Int, true, false)

	// init key variable with 0
	keyMeta := meta.ForRangeStmt.Key
	if keyMeta != nil {
		if !sema.IsBlankIdentifierMeta(keyMeta) {
			emitAddr(keyMeta) // lhs
			emitZeroValue(types.Int)
			emitStore(types.Int, true, false)
		}
	}

	// Condition
	// if (indexvar < lenvar) then
	//   execute body
	// else
	//   exit
	emitComment(2, "ForRangeStmt Condition\n")
	printf("  %s:\n", labelCond)

	emitVariableAddr(meta.ForRangeStmt.Indexvar)
	emitLoadAndPush(types.Int)
	emitVariableAddr(meta.ForRangeStmt.LenVar)
	emitLoadAndPush(types.Int)
	emitCompExpr("setl")
	emitPopBool(" indexvar < lenvar")
	printf("  cmpq $1, %%rax\n")
	printf("  jne %s # jmp if false\n", labelExit)

	emitComment(2, "assign list[indexvar] value variables\n")
	elemType := sema.GetTypeOfExpr(meta.ForRangeStmt.Value)
	emitAddr(meta.ForRangeStmt.Value) // lhs

	emitVariableAddr(meta.ForRangeStmt.Indexvar)
	emitLoadAndPush(types.Int) // index value
	emitListElementAddr(meta.ForRangeStmt.X, elemType)

	emitLoadAndPush(elemType)
	emitStore(elemType, true, false)

	// Body
	emitComment(2, "ForRangeStmt Body\n")
	emitBlockStmt(meta.Body)

	// Post statement: Increment indexvar and go next
	emitComment(2, "ForRangeStmt Post statement\n")
	printf("  %s:\n", labelPost)                 // used for "continue"
	emitVariableAddr(meta.ForRangeStmt.Indexvar) // lhs
	emitVariableAddr(meta.ForRangeStmt.Indexvar) // rhs
	emitLoadAndPush(types.Int)
	emitAddConst(1, "indexvar value ++")
	emitStore(types.Int, true, false)

	// incr key variable
	if keyMeta != nil {
		if !sema.IsBlankIdentifierMeta(keyMeta) {
			emitAddr(keyMeta)                            // lhs
			emitVariableAddr(meta.ForRangeStmt.Indexvar) // rhs
			emitLoadAndPush(types.Int)
			emitStore(types.Int, true, false)
		}
	}

	printf("  jmp %s\n", labelCond)

	printf("  %s:\n", labelExit)
}

func emitSwitchStmt(s *ir.MetaSwitchStmt) {
	labelid++
	labelEnd := fmt.Sprintf(".L.switch.%d.exit", labelid)
	if s.Init != nil {
		panic("TBI 2186")
	}
	if s.Tag == nil {
		panic("Omitted tag is not supported yet")
	}
	emitExpr(s.Tag)
	condType := sema.GetTypeOfExpr(s.Tag)
	cases := s.Cases
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
			assert(sema.GetSizeOfType(condType) <= 8 || sema.Kind(condType) == types.T_STRING, "should be one register size or string", __func__)
			switch sema.Kind(condType) {
			case types.T_STRING:
				ff := sema.LookupForeignFunc(sema.NewQI("runtime", "cmpstrings"))
				emitAllocReturnVarsAreaFF(ff)

				emitPushStackTop(condType, sema.SizeOfInt, "switch expr")
				emitExpr(m)

				emitCallFF(ff)
			case types.T_INTERFACE:
				ff := sema.LookupForeignFunc(sema.NewQI("runtime", "cmpinterface"))

				emitAllocReturnVarsAreaFF(ff)

				emitPushStackTop(condType, sema.SizeOfInt, "switch expr")
				emitExpr(m)

				emitCallFF(ff)
			case types.T_INT, types.T_UINT8, types.T_UINT16, types.T_UINTPTR, types.T_POINTER:
				emitPushStackTop(condType, 0, "switch expr")
				emitExpr(m)
				emitCompExpr("sete")
			default:
				unexpectedKind(sema.Kind(condType))
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
	for i, cc := range s.Cases {
		printf("  %s:\n", labels[i])
		for _, _s := range cc.Body {
			emitStmt(_s)
		}
		printf("  jmp %s\n", labelEnd)
	}
	printf("  %s:\n", labelEnd)
}
func emitTypeSwitchStmt(meta *ir.MetaTypeSwitchStmt) {
	labelid++
	labelEnd := fmt.Sprintf(".L.typeswitch.%d.exit", labelid)

	// subjectVariable = subject
	emitVariableAddr(meta.SubjectVariable)
	emitExpr(meta.Subject)
	emitStore(types.Eface, true, false)

	cases := meta.Cases
	var labels = make([]string, len(cases), len(cases))
	var defaultLabel string
	emitComment(2, "Start comparison with cases\n")
	for i, c := range meta.Cases {
		labelid++
		labelCase := ".L.case." + strconv.Itoa(labelid)
		labels[i] = labelCase
		if len(c.Types) == 0 {
			defaultLabel = labelCase
			continue
		}
		for _, t := range c.Types {
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
			sema.SetVariable(meta.AssignObj, c.Variable)
		}
		printf("  %s:\n", labels[i])

		var _isNil bool
		for _, typ := range c.Types {
			if typ == nil {
				_isNil = true
			}
		}

		if c.Variable != nil {
			// do assignment
			if _isNil {
				// @TODO: assign nil to the AssignObj of interface type
			} else {
				emitVariableAddr(c.Variable) // push lhs

				// push rhs
				emitVariableAddr(meta.SubjectVariable)
				emitLoadAndPush(types.Eface)
				printf("  popq %%rax # ifc.dtype\n")
				printf("  popq %%rcx # ifc.data\n")
				printf("  pushq %%rcx # ifc.data\n")
				emitLoadAndPush(c.Variable.Typ)

				// assign
				emitStore(c.Variable.Typ, true, false)
			}
		}

		for _, _s := range c.Body {
			emitStmt(_s)
		}
		printf("  jmp %s\n", labelEnd)
	}
	printf("  %s:\n", labelEnd)
}

func emitBranchStmt(meta *ir.MetaBranchStmt) {
	containerFor := meta.ContainerForStmt
	switch meta.ContinueOrBreak {
	case 1: // continue
		printf("  jmp %s # continue\n", containerFor.LabelPost)
	case 2: // break
		printf("  jmp %s # break\n", containerFor.LabelExit)
	default:
		throw(meta.ContinueOrBreak)
	}
}

func emitGoStmt(m *ir.MetaGoStmt) {
	emitCallMalloc(sema.SizeOfPtr) // area := new(func())
	emitExpr(m.Fun)
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

func emitStmt(meta ir.MetaStmt) {
	loc := getLoc(sema.Pos(meta))
	if loc == "" {
		printf("  # noloc %T\n", meta)
	} else {
		printf("  .loc %s # %T\n", loc, meta)
	}
	switch m := meta.(type) {
	case *ir.MetaBlockStmt:
		emitBlockStmt(m)
	case *ir.MetaExprStmt:
		emitExprStmt(m)
	case *ir.MetaVarDecl:
		emitDeclStmt(m)
	case *ir.MetaSingleAssign:
		emitSingleAssign(m.Lhs, m.Rhs)
	case *ir.MetaTupleAssign:
		if m.IsOK {
			emitOkAssignment(m)
		} else {
			emitFuncallAssignment(m)
		}
	case *ir.MetaReturnStmt:
		emitReturnStmt(m)
	case *ir.MetaIfStmt:
		emitIfStmt(m)
	case *ir.MetaForContainer:
		if m.ForRangeStmt != nil {
			if m.ForRangeStmt.IsMap {
				emitRangeMap(m)
			} else {
				emitRangeStmt(m)
			}
		} else {
			emitForStmt(m)
		}
	case *ir.MetaSwitchStmt:
		emitSwitchStmt(m)
	case *ir.MetaTypeSwitchStmt:
		emitTypeSwitchStmt(m)
	case *ir.MetaBranchStmt:
		emitBranchStmt(m)
	case *ir.MetaGoStmt:
		emitGoStmt(m)
	default:
		panic(fmt.Sprintf("unknown type:%T", meta))
	}
}

func emitRevertStackTop(t *types.Type) {
	printf("  addq $%d, %%rsp # revert stack top\n", sema.GetSizeOfType(t))
}

var labelid int

func emitFuncDecl(pkgName string, fnc *ir.Func) {
	printf("\n")
	//logf("[package %s][emitFuncDecl], fnc.name=\"%s\"\n", pkgName, fnc.Name)
	var symbol string
	if fnc.Method != nil {
		symbol = sema.GetMethodSymbol(fnc.Method)
		printf("# Method %s\n", symbol)
	} else {
		symbol = sema.GetPackageSymbol(pkgName, fnc.Name)
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

func emitGlobalVariable(pkg *ir.PkgContainer, vr *ir.PackageVals) {
	name := vr.Name.Name
	t := vr.Type
	typeKind := sema.Kind(vr.Type)
	val := vr.Val
	printf(".global %s.%s\n", pkg.Name, name)
	printf("%s.%s: # T %s\n", pkg.Name, name, string(typeKind))

	metaVal := vr.MetaVal
	_ = metaVal
	switch typeKind {
	case types.T_STRING:
		if metaVal == nil {
			// no value
			printf("  .quad 0\n")
			printf("  .quad 0\n")
		} else {
			lit, ok := metaVal.(*ir.MetaBasicLit)
			if !ok {
				panic("only BasicLit is supported")
			}
			sl := lit.StrVal
			printf("  .quad %s\n", sl.Label)
			printf("  .quad %d\n", sl.Strlen)
		}
	case types.T_BOOL:
		switch vl := val.(type) {
		case nil:
			printf("  .quad 0 # bool zero value\n")
		case *ast.Ident:
			switch vl.Obj {
			case universe.True:
				printf("  .quad 1 # bool true\n")
			case universe.False:
				printf("  .quad 0 # bool false\n")
			default:
				throw(val)
			}
		default:
			throw(val)
		}
	case types.T_INT, types.T_UINTPTR:
		switch vl := val.(type) {
		case nil:
			printf("  .quad 0\n")
		case *ast.BasicLit:
			printf("  .quad %s\n", vl.Value)
		default:
			throw(val)
		}
	case types.T_UINT8:
		switch vl := val.(type) {
		case nil:
			printf("  .byte 0\n")
		case *ast.BasicLit:
			printf("  .byte %s\n", vl.Value)
		default:
			throw(val)
		}
	case types.T_UINT16:
		switch vl := val.(type) {
		case nil:
			printf("  .word 0\n")
		case *ast.BasicLit:
			printf("  .word %s\n", vl.Value)
		default:
			throw(val)
		}
	case types.T_INT32:
		switch vl := val.(type) {
		case nil:
			printf("  .long 0\n")
		case *ast.BasicLit:
			printf("  .long %s\n", vl.Value)
		default:
			throw(val)
		}
	case types.T_SLICE:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		printf("  .quad 0 # ptr\n")
		printf("  .quad 0 # len\n")
		printf("  .quad 0 # cap\n")
	case types.T_STRUCT:
		if val != nil {
			panic("Unsupported global value")
		}
		for i := 0; i < sema.GetSizeOfType(t); i++ {
			printf("  .byte 0 # struct zero value\n")
		}
	case types.T_ARRAY:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		arrayType := t.E.(*ast.ArrayType)
		assert(arrayType.Len != nil, "slice type is not expected", __func__)
		length := sema.EvalInt(arrayType.Len)
		var zeroValue string
		knd := sema.Kind(sema.E2T(arrayType.Elt))
		switch knd {
		case types.T_INT:
			zeroValue = "  .quad 0 # int zero value\n"
		case types.T_UINT8:
			zeroValue = "  .byte 0 # uint8 zero value\n"
		case types.T_STRING:
			zeroValue = "  .quad 0 # string zero value (ptr)\n"
			zeroValue += "  .quad 0 # string zero value (len)\n"
		case types.T_INTERFACE:
			zeroValue = "  .quad 0 # eface zero value (dtype)\n"
			zeroValue += "  .quad 0 # eface zero value (data)\n"
		default:
			unexpectedKind(knd)
		}
		for i := 0; i < length; i++ {
			printf(zeroValue)
		}
	case types.T_POINTER, types.T_FUNC:
		// will be set in the initGlobal func
		printf("  .quad 0\n")
	case types.T_MAP:
		// will be set in the initGlobal func
		printf("  .quad 0\n")
	case types.T_INTERFACE:
		// will be set in the initGlobal func
		printf("  .quad 0\n")
		printf("  .quad 0\n")
	default:
		unexpectedKind(typeKind)
	}
}

func GenerateCode(pkg *ir.PkgContainer, fout *os.File) {
	Fout = fout
	TypesMap = make(map[string]*DtypeEntry)
	TypeId = 1

	printf("#--- string literals\n")
	printf(".data\n")
	for _, sl := range pkg.StringLiterals {
		//fmt.Fprintf(os.Stderr, "[emit string] %s index=%d\n", pkg.Name, idx)

		printf("%s:\n", sl.Label)
		printf("  .string %s\n", sl.Value)
	}

	printf("#--- global vars (static values)\n")
	for _, vr := range pkg.Vals {
		if vr.Type == nil {
			panic("type cannot be nil for global variable: " + vr.Name.Name)
		}
		emitGlobalVariable(pkg, vr)
	}

	printf("\n")
	printf("#--- global vars (dynamic value setting)\n")
	printf(".text\n")
	printf(".global %s.__initVars\n", pkg.Name)
	printf("%s.__initVars:\n", pkg.Name)
	for _, vr := range pkg.Vals {
		if vr.MetaVal == nil {
			continue
		}
		printf("  \n")
		typeKind := sema.Kind(vr.Type)
		switch typeKind {
		case types.T_POINTER, types.T_MAP, types.T_INTERFACE:
			printf("  # init global %s:\n", vr.Name.Name)
			emitSingleAssign(vr.MetaVar, vr.MetaVal)
		}
	}
	printf("  ret\n")

	for _, fnc := range pkg.Funcs {
		emitFuncDecl(pkg.Name, fnc)
	}

	emitDynamicTypes(TypesMap)
	printf("\n")

	Fout = nil
	TypesMap = nil
	TypeId = 0
}

func emitDynamicTypes(mapDtypes map[string]*DtypeEntry) {
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

func serializeType(t *types.Type) string {
	if t == nil {
		panic("nil type is not expected")
	}
	if t.E == sema.GeneralSlice {
		panic("TBD: GeneralSlice")
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
			case universe.Uintptr:
				return "uintptr"
			case universe.Int:
				return "int"
			case universe.String:
				return "string"
			case universe.Uint8:
				return "uint8"
			case universe.Uint16:
				return "uint16"
			case universe.Bool:
				return "bool"
			case universe.Error:
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
				typ := sema.E2T(field.Type)
				r += fmt.Sprintf("%s %s;", name, serializeType(typ))
			}
		}
		return r + "}"
	case *ast.ArrayType:
		if e.Len == nil {
			if e.Elt == nil {
				panic(e)
			}
			return "[]" + serializeType(sema.E2T(e.Elt))
		} else {
			return "[" + strconv.Itoa(sema.EvalInt(e.Len)) + "]" + serializeType(sema.E2T(e.Elt))
		}
	case *ast.StarExpr:
		return "*" + serializeType(sema.E2T(e.X))
	case *ast.Ellipsis: // x ...T
		panic("TBD: Ellipsis")
	case *ast.InterfaceType:
		return "interface{}" // @TODO list methods
	case *ast.MapType:
		return "map[" + serializeType(sema.E2T(e.Key)) + "]" + serializeType(sema.E2T(e.Value))
	case *ast.SelectorExpr:
		qi := sema.Selector2QI(e)
		return string(qi)
	case *ast.FuncType:
		return "func"
	default:
		throw(t)
	}
	return ""
}
