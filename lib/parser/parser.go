package parser

import (
	"os"

	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/scanner"
	"github.com/DQNEO/babygo/lib/token"
)

var ImportsOnly uint8 = 2

var __func__ = "__func__"

func (p *parser) init(fset *token.FileSet, filename string, src []uint8) {
	f := fset.AddFile(filename, -1, len(src))
	p.fset = fset
	p.filename = filename
	p.scanner = &scanner.Scanner{}
	p.scanner.Init(f, src)
	p.next()
}

type parser struct {
	pos        token.Pos
	tok        string
	lit        string
	unresolved []*ast.Ident
	topScope   *ast.Scope
	pkgScope   *ast.Scope
	scanner    *scanner.Scanner
	imports    []*ast.ImportSpec
	filename   string
	fset       *token.FileSet
}

func (p *parser) openScope() {
	p.topScope = ast.NewScope(p.topScope)
}

func (p *parser) closeScope() {
	p.topScope = p.topScope.Outer
}

func (p *parser) consumeComment() {
	p.next0()
}

func (p *parser) next0() {
	var s = p.scanner
	tc := s.Scan()
	p.pos = tc.Pos
	p.tok = tc.Tok
	p.lit = tc.Lit
	//logf("[parser.next0] pos=%d\n", p.tok.pos)
}

func (p *parser) next() {
	p.next0()
	if p.tok == ";" {
		//logff(" [parser] pointing at : \"%s\" newline (%s)\n", p.tok, strconv.Itoa(p.scanner.offset))
	} else if p.tok == "IDENT" {
		//logff(" [parser] pointing at: IDENT \"%s\" (%s)\n", p.lit, strconv.Itoa(p.scanner.offset))
	} else {
		//logff(" [parser] pointing at: \"%s\" %s (%s)\n", p.tok, p.lit, strconv.Itoa(p.scanner.offset))
	}

	if p.tok == "COMMENT" {
		for p.tok == "COMMENT" {
			p.consumeComment()
		}
	}
}

func (p *parser) expect(tok string, who string) {
	if p.tok != tok {
		var s = fmt.Sprintf("%s expected, but got %s", tok, p.tok)
		panic2(who, s)
	}
	p.next()
}

func (p *parser) expectSemi(caller string) {
	if p.tok != ")" && p.tok != "}" {
		switch p.tok {
		case ";":
			p.next()
		default:
			panic2(caller, "semicolon expected, but got token "+p.tok)
		}
	}
}

func (p *parser) parseIdent() *ast.Ident {
	pos := p.pos
	var name string
	if p.tok == "IDENT" {
		name = p.lit
		p.next()
	} else {
		panic2(__func__, "IDENT expected, but got "+p.tok)
	}

	return &ast.Ident{
		NamePos: pos,
		Name:    name,
	}
}

func (p *parser) parseImportSpec() *ast.ImportSpec {
	var pth = p.lit
	pos := p.pos
	p.next()
	spec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			ValuePos: pos,
			Kind:     token.STRING,
			Value:    pth,
		},
	}
	p.imports = append(p.imports, spec)
	return spec
}

func (p *parser) tryVarType(ellipsisOK bool) ast.Expr {
	if ellipsisOK && p.tok == "..." {
		pos := p.pos
		p.next() // consume "..."
		var typ = p.tryIdentOrType()
		if typ != nil {
			p.resolve(typ)
		} else {
			panic2(__func__, "Syntax error")
		}

		return &ast.Ellipsis{
			Ellipsis: pos,
			Elt:      typ,
		}
	}
	return p.tryIdentOrType()
}

func (p *parser) parseVarType(ellipsisOK bool) ast.Expr {

	var typ = p.tryVarType(ellipsisOK)
	if typ == nil {
		panic2(__func__, "nil is not expected")
	}

	return typ
}

func (p *parser) tryType() ast.Expr {

	var typ = p.tryIdentOrType()
	if typ != nil {
		p.resolve(typ)
	}

	return typ
}

func (p *parser) parseType() ast.Expr {
	var typ = p.tryType()
	return typ
}

func (p *parser) parsePointerType() ast.Expr {
	pos := p.pos
	p.expect("*", __func__)
	var base = p.parseType()
	return &ast.StarExpr{
		Star: pos,
		X:    base,
	}
}

func (p *parser) parseArrayType() ast.Expr {
	pos := p.pos
	p.expect("[", __func__)
	var ln ast.Expr
	if p.tok != "]" {
		ln = p.parseRhs()
	}
	p.expect("]", __func__)
	var elt = p.parseType()

	return &ast.ArrayType{
		Lbrack: pos,
		Len:    ln,
		Elt:    elt,
	}
}

func (p *parser) parseFieldDecl(scope *ast.Scope) *ast.Field {
	var varType = p.parseVarType(false)
	var typ = p.tryVarType(false)

	p.expectSemi(__func__)
	ident := varType.(*ast.Ident)
	var field = &ast.Field{
		Type:  typ,
		Names: []*ast.Ident{ident},
	}
	declareField(field, scope, ast.Var, ident)
	p.resolve(typ)
	return field
}

func (p *parser) parseStructType() ast.Expr {
	pos := p.pos
	p.expect("struct", __func__)
	p.expect("{", __func__)

	var _nil *ast.Scope
	var scope = ast.NewScope(_nil)

	var list []*ast.Field
	for p.tok == "IDENT" || p.tok == "*" {
		var field *ast.Field = p.parseFieldDecl(scope)
		list = append(list, field)
	}
	p.expect("}", __func__)

	return &ast.StructType{
		Struct: pos,
		Fields: &ast.FieldList{
			List: list,
		},
	}
}

func (p *parser) parseMaptype() ast.Expr {
	pos := p.pos
	p.expect("map", __func__)
	p.expect("[", __func__)
	keyType := p.parseType()
	p.expect("]", __func__)
	valueType := p.parseType()
	return &ast.MapType{
		Map:   pos,
		Key:   keyType,
		Value: valueType,
	}
}

func (p *parser) parseTypeName() ast.Expr {

	var ident = p.parseIdent()
	if p.tok == "." {
		// ident is a package name
		p.next() // consume "."
		eIdent := ident
		p.resolve(eIdent)
		sel := p.parseIdent()
		selectorExpr := &ast.SelectorExpr{
			X:   eIdent,
			Sel: sel,
		}
		return selectorExpr
	}

	return ident
}

func (p *parser) tryIdentOrType() ast.Expr {

	switch p.tok {
	case "IDENT":
		return p.parseTypeName()
	case "[":
		return p.parseArrayType()
	case "struct":
		return p.parseStructType()
	case "map":
		return p.parseMaptype()
	case "*":
		return p.parsePointerType()
	case "interface":
		pos := p.pos
		p.next()
		p.expect("{", __func__)
		// @TODO parser method sets
		p.expect("}", __func__)
		return &ast.InterfaceType{
			Interface: pos,
			Methods:   nil,
		}
	case "func":
		return p.parseFuncType()
	case "(":
		pos := p.pos
		p.next()
		var _typ = p.parseType()
		p.expect(")", __func__)
		return &ast.ParenExpr{
			Lparen: pos,
			X:      _typ,
		}
	case "type":
		p.next()
		return nil
	}

	return nil
}

func (p *parser) parseParameterList(scope *ast.Scope, ellipsisOK bool) []*ast.Field {

	var list []ast.Expr
	for {
		var varType = p.parseVarType(ellipsisOK)
		list = append(list, varType)
		if p.tok != "," {
			break
		}
		p.next()
		if p.tok == ")" {
			break
		}
	}

	var params []*ast.Field

	var typ = p.tryVarType(ellipsisOK)
	if typ != nil {
		if len(list) > 1 {
			panic2(__func__, "Ident list is not supported")
		}
		var eIdent = list[0]
		ident := eIdent.(*ast.Ident)

		field := &ast.Field{
			Names: []*ast.Ident{ident},
			Type:  typ,
		}
		params = append(params, field)
		declareField(field, scope, ast.Var, ident)
		p.resolve(typ)
		if p.tok != "," {

			return params
		}
		p.next()
		for p.tok != ")" && p.tok != "EOF" {
			ident = p.parseIdent()
			typ = p.parseVarType(ellipsisOK)
			field = &ast.Field{
				Names: []*ast.Ident{ident},
				Type:  typ,
			}
			params = append(params, field)
			declareField(field, scope, ast.Var, ident)
			p.resolve(typ)
			if p.tok != "," {
				break
			}
			p.next()
		}

		return params
	}

	// Type { "," Type } (anonymous parameters)
	params = make([]*ast.Field, len(list), len(list))

	for i, typ := range list {
		p.resolve(typ)
		params[i] = &ast.Field{
			Type: typ,
		}
	}

	return params
}

func (p *parser) parseParameters(scope *ast.Scope, ellipsisOk bool) *ast.FieldList {

	var params []*ast.Field
	pos := p.pos
	p.expect("(", __func__)
	if p.tok != ")" {
		params = p.parseParameterList(scope, ellipsisOk)
	}
	p.expect(")", __func__)

	return &ast.FieldList{
		Opening: pos,
		List:    params,
	}
}

func (p *parser) parseResult(scope *ast.Scope) *ast.FieldList {

	pos := p.pos
	if p.tok == "(" {
		var r = p.parseParameters(scope, false)

		return r
	}

	if p.tok == "{" {

		var _r *ast.FieldList = nil
		return _r
	}
	var typ = p.tryType()
	var list []*ast.Field
	if typ != nil {
		list = append(list, &ast.Field{
			Type: typ,
		})
	}

	return &ast.FieldList{
		Opening: pos,
		List:    list,
	}
}

func (p *parser) parseSignature(scope *ast.Scope) *ast.Signature {

	pos := p.pos
	var params *ast.FieldList
	var results *ast.FieldList
	params = p.parseParameters(scope, true)
	results = p.parseResult(scope)
	return &ast.Signature{
		StartPos: pos,
		Params:   params,
		Results:  results,
	}
}

func declareField(decl *ast.Field, scope *ast.Scope, kind ast.ObjKind, ident *ast.Ident) {
	// declare
	var obj = &ast.Object{
		Decl: decl,
		Name: ident.Name,
		Kind: kind,
	}

	ident.Obj = obj

	// scope insert
	if ident.Name != "_" {
		scope.Insert(obj)
	}
}

func declare(decl interface{}, scope *ast.Scope, kind ast.ObjKind, ident *ast.Ident) {

	//valSpec.Name.Obj
	var obj = &ast.Object{
		Decl: decl,
		Name: ident.Name,
		Kind: kind,
	}
	ident.Obj = obj

	// scope insert
	if ident.Name != "_" {
		scope.Insert(obj)
	}

}

func (p *parser) resolve(x ast.Expr) {
	p.tryResolve(x, true)
}
func (p *parser) tryResolve(x ast.Expr, collectUnresolved bool) {
	if !isExprIdent(x) {
		return
	}
	ident := x.(*ast.Ident)
	if ident.Name == "_" {
		return
	}

	var s *ast.Scope
	for s = p.topScope; s != nil; s = s.Outer {
		var obj = s.Lookup(ident.Name)
		if obj != nil {
			ident.Obj = obj
			return
		}
	}

	if collectUnresolved {
		p.unresolved = append(p.unresolved, ident)
	}
}

func (p *parser) parseOperand(lhs bool) ast.Expr {

	switch p.tok {
	case "IDENT":
		var ident = p.parseIdent()
		var eIdent = (ident)
		if !lhs {
			p.tryResolve(eIdent, true)
		}
		return eIdent
	case "INT", "STRING", "CHAR":
		pos := p.pos
		var basicLit = &ast.BasicLit{
			Kind:     token.Token(p.tok),
			Value:    p.lit,
			ValuePos: pos,
		}
		p.next()

		return basicLit
	case "(":
		pos := p.pos
		p.next() // consume "("
		parserExprLev++
		var x = p.parseRhsOrType()
		parserExprLev--
		p.expect(")", __func__)
		return &ast.ParenExpr{
			Lparen: pos,
			X:      x,
		}
	}

	var typ = p.tryIdentOrType()
	if typ == nil {
		panic2(__func__, "# typ should not be nil\n")
	}

	return typ
}

func (p *parser) parseRhsOrType() ast.Expr {
	var x = p.parseExpr(false)
	return x
}

func (p *parser) parseCallExpr(fn ast.Expr) ast.Expr {
	p.expect("(", __func__)

	var list []ast.Expr
	var ellipsis token.Pos
	for p.tok != ")" {
		var arg = p.parseExpr(false)
		list = append(list, arg)
		if p.tok == "," {
			p.next()
		} else if p.tok == ")" {
			break
		} else if p.tok == "..." {
			// f(a, b, c...)
			//          ^ this
			break
		}
	}

	if p.tok == "..." {
		p.next()
		ellipsis = 1 // this means true
	}

	p.expect(")", __func__)
	return &ast.CallExpr{
		Fun:      fn,
		Args:     list,
		Ellipsis: ellipsis,
	}
}

var parserExprLev int // < 0: in control clause, >= 0: in expression

func (p *parser) parsePrimaryExpr(lhs bool) ast.Expr {

	var x = p.parseOperand(lhs)

	var cnt int

	for {
		cnt++

		if cnt > 100 {
			panic2(__func__, "too many iteration")
		}

		switch p.tok {
		case ".":
			p.next() // consume "."
			if lhs {
				p.resolve(x)
			}

			switch p.tok {
			case "IDENT":
				// Assume CallExpr
				var secondIdent = p.parseIdent()
				var sel = &ast.SelectorExpr{
					X:   x,
					Sel: secondIdent,
				}
				if p.tok == "(" {
					var fn = (sel)
					x = p.parseCallExpr(fn)
				} else {
					x = (sel)
				}
			case "(": // type assertion
				x = p.parseTypeAssertion(x)
			default:
				panic2(__func__, "Unexpected token:"+p.tok)
			}
		case "(":
			// a simpleStmt like x() is parsed in lhs=true mode.
			// So we must resolve "x" here
			if lhs {
				p.resolve(x)
			}
			x = p.parseCallExpr(x)
		case "[":
			p.resolve(x)
			x = p.parseIndexOrSlice(x)
		case "{":
			if isLiteralType(x) && parserExprLev >= 0 {
				x = p.parseLiteralValue(x)
			} else {
				return x
			}
		default:

			return x
		}
	}

	return x
}

func (p *parser) parseTypeAssertion(x ast.Expr) ast.Expr {
	p.expect("(", __func__)
	typ := p.parseType()
	p.expect(")", __func__)
	return &ast.TypeAssertExpr{
		X:    x,
		Type: typ,
	}
}

func (p *parser) parseElement() ast.Expr {
	var x = p.parseExpr(false) // key or value
	var v ast.Expr
	var kvExpr *ast.KeyValueExpr
	if p.tok == ":" {
		p.next() // skip ":"
		v = p.parseExpr(false)
		kvExpr = &ast.KeyValueExpr{
			Key:   x,
			Value: v,
		}
		x = (kvExpr)
	}
	return x
}

func (p *parser) parseElementList() []ast.Expr {
	var list []ast.Expr
	var e ast.Expr
	for p.tok != "}" {
		e = p.parseElement()
		list = append(list, e)
		if p.tok != "," {
			break
		}
		p.expect(",", __func__)
	}
	return list
}

func (p *parser) parseLiteralValue(typ ast.Expr) ast.Expr {

	pos := p.pos
	p.expect("{", __func__)
	var elts []ast.Expr
	if p.tok != "}" {
		elts = p.parseElementList()
	}
	p.expect("}", __func__)

	return &ast.CompositeLit{
		Lbrace: pos,
		Type:   typ,
		Elts:   elts,
	}
}

func isLiteralType(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
	case *ast.SelectorExpr:
		return isExprIdent(e.X)
	case *ast.ArrayType:
	case *ast.StructType:
	//case *ast.MapType:
	default:
		return false
	}

	return true
}

func (p *parser) parseIndexOrSlice(x ast.Expr) ast.Expr {
	p.expect("[", __func__)
	var index = make([]ast.Expr, 3, 3)
	if p.tok != ":" {
		index[0] = p.parseRhs()
	}
	var ncolons int
	for p.tok == ":" && ncolons < 2 {
		ncolons++
		p.next() // consume ":"
		if p.tok != ":" && p.tok != "]" {
			index[ncolons] = p.parseRhs()
		}
	}
	p.expect("]", __func__)

	if ncolons > 0 {
		// slice expression
		var sliceExpr = &ast.SliceExpr{
			Slice3: false,
			X:      x,
			Low:    index[0],
			High:   index[1],
		}
		if ncolons == 2 {
			sliceExpr.Max = index[2]
		}
		return (sliceExpr)
	}

	var indexExpr = &ast.IndexExpr{}
	indexExpr.X = x
	indexExpr.Index = index[0]
	return (indexExpr)
}

func (p *parser) parseUnaryExpr(lhs bool) ast.Expr {
	pos := p.pos
	var r ast.Expr
	switch p.tok {
	case "+", "-", "!", "&":
		var tok = p.tok
		p.next()
		var x = p.parseUnaryExpr(false)
		r = &ast.UnaryExpr{
			OpPos: pos,
			X:     x,
			Op:    token.Token(tok),
		}
		return r
	case "*":
		p.next() // consume "*"
		var x = p.parseUnaryExpr(false)
		r = &ast.StarExpr{
			Star: pos,
			X:    x,
		}
		return r
	}
	r = p.parsePrimaryExpr(lhs)
	return r
}

const LowestPrec int = 0

// https://golang.org/ref/spec#Operators
func precedence(op string) int {
	switch op {
	case "*", "/", "%", "&":
		return 5
	case "+", "-", "|":
		return 4
	case "==", "!=", "<", "<=", ">", ">=":
		return 3
	case "&&":
		return 2
	case "||":
		return 1
	default:
		return 0
	}
	return 0
}

func (p *parser) parseBinaryExpr(lhs bool, prec1 int) ast.Expr {

	var x = p.parseUnaryExpr(lhs)
	var oprec int
	for {
		var op = p.tok
		oprec = precedence(op)
		if oprec < prec1 {

			return x
		}
		p.expect(op, __func__)
		if lhs {
			// x + y
			// x is not lhs at all !!!
			p.resolve(x)
			lhs = false
		}

		var y = p.parseBinaryExpr(false, oprec+1)
		var binaryExpr = &ast.BinaryExpr{}
		binaryExpr.X = x
		binaryExpr.Y = y
		binaryExpr.Op = token.Token(op)
		var r = (binaryExpr)
		x = r
	}

	return x
}

func (p *parser) parseExpr(lhs bool) ast.Expr {

	var e = p.parseBinaryExpr(lhs, 1)

	return e
}

func (p *parser) parseRhs() ast.Expr {
	var x = p.parseExpr(false)
	return x
}

// Extract ast.Expr from ExprStmt. Returns nil if input is nil
func makeExpr(s ast.Stmt) ast.Expr {

	if s == nil {
		var r ast.Expr
		return r
	}
	return s.(*ast.ExprStmt).X
}

func (p *parser) parseGoStmt() ast.Stmt {
	pos := p.pos
	p.expect("go", __func__)
	expr := p.parsePrimaryExpr(false)
	p.expectSemi(__func__)
	return &ast.GoStmt{
		Go:   pos,
		Call: expr.(*ast.CallExpr),
	}
}

func (p *parser) parseForStmt() ast.Stmt {
	pos := p.pos
	p.expect("for", __func__)
	p.openScope()

	var s1 ast.Stmt
	var s2 ast.Stmt
	var s3 ast.Stmt
	var isRange bool
	parserExprLev = -1
	if p.tok != "{" {
		if p.tok != ";" {
			s2 = p.parseSimpleStmt(true)
			var isAssign bool
			var assign *ast.AssignStmt
			assign, isAssign = s2.(*ast.AssignStmt)
			isRange = isAssign && assign.IsRange

		}
		if !isRange && p.tok == ";" {
			p.next() // consume ";"
			s1 = s2
			s2 = nil
			if p.tok != ";" {
				s2 = p.parseSimpleStmt(false)
			}
			p.expectSemi(__func__)
			if p.tok != "{" {
				s3 = p.parseSimpleStmt(false)
			}
		}
	}

	parserExprLev = 0
	var body = p.parseBlockStmt()
	p.expectSemi(__func__)

	var as *ast.AssignStmt
	var rangeX ast.Expr
	if isRange {
		as = s2.(*ast.AssignStmt)

		var key ast.Expr
		var value ast.Expr
		switch len(as.Lhs) {
		case 0:
		case 1:
			key = as.Lhs[0]
		case 2:
			key = as.Lhs[0]
			value = as.Lhs[1]
		default:
			panic2(__func__, "Unexpected len of as.Lhs")
		}

		rangeX = as.Rhs[0].(*ast.UnaryExpr).X
		var rangeStmt = &ast.RangeStmt{
			For: pos,
		}
		rangeStmt.Key = key
		rangeStmt.Value = value
		rangeStmt.X = rangeX
		rangeStmt.Body = body
		rangeStmt.Tok = token.Token(as.Tok)
		p.closeScope()

		return rangeStmt
	}
	var forStmt = &ast.ForStmt{
		For: pos,
	}
	forStmt.Init = s1
	forStmt.Cond = makeExpr(s2)
	forStmt.Post = s3
	forStmt.Body = body
	p.closeScope()

	return forStmt
}

func (p *parser) parseIfStmt() ast.Stmt {
	pos := p.pos
	p.expect("if", __func__)
	parserExprLev = -1
	var condStmt ast.Stmt = p.parseSimpleStmt(false)
	exprStmt := condStmt.(*ast.ExprStmt)
	var cond = exprStmt.X
	parserExprLev = 0
	var body = p.parseBlockStmt()
	var else_ ast.Stmt
	if p.tok == "else" {
		p.next()
		if p.tok == "if" {
			else_ = p.parseIfStmt()
		} else {
			var elseblock = p.parseBlockStmt()
			p.expectSemi(__func__)
			else_ = elseblock
		}
	} else {
		p.expectSemi(__func__)
	}
	var ifStmt = &ast.IfStmt{
		If: pos,
	}
	ifStmt.Cond = cond
	ifStmt.Body = body
	ifStmt.Else = else_

	return ifStmt
}

func (p *parser) parseCaseClause() *ast.CaseClause {

	var list []ast.Expr
	if p.tok == "case" {
		p.next() // consume "case"
		list = p.parseRhsList()
	} else {
		p.expect("default", __func__)
	}
	p.expect(":", __func__)
	p.openScope()
	var body = p.parseStmtList()
	var r = &ast.CaseClause{}
	r.Body = body
	r.List = list
	p.closeScope()

	return r
}

func isTypeSwitchAssert(e ast.Expr) bool {
	_, ok := e.(*ast.TypeAssertExpr)
	return ok
}

func isTypeSwitchGuard(stmt ast.Stmt) bool {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		if isTypeSwitchAssert(s.X) {
			return true
		}
	case *ast.AssignStmt:
		if len(s.Lhs) == 1 && len(s.Rhs) == 1 && isTypeSwitchAssert(s.Rhs[0]) {
			return true
		}
	}
	return false
}

func (p *parser) parseSwitchStmt() ast.Stmt {
	pos := p.pos
	p.expect("switch", __func__)
	p.openScope()

	var s2 ast.Stmt
	parserExprLev = -1
	s2 = p.parseSimpleStmt(false)
	parserExprLev = 0

	p.expect("{", __func__)
	var list []ast.Stmt
	var cc *ast.CaseClause
	var ccs ast.Stmt
	for p.tok == "case" || p.tok == "default" {
		cc = p.parseCaseClause()
		ccs = cc
		list = append(list, ccs)
	}
	p.expect("}", __func__)
	p.expectSemi(__func__)
	var body = &ast.BlockStmt{}
	body.List = list

	typeSwitch := isTypeSwitchGuard(s2)

	p.closeScope()
	if typeSwitch {
		return &ast.TypeSwitchStmt{
			Switch: pos,
			Assign: s2,
			Body:   body,
		}
	} else {
		return &ast.SwitchStmt{
			Switch: pos,
			Body:   body,
			Tag:    makeExpr(s2),
		}
	}
}

func (p *parser) parseLhsList() []ast.Expr {

	var list = p.parseExprList(true)

	if p.tok != ":=" {
		// x = y
		// x is declared earlier and it should be resolved here
		for _, lhs := range list {
			p.resolve(lhs)
		}
	}

	return list
}

func (p *parser) parseSimpleStmt(isRangeOK bool) ast.Stmt {

	var x = p.parseLhsList()
	tokPos := p.pos
	stok := p.tok
	var isRange = false
	var y ast.Expr
	var rangeX ast.Expr
	var rangeUnary *ast.UnaryExpr
	switch stok {
	case ":=", "=", "+=", "-=":
		var assignToken = stok
		p.next() // consume =
		if isRangeOK && p.tok == "range" {
			p.next() // consume "range"
			rangeX = p.parseRhs()
			rangeUnary = &ast.UnaryExpr{}
			rangeUnary.Op = "range"
			rangeUnary.X = rangeX
			y = (rangeUnary)
			isRange = true
		} else {
			y = p.parseExpr(false) // rhs
		}
		var as = &ast.AssignStmt{
			TokPos: tokPos,
		}
		as.Tok = token.Token(assignToken)
		as.Lhs = x
		as.Rhs = make([]ast.Expr, 1, 1)
		as.Rhs[0] = y
		as.IsRange = isRange
		s := as
		if as.Tok == ":=" {
			lhss := x
			for _, lhs := range lhss {
				idnt := lhs.(*ast.Ident)
				declare(as, p.topScope, ast.Var, idnt)
			}
		}

		return s
	case ";":
		var exprStmt = &ast.ExprStmt{}
		exprStmt.X = x[0]

		return exprStmt
	}

	switch stok {
	case "++", "--":
		var sInc = &ast.IncDecStmt{
			TokPos: tokPos,
		}
		sInc.X = x[0]
		sInc.Tok = token.Token(stok)
		p.next() // consume "++" or "--"
		return sInc
	}
	var exprStmt = &ast.ExprStmt{}
	exprStmt.X = x[0]

	return exprStmt
}

func (p *parser) parseStmt() ast.Stmt {

	var s ast.Stmt
	switch p.tok {
	case "var":
		s = &ast.DeclStmt{
			Decl: p.parseDecl("var"),
		}

	case "IDENT", "*":
		s = p.parseSimpleStmt(false)
		p.expectSemi(__func__)
	case "return":
		s = p.parseReturnStmt()
	case "break", "continue":
		s = p.parseBranchStmt(p.tok)
	case "if":
		s = p.parseIfStmt()
	case "switch":
		s = p.parseSwitchStmt()
	case "for":
		s = p.parseForStmt()
	case "go":
		s = p.parseGoStmt()
	default:
		panic2(__func__, "TBI 3:"+p.tok)
	}

	return s
}

func (p *parser) parseExprList(lhs bool) []ast.Expr {

	var list []ast.Expr
	var e = p.parseExpr(lhs)
	list = append(list, e)
	for p.tok == "," {
		p.next() // consume ","
		e = p.parseExpr(lhs)
		list = append(list, e)
	}

	return list
}

func (p *parser) parseRhsList() []ast.Expr {
	var list = p.parseExprList(false)
	return list
}

func (p *parser) parseBranchStmt(tok string) ast.Stmt {
	pos := p.pos
	p.expect(tok, __func__)

	p.expectSemi(__func__)

	return &ast.BranchStmt{
		TokPos: pos,
		Tok:    token.Token(tok),
	}
}

func (p *parser) parseReturnStmt() ast.Stmt {
	pos := p.pos
	p.expect("return", __func__)
	var x []ast.Expr
	if p.tok != ";" && p.tok != "}" {
		x = p.parseRhsList()
	}
	p.expectSemi(__func__)
	var returnStmt = &ast.ReturnStmt{
		Return:  pos,
		Results: x,
	}
	return returnStmt
}

func (p *parser) parseStmtList() []ast.Stmt {
	var list []ast.Stmt
	for p.tok != "}" && p.tok != "EOF" && p.tok != "case" && p.tok != "default" {
		var stmt = p.parseStmt()
		list = append(list, stmt)
	}
	return list
}

func (p *parser) parseBody(scope *ast.Scope) *ast.BlockStmt {
	pos := p.pos
	p.expect("{", __func__)
	p.topScope = scope

	var list = p.parseStmtList()

	p.closeScope()
	p.expect("}", __func__)
	return &ast.BlockStmt{
		Lbrace: pos,
		List:   list,
	}
}

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	pos := p.pos
	p.expect("{", __func__)
	p.openScope()

	var list = p.parseStmtList()

	p.closeScope()
	p.expect("}", __func__)
	return &ast.BlockStmt{
		Lbrace: pos,
		List:   list,
	}
}

func (p *parser) parseDecl(keyword string) *ast.GenDecl {
	pos := p.pos
	var r *ast.GenDecl
	switch p.tok {
	case "var":
		p.expect(keyword, __func__)
		var ident = p.parseIdent()
		var typ = p.parseType()
		var value ast.Expr
		var values []ast.Expr
		if p.tok == "=" {
			p.next()
			value = p.parseExpr(false)
			values = []ast.Expr{value}
		}
		p.expectSemi(__func__)
		names := []*ast.Ident{ident}
		valSpec := &ast.ValueSpec{
			Names:  names,
			Type:   typ,
			Values: values,
		}
		declare(valSpec, p.topScope, ast.Var, ident)
		specs := []ast.Spec{valSpec}
		return &ast.GenDecl{
			TokPos: pos,
			Specs:  specs,
		}
	default:
		panic2(__func__, "TBI\n")
	}
	return r
}

func (p *parser) parserTypeSpec() *ast.TypeSpec {

	p.expect("type", __func__)
	var ident = p.parseIdent()

	var spec = &ast.TypeSpec{}
	spec.Name = ident
	declare(spec, p.topScope, ast.Typ, ident)
	if p.tok == "=" {
		// type alias
		p.next()
		spec.IsAssign = true
	}
	var typ = p.parseType()

	p.expectSemi(__func__)
	spec.Type = typ
	return spec
}

func (p *parser) parseValueSpec(keyword string) *ast.ValueSpec {

	p.expect(keyword, __func__)
	var ident = p.parseIdent()

	var typ = p.parseType()
	var value ast.Expr
	var values []ast.Expr
	if p.tok == "=" {
		p.next()
		value = p.parseExpr(false)
		values = []ast.Expr{value}
	}
	p.expectSemi(__func__)
	names := []*ast.Ident{ident}
	spec := &ast.ValueSpec{
		Names:  names,
		Type:   typ,
		Values: values,
	}
	var kind = ast.Con
	if keyword == "var" {
		kind = ast.Var
	}
	declare(spec, p.topScope, kind, ident)

	return spec
}

func (p *parser) parseFuncType() ast.Expr {
	pos := p.pos
	p.next()
	var scope = ast.NewScope(p.topScope) // function scope
	var sig = p.parseSignature(scope)
	var params = sig.Params
	var results = sig.Results
	ft := &ast.FuncType{
		Func:    pos,
		Params:  params,
		Results: results,
	}
	return ft
}

func (p *parser) parseFuncDecl() ast.Decl {
	pos := p.pos
	p.expect("func", __func__)
	var scope = ast.NewScope(p.topScope) // function scope
	var receivers *ast.FieldList
	if p.tok == "(" {

		receivers = p.parseParameters(scope, false)
	} else {

	}
	var ident = p.parseIdent() // func name
	var sig = p.parseSignature(scope)
	var params = sig.Params
	var results = sig.Results
	if results == nil {

	} else {

	}
	var body *ast.BlockStmt
	if p.tok == "{" {

		body = p.parseBody(scope)

		p.expectSemi(__func__)
	} else {

		p.expectSemi(__func__)
	}
	var decl ast.Decl

	//logf("[parser] p.tok.pos=%d\n", p.tok.pos)

	var funcDecl = &ast.FuncDecl{}
	funcDecl.Recv = receivers
	funcDecl.Name = ident
	funcDecl.Type = &ast.FuncType{
		Func: pos,
	}
	funcDecl.Type.Params = params
	funcDecl.Type.Results = results
	funcDecl.Body = body
	decl = funcDecl
	if receivers == nil {
		declare(funcDecl, p.pkgScope, ast.Fun, ident)
	}
	return decl
}

func (p *parser) parseFile(importsOnly bool) *ast.File {
	// expect "package" keyword
	pos := p.pos
	p.expect("package", __func__)
	p.unresolved = nil
	var ident = p.parseIdent()
	packageName := ident
	p.expectSemi(__func__)

	p.topScope = ast.NewScope(nil) // open scope
	p.pkgScope = p.topScope

	for p.tok == "import" {
		p.expect("import", __func__)
		if p.tok == "(" {
			p.next()
			for p.tok != ")" {
				p.parseImportSpec()
				p.expectSemi(__func__)
			}
			p.next()
			p.expectSemi(__func__)
		} else {
			p.parseImportSpec()
			p.expectSemi(__func__)
		}
	}

	var decls []ast.Decl
	var decl ast.Decl

	for !importsOnly && p.tok != "EOF" {
		switch p.tok {
		case "var", "const":
			var spec = p.parseValueSpec(p.tok)
			specs := []ast.Spec{spec}
			decl = &ast.GenDecl{
				Specs: specs,
			}
		case "func":

			decl = p.parseFuncDecl()
			//logff(" func decl parsed:%s\n", decl.funcDecl.Name.Name)
		case "type":
			spec := p.parserTypeSpec()
			specs := []ast.Spec{spec}
			decl = &ast.GenDecl{
				Specs: specs,
			}

		default:
			panic2(__func__, "TBI:"+p.tok)
		}
		decls = append(decls, decl)
	}

	p.topScope = nil

	var unresolved []*ast.Ident
	for _, idnt := range p.unresolved {
		ps := p.pkgScope
		var obj *ast.Object = ps.Lookup(idnt.Name)
		if obj != nil {
			idnt.Obj = obj
		} else {
			unresolved = append(unresolved, idnt)
		}
	}

	var f = &ast.File{
		Package: pos,
	}
	f.Name = packageName
	f.Scope = p.pkgScope
	f.Decls = decls
	f.Unresolved = unresolved
	f.Imports = p.imports
	return f
}

func readSource(filename string) []uint8 {
	buf, _ := os.ReadFile(filename)
	return buf
}

func ParseFile(fset *token.FileSet, filename string, src interface{}, mode uint8) (*ast.File, *ParserError) {
	//logff("[ParseFile] Start file %s\n", filename)
	var importsOnly bool
	if mode == ImportsOnly {
		importsOnly = true
	}

	text := readSource(filename)
	var p = &parser{}
	packagePos := token.Pos(fset.Base)
	p.init(fset, filename, text)
	astFile := p.parseFile(importsOnly)
	astFile.Package = packagePos
	return astFile, nil
}

func isExprIdent(e ast.Expr) bool {
	_, ok := e.(*ast.Ident)
	return ok
}

func panic2(caller string, x string) {
	panic(caller + ": " + x)
}

type ParserError struct {
	msg string
}

func (err *ParserError) Error() string {
	return err.msg
}
