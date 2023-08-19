package sema

import (
	"unsafe"

	"github.com/DQNEO/babygo/internal/ir"
	"github.com/DQNEO/babygo/internal/types"
	"github.com/DQNEO/babygo/internal/universe"
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/strconv"
	"github.com/DQNEO/babygo/lib/token"
)

var __func__ = "__func__"

var Fset *token.FileSet
var CurrentPkg *ir.PkgContainer
var exportedIdents = make(map[string]*ir.ExportedIdent)
var currentFor *ir.MetaForContainer
var currentFunc *ir.Func
var mapFieldOffset = make(map[unsafe.Pointer]int)

func Clear() {
	Fset = nil
	CurrentPkg = nil
	exportedIdents = nil
	currentFor = nil
	currentFunc = nil
	mapFieldOffset = nil
}

func assert(bol bool, msg string, caller string) {
	if !bol {
		panic(caller + ": " + msg)
	}
}

const ThrowFormat string = "%T"

func throw(x interface{}) {
	panic(fmt.Sprintf(ThrowFormat, x))
}

func unexpectedKind(knd types.TypeKind) {
	panic("Unexpected Kind: " + string(knd))
}

func panicPos(s string, pos token.Pos) {
	position := Fset.Position(pos)
	panic(fmt.Sprintf("%s\n\t%s", s, position.String()))
}

func LinePosition(pos token.Pos) string {
	posit := Fset.Position(pos)
	return fmt.Sprintf("%s:%d", posit.Filename, posit.Line)
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
			qi := Selector2QI(e)
			ei := LookupForeignIdent(qi, e.Pos())
			return ei.IsType
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

type astArgAndParam struct {
	e         ast.Expr
	paramType *types.Type // expected type
}

type argAndParamType struct {
	Meta      ir.MetaExpr
	ParamType *types.Type // expected type
}

func prepareArgsAndParams(funcType *ast.FuncType, receiver ir.MetaExpr, eArgs []ast.Expr, expandElipsis bool) []*argAndParamType {
	if funcType == nil {
		panic("no funcType")
	}
	var args []*astArgAndParam
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

		paramType := E2T(param.Type)
		arg := &astArgAndParam{
			e:         eArg,
			paramType: paramType,
		}
		args = append(args, arg)
	}

	if variadicElmType != nil && !expandElipsis {
		// collect args as a slice
		pos := funcType.Pos()
		sliceType := &ast.ArrayType{
			Elt:    variadicElmType,
			Lbrack: pos,
		}
		vargsSliceWrapper := &ast.CompositeLit{
			Type:   sliceType,
			Elts:   variadicArgs,
			Lbrace: pos,
		}
		args = append(args, &astArgAndParam{
			e:         vargsSliceWrapper,
			paramType: E2T(sliceType),
		})
	} else if len(args) < len(params) {
		// Add nil as a variadic arg
		param := params[len(args)]
		elp, ok := param.Type.(*ast.Ellipsis)
		if !ok {
			panicPos("Should be Ellipsis", param.Type.Pos())
		}
		paramType := E2T(elp)
		iNil := &ast.Ident{
			Obj:     universe.Nil,
			Name:    "nil",
			NamePos: funcType.Pos(),
		}
		//		exprTypeMeta[unsafe.Pointer(iNil)] = E2T(elp)
		args = append(args, &astArgAndParam{
			e:         iNil,
			paramType: paramType,
		})
	}

	var metaArgs []*argAndParamType
	for _, arg := range args {
		ctx := &ir.EvalContext{Type: arg.paramType}
		m := walkExpr(arg.e, ctx)
		a := &argAndParamType{
			Meta:      m,
			ParamType: arg.paramType,
		}
		metaArgs = append(metaArgs, a)
	}

	if receiver != nil { // method call
		paramType := GetTypeOfExpr(receiver)
		if paramType == nil {
			panic("[prepaareArgs] param type must not be nil")
		}
		var receiverAndArgs []*argAndParamType = []*argAndParamType{
			&argAndParamType{
				ParamType: paramType,
				Meta:      receiver,
			},
		}
		for _, arg := range metaArgs {
			receiverAndArgs = append(receiverAndArgs, arg)
		}
		return receiverAndArgs
	}

	return metaArgs
}

func NewFuncValueFromSymbol(symbol string) *ir.FuncValue {
	return &ir.FuncValue{
		IsDirect: true,
		Symbol:   symbol,
	}
}

func IsNil(meta ir.MetaExpr) bool {
	m, ok := meta.(*ir.MetaIdent)
	if !ok {
		return false
	}
	return isUniverseNil(m)
}

func NewNumberLiteral(x int, pos token.Pos) *ir.MetaBasicLit {
	return &ir.MetaBasicLit{
		Tpos:   pos,
		Type:   types.Int,
		Kind:   "INT",
		IntVal: x,
	}
}

func IsBlankIdentifierMeta(m ir.MetaExpr) bool {
	ident, isIdent := m.(*ir.MetaIdent)
	if !isIdent {
		return false
	}
	return ident.Kind == "blank"
}

func GetMethodSymbol(method *ir.Method) string {
	rcvTypeName := method.RcvNamedType
	var subsymbol string
	if method.IsPtrMethod {
		subsymbol = "$" + rcvTypeName.Name + "." + method.Name // pointer
	} else {
		subsymbol = rcvTypeName.Name + "." + method.Name // value
	}

	return GetPackageSymbol(method.PkgName, subsymbol)
}

func GetPackageSymbol(pkgName string, subsymbol string) string {
	return pkgName + "." + subsymbol
}

func GetGoTypeOfExpr(meta ir.MetaExpr) types.GoType {
	t := GetTypeOfExpr(meta)
	return t.GoType
}

// Types of an expr in Single value context
func GetTypeOfExpr(meta ir.MetaExpr) *types.Type {
	var t *types.Type
	switch m := meta.(type) {
	case *ir.MetaBasicLit:
		t = m.Type
	case *ir.MetaCompositLit:
		t = m.Type
	case *ir.MetaIdent:
		t = m.Type
	case *ir.Variable:
		t = m.Type
	case *ir.MetaSelectorExpr:
		t = m.Type
	case *ir.MetaConversionExpr:
		t = m.Type
	case *ir.MetaCallLen:
		t = m.Type
	case *ir.MetaCallCap:
		t = m.Type
	case *ir.MetaCallNew:
		t = m.Type
	case *ir.MetaCallMake:
		t = m.Type
	case *ir.MetaCallAppend:
		t = m.Type
	case *ir.MetaCallPanic:
		t = m.Type
	case *ir.MetaCallDelete:
		t = m.Type
	case *ir.MetaCallExpr: // funcall
		t = m.Type // can be nil (e.g. panic()). if Tuple , m.Types has Types
	case *ir.MetaIndexExpr:
		t = m.Type
	case *ir.MetaSliceExpr:
		t = m.Type
	case *ir.MetaStarExpr:
		t = m.Type
	case *ir.MetaUnaryExpr:
		t = m.Type
	case *ir.MetaBinaryExpr:
		t = m.Type
	case *ir.MetaTypeAssertExpr:
		t = m.Type
	case *ir.IfcConversion:
		t = m.Type
	}
	if t == nil {
		panic(fmt.Sprintf("bad type:%T\n", meta))
	}
	if t.GoType == nil {
		panic(fmt.Sprintf("GoType is not set:%T\n", meta))
	}
	return t
}

func FieldList2Types(fieldList *ast.FieldList) []*types.Type {
	if fieldList == nil {
		return nil
	}
	var r []*types.Type
	for _, e2 := range fieldList.List {
		t := E2T(e2.Type)
		r = append(r, t)
	}
	return r
}

func FieldList2Tuple(fieldList *ast.FieldList) *types.Tuple {
	if fieldList == nil {
		return nil
	}
	var r = &types.Tuple{}
	for _, e2 := range fieldList.List {
		ident, isIdent := e2.Type.(*ast.Ident)
		var t types.GoType
		if isIdent && ident.Name == inNamed {
			t = inNamedType
		} else {
			t = E2G(e2.Type)
		}

		r.Types = append(r.Types, t)
	}
	return r
}

func GetTupleTypes(rhsMeta ir.MetaExpr) []*types.Type {
	if IsOkSyntax(rhsMeta) {
		return []*types.Type{GetTypeOfExpr(rhsMeta), types.Bool}
	} else {
		rhs, ok := rhsMeta.(*ir.MetaCallExpr)
		if !ok {
			panic("is not *MetaCallExpr")
		}
		return rhs.Types
	}
}

var inNamed string
var inNamedType *types.Named

func E2G(typeExpr ast.Expr) types.GoType {
	switch t := typeExpr.(type) {
	case *ast.Ident:
		obj := t.Obj
		if obj == nil {
			panicPos("t.Obj should not be nil", typeExpr.Pos())
		}
		switch obj {
		case universe.Uintptr:
			return types.Uintptr.GoType
		case universe.Int:
			return types.Int.GoType
		case universe.Int32:
			return types.Int32.GoType
		case universe.String:
			return types.String.GoType
		case universe.Uint8:
			return types.Uint8.GoType
		case universe.Uint16:
			return types.Uint16.GoType
		case universe.Bool:
			return types.Bool.GoType
		case universe.Error:
			dcl := universe.Error.Decl.(*ast.TypeSpec)
			ut := E2G(dcl.Type)
			named := types.NewNamed(universe.Error.Name, ut)
			return named
		default:
			switch dcl := t.Obj.Decl.(type) {
			case *ast.TypeSpec:
				typeSpec := dcl
				//util.Logf("[E2G] type %s\n", typeSpec.Name.Name)
				gt := types.NewNamed(typeSpec.Name.Name, nil)
				if typeSpec.Name.Obj.Data != nil {
					gt.PkgName = typeSpec.Name.Obj.Data.(string)
				}
				inNamed = typeSpec.Name.Name
				inNamedType = gt
				ut := E2G(typeSpec.Type)
				gt.Uunderlying = ut
				inNamedType = nil
				inNamed = ""
				return gt
			default:
				panicPos(fmt.Sprintf("Unexpeced:%T ident=%s", t.Obj.Decl, t.Name), t.Pos())
			}
			panic("Unexpected flow")
		}
	case *ast.ArrayType:
		if t.Len == nil {
			return types.NewSlice(E2G(t.Elt))
		} else {
			return types.NewArray(E2G(t.Elt), EvalInt(t.Len))
		}
	case *ast.StructType:
		var fields []*types.Var
		var astFields []*ast.Field
		if t.Fields != nil {
			for _, fld := range t.Fields.List {
				astFields = append(astFields, fld)
				ft := E2G(fld.Type)
				v := &types.Var{
					Name: fld.Names[0].Name,
					Typ:  ft,
				}
				fields = append(fields, v)
			}
		}
		return types.NewStruct(fields, astFields)
	case *ast.StarExpr:
		if inNamedType != nil {
			ident, ok := t.X.(*ast.Ident)
			if ok && ident.Name == inNamed {
				p := types.NewPointer(inNamedType)
				return p
			}
		}
		return types.NewPointer(E2G(t.X))
	case *ast.Ellipsis:
		slc := types.NewSlice(E2G(t.Elt))
		slc.Elp = true
		return slc
	case *ast.MapType:
		return types.NewMap(E2G(t.Key), E2G(t.Value))
	case *ast.InterfaceType:
		var methods []*types.Func
		if t.Methods != nil {
			for _, m := range t.Methods.List {
				methodName := m.Names[0].Name
				t := E2G(m.Type)
				f := &types.Func{
					Typ:  t,
					Name: methodName,
				}
				methods = append(methods, f)
			}
		}
		return types.NewInterfaceType(methods)
	case *ast.FuncType:
		sig := &types.Signature{}
		if t.Params != nil {
			sig.Params = FieldList2Tuple(t.Params)
		}
		if t.Results != nil {
			sig.Results = FieldList2Tuple(t.Results)
		}
		//util.Logf("%s:[E2G] handling *ast.FuncType\n", Fset.Position(t.Pos()).String())
		return types.NewFunc(sig)
	case *ast.ParenExpr:
		typeExpr = t.X
		return E2G(typeExpr)
	case *ast.SelectorExpr:
		if isQI(t) { // e.g. unsafe.Pointer
			ei := LookupForeignIdent(Selector2QI(t), t.Pos())
			return ei.Type.GoType
		} else {
			panic("@TBI")
		}
	}

	panic(fmt.Sprintf("should not reach here: %T\n", typeExpr))

	return nil
}

func E2T(typeExpr ast.Expr) *types.Type {
	if typeExpr == nil {
		panic("nil is not allowed")
	}

	// unwrap paren
	switch e := typeExpr.(type) {
	case *ast.ParenExpr:
		typeExpr = e.X
		return E2T(typeExpr)
	}
	g := E2G(typeExpr)
	return &types.Type{
		E:      typeExpr,
		GoType: g,
	}
}

func GetArrayLen(t types.GoType) int {
	t = t.Underlying()
	arrayType := t.(*types.Array)
	return arrayType.Len()
}

func GetUnderlyingStructType(t *types.Type) *ast.StructType {
	ut := GetUnderlyingType(t)
	return ut.E.(*ast.StructType)
}

func GetUnderlyingType(t *types.Type) *types.Type {
	if t == nil {
		panic("nil type is not expected")
	}
	if t == types.GeneralSliceType {
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
		case universe.Uintptr, universe.Int, universe.Int32, universe.String, universe.Uint8, universe.Uint16, universe.Bool:
			return t
		}
		if e.Obj.Decl == nil {
			panic("e.Obj.Decl should not be nil: Obj.Name=" + e.Obj.Name)
		}

		// defined type or alias
		typeSpec := e.Obj.Decl.(*ast.TypeSpec)
		specType := typeSpec.Type
		t := E2T(specType)
		// get RHS in its type definition recursively
		return GetUnderlyingType(t)
	case *ast.SelectorExpr:
		ei := LookupForeignIdent(Selector2QI(e), e.Pos())
		assert(ei.IsType, "should be a type", __func__)
		return GetUnderlyingType(ei.Type)
	case *ast.ParenExpr:
		return GetUnderlyingType(E2T(e.X))
	case *ast.FuncType:
		return t
	}
	throw(t.E)
	return nil
}

func Kind2(gType types.GoType) types.TypeKind {
	if gType == nil {
		panic(fmt.Sprintf("[Kind2] Unexpected nil:\n"))
	}

	switch gt := gType.(type) {
	case *types.Basic:
		switch gt.Kind() {
		case types.GBool:
			return types.T_BOOL
		case types.GInt:
			return types.T_INT
		case types.GInt32:
			return types.T_INT32
		case types.GUint8:
			return types.T_UINT8
		case types.GUint16:
			return types.T_UINT16
		case types.GUintptr:
			return types.T_UINTPTR
		case types.GString:
			return types.T_STRING
		default:
			panicPos("TBI: unknown gt.Kind", 1)
		}
	case *types.Array:
		return types.T_ARRAY
	case *types.Slice:
		return types.T_SLICE
	case *types.Struct:
		return types.T_STRUCT
	case *types.Pointer:
		return types.T_POINTER
	case *types.Map:
		return types.T_MAP
	case *types.Interface:
		return types.T_INTERFACE
	case *types.Func:
		return types.T_FUNC
	case *types.Signature:
		return types.T_FUNC
	case *types.Tuple:
		panic(fmt.Sprintf("Tuple is not expected: type %T\n", gType))
	case *types.Named:
		ut := gt.Underlying()
		if ut == nil {
			panic(fmt.Sprintf("nil is not expected: NamedType %s\n", gt.String()))
		}
		//t := &types.Type{GoType: ut}
		return Kind2(ut)
	default:
		panic(fmt.Sprintf("[Kind2] Unexpected type: %T\n", gType))
		//panicPos(fmt.Sprintf("Unexpected type %T\n", gType), t.E.Pos())
	}
	return "UNKNOWN_KIND"
}

func Kind(t *types.Type) types.TypeKind {
	if t == nil {
		panicPos("nil type is not expected", t.E.Pos())
	}
	if t.GoType == nil {
		panic(fmt.Sprintf("[Kind] Unexpected GoType nil: %T\n", t.E))
	}
	return Kind2(t.GoType)
}

func IsInterface(t types.GoType) bool {
	return Kind2(t) == types.T_INTERFACE
}

func HasIfcMethod(t types.GoType) bool {
	if !IsInterface(t) {
		panic("type should be an interface")
	}

	ut := t.Underlying()
	ifc, ok := ut.(*types.Interface)
	if !ok {
		panic("type should be an interface")
	}
	if len(ifc.Methods) > 0 {
		return true
	}
	return false
}

func GetElementTypeOfCollectionType(t *types.Type) *types.Type {
	ut := GetUnderlyingType(t)
	switch Kind(ut) {
	case types.T_SLICE, types.T_ARRAY:
		switch e := ut.E.(type) {
		case *ast.ArrayType:
			return E2T(e.Elt)
		case *ast.Ellipsis:
			return E2T(e.Elt)
		default:
			throw(t.E)
		}
	case types.T_STRING:
		return types.Uint8
	case types.T_MAP:
		mapType := ut.E.(*ast.MapType)
		return E2T(mapType.Value)
	default:
		unexpectedKind(Kind(t))
	}
	return nil
}

func getKeyTypeOfCollectionType(t *types.Type) *types.Type {
	ut := GetUnderlyingType(t)
	switch Kind(ut) {
	case types.T_SLICE, types.T_ARRAY, types.T_STRING:
		return types.Int
	case types.T_MAP:
		mapType := ut.E.(*ast.MapType)
		return E2T(mapType.Key)
	default:
		unexpectedKind(Kind(t))
	}
	return nil
}

func LookupStructField(structType *ast.StructType, selName string) *ast.Field {
	for _, field := range structType.Fields.List {
		if field.Names[0].Name == selName {
			return field
		}
	}
	//	panicPos("Unexpected flow: struct field not found:  "+selName, structType.Pos())
	return nil
}

func registerParamVariable(fnc *ir.Func, name string, t *types.Type) *ir.Variable {
	vr := newLocalVariable(name, fnc.Argsarea, t)
	size := GetSizeOfType(t)
	fnc.Argsarea += size
	fnc.Params = append(fnc.Params, vr)
	return vr
}

func registerReturnVariable(fnc *ir.Func, name string, t *types.Type) *ir.Variable {
	vr := newLocalVariable(name, fnc.Argsarea, t)
	size := GetSizeOfType(t)
	fnc.Argsarea += size
	fnc.Retvars = append(fnc.Retvars, vr)
	return vr
}

func registerLocalVariable(fnc *ir.Func, name string, t *types.Type) *ir.Variable {
	assert(t != nil && t.E != nil, "type of local var should not be nil", __func__)
	fnc.Localarea -= GetSizeOfType(t)
	vr := newLocalVariable(name, currentFunc.Localarea, t)
	fnc.LocalVars = append(fnc.LocalVars, vr)
	return vr
}

func registerStringLiteral(lit *ast.BasicLit) *ir.SLiteral {
	if CurrentPkg.Name == "" {
		panic("no pkgName")
	}

	var strlen int
	for _, c := range []uint8(lit.Value) {
		if c != '\\' {
			strlen++
		}
	}

	label := fmt.Sprintf(".string_%d", CurrentPkg.StringIndex)
	CurrentPkg.StringIndex++
	sl := &ir.SLiteral{
		Label:  label,
		Strlen: strlen - 2,
		Value:  lit.Value,
	}
	CurrentPkg.StringLiterals = append(CurrentPkg.StringLiterals, sl)
	return sl
}

func newGlobalVariable(pkgName string, name string, t *types.Type) *ir.Variable {
	return &ir.Variable{
		Name:         name,
		IsGlobal:     true,
		GlobalSymbol: pkgName + "." + name,
		Type:         t,
	}
}

func newLocalVariable(name string, localoffset int, t *types.Type) *ir.Variable {
	return &ir.Variable{
		Name:        name,
		IsGlobal:    false,
		LocalOffset: localoffset,
		Type:        t,
	}
}

func NewQI(pkg string, ident string) ir.QualifiedIdent {
	return ir.QualifiedIdent(pkg + "." + ident)
}

func isQI(e *ast.SelectorExpr) bool {
	ident, isIdent := e.X.(*ast.Ident)
	if !isIdent {
		return false
	}
	if ident.Obj == nil {
		panicPos("ident.Obj should not be nil:"+ident.Name, e.Pos())
	}
	return ident.Obj.Kind == ast.Pkg
}

func Selector2QI(e *ast.SelectorExpr) ir.QualifiedIdent {
	pkgName := e.X.(*ast.Ident)
	assert(pkgName.Obj.Kind == ast.Pkg, "should be ast.Pkg", __func__)
	return NewQI(pkgName.Name, e.Sel.Name)
}

func newMethod(pkgName string, funcDecl *ast.FuncDecl) *ir.Method {
	rcvType := funcDecl.Recv.List[0].Type
	rcvPointerType, isPtr := rcvType.(*ast.StarExpr)
	if isPtr {
		rcvType = rcvPointerType.X
	}
	rcvNamedType := rcvType.(*ast.Ident)
	method := &ir.Method{
		PkgName:      pkgName,
		RcvNamedType: rcvNamedType,
		IsPtrMethod:  isPtr,
		Name:         funcDecl.Name.Name,
		FuncType:     funcDecl.Type,
	}
	return method
}

// https://golang.org/ref/spec#Method_sets
var namedTypes = make(map[string]*ir.NamedType)

// @TODO: enable to register ifc method
func registerMethod(pkgName string, method *ir.Method) {
	namedTypeId := pkgName + "." + method.RcvNamedType.Name
	namedType, ok := namedTypes[namedTypeId]
	if !ok {
		namedType = &ir.NamedType{
			MethodSet: make(map[string]*ir.Method),
		}
		namedTypes[namedTypeId] = namedType
	}
	//util.Logf("registerMethod: pkg=%s namedTypeId=%s namedType=%s\n", pkgName, namedTypeId, method.RcvNamedType.Obj.Name)
	namedType.MethodSet[method.Name] = method
}

// @TODO: enable to lookup ifc method
func LookupMethod(rcvT types.GoType, methodName string) *ir.Method {
	rcvPointerType, isPtr := rcvT.(*types.Pointer)
	if isPtr {
		rcvT = rcvPointerType.Elem()
	}
	var namedTypeId string

	switch typ := rcvT.(type) {
	case *types.Named:
		if typ.PkgName == "" && typ.String() == "error" {
			namedTypeId = "error"
			return &ir.Method{
				PkgName: "",
				RcvNamedType: &ast.Ident{
					Name: "",
				},
				IsPtrMethod: false,
				Name:        "error",
				FuncType:    universe.ErrorMethodFuncType,
			}
		} else {
			pkgName := typ.PkgName
			namedTypeId = pkgName + "." + typ.String()
			//util.Logf("[LookupMethod] ident: namedTypeId=%s\n", namedTypeId)
		}
	default:
		panic("Unexpected type")
	}

	namedType, ok := namedTypes[namedTypeId]
	if !ok {
		panic(namedTypeId + " has no methodSet")
	}
	method, ok := namedType.MethodSet[methodName]
	if !ok {
		panic("method not found: " + methodName)
	}
	return method
}

func walkExprStmt(s *ast.ExprStmt) *ir.MetaExprStmt {
	m := walkExpr(s.X, nil)
	return &ir.MetaExprStmt{
		X:    m,
		Tpos: s.Pos(),
	}
}

func walkDeclStmt(s *ast.DeclStmt) *ir.MetaVarDecl {
	genDecl := s.Decl.(*ast.GenDecl)
	declSpec := genDecl.Specs[0]
	switch spec := declSpec.(type) {
	case *ast.ValueSpec:
		lhsIdent := spec.Names[0]
		var rhsMeta ir.MetaExpr
		var t *types.Type
		if spec.Type != nil { // var x T = e
			// walkExpr(spec.Type, nil) // Do we need to walk type ?
			t = E2T(spec.Type)
			if len(spec.Values) > 0 {
				rhs := spec.Values[0]
				ctx := &ir.EvalContext{Type: t}
				rhsMeta = walkExpr(rhs, ctx)
				rhsMeta = CheckIfcConversion(rhs.Pos(), rhsMeta, t)
			}
		} else { // var x = e  infer lhs type from rhs
			if len(spec.Values) == 0 {
				panic("invalid syntax")
			}

			rhs := spec.Values[0]
			rhsMeta = walkExpr(rhs, nil)
			t = GetTypeOfExpr(rhsMeta)
			if t == nil {
				panic("rhs should have a type")
			}
		}
		spec.Type = t.E // set lhs type

		obj := lhsIdent.Obj
		SetVariable(obj, registerLocalVariable(currentFunc, obj.Name, t))
		lhsMeta := WalkIdent(lhsIdent, nil)
		single := &ir.MetaSingleAssign{
			Tpos: lhsIdent.Pos(),
			Lhs:  lhsMeta,
			Rhs:  rhsMeta,
		}
		return &ir.MetaVarDecl{
			Tpos:    lhsIdent.Pos(),
			Single:  single,
			LhsType: t,
		}
	default:
		// @TODO type, const, etc
	}

	panic("TBI 3366")
}

func IsOkSyntax(rhs ir.MetaExpr) bool {
	typeAssertion, isTypeAssertion := rhs.(*ir.MetaTypeAssertExpr)
	if isTypeAssertion && typeAssertion.NeedsOK {
		return true
	}
	indexExpr, isIndexExpr := rhs.(*ir.MetaIndexExpr)
	if isIndexExpr && indexExpr.NeedsOK {
		return true
	}
	return false
}

func walkAssignStmt(s *ast.AssignStmt) ir.MetaStmt {
	pos := s.Pos()
	stok := s.Tok.String()
	switch stok {
	case "=":
		if len(s.Lhs) == 1 && len(s.Rhs) == 1 {
			// Single assignment
			var lhsMetas []ir.MetaExpr
			for _, lhs := range s.Lhs {
				lm := walkExpr(lhs, nil)
				lhsMetas = append(lhsMetas, lm)
			}
			var ctx *ir.EvalContext
			var t *types.Type
			if !IsBlankIdentifierMeta(lhsMetas[0]) {
				t = GetTypeOfExpr(lhsMetas[0])
				ctx = &ir.EvalContext{
					Type: t,
				}
			}
			rhsMeta := walkExpr(s.Rhs[0], ctx)
			if t == nil {
				t = GetTypeOfExpr(rhsMeta)
			}
			mc := CheckIfcConversion(rhsMeta.Pos(), rhsMeta, t)
			//checkIfcConversion(mc)
			return &ir.MetaSingleAssign{
				Tpos: pos,
				Lhs:  lhsMetas[0],
				Rhs:  mc,
			}
		} else if len(s.Lhs) == len(s.Rhs) {
			panic("TBI 3404")
		} else if len(s.Lhs) > 1 && len(s.Rhs) == 1 {
			// Tuple assignment
			maybeOkContext := len(s.Lhs) == 2
			rhsMeta := walkExpr(s.Rhs[0], &ir.EvalContext{MaybeOK: maybeOkContext})
			isOK := len(s.Lhs) == 2 && IsOkSyntax(rhsMeta)
			rhsTypes := GetTupleTypes(rhsMeta)
			assert(len(s.Lhs) == len(rhsTypes), fmt.Sprintf("length unmatches %d <=> %d", len(s.Lhs), len(rhsTypes)), __func__)

			var lhsMetas []ir.MetaExpr
			for _, lhs := range s.Lhs {
				lm := walkExpr(lhs, nil)
				lhsMetas = append(lhsMetas, lm)
			}
			return &ir.MetaTupleAssign{
				Tpos:     pos,
				IsOK:     isOK,
				Lhss:     lhsMetas,
				Rhs:      rhsMeta,
				RhsTypes: rhsTypes,
			}
		} else {
			panic("Bad syntax")
		}
	case ":=":
		if len(s.Lhs) == 1 && len(s.Rhs) == 1 {
			// Single assignment
			rhsMeta := walkExpr(s.Rhs[0], nil) // FIXME
			rhsType := GetTypeOfExpr(rhsMeta)
			lhsTypes := []*types.Type{rhsType}
			var lhsMetas []ir.MetaExpr
			for i, lhs := range s.Lhs {
				typ := lhsTypes[i]
				obj := lhs.(*ast.Ident).Obj
				SetVariable(obj, registerLocalVariable(currentFunc, obj.Name, typ))
				lm := walkExpr(lhs, nil)
				lhsMetas = append(lhsMetas, lm)
			}

			return &ir.MetaSingleAssign{
				Tpos: pos,
				Lhs:  lhsMetas[0],
				Rhs:  rhsMeta,
			}
		} else if len(s.Lhs) == len(s.Rhs) {
			panic("TBI 3447")
		} else if len(s.Lhs) > 1 && len(s.Rhs) == 1 {
			// Tuple assignment
			maybeOkContext := len(s.Lhs) == 2
			rhsMeta := walkExpr(s.Rhs[0], &ir.EvalContext{MaybeOK: maybeOkContext})
			isOK := len(s.Lhs) == 2 && IsOkSyntax(rhsMeta)
			rhsTypes := GetTupleTypes(rhsMeta)
			assert(len(s.Lhs) == len(rhsTypes), fmt.Sprintf("length unmatches %d <=> %d", len(s.Lhs), len(rhsTypes)), __func__)

			lhsTypes := rhsTypes
			for i, lhs := range s.Lhs {
				typ := lhsTypes[i]
				obj := lhs.(*ast.Ident).Obj
				SetVariable(obj, registerLocalVariable(currentFunc, obj.Name, typ))
			}

			var lhsMetas []ir.MetaExpr
			for _, lhs := range s.Lhs {
				lm := walkExpr(lhs, nil)
				lhsMetas = append(lhsMetas, lm)
			}
			return &ir.MetaTupleAssign{
				Tpos:     pos,
				IsOK:     isOK,
				Lhss:     lhsMetas,
				Rhs:      rhsMeta,
				RhsTypes: rhsTypes,
			}
		} else {
			panic("Bad syntax")
		}
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
		return &ir.MetaSingleAssign{
			Tpos: pos,
			Lhs:  lhsMeta,
			Rhs:  rhsMeta,
		}
	default:
		panic("TBI 3497 ")
	}
	return nil
}

func walkReturnStmt(s *ast.ReturnStmt) *ir.MetaReturnStmt {
	funcDef := currentFunc

	if len(funcDef.Retvars) > 1 && len(s.Results) == 1 {
		// Tuple assign:
		// return f()
		funcall, ok := s.Results[0].(*ast.CallExpr)
		if !ok {
			panic("syntax error in return statement")
		}
		m := walkExpr(funcall, nil)
		tupleTypes := GetTupleTypes(m)
		assert(len(funcDef.Retvars) == len(tupleTypes), "number of return exprs should match", __func__)
		var lhss []ir.MetaExpr
		for _, v := range funcDef.Retvars {
			lhss = append(lhss, v)
		}
		//@TODO: CheckIfcConversion
		ta := &ir.MetaTupleAssign{
			Tpos:     s.Pos(),
			IsOK:     false,
			Lhss:     lhss,
			Rhs:      m,
			RhsTypes: tupleTypes,
		}
		return &ir.MetaReturnStmt{
			Tpos:        s.Pos(),
			IsTuple:     true,
			TupleAssign: ta,
		}
	}

	_len := len(funcDef.Retvars)
	var sas []*ir.MetaSingleAssign
	for i := 0; i < _len; i++ {
		expr := s.Results[i]
		retTyp := funcDef.Retvars[i].Type
		ctx := &ir.EvalContext{
			Type: retTyp,
		}
		m := walkExpr(expr, ctx)
		mc := CheckIfcConversion(expr.Pos(), m, retTyp)
		as := &ir.MetaSingleAssign{
			Tpos: s.Pos(),
			Lhs:  funcDef.Retvars[i],
			Rhs:  mc,
		}
		sas = append(sas, as)
	}
	return &ir.MetaReturnStmt{
		Tpos:              s.Pos(),
		IsTuple:           false,
		SingleAssignments: sas,
	}
}

func walkIfStmt(s *ast.IfStmt) *ir.MetaIfStmt {
	var mInit ir.MetaStmt
	var mElse ir.MetaStmt
	var condMeta ir.MetaExpr
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
	return &ir.MetaIfStmt{
		Tpos: s.Pos(),
		Init: mInit,
		Cond: condMeta,
		Body: mtBlock,
		Else: mElse,
	}
}

func walkBlockStmt(s *ast.BlockStmt) *ir.MetaBlockStmt {
	mt := &ir.MetaBlockStmt{
		Tpos: s.Pos(),
	}
	for _, stmt := range s.List {
		meta := walkStmt(stmt)
		mt.List = append(mt.List, meta)
	}
	return mt
}

func walkForStmt(s *ast.ForStmt) *ir.MetaForContainer {
	meta := &ir.MetaForContainer{
		Tpos:    s.Pos(),
		Outer:   currentFor,
		ForStmt: &ir.MetaForForStmt{},
	}
	currentFor = meta

	if s.Init != nil {
		meta.ForStmt.Init = walkStmt(s.Init)
	}
	if s.Cond != nil {
		meta.ForStmt.Cond = walkExpr(s.Cond, nil)
	}
	if s.Post != nil {
		meta.ForStmt.Post = walkStmt(s.Post)
	}
	meta.Body = walkBlockStmt(s.Body)
	currentFor = meta.Outer
	return meta
}
func walkRangeStmt(s *ast.RangeStmt) *ir.MetaForContainer {
	meta := &ir.MetaForContainer{
		Tpos:  s.Pos(),
		Outer: currentFor,
	}
	currentFor = meta
	metaX := walkExpr(s.X, nil)

	collectionType := GetUnderlyingType(GetTypeOfExpr(metaX))
	keyType := getKeyTypeOfCollectionType(collectionType)
	elmType := GetElementTypeOfCollectionType(collectionType)
	walkExpr(types.Int.E, nil)
	switch Kind(collectionType) {
	case types.T_SLICE, types.T_ARRAY:
		meta.ForRangeStmt = &ir.MetaForRangeStmt{
			Tpos:     s.Pos(),
			IsMap:    false,
			LenVar:   registerLocalVariable(currentFunc, ".range.len", types.Int),
			Indexvar: registerLocalVariable(currentFunc, ".range.index", types.Int),
			X:        metaX,
		}
	case types.T_MAP:
		mapVar := registerLocalVariable(currentFunc, ".range.map", types.Uintptr)
		mapVarAssign := &ir.MetaSingleAssign{
			Tpos: s.Pos(),
			Lhs:  mapVar,
			Rhs:  metaX,
		}
		meta.ForRangeStmt = &ir.MetaForRangeStmt{
			Tpos:         s.Pos(),
			IsMap:        true,
			MapVar:       mapVar,
			ItemVar:      registerLocalVariable(currentFunc, ".range.item", types.Uintptr),
			MapVarAssign: mapVarAssign,
		}
	default:
		throw(collectionType)
	}
	if s.Tok.String() == ":=" {
		// declare local variables
		keyIdent := s.Key.(*ast.Ident)
		SetVariable(keyIdent.Obj, registerLocalVariable(currentFunc, keyIdent.Name, keyType))

		valueIdent := s.Value.(*ast.Ident)
		SetVariable(valueIdent.Obj, registerLocalVariable(currentFunc, valueIdent.Name, elmType))
	}
	if s.Key != nil {
		meta.ForRangeStmt.Key = walkExpr(s.Key, nil)
	}
	if s.Value != nil {
		meta.ForRangeStmt.Value = walkExpr(s.Value, nil)
	}

	mtBlock := walkBlockStmt(s.Body)
	meta.Body = mtBlock
	currentFor = meta.Outer
	return meta
}

func walkIncDecStmt(s *ast.IncDecStmt) *ir.MetaSingleAssign {
	var binop token.Token
	switch s.Tok.String() {
	case "++":
		binop = token.ADD
	case "--":
		binop = token.SUB
	default:
		panic("Unexpected Tok=" + s.Tok.String())
	}
	exprOne := &ast.BasicLit{
		Kind:     token.INT,
		Value:    "1",
		ValuePos: 1,
	}
	newRhs := &ast.BinaryExpr{
		X:  s.X,
		Y:  exprOne,
		Op: binop,
	}
	rhsMeta := walkExpr(newRhs, nil)
	lhsMeta := walkExpr(s.X, nil)
	return &ir.MetaSingleAssign{
		Tpos: s.Pos(),
		Lhs:  lhsMeta,
		Rhs:  rhsMeta,
	}
}

func walkSwitchStmt(s *ast.SwitchStmt) *ir.MetaSwitchStmt {
	meta := &ir.MetaSwitchStmt{
		Tpos: s.Pos(),
	}
	if s.Init != nil {
		meta.Init = walkStmt(s.Init)
	}
	if s.Tag != nil {
		meta.Tag = walkExpr(s.Tag, nil)
	}
	var cases []*ir.MetaCaseClause
	for _, _case := range s.Body.List {
		cc := _case.(*ast.CaseClause)
		_cc := walkCaseClause(cc)
		cases = append(cases, _cc)
	}
	meta.Cases = cases

	return meta
}

func walkTypeSwitchStmt(e *ast.TypeSwitchStmt) *ir.MetaTypeSwitchStmt {
	typeSwitch := &ir.MetaTypeSwitchStmt{
		Tpos: e.Pos(),
	}
	var assignIdent *ast.Ident

	switch assign := e.Assign.(type) {
	case *ast.ExprStmt:
		typeAssertExpr := assign.X.(*ast.TypeAssertExpr)
		typeSwitch.Subject = walkExpr(typeAssertExpr.X, nil)
	case *ast.AssignStmt:
		lhs := assign.Lhs[0]
		assignIdent = lhs.(*ast.Ident)
		typeSwitch.AssignObj = assignIdent.Obj
		// ident will be a new local variable in each case clause
		typeAssertExpr := assign.Rhs[0].(*ast.TypeAssertExpr)
		typeSwitch.Subject = walkExpr(typeAssertExpr.X, nil)
	default:
		throw(e.Assign)
	}

	typeSwitch.SubjectVariable = registerLocalVariable(currentFunc, ".switch_expr", types.Eface)

	var cases []*ir.MetaTypeSwitchCaseClose
	for _, _case := range e.Body.List {
		cc := _case.(*ast.CaseClause)
		tscc := &ir.MetaTypeSwitchCaseClose{
			Tpos: cc.Pos(),
		}
		cases = append(cases, tscc)

		if assignIdent != nil {
			if len(cc.List) > 0 {
				var varType *types.Type
				if isNilIdent(cc.List[0]) {
					varType = GetTypeOfExpr(typeSwitch.Subject)
				} else {
					varType = E2T(cc.List[0])
				}
				// inject a variable of that type
				vr := registerLocalVariable(currentFunc, assignIdent.Name, varType)
				tscc.Variable = vr
				SetVariable(assignIdent.Obj, vr)
			} else {
				// default clause
				// inject a variable of subject type
				varType := GetTypeOfExpr(typeSwitch.Subject)
				vr := registerLocalVariable(currentFunc, assignIdent.Name, varType)
				tscc.Variable = vr
				SetVariable(assignIdent.Obj, vr)
			}
		}
		var body []ir.MetaStmt
		for _, stmt := range cc.Body {
			m := walkStmt(stmt)
			body = append(body, m)
		}
		tscc.Body = body
		var typs []*types.Type
		for _, e := range cc.List {
			var typ *types.Type
			if !isNilIdent(e) {
				typ = E2T(e)
				RegisterDtype(typ, GetTypeOfExpr(typeSwitch.Subject))
			}
			typs = append(typs, typ) // universe nil can be appended
		}
		tscc.Types = typs
		if assignIdent != nil {
			SetVariable(assignIdent.Obj, nil)
		}
	}
	typeSwitch.Cases = cases

	return typeSwitch
}
func isNilIdent(e ast.Expr) bool {
	ident, ok := e.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Obj == universe.Nil
}

func walkCaseClause(s *ast.CaseClause) *ir.MetaCaseClause {
	var listMeta []ir.MetaExpr
	for _, e := range s.List {
		m := walkExpr(e, nil)
		listMeta = append(listMeta, m)
	}
	var body []ir.MetaStmt
	for _, stmt := range s.Body {
		metaStmt := walkStmt(stmt)
		body = append(body, metaStmt)
	}
	return &ir.MetaCaseClause{
		Tpos:     s.Pos(),
		ListMeta: listMeta,
		Body:     body,
	}
}

func walkBranchStmt(s *ast.BranchStmt) *ir.MetaBranchStmt {
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

	return &ir.MetaBranchStmt{
		Tpos:             s.Pos(),
		ContainerForStmt: currentFor,
		ContinueOrBreak:  continueOrBreak,
	}
}

func walkGoStmt(s *ast.GoStmt) *ir.MetaGoStmt {
	fun := walkExpr(s.Call.Fun, nil)
	return &ir.MetaGoStmt{
		Tpos: s.Pos(),
		Fun:  fun,
	}
}

func walkDeferStmt(s *ast.DeferStmt) *ir.MetaDeferStmt {
	funcDef := currentFunc
	fun := walkExpr(s.Call.Fun, nil)
	deferVar := registerLocalVariable(funcDef, ".defer.var", types.Uintptr)
	funcDef.HasDefer = true
	funcDef.DeferVar = deferVar
	singleAssing := &ir.MetaSingleAssign{
		Tpos: s.Pos(),
		Lhs:  deferVar,
		Rhs:  fun,
	}
	return &ir.MetaDeferStmt{
		Tpos:       s.Pos(),
		Fun:        fun,
		FuncAssign: singleAssing,
	}
}

func walkStmt(stmt ast.Stmt) ir.MetaStmt {
	var mt ir.MetaStmt
	assert(stmt.Pos() != 0, "stmt.Pos() should not be zero", __func__)
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		mt = walkBlockStmt(s)
	case *ast.ExprStmt:
		mt = walkExprStmt(s)
	case *ast.DeclStmt:
		mt = walkDeclStmt(s)
	case *ast.AssignStmt:
		mt = walkAssignStmt(s)
	case *ast.IncDecStmt:
		mt = walkIncDecStmt(s)
	case *ast.ReturnStmt:
		mt = walkReturnStmt(s)
	case *ast.IfStmt:
		mt = walkIfStmt(s)
	case *ast.ForStmt:
		mt = walkForStmt(s)
	case *ast.RangeStmt:
		mt = walkRangeStmt(s)
	case *ast.BranchStmt:
		mt = walkBranchStmt(s)
	case *ast.SwitchStmt:
		mt = walkSwitchStmt(s)
	case *ast.TypeSwitchStmt:
		mt = walkTypeSwitchStmt(s)
	case *ast.GoStmt:
		mt = walkGoStmt(s)
	case *ast.DeferStmt:
		mt = walkDeferStmt(s)
	default:
		throw(stmt)
	}

	assert(mt != nil, "meta should not be nil", __func__)
	assert(mt.Pos() != 0, "mt.Pos() should not be zero", __func__)
	return mt
}

func isUniverseNil(m *ir.MetaIdent) bool {
	return m.Kind == "nil"
}

func WalkIdent(e *ast.Ident, ctx *ir.EvalContext) *ir.MetaIdent {
	meta := &ir.MetaIdent{
		Tpos: e.Pos(),
		Name: e.Name,
	}
	logfncname := "(toplevel)"
	if currentFunc != nil {
		logfncname = currentFunc.Name
	}
	_ = logfncname
	if e.Name == "_" {
		// blank identifier
		// e.Obj is nil in this case.
		// @TODO do something
		meta.Kind = "blank"
		meta.Type = nil
		return meta
	}
	assert(e.Obj != nil, CurrentPkg.Name+" ident.Obj should not be nil:"+e.Name, __func__)
	switch e.Obj {
	case universe.Nil:
		assert(ctx != nil, "ctx of nil is not passed", __func__)
		assert(ctx.Type != nil, "ctx.Type of nil is not passed", __func__)
		meta.Type = ctx.Type
		meta.Kind = "nil"
	case universe.True:
		meta.Kind = "true"
		meta.Type = types.Bool
	case universe.False:
		meta.Kind = "false"
		meta.Type = types.Bool
	default:
		switch e.Obj.Kind {
		case ast.Var:
			meta.Kind = "var"
			if e.Obj.Data == nil {
				panic("ident.Obj.Data should not be nil: name=" + meta.Name)
			}
			meta.Variable = e.Obj.Data.(*ir.Variable)
			meta.Type = meta.Variable.Type
		case ast.Con:
			meta.Kind = "con"
			if e.Obj.Data == nil {
				panic("ident.Obj.Data should not be nil: name=" + meta.Name)
			}
			cnst := e.Obj.Data.(*ir.Const)
			meta.Type = cnst.Type
			meta.Const = cnst
		case ast.Fun:
			meta.Kind = "fun"
			switch e.Obj {
			case universe.Len, universe.Cap, universe.New, universe.Make, universe.Append, universe.Panic, universe.Delete:
				// builtin funcs have no func type
			default:
				meta.Type = E2T(e.Obj.Decl.(*ast.FuncDecl).Type)
			}
		case ast.Typ:
			// this can happen when walking type nodes intentionally
			meta.Kind = "typ"
			meta.Type = E2T(e)
		default: // ast.Pkg
			panic("Unexpected ident Kind:" + e.Obj.Kind.String() + " name:" + e.Name)
		}

	}
	return meta
}

func walkSelectorExpr(e *ast.SelectorExpr, ctx *ir.EvalContext) *ir.MetaSelectorExpr {
	meta := &ir.MetaSelectorExpr{
		Tpos: e.Pos(),
	}
	if isQI(e) {
		meta.IsQI = true
		// pkg.ident
		qi := Selector2QI(e)
		meta.QI = qi
		ei := LookupForeignIdent(qi, e.Pos())
		if ei.Func != nil { // fun
			foreignMeta := &ir.MetaForeignFuncWrapper{
				Tpos: e.Pos(),
				QI:   qi,
			}
			meta.ForeignValue = foreignMeta
			meta.Type = E2T(ei.Func.Decl.Type)
		} else { // var|con
			meta.ForeignValue = ei.MetaIdent
			meta.Type = ei.Type
		}
	} else {
		// expr.field
		meta.X = walkExpr(e.X, ctx)
		typ, field, offset, needDeref := getTypeOfSelector(meta.X, e)
		meta.Type = typ
		if field != nil {
			// struct.field
			meta.Field = field
			meta.Offset = offset
			meta.NeedDeref = needDeref
		}

	}
	meta.SelName = e.Sel.Name
	return meta
}

func getTypeOfSelector(x ir.MetaExpr, e *ast.SelectorExpr) (*types.Type, *ast.Field, int, bool) {
	// (strct).field | (ptr).field | (obj).method
	var needDeref bool
	typeOfLeft := GetTypeOfExpr(x)
	utLeft := GetUnderlyingType(typeOfLeft)
	var structTypeLiteral *ast.StructType
	switch typ := utLeft.E.(type) {
	case *ast.StructType: // strct.field
		structTypeLiteral = typ
	case *ast.StarExpr: // ptr.field
		needDeref = true
		origType := E2T(typ.X)
		if Kind(origType) == types.T_STRUCT {
			structTypeLiteral = GetUnderlyingStructType(origType)
		} else {
			_, isIdent := typ.X.(*ast.Ident)
			if isIdent {
				typeOfLeft = origType
				method := LookupMethod(typeOfLeft.GoType, e.Sel.Name)
				funcType := method.FuncType
				if funcType.Results == nil || len(funcType.Results.List) == 0 {
					return nil, nil, 0, needDeref
				}
				types := FieldList2Types(funcType.Results)
				return types[0], nil, 0, needDeref
			}
		}
	default: // obj.method
		method := LookupMethod(typeOfLeft.GoType, e.Sel.Name)
		funcType := method.FuncType
		if funcType == nil {
			panic("funcType should not be nil:" + method.Name)
		}
		if funcType.Results == nil || len(funcType.Results.List) == 0 {
			return nil, nil, 0, false
		}
		types := FieldList2Types(funcType.Results)
		return types[0], nil, 0, false
	}

	field := LookupStructField(structTypeLiteral, e.Sel.Name)
	if field != nil {
		offset := GetStructFieldOffset(field)
		return E2T(field.Type), field, offset, needDeref
	}
	if field == nil { // try to find method
		method := LookupMethod(typeOfLeft.GoType, e.Sel.Name)
		funcType := method.FuncType
		if funcType.Results == nil || len(funcType.Results.List) == 0 {
			return nil, nil, 0, needDeref
		}
		types := FieldList2Types(funcType.Results)
		return types[0], field, 0, needDeref
	}

	panic("Bad type")
}

func walkConversion(pos token.Pos, toType *types.Type, arg0 ir.MetaExpr) ir.MetaExpr {

	meta := &ir.MetaConversionExpr{
		Tpos: pos,
		Type: toType,
		Arg0: arg0,
	}
	fromType := GetTypeOfExpr(arg0)
	fromKind := Kind(fromType)
	toKind := Kind(toType)
	if toKind == types.T_INTERFACE && fromKind != types.T_INTERFACE {
		RegisterDtype(fromType, toType)
	}
	return meta
}

func walkCallExpr(e *ast.CallExpr, ctx *ir.EvalContext) ir.MetaExpr {
	if isType(e.Fun) {
		assert(len(e.Args) == 1, "convert must take only 1 argument", __func__)
		toType := E2T(e.Fun)
		ctx := &ir.EvalContext{
			Type: toType,
		}
		arg0 := walkExpr(e.Args[0], ctx)
		return walkConversion(e.Pos(), toType, arg0)
	}

	meta := &ir.MetaCallExpr{
		Tpos: e.Pos(),
	}

	meta.HasEllipsis = e.Ellipsis != token.NoPos

	// function call
	metaFun := walkExpr(e.Fun, nil)

	// Replace __func__ ident by a string literal
	//for i, arg := range meta.args {
	//	ident, ok := arg.(*ast.Ident)
	//	if ok {
	//		if ident.Name == "__func__" && ident.Obj.Kind == ast.Var {
	//			basicLit := &ast.BasicLit{
	//				Kind:  token.STRING,
	//				RawValue: "\"" + currentFunc.Name + "\"",
	//			}
	//			arg = basicLit
	//			e.Args[i] = arg
	//		}
	//	}
	//
	//}

	identFun, isIdent := e.Fun.(*ast.Ident)
	if isIdent {
		switch identFun.Obj {
		case universe.Len:
			a0 := walkExpr(e.Args[0], nil)
			return &ir.MetaCallLen{
				Tpos: e.Pos(),
				Type: types.Int,
				Arg0: a0,
			}
		case universe.Cap:
			a0 := walkExpr(e.Args[0], nil)
			return &ir.MetaCallCap{
				Tpos: e.Pos(),
				Type: types.Int,
				Arg0: a0,
			}
		case universe.New:
			walkExpr(e.Args[0], nil) // Do we need this ?
			typeArg0 := E2T(e.Args[0])
			ptrType := &ast.StarExpr{
				X:    e.Args[0],
				Star: 1,
			}
			typ := E2T(ptrType)
			return &ir.MetaCallNew{
				Tpos:     e.Pos(),
				Type:     typ,
				TypeArg0: typeArg0,
			}
		case universe.Make:
			walkExpr(e.Args[0], nil) // Do we need this ?
			typeArg0 := E2T(e.Args[0])
			typ := typeArg0
			ctx := &ir.EvalContext{Type: types.Int}
			var a1 ir.MetaExpr
			var a2 ir.MetaExpr
			if len(e.Args) > 1 {
				a1 = walkExpr(e.Args[1], ctx)
			}

			if len(e.Args) > 2 {
				a2 = walkExpr(e.Args[2], ctx)
			}
			return &ir.MetaCallMake{
				Tpos:     e.Pos(),
				Type:     typ,
				TypeArg0: typeArg0,
				Arg1:     a1,
				Arg2:     a2,
			}
		case universe.Append:
			a0 := walkExpr(e.Args[0], nil)
			a1 := walkExpr(e.Args[1], nil)
			typ := GetTypeOfExpr(a0)
			return &ir.MetaCallAppend{
				Tpos: e.Pos(),
				Type: typ,
				Arg0: a0,
				Arg1: a1,
			}
		case universe.Panic:
			a0 := walkExpr(e.Args[0], nil)
			return &ir.MetaCallPanic{
				Tpos: e.Pos(),
				Type: nil,
				Arg0: a0,
			}
		case universe.Delete:
			a0 := walkExpr(e.Args[0], nil)
			a1 := walkExpr(e.Args[1], nil)
			return &ir.MetaCallDelete{
				Tpos: e.Pos(),
				Type: nil,
				Arg0: a0,
				Arg1: a1,
			}
		}
	}

	var funcType *ast.FuncType
	var funcVal *ir.FuncValue
	var receiver ast.Expr
	var receiverMeta ir.MetaExpr
	switch fn := e.Fun.(type) {
	case *ast.Ident:
		// general function call
		symbol := GetPackageSymbol(CurrentPkg.Name, fn.Name)
		switch CurrentPkg.Name {
		case "os":
			switch fn.Name {
			case "runtime_args":
				symbol = GetPackageSymbol("runtime", "runtime_args")
			}
		case "runtime":
			if fn.Name == "makeSlice1" || fn.Name == "makeSlice8" || fn.Name == "makeSlice16" || fn.Name == "makeSlice24" {
				fn.Name = "makeSlice"
				symbol = GetPackageSymbol("runtime", fn.Name)
			}
		}
		funcVal = NewFuncValueFromSymbol(symbol)
		switch dcl := fn.Obj.Decl.(type) {
		case *ast.FuncDecl:
			funcType = dcl.Type
		case *ast.ValueSpec: // var f func()
			funcType = dcl.Type.(*ast.FuncType)
			funcVal = &ir.FuncValue{
				Expr: metaFun,
			}
		case *ast.AssignStmt: // f := staticF
			assert(fn.Obj.Data != nil, "funcvalue should be a variable:"+fn.Name, __func__)
			rhs := dcl.Rhs[0]
			switch r := rhs.(type) {
			case *ast.SelectorExpr:
				assert(isQI(r), "expect QI", __func__)
				qi := Selector2QI(r)
				ff := LookupForeignFunc(qi)
				funcType = ff.Decl.Type
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
			qi := Selector2QI(fn)
			funcVal = NewFuncValueFromSymbol(string(qi))
			ff := LookupForeignFunc(qi)
			funcType = ff.Decl.Type
		} else {
			// method call
			receiver = fn.X
			receiverMeta = walkExpr(fn.X, nil)
			receiverType := GetTypeOfExpr(receiverMeta)
			method := LookupMethod(receiverType.GoType, fn.Sel.Name)
			funcType = method.FuncType
			funcVal = NewFuncValueFromSymbol(GetMethodSymbol(method))
			funcVal.MethodName = fn.Sel.Name
			if Kind(receiverType) == types.T_POINTER {
				if method.IsPtrMethod {
					// p.mp() => as it is
				} else {
					// p.mv()
					panic("TBI 4190")
				}
			} else if Kind(receiverType) == types.T_INTERFACE {
				funcVal.IfcMethodCal = true
				funcVal.IfcType = receiverType
				funcVal.IsDirect = false
			} else {
				if method.IsPtrMethod {
					// v.mp() => (&v).mp()
					// @TODO we should check addressable
					rcvr := &ast.UnaryExpr{
						Op:    token.AND,
						X:     receiver,
						OpPos: 1,
					}
					eTyp := &ast.StarExpr{
						X:    receiverType.E,
						Star: 1,
					}
					receiverMeta = &ir.MetaUnaryExpr{
						X:    receiverMeta,
						Type: E2T(eTyp),
						Op:   rcvr.Op.String(),
					}
				} else {
					// v.mv() => as it is
				}
			}
		}
	default:
		throw(e.Fun)
	}

	meta.Types = FieldList2Types(funcType.Results)
	if len(meta.Types) > 0 {
		meta.Type = meta.Types[0]
	}
	meta.FuncVal = funcVal
	argsAndParams := prepareArgsAndParams(funcType, receiverMeta, e.Args, meta.HasEllipsis)
	var paramTypes []*types.Type
	var args []ir.MetaExpr
	for _, a := range argsAndParams {
		paramTypes = append(paramTypes, a.ParamType)
		arg := CheckIfcConversion(a.Meta.Pos(), a.Meta, a.ParamType)
		args = append(args, arg)
	}

	meta.ParamTypes = paramTypes
	meta.Args = args
	return meta
}

func walkBasicLit(e *ast.BasicLit, ctx *ir.EvalContext) *ir.MetaBasicLit {
	m := &ir.MetaBasicLit{
		Tpos:     e.Pos(),
		Kind:     e.Kind.String(),
		RawValue: e.Value,
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
		m.CharVal = int(char)
		m.Type = types.Int32 // @TODO: This is not correct
	case "INT":
		ival, err := strconv.ParseInt(m.RawValue, 0, 64)
		if err != nil {
			panic("strconv.ParseInt failed")
		}
		m.IntVal = int(ival)
		m.Type = types.Int // @TODO: This is not correct
	case "STRING":
		m.StrVal = registerStringLiteral(e)
		m.Type = types.String // @TODO: This is not correct
	default:
		panic("Unexpected literal Kind:" + e.Kind.String())
	}
	return m
}

func walkCompositeLit(e *ast.CompositeLit, ctx *ir.EvalContext) *ir.MetaCompositLit {
	//walkExpr(e.Type, nil) // a[len("foo")]{...} // "foo" should be walked
	typ := E2T(e.Type)
	ut := GetUnderlyingType(typ)
	var knd string
	switch Kind(ut) {
	case types.T_STRUCT:
		knd = "struct"
	case types.T_ARRAY:
		knd = "array"
	case types.T_SLICE:
		knd = "slice"
	default:
		unexpectedKind(Kind(typ))
	}
	meta := &ir.MetaCompositLit{
		Tpos: e.Pos(),
		Kind: knd,
		Type: typ,
	}

	switch Kind(ut) {
	case types.T_STRUCT:
		structType := meta.Type
		var metaElms []*ir.MetaStructLiteralElement
		for _, elm := range e.Elts {
			kvExpr := elm.(*ast.KeyValueExpr)
			fieldName := kvExpr.Key.(*ast.Ident)
			field := LookupStructField(GetUnderlyingStructType(structType), fieldName.Name)
			fieldType := E2T(field.Type)
			ctx := &ir.EvalContext{Type: fieldType}
			// attach type to nil : STRUCT{Key:nil}
			valueMeta := walkExpr(kvExpr.Value, ctx)
			mc := CheckIfcConversion(kvExpr.Pos(), valueMeta, fieldType)
			metaElm := &ir.MetaStructLiteralElement{
				Tpos:      kvExpr.Pos(),
				Field:     field,
				FieldType: fieldType,
				Value:     mc,
			}

			metaElms = append(metaElms, metaElm)
		}
		meta.StructElements = metaElms
	case types.T_ARRAY:
		arrayType := ut.E.(*ast.ArrayType)
		meta.Len = EvalInt(arrayType.Len)
		meta.ElmType = E2T(arrayType.Elt)
		ctx := &ir.EvalContext{Type: meta.ElmType}
		var ms []ir.MetaExpr
		for _, v := range e.Elts {
			m := walkExpr(v, ctx)
			mc := CheckIfcConversion(v.Pos(), m, meta.ElmType)
			ms = append(ms, mc)
		}
		meta.Elms = ms
	case types.T_SLICE:
		arrayType := ut.E.(*ast.ArrayType)
		meta.Len = len(e.Elts)
		meta.ElmType = E2T(arrayType.Elt)
		ctx := &ir.EvalContext{Type: meta.ElmType}
		var ms []ir.MetaExpr
		for _, v := range e.Elts {
			m := walkExpr(v, ctx)
			mc := CheckIfcConversion(v.Pos(), m, meta.ElmType)
			ms = append(ms, mc)
		}
		meta.Elms = ms
	}
	return meta
}

func walkUnaryExpr(e *ast.UnaryExpr, ctx *ir.EvalContext) *ir.MetaUnaryExpr {
	meta := &ir.MetaUnaryExpr{
		Tpos: e.Pos(),
		Op:   e.Op.String(),
	}
	meta.X = walkExpr(e.X, nil)
	switch meta.Op {
	case "+", "-":
		meta.Type = GetTypeOfExpr(meta.X)
	case "!":
		meta.Type = types.Bool
	case "&":
		xTyp := GetTypeOfExpr(meta.X)
		ptrType := &ast.StarExpr{
			Star: e.Pos(),
			X:    xTyp.E,
		}
		meta.Type = E2T(ptrType)
	}

	return meta
}

func walkBinaryExpr(e *ast.BinaryExpr, ctx *ir.EvalContext) *ir.MetaBinaryExpr {
	meta := &ir.MetaBinaryExpr{
		Tpos: e.Pos(),
		Op:   e.Op.String(),
	}
	if isNilIdent(e.X) {
		// Y should be typed
		meta.Y = walkExpr(e.Y, nil) // right
		xCtx := &ir.EvalContext{Type: GetTypeOfExpr(meta.Y)}

		meta.X = walkExpr(e.X, xCtx) // left
	} else {
		// X should be typed
		meta.X = walkExpr(e.X, nil) // left
		xTyp := GetTypeOfExpr(meta.X)
		if xTyp == nil {
			panicPos("xTyp should not be nil", e.Pos())
		}
		yCtx := &ir.EvalContext{Type: xTyp}
		meta.Y = walkExpr(e.Y, yCtx) // right
	}
	switch meta.Op {
	case "==", "!=", "<", ">", "<=", ">=":
		meta.Type = types.Bool
	default:
		// @TODO type of (1 + x) should be type of x
		if isNilIdent(e.X) {
			meta.Type = GetTypeOfExpr(meta.Y)
		} else {
			meta.Type = GetTypeOfExpr(meta.X)
		}
	}
	return meta
}

func walkIndexExpr(e *ast.IndexExpr, ctx *ir.EvalContext) *ir.MetaIndexExpr {
	meta := &ir.MetaIndexExpr{
		Tpos: e.Pos(),
	}
	meta.Index = walkExpr(e.Index, nil) // @TODO pass context for map,slice,array
	meta.X = walkExpr(e.X, nil)
	collectionTyp := GetTypeOfExpr(meta.X)
	if Kind(collectionTyp) == types.T_MAP {
		meta.IsMap = true
		if ctx != nil && ctx.MaybeOK {
			meta.NeedsOK = true
		}
	}

	meta.Type = GetElementTypeOfCollectionType(collectionTyp)
	return meta
}

func walkSliceExpr(e *ast.SliceExpr, ctx *ir.EvalContext) *ir.MetaSliceExpr {
	meta := &ir.MetaSliceExpr{
		Tpos: e.Pos(),
	}

	// For convenience, any of the indices may be omitted.

	// A missing low index defaults to zero;
	if e.Low != nil {
		meta.Low = walkExpr(e.Low, nil)
	} else {
		eZeroInt := &ast.BasicLit{
			Value:    "0",
			Kind:     token.INT,
			ValuePos: 1,
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
	listType := GetTypeOfExpr(meta.X)
	if Kind(listType) == types.T_STRING {
		// str2 = str1[n:m]
		meta.Type = types.String
	} else {
		elmType := GetElementTypeOfCollectionType(listType)
		r := &ast.ArrayType{
			Len:    nil, // slice
			Elt:    elmType.E,
			Lbrack: e.Pos(),
		}
		meta.Type = E2T(r)
	}
	return meta
}

func walkStarExpr(e *ast.StarExpr, ctx *ir.EvalContext) *ir.MetaStarExpr {
	meta := &ir.MetaStarExpr{
		Tpos: e.Pos(),
	}
	meta.X = walkExpr(e.X, nil)
	xType := GetTypeOfExpr(meta.X)
	origType := xType.E.(*ast.StarExpr)
	meta.Type = E2T(origType.X)
	return meta
}

func walkTypeAssertExpr(e *ast.TypeAssertExpr, ctx *ir.EvalContext) *ir.MetaTypeAssertExpr {
	meta := &ir.MetaTypeAssertExpr{
		Tpos: e.Pos(),
	}
	if ctx != nil && ctx.MaybeOK {
		meta.NeedsOK = true
	}
	meta.X = walkExpr(e.X, nil)
	meta.Type = E2T(e.Type)
	if meta.Type.GoType == nil {
		panic(fmt.Sprintf("[walkTypeAssertExpr] GoType is not set:%T\n", e.Type))
	}

	RegisterDtype(meta.Type, GetTypeOfExpr(meta.X))
	return meta
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
func walkExpr(expr ast.Expr, ctx *ir.EvalContext) ir.MetaExpr {
	assert(expr.Pos() != 0, "e.Pos() should not be zero", __func__)
	switch e := expr.(type) {
	case *ast.BasicLit:
		return walkBasicLit(e, ctx)
	case *ast.CompositeLit:
		return walkCompositeLit(e, ctx)
	case *ast.Ident:
		return WalkIdent(e, ctx)
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
	case *ast.ParenExpr:
		return walkExpr(e.X, ctx)
	// Each one below is not an expr but a type
	case *ast.ArrayType: // type
		return nil
	case *ast.MapType: // type
		return nil
	case *ast.InterfaceType: // type
		return nil
	case *ast.FuncType:
		return nil // @TODO walk
	default:
		panic(fmt.Sprintf("unknown type %T", expr))
	}
}

func CheckIfcConversion(pos token.Pos, expr ir.MetaExpr, trgtType *types.Type) ir.MetaExpr {
	if IsNil(expr) {
		return expr
	}
	if trgtType == nil {
		return expr
	}
	if !IsInterface(trgtType.GoType) {
		return expr
	}
	fromType := GetTypeOfExpr(expr)
	if IsInterface(fromType.GoType) {
		return expr
	}

	RegisterDtype(fromType, trgtType)

	return &ir.IfcConversion{
		Tpos:  pos,
		Value: expr,
		Type:  trgtType,
	}
}

func LookupForeignIdent(qi ir.QualifiedIdent, pos token.Pos) *ir.ExportedIdent {
	ei, ok := exportedIdents[string(qi)]
	if !ok {
		panicPos(string(qi)+" Not found in exportedIdents", pos)
	}
	return ei
}

func LookupForeignFunc(qi ir.QualifiedIdent) *ir.Func {
	ei := LookupForeignIdent(qi, 1)
	return ei.Func
}

func SetVariable(obj *ast.Object, vr *ir.Variable) {
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
// - determine Types of variable declarations
// - attach type to every expression
// - transmit ok syntax context
// - (hope) attach type to untyped constants
// - (hope) transmit the need of interface conversion
func Walk(pkg *ir.PkgContainer) *ir.AnalyzedPackage {
	pkg.StringIndex = 0
	pkg.StringLiterals = nil
	CurrentPkg = pkg

	ITab = make(map[string]*ITabEntry)
	ITabID = 1

	var hasInitFunc bool
	var typs []*types.Type
	var funcs []*ir.Func
	var consts []*ir.PackageVarConst
	var vars []*ir.PackageVarConst

	var typeSpecs []*ast.TypeSpec
	var funcDecls []*ast.FuncDecl
	var varSpecs []*ast.ValueSpec
	var constSpecs []*ast.ValueSpec

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
		//@TODO check serializeType()'s *ast.Ident case
		typeSpec.Name.Obj.Data = pkg.Name // package the type belongs to
		eType := &ast.Ident{
			Name:    typeSpec.Name.Name,
			NamePos: typeSpec.Pos(),
			Obj: &ast.Object{
				Kind: ast.Typ,
				Decl: typeSpec,
				Data: pkg.Name,
			},
		}
		t := E2T(eType)
		gt := t.GoType.(*types.Named)
		gt.PkgName = pkg.Name
		typs = append(typs, t)
		switch Kind(t) {
		case types.T_STRUCT:
			//structType := GetUnderlyingType(t)
			st := t.GoType.Underlying().Underlying()
			calcStructSizeAndSetFieldOffset2(st.(*types.Struct))
			//			calcStructSizeAndSetFieldOffset(structType.E.(*ast.StructType))
		case types.T_INTERFACE:
			// register ifc method
			it := typeSpec.Type.(*ast.InterfaceType)
			if it.Methods != nil {
				for _, m := range it.Methods.List {
					funcType := m.Type.(*ast.FuncType)
					method := &ir.Method{
						PkgName:      pkg.Name,
						RcvNamedType: typeSpec.Name,
						Name:         m.Names[0].Name,
						FuncType:     funcType,
					}
					registerMethod(pkg.Name, method)
				}
			}
		}
		// named type
		ei := &ir.ExportedIdent{
			Name:    typeSpec.Name.Name,
			IsType:  true,
			Obj:     typeSpec.Name.Obj,
			Type:    t,
			PkgName: pkg.Name,
			Pos:     typeSpec.Pos(),
		}
		exportedIdents[string(NewQI(pkg.Name, typeSpec.Name.Name))] = ei
	}

	for _, spec := range constSpecs {
		assert(len(spec.Values) == 1, "only 1 value is supported", __func__)
		lhsIdent := spec.Names[0]
		rhs := spec.Values[0]
		var rhsMeta ir.MetaExpr
		var t *types.Type
		if spec.Type != nil { // const x T = e
			t = E2T(spec.Type)
			ctx := &ir.EvalContext{Type: t}
			rhsMeta = walkExpr(rhs, ctx)
		} else { // const x = e
			rhsMeta = walkExpr(rhs, nil)
			t = GetTypeOfExpr(rhsMeta)
			spec.Type = t.E
		}
		// treat package const as global var for now

		rhsLiteral, isLiteral := rhsMeta.(*ir.MetaBasicLit)
		if !isLiteral {
			panic("const decl value should be literal:" + lhsIdent.Name)
		}
		cnst := &ir.Const{
			Tpos:         lhsIdent.Pos(),
			Name:         lhsIdent.Name,
			IsGlobal:     true,
			GlobalSymbol: CurrentPkg.Name + "." + lhsIdent.Name,
			Literal:      rhsLiteral,
			Type:         t,
		}
		lhsIdent.Obj.Data = cnst
		metaVar := WalkIdent(lhsIdent, nil)
		pconst := &ir.PackageVarConst{
			Spec:    spec,
			Name:    lhsIdent,
			Val:     rhs,
			MetaVal: rhsMeta, // cannot be nil
			MetaVar: metaVar,
			Type:    t,
		}
		consts = append(consts, pconst)

		ei := &ir.ExportedIdent{
			Name:      lhsIdent.Name,
			PkgName:   pkg.Name,
			Pos:       lhsIdent.Pos(),
			MetaIdent: metaVar,
			Type:      t,
		}

		exportedIdents[string(NewQI(pkg.Name, lhsIdent.Name))] = ei
	}

	for _, spec := range varSpecs {
		lhsIdent := spec.Names[0]
		assert(lhsIdent.Obj.Kind == ast.Var, "should be Var", __func__)
		var rhsMeta ir.MetaExpr
		var t *types.Type
		if spec.Type != nil { // var x T = e
			// walkExpr(spec.Type, nil) // Do we need walk type ?s
			t = E2T(spec.Type)
			if len(spec.Values) > 0 {
				rhs := spec.Values[0]
				ctx := &ir.EvalContext{Type: t}
				rhsMeta = walkExpr(rhs, ctx)
				rhsMeta = CheckIfcConversion(rhs.Pos(), rhsMeta, t)
			}
		} else { // var x = e  infer lhs type from rhs
			if len(spec.Values) == 0 {
				panic("invalid syntax")
			}

			rhs := spec.Values[0]
			rhsMeta = walkExpr(rhs, nil)
			t = GetTypeOfExpr(rhsMeta)
			if t == nil {
				panic("variable type is not determined : " + lhsIdent.Name)
			}
		}
		spec.Type = t.E

		variable := newGlobalVariable(pkg.Name, lhsIdent.Obj.Name, t)
		SetVariable(lhsIdent.Obj, variable)
		metaVar := WalkIdent(lhsIdent, nil)

		var rhs ast.Expr
		if len(spec.Values) > 0 {
			rhs = spec.Values[0]
			// collect string literals
		}
		pkgVar := &ir.PackageVarConst{
			Spec:    spec,
			Name:    lhsIdent,
			Val:     rhs,
			MetaVal: rhsMeta, // can be nil
			MetaVar: metaVar,
			Type:    t,
		}
		vars = append(vars, pkgVar)
		ei := &ir.ExportedIdent{
			PkgName:   pkg.Name,
			Name:      lhsIdent.Name,
			Pos:       lhsIdent.Pos(),
			Type:      t,
			MetaIdent: metaVar,
		}
		exportedIdents[string(NewQI(pkg.Name, lhsIdent.Name))] = ei
	}

	// collect methods in advance
	for _, funcDecl := range funcDecls {
		if funcDecl.Recv != nil {
			// is method
			method := newMethod(pkg.Name, funcDecl)
			registerMethod(pkg.Name, method)
		}
	}

	for _, funcDecl := range funcDecls {
		fnc := &ir.Func{
			PkgName:   CurrentPkg.Name,
			Name:      funcDecl.Name.Name,
			Decl:      funcDecl,
			Signature: FuncTypeToSignature(funcDecl.Type),
			Localarea: 0,
			Argsarea:  16, // return address + previous rbp
		}
		currentFunc = fnc
		funcs = append(funcs, fnc)

		if funcDecl.Recv == nil {
			// non-method function
			if funcDecl.Name.Name == "init" {
				hasInitFunc = true
			}
			qi := NewQI(pkg.Name, funcDecl.Name.Name)
			ei := &ir.ExportedIdent{
				Name:    funcDecl.Name.Name,
				PkgName: pkg.Name,
				Pos:     funcDecl.Pos(),
				Func:    fnc,
			}
			exportedIdents[string(qi)] = ei
		}

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
			// param names can be ommitted.
			if len(field.Names) > 0 {
				obj := field.Names[0].Obj
				SetVariable(obj, registerParamVariable(fnc, obj.Name, E2T(field.Type)))
			}
		}

		for i, field := range resultFields {
			if len(field.Names) == 0 {
				// unnamed retval
				registerReturnVariable(fnc, ".r"+strconv.Itoa(i), E2T(field.Type))
			} else {
				panic("TBI: named return variable is not supported")
			}
		}

		if funcDecl.Body != nil {
			fnc.HasBody = true
			var ms []ir.MetaStmt
			for _, stmt := range funcDecl.Body.List {
				m := walkStmt(stmt)
				ms = append(ms, m)
			}
			fnc.Stmts = ms

			if funcDecl.Recv != nil { // is Method
				fnc.Method = newMethod(pkg.Name, funcDecl)
			}
		}
		currentFunc = nil
	}

	return &ir.AnalyzedPackage{
		Path:           pkg.Path,
		Name:           pkg.Name,
		Imports:        pkg.Imports,
		Types:          typs,
		Funcs:          funcs,
		Consts:         consts,
		Vars:           vars,
		HasInitFunc:    hasInitFunc,
		StringLiterals: pkg.StringLiterals,
		Fset:           pkg.Fset,
		FileNoMap:      pkg.FileNoMap,
	}
}

const SizeOfSlice int = 24
const SizeOfString int = 16
const SizeOfInt int = 8
const SizeOfUint8 int = 1
const SizeOfUint16 int = 2
const SizeOfPtr int = 8
const SizeOfInterface int = 16

func GetSizeOfType2(t types.GoType) int {
	t = t.Underlying()
	t = t.Underlying()
	switch Kind2(t) {
	case types.T_SLICE:
		return SizeOfSlice
	case types.T_STRING:
		return SizeOfString
	case types.T_INT:
		return SizeOfInt
	case types.T_UINTPTR, types.T_POINTER, types.T_MAP:
		return SizeOfPtr
	case types.T_UINT8:
		return SizeOfUint8
	case types.T_UINT16:
		return SizeOfUint16
	case types.T_BOOL:
		return SizeOfInt
	case types.T_INTERFACE:
		return SizeOfInterface
	case types.T_ARRAY:
		arrayType := t.(*types.Array)
		elmSize := GetSizeOfType2(arrayType.Elem())
		return elmSize * arrayType.Len()
	case types.T_STRUCT:
		return calcStructSizeAndSetFieldOffset2(t.(*types.Struct))
	case types.T_FUNC:
		return SizeOfPtr
	default:
		unexpectedKind(Kind2(t))
	}
	return 0
}

func GetSizeOfType(t *types.Type) int {
	return GetSizeOfType2(t.GoType)
}

func calcStructSizeAndSetFieldOffset2(structType *types.Struct) int {
	var offset int = 0
	for i, field := range structType.Fields {
		setStructFieldOffset(structType.AstFields[i], offset)
		size := GetSizeOfType2(field.Typ)
		offset += size
	}
	return offset
}

func GetStructFieldOffset(field *ast.Field) int {
	return mapFieldOffset[unsafe.Pointer(field)]
}

func setStructFieldOffset(field *ast.Field, offset int) {
	mapFieldOffset[unsafe.Pointer(field)] = offset
}

func EvalInt(expr ast.Expr) int {
	if expr == nil {
		panic("EvanInt: nil is not expected")
	}
	switch e := expr.(type) {
	case *ast.BasicLit:
		return strconv.Atoi(e.Value)
	}
	panic(fmt.Sprintf("Unknown type:%T", expr))
}

func SerializeType(goType types.GoType, showPkgPrefix bool, showOnlyForeignPrefix bool, currentPkgName string) string {
	switch g := goType.(type) {
	case *types.Basic:
		return g.Name()
	case *types.Named:
		if g.PkgName == "" && g.String() == "error" {
			return "error"
		}
		if showPkgPrefix {
			if showOnlyForeignPrefix {
				if g.PkgName == currentPkgName {
					return g.String()
				} else {
					return g.PkgName + "." + g.String()
				}
			} else {
				return g.PkgName + "." + g.String()
			}
		} else {
			return g.String()
		}
	case *types.Pointer:
		return "*" + SerializeType(g.Elem(), showPkgPrefix, showOnlyForeignPrefix, currentPkgName)
	case *types.Array:
		return "[" + strconv.Itoa(g.Len()) + "]" + SerializeType(g.Elem(), showPkgPrefix, showOnlyForeignPrefix, currentPkgName)
	case *types.Slice:
		if g.Elp {
			return "..." + SerializeType(g.Elem(), showPkgPrefix, showOnlyForeignPrefix, currentPkgName)
		} else {
			return "[]" + SerializeType(g.Elem(), showPkgPrefix, showOnlyForeignPrefix, currentPkgName)
		}
	case *types.Map:
		return "map[" + SerializeType(g.Key(), showPkgPrefix, showOnlyForeignPrefix, currentPkgName) + "]" + SerializeType(g.Elem(), showPkgPrefix, showOnlyForeignPrefix, currentPkgName)
	case *types.Func:
		return "func()"
	case *types.Struct:
		r := "struct{"
		if len(g.Fields) > 0 {
			for _, field := range g.Fields {
				name := field.Name
				typ := field.Typ
				r += fmt.Sprintf("%s %s; ", name, SerializeType(typ, showPkgPrefix, showOnlyForeignPrefix, currentPkgName))
			}
		}
		return r + "}"
	case *types.Interface:
		if len(g.Methods) == 0 {
			return "interface{}"
		}
		r := "interface{ "
		for _, m := range g.Methods {
			mdcl := RestoreMethodDecl(m, showPkgPrefix, showOnlyForeignPrefix, currentPkgName)
			r += mdcl + "; "
		}
		r += " }"
		return r
	}
	panic(fmt.Sprintf("@TBI: GoType=%T", goType))
	return ""
}

func FuncTypeToSignature(funcType *ast.FuncType) *ir.Signature {
	p := FieldList2Types(funcType.Params)
	r := FieldList2Types(funcType.Results)
	return &ir.Signature{
		ParamTypes:  p,
		ReturnTypes: r,
	}
}

func RestoreMethodDecl(m *types.Func, showPkgPrefix bool, showOnlyForeignPrefix bool, currentPkgName string) string {
	var p string
	var r string
	name := m.Name
	fun, ok := m.Typ.(*types.Func)
	if !ok {
		panic(fmt.Sprintf("[SerializeType] Invalid type:%T\n", m.Typ))
	}
	sig := fun.Typ.(*types.Signature)
	if sig.Params != nil && len(sig.Params.Types) > 0 {
		for _, t := range sig.Params.Types {
			//name := field.Names[0].Name
			if p != "" {
				p += ","
			}
			p += SerializeType(t, showPkgPrefix, showOnlyForeignPrefix, currentPkgName)
		}
	}

	if sig.Results != nil && len(sig.Results.Types) > 0 {
		for _, t := range sig.Results.Types {
			//name := field.Names[0].Name
			if r != "" {
				r += ","
			}
			r += SerializeType(t, showPkgPrefix, showOnlyForeignPrefix, currentPkgName)
		}
	}

	decl := fmt.Sprintf("%s(%s) (%s)", name, p, r)
	return decl
}

func RestoreFuncDecl(fnc *ir.Func, showPkgPrefix bool, showOnlyForeignPrefix bool, currentPkgName string) string {
	var p string
	var r string
	for _, t := range fnc.Signature.ParamTypes {
		if p != "" {
			p += ","
		}
		p += SerializeType(t.GoType, showPkgPrefix, showOnlyForeignPrefix, currentPkgName)
	}
	for _, t := range fnc.Signature.ReturnTypes {
		if r != "" {
			r += ","
		}
		r += SerializeType(t.GoType, showPkgPrefix, showOnlyForeignPrefix, currentPkgName)
	}
	var m string
	var star string
	if fnc.Method != nil {
		if fnc.Method.IsPtrMethod {
			star = "*"
		}
		m = fmt.Sprintf("(%s%s) ", star, fnc.Method.RcvNamedType.Name)
	}
	return fmt.Sprintf("func %s%s (%s) (%s)",
		m, fnc.Name, p, r)
}

func NewLenMapSignature(arg0 ir.MetaExpr) *ir.Signature {
	return &ir.Signature{
		ParamTypes:  []*types.Type{types.Int},
		ReturnTypes: []*types.Type{GetTypeOfExpr(arg0)},
	}
}

func NewAppendSignature(elmType *types.Type) *ir.Signature {
	return &ir.Signature{
		ParamTypes:  []*types.Type{types.GeneralSliceType, elmType},
		ReturnTypes: []*types.Type{types.GeneralSliceType},
	}
}

func NewDeleteSignature(arg0 ir.MetaExpr) *ir.Signature {
	return &ir.Signature{
		ParamTypes:  []*types.Type{GetTypeOfExpr(arg0), types.Eface},
		ReturnTypes: nil,
	}
}

func GetConstRawValue(cnstExpr ir.MetaExpr) string {
	switch v := cnstExpr.(type) {
	case *ir.MetaBasicLit:
		return v.RawValue
	case *ir.MetaIdent:
		return GetConstRawValue(v.Const)
	default:
		panic("TBI")
	}
}

var ITabID int
var ITab map[string]*ITabEntry

type ITabEntry struct {
	Id          int
	DSerialized string
	ISeralized  string
	Itype       *types.Type
	Dtype       *types.Type
	Label       string
}

// "**[1][]*int" => ".dtype.8"
func RegisterDtype(dtype *types.Type, itype *types.Type) {
	ds := SerializeType(dtype.GoType, true, false, "")
	is := SerializeType(itype.GoType, true, false, "")

	key := ds + "-" + is
	_, ok := ITab[key]
	if ok {
		return
	}

	id := ITabID
	e := &ITabEntry{
		Id:          id,
		DSerialized: ds,
		ISeralized:  is,
		Itype:       itype,
		Dtype:       dtype,
		Label:       "." + "itab_" + strconv.Itoa(id),
	}
	ITab[key] = e
	ITabID++
}

func GetITabEntry(d types.GoType, i types.GoType) *ITabEntry {
	ds := SerializeType(d, true, false, "")
	is := SerializeType(i, true, false, "")
	key := ds + "-" + is
	ent, ok := ITab[key]
	if !ok {
		panic("dtype is not set:" + key)
	}
	return ent
}

func GetInterfaceMethods(iType types.GoType) []*types.Func {
	ut := iType.Underlying()
	it, ok := ut.(*types.Interface)
	if !ok {
		panic("not interface type")
	}
	if len(it.Methods) == 0 {
		return nil
	}
	return it.Methods
}
