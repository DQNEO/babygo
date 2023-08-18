package codegen

import (
	"os"

	"github.com/DQNEO/babygo/internal/ir"
	"github.com/DQNEO/babygo/internal/sema"
	"github.com/DQNEO/babygo/internal/types"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/strconv"
	"github.com/DQNEO/babygo/lib/token"
)

var Fout *os.File
var labelid int
var DebugCodeGen bool = true

var __func__ = "__func__"

func assert(bol bool, msg string, caller string) {
	if !bol {
		panic(sema.CurrentPkg.Name + ":" + caller + ": " + msg)
	}
}

func panicPos(s string, pos token.Pos) {
	position := sema.Fset.Position(pos)
	panic(fmt.Sprintf("%s\n\t%s", s, position.String()))
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

func emitPushStackTop(t types.GoType, offset int, comment string) {
	knd := sema.Kind2(t)
	switch knd {
	case types.T_STRING:
		printf("  movq %d+8(%%rsp), %%rcx # copy str.len from stack top (%s)\n", offset, comment)
		printf("  movq %d+0(%%rsp), %%rax # copy str.ptr from stack top (%s)\n", offset, comment)
		printf("  pushq %%rcx # str.len\n")
		printf("  pushq %%rax # str.ptr\n")
	case types.T_POINTER, types.T_UINTPTR, types.T_BOOL, types.T_INT, types.T_UINT8, types.T_UINT16:
		printf("  movq %d(%%rsp), %%rax # copy stack top value (%s) \n", offset, comment)
		printf("  pushq %%rax\n")
	default:
		unexpectedKind(knd)
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
	emitLoadAndPush(variable.Type)
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
	case *ir.Variable:
		emitVariableAddr(m)
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
		if m.IsQI { // pkg.Var|pkg.Con|pkg.Fun
			emitAddr(m.ForeignValue)
		} else { // (e).field
			if m.NeedDeref {
				emitExpr(m.X)
			} else {
				emitAddr(m.X)
			}
			emitAddConst(m.Offset, "struct head address + struct.field offset")
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
	fromType := sema.GetTypeOfExpr(arg0)
	fromKind := sema.Kind(fromType)
	toKind := sema.Kind(toType)
	switch toKind {
	case types.T_STRING:
		if fromKind == types.T_SLICE {
			emitExpr(arg0) // slice
			emitPopSlice()
			printf("  pushq %%rcx # str len\n")
			printf("  pushq %%rax # str ptr\n")
			return
		} else {
			emitExpr(arg0)
			return
		}
	case types.T_SLICE:
		assert(fromKind == types.T_STRING, "source type should be slice", __func__)
		emitComment(2, "Conversion of string => slice \n")
		// @FIXME: As string should be immutable, we must copy data.
		emitExpr(arg0)
		emitPopString()
		printf("  pushq %%rcx # cap\n")
		printf("  pushq %%rcx # len\n")
		printf("  pushq %%rax # ptr\n")
		return
	case types.T_INTERFACE:
		emitExpr(arg0)
		if sema.IsInterface(fromType) {
			return // do nothing
		} else {
			// Convert dynamic value to interface
			emitConvertToInterface(fromType, toType)
			return
		}
	default:
		emitExpr(arg0)
		return
	}
}

func emitIfcConversion(ic *ir.IfcConversion) {
	emitExpr(ic.Value)
	emitComment(2, "emitIfcConversion\n")
	emitConvertToInterface(sema.GetTypeOfExpr(ic.Value), ic.Type)
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
	t := sema.GetTypeOfExpr(arg)
	switch sema.Kind(t) {
	case types.T_ARRAY:
		arrayLen := sema.GetArrayLen(t)
		printf("  pushq $%d # array len\n", arrayLen)
	case types.T_SLICE:
		emitExpr(arg)
		emitPopSlice()
		printf("  pushq %%rcx # len\n")
	case types.T_STRING:
		emitExpr(arg)
		emitPopString()
		printf("  pushq %%rcx # len\n")
	case types.T_MAP:
		sig := sema.NewLenMapSignature(arg)
		emitCallDirect("runtime.lenMap", []ir.MetaExpr{arg}, sig)

	default:
		unexpectedKind(sema.Kind(sema.GetTypeOfExpr(arg)))
	}
}

func emitCap(arg ir.MetaExpr) {
	t := sema.GetTypeOfExpr(arg)
	switch sema.Kind(t) {
	case types.T_ARRAY:
		arrayLen := sema.GetArrayLen(t)
		printf("  pushq $%d # array len\n", arrayLen)
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
		emitPushStackTop(types.Uintptr.GoType, 0, "address of struct heaad")

		fieldOffset := sema.GetStructFieldOffset(metaElm.Field)
		emitAddConst(fieldOffset, "address of struct field")

		// push rhs value
		emitExpr(metaElm.Value)

		// assign
		emitStore(metaElm.FieldType, true, false)
	}
}

func emitArrayLiteral(meta *ir.MetaCompositLit) {
	elmType := meta.ElmType
	elmSize := sema.GetSizeOfType(elmType)
	memSize := elmSize * meta.Len

	emitCallMalloc(memSize) // push
	for i, elm := range meta.Elms {
		// push lhs address
		emitPushStackTop(types.Uintptr.GoType, 0, "malloced address")
		emitAddConst(elmSize*i, "malloced address + elmSize * index")
		// push rhs value
		emitExpr(elm)

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

func emitCallDirect(symbol string, args []ir.MetaExpr, sig *ir.Signature) {
	emitCall(sema.NewFuncValueFromSymbol(symbol), args, sig.ParamTypes, sig.ReturnTypes)
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
func emitCall(fv *ir.FuncValue, args []ir.MetaExpr, paramTypes []*types.Type, returnTypes []*types.Type) {
	emitComment(2, "emitCall len(args)=%d\n", len(args))
	var totalParamSize int
	var offsets []int
	for i, paramType := range paramTypes {
		offsets = append(offsets, totalParamSize)
		var size int
		if fv.IfcMethodCal && i == 0 {
			size = sema.GetSizeOfType(types.Uintptr) // @FIXME
		} else {
			size = sema.GetSizeOfType(paramType)
		}
		totalParamSize += size
	}

	emitAllocReturnVarsArea(getTotalSizeOfType(returnTypes))
	printf("  subq $%d, %%rsp # alloc parameters area\n", totalParamSize)
	for i, arg := range args {
		if i == 0 && fv.IfcMethodCal {
			// tweak recevier
			paramType := paramTypes[i]
			emitExpr(arg)
			emitPop(sema.Kind(paramType))
			printf("  leaq %d(%%rsp), %%rsi # place to save\n", offsets[i])
			printf("  movq 0(%%rcx), %%rcx # load eface.data\n", 0)
			printf("  movq %%rcx, %d(%%rsi) # store eface.data\n", 0)
			printf("  movq %%rax, %%r12 # copy eface.dtype\n", 0) //@TODO %r12 can be overwritten by another expr
		} else {
			paramType := paramTypes[i]
			emitExpr(arg)
			emitPop(sema.Kind(paramType))
			printf("  leaq %d(%%rsp), %%rsi # place to save\n", offsets[i])
			printf("  pushq %%rsi # place to save\n")
			emitRegiToMem(paramType)
		}
	}

	emitCallQ(fv, totalParamSize, returnTypes)
}

func emitAllocReturnVarsAreaFF(ff *ir.Func) {
	emitAllocReturnVarsArea(getTotalSizeOfType(ff.Signature.ReturnTypes))
}

func getTotalSizeOfType(types []*types.Type) int {
	var r int
	for _, t := range types {
		r += sema.GetSizeOfType(t)
	}
	return r
}

func emitCallFF(ff *ir.Func) {
	totalParamSize := getTotalSizeOfType(ff.Signature.ParamTypes)
	symbol := ff.PkgName + "." + ff.Name
	emitCallQ(sema.NewFuncValueFromSymbol(symbol), totalParamSize, ff.Signature.ReturnTypes)
}

func emitCallQ(fv *ir.FuncValue, totalParamSize int, returnTypes []*types.Type) {
	if fv.IsDirect {
		if fv.Symbol == "" {
			panic("callq target must not be empty")
		}
		printf("  callq %s\n", fv.Symbol)
	} else {
		if fv.IfcMethodCal {
			emitComment(2, "fv.IfcMethodCal\n")
			printf("  movq %%r12, %%rax # eface.dtype\n")
			printf("  addq $24, %%rax # addr(eface.dtype)  8*3 \n")

			var methodId int
			methods := sema.GetInterfaceMethods(fv.IfcType)
			for i, m := range methods {
				if m.Names[0].Name == fv.MethodName {
					methodId = i
				}
			}
			methodOffset := methodId*8*3 + 16
			printf("  addq $%d, %%rax # addr(eface.dtype)\n", methodOffset)
			printf("  movq (%%rax), %%rax # load method addr\n")
			printf("  callq *%%rax\n")

		} else {

			emitExpr(fv.Expr)
			printf("  popq %%rax\n")
			printf("  callq *%%rax\n")

		}
	}

	emitFreeParametersArea(totalParamSize)
	printf("  #  totalReturnSize=%d\n", getTotalSizeOfType(returnTypes))
	emitFreeAndPushReturnedValue(returnTypes)
}

// callee
func emitReturnStmt(meta *ir.MetaReturnStmt) {
	if meta.IsTuple {
		emitFuncallAssignment(meta.TupleAssign)
	} else {
		for _, sa := range meta.SingleAssignments {
			emitSingleAssign(sa)
		}
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

func emitMetaCallNew(m *ir.MetaCallNew) {
	// size to malloc
	size := sema.GetSizeOfType(m.TypeArg0)
	emitCallMalloc(size)
}

func emitMetaCallMake(m *ir.MetaCallMake) {
	typeArg := m.TypeArg0
	switch sema.Kind(typeArg) {
	case types.T_MAP:
		mapValueType := sema.GetElementTypeOfCollectionType(typeArg)
		valueSize := sema.NewNumberLiteral(sema.GetSizeOfType(mapValueType), m.Pos())
		// A new, empty map value is made using the built-in function make,
		// which takes the map type and an optional capacity hint as arguments:
		length := sema.NewNumberLiteral(0, m.Pos())
		emitCallDirect("runtime.makeMap", []ir.MetaExpr{length, valueSize}, ir.RuntimeMakeMapSignature)
		return
	case types.T_SLICE:
		// make([]T, ...)
		elmType := sema.GetElementTypeOfCollectionType(typeArg)
		elmSize := sema.GetSizeOfType(elmType)
		numlit := sema.NewNumberLiteral(elmSize, m.Pos())
		args := []ir.MetaExpr{numlit, m.Arg1, m.Arg2}
		emitCallDirect("runtime.makeSlice", args, ir.RuntimeMakeSliceSignature)
		return
	default:
		throw(typeArg)
	}

}

func emitMetaCallAppend(m *ir.MetaCallAppend) {
	sliceArg := m.Arg0
	elemArg := m.Arg1
	elmType := sema.GetElementTypeOfCollectionType(sema.GetTypeOfExpr(sliceArg))
	elmSize := sema.GetSizeOfType(elmType)

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
	arg1 := sema.CheckIfcConversion(m.Pos(), elemArg, elmType)
	args := []ir.MetaExpr{sliceArg, arg1}
	sig := sema.NewAppendSignature(elmType)
	emitCallDirect(symbol, args, sig)
	return

}

func emitMetaCallPanic(m *ir.MetaCallPanic) {
	funcVal := "runtime.panic"
	arg0 := sema.CheckIfcConversion(m.Pos(), m.Arg0, types.Eface)
	args := []ir.MetaExpr{arg0}
	emitCallDirect(funcVal, args, ir.BuiltinPanicSignature)
	return
}

func emitMetaCallDelete(m *ir.MetaCallDelete) {
	funcVal := "runtime.deleteMap"
	sig := sema.NewDeleteSignature(m.Arg0)
	mc := sema.CheckIfcConversion(m.Pos(), m.Arg1, sig.ParamTypes[1])
	args := []ir.MetaExpr{m.Arg0, mc}
	emitCallDirect(funcVal, args, sig)
	return

}

// 1 value
func emitIdent(meta *ir.MetaIdent) {
	switch meta.Kind {
	case "true": // true constant
		emitTrue()
	case "false": // false constant
		emitFalse()
	case "nil": // zero value
		metaType := meta.Type
		if metaType == nil {
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
		emitExpr(meta.ForeignValue) // var,con,fun
	} else {
		// strct.field
		emitAddr(meta)
		emitLoadAndPush(sema.GetTypeOfExpr(meta))
	}
}

func emitForeignFuncAddr(meta *ir.MetaForeignFuncWrapper) {
	emitFuncAddr(meta.QI)
}

// 1 value
func emitBasicLit(mt *ir.MetaBasicLit) {
	switch mt.Kind {
	case "CHAR":
		printf("  pushq $%d # convert char literal to int\n", mt.CharVal)
	case "INT":
		printf("  pushq $%d # number literal %s\n", mt.IntVal, mt.RawValue)
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
	mc := sema.CheckIfcConversion(m.Pos(), key, ir.RuntimeGetAddrForMapGetSignature.ParamTypes[1])
	args := []ir.MetaExpr{mp, mc}
	emitCallDirect("runtime.getAddrForMapGet", args, ir.RuntimeGetAddrForMapGetSignature)
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
	emitDtypeLabelAddr(meta.Type, sema.GetTypeOfExpr(meta.X))
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
	loc := getLoc(meta.Pos())
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
	case *ir.MetaForeignFuncWrapper:
		emitForeignFuncAddr(m)
	case *ir.MetaCallLen:
		emitLen(m.Arg0)
	case *ir.MetaCallCap:
		emitCap(m.Arg0)
	case *ir.MetaCallNew:
		emitMetaCallNew(m)
	case *ir.MetaCallMake:
		emitMetaCallMake(m)
	case *ir.MetaCallAppend:
		emitMetaCallAppend(m)
	case *ir.MetaCallPanic:
		emitMetaCallPanic(m)
	case *ir.MetaCallDelete:
		emitMetaCallDelete(m)
	case *ir.MetaConversionExpr:
		emitConversion(m.Type, m.Arg0)
	case *ir.MetaCallExpr:
		emitCall(m.FuncVal, m.Args, m.ParamTypes, m.Types) // can be Tuple
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
	case *ir.IfcConversion:
		emitIfcConversion(m)
	default:
		panic(fmt.Sprintf("meta type:%T", meta))
	}
}

// convert stack top value to interface
func emitConvertToInterface(fromType *types.Type, toType *types.Type) {
	emitComment(2, "ConversionToInterface\n")
	if sema.HasIfcMethod(toType) {
		emitComment(2, "@@@ toType has methods\n")
	}
	memSize := sema.GetSizeOfType(fromType)
	// copy data to heap
	emitCallMalloc(memSize)
	emitStore(fromType, false, true) // heap addr pushed
	// push dtype label's address
	emitDtypeLabelAddr(fromType, toType)
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

func emitDtypeLabelAddr(t *types.Type, it *types.Type) {
	de := sema.GetITabEntry(t, it)
	dtypeLabel := de.Label
	sr := de.DSerialized
	printf("  leaq %s(%%rip), %%rax # dtype label address \"%s\"\n", dtypeLabel, sr)
	printf("  pushq %%rax           # dtype label address\n")
}

func emitAddrForMapSet(indexExpr *ir.MetaIndexExpr) {
	// alloc heap for map value
	//size := GetSizeOfType(elmType)
	emitComment(2, "[emitAddrForMapSet]\n")
	keyExpr := sema.CheckIfcConversion(indexExpr.Pos(), indexExpr.Index, ir.RuntimeGetAddrForMapSetSignature.ParamTypes[1])
	args := []ir.MetaExpr{indexExpr.X, keyExpr}
	emitCallDirect("runtime.getAddrForMapSet", args, ir.RuntimeGetAddrForMapSetSignature)
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
	args := []ir.MetaExpr{left, right}
	emitCallDirect("runtime.catstrings", args, ir.RuntimeCatStringsSignature)
}

func emitCompStrings(left ir.MetaExpr, right ir.MetaExpr) {
	args := []ir.MetaExpr{left, right}
	emitCallDirect("runtime.cmpstrings", args, ir.RuntimeCmpStringFuncSignature)
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

func emitAssignZeroValue(lhs ir.MetaExpr, lhsType *types.Type) {
	emitComment(2, "emitAssignZeroValue\n")
	emitComment(2, "lhs addresss\n")
	emitAddr(lhs)
	emitComment(2, "emitZeroValue\n")
	emitZeroValue(lhsType)
	emitStore(lhsType, true, false)

}

func emitSingleAssign(a *ir.MetaSingleAssign) {
	_emitSingleAssign(a.Lhs, a.Rhs)
}

func _emitSingleAssign(lhs ir.MetaExpr, rhs ir.MetaExpr) {
	//	lhs := metaSingle.lhs
	//	rhs := metaSingle.rhs
	if sema.IsBlankIdentifierMeta(lhs) {
		emitExpr(rhs)
		rhsType := sema.GetTypeOfExpr(rhs)
		if rhsType == nil {
			panicPos("rhs type should not be nil", rhs.Pos())
		}
		emitPop(sema.Kind(rhsType))
		return
	}
	emitComment(2, "Assignment: emitAddr(lhs)\n")
	emitAddr(lhs)
	emitComment(2, "Assignment: emitExpr(rhs)\n")
	emitExpr(rhs)
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
		emitSingleAssign(meta.Single)
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

	// _mp = EXPRs
	emitSingleAssign(meta.ForRangeStmt.MapVarAssign)

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

				emitPushStackTop(condType.GoType, sema.SizeOfInt, "switch expr")
				emitExpr(m)

				emitCallFF(ff)
			case types.T_INTERFACE:
				ff := sema.LookupForeignFunc(sema.NewQI("runtime", "cmpinterface"))

				emitAllocReturnVarsAreaFF(ff)

				emitPushStackTop(condType.GoType, sema.SizeOfInt, "switch expr")
				emitExpr(m)

				emitCallFF(ff)
			case types.T_INT, types.T_UINT8, types.T_UINT16, types.T_UINTPTR, types.T_POINTER:
				emitPushStackTop(condType.GoType, 0, "switch expr")
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
				emitDtypeLabelAddr(t, sema.GetTypeOfExpr(meta.Subject))
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
				emitLoadAndPush(c.Variable.Type)

				// assign
				emitStore(c.Variable.Type, true, false)
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

func emitDeferStmt(m *ir.MetaDeferStmt) {
	emitSingleAssign(m.FuncAssign)
}

func emitStmt(meta ir.MetaStmt) {
	loc := getLoc(meta.Pos())
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
		emitSingleAssign(m)
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
	case *ir.MetaDeferStmt:
		emitDeferStmt(m)
	default:
		panic(fmt.Sprintf("unknown type:%T", meta))
	}
}

func emitRevertStackTop(t *types.Type) {
	printf("  addq $%d, %%rsp # revert stack top\n", sema.GetSizeOfType(t))
}

var deferLabelId int = 1

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

	if fnc.HasDefer {
		printf(".L.defer.%d:\n", deferLabelId)
		deferLabelId++
		emitVariable(fnc.DeferVar) // defer func addr
		printf("  popq %%rax # defer func addr\n")
		printf("  callq *%%rax # defer func addr\n")
	}

	printf("  leave\n")
	printf("  ret\n")
}

func emitZeroData(t *types.Type) {
	for i := 0; i < sema.GetSizeOfType(t); i++ {
		printf("  .byte 0 # zero value\n")
	}
}

func emitData(dotSize string, val ir.MetaExpr) {
	if val == nil {
		printf("  %s 0\n", dotSize)
	} else {
		var lit *ir.MetaBasicLit
		switch m := val.(type) {
		case *ir.MetaBasicLit:
			lit = m
		}

		printf("  %s %s\n", dotSize, lit.RawValue)
	}
}

func emitGlobalVarConst(pkgName string, vr *ir.PackageVarConst) {
	name := vr.Name.Name
	t := vr.Type
	typeKind := sema.Kind(vr.Type)
	val := vr.Val
	printf(".global %s.%s\n", pkgName, name)
	printf("%s.%s: # T %s\n", pkgName, name, string(typeKind))

	metaVal := vr.MetaVal
	switch typeKind {
	case types.T_STRING:
		if metaVal == nil {
			// no value
			emitZeroData(t)
		} else {
			var lit *ir.MetaBasicLit
			switch m := metaVal.(type) {
			case *ir.MetaBasicLit:
				lit = m
			}
			sl := lit.StrVal
			printf("  .quad %s\n", sl.Label)
			printf("  .quad %d\n", sl.Strlen)
		}
	case types.T_BOOL:
		if metaVal == nil {
			printf("  .quad 0 # bool zero value\n")
		} else {
			var i *ir.MetaIdent
			switch m := metaVal.(type) {
			case *ir.MetaIdent:
				i = m
			}

			switch i.Kind {
			case "true":
				printf("  .quad 1 # bool true\n")
			case "false":
				printf("  .quad 0 # bool false\n")
			default:
				panic("Unexpected bool value")
			}
		}
	case types.T_UINT8:
		emitData(".byte", metaVal)
	case types.T_UINT16:
		emitData(".word", metaVal)
	case types.T_INT32:
		emitData(".long", metaVal)
	case types.T_INT, types.T_UINTPTR:
		emitData(".quad", metaVal)
	case types.T_STRUCT, types.T_ARRAY, types.T_SLICE:
		// only zero value
		if val != nil {
			panic("Unsupported global value:" + typeKind)
		}
		emitZeroData(t)
	case types.T_POINTER, types.T_FUNC, types.T_MAP, types.T_INTERFACE:
		// will be set in the initGlobal func
		emitZeroData(t)
	default:
		unexpectedKind(typeKind)
	}
}

func GenerateDecls(pkg *ir.AnalyzedPackage, declFilePath string) {
	fout, err := os.Create(declFilePath)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(fout, "package %s\n", pkg.Name)

	// list import
	for _, im := range pkg.Imports {
		fmt.Fprintf(fout, "import \"%s\"\n", im)
	}
	// Type, Con, Var, Func
	for _, typ := range pkg.Types {
		ut := sema.GetUnderlyingType(typ)
		fmt.Fprintf(fout, "type %s %s\n", typ.Name, sema.SerializeType(ut, false))
	}
	for _, vr := range pkg.Vars {
		fmt.Fprintf(fout, "var %s %s\n", vr.Name.Name, sema.SerializeType(vr.Type, false))
	}
	for _, cnst := range pkg.Consts {
		fmt.Fprintf(fout, "const %s %s = %s\n", cnst.Name.Name, sema.SerializeType(cnst.Type, false), sema.GetConstRawValue(cnst.MetaVal))
	}
	for _, fnc := range pkg.Funcs {
		fmt.Fprintf(fout, "%s\n", sema.RestoreFuncDecl(fnc))
	}

	fout.Close()
}

func GenerateCode(pkg *ir.AnalyzedPackage, fout *os.File) {
	Fout = fout
	labelid = 1

	printf("#--- string literals\n")
	printf(".data\n")
	for _, sl := range pkg.StringLiterals {
		printf("%s:\n", sl.Label)
		printf("  .string %s\n", sl.Value)
	}

	printf("#--- global vars (static values)\n")
	for _, vr := range pkg.Vars {
		if vr.Type == nil {
			panic("type cannot be nil for global variable: " + vr.Name.Name)
		}
		emitGlobalVarConst(pkg.Name, vr)
	}

	printf("#--- global consts\n")
	for _, cnst := range pkg.Consts {
		if cnst.Type == nil {
			panic("type cannot be nil for global const: " + cnst.Name.Name)
		}
		emitGlobalVarConst(pkg.Name, cnst)
	}

	printf("\n")
	printf("#--- global vars (dynamic value setting)\n")
	printf(".text\n")
	printf(".global %s.__initVars\n", pkg.Name)
	printf("%s.__initVars:\n", pkg.Name)
	for _, vr := range pkg.Vars {
		if vr.MetaVal == nil {
			continue
		}
		printf("  \n")
		typeKind := sema.Kind(vr.Type)
		switch typeKind {
		case types.T_POINTER, types.T_MAP, types.T_INTERFACE:
			printf("  # init global %s:\n", vr.Name.Name)
			_emitSingleAssign(vr.MetaVar, vr.MetaVal)
		}
	}
	printf("  ret\n")

	for _, fnc := range pkg.Funcs {
		if fnc.HasBody { // if not hasBody, func decl should be in asm code.
			emitFuncDecl(pkg.Name, fnc)
		}
	}

	emitInterfaceTables(sema.ITab)
	printf("\n")

	sema.ITab = nil
	sema.ITabID = 0
}

func emitInterfaceTables(itab map[string]*sema.ITabEntry) {
	printf("# ------- Dynamic Types (len = %d)------\n", len(itab))
	printf(".data\n")

	entries := make([]string, len(itab)+1, len(itab)+1)

	// sort map in order to assure the deterministic results
	for key, ent := range itab {
		entries[ent.Id] = key
	}

	// skip id=0
	for id := 1; id < len(entries); id++ {
		key := entries[id]
		ent := itab[key]

		printf(".string_dtype_%d:\n", id)
		printf("  .string \"%s\"\n", ent.DSerialized)
		printf("%s: # %s\n", ent.Label, key)
		printf("  .quad %d\n", id)
		printf("  .quad .string_dtype_%d\n", id)
		printf("  .quad %d\n", len(ent.DSerialized))

		methods := sema.GetInterfaceMethods(ent.Itype)
		if len(methods) == 0 {
			printf("  # no methods for %s\n", sema.SerializeType(ent.Itype, true))
		}
		for mi, m := range methods {
			dmethod := sema.LookupMethod(ent.Dtype, m.Names[0])
			sym := sema.GetMethodSymbol(dmethod)
			printf("  .quad .method_name_%d_%d # %s \n", id, mi, m.Names[0].Name)
			printf("  .quad %d # method name len\n", len(m.Names[0].Name))
			printf("  .quad %s # method ref %s\n", sym, m.Names[0].Name)
		}
		printf("  .quad 0 # End of methods\n")

		for mi, m := range methods {
			printf(".method_name_%d_%d:\n", id, mi)
			printf("  .string \"%s\"\n", m.Names[0].Name)
		}
	}
	printf("\n")
}
