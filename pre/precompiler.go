package main

import (
	"fmt"
	"syscall"

	"go/ast"
	"go/parser"
	"go/token"
)

// --- foundation ---
var __func__ = "__func__"

func assert(bol bool, msg string) {
	if !bol {
		panic(msg)
	}
}

func throw(x interface{}) {
	panic(fmt.Sprintf("%#v", x))
}

func panic2(caller string, x string) {
	panic("[" + caller + "] " + x)
}

var debugFrontEnd bool = true

func logf(format string, a ...string) {
	if !debugFrontEnd {
		return
	}
	var f = "# " + format
	var s = fmtSprintf(f, a)
	syscall.Write(1, []uint8(s))
}

// --- libs ---
func fmtSprintf(format string, a []string) string {
	var buf []uint8
	var inPercent bool
	var argIndex int
	for _, c := range []uint8(format) {
		if inPercent {
			if c == '%' {
				buf = append(buf, c)
			} else {
				arg := a[argIndex]
				argIndex++
				s := arg // // p.printArg(arg, c)
				for _, _c := range []uint8(s) {
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
	var s = fmtSprintf(format, a)
	syscall.Write(1, []uint8(s))
}

func Atoi(gs string) int {
	if len(gs) == 0 {
		return 0
	}
	var n int

	var isMinus bool
	for _, b := range []uint8(gs) {
		if b == '.' {
			return -999 // @FIXME all no number should return error
		}
		if b == '-' {
			isMinus = true
			continue
		}
		var x uint8 = b - uint8('0')
		n = n * 10
		n = n + int(x)
	}
	if isMinus {
		n = -n
	}

	return n
}

func Itoa(ival int) string {
	if ival == 0 {
		return "0"
	}

	var buf = make([]uint8, 100, 100)
	var r = make([]uint8, 100, 100)

	var next int
	var right int
	var ix = 0
	var minus bool
	minus = false
	for ix = 0; ival != 0; ix = ix + 1 {
		if ival < 0 {
			ival = -1 * ival
			minus = true
			r[0] = '-'
		} else {
			next = ival / 10
			right = ival - next*10
			ival = next
			buf[ix] = uint8('0' + right)
		}
	}

	var j int
	var c uint8
	for j = 0; j < ix; j = j + 1 {
		c = buf[ix-j-1]
		if minus {
			r[j+1] = c
		} else {
			r[j] = c
		}
	}

	return string(r[0:ix])
}

// --- parser ---
func parseFile(fset *token.FileSet, filename string) *ast.File {
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		panic(err)
	}
	return f
}

// --- codegen ---
var debugCodeGen bool = true

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

func evalInt(expr ast.Expr) int {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return Atoi(e.Value)
	}
	return 0
}

func emitPopPrimitive(comment string) {
	fmtPrintf("  popq %%rax # result of %s\n", comment)
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

func emitPopInterFace() {
	fmtPrintf("  popq %%rax # eface.dtype\n")
	fmtPrintf("  popq %%rcx # eface.data\n")
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
		throw(kind(condType))
	}
}

func emitRevertStackPointer(size int) {
	fmtPrintf("  addq $%s, %%rsp # revert stack pointer\n", Itoa(size))
}

func emitAddConst(addValue int, comment string) {
	emitComment(2, "Add const: %s\n", comment)
	fmtPrintf("  popq %%rax\n")
	fmtPrintf("  addq $%s, %%rax\n", Itoa(addValue))
	fmtPrintf("  pushq %%rax\n")
}

// "Load" means copy data from memory to registers
func emitLoadFromMemoryAndPush(t *Type) {
	if t == nil {
		panic2(__func__, "nil type error\n")
	}
	emitPopAddress(string(kind(t)))
	switch kind(t) {
	case T_SLICE:
		fmtPrintf("  movq %d(%%rax), %%rdx\n", Itoa(16))
		fmtPrintf("  movq %d(%%rax), %%rcx\n", Itoa(8))
		fmtPrintf("  movq %d(%%rax), %%rax\n", Itoa(0))
		fmtPrintf("  pushq %%rdx # cap\n")
		fmtPrintf("  pushq %%rcx # len\n")
		fmtPrintf("  pushq %%rax # ptr\n")
	case T_STRING:
		fmtPrintf("  movq %d(%%rax), %%rdx # len\n", Itoa(8))
		fmtPrintf("  movq %d(%%rax), %%rax # ptr\n", Itoa(0))
		fmtPrintf("  pushq %%rdx # len\n")
		fmtPrintf("  pushq %%rax # ptr\n")
	case T_INTERFACE:
		fmtPrintf("  movq %d(%%rax), %%rdx # data\n", Itoa(8))
		fmtPrintf("  movq %d(%%rax), %%rax # dtype\n", Itoa(0))
		fmtPrintf("  pushq %%rdx # data\n")
		fmtPrintf("  pushq %%rax # dtype\n")
	case T_UINT8:
		fmtPrintf("  movzbq %d(%%rax), %%rax # load uint8\n", Itoa(0))
		fmtPrintf("  pushq %%rax\n")
	case T_UINT16:
		fmtPrintf("  movzwq %d(%%rax), %%rax # load uint16\n", Itoa(0))
		fmtPrintf("  pushq %%rax\n")
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		fmtPrintf("  movq %d(%%rax), %%rax # load int\n", Itoa(0))
		fmtPrintf("  pushq %%rax\n")
	case T_ARRAY, T_STRUCT:
		// pure proxy
		fmtPrintf("  pushq %%rax\n")
	default:
		panic2(__func__, "TBI:kind="+string(kind(t)))
	}
}

func emitVariableAddr(variable *Variable) {
	emitComment(2, "emit Addr of variable \"%s\" \n", variable.name)

	if variable.isGlobal {
		fmtPrintf("  leaq %s(%%rip), %%rax # global variable \"%s\"\n", variable.globalSymbol, variable.name)
	} else {
		fmtPrintf("  leaq %d(%%rbp), %%rax # local variable \"%s\"\n", Itoa(int(variable.localOffset)), variable.name)
	}

	fmtPrintf("  pushq %%rax # variable address\n")
}

func emitListHeadAddr(list ast.Expr) {
	var t = getTypeOfExpr(list)
	switch kind(t) {
	case T_ARRAY:
		emitAddr(list) // array head
	case T_SLICE:
		emitExpr(list, nil)
		emitPopSlice()
		fmtPrintf("  pushq %%rax # slice.ptr\n")
	case T_STRING:
		emitExpr(list, nil)
		emitPopString()
		fmtPrintf("  pushq %%rax # string.ptr\n")
	default:
		panic2(__func__, "kind="+string(kind(getTypeOfExpr(list))))
	}
}

func isOsArgs(e *ast.SelectorExpr) bool {
	xIdent, isIdent := e.X.(*ast.Ident)
	return isIdent && xIdent.Name == "os" && e.Sel.Name == "Args"
}

func emitAddr(expr ast.Expr) {
	emitComment(2, "[emitAddr] %T\n", expr)
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Name == "_" {
			panic(" \"_\" has no address")
		}
		if e.Obj == nil {
			throw(expr)
		}
		if e.Obj.Kind == ast.Var {
			assert(e.Obj.Data != nil, "e.Obj.Data should not be nil: Obj.Name=" + e.Obj.Name)
			vr, ok := e.Obj.Data.(*Variable)
			if !ok {
				throw(e.Obj.Data)
			}
			assert(ok, "should be *Variable")
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
		if isOsArgs(e) {
			fmtPrintf("  leaq %s(%%rip), %%rax # hack for os.Args\n", "runtime.__args__")
			fmtPrintf("  pushq %%rax\n")
			return
		}
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
	case *ast.CompositeLit:
		knd := kind(getTypeOfExpr(e))
		switch knd {
		case T_STRUCT:
			// result of evaluation of a struct literal is its address
			emitExpr(expr, nil)
		default:
			panic(string(knd))
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
	emitComment(2, "Conversion %s <= %s\n", toType.e, getTypeOfExpr(arg0))
	switch to := toType.e.(type) {
	case *ast.Ident:
		ident := to
		switch ident.Obj {
		case gString: // string(e)
			switch kind(getTypeOfExpr(arg0)) {
			case T_SLICE: // string(slice)
				emitExpr(arg0, nil)
				emitPopSlice()
				fmtPrintf("  pushq %%rcx # str len\n")
				fmtPrintf("  pushq %%rax # str ptr\n")
			}
		case gInt, gUint8, gUint16, gUintptr: // int(e)
			emitExpr(arg0, nil)
		default:
			if ident.Obj.Kind == ast.Typ {
				// define type.  e.g. MyType(10)
				//typeSpec, ok := ident.Obj.Decl.(*ast.TypeSpec)
				//if !ok {
				//	throw(ident.Obj.Decl)
				//}
				// What should we do if MyType is an interface ?
				emitExpr(arg0, nil)
			} else {
				throw(ident.Obj)
			}
		}
	case *ast.ArrayType: // Conversion to slice
		arrayType := to
		if arrayType.Len != nil {
			throw(to)
		}
		assert(kind(getTypeOfExpr(arg0)) == T_STRING, "source type should be slice")
		emitComment(2, "Conversion to slice %s <= %s\n", arrayType.Elt, getTypeOfExpr(arg0))
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
		if isInterface(getTypeOfExpr(arg0))  {
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
		fmtPrintf("  pushq $0 # slice cap\n")
		fmtPrintf("  pushq $0 # slice len\n")
		fmtPrintf("  pushq $0 # slice ptr\n")
	case T_STRING:
		fmtPrintf("  pushq $0 # string len\n")
		fmtPrintf("  pushq $0 # string ptr\n")
	case T_INTERFACE:
		fmtPrintf("  pushq $0 # interface data\n")
		fmtPrintf("  pushq $0 # interface dtype\n")
	case T_INT, T_UINTPTR, T_UINT8, T_POINTER, T_BOOL:
		fmtPrintf("  pushq $0 # %s zero value\n", string(kind(t)))
	case T_STRUCT:
		structSize := getSizeOfType(t)
		fmtPrintf("  # zero value of a struct. size=%s (allocating on heap)\n", Itoa(structSize))
		emitCallMalloc(structSize)
	default:
		throw(t)
	}
}

func emitLen(arg ast.Expr) {
	switch kind(getTypeOfExpr(arg)) {
	case T_ARRAY:
		arrayType, ok := getTypeOfExpr(arg).e.(*ast.ArrayType)
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
		throw(kind(getTypeOfExpr(arg)))
	}
}

func emitCap(arg ast.Expr) {
	switch kind(getTypeOfExpr(arg)) {
	case T_ARRAY:
		arrayType, ok := getTypeOfExpr(arg).e.(*ast.ArrayType)
		assert(ok, "should be *ast.ArrayType")
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
	fmtPrintf("  pushq $%s\n", Itoa(size))
	// call malloc and return pointer
	var resultList = []*ast.Field{
		&ast.Field{
			Names:   nil,
			Type:    tUintptr.e,
		},
	}
	fmtPrintf("  callq runtime.malloc\n") // no need to invert args orders
	emitRevertStackPointer(intSize)
	emitReturnedValue(resultList)
}

func emitStructLiteral(e *ast.CompositeLit) {
	// allocate heap area with zero value
	fmtPrintf("  # Struct literal\n")
	structType := e2t(e.Type)
	emitZeroValue(structType) // push address of the new storage
	for i, elm := range e.Elts {
		kvExpr , ok := elm.(*ast.KeyValueExpr)
		assert(ok, "expect *ast.KeyValueExpr")
		fieldName, ok := kvExpr.Key.(*ast.Ident)
		assert(ok, "expect *ast.Ident")
		fmt.Printf("  #  - [%d] : key=%s, value=%T\n", i, fieldName.Name, kvExpr.Value)
		field := lookupStructField(getStructTypeSpec(structType), fieldName.Name)
		fieldType := e2t(field.Type)
		fieldOffset := getStructFieldOffset(field)
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

func emitArrayLiteral(arrayType *ast.ArrayType, arrayLen int, elts []ast.Expr) {
	elmType := e2t(arrayType.Elt)
	elmSize := getSizeOfType(elmType)
	memSize := elmSize * arrayLen
	emitCallMalloc(memSize) // push
	for i, elm := range elts {
		// push lhs address
		emitPushStackTop(tUintptr, "malloced address")
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
	fmtPrintf("  xor $1, %%rax\n")
	fmtPrintf("  pushq %%rax\n")
}

func emitTrue() {
	fmt.Printf("  pushq $1 # true\n")
}

func emitFalse() {
	fmt.Printf("  pushq $0 # false\n")
}

type Arg struct {
	e      ast.Expr
	t      *Type // expected type
	offset int
}

func emitArgs(args []*Arg) int {
	var totalPushedSize int
	for _, arg := range args {
		var t *Type
		if arg.t != nil {
			t = arg.t
		} else {
			t = getTypeOfExpr(arg.e)
		}
		arg.offset = totalPushedSize
		totalPushedSize += getPushSizeOfType(t)
	}
	fmtPrintf("  subq $%d, %%rsp # for args\n", Itoa(totalPushedSize))
	for _, arg := range args {
		ctx := &evalContext{
			_type: arg.t,
		}
		emitExprIfc(arg.e, ctx)
	}
	fmtPrintf("  addq $%d, %%rsp # for args\n", Itoa(totalPushedSize))

	for _, arg := range args {
		var t *Type
		if arg.t != nil {
			t = arg.t
		} else {
			t = getTypeOfExpr(arg.e)
		}
		switch kind(t) {
		case T_BOOL, T_INT, T_UINT8, T_POINTER, T_UINTPTR:
			fmtPrintf("  movq %d-8(%%rsp) , %%rax # load\n", Itoa(-arg.offset))
			fmtPrintf("  movq %%rax, %d(%%rsp) # store\n", Itoa(+arg.offset))
		case T_STRING, T_INTERFACE:
			fmtPrintf("  movq %d-16(%%rsp), %%rax\n", Itoa(-arg.offset))
			fmtPrintf("  movq %d-8(%%rsp), %%rcx\n", Itoa(-arg.offset))
			fmtPrintf("  movq %%rax, %d(%%rsp)\n", Itoa(+arg.offset))
			fmtPrintf("  movq %%rcx, %d+8(%%rsp)\n", Itoa(+arg.offset))
		case T_SLICE:
			fmtPrintf("  movq %d-24(%%rsp), %%rax\n", Itoa(-arg.offset)) // arg1: slc.ptr
			fmtPrintf("  movq %d-16(%%rsp), %%rcx\n", Itoa(-arg.offset)) // arg1: slc.len
			fmtPrintf("  movq %d-8(%%rsp), %%rdx\n", Itoa(-arg.offset))  // arg1: slc.cap
			fmtPrintf("  movq %%rax, %d+0(%%rsp)\n", Itoa(+arg.offset))  // arg1: slc.ptr
			fmtPrintf("  movq %%rcx, %d+8(%%rsp)\n", Itoa(+arg.offset))  // arg1: slc.len
			fmtPrintf("  movq %%rdx, %d+16(%%rsp)\n", Itoa(+arg.offset)) // arg1: slc.cap
		default:
			throw(kind(t))
		}
	}

	return totalPushedSize
}

func prepareArgs(funcType *ast.FuncType, receiver ast.Expr, eArgs []ast.Expr) []*Arg {
	if funcType == nil {
		panic("no funcType")
	}
	var args []*Arg
	params := funcType.Params.List
	var variadicArgs []ast.Expr // nil means there is no varadic in funcdecl
	var variadicElp *ast.Ellipsis
	var argIndex int
	var eArg ast.Expr
	var param *ast.Field
	for argIndex, eArg = range eArgs {
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
		// collect args as a slice
		sliceType := &ast.ArrayType{Elt: variadicElp.Elt}
		sliceLiteral := &ast.CompositeLit{
			Type:       sliceType,
			Lbrace:     0,
			Elts:       variadicArgs,
			Rbrace:     0,
			Incomplete: false,
		}
		args = append(args, &Arg{
			e:      sliceLiteral,
			t:      e2t(sliceType),
		})
	} else if len(args) < len(params) {
		// Add nil as a variadic arg
		param := params[argIndex+1]
		elp, ok := param.Type.(*ast.Ellipsis)
		assert(ok, "compile error")
		args = append(args, &Arg{
			e:      eNil,
			t:      e2t(elp),
		})
	}

	if receiver != nil { // method call
		var receiverAndArgs []*Arg = []*Arg{
			&Arg{
				e: receiver,
				t: getTypeOfExpr(receiver),
			},
		}
		for _, arg := range args {
			receiverAndArgs = append(receiverAndArgs, arg)
		}
		return receiverAndArgs
	}

	return args
}

func emitCall(symbol string, args []*Arg, results []*ast.Field) {
	totalPushedSize := emitArgs(args)
	fmtPrintf("  callq %s\n", symbol)
	emitRevertStackPointer(totalPushedSize)
	emitReturnedValue(results)
}

func emitReturnedValue(resultList []*ast.Field) {
	switch len(resultList) {
	case 0:
		// do nothing
	case 1:
		emitComment(2, "emit return value\n")
		retval0 := resultList[0]
		switch kind(e2t(retval0.Type)) {
		case T_STRING:
			fmt.Printf("  pushq %%rdi # str len\n")
			fmt.Printf("  pushq %%rax # str ptr\n")
		case T_INTERFACE:
			fmtPrintf("  pushq %%rdi # ifc data\n")
			fmtPrintf("  pushq %%rax # ifc dtype\n")
		case T_BOOL, T_INT, T_UINTPTR, T_POINTER:
			fmt.Printf("  pushq %%rax\n")
		case T_SLICE:
			fmt.Printf("  pushq %%rsi # slice cap\n")
			fmt.Printf("  pushq %%rdi # slice len\n")
			fmt.Printf("  pushq %%rax # slice ptr\n")
		default:
			throw(kind(e2t(retval0.Type)))
		}
	default:
		panic("multipul returned values is not supported ")
	}
}

func emitFuncall(fun ast.Expr, eArgs []ast.Expr) {
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
				arrayType, ok := typeArg.e.(*ast.ArrayType)
				assert(ok, "should be *ast.ArrayType")
				var elmSize = getSizeOfType(e2t(arrayType.Elt))
				var numlit = newNumberLiteral(elmSize)

				var args []*Arg = []*Arg{
					// elmSize
					&Arg{
						e: numlit,
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

				var resultList = []*ast.Field{
					&ast.Field{
						Names: nil,
						Type:  generalSlice,
					},
				}
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
			var resultList = []*ast.Field{
				&ast.Field{
					Names: nil,
					Type:  generalSlice,
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
				panic("TBI")
			}
			emitCall(symbol, _args, nil)
			return
		}

		if fn.Name == "makeSlice1" || fn.Name == "makeSlice8" || fn.Name == "makeSlice16" || fn.Name == "makeSlice24" {
			fn.Name = "makeSlice"
		}

		// general function call
		symbol = getFuncSymbol(pkg.name, fn.Name)
		obj := fn.Obj //.Kind == FN
		fndecl, ok := obj.Decl.(*ast.FuncDecl)
		if !ok {
			throw(fn.Obj)
		}
		funcType = fndecl.Type
	case *ast.SelectorExpr:
		symbol = fmt.Sprintf("%s.%s", fn.X, fn.Sel)
		switch symbol {
		case "unsafe.Pointer":
			// This is actually not a call
			emitExpr(eArgs[0], nil)
			return
		case "os.Exit":
			funcType = funcTypeOsExit
		case "syscall.Open":
			// func body is in runtime.s
			funcType = funcTypeSyscallOpen
		case "syscall.Read":
			// func body is in runtime.s
			funcType = funcTypeSyscallRead
		case "syscall.Write":
			// func body is in runtime.s
			funcType = funcTypeSyscallWrite
		case "syscall.Syscall":
			// func body is in runtime.s
			funcType = funcTypeSyscallSyscall
		default:
			// Assume method call
			receiver = fn.X
			receiverType := getTypeOfExpr(receiver)
			method := lookupMethod(receiverType, fn.Sel)
			funcType = method.funcType
			subsymbol := getMethodSymbol(method)
			symbol = getFuncSymbol(pkg.name, subsymbol)
		}
	default:
		throw(fun)
	}

	args := prepareArgs(funcType, receiver, eArgs)
	var resultList []*ast.Field
	if funcType.Results != nil {
		resultList = funcType.Results.List
	}
	emitCall(symbol, args, resultList)
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

func emitNil(targetType *Type) {
	if targetType == nil {
		panic("Type is required to emit nil")
	}
	switch kind(targetType) {
	case T_SLICE, T_POINTER, T_INTERFACE:
		emitZeroValue(targetType)
	default:
		throw(kind(targetType))
	}
}

func emitNamedConst(e *ast.Ident, ctx *evalContext) {
	valSpec, ok := e.Obj.Decl.(*ast.ValueSpec)
	assert(ok, "should be *ast.ValueSpec")
	lit, ok := valSpec.Values[0].(*ast.BasicLit)
	assert(ok, "should be *ast.BasicLit")
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
			emitNil(ctx._type)
			isNilObj = true
		default:
			if e.Obj == nil {
				panic(fmt.Sprintf("ident %s is unresolved", e.Name))
			}
			switch e.Obj.Kind {
			case ast.Var:
				emitAddr(e)
				emitLoadFromMemoryAndPush(getTypeOfExpr(e))
			case ast.Con:
				emitNamedConst(e, ctx)
			default:
				panic("Unexpected ident kind:" + e.Obj.Kind.String())
			}
		}
	case *ast.IndexExpr: // 1 or 2 values
		emitAddr(e)
		emitLoadFromMemoryAndPush(getTypeOfExpr(e))
	case *ast.StarExpr: // 1 value
		emitAddr(e)
		emitLoadFromMemoryAndPush(getTypeOfExpr(e))
	case *ast.SelectorExpr: // 1 value X.Sel
		emitComment(2, "emitExpr *ast.SelectorExpr %s.%s\n", e.X, e.Sel)
		emitAddr(e)
		emitLoadFromMemoryAndPush(getTypeOfExpr(e))
	case *ast.CallExpr: // multi values Fun(Args)
		var fun = e.Fun
		emitComment(2, "callExpr=%#v\n", fun)
		// check if it's a conversion
		if isType(fun) {
			emitConversion(e2t(fun), e.Args[0])
		} else {
			emitFuncall(fun, e.Args)
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
			fmtPrintf("  pushq $%d # convert char literal to int\n", Itoa(int(char)))
		case "INT":
			ival := Atoi(e.Value)
			fmtPrintf("  pushq $%d # number literal\n", Itoa(ival))
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
			fmtPrintf("  popq %%rax # e.X\n")
			fmtPrintf("  imulq $-1, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		case "&":
			emitAddr(e.X)
		case "!":
			emitExpr(e.X, nil)
			emitInvertBoolValue()
		default:
			throw(e.Op.String())
		}
	case *ast.BinaryExpr: // 1 value
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
				var resultList = []*ast.Field{
					&ast.Field{
						Names:   nil,
						Type:    tString.e,
					},
				}

				emitCall("runtime.catstrings", args, resultList)
			case "==":
				emitArgs(args)
				emitCompEq(getTypeOfExpr(e.X))
			case "!=":
				emitArgs(args)
				emitCompEq(getTypeOfExpr(e.X))
				emitInvertBoolValue()
			default:
				throw(e.Op.String())
			}
		} else {
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
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				emitExpr(e.Y, nil)   // right
				fmtPrintf("  popq %%rcx # right\n")
				fmtPrintf("  popq %%rax # left\n")
				fmtPrintf("  addq %%rcx, %%rax\n")
				fmtPrintf("  pushq %%rax\n")
			case "-":
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				emitExpr(e.Y, nil)   // right
				fmtPrintf("  popq %%rcx # right\n")
				fmtPrintf("  popq %%rax # left\n")
				fmtPrintf("  subq %%rcx, %%rax\n")
				fmtPrintf("  pushq %%rax\n")
			case "*":
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				emitExpr(e.Y, nil)   // right
				fmtPrintf("  popq %%rcx # right\n")
				fmtPrintf("  popq %%rax # left\n")
				fmtPrintf("  imulq %%rcx, %%rax\n")
				fmtPrintf("  pushq %%rax\n")
			case "%":
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				emitExpr(e.Y, nil)   // right
				fmtPrintf("  popq %%rcx # right\n")
				fmtPrintf("  popq %%rax # left\n")
				fmtPrintf("  movq $0, %%rdx # init %%rdx\n")
				fmtPrintf("  divq %%rcx\n")
				fmtPrintf("  movq %%rdx, %%rax\n")
				fmtPrintf("  pushq %%rax\n")
			case "/":
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				emitExpr(e.Y, nil)   // right
				fmtPrintf("  popq %%rcx # right\n")
				fmtPrintf("  popq %%rax # left\n")
				fmtPrintf("  movq $0, %%rdx # init %%rdx\n")
				fmtPrintf("  divq %%rcx\n")
				fmtPrintf("  pushq %%rax\n")
			case "==":
				var t = getTypeOfExpr(e.X)
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				ctx := &evalContext{_type: t}
				emitExprIfc(e.Y, ctx) // right
				emitCompEq(t)
			case "!=":
				var t = getTypeOfExpr(e.X)
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				ctx := &evalContext{_type: t}
				emitExprIfc(e.Y, ctx) // right
				emitCompEq(t)
				emitInvertBoolValue()
			case "<":
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				emitExpr(e.Y, nil)   // right
				emitCompExpr("setl")
			case "<=":
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				emitExpr(e.Y, nil)   // right
				emitCompExpr("setle")
			case ">":
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				emitExpr(e.Y, nil)   // right
				emitCompExpr("setg")
			case ">=":
				emitComment(2, "start %T\n", e)
				emitExpr(e.X, nil) // left
				emitExpr(e.Y, nil)   // right
				emitCompExpr("setge")
			default:
				panic(fmt.Sprintf("TBI: binary operation for '%s'", e.Op.String()))
			}
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
			panic(string(kind(e2t(e.Type))))
		}
	case *ast.SliceExpr: // 1 value list[low:high]
		list := e.X
		listType := getTypeOfExpr(list)
		emitExpr(e.High, nil) // intval
		emitExpr(e.Low, nil)  // intval
		fmtPrintf("  popq %%rcx # low\n")
		fmtPrintf("  popq %%rax # high\n")
		fmtPrintf("  subq %%rcx, %%rax # high - low\n")
		switch kind(listType) {
		case T_SLICE, T_ARRAY:
			fmtPrintf("  pushq %%rax # cap\n")
			fmtPrintf("  pushq %%rax # len\n")
		case T_STRING:
			fmtPrintf("  pushq %%rax # len\n")
			// no cap
		default:
			throw(list)
		}

		emitExpr(e.Low, nil) // index number
		elmType := getElementTypeOfListType(listType)
		emitListElementAddr(list, elmType)
	case *ast.TypeAssertExpr:
		emitExpr(e.X, nil)
		fmtPrintf("  popq  %%rax # ifc.dtype\n")
		fmtPrintf("  popq  %%rcx # ifc.data\n")
		fmtPrintf("  pushq %%rax # ifc.data\n")

		typ := e2t(e.Type)
		sType := serializeType(typ)
		typeId := getTypeId(sType)
		typeSymbol := typeIdToSymbol(typeId)
		// check if type matches
		fmtPrintf("  leaq %s(%%rip), %%rax # ifc.dtype\n", typeSymbol)
		fmtPrintf("  pushq %%rax           # ifc.dtype\n")

		emitCompExpr("sete") // this pushes 1 or 0 in the end
		emitPopBool("type assertion ok value")
		fmt.Printf("  cmpq $1, %%rax\n")

		labelid++
		labelTypeAssertionEnd := fmt.Sprintf(".L.end_type_assertion.%d", labelid)
		labelElse := fmt.Sprintf(".L.unmatch.%d", labelid)
		fmtPrintf("  jne %s # jmp if false\n", labelElse)

		// if matched
		if ctx.okContext != nil {
			emitComment(2, " double value context\n")
			if ctx.okContext.needMain {
				emitExpr(e.X, nil)
				fmtPrintf("  popq %%rax # garbage\n")
				emitLoadFromMemoryAndPush(e2t(e.Type)) // load dynamic data
			}
			if ctx.okContext.needOk {
				fmtPrintf("  pushq $1 # ok = true\n")
			}
		} else {
			emitComment(2, " single value context\n")
			emitExpr(e.X, nil)
			fmtPrintf("  popq %%rax # garbage\n")
			emitLoadFromMemoryAndPush(e2t(e.Type)) // load dynamic data
		}

		// exit
		fmtPrintf("  jmp %s\n", labelTypeAssertionEnd)

		// if not matched
		fmtPrintf("  %s:\n", labelElse)
		if ctx.okContext != nil {
			emitComment(2, " double value context\n")
			if ctx.okContext.needMain {
				emitZeroValue(typ)
			}
			if ctx.okContext.needOk {
				fmtPrintf("  pushq $0 # ok = false\n")
			}
		} else {
			emitComment(2, " single value context\n")
			emitZeroValue(typ)
		}

		fmtPrintf("  %s:\n", labelTypeAssertionEnd)
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
	return "dtype." + Itoa(id)
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
	fmtPrintf("  leaq %s(%%rip), %%rax # type symbol \"%s\"\n", typeSymbol, str)
	fmtPrintf("  pushq %%rax           # type symbol %s\n", typeSymbol)
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
	fmtPrintf("  popq %%rcx # index id\n")
	fmtPrintf("  movq $%s, %%rdx # elm size\n", Itoa(getSizeOfType(elmType)))
	fmtPrintf("  imulq %%rdx, %%rcx\n")
	fmtPrintf("  addq %%rcx, %%rax\n")
	fmtPrintf("  pushq %%rax # addr of element\n")
}

func emitCompEq(t *Type) {
	switch kind(t) {
	case T_STRING:
		var resultList = []*ast.Field{
			&ast.Field{
				Names:   nil,
				Type:    tBool.e,
			},
		}
		fmtPrintf("  callq runtime.cmpstrings\n")
		emitRevertStackPointer(stringSize * 2)
		emitReturnedValue(resultList)
	case T_INTERFACE:
		var resultList = []*ast.Field{
			&ast.Field{
				Names:   nil,
				Type:    tBool.e,
			},
		}
		fmtPrintf("  callq runtime.cmpinterface\n")
		emitRevertStackPointer(interfaceSize * 2)
		emitReturnedValue(resultList)
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
	fmtPrintf("  popq %%rcx # right\n")
	fmtPrintf("  popq %%rax # left\n")
	fmtPrintf("  cmpq %%rcx, %%rax\n")
	fmtPrintf("  %s %%al\n", inst)
	fmtPrintf("  movzbq %%al, %%rax\n") // true:1, false:0
	fmtPrintf("  pushq %%rax\n")
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
		panic("TBI:" + knd)
	}
}

func emitStore(t *Type, rhsTop bool, pushLhs bool) {
	knd := kind(t)
	emitComment(2, "emitStore(%s)\n", knd)
	if rhsTop {
		emitPop(knd) // rhs
		fmtPrintf("  popq %%rsi # lhs addr\n")
	} else {
		fmtPrintf("  popq %%rsi # lhs addr\n")
		emitPop(knd) // rhs
	}
	if pushLhs {
		fmtPrintf("  pushq %%rsi # lhs addr\n")
	}

	switch knd {
	case T_SLICE:
		fmtPrintf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", Itoa(0))
		fmtPrintf("  movq %%rcx, %d(%%rsi) # len to len\n", Itoa(8))
		fmtPrintf("  movq %%rdx, %d(%%rsi) # cap to cap\n", Itoa(16))
	case T_STRING:
		fmtPrintf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", Itoa(0))
		fmtPrintf("  movq %%rcx, %d(%%rsi) # len to len\n", Itoa(8))
	case T_INTERFACE:
		fmtPrintf("  movq %%rax, %d(%%rsi) # store dtype\n", Itoa(0))
		fmtPrintf("  movq %%rcx, %d(%%rsi) # store data\n", Itoa(8))
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		fmtPrintf("  movq %%rax, (%%rsi) # assign\n")
	case T_UINT16:
		fmtPrintf("  movw %%ax, (%%rsi) # assign word\n")
	case T_UINT8:
		fmtPrintf("  movb %%al, (%%rsi) # assign byte\n")
	case T_STRUCT, T_ARRAY:
		fmtPrintf("  pushq $%d # size\n", Itoa(getSizeOfType(t)))
		fmtPrintf("  pushq %%rsi # dst lhs\n")
		fmtPrintf("  pushq %%rax # src rhs\n")
		fmtPrintf("  callq runtime.memcopy\n")
		emitRevertStackPointer(ptrSize*2 + intSize)
	default:
		panic("TBI:" + knd)
	}
}

func isBlankIdentifier(e ast.Expr) bool {
	ident, isIdent := e.(*ast.Ident)
	if !isIdent {
		return false
	}
	return ident.Name == "_"
}

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
		_type:     nil,
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
	fmtPrintf("\n")
	emitComment(2, "== Statement %T ==\n", stmt)
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
				var valSpec = ds
				var t = e2t(valSpec.Type)
				emitComment(2, "Decl.Specs[0]: Names[0]=%#v, Type=%#v\n", ds.Names[0], t.e)
				lhs := valSpec.Names[0]
				var rhs ast.Expr
				if len(valSpec.Values) == 0 {
					emitAddr(lhs)
					emitComment(2, "emitZeroValue\n")
					emitZeroValue(t)
					emitComment(2, "Assignment: zero value\n")
					emitStore(t, true, false)
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
		case "=", ":=":
			lhs := s.Lhs[0]
			rhs := s.Rhs[0]
			_ , isTypeAssertion := rhs.(*ast.TypeAssertExpr)
			if len(s.Lhs) == 2 && isTypeAssertion {
				emitAssignWithOK(s.Lhs, rhs)
			} else {
				ident , isIdent := lhs.(*ast.Ident)
				if isIdent && ident.Name == "_" {
					panic(" _ is not supported yet")
				}
				emitAssign(lhs, rhs)
			}
		default:
			panic("TBI: assignment of " + s.Tok.String())
		}
	case *ast.ReturnStmt:
		node := mapReturnStmt[s]
		funcType := node.fnc.funcType
		if len(s.Results) == 0 {
			fmtPrintf("  leave\n")
			fmtPrintf("  ret\n")
		} else if len(s.Results) == 1 {
			//funcType := nil
			targetType := e2t(funcType.Results.List[0].Type)
			ctx := &evalContext{
				_type:     targetType,
			}
			emitExprIfc(s.Results[0], ctx)
			var knd = kind(targetType)
			switch knd {
			case T_BOOL, T_INT, T_UINTPTR, T_POINTER:
				fmtPrintf("  popq %%rax # return 64bit\n")
			case T_STRING, T_INTERFACE:
				fmtPrintf("  popq %%rax # return string (head)\n")
				fmtPrintf("  popq %%rdi # return string (tail)\n")
			case T_SLICE:
				fmtPrintf("  popq %%rax # return string (head)\n")
				fmtPrintf("  popq %%rdi # return string (body)\n")
				fmtPrintf("  popq %%rsi # return string (tail)\n")
			default:
				panic("TBI:" + knd)
			}

			fmtPrintf("  leave\n")
			fmtPrintf("  ret\n")
		} else if len(s.Results) == 3 {
			// Special treatment to return a slice by makeSlice()
			for _, result := range s.Results {
				assert(getSizeOfType(getTypeOfExpr(result)) == 8, "TBI")
			}
			emitExpr(s.Results[2], nil) // @FIXME
			emitExpr(s.Results[1], nil) // @FIXME
			emitExpr(s.Results[0], nil) // @FIXME
			fmt.Printf("  popq %%rax # return 64bit\n")
			fmt.Printf("  popq %%rdi # return 64bit\n")
			fmt.Printf("  popq %%rsi # return 64bit\n")
		} else {
			panic("TBI")
		}
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
		emitLoadFromMemoryAndPush(tInt)
		emitVariableAddr(rngMisc.lenvar)
		emitLoadFromMemoryAndPush(tInt)
		emitCompExpr("setl")
		emitPopBool(" indexvar < lenvar")
		fmt.Printf("  cmpq $1, %%rax\n")
		fmt.Printf("  jne %s # jmp if false\n", labelExit)

		emitComment(2, "assign list[indexvar] value variables\n")
		elemType := getTypeOfExpr(s.Value)
		emitAddr(s.Value) // lhs

		emitVariableAddr(rngMisc.indexvar)
		emitLoadFromMemoryAndPush(tInt) // index value
		emitListElementAddr(s.X, elemType)

		emitLoadFromMemoryAndPush(elemType)
		emitStore(elemType, true, false)

		// Body
		emitComment(2, "ForRange Body\n")
		emitStmt(s.Body)

		// Post statement: Increment indexvar and go next
		emitComment(2, "ForRange Post statement\n")
		fmt.Printf("  %s:\n", labelPost)   // used for "continue"
		emitVariableAddr(rngMisc.indexvar) // lhs
		emitVariableAddr(rngMisc.indexvar) // rhs
		emitLoadFromMemoryAndPush(tInt)
		emitAddConst(1, "indexvar value ++")
		emitStore(tInt, true, false)

		// incr key variable
		if s.Key != nil {
			keyIdent, ok := s.Key.(*ast.Ident)
			assert(ok, "key expr should be an ident")
			if keyIdent.Name != "_" {
				emitAddr(s.Key)                    // lhs
				emitVariableAddr(rngMisc.indexvar) // rhs
				emitLoadFromMemoryAndPush(tInt)
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
				emitPushStackTop(condType, "switch expr")
				emitExpr(e, nil)
				emitCompEq(condType)

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
		emitVariableAddr(typeSwitch.subjectVariable)
		emitExpr(typeSwitch.subject, nil)
		emitStore(tEface, true, false)

		cases := s.Body.List
		var labels = make([]string, len(cases))
		var defaultLabel string
		emitComment(2, "Start comparison with cases\n")
		for i, c := range cases {
			cc, ok := c.(*ast.CaseClause)
			assert(ok, "should be *ast.CaseClause")
			labelid++
			labelCase := ".L.case." + Itoa(labelid)
			labels[i] = labelCase
			if len(cc.List) == 0 {
				defaultLabel = labelCase
				continue
			}
			for _, e := range cc.List {
				emitVariableAddr(typeSwitch.subjectVariable)
				emitLoadFromMemoryAndPush(tEface)

				emitDtypeSymbol(e2t(e))
				emitCompExpr("sete") // this pushes 1 or 0 in the end

				emitPopBool(" of switch-case comparison")
				fmtPrintf("  cmpq $1, %%rax\n")
				fmtPrintf("  je %s # jump if match\n", labelCase)
			}
		}
		emitComment(2, "End comparison with cases\n")

		// if no case matches, then jump to
		if defaultLabel != "" {
			// default
			fmtPrintf("  jmp %s\n", defaultLabel)
		} else {
			// exit
			fmtPrintf("  jmp %s\n", labelEnd)
		}

		for i, typeSwitchCaseClose := range typeSwitch.cases {
			// Injecting variable and type to the subject
			if typeSwitchCaseClose.variable != nil {
				typeSwitch.assignIdent.Obj.Data = typeSwitchCaseClose.variable
			}
			fmtPrintf("%s:\n", labels[i])

			for _, _s := range typeSwitchCaseClose.orig.Body {
				if typeSwitchCaseClose.variable != nil {
					// do assignment
					emitAddr(typeSwitch.assignIdent)

					emitVariableAddr(typeSwitch.subjectVariable)
					emitLoadFromMemoryAndPush(tEface)
					fmtPrintf("  popq %%rax # ifc.dtype\n")
					fmtPrintf("  popq %%rcx # ifc.data\n")
					fmtPrintf("  push %%rcx # ifc.data\n")
					emitLoadFromMemoryAndPush(typeSwitchCaseClose.variableType)

					emitStore(typeSwitchCaseClose.variableType, true, false)
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
			fmtPrintf("jmp %s # continue\n", containerFor.labelPost)
		case token.BREAK:
			fmtPrintf("jmp %s # break\n", containerFor.labelExit)
		default:
			throw(s.Tok)
		}
	default:
		throw(stmt)
	}
}

func emitRevertStackTop(t *Type) {
	fmtPrintf("  addq $%s, %%rsp # revert stack top\n", Itoa(getSizeOfType(t)))
}

var labelid int

func getMethodSymbol(method *Method) string {
	rcvTypeName := method.rcvNamedType
	if method.isPtrMethod {
		return "$" + rcvTypeName.Name + "." + method.name // pointer
	} else {
		return rcvTypeName.Name + "." + method.name // value
	}
}

func getFuncSubSymbol(fnc *Func) string {
	var subsymbol string
	if fnc.method != nil {
		subsymbol = getMethodSymbol(fnc.method)
	} else {
		subsymbol = fnc.name
	}
	return subsymbol
}

func getFuncSymbol(pkgPrefix string, subsymbol string) string {
	return pkgPrefix + "." + subsymbol
}

func emitFuncDecl(pkgPrefix string, fnc *Func) {
	var localarea int = int(fnc.localarea)
	fmtPrintf("\n")
	subsymbol := getFuncSubSymbol(fnc)
	symbol := getFuncSymbol(pkgPrefix, subsymbol)
	fmtPrintf("%s: # args %d, locals %d\n",
		symbol, Itoa(int(fnc.argsarea)), Itoa(int(fnc.localarea)))

	fmtPrintf("  pushq %%rbp\n")
	fmtPrintf("  movq %%rsp, %%rbp\n")
	if localarea != 0 {
		fmtPrintf("  subq $%d, %%rsp # local area\n", Itoa(-localarea))
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
		fmt.Printf("# init global %s: # T %s\n", name.Name, typeKind)
		emitAssign(name, val)
	}
}

func emitGlobalVariable(pkg *PkgContainer, name *ast.Ident, t *Type, val ast.Expr) {
	typeKind := kind(t)
	fmtPrintf("%s.%s: # T %s\n", pkg.name, name.Name, string(typeKind))
	switch typeKind {
	case T_STRING:
		switch vl := val.(type) {
		case nil:
			fmtPrintf("  .quad 0\n")
			fmtPrintf("  .quad 0\n")
		case *ast.BasicLit:
			sl := getStringLiteral(vl)
			fmtPrintf("  .quad %s\n", sl.label)
			fmtPrintf("  .quad %d\n", Itoa(sl.strlen))
		default:
			panic("Unsupported global string value")
		}
	case T_INTERFACE:
		// only zero value
		fmtPrintf("  .quad 0 # dtype\n")
		fmtPrintf("  .quad 0 # data\n")
	case T_BOOL:
		switch vl := val.(type) {
		case nil:
			fmtPrintf("  .quad 0 # bool zero value\n")
		case *ast.Ident:
			switch vl.Obj {
			case gTrue:
				fmtPrintf("  .quad 1 # bool true\n")
			case gFalse:
				fmtPrintf("  .quad 0 # bool false\n")
			default:
				throw(val)
			}
		default:
			throw(val)
		}
	case T_INT:
		switch vl := val.(type) {
		case nil:
			fmtPrintf("  .quad 0\n")
		case *ast.BasicLit:
			fmtPrintf("  .quad %s\n", vl.Value)
		default:
			throw(val)
		}
	case T_UINT8:
		switch vl := val.(type) {
		case nil:
			fmtPrintf("  .byte 0\n")
		case *ast.BasicLit:
			fmtPrintf("  .byte %s\n", vl.Value)
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
		fmtPrintf("  .quad 0\n")
	case T_UINTPTR:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		fmtPrintf("  .quad 0\n")
	case T_SLICE:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		fmtPrintf("  .quad 0 # ptr\n")
		fmtPrintf("  .quad 0 # len\n")
		fmtPrintf("  .quad 0 # cap\n")
	case T_ARRAY:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		arrayType, ok := t.e.(*ast.ArrayType)
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
			throw(arrayType.Elt)
		}
		for i := 0; i < length; i++ {
			fmt.Printf(zeroValue)
		}
	default:
		throw(typeKind)
	}
}

func generateCode(pkg *PkgContainer) {
	fmt.Printf(".data\n")
	for _, con := range stringLiterals {
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
	fmtPrintf("\n")
	fmtPrintf(".text\n")
	fmtPrintf("%s.__initGlobals:\n", pkg.name)
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

	fmtPrintf("\n")
}

// --- type ---
const sliceSize int = 24
const stringSize int = 16
const intSize int = 8
const ptrSize int = 8
const interfaceSize int = 16

type Type struct {
	e ast.Expr // original expr
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


// Rune
var tInt32 *Type = &Type{
	e: &ast.Ident{
		NamePos: 0,
		Name:    "int",
		Obj:     gInt32,
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
		Name:    "uint8",
		Obj:     gUint8,
	},
}

var tSliceOfString *Type = &Type{
	e: &ast.ArrayType{
		Len: nil,
		Elt: &ast.Ident{
			NamePos: 0,
			Name:    "string",
			Obj:     gString,
		},
	},
}

var tString *Type = &Type{
	e: &ast.Ident{
		NamePos: 0,
		Name:    "string",
		Obj:     gString,
	},
}

var tEface *Type = &Type{
	e: &ast.InterfaceType{},
}

var generalSlice ast.Expr = &ast.Ident{}

func getTypeOfExpr(expr ast.Expr) *Type {
	switch e := expr.(type) {
	case *ast.Ident:
		assert(e.Obj != nil, "Obj is nil in ident '" + e.Name + "'")
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
				return variable.typ
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
			starExpr.X = t.e
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
				case gNew:
					return e2t(&ast.StarExpr{
						Star: 0,
						X:    e.Args[0],
					})
				case gMake:
					return e2t(e.Args[0])
				case gAppend:
					return e2t(e.Args[0])
				}
				switch decl := fn.Obj.Decl.(type) {
				case *ast.FuncDecl:
					assert(len(decl.Type.Results.List) == 1, "func is expected to return a single value")
					return e2t(decl.Type.Results.List[0].Type)
				default:
					throw(fn.Obj)
				}
			}
		case *ast.ParenExpr: // (X)(e) funcall or conversion
			if isType(fn.X) {
				return e2t(fn.X)
			} else {
				panic("TBI: what should we do ?")
			}
		case *ast.ArrayType: // conversion [n]T(e) or []T(e)
			return e2t(fn)
		case *ast.SelectorExpr: // (X).Sel()
			xIdent, ok := fn.X.(*ast.Ident)
			if !ok {
				throw(fn)
			}
			var funcType *ast.FuncType
			symbol := fmt.Sprintf("%s.%s", xIdent.Name, fn.Sel)
			switch symbol {
			case "unsafe.Pointer":
				// unsafe.Pointer(x)
				return tUintptr
			case "os.Exit":
				funcType = funcTypeOsExit
				return nil
			case "syscall.Open":
				// func body is in runtime.s
				funcType = funcTypeSyscallOpen
				return	e2t(funcType.Results.List[0].Type)
			case "syscall.Read":
				// func body is in runtime.s
				funcType = funcTypeSyscallRead
				return	e2t(funcType.Results.List[0].Type)
			case "syscall.Write":
				// func body is in runtime.s
				funcType = funcTypeSyscallWrite
				return	e2t(funcType.Results.List[0].Type)
			case "syscall.Syscall":
				// func body is in runtime.s
				funcType = funcTypeSyscallSyscall
				return	e2t(funcType.Results.List[0].Type)
			default:
				// Assume method call
				xType := getTypeOfExpr(fn.X)
				method := lookupMethod(xType, fn.Sel)
				assert(len(method.funcType.Results.List) == 1, "func is expected to return a single value")
				return e2t(method.funcType.Results.List[0].Type)
			}
		case *ast.InterfaceType:
			return tEface
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
		// X.Sel
		emitComment(2, "getTypeOfExpr(%s.%s)\n", e.X, e.Sel)
		if isOsArgs(e) {
			// os.Args
			return tSliceOfString
		}
		structType := getStructTypeOfX(e)
		field := lookupStructField(getStructTypeSpec(structType), e.Sel.Name)
		return e2t(field.Type)
	case *ast.CompositeLit:
		return e2t(e.Type)
	case *ast.ParenExpr:
		return getTypeOfExpr(e.X)
	case *ast.TypeAssertExpr:
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

func serializeType(t *Type) string {
	if t == nil {
		panic("nil type is not expected")
	}
	if t.e == generalSlice {
		panic("TBD: generalSlice")
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
				return "main." + typeSpec.Name.Name
			}
		}
	case *ast.StructType:
		return "struct"
	case *ast.ArrayType:
		if e.Len == nil {
			return "[]" + serializeType(e2t(e.Elt))
		} else {
			return "[" + Itoa(evalInt(e.Len)) + "]" + serializeType(e2t(e.Elt))
		}
	case *ast.StarExpr:
		return "*" + serializeType(e2t(e.X))
	case *ast.Ellipsis: // x ...T
		panic("TBD: Ellipsis")
	case *ast.InterfaceType:
		return "interface"
	default:
		throw(t)
	}
	return ""
}

func kind(t *Type) TypeKind {
	if t == nil {
		panic("nil type is not expected")
	}
	if t.e == generalSlice {
		return T_SLICE
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
		ptrType, ok := typeOfX.e.(*ast.StarExpr)
		assert(ok, "should be *ast.StarExpr")
		structType = e2t(ptrType.X)
	default:
		throw("TBI")
	}
	return structType
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
	case T_INTERFACE:
		return interfaceSize
	default:
		throw(t)
	}
	return varSize
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
		throw(t)
	}
	throw(t)
	return 0
}

func getStructFieldOffset(field *ast.Field) int {
	if field.Doc == nil {
		panic("Doc is nil:" + field.Names[0].Name)
	}
	text := field.Doc.List[0].Text
	offset := Atoi(text)
	return offset
}

func setStructFieldOffset(field *ast.Field, offset int) {
	comment := &ast.Comment{
		Text: Itoa(offset),
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
	subject         ast.Expr
	subjectVariable *Variable
	assignIdent     *ast.Ident
	cases           []*TypeSwitchCaseClose
}

type TypeSwitchCaseClose struct {
	variable     *Variable
	variableType *Type
	orig         *ast.CaseClause
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
	funcType  *ast.FuncType
	method *Method
}

type Method struct {
	rcvNamedType *ast.Ident
	isPtrMethod  bool
	name string
	funcType *ast.FuncType
}

type Variable struct {
	name         string
	isGlobal     bool
	globalSymbol string
	localOffset  localoffsetint
	typ *Type
}

type localoffsetint int

func (fnc *Func) registerParamVariable(name string, t *Type) *Variable {
	vr := newLocalVariable(name, fnc.argsarea,t)
	fnc.argsarea += localoffsetint(getSizeOfType(t))
	return vr
}

func (fnc *Func) registerLocalVariable(name string, t *Type) *Variable {
	assert(t != nil && t.e != nil, "type of local var should not be nil")
	fnc.localarea -= localoffsetint(getSizeOfType(t))
	return newLocalVariable(name, currentFunc.localarea, t)
}

var stringLiterals []*stringLiteralsContainer
var stringIndex int
var currentFor *ForStmt

var mapForNodeToFor map[*ast.ForStmt]*ForStmt = map[*ast.ForStmt]*ForStmt{}
var mapRangeNodeToFor map[*ast.RangeStmt]*ForStmt = map[*ast.RangeStmt]*ForStmt{}
var mapBranchToFor map[*ast.BranchStmt]*ForStmt = map[*ast.BranchStmt]*ForStmt{}
var mapRangeStmt map[*ast.RangeStmt]*RangeStmtMisc = map[*ast.RangeStmt]*RangeStmtMisc{}
var mapTypeSwitchStmtMeta = map[*ast.TypeSwitchStmt]*TypeSwitchStmt{}
var mapReturnStmt = map[*ast.ReturnStmt]*nodeReturnStmt{}

var currentFunc *Func
func getStringLiteral(lit *ast.BasicLit) *sliteral {
	for _, container := range stringLiterals {
		if container.lit == lit {
			return container.sl
		}
	}

	panic(lit.Value)
	return nil
}

func registerStringLiteral(lit *ast.BasicLit) {
	if pkg.name == "" {
		panic("no pkgName")
	}

	var strlen int
	for _, c := range []uint8(lit.Value) {
		if c != '\\' {
			strlen++
		}
	}

	label := fmt.Sprintf(".%s.S%d", pkg.name, stringIndex)
	stringIndex++

	sl := &sliteral{
		label:  label,
		strlen: strlen - 2,
		value:  lit.Value,
	}
	var cont *stringLiteralsContainer = new(stringLiteralsContainer)
	cont.sl = sl
	cont.lit = lit
	stringLiterals = append(stringLiterals, cont)
}

func newGlobalVariable(pkgName string, name string, t *Type) *Variable {
	return &Variable{
		name:         name,
		isGlobal:     true,
		globalSymbol: pkgName + "." +name,
		localOffset:  0,
		typ: t,
	}
}

func newLocalVariable(name string, localoffset localoffsetint, t *Type) *Variable {
	return &Variable{
		name:         name,
		isGlobal:     false,
		globalSymbol: "",
		localOffset:  localoffset,
		typ: t,
	}
}

// https://golang.org/ref/spec#Method_sets
var typesWithMethods map[string]map[string]*Method = map[string]map[string]*Method{} // map[TypeName][MethodName]*Func

func newMethod(funcDecl *ast.FuncDecl) *Method {
	rcvType := funcDecl.Recv.List[0].Type
	rcvPointerType , ok := rcvType.(*ast.StarExpr)
	var isPtr bool
	if ok {
		isPtr = true
		rcvType = rcvPointerType.X
	}
	rcvNamedType,ok := rcvType.(*ast.Ident)
	if !ok {
		throw(rcvType)
	}
	method := &Method{
		rcvNamedType: rcvNamedType,
		isPtrMethod : isPtr,
		name: funcDecl.Name.Name,
		funcType: funcDecl.Type,
	}
	return method
}

func registerMethod(method *Method) {
	methodSet, ok := typesWithMethods[method.rcvNamedType.Name]
	if !ok {
		methodSet = map[string]*Method{}
		typesWithMethods[method.rcvNamedType.Name] = methodSet
	}
	methodSet[method.name] = method
}

func lookupMethod(rcvT *Type, methodName *ast.Ident) *Method {
	rcvType := rcvT.e
	rcvPointerType , ok := rcvType.(*ast.StarExpr)
	if ok {
		rcvType = rcvPointerType.X
	}
	rcvTypeName, ok := rcvType.(*ast.Ident)
	if !ok {
		throw(rcvType)
	}
	methodSet,ok := typesWithMethods[rcvTypeName.Name]
	if !ok {
		panic(rcvTypeName.Name + " has no moethodeiverTypeName:")
	}
	method, ok := methodSet[methodName.Name]
	if !ok {
		panic("method not found")
	}
	return method
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
				if varSpec.Type == nil { // var x = e
					if len(ds.Values) > 0 {
						// infer type from rhs
						val := ds.Values[0]
						logf("nfering type of variable %s\n", obj.Name)
						typ := getTypeOfExpr(val)
						if typ != nil && typ.e != nil {
							varSpec.Type = typ.e
						}
					}
				}

				t := e2t(varSpec.Type)
				obj.Data = currentFunc.registerLocalVariable(obj.Name, t)
				for _, v := range ds.Values {
					walkExpr(v)
				}
			}
		default:
			throw(decl)
		}

	case *ast.AssignStmt:
		lhs := s.Lhs[0]
		rhs := s.Rhs[0]
		if s.Tok.String() == ":=" {
			// short var decl
			ident, ok := lhs.(*ast.Ident)
			assert(ok, "should be ident")
			obj := ident.Obj
			assert(obj.Kind == ast.Var, "should be ast.Var")
			walkExpr(rhs)
			// infer type
			typ := getTypeOfExpr(rhs)
			obj.Data = currentFunc.registerLocalVariable(obj.Name, typ)

		} else {
			walkExpr(rhs)
		}
	case *ast.ReturnStmt:
		mapReturnStmt[s] = &nodeReturnStmt{
			fnc:currentFunc,
		}
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
	case *ast.RangeStmt:
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
			keyIdent.Obj.Data = currentFunc.registerLocalVariable(keyIdent.Name, keyType)

			// determine type of Value
			elmType := getElementTypeOfListType(listType)
			valueIdent := s.Value.(*ast.Ident)
			valueIdent.Obj.Data = currentFunc.registerLocalVariable(valueIdent.Name, elmType)
		}
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
	case *ast.TypeSwitchStmt:
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
			typeSwitch.subject = typeAssertExpr.X
			walkExpr(typeAssertExpr.X)
		case *ast.AssignStmt:
			lhs := assign.Lhs[0]
			var ok bool
			assignIdent, ok = lhs.(*ast.Ident)
			assert(ok, "lhs should be ident")
			typeSwitch.assignIdent = assignIdent
			// ident will be a new local variable in each case clause
			typeAssertExpr, ok := assign.Rhs[0].(*ast.TypeAssertExpr)
			assert(ok, "should be *ast.TypeAssertExpr")
			typeSwitch.subject = typeAssertExpr.X
			walkExpr(typeAssertExpr.X)
		default:
			throw(s.Assign)
		}

		typeSwitch.subjectVariable = currentFunc.registerLocalVariable(".switch_expr", tEface)
		for _, _case := range s.Body.List {
			cc := _case.(*ast.CaseClause)
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
				assignIdent.Obj.Data = vr
			}

			for _, stmt := range cc.Body {
				walkStmt(stmt)
			}

			if assignIdent != nil {
				assignIdent.Obj.Data = nil
			}
		}
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

func walkExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.Ident:
		// what to do ?
	case *ast.SelectorExpr:
		walkExpr(e.X)
	case *ast.CallExpr:
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
	case *ast.ArrayType: // first argument of builtin func like make()
		// do nothing
	case *ast.StarExpr:
		walkExpr(e.X)
	case *ast.KeyValueExpr:
		walkExpr(e.Key)
		walkExpr(e.Value)
	case *ast.InterfaceType:
		// interface{}(e)  conversion. Nothing to do.
	case *ast.TypeAssertExpr:
		walkExpr(e.X)
	default:
		throw(expr)
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
func walk(pkg *PkgContainer, f *ast.File) {
	var typeSpecs []*ast.TypeSpec
	var funcDecls []*ast.FuncDecl
	var varSpecs []*ast.ValueSpec
	var constSpecs []*ast.ValueSpec

	// grouping declarations by type
	for _, decl := range f.Decls {
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
		switch kind(e2t(typeSpec.Type)) {
		case T_STRUCT:
			calcStructSizeAndSetFieldOffset(typeSpec)
		}
	}

	// collect methods in advance
	for _, funcDecl := range funcDecls {
		if funcDecl.Body != nil {
			if funcDecl.Recv != nil { // is Method
				method := newMethod(funcDecl)
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
			varSpec.Type = t.e
		}
		nameIdent.Obj.Data = newGlobalVariable(pkg.name, nameIdent.Obj.Name, e2t(varSpec.Type))
		pkg.vars = append(pkg.vars, varSpec)
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

		if funcDecl.Recv != nil { // Method
			paramFields = append(paramFields, funcDecl.Recv.List[0])
		}
		for _, field := range funcDecl.Type.Params.List {
			paramFields = append(paramFields, field)
		}

		for _, field := range paramFields {
			obj := field.Names[0].Obj
			obj.Data = fnc.registerParamVariable(obj.Name, e2t(field.Type))
		}
		if funcDecl.Body != nil {
			fnc.stmts = funcDecl.Body.List
			for _, stmt := range fnc.stmts {
				walkStmt(stmt)
			}

			if funcDecl.Recv != nil { // is Method
				fnc.method = newMethod(funcDecl)
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

var gInt32 = &ast.Object{
	Kind: ast.Typ,
	Name: "int",
	Decl: nil,
	Data: 0,
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

var gPanic = &ast.Object{
	Kind: ast.Fun,
	Name: "panic",
	Decl: nil,
	Data: nil,
	Type: nil,
}

// func type of runtime functions
var funcTypeOsExit = &ast.FuncType{
	Params: &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type:  tInt.e,
			},
		},
	},
	Results: nil,
}

var funcTypeSyscallOpen = &ast.FuncType{
	Params: &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type:  tString.e,
			},
			&ast.Field{
				Type:  tInt.e,
			},
			&ast.Field{
				Type:  tInt.e,
			},
		},
	},
	Results: &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type:  tInt.e,
			},
		},
	},
}

var funcTypeSyscallRead = &ast.FuncType{
	Params: &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type:  tInt.e,
			},
			&ast.Field{
				Type: generalSlice,
			},
		},
	},
	Results: &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type:  tInt.e,
			},
		},
	},
}

var funcTypeSyscallWrite = &ast.FuncType{
	Params: &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type:  tInt.e,
			},
			&ast.Field{
				Type: generalSlice,
			},
		},
	},
	Results: &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type:  tInt.e,
			},
		},
	},
}

var funcTypeSyscallSyscall = &ast.FuncType{
	Params: &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type:  tUintptr.e,
			},
			&ast.Field{
				Type:  tUintptr.e,
			},
			&ast.Field{
				Type:  tUintptr.e,
			},
			&ast.Field{
				Type:  tUintptr.e,
			},
		},
	},
	Results: &ast.FieldList{
		List: []*ast.Field{
			&ast.Field{
				Type:  tUintptr.e,
			},
		},
	},
}
func createUniverse() *ast.Scope {
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
	universe.Insert(gPanic)

	universe.Insert(&ast.Object{
		Kind: ast.Fun,
		Name: "print",
		Decl: nil,
		Data: nil,
		Type: nil,
	})

	// inject os identifier @TODO this should come from imports
	universe.Insert(&ast.Object{
		Kind: ast.Pkg,
		Name: "os",
		Decl: nil,
		Data: nil,
		Type: nil,
	})

	return universe
}

func resolveUniverse(fiile *ast.File, universe *ast.Scope) {
	var unresolved []*ast.Ident
	for _, ident := range fiile.Unresolved {
		if obj := universe.Lookup(ident.Name); obj != nil {
			ident.Obj = obj
		} else {
			unresolved = append(unresolved, ident)
		}
	}
}

// --- main ---
var pkg *PkgContainer

type PkgContainer struct {
	name string
	vars []*ast.ValueSpec
	funcs []*Func
}

func main() {
	var universe = createUniverse()
	var sourceFiles = []string{"runtime.go", "/dev/stdin"}

	for _, sourceFile := range sourceFiles {
		fmtPrintf("# file: %s\n", sourceFile)
		stringIndex = 0
		stringLiterals = nil
		fset := &token.FileSet{}
		astFile := parseFile(fset, sourceFile)
		resolveUniverse(astFile, universe)
		pkg = &PkgContainer{
			name: astFile.Name.Name,
		}
		logf("Package:   %s\n", pkg.name)
		walk(pkg, astFile)
		generateCode(pkg)
	}

	// emitting dynamic types
	fmtPrintf("# ------- Dynamic Types ------\n")
	fmtPrintf(".data\n")
	for name, id := range typeMap {
		symbol := typeIdToSymbol(id)
		fmtPrintf("%s: # %s\n", symbol, name)
		fmtPrintf("  .quad %s\n", Itoa(id))
		fmtPrintf("  .quad .S.dtype.%s\n", Itoa(id))
		fmtPrintf(".S.dtype.%s:\n", Itoa(id))
		fmtPrintf("  .string \"%s\"\n", name)
	}
	fmtPrintf("\n")
}
