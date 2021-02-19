package main

import (
	"fmt"
	"os"
	"syscall"

	"go/ast"
	"go/parser"
	"go/token"

	//"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/path"
	"github.com/DQNEO/babygo/lib/strconv"
	"github.com/DQNEO/babygo/lib/strings"
)

func assert(bol bool, msg string) {
	if !bol {
		panic(msg)
	}
}

func unexpectedKind(knd TypeKind) {
	panic("Unexpected Kind: " + string(knd))
}

func throw(x interface{}) {
	panic(fmt.Sprintf("%#v", x))
}

func parseImports(fset *token.FileSet, filename string) *ast.File {
	f, err := parser.ParseFile(fset, filename, nil, parser.ImportsOnly)
	if err != nil {
		panic(filename + ":" + err.Error())
	}
	return f
}

func parseFile(fset *token.FileSet, filename string) *ast.File {
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		panic(err)
	}
	return f
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
	assert(t != nil, "type should not be nil")
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
		fmt.Printf("  leaq %d(%%rbp), %%rax # local variable \"%s\"\n", int(variable.LocalOffset), variable.Name)
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
			panic("Unexpected ident kind")
		}
	case *ast.IndexExpr:
		emitExpr(e.Index, nil) // index number
		list := e.X
		elmType := getTypeOfExpr(e)
		emitListElementAddr(list, elmType)
	case *ast.StarExpr:
		emitExpr(e.X, nil)
	case *ast.SelectorExpr: // (X).Sel
		typeOfX := getTypeOfExpr(e.X)
		var structType *Type
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
		field := lookupStructField(getStructTypeSpec(structType), e.Sel.Name)
		offset := getStructFieldOffset(field)
		emitAddConst(offset, "struct head address + struct.field offset")
	case *ast.CompositeLit:
		knd := kind(getTypeOfExpr(e))
		switch knd {
		case T_STRUCT:
			// result of evaluation of a struct literal is its address
			emitExpr(expr, nil)
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
		if e.Obj == nil {
			panic("unresolved ident: " + e.String())
		}
		return e.Obj.Kind == ast.Typ
	case *ast.SelectorExpr:
		if isQI(e) {
			qi := selector2QI(e)
			ident := lookupForeignIdent(qi)
			if ident.Obj.Kind == ast.Typ {
				return true
			}
		}
	case *ast.ParenExpr: // We assume (T)(e) is conversion
		return isType(e.X)
	case *ast.StarExpr:
		return isType(e.X)
	case *ast.InterfaceType:
		return true
	}
	emitComment(0, "[isType][%T] is not considered a type\n", expr)
	return false
}

// explicit conversion T(e)
func emitConversion(toType *Type, arg0 ast.Expr) {
	emitComment(2, "emitConversion\n")
	switch to := toType.E.(type) {
	case *ast.Ident:
		switch to.Obj {
		case gString: // string(e)
			switch kind(getTypeOfExpr(arg0)) {
			case T_SLICE: // string(slice)
				emitExpr(arg0, nil)
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
			panic("TBI")
		}
	case *ast.ArrayType: // Conversion to slice
		arrayType := to
		if arrayType.Len != nil {
			throw(to)
		}
		assert(kind(getTypeOfExpr(arg0)) == T_STRING, "source type should be slice")
		emitComment(2, "Conversion to slice\n")
		emitExpr(arg0, nil)
		emitPopString()
		fmt.Printf("  pushq %%rcx # cap\n")
		fmt.Printf("  pushq %%rcx # len\n")
		fmt.Printf("  pushq %%rax # ptr\n")
	case *ast.ParenExpr: // (T)(arg0)
		emitConversion(e2t(to.X), arg0)
	case *ast.StarExpr: // (*T)(arg0)
		// go through
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
	return
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
		arrayType, ok := getTypeOfExpr(arg).E.(*ast.ArrayType)
		assert(ok, "should be *ast.ArrayType")
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
		arrayType, ok := getTypeOfExpr(arg).E.(*ast.ArrayType)
		assert(ok, "should be *ast.ArrayType")
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
	for i, elm := range e.Elts {
		kvExpr, ok := elm.(*ast.KeyValueExpr)
		assert(ok, "expect *ast.KeyValueExpr")
		fieldName, ok := kvExpr.Key.(*ast.Ident)
		assert(ok, "expect *ast.Ident")
		emitComment(2, "- [%d] : key=%s, value=%T\n", i, fieldName.Name, kvExpr.Value)
		field := lookupStructField(getStructTypeSpec(structType), fieldName.Name)
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
	var argIndex int
	var eArg ast.Expr
	var param *ast.Field
	lenParams := len(params)
	for argIndex, eArg = range eArgs {
		if argIndex < lenParams {
			param = params[argIndex]
			elp, ok := param.Type.(*ast.Ellipsis)
			if ok {
				variadicElmType = elp.Elt
				variadicArgs = make([]ast.Expr, 0)
			}
		}
		if variadicArgs != nil && !expandElipsis {
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

	if variadicArgs != nil && !expandElipsis {
		// collect args as a slice
		sliceType := &ast.ArrayType{Elt: variadicElmType}
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
		param := params[argIndex+1]
		elp, ok := param.Type.(*ast.Ellipsis)
		assert(ok, "compile error")
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
	node := mapReturnStmt[s]
	fnc := node.fnc
	if len(fnc.retvars) != len(s.Results) {
		panic("length of return and func type do not match")
	}

	var i int
	_len := len(s.Results)
	for i = 0; i < _len; i++ {
		emitAssignToVar(fnc.retvars[i], s.Results[i])
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
		emitComment(2, "emit return value\n")
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
			assert(len(eArgs) == 1, "builtin len should take only 1 args")
			var arg ast.Expr = eArgs[0]
			emitLen(arg)
			return
		case gCap:
			assert(len(eArgs) == 1, "builtin len should take only 1 args")
			var arg ast.Expr = eArgs[0]
			emitCap(arg)
			return
		case gNew:
			typeArg := e2t(eArgs[0])
			// size to malloc
			size := getSizeOfType(typeArg)
			emitCallMalloc(size)
			return
		case gMake:
			var typeArg = e2t(eArgs[0])
			switch kind(typeArg) {
			case T_SLICE:
				// make([]T, ...)
				arrayType, ok := typeArg.E.(*ast.ArrayType)
				assert(ok, "should be *ast.ArrayType")
				var elmSize = getSizeOfType(e2t(arrayType.Elt))
				var numlit = newNumberLiteral(elmSize)

				var args []*Arg = []*Arg{
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

				var resultList = &ast.FieldList{List: []*ast.Field{
					&ast.Field{
						Names: nil,
						Type:  generalSlice,
					},
				}}
				emitCall("runtime.makeSlice", args, resultList)
				return
			default:
				throw(typeArg)
			}
		case gAppend:
			var sliceArg ast.Expr = eArgs[0]
			var elemArg ast.Expr = eArgs[1]
			var elmType *Type = getElementTypeOfListType(getTypeOfExpr(sliceArg))
			var elmSize int = getSizeOfType(elmType)

			var args []*Arg = []*Arg{
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

			var resultList = &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: nil,
						Type:  generalSlice,
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
			// builtin print
			_args := []*Arg{&Arg{
				e: eArgs[0],
			}}
			var symbol string
			switch kind(getTypeOfExpr(eArgs[0])) {
			case T_STRING:
				symbol = "runtime.printstring"
			case T_INT:
				symbol = "runtime.printint"
			default:
				unexpectedKind(kind(getTypeOfExpr(eArgs[0])))
			}
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
			funcdecl := lookupForeignFunc(qi)
			funcType = funcdecl.decl.Type
		} else {
			// Assume method call
			receiver = fn.X
			receiverType := getTypeOfExpr(receiver)
			method := lookupMethod(receiverType, fn.Sel)
			funcType = method.funcType
			symbol = getMethodSymbol(method)
		}
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
	emitExpr(lit, ctx)
}

type okContext struct {
	needMain bool
	needOk   bool
}

type evalContext struct {
	okContext *okContext
	_type     *Type
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
	var isNilObj bool
	emitComment(2, "[emitExpr] dtype=%T\n", expr)
	switch e := expr.(type) {
	case *ast.Ident: // 1 value
		switch e.Obj {
		case gTrue: // true constant
			emitTrue()
		case gFalse: // false constant
			emitFalse()
		case gNil:
			assert(ctx._type != nil, "context of nil is not passed")
			emitNil(ctx._type)
			isNilObj = true
		default:
			assert(e.Obj != nil, "should not be nil")
			switch e.Obj.Kind {
			case ast.Var:
				emitAddr(e)
				emitLoadAndPush(getTypeOfExpr(e))
			case ast.Con:
				emitNamedConst(e, ctx)
			default:
				panic("Unexpected ident kind:" + e.Obj.Kind.String())
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
			emitAddr(e)
			emitLoadAndPush(getTypeOfExpr(e))
		}
	case *ast.CallExpr: // multi values Fun(Args)
		var fun = e.Fun
		// check if it's a conversion
		if isType(fun) {
			emitConversion(e2t(fun), e.Args[0])
		} else {
			emitFuncall(fun, e.Args, e.Ellipsis != token.NoPos)
		}
	case *ast.ParenExpr: // multi values (e)
		emitExpr(e.X, ctx)
	case *ast.BasicLit: // 1 value
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
	case *ast.UnaryExpr: // 1 value
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
			throw(e.Op.String())
		}
	case *ast.BinaryExpr: // 1 value
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
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				emitExpr(e.Y, nil) // right
				fmt.Printf("  popq %%rcx # right\n")
				fmt.Printf("  popq %%rax # left\n")
				fmt.Printf("  addq %%rcx, %%rax\n")
				fmt.Printf("  pushq %%rax\n")
			}
		case "-":
			emitComment(2, "start %T\n", e)
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			fmt.Printf("  popq %%rcx # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  subq %%rcx, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "*":
			emitComment(2, "start %T\n", e)
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			fmt.Printf("  popq %%rcx # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  imulq %%rcx, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "%":
			emitComment(2, "start %T\n", e)
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			fmt.Printf("  popq %%rcx # right\n")
			fmt.Printf("  popq %%rax # left\n")
			fmt.Printf("  movq $0, %%rdx # init %%rdx\n")
			fmt.Printf("  divq %%rcx\n")
			fmt.Printf("  movq %%rdx, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		case "/":
			emitComment(2, "start %T\n", e)
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
			emitComment(2, "start %T\n", e)
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			emitCompExpr("setl")
		case "<=":
			emitComment(2, "start %T\n", e)
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			emitCompExpr("setle")
		case ">":
			emitComment(2, "start %T\n", e)
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			emitCompExpr("setg")
		case ">=":
			emitComment(2, "start %T\n", e)
			emitExpr(e.X, nil) // left
			emitExpr(e.Y, nil) // right
			emitCompExpr("setge")
		default:
			panic(fmt.Sprintf("TBI: binary operation for '%s'", e.Op.String()))
		}
	case *ast.CompositeLit: // 1 value
		// slice , array, map or struct
		switch kind(e2t(e.Type)) {
		case T_STRUCT:
			emitStructLiteral(e)
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
			unexpectedKind(kind(e2t(e.Type)))
		}
	case *ast.SliceExpr: // 1 value list[low:high]
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

		// a missing high index defaults to the length of the sliced operand:
		// @TODO
		switch kind(listType) {
		case T_SLICE, T_ARRAY:
			// len = high - low
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
			emitExpr(low, nil) // intval
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
	case *ast.TypeAssertExpr:
		emitExpr(e.X, nil)
		fmt.Printf("  popq  %%rax # ifc.dtype\n")
		fmt.Printf("  popq  %%rcx # ifc.data\n")
		fmt.Printf("  pushq %%rax # ifc.data\n")

		typ := e2t(e.Type)
		sType := serializeType(typ)
		typeId := getTypeId(sType)
		typeSymbol := typeIdToSymbol(typeId)
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
	default:
		throw(expr)
	}

	return isNilObj
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

var typeMap map[string]int = map[string]int{}
var typeId int = 1

func typeIdToSymbol(id int) string {
	return "dtype." + strconv.Itoa(id)
}

func getTypeId(s string) int {
	id, ok := typeMap[s]
	if !ok {
		typeMap[s] = typeId
		r := typeId
		typeId++
		return r
	}
	return id
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
		ValuePos: 0,
		Kind:     token.INT,
		Value:    fmt.Sprintf("%d", x),
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
			offset:    0,
		},
		&Arg{
			e:         right,
			paramType: tString,
			offset:    0,
		},
	}

	fList := &ast.FieldList{List: []*ast.Field{
		&ast.Field{
			Names: nil,
			Type:  tString.E,
		},
	}}
	emitCall("runtime.catstrings", args, fList)
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

	rList := &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Names: nil,
				Type:  tBool.E,
			},
		},
	}

	emitCall("runtime.cmpstrings", args, rList)
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
	emitComment(2, "emitStore\n")
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
		_type: nil,
	}
	emitExprIfc(rhs, ctx)
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
	emitComment(2, "Assignment: emitStore(getTypeOfExpr(lhs))\n")
	emitStore(getTypeOfExpr(lhs), true, false)
}

func emitStmt(stmt ast.Stmt) {
	emitComment(2, "== Statement %T ==\n", stmt)
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		for _, s := range s.List {
			emitStmt(s)
		}
	case *ast.ExprStmt:
		emitExpr(s.X, nil)
	case *ast.DeclStmt:
		genDecl := s.Decl.(*ast.GenDecl)
		declSpec := genDecl.Specs[0]
		switch spec := declSpec.(type) {
		case *ast.ValueSpec:
			valSpec := spec
			t := e2t(valSpec.Type)
			lhs := valSpec.Names[0]
			if len(valSpec.Values) == 0 {
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
	case *ast.AssignStmt:
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
					callExpr, ok := rhs0.(*ast.CallExpr)
					assert(ok, "should be a CallExpr")
					returnTypes := getCallResultTypes(callExpr)
					fmt.Printf("# len lhs=%d\n", len(s.Lhs))
					fmt.Printf("# returnTypes=%d\n", len(returnTypes))
					assert(len(returnTypes) == len(s.Lhs), fmt.Sprintf("length unmatches %d <=> %d", len(s.Lhs), len(returnTypes)))
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
							default:
							}
							emitAddr(lhs)
							emitStore(getTypeOfExpr(lhs), false, false)
						}
					}

				}
			}
		default:
			panic("TBI: assignment of " + s.Tok.String())
		}
	case *ast.ReturnStmt:
		emitReturnStmt(s)
	case *ast.IfStmt:
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
	case *ast.ForStmt:
		labelid++
		labelCond := fmt.Sprintf(".L.for.cond.%d", labelid)
		labelPost := fmt.Sprintf(".L.for.post.%d", labelid)
		labelExit := fmt.Sprintf(".L.for.exit.%d", labelid)
		forStmt, ok := mapForNodeToFor[s]
		assert(ok, "map value should exist")
		forStmt.labelPost = labelPost
		forStmt.labelExit = labelExit

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
		labelCond := fmt.Sprintf(".L.range.cond.%d", labelid)
		labelPost := fmt.Sprintf(".L.range.post.%d", labelid)
		labelExit := fmt.Sprintf(".L.range.exit.%d", labelid)

		forStmt, ok := mapRangeNodeToFor[s]
		assert(ok, "map value should exist")
		forStmt.labelPost = labelPost
		forStmt.labelExit = labelExit
		// initialization: store len(rangeexpr)
		emitComment(2, "ForRange Initialization\n")

		emitComment(2, "  assign length to lenvar\n")
		// lenvar = len(s.X)
		rngMisc, ok := mapRangeStmt[s]
		assert(ok, "lenVar should exist")
		// lenvar = len(s.X)
		emitVariableAddr(rngMisc.lenvar)
		emitLen(s.X)
		emitStore(tInt, true, false)

		emitComment(2, "  assign 0 to indexvar\n")
		// indexvar = 0
		emitVariableAddr(rngMisc.indexvar)
		emitZeroValue(tInt)
		emitStore(tInt, true, false)

		// init key variable with 0
		if s.Key != nil {
			keyIdent, ok := s.Key.(*ast.Ident)
			assert(ok, "key expr should be an ident")
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

		emitVariableAddr(rngMisc.indexvar)
		emitLoadAndPush(tInt)
		emitVariableAddr(rngMisc.lenvar)
		emitLoadAndPush(tInt)
		emitCompExpr("setl")
		emitPopBool(" indexvar < lenvar")
		fmt.Printf("  cmpq $1, %%rax\n")
		fmt.Printf("  jne %s # jmp if false\n", labelExit)

		emitComment(2, "assign list[indexvar] value variables\n")
		elemType := getTypeOfExpr(s.Value)
		emitAddr(s.Value) // lhs

		emitVariableAddr(rngMisc.indexvar)
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
		emitVariableAddr(rngMisc.indexvar) // lhs
		emitVariableAddr(rngMisc.indexvar) // rhs
		emitLoadAndPush(tInt)
		emitAddConst(1, "indexvar value ++")
		emitStore(tInt, true, false)

		// incr key variable
		if s.Key != nil {
			keyIdent, ok := s.Key.(*ast.Ident)
			assert(ok, "key expr should be an ident")
			if keyIdent.Name != "_" {
				emitAddr(s.Key)                    // lhs
				emitVariableAddr(rngMisc.indexvar) // rhs
				emitLoadAndPush(tInt)
				emitStore(tInt, true, false)
			}
		}

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
		emitStore(getTypeOfExpr(s.X), true, false)
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
		var labels = make([]string, len(cases))
		var defaultLabel string
		emitComment(2, "Start comparison with cases\n")
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
			cc, ok := c.(*ast.CaseClause)
			assert(ok, "should be *ast.CaseClause")
			fmt.Printf("%s:\n", labels[i])
			for _, _s := range cc.Body {
				emitStmt(_s)
			}
			fmt.Printf("  jmp %s\n", labelEnd)
		}
		fmt.Printf("%s:\n", labelEnd)
	case *ast.TypeSwitchStmt:
		typeSwitch, ok := mapTypeSwitchStmtMeta[s]
		assert(ok, "should exist")
		labelid++
		labelEnd := fmt.Sprintf(".L.typeswitch.%d.exit", labelid)

		// subjectVariable = subject
		emitVariableAddr(typeSwitch.SubjectVariable)
		emitExpr(typeSwitch.Subject, nil)
		emitStore(tEface, true, false)

		cases := s.Body.List
		var labels = make([]string, len(cases))
		var defaultLabel string
		emitComment(2, "Start comparison with cases\n")
		for i, c := range cases {
			cc, ok := c.(*ast.CaseClause)
			assert(ok, "should be *ast.CaseClause")
			labelid++
			labelCase := ".L.case." + strconv.Itoa(labelid)
			labels[i] = labelCase
			if len(cc.List) == 0 {
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
			// Injecting variable and type to the subject
			if typeSwitchCaseClose.Variable != nil {
				setVariable(typeSwitch.AssignIdent.Obj, typeSwitchCaseClose.Variable)
			}
			fmt.Printf("%s:\n", labels[i])

			for _, _s := range typeSwitchCaseClose.Orig.Body {
				if typeSwitchCaseClose.Variable != nil {
					// do assignment
					emitAddr(typeSwitch.AssignIdent)

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

func getMethodSymbol(method *Method) string {
	rcvTypeName := method.rcvNamedType
	var subsymbol string
	if method.isPtrMethod {
		subsymbol = "$" + rcvTypeName.Name + "." + method.name // pointer
	} else {
		subsymbol = rcvTypeName.Name + "." + method.name // value
	}

	return getPackageSymbol(method.pkgName, subsymbol)
}

func getPackageSymbol(pkgPrefix string, subsymbol string) string {
	return pkgPrefix + "." + subsymbol
}

func emitFuncDecl(pkgPrefix string, fnc *Func) {
	fmt.Printf("# emitFuncDecl\n")
	if len(fnc.params) > 0 {
		for i := 0; i < len(fnc.params); i++ {
			v := fnc.params[i]
			logf("  #       params %d %d \"%s\" %s\n", int(v.LocalOffset), getSizeOfType(v.Typ), v.Name, string(kind(v.Typ)))
		}
	}
	if len(fnc.retvars) > 0 {
		for i := 0; i < len(fnc.retvars); i++ {
			v := fnc.retvars[i]
			logf("  #       retvars %d %d \"%s\" %s\n", int(v.LocalOffset), getSizeOfType(v.Typ), v.Name, string(kind(v.Typ)))
		}
	}

	var symbol string
	if fnc.method != nil {
		symbol = getMethodSymbol(fnc.method)
	} else {
		symbol = getPackageSymbol(pkgPrefix, fnc.name)
	}
	fmt.Printf("%s: # args %d, locals %d\n", symbol, int(fnc.argsarea), int(fnc.localarea))
	fmt.Printf("  pushq %%rbp\n")
	fmt.Printf("  movq %%rsp, %%rbp\n")
	if len(fnc.localvars) > 0 {
		for i := len(fnc.localvars) - 1; i >= 0; i-- {
			v := fnc.localvars[i]
			logf("  # -%d(%%rbp) local variable %d \"%s\"\n", -int(v.LocalOffset), getSizeOfType(v.Typ), v.Name)
		}
	}
	logf("  #  0(%%rbp) previous rbp\n")
	logf("  #  8(%%rbp) return address\n")

	if int(fnc.localarea) != 0 {
		fmt.Printf("  subq $%d, %%rsp # local area\n", int(-int(fnc.localarea)))

	}
	for _, stmt := range fnc.stmts {
		emitStmt(stmt)
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
		switch vl := val.(type) {
		case nil:
			fmt.Printf("  .quad 0\n")
			fmt.Printf("  .quad 0\n")
		case *ast.BasicLit:
			sl := getStringLiteral(vl)
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
		switch vl := val.(type) {
		case nil:
			fmt.Printf("  .quad 0 # bool zero value\n")
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
		switch vl := val.(type) {
		case nil:
			fmt.Printf("  .quad 0\n")
		case *ast.BasicLit:
			fmt.Printf("  .quad %s\n", vl.Value)
		default:
			throw(val)
		}
	case T_UINT8:
		switch vl := val.(type) {
		case nil:
			fmt.Printf("  .byte 0\n")
		case *ast.BasicLit:
			fmt.Printf("  .byte %s\n", vl.Value)
		default:
			throw(val)
		}
	case T_UINT16:
		switch vl := val.(type) {
		case nil:
			fmt.Printf("  .word 0\n")
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
		arrayType, ok := t.E.(*ast.ArrayType)
		assert(ok, "should be *ast.ArrayType")
		assert(arrayType.Len != nil, "slice type is not expected")
		length := evalInt(arrayType.Len)
		var zeroValue string
		switch kind(e2t(arrayType.Elt)) {
		case T_INT:
			zeroValue = "  .quad 0 # int zero value\n"
		case T_UINT8:
			zeroValue = fmt.Sprintf("  .byte 0 # uint8 zero value\n")
		case T_STRING:
			zeroValue = fmt.Sprintf("  .quad 0 # string zero value (ptr)\n")
			zeroValue += fmt.Sprintf("  .quad 0 # string zero value (len)\n")
		case T_INTERFACE:
			zeroValue = fmt.Sprintf("  .quad 0 # eface zero value (dtype)\n")
			zeroValue += fmt.Sprintf("  .quad 0 # eface zero value (data)\n")
		default:
			unexpectedKind(kind(e2t(arrayType.Elt)))
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
		if t == nil {
			panic("type cannot be nil for global variable: " + spec.Names[0].Name)
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

	var fnc *Func
	for _, fnc = range pkg.funcs {
		emitFuncDecl(pkg.name, fnc)
	}

	fmt.Printf("\n")
}

func emitDynamicTypes(typeMap map[string]int) {
	fmt.Printf("# ------- Dynamic Types ------\n")
	fmt.Printf(".data\n")

	sliceTypeMap := make([]string, len(typeMap)+1)

	// sort map in order to assure the deterministic results
	for name, id := range typeMap {
		sliceTypeMap[id] = name
	}
	for id, name := range sliceTypeMap {
		if name == "" {
			continue
		}
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
type Type struct {
	E ast.Expr // original expr
}

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

var tBool *Type = &Type{
	E: &ast.Ident{
		NamePos: 0,
		Name:    "bool",
		Obj:     gBool,
	},
}

var tInt *Type = &Type{
	E: &ast.Ident{
		NamePos: 0,
		Name:    "int",
		Obj:     gInt,
	},
}

// Rune
var tInt32 *Type = &Type{
	E: &ast.Ident{
		NamePos: 0,
		Name:    "int",
		Obj:     gInt32,
	},
}

var tUintptr *Type = &Type{
	E: &ast.Ident{
		NamePos: 0,
		Name:    "uintptr",
		Obj:     gUintptr,
	},
}

var tUint8 *Type = &Type{
	E: &ast.Ident{
		NamePos: 0,
		Name:    "uint8",
		Obj:     gUint8,
	},
}

var tByte *Type = &Type{
	E: &ast.Ident{
		NamePos: 0,
		Name:    "byte",
		Obj:     gUint8,
	},
}

var tSliceOfString *Type = &Type{
	E: &ast.ArrayType{
		Len: nil,
		Elt: &ast.Ident{
			NamePos: 0,
			Name:    "string",
			Obj:     gString,
		},
	},
}

var tString *Type = &Type{
	E: &ast.Ident{
		NamePos: 0,
		Name:    "string",
		Obj:     gString,
	},
}

var tEface *Type = &Type{
	E: &ast.InterfaceType{},
}

var generalSlice ast.Expr = &ast.Ident{}

func getTypeOfExpr(expr ast.Expr) *Type {
	switch e := expr.(type) {
	case *ast.Ident:
		assert(e.Obj != nil, "Obj is nil in ident '"+e.Name+"'")
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
				throw(e.Obj.Decl)
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
			throw(e.Obj)
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
			throw(e.Kind.String())
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
			var starExpr = &ast.StarExpr{}
			var t = getTypeOfExpr(e.X)
			starExpr.X = t.E
			return e2t(starExpr)
		case "range":
			listType := getTypeOfExpr(e.X)
			elmType := getElementTypeOfListType(listType)
			return elmType
		default:
			throw(e.Op.String())
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
		assert(len(types) == 1, "single value is expected")
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
		ptrType, ok := t.E.(*ast.StarExpr)
		if !ok {
			throw(t)
		}
		return e2t(ptrType.X)
	case *ast.SelectorExpr:
		emitComment(2, "getTypeOfExpr(X.%s)\n", e.Sel.Name)
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
		return e2t(e.Type)
	default:
		throw(expr)
	}
	throw(expr)
	return nil
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
				return []*Type{e2t(&ast.StarExpr{
					Star: 0,
					X:    e.Args[0],
				})}
			case gMake:
				return []*Type{e2t(e.Args[0])}
			case gAppend:
				return []*Type{e2t(e.Args[0])}
			}
			switch decl := fn.Obj.Decl.(type) {
			case *ast.FuncDecl:
				return fieldList2Types(decl.Type.Results)
			default:
				throw(fn.Obj)
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
			return fieldList2Types(method.funcType.Results)
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
				typeSpec, ok := decl.(*ast.TypeSpec)
				if !ok {
					throw(decl)
				}
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
	case *ast.SelectorExpr:
		qi := selector2QI(e)
		return string(qi)
	default:
		throw(t)
	}
	return ""
}

func kind(t *Type) TypeKind {
	if t == nil {
		panic("nil type is not expected")
	}
	if t.E == generalSlice {
		return T_SLICE
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
	case *ast.InterfaceType:
		return T_INTERFACE
	case *ast.SelectorExpr:
		ident := lookupForeignIdent(selector2QI(e))
		return kind(e2t(ident))
	default:
		throw(t)
	}
	return ""
}

func isInterface(t *Type) bool {
	return kind(t) == T_INTERFACE
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
		ptrType, ok := typeOfX.E.(*ast.StarExpr)
		assert(ok, "should be *ast.StarExpr")
		structType = e2t(ptrType.X)
	default:
		unexpectedKind(kind(typeOfX))
	}
	return structType
}

func getElementTypeOfListType(t *Type) *Type {
	switch kind(t) {
	case T_SLICE, T_ARRAY:
		switch tt := t.E.(type) {
		case *ast.ArrayType:
			return e2t(tt.Elt)
		case *ast.Ellipsis:
			return e2t(tt.Elt)
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
		elmSize := getSizeOfType(e2t(arrayType.Elt))
		return elmSize * evalInt(arrayType.Len)
	case T_STRUCT:
		return calcStructSizeAndSetFieldOffset(getStructTypeSpec(t))
	default:
		unexpectedKind(kind(t))
	}
	return 0
}

func getStructFieldOffset(field *ast.Field) int {
	if field.Doc == nil {
		panic("Doc is nil:" + field.Names[0].Name)
	}
	text := field.Doc.List[0].Text
	offset := strconv.Atoi(text)
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

func getStructTypeSpec(typ *Type) *ast.TypeSpec {
	if kind(typ) != T_STRUCT {
		throw(typ)
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

	typeSpec, ok := typeName.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		throw(typeName.Obj.Decl)
	}
	return typeSpec

}

func lookupStructField(structTypeSpec *ast.TypeSpec, selName string) *ast.Field {
	for _, field := range getStructFields(structTypeSpec) {
		if field.Names[0].Name == selName {
			return field
		}
	}
	panic("Unexpected flow: struct field not found:" + selName)
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

type ForStmt struct {
	kind      int    // 1:for 2:range
	labelPost string // for continue
	labelExit string // for break
	outer     *ForStmt
	astFor    *ast.ForStmt
	astRange  *ast.RangeStmt
}

type TypeSwitchStmt struct {
	Subject         ast.Expr
	SubjectVariable *Variable
	AssignIdent     *ast.Ident
	Cases           []*TypeSwitchCaseClose
}

type TypeSwitchCaseClose struct {
	Variable     *Variable
	VariableType *Type
	Orig         *ast.CaseClause
}

type nodeReturnStmt struct {
	fnc *Func
}

type RangeStmtMisc struct {
	lenvar   *Variable
	indexvar *Variable
}

type Func struct {
	name      string
	stmts     []ast.Stmt
	localarea localoffsetint
	argsarea  localoffsetint
	localvars []*Variable
	params    []*Variable
	retvars   []*Variable
	funcType  *ast.FuncType
	method    *Method
}

type Method struct {
	pkgName      string
	rcvNamedType *ast.Ident
	isPtrMethod  bool
	name         string
	funcType     *ast.FuncType
}

type Variable struct {
	Name         string
	IsGlobal     bool
	GlobalSymbol string
	LocalOffset  localoffsetint
	Typ          *Type
}

type localoffsetint int

func (fnc *Func) registerParamVariable(name string, t *Type) *Variable {
	vr := newLocalVariable(name, fnc.argsarea, t)
	size := getSizeOfType(t)
	fnc.argsarea += localoffsetint(size)
	fnc.params = append(fnc.params, vr)
	return vr
}

func (fnc *Func) registerReturnVariable(name string, t *Type) *Variable {
	vr := newLocalVariable(name, fnc.argsarea, t)
	size := getSizeOfType(t)
	fnc.argsarea += localoffsetint(size)
	fnc.retvars = append(fnc.retvars, vr)
	return vr
}

func (fnc *Func) registerLocalVariable(name string, t *Type) *Variable {
	assert(t != nil && t.E != nil, "type of local var should not be nil")
	fnc.localarea -= localoffsetint(getSizeOfType(t))
	vr := newLocalVariable(name, currentFunc.localarea, t)
	fnc.localvars = append(fnc.localvars, vr)
	return vr
}

var currentFor *ForStmt

var mapForNodeToFor map[*ast.ForStmt]*ForStmt = map[*ast.ForStmt]*ForStmt{}
var mapRangeNodeToFor map[*ast.RangeStmt]*ForStmt = map[*ast.RangeStmt]*ForStmt{}
var mapBranchToFor map[*ast.BranchStmt]*ForStmt = map[*ast.BranchStmt]*ForStmt{}
var mapRangeStmt map[*ast.RangeStmt]*RangeStmtMisc = map[*ast.RangeStmt]*RangeStmtMisc{}
var mapTypeSwitchStmtMeta = map[*ast.TypeSwitchStmt]*TypeSwitchStmt{}
var mapReturnStmt = map[*ast.ReturnStmt]*nodeReturnStmt{}

var currentFunc *Func

func getStringLiteral(lit *ast.BasicLit) *sliteral {
	for _, container := range currentPkg.stringLiterals {
		if container.lit == lit {
			return container.sl
		}
	}

	panic(lit.Value)
	return nil
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
	var cont *stringLiteralsContainer = new(stringLiteralsContainer)
	cont.sl = sl
	cont.lit = lit
	currentPkg.stringLiterals = append(currentPkg.stringLiterals, cont)
}

func newGlobalVariable(pkgName string, name string, t *Type) *Variable {
	return &Variable{
		Name:         name,
		IsGlobal:     true,
		GlobalSymbol: pkgName + "." + name,
		LocalOffset:  0,
		Typ:          t,
	}
}

func newLocalVariable(name string, localoffset localoffsetint, t *Type) *Variable {
	return &Variable{
		Name:         name,
		IsGlobal:     false,
		GlobalSymbol: "",
		LocalOffset:  localoffset,
		Typ:          t,
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
	pkgName, isIdent := e.X.(*ast.Ident)
	if !isIdent {
		throw(e)
	}
	assert(pkgName.Obj.Kind == ast.Pkg, "should be ast.Pkg")
	return newQI(pkgName.Name, e.Sel.Name)
}

// https://golang.org/ref/spec#Method_sets
// @TODO map key should be a QI ?
var MethodSets = map[*ast.Object]map[string]*Method{}

func newMethod(pkgName string, funcDecl *ast.FuncDecl) *Method {
	rcvType := funcDecl.Recv.List[0].Type
	rcvPointerType, ok := rcvType.(*ast.StarExpr)
	var isPtr bool
	if ok {
		isPtr = true
		rcvType = rcvPointerType.X
	}
	rcvNamedType, ok := rcvType.(*ast.Ident)
	if !ok {
		throw(rcvType)
	}
	method := &Method{
		pkgName:      pkgName,
		rcvNamedType: rcvNamedType,
		isPtrMethod:  isPtr,
		name:         funcDecl.Name.Name,
		funcType:     funcDecl.Type,
	}
	return method
}

func registerMethod(method *Method) {
	methodSet, ok := MethodSets[method.rcvNamedType.Obj]
	if !ok {
		methodSet = map[string]*Method{}
		MethodSets[method.rcvNamedType.Obj] = methodSet
	}
	methodSet[method.name] = method
}

func lookupMethod(rcvT *Type, methodName *ast.Ident) *Method {
	rcvType := rcvT.E
	rcvPointerType, ok := rcvType.(*ast.StarExpr)
	if ok {
		rcvType = rcvPointerType.X
	}
	var methodSet map[string]*Method
	switch typ := rcvType.(type) {
	case *ast.Ident:
		var ok bool
		methodSet, ok = MethodSets[typ.Obj]
		if !ok {
			panic(typ.Name + " has no methodSet (1)")
		}
	case *ast.SelectorExpr:
		t := lookupForeignIdent(selector2QI(typ))
		var ok bool
		methodSet, ok = MethodSets[t.Obj]
		if !ok {
			panic(t.Name + " has no methodSet (2)")
		}
	}

	method, ok := methodSet[methodName.Name]
	if !ok {
		panic("method not found")
	}
	return method
}


func walkExprStmt(s *ast.ExprStmt) {
	expr := s.X
	walkExpr(expr)
}
func walkDeclStmt(s *ast.DeclStmt) {
	decl := s.Decl
	switch dcl := decl.(type) {
	case *ast.GenDecl:
		declSpec := dcl.Specs[0]
		switch ds := declSpec.(type) {
		case *ast.ValueSpec:
			varSpec := ds
			obj := varSpec.Names[0].Obj
			if varSpec.Type == nil { // var x = e
				if len(ds.Values) > 0 {
					// infer type from rhs
					val := ds.Values[0]
					logf("nfering type of variable %s\n", obj.Name)
					typ := getTypeOfExpr(val)
					if typ != nil && typ.E != nil {
						varSpec.Type = typ.E
					}
				}
			}

			t := e2t(varSpec.Type)
			setVariable(obj, currentFunc.registerLocalVariable(obj.Name, t))
			for _, v := range ds.Values {
				walkExpr(v)
			}
		}
	default:
		throw(decl)
	}
}
func walkAssignStmt(s *ast.AssignStmt) {
	if s.Tok.String() == ":=" {
		// short var decl
		rhs0 := s.Rhs[0]
		walkExpr(rhs0)
		// infer type
		var typ0 *Type
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
				setVariable(okObj, currentFunc.registerLocalVariable(okObj.Name, tBool))
			}
		default:
			typ0 = getTypeOfExpr(rhs0)
		}

		if typ0 != nil && typ0.E != nil {
		} else {
			panic("type inference is not supported")
		}
		obj0 := s.Lhs[0].(*ast.Ident).Obj
		setVariable(obj0, currentFunc.registerLocalVariable(obj0.Name, typ0))
	} else {
		walkExpr(s.Rhs[0])
	}
}
func walkReturnStmt(s *ast.ReturnStmt) {
	mapReturnStmt[s] = &nodeReturnStmt{
		fnc: currentFunc,
	}
	for _, r := range s.Results {
		walkExpr(r)
	}
}
func walkIfStmt(s *ast.IfStmt) {
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
	forStmt := new(ForStmt)
	forStmt.astFor = s
	forStmt.outer = currentFor
	currentFor = forStmt
	mapForNodeToFor[s] = forStmt
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
	currentFor = forStmt.outer
}
func walkRangeStmt(s *ast.RangeStmt) {
	forStmt := new(ForStmt)
	forStmt.astRange = s
	forStmt.outer = currentFor
	currentFor = forStmt
	mapRangeNodeToFor[s] = forStmt
	walkExpr(s.X)
	walkStmt(s.Body)
	lenvar := currentFunc.registerLocalVariable(".range.len", tInt)
	indexvar := currentFunc.registerLocalVariable(".range.index", tInt)
	if s.Tok.String() == ":=" {
		// short var decl
		listType := getTypeOfExpr(s.X)

		keyIdent := s.Key.(*ast.Ident)
		//@TODO map key can be any type
		//keyType := getKeyTypeOfListType(listType)
		keyType := tInt
		setVariable(keyIdent.Obj, currentFunc.registerLocalVariable(keyIdent.Name, keyType))

		// determine type of Value
		elmType := getElementTypeOfListType(listType)
		valueIdent := s.Value.(*ast.Ident)
		setVariable(valueIdent.Obj, currentFunc.registerLocalVariable(valueIdent.Name, elmType))
	}
	mapRangeStmt[s] = &RangeStmtMisc{
		lenvar:   lenvar,
		indexvar: indexvar,
	}
	currentFor = forStmt.outer
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
	typeSwitch := &TypeSwitchStmt{}
	mapTypeSwitchStmtMeta[s] = typeSwitch
	if s.Init != nil {
		walkStmt(s.Init)
	}
	var assignIdent *ast.Ident
	switch assign := s.Assign.(type) {
	case *ast.ExprStmt:
		typeAssertExpr, ok := assign.X.(*ast.TypeAssertExpr)
		assert(ok, "should be *ast.TypeAssertExpr")
		typeSwitch.Subject = typeAssertExpr.X
		walkExpr(typeAssertExpr.X)
	case *ast.AssignStmt:
		lhs := assign.Lhs[0]
		var ok bool
		assignIdent, ok = lhs.(*ast.Ident)
		assert(ok, "lhs should be ident")
		typeSwitch.AssignIdent = assignIdent
		// ident will be a new local variable in each case clause
		typeAssertExpr, ok := assign.Rhs[0].(*ast.TypeAssertExpr)
		assert(ok, "should be *ast.TypeAssertExpr")
		typeSwitch.Subject = typeAssertExpr.X
		walkExpr(typeAssertExpr.X)
	default:
		throw(s.Assign)
	}

	typeSwitch.SubjectVariable = currentFunc.registerLocalVariable(".switch_expr", tEface)
	for _, _case := range s.Body.List {
		cc := _case.(*ast.CaseClause)
		tscc := &TypeSwitchCaseClose{
			Orig: cc,
		}
		typeSwitch.Cases = append(typeSwitch.Cases, tscc)
		if assignIdent != nil && len(cc.List) > 0 {
			// inject a variable of that type
			varType := e2t(cc.List[0])
			vr := currentFunc.registerLocalVariable(assignIdent.Name, varType)
			tscc.Variable = vr
			tscc.VariableType = varType
			setVariable(assignIdent.Obj, vr)
		}

		for _, stmt := range cc.Body {
			walkStmt(stmt)
		}

		if assignIdent != nil {
			assignIdent.Obj.Data = nil
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
func walkBranchStmt(s *ast.BranchStmt) {
	assert(currentFor != nil, "break or continue should be in for body")
	mapBranchToFor[s] = currentFor
}

func walkStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.ExprStmt: walkExprStmt(s)
	case *ast.DeclStmt: walkDeclStmt(s)
	case *ast.AssignStmt: walkAssignStmt(s)
	case *ast.ReturnStmt: walkReturnStmt(s)
	case *ast.IfStmt: walkIfStmt(s)
	case *ast.BlockStmt: walkBlockStmt(s)
	case *ast.ForStmt: walkForStmt(s)
	case *ast.RangeStmt: walkRangeStmt(s)
	case *ast.IncDecStmt: walkIncDecStmt(s)
	case *ast.SwitchStmt: walkSwitchStmt(s)
	case *ast.TypeSwitchStmt: walkTypeSwitchStmt(s)
	case *ast.CaseClause: walkCaseClause(s)
	case *ast.BranchStmt: walkBranchStmt(s)
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
					ValuePos: 0,
					Kind:     token.STRING,
					Value:    "\"" + currentFunc.name + "\"",
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
func walkArrayType(e *ast.ArrayType) {
	// first argument of builtin func like make()
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
	case *ast.Ident: walkIdent(e)
	case *ast.SelectorExpr: walkSelectorExpr(e)
	case *ast.CallExpr: walkCallExpr(e)
	case *ast.ParenExpr: walkParenExpr(e)
	case *ast.BasicLit: walkBasicLit(e)
	case *ast.CompositeLit: walkCompositeLit(e)
	case *ast.UnaryExpr: walkUnaryExpr(e)
	case *ast.BinaryExpr: walkBinaryExpr(e)
	case *ast.IndexExpr: walkIndexExpr(e)
	case *ast.SliceExpr: walkSliceExpr(e)
	case *ast.ArrayType: walkArrayType(e)
	case *ast.StarExpr: walkStarExpr(e)
	case *ast.KeyValueExpr: walkKeyValueExpr(e)
	case *ast.InterfaceType: walkInterfaceType(e)
	case *ast.TypeAssertExpr: walkTypeAssertExpr(e)
	default:
		throw(expr)
	}
}

var ExportedQualifiedIdents = map[QualifiedIdent]*ast.Ident{}

func lookupForeignIdent(qi QualifiedIdent) *ast.Ident {
	ident, found := ExportedQualifiedIdents[qi]
	if !found {
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
	if ident.Obj.Kind != ast.Fun {
		panic("Not ast.Fun: " + qi)
	}
	decl, ok := ident.Obj.Decl.(*ast.FuncDecl)
	if !ok {
		panic("Function not found: " + qi)
	}
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
		switch kind(e2t(typeSpec.Type)) {
		case T_STRUCT:
			calcStructSizeAndSetFieldOffset(typeSpec)
		}
		ExportedQualifiedIdents[newQI(pkg.name, typeSpec.Name.Name)] = typeSpec.Name
	}

	// collect methods in advance
	for _, funcDecl := range funcDecls {
		if funcDecl.Recv == nil {
			qi := newQI(pkg.name, funcDecl.Name.Name)
			logf("ExportedQualifiedIdents added: %s\n", string(qi))
			ExportedQualifiedIdents[qi] = funcDecl.Name
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
		assert(nameIdent.Obj.Kind == ast.Var, "should be Var")
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
		ExportedQualifiedIdents[newQI(pkg.name, nameIdent.Obj.Name)] = nameIdent
		for _, v := range varSpec.Values {
			// mainly to collect string literals
			walkExpr(v)
		}
	}

	for _, funcDecl := range funcDecls {
		fnc := &Func{
			name:      funcDecl.Name.Name,
			funcType:  funcDecl.Type,
			localarea: 0,
			argsarea:  16, // return address + previous rbp
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
			setVariable(obj, fnc.registerParamVariable(obj.Name, e2t(field.Type)))
		}

		for i, field := range resultFields {
			if len(field.Names) == 0 {
				// unnamed retval
				fnc.registerReturnVariable(".r"+strconv.Itoa(i), e2t(field.Type))
			} else {
				panic("TBI: named return variable is not supported")
			}
		}

		if funcDecl.Body != nil {
			fnc.stmts = funcDecl.Body.List
			for _, stmt := range fnc.stmts {
				walkStmt(stmt)
			}

			if funcDecl.Recv != nil { // is Method
				fnc.method = newMethod(pkg.name, funcDecl)
			}
			pkg.funcs = append(pkg.funcs, fnc)
		}
	}
}

// --- universe ---
var gNil = &ast.Object{
	Kind: ast.Con, // is nil a constant ?
	Name: "nil",
	Decl: nil,
	Data: nil,
	Type: nil,
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

var gPanic = &ast.Object{
	Kind: ast.Fun,
	Name: "panic",
	Decl: nil,
	Data: nil,
	Type: nil,
}

func createUniverse() *ast.Scope {
	universe := &ast.Scope{
		Outer:   nil,
		Objects: make(map[string]*ast.Object),
	}

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
	var mapImports = map[string]bool{}
	for _, imprt := range file.Imports {
		// unwrap double quote "..."
		rawValue := imprt.Path.Value
		pth := rawValue[1 : len(rawValue)-1]
		base := path.Base(pth)
		mapImports[base] = true
	}
	for _, ident := range file.Unresolved {
		// lookup imported package name
		if mapImports[ident.Name] {
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
	//fname := path.Base(dir) + ".go"
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

func getImportPathsFromFile(file string) map[string]bool {
	fset := &token.FileSet{}
	astFile0 := parseImports(fset, file)
	var paths = map[string]bool{}
	for _, importSpec := range astFile0.Imports {
		rawValue := importSpec.Path.Value
		logf("import %s\n", rawValue)
		path := rawValue[1 : len(rawValue)-1]
		paths[path] = true
	}
	return paths
}

func removeNode(tree map[string]map[string]bool, node string) {
	for _, paths := range tree {
		delete(paths, node)
	}

	delete(tree, node)
}

func getKeys(tree map[string]map[string]bool) []string {
	var keys []string
	for k, _ := range tree {
		keys = append(keys, k)
	}
	return keys
}

type DependencyTree map[string]map[string]bool

// Do topological sort
// In the result list, the independent (lowest level) packages come first.
func (tree DependencyTree) sortTopologically() []string {
	logf("sortTopologically start\n")
	var sorted []string
	for len(tree) > 0 {
		keys := getKeys(tree)
		mylib.SortStrings(keys)
		for _, _path := range keys {
			children := tree[_path]
			if len(children) == 0 {
				// leaf node
				logf("Found leaf node: %s\n", _path)
				sorted = append(sorted, _path)
				removeNode(tree, _path)
			}
		}

	}

	logf("sortTopologically end\n")
	return sorted
}

func getPackageDir(importPath string) string {
	if isStdLib(importPath) {
		return prjSrcPath + "/" + importPath
	} else {
		return srcPath + "/" + importPath
	}
}

func (tree DependencyTree) collectDependency(paths map[string]bool) {
	for pkgPath, _ := range paths {
		if pkgPath == "unsafe" || pkgPath == "runtime" {
			continue
		}
		logf("collectDependency in %s\n", pkgPath)
		packageDir := getPackageDir(pkgPath)
		fnames := findFilesInDir(packageDir)
		var children = map[string]bool{}
		for _, fname := range fnames {
			importPathsOfFile := getImportPathsFromFile(packageDir + "/" + fname)
			for _path, _ := range importPathsOfFile {
				if _path == "unsafe" || _path == "runtime" {
					continue
				}
				logf("  found %s\n", _path)
				children[_path] = true
			}
		}
		tree[pkgPath] = children
		tree.collectDependency(children)
	}
}

var srcPath string
var prjSrcPath string

func collectAllPackages(inputFiles []string) []string {
	var tree DependencyTree = map[string]map[string]bool{}
	directChildren := collectDirectDependents(inputFiles)
	tree.collectDependency(directChildren)
	sortedPaths := tree.sortTopologically()

	// sort packages by this order
	// 1: pseudo
	// 2: stdlib
	// 3: external
	paths := []string{"unsafe", "runtime"}
	for _, _path := range sortedPaths {
		if isStdLib(_path) {
			paths = append(paths, _path)
		}
	}
	for _, _path := range sortedPaths {
		if !isStdLib(_path) {
			paths = append(paths, _path)
		}
	}
	return paths
}

func collectDirectDependents(inputFiles []string) map[string]bool {
	importPaths := map[string]bool{}

	for _, inputFile := range inputFiles {
		logf("input file: \"%s\"\n", inputFile)
		logf("Parsing imports\n")
		_paths := getImportPathsFromFile(inputFile)
		for k, _ := range _paths {
			importPaths[k] = true
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
	fset := &token.FileSet{}
	pkgScope := ast.NewScope(universe)
	for _, file := range _pkg.files {
		logf("Parsing file: %s\n", file)
		astFile := parseFile(fset, file)
		_pkg.name = astFile.Name.Name
		_pkg.astFiles = append(_pkg.astFiles, astFile)
		for _, obj := range astFile.Scope.Objects {
			pkgScope.Objects[obj.Name] = obj
		}
	}
	for _, astFile := range _pkg.astFiles {
		resolveImports(astFile)
		var unresolved []*ast.Ident
		for _, ident := range astFile.Unresolved {
			if obj := pkgScope.Lookup(ident.Name); obj != nil {
				ident.Obj = obj
			} else {
				logf("# unresolved: %s\n", ident.Name)
				if obj := universe.Lookup(ident.Name); obj != nil {
					ident.Obj = obj
				} else {
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
	fmt.Printf("    pre version:  show version\n")
	fmt.Printf("    pre [-DF] [-DG] filename\n")
}

func main() {
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

// --- util ---
func obj2var(obj *ast.Object) *Variable {
	assert(obj.Kind == ast.Var, "should be ast.Var")
	vr, ok := obj.Data.(*Variable)
	if !ok {
		throw(obj.Data)
	}
	return vr
}

func setVariable(obj *ast.Object, vr *Variable) {
	obj.Data = vr
}
