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
		elp := param.Type.(*ast.Ellipsis)
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

func NewNumberLiteral(x int) *ir.MetaBasicLit {
	return &ir.MetaBasicLit{
		Pos:    1,
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

// Types of an expr in Single value context
func GetTypeOfExpr(meta ir.MetaExpr) *types.Type {
	switch m := meta.(type) {
	case *ir.MetaBasicLit:
		return m.Type
	case *ir.MetaCompositLit:
		return m.Type
	case *ir.MetaIdent:
		return m.Type
	case *ir.MetaSelectorExpr:
		return m.Type
	case *ir.MetaConversionExpr:
		return m.Type
	case *ir.MetaCallLen:
		return m.Type
	case *ir.MetaCallCap:
		return m.Type
	case *ir.MetaCallNew:
		return m.Type
	case *ir.MetaCallMake:
		return m.Type
	case *ir.MetaCallAppend:
		return m.Type
	case *ir.MetaCallPanic:
		return m.Type
	case *ir.MetaCallDelete:
		return m.Type
	case *ir.MetaCallExpr: // funcall
		return m.Type // can be nil (e.g. panic()). if Tuple , m.Types has Types
	case *ir.MetaIndexExpr:
		return m.Type
	case *ir.MetaSliceExpr:
		return m.Type
	case *ir.MetaStarExpr:
		return m.Type
	case *ir.MetaUnaryExpr:
		return m.Type
	case *ir.MetaBinaryExpr:
		return m.Type
	case *ir.MetaTypeAssertExpr:
		return m.Type
	}
	panic(fmt.Sprintf("bad type:%T\n", meta))
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
	return &types.Type{
		E: typeExpr,
	}
}

func GetArrayLen(t *types.Type) int {
	arrayType := t.E.(*ast.ArrayType)
	return EvalInt(arrayType.Len)
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
		case universe.Error:
			return E2T(&ast.InterfaceType{
				Methods:   nil, //  @FIXME
				Interface: 1,
			})
		}
		if e.Obj.Decl == nil {
			//logf("universe.Int=%d\n", uintptr(unsafe.Pointer(universe.Int)))
			//logf("e.Obj=%d\n", uintptr(unsafe.Pointer(e.Obj)))
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

func Kind(t *types.Type) types.TypeKind {
	if t == nil {
		panic("nil type is not expected")
	}

	ut := GetUnderlyingType(t)
	if ut == types.GeneralSliceType {
		return types.T_SLICE
	}

	switch e := ut.E.(type) {
	case *ast.Ident:
		assert(e.Obj.Kind == ast.Typ, "should be ast.Typ", __func__)
		switch e.Obj {
		case universe.Uintptr:
			return types.T_UINTPTR
		case universe.Int:
			return types.T_INT
		case universe.Int32:
			return types.T_INT32
		case universe.String:
			return types.T_STRING
		case universe.Uint8:
			return types.T_UINT8
		case universe.Uint16:
			return types.T_UINT16
		case universe.Bool:
			return types.T_BOOL
		case universe.Error:
			return types.T_INTERFACE
		default:
			panic("Unexpected type")
		}
	case *ast.StructType:
		return types.T_STRUCT
	case *ast.ArrayType:
		if e.Len == nil {
			return types.T_SLICE
		} else {
			return types.T_ARRAY
		}
	case *ast.StarExpr:
		return types.T_POINTER
	case *ast.Ellipsis: // x ...T
		return types.T_SLICE // @TODO is this right ?
	case *ast.MapType:
		return types.T_MAP
	case *ast.InterfaceType:
		return types.T_INTERFACE
	case *ast.FuncType:
		return types.T_FUNC
	}
	panic("should not reach here")
}

func IsInterface(t *types.Type) bool {
	return Kind(t) == types.T_INTERFACE
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

	//fmt.Fprintf(os.Stderr, "[register string] %s index=%d\n", CurrentPkg.Name, CurrentPkg.StringIndex)
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
		Typ:          t,
	}
}

func newLocalVariable(name string, localoffset int, t *types.Type) *ir.Variable {
	return &ir.Variable{
		Name:        name,
		IsGlobal:    false,
		LocalOffset: localoffset,
		Typ:         t,
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
	assert(ident.Obj != nil, "ident.Obj is nil:"+ident.Name, __func__)
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
// @TODO map key should be a QI ?
var MethodSets = make(map[unsafe.Pointer]*ir.NamedType)

func registerMethod(method *ir.Method) {
	key := unsafe.Pointer(method.RcvNamedType.Obj)
	namedType, ok := MethodSets[key]
	if !ok {
		namedType = &ir.NamedType{
			MethodSet: make(map[string]*ir.Method),
		}
		MethodSets[key] = namedType
	}
	//	util.Logf("registerMethod: type=%s name=%s\n", method.RcvNamedType.Name, method.Name)
	namedType.MethodSet[method.Name] = method
}

func lookupMethod(rcvT *types.Type, methodName *ast.Ident) *ir.Method {
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
		ei := LookupForeignIdent(Selector2QI(typ), methodName.Pos())
		typeObj = ei.Ident.Obj
	default:
		panic(rcvType)
	}

	namedType, ok := MethodSets[unsafe.Pointer(typeObj)]
	if !ok {
		panicPos(typeObj.Name+" has no methodSet", methodName.Pos())
	}
	method, ok := namedType.MethodSet[methodName.Name]
	if !ok {
		panicPos("method not found: "+methodName.Name, methodName.Pos())
	}
	return method
}

func walkExprStmt(s *ast.ExprStmt) *ir.MetaExprStmt {
	m := walkExpr(s.X, nil)
	return &ir.MetaExprStmt{
		X:   m,
		Pos: s.Pos(),
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
			Pos: lhsIdent.Pos(),
			Lhs: lhsMeta,
			Rhs: rhsMeta,
		}
		return &ir.MetaVarDecl{
			Pos:     lhsIdent.Pos(),
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
			if !IsBlankIdentifierMeta(lhsMetas[0]) {
				ctx = &ir.EvalContext{
					Type: GetTypeOfExpr(lhsMetas[0]),
				}
			}
			rhsMeta := walkExpr(s.Rhs[0], ctx)
			return &ir.MetaSingleAssign{
				Pos: pos,
				Lhs: lhsMetas[0],
				Rhs: rhsMeta,
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
				Pos:      pos,
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
				Pos: pos,
				Lhs: lhsMetas[0],
				Rhs: rhsMeta,
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
				Pos:      pos,
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
			Pos: pos,
			Lhs: lhsMeta,
			Rhs: rhsMeta,
		}
	default:
		panic("TBI 3497 ")
	}
	return nil
}

func walkReturnStmt(s *ast.ReturnStmt) *ir.MetaReturnStmt {
	funcDef := currentFunc
	if len(funcDef.Retvars) != len(s.Results) {
		panic("length of return and func type do not match")
	}

	_len := len(funcDef.Retvars)
	var results []ir.MetaExpr
	for i := 0; i < _len; i++ {
		expr := s.Results[i]
		retTyp := funcDef.Retvars[i].Typ
		ctx := &ir.EvalContext{
			Type: retTyp,
		}
		m := walkExpr(expr, ctx)
		results = append(results, m)
	}
	return &ir.MetaReturnStmt{
		Pos:     s.Pos(),
		Fnc:     funcDef,
		Results: results,
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
		Pos:  s.Pos(),
		Init: mInit,
		Cond: condMeta,
		Body: mtBlock,
		Else: mElse,
	}
}

func walkBlockStmt(s *ast.BlockStmt) *ir.MetaBlockStmt {
	mt := &ir.MetaBlockStmt{
		Pos: s.Pos(),
	}
	for _, stmt := range s.List {
		meta := walkStmt(stmt)
		mt.List = append(mt.List, meta)
	}
	return mt
}

func walkForStmt(s *ast.ForStmt) *ir.MetaForContainer {
	meta := &ir.MetaForContainer{
		Pos:     s.Pos(),
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
		Pos:   s.Pos(),
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
			Pos:      s.Pos(),
			IsMap:    false,
			LenVar:   registerLocalVariable(currentFunc, ".range.len", types.Int),
			Indexvar: registerLocalVariable(currentFunc, ".range.index", types.Int),
			X:        metaX,
		}
	case types.T_MAP:
		meta.ForRangeStmt = &ir.MetaForRangeStmt{
			Pos:     s.Pos(),
			IsMap:   true,
			MapVar:  registerLocalVariable(currentFunc, ".range.map", types.Uintptr),
			ItemVar: registerLocalVariable(currentFunc, ".range.item", types.Uintptr),
			X:       metaX,
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
		Pos: s.Pos(),
		Lhs: lhsMeta,
		Rhs: rhsMeta,
	}
}

func walkSwitchStmt(s *ast.SwitchStmt) *ir.MetaSwitchStmt {
	meta := &ir.MetaSwitchStmt{
		Pos: s.Pos(),
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
		Pos: e.Pos(),
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
			Pos: cc.Pos(),
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
		Pos:      s.Pos(),
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
		Pos:              s.Pos(),
		ContainerForStmt: currentFor,
		ContinueOrBreak:  continueOrBreak,
	}
}

func walkGoStmt(s *ast.GoStmt) *ir.MetaGoStmt {
	fun := walkExpr(s.Call.Fun, nil)
	return &ir.MetaGoStmt{
		Pos: s.Pos(),
		Fun: fun,
	}
}

func walkStmt(stmt ast.Stmt) ir.MetaStmt {
	var mt ir.MetaStmt
	//logf("walkStmt : %s\n", fset.Position(Pos(stmt)).String())
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkBlockStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.ExprStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkExprStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.DeclStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkDeclStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.AssignStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkAssignStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.IncDecStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkIncDecStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.ReturnStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkReturnStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.IfStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkIfStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.ForStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkForStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.RangeStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkRangeStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.BranchStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkBranchStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.SwitchStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkSwitchStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.TypeSwitchStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkTypeSwitchStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	case *ast.GoStmt:
		assert(Pos(s) != 0, "s.Pos() should not be zero", __func__)
		mt = walkGoStmt(s)
		assert(Pos(mt) != 0, "mt.Pos() should not be zero", __func__)
	default:
		throw(stmt)
	}

	assert(mt != nil, "meta should not be nil", __func__)
	return mt
}

func isUniverseNil(m *ir.MetaIdent) bool {
	return m.Kind == "nil"
}

func WalkIdent(e *ast.Ident, ctx *ir.EvalContext) *ir.MetaIdent {
	//	logf("(%s) [WalkIdent] Pos=%d ident=\"%s\"\n", CurrentPkg.name, int(e.Pos()), e.Name)
	meta := &ir.MetaIdent{
		Pos:  e.Pos(),
		Name: e.Name,
	}
	logfncname := "(toplevel)"
	if currentFunc != nil {
		logfncname = currentFunc.Name
	}
	//logf("WalkIdent: pkg=%s func=%s, ident=%s\n", CurrentPkg.name, logfncname, e.Name)
	_ = logfncname
	// what to do ?
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
			meta.Type = meta.Variable.Typ
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
				//logf("ast.Fun=%s\n", e.Name)
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
		Pos: e.Pos(),
	}
	if isQI(e) {
		meta.IsQI = true
		// pkg.ident
		qi := Selector2QI(e)
		meta.QI = qi
		ei := LookupForeignIdent(qi, e.Pos())
		switch ei.Ident.Obj.Kind {
		case ast.Var:
			meta.ForeignValue = ei.MetaIdent
			meta.Type = ei.Type
		case ast.Con:
			meta.ForeignValue = ei.MetaIdent
			meta.Type = ei.Type
		case ast.Fun:
			ff := newForeignFunc(ei, qi)
			foreignMeta := &ir.MetaForeignFuncWrapper{
				Pos: e.Pos(),
				QI:  qi,
				FF:  ff,
				// @TODO: set Type from ff.FuncType
			}
			meta.ForeignValue = foreignMeta
			meta.Type = E2T(ei.Ident.Obj.Decl.(*ast.FuncDecl).Type)
		default:
			panic("Unexpected")
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
	//logf("%s: walkSelectorExpr %s\n", fset.Position(e.Sel.Pos()), e.Sel.Name)
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
				method := lookupMethod(typeOfLeft, e.Sel)
				funcType := method.FuncType
				//logf("%s: Looking for method ... %s\n", e.Sel.Pos(), e.Sel.Name)
				if funcType.Results == nil || len(funcType.Results.List) == 0 {
					return nil, nil, 0, needDeref
				}
				types := FieldList2Types(funcType.Results)
				return types[0], nil, 0, needDeref
			}
		}
	default: // can be a named type of recevier
		method := lookupMethod(typeOfLeft, e.Sel)
		funcType := method.FuncType
		//logf("%s: Looking for method ... %s\n", e.Sel.Pos(), e.Sel.Name)
		if funcType.Results == nil || len(funcType.Results.List) == 0 {
			return nil, nil, 0, false
		}
		types := FieldList2Types(funcType.Results)
		return types[0], nil, 0, false
	}

	//logf("%s: Looking for struct  ... \n", fset.Position(structTypeLiteral.Pos()))
	field := LookupStructField(structTypeLiteral, e.Sel.Name)
	if field != nil {
		offset := GetStructFieldOffset(field)
		return E2T(field.Type), field, offset, needDeref
	}
	if field == nil { // try to find method
		method := lookupMethod(typeOfLeft, e.Sel)
		funcType := method.FuncType
		//logf("%s: Looking for method ... %s\n", e.Sel.Pos(), e.Sel.Name)
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
		Pos:  pos,
		Type: toType,
		Arg0: arg0,
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
		Pos: e.Pos(),
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
		//logf("  fun=%s\n", identFun.Name)
		switch identFun.Obj {
		case universe.Len:
			a0 := walkExpr(e.Args[0], nil)
			return &ir.MetaCallLen{
				Pos:  e.Pos(),
				Type: types.Int,
				Arg0: a0,
			}
		case universe.Cap:
			a0 := walkExpr(e.Args[0], nil)
			return &ir.MetaCallCap{
				Pos:  e.Pos(),
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
				Pos:      e.Pos(),
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
				Pos:      e.Pos(),
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
				Pos:  e.Pos(),
				Type: typ,
				Arg0: a0,
				Arg1: a1,
			}
		case universe.Panic:
			a0 := walkExpr(e.Args[0], nil)
			return &ir.MetaCallPanic{
				Pos:  e.Pos(),
				Type: nil,
				Arg0: a0,
			}
		case universe.Delete:
			a0 := walkExpr(e.Args[0], nil)
			a1 := walkExpr(e.Args[1], nil)
			return &ir.MetaCallDelete{
				Pos:  e.Pos(),
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
			case "runtime_getenv":
				symbol = GetPackageSymbol("runtime", "runtime_getenv")
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
				funcType = ff.FuncType
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
			funcType = ff.FuncType
		} else {
			// method call
			receiver = fn.X
			receiverMeta = walkExpr(fn.X, nil)
			receiverType := GetTypeOfExpr(receiverMeta)
			method := lookupMethod(receiverType, fn.Sel)
			funcType = method.FuncType
			funcVal = NewFuncValueFromSymbol(GetMethodSymbol(method))

			if Kind(receiverType) == types.T_POINTER {
				if method.IsPtrMethod {
					// p.mp() => as it is
				} else {
					// p.mv()
					panic("TBI 4190")
				}
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
						X:    receiver,
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
		args = append(args, a.Meta)
	}

	meta.ParamTypes = paramTypes
	meta.Args = args
	return meta
}

func walkBasicLit(e *ast.BasicLit, ctx *ir.EvalContext) *ir.MetaBasicLit {
	m := &ir.MetaBasicLit{
		Pos:      e.Pos(),
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
		m.IntVal = strconv.Atoi(m.RawValue)
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
		Pos:  e.Pos(),
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

			metaElm := &ir.MetaStructLiteralElement{
				Pos:       kvExpr.Pos(),
				Field:     field,
				FieldType: fieldType,
				ValueMeta: valueMeta,
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
			ms = append(ms, m)
		}
		meta.MetaElms = ms
	case types.T_SLICE:
		arrayType := ut.E.(*ast.ArrayType)
		meta.Len = len(e.Elts)
		meta.ElmType = E2T(arrayType.Elt)
		ctx := &ir.EvalContext{Type: meta.ElmType}
		var ms []ir.MetaExpr
		for _, v := range e.Elts {
			m := walkExpr(v, ctx)
			ms = append(ms, m)
		}
		meta.MetaElms = ms
	}
	return meta
}

func walkUnaryExpr(e *ast.UnaryExpr, ctx *ir.EvalContext) *ir.MetaUnaryExpr {
	meta := &ir.MetaUnaryExpr{
		Pos: e.Pos(),
		Op:  e.Op.String(),
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
			Star: Pos(e),
			X:    xTyp.E,
		}
		meta.Type = E2T(ptrType)
	}

	return meta
}

func walkBinaryExpr(e *ast.BinaryExpr, ctx *ir.EvalContext) *ir.MetaBinaryExpr {
	meta := &ir.MetaBinaryExpr{
		Pos: e.Pos(),
		Op:  e.Op.String(),
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
			panicPos("xTyp should not be nil", Pos(e))
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
		Pos: e.Pos(),
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
		Pos: e.Pos(),
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
func walkStarExpr(e *ast.StarExpr, ctx *ir.EvalContext) *ir.MetaStarExpr {
	meta := &ir.MetaStarExpr{
		Pos: e.Pos(),
	}
	meta.X = walkExpr(e.X, nil)
	xType := GetTypeOfExpr(meta.X)
	origType := xType.E.(*ast.StarExpr)
	meta.Type = E2T(origType.X)
	return meta
}

func walkInterfaceType(e *ast.InterfaceType) {
	// interface{}(e)  conversion. Nothing to do.
}

func walkTypeAssertExpr(e *ast.TypeAssertExpr, ctx *ir.EvalContext) *ir.MetaTypeAssertExpr {
	meta := &ir.MetaTypeAssertExpr{
		Pos: e.Pos(),
	}
	if ctx != nil && ctx.MaybeOK {
		meta.NeedsOK = true
	}
	meta.X = walkExpr(e.X, nil)
	meta.Type = E2T(e.Type)
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
	switch e := expr.(type) {
	case *ast.BasicLit:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return walkBasicLit(e, ctx)
	case *ast.CompositeLit:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return walkCompositeLit(e, ctx)
	case *ast.Ident:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return WalkIdent(e, ctx)
	case *ast.SelectorExpr:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return walkSelectorExpr(e, ctx)
	case *ast.CallExpr:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return walkCallExpr(e, ctx)
	case *ast.IndexExpr:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return walkIndexExpr(e, ctx)
	case *ast.SliceExpr:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return walkSliceExpr(e, ctx)
	case *ast.StarExpr:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return walkStarExpr(e, ctx)
	case *ast.UnaryExpr:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return walkUnaryExpr(e, ctx)
	case *ast.BinaryExpr:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return walkBinaryExpr(e, ctx)
	case *ast.TypeAssertExpr:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return walkTypeAssertExpr(e, ctx)
	case *ast.ParenExpr:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return walkExpr(e.X, ctx)
	case *ast.KeyValueExpr:
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
		panic("the compiler should no reach here")

	// Each one below is not an expr but a type
	case *ast.ArrayType: // type
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return nil
	case *ast.MapType: // type
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return nil
	case *ast.InterfaceType: // type
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return nil
	case *ast.FuncType:
		assert(Pos(e) != 0, "e.Pos() should not be zero", __func__)
		return nil // @TODO walk
	default:
		panic(fmt.Sprintf("unknown type %T", expr))
	}
}

func Pos(node interface{}) token.Pos {
	switch n := node.(type) {
	// Expr
	case *ast.Ident:
		return n.Pos()
	case *ast.Ellipsis:
		return n.Pos()
	case *ast.BasicLit:
		return n.Pos()
	case *ast.CompositeLit:
		return n.Pos()
	case *ast.ParenExpr:
		return n.Pos()
	case *ast.SelectorExpr:
		return n.Pos()
	case *ast.IndexExpr:
		return n.Pos()
	case *ast.SliceExpr:
		return n.Pos()
	case *ast.TypeAssertExpr:
		return n.Pos()
	case *ast.CallExpr:
		return n.Pos()
	case *ast.StarExpr:
		return n.Pos()
	case *ast.UnaryExpr:
		return n.Pos()
	case *ast.BinaryExpr:
		return n.Pos()
	case *ast.KeyValueExpr:
		return n.Pos()
	case *ast.ArrayType:
		return n.Pos()
	case *ast.MapType:
		return n.Pos()
	case *ast.StructType:
		return n.Pos()
	case *ast.FuncType:
		return n.Pos()
	case *ast.InterfaceType:
		return n.Pos()

	// Stmt
	case *ast.DeclStmt:
		return n.Pos()
	case *ast.ExprStmt:
		return n.Pos()
	case *ast.IncDecStmt:
		return n.Pos()
	case *ast.AssignStmt:
		return n.Pos()
	case *ast.GoStmt:
		return n.Pos()
	case *ast.ReturnStmt:
		return n.Pos()
	case *ast.BranchStmt:
		return n.Pos()
	case *ast.BlockStmt:
		return n.Pos()
	case *ast.IfStmt:
		return n.Pos()
	case *ast.CaseClause:
		return n.Pos()
	case *ast.SwitchStmt:
		return n.Pos()
	case *ast.TypeSwitchStmt:
		return n.Pos()
	case *ast.ForStmt:
		return n.Pos()
	case *ast.RangeStmt:
		return n.Pos()
	// IR
	case *ir.MetaBasicLit:
		return n.Pos
	case *ir.MetaCompositLit:
		return n.Pos
	case *ir.MetaIdent:
		return n.Pos
	case *ir.MetaSelectorExpr:
		return n.Pos
	case *ir.MetaForeignFuncWrapper:
		return n.Pos
	case *ir.MetaConversionExpr:
		return n.Pos
	case *ir.MetaCallLen:
		return n.Pos
	case *ir.MetaCallCap:
		return n.Pos
	case *ir.MetaCallNew:
		return n.Pos
	case *ir.MetaCallMake:
		return n.Pos
	case *ir.MetaCallAppend:
		return n.Pos
	case *ir.MetaCallPanic:
		return n.Pos
	case *ir.MetaCallDelete:
		return n.Pos
	case *ir.MetaCallExpr:
		return n.Pos
	case *ir.MetaIndexExpr:
		return n.Pos
	case *ir.MetaSliceExpr:
		return n.Pos
	case *ir.MetaStarExpr:
		return n.Pos
	case *ir.MetaUnaryExpr:
		return n.Pos
	case *ir.MetaBinaryExpr:
		return n.Pos
	case *ir.MetaTypeAssertExpr:
		return n.Pos
	case *ir.MetaBlockStmt:
		return n.Pos
	case *ir.MetaExprStmt:
		return n.Pos
	case *ir.MetaVarDecl:
		return n.Pos
	case *ir.MetaSingleAssign:
		return n.Pos
	case *ir.MetaTupleAssign:
		return n.Pos
	case *ir.MetaReturnStmt:
		return n.Pos
	case *ir.MetaIfStmt:
		return n.Pos
	case *ir.MetaForContainer:
		return n.Pos
	case *ir.MetaForForStmt:
		return n.Pos
	case *ir.MetaForRangeStmt:
		return n.Pos
	case *ir.MetaBranchStmt:
		return n.Pos
	case *ir.MetaSwitchStmt:
		return n.Pos
	case *ir.MetaCaseClause:
		return n.Pos
	case *ir.MetaTypeSwitchStmt:
		return n.Pos
	case *ir.MetaTypeSwitchCaseClose:
		return n.Pos
	case *ir.MetaGoStmt:
		return n.Pos

	}

	panic(fmt.Sprintf("Unknown type:%T", node))
}

func LookupForeignIdent(qi ir.QualifiedIdent, pos token.Pos) *ir.ExportedIdent {
	ei, ok := exportedIdents[string(qi)]
	if !ok {
		panicPos(string(qi)+" Not found in exportedIdents", pos)
	}
	return ei
}

func LookupForeignFunc(qi ir.QualifiedIdent) *ir.ForeignFunc {
	ei := LookupForeignIdent(qi, 1)
	return newForeignFunc(ei, qi)
}

func newForeignFunc(ei *ir.ExportedIdent, qi ir.QualifiedIdent) *ir.ForeignFunc {
	assert(ei.Ident.Obj.Kind == ast.Fun, "should be Fun", __func__)
	decl := ei.Ident.Obj.Decl.(*ast.FuncDecl)
	returnTypes := FieldList2Types(decl.Type.Results)
	paramTypes := FieldList2Types(decl.Type.Params)
	return &ir.ForeignFunc{
		Symbol:      string(qi),
		FuncType:    decl.Type,
		ParamTypes:  paramTypes,
		ReturnTypes: returnTypes,
	}
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

	var hasInitFunc bool
	var typs []*types.Type
	var funcs []*ir.Func
	var consts []*ir.PackageVarConst
	var vars []*ir.PackageVarConst

	var typeSpecs []*ast.TypeSpec
	var funcDecls []*ast.FuncDecl
	var varSpecs []*ast.ValueSpec
	var constSpecs []*ast.ValueSpec

	var exportedTpyes []*types.Type

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
			NamePos: typeSpec.Pos(),
			Obj: &ast.Object{
				Kind: ast.Typ,
				Decl: typeSpec,
			},
		}
		t := E2T(eType)
		t.PkgName = pkg.Name
		t.Name = typeSpec.Name.Name
		typs = append(typs, t)
		exportedTpyes = append(exportedTpyes, t)
		switch Kind(t) {
		case types.T_STRUCT:
			structType := GetUnderlyingType(t)
			calcStructSizeAndSetFieldOffset(structType.E.(*ast.StructType))
		}
		ei := &ir.ExportedIdent{
			Ident:   typeSpec.Name,
			IsType:  true,
			Type:    t,
			PkgName: pkg.Name,
			Pos:     typeSpec.Pos(),
		}
		exportedIdents[string(NewQI(pkg.Name, typeSpec.Name.Name))] = ei
	}

	// collect methods in advance
	for _, funcDecl := range funcDecls {
		if funcDecl.Recv == nil { // non-method function
			if funcDecl.Name.Name == "init" {
				hasInitFunc = true
			}
			qi := NewQI(pkg.Name, funcDecl.Name.Name)
			ei := &ir.ExportedIdent{
				Ident:   funcDecl.Name,
				PkgName: pkg.Name,
				Pos:     funcDecl.Pos(),
			}

			exportedIdents[string(qi)] = ei
		} else { // is method
			method := newMethod(pkg.Name, funcDecl)
			registerMethod(method)
		}
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
			Ident:     lhsIdent,
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
			Ident:     lhsIdent,
			Pos:       lhsIdent.Pos(),
			Type:      t,
			MetaIdent: metaVar,
		}
		exportedIdents[string(NewQI(pkg.Name, lhsIdent.Name))] = ei
	}

	for _, funcDecl := range funcDecls {

		fnc := &ir.Func{
			Name:      funcDecl.Name.Name,
			Decl:      funcDecl,
			Signature: FuncTypeToSignature(funcDecl.Type),
			Localarea: 0,
			Argsarea:  16, // return address + previous rbp
		}
		currentFunc = fnc
		funcs = append(funcs, fnc)

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

func GetSizeOfType(t *types.Type) int {
	ut := GetUnderlyingType(t)
	switch Kind(ut) {
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
		arrayType := ut.E.(*ast.ArrayType)
		elmSize := GetSizeOfType(E2T(arrayType.Elt))
		return elmSize * EvalInt(arrayType.Len)
	case types.T_STRUCT:
		return calcStructSizeAndSetFieldOffset(ut.E.(*ast.StructType))
	case types.T_FUNC:
		return SizeOfPtr
	default:
		unexpectedKind(Kind(t))
	}
	return 0
}

func calcStructSizeAndSetFieldOffset(structType *ast.StructType) int {
	var offset int = 0
	for _, field := range structType.Fields.List {
		setStructFieldOffset(field, offset)
		size := GetSizeOfType(E2T(field.Type))
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
	switch e := expr.(type) {
	case *ast.BasicLit:
		return strconv.Atoi(e.Value)
	}
	panic("Unknown type")
}

func SerializeType(t *types.Type, showPkgPrefix bool) string {
	if t == nil {
		panic("nil type is not expected")
	}
	if t == types.GeneralSliceType {
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
			panic(e.Obj)
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
				typeName := typeSpec.Name.Name
				if showPkgPrefix {
					pkgName := typeSpec.Name.Obj.Data.(string)
					return pkgName + "." + typeName
				} else {
					return typeName
				}
			}
		}
	case *ast.StructType:
		r := "struct{"
		if e.Fields != nil {
			for _, field := range e.Fields.List {
				name := field.Names[0].Name
				typ := E2T(field.Type)
				r += fmt.Sprintf("%s %s;", name, SerializeType(typ, showPkgPrefix))
			}
		}
		return r + "}"
	case *ast.ArrayType:
		if e.Len == nil {
			if e.Elt == nil {
				panic(e)
			}
			return "[]" + SerializeType(E2T(e.Elt), showPkgPrefix)
		} else {
			return "[" + strconv.Itoa(EvalInt(e.Len)) + "]" + SerializeType(E2T(e.Elt), showPkgPrefix)
		}
	case *ast.StarExpr:
		return "*" + SerializeType(E2T(e.X), showPkgPrefix)
	case *ast.Ellipsis: // x ...T
		return "..." + SerializeType(E2T(e.Elt), showPkgPrefix)
	case *ast.InterfaceType:
		return "interface{}" // @TODO list methods
	case *ast.MapType:
		return "map[" + SerializeType(E2T(e.Key), showPkgPrefix) + "]" + SerializeType(E2T(e.Value), showPkgPrefix)
	case *ast.SelectorExpr:
		qi := Selector2QI(e)
		return string(qi)
	case *ast.FuncType:
		return "func()"
	default:
		panic(t)
	}
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

func RestoreFuncDecl(fnc *ir.Func) string {
	var p string
	var r string
	for _, t := range fnc.Signature.ParamTypes {
		if p != "" {
			p += ","
		}
		p += SerializeType(t, false)
	}
	for _, t := range fnc.Signature.ReturnTypes {
		if r != "" {
			r += ","
		}
		r += SerializeType(t, false)
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
