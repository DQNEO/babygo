package parser

import (
	"os"

	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/scanner"
	"github.com/DQNEO/babygo/lib/token"
)

const ImportsOnly uint8 = 2

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
	lit, tok, pos := p.scanner.Scan()
	p.lit = lit
	p.tok = tok
	p.pos = pos
}

func (p *parser) next() {
	p.next0()
	if p.tok == ";" {
	} else if p.tok == "IDENT" {
	} else {
	}

	if p.tok == "COMMENT" {
		for p.tok == "COMMENT" {
			p.consumeComment()
		}
	}
}

func (p *parser) expect(tok string, who string) {
	if p.tok != tok {
		s := fmt.Sprintf("%s %s expected, but got %s", p.fset.Position(p.pos).String(), tok, p.tok)
		p.panic(who, s)
	}
	p.next()
}

func (p *parser) expectSemi(caller string) {
	if p.tok != ")" && p.tok != "}" {
		switch p.tok {
		case ";":
			p.next()
		default:
			p.panic(caller, "semicolon expected, but got token "+p.tok+" lit:"+p.lit)
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
		p.panic(__func__, "IDENT expected, but got "+p.tok)
	}

	return &ast.Ident{
		NamePos: pos,
		Name:    name,
	}
}

func (p *parser) parseImportSpec() *ast.ImportSpec {
	pth := p.lit
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
			p.panic(__func__, "Syntax error")
		}

		return &ast.Ellipsis{
			Ellipsis: pos,
			Elt:      typ,
		}
	}
	return p.tryIdentOrType()
}

func (p *parser) parseVarType(ellipsisOK bool) ast.Expr {
	typ := p.tryVarType(ellipsisOK)
	if typ == nil {
		p.panic(__func__, "nil is not expected")
	}

	return typ
}

func (p *parser) tryType() ast.Expr {
	typ := p.tryIdentOrType()
	if typ != nil {
		p.resolve(typ)
	}

	return typ
}

func (p *parser) parseType() ast.Expr {
	typ := p.tryType()
	return typ
}

func (p *parser) parsePointerType() ast.Expr {
	pos := p.pos
	p.expect("*", __func__)
	base := p.parseType()
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
	elt := p.parseType()

	return &ast.ArrayType{
		Lbrack: pos,
		Len:    ln,
		Elt:    elt,
	}
}

func (p *parser) parseFieldDecl(scope *ast.Scope) *ast.Field {
	varType := p.parseVarType(false)
	typ := p.tryVarType(false)

	p.expectSemi(__func__)
	ident := varType.(*ast.Ident)
	field := &ast.Field{
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
	scope := ast.NewScope(_nil)

	var list []*ast.Field
	for p.tok == "IDENT" || p.tok == "*" {
		field := p.parseFieldDecl(scope)
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
	ident := p.parseIdent()
	if p.tok == "." {
		// ident is a package name
		p.next() // consume "."
		eIdent := ident
		p.resolve(eIdent)
		sel := p.parseIdent()
		return &ast.SelectorExpr{
			X:   eIdent,
			Sel: sel,
		}
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
		p.expect("interface", __func__)
		p.expect("{", __func__)
		methods := p.parseMethods()
		return &ast.InterfaceType{
			Interface: pos,
			Methods:   methods,
		}
	case "func":
		return p.parseFuncType()
	case "(":
		pos := p.pos
		p.next()
		_typ := p.parseType()
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

	typ := p.tryVarType(ellipsisOK)
	if typ != nil {
		if len(list) > 1 {
			p.panic(__func__, "Ident list is not supported")
		}
		eIdent := list[0]
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
		return p.parseParameters(scope, false)
	}

	if p.tok == "{" {
		return nil
	}
	typ := p.tryType()
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
	params := p.parseParameters(scope, true)
	results := p.parseResult(scope)
	return &ast.Signature{
		StartPos: pos,
		Params:   params,
		Results:  results,
	}
}

func declareField(decl *ast.Field, scope *ast.Scope, kind ast.ObjKind, ident *ast.Ident) {
	// declare
	obj := &ast.Object{
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
	obj := &ast.Object{
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

	for s := p.topScope; s != nil; s = s.Outer {
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

	typ := p.tryIdentOrType()
	if typ == nil {
		p.panic(__func__, "# typ should not be nil\n")
	}

	return typ
}

func (p *parser) parseRhsOrType() ast.Expr {
	x := p.parseExpr(false)
	return x
}

func (p *parser) parseCallExpr(fn ast.Expr) ast.Expr {
	p.expect("(", __func__)

	var list []ast.Expr
	var ellipsis token.Pos
	for p.tok != ")" {
		arg := p.parseExpr(false)
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
	x := p.parseOperand(lhs)

	var cnt int

	for {
		cnt++

		if cnt > 100 {
			p.panic(__func__, "too many iteration")
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
				secondIdent := p.parseIdent()
				sel := &ast.SelectorExpr{
					X:   x,
					Sel: secondIdent,
				}
				if p.tok == "(" {
					x = p.parseCallExpr(sel)
				} else {
					x = sel
				}
			case "(": // type assertion
				x = p.parseTypeAssertion(x)
			default:
				p.panic(__func__, "Unexpected token:"+p.tok)
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
	x := p.parseExpr(false) // key or value
	var v ast.Expr
	if p.tok == ":" {
		p.next() // skip ":"
		v = p.parseExpr(false)
		x = &ast.KeyValueExpr{
			Key:   x,
			Value: v,
		}
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
		return sliceExpr
	}

	return &ast.IndexExpr{
		X:     x,
		Index: index[0],
	}
}

func (p *parser) parseUnaryExpr(lhs bool) ast.Expr {
	pos := p.pos
	switch p.tok {
	case "+", "-", "!", "&":
		tok := p.tok
		p.next()
		x := p.parseUnaryExpr(false)
		return &ast.UnaryExpr{
			OpPos: pos,
			X:     x,
			Op:    token.Token(tok),
		}
	case "*":
		p.next() // consume "*"
		x := p.parseUnaryExpr(false)
		return &ast.StarExpr{
			Star: pos,
			X:    x,
		}
	}
	return p.parsePrimaryExpr(lhs)
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
}

func (p *parser) parseBinaryExpr(lhs bool, prec1 int) ast.Expr {
	x := p.parseUnaryExpr(lhs)
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

		y := p.parseBinaryExpr(false, oprec+1)
		x = &ast.BinaryExpr{
			X:  x,
			Y:  y,
			Op: token.Token(op),
		}
	}
}

func (p *parser) parseExpr(lhs bool) ast.Expr {
	e := p.parseBinaryExpr(lhs, 1)
	return e
}

func (p *parser) parseRhs() ast.Expr {
	x := p.parseExpr(false)
	return x
}

// Extract ast.Expr from ExprStmt. Returns nil if input is nil
func makeExpr(s ast.Stmt) ast.Expr {
	if s == nil {
		return nil
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
			assign, isAssign := s2.(*ast.AssignStmt)
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
	body := p.parseBlockStmt()
	p.expectSemi(__func__)

	if isRange {
		as := s2.(*ast.AssignStmt)
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
			p.panic(__func__, "Unexpected len of as.Lhs")
		}

		rangeX := as.Rhs[0].(*ast.UnaryExpr).X
		p.closeScope()
		return &ast.RangeStmt{
			For:   pos,
			Key:   key,
			Value: value,
			X:     rangeX,
			Body:  body,
			Tok:   as.Tok,
		}
	}
	cond := makeExpr(s2)
	p.closeScope()
	return &ast.ForStmt{
		For:  pos,
		Init: s1,
		Cond: cond,
		Post: s3,
		Body: body,
	}
}

func (p *parser) parseIfStmt() ast.Stmt {
	pos := p.pos
	p.expect("if", __func__)
	parserExprLev = -1
	condStmt := p.parseSimpleStmt(false)
	exprStmt := condStmt.(*ast.ExprStmt)
	cond := exprStmt.X
	parserExprLev = 0
	body := p.parseBlockStmt()
	var else_ ast.Stmt
	if p.tok == "else" {
		p.next()
		if p.tok == "if" {
			else_ = p.parseIfStmt()
		} else {
			elseblock := p.parseBlockStmt()
			p.expectSemi(__func__)
			else_ = elseblock
		}
	} else {
		p.expectSemi(__func__)
	}
	return &ast.IfStmt{
		If:   pos,
		Cond: cond,
		Body: body,
		Else: else_,
	}
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
	body := p.parseStmtList()
	p.closeScope()
	return &ast.CaseClause{
		List: list,
		Body: body,
	}
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

	parserExprLev = -1
	s2 := p.parseSimpleStmt(false)
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
	body := &ast.BlockStmt{
		List: list,
	}
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
	list := p.parseExprList(true)
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
	x := p.parseLhsList()
	tokPos := p.pos
	stok := p.tok
	var isRange bool
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
		s := &ast.AssignStmt{
			TokPos:  tokPos,
			Tok:     token.Token(assignToken),
			Lhs:     x,
			Rhs:     []ast.Expr{y},
			IsRange: isRange,
		}
		if s.Tok == ":=" {
			lhss := x
			for _, lhs := range lhss {
				idnt := lhs.(*ast.Ident)
				declare(s, p.topScope, ast.Var, idnt)
			}
		}
		return s
	case ";":
		return &ast.ExprStmt{
			X: x[0],
		}
	}

	switch stok {
	case "++", "--":
		p.next() // consume "++" or "--"
		return &ast.IncDecStmt{
			TokPos: tokPos,
			X:      x[0],
			Tok:    token.Token(stok),
		}
	}
	return &ast.ExprStmt{
		X: x[0],
	}
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
		p.panic(__func__, "TBI 3:"+p.tok)
	}

	return s
}

func (p *parser) parseExprList(lhs bool) []ast.Expr {
	e := p.parseExpr(lhs)
	list := []ast.Expr{e}
	for p.tok == "," {
		p.next() // consume ","
		e = p.parseExpr(lhs)
		list = append(list, e)
	}

	return list
}

func (p *parser) parseRhsList() []ast.Expr {
	list := p.parseExprList(false)
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
	return &ast.ReturnStmt{
		Return:  pos,
		Results: x,
	}
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

	list := p.parseStmtList()

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
	list := p.parseStmtList()
	p.closeScope()
	p.expect("}", __func__)
	return &ast.BlockStmt{
		Lbrace: pos,
		List:   list,
	}
}

func (p *parser) parseDecl(keyword string) *ast.GenDecl {
	pos := p.pos
	switch p.tok {
	case "var":
		p.expect(keyword, __func__)
		ident := p.parseIdent()
		typ := p.parseType()
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
		p.panic(__func__, "TBI\n")
	}
	return nil
}

func (p *parser) parserTypeSpec() *ast.TypeSpec {
	p.expect("type", __func__)
	ident := p.parseIdent()
	spec := &ast.TypeSpec{
		Name: ident,
	}
	declare(spec, p.topScope, ast.Typ, ident)
	if p.tok == "=" {
		// type alias
		p.next()
		spec.IsAssign = true
	}
	typ := p.parseType()
	p.expectSemi(__func__)
	spec.Type = typ
	return spec
}

func (p *parser) parseValueSpec(keyword string) *ast.ValueSpec {
	p.expect(keyword, __func__)
	ident := p.parseIdent()

	typ := p.parseType()
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
	var kind ast.ObjKind
	if keyword == "var" {
		kind = ast.Var
	} else {
		kind = ast.Con
	}
	declare(spec, p.topScope, kind, ident)

	return spec
}

func (p *parser) parseFuncType() ast.Expr {
	pos := p.pos
	p.expect("func", __func__)
	scope := ast.NewScope(p.topScope) // function scope
	sig := p.parseSignature(scope)
	return &ast.FuncType{
		Func:    pos,
		Params:  sig.Params,
		Results: sig.Results,
	}
}

func (p *parser) parseMethods() *ast.FieldList {
	pos := p.pos

	var list []*ast.Field
	for p.tok == "IDENT" {
		ident := p.parseIdent() // method name

		scope := ast.NewScope(p.topScope) // function scope
		sig := p.parseSignature(scope)

		funcType := &ast.FuncType{
			Func:    pos,
			Params:  sig.Params,
			Results: sig.Results,
		}
		field := &ast.Field{
			Names: []*ast.Ident{ident},
			Type:  funcType,
		}
		list = append(list, field)
		p.expectSemi(__func__)
	}
	p.expect("}", __func__)

	return &ast.FieldList{
		Opening: pos,
		List:    list,
	}
}

func (p *parser) parseFuncDecl() ast.Decl {
	pos := p.pos
	p.expect("func", __func__)
	scope := ast.NewScope(p.topScope) // function scope
	var receivers *ast.FieldList
	if p.tok == "(" {
		receivers = p.parseParameters(scope, false)
	}
	ident := p.parseIdent() // func name
	sig := p.parseSignature(scope)

	var body *ast.BlockStmt
	if p.tok == "{" {
		body = p.parseBody(scope)
	}
	p.expectSemi(__func__)
	funcType := &ast.FuncType{
		Func:    pos,
		Params:  sig.Params,
		Results: sig.Results,
	}
	funcDecl := &ast.FuncDecl{
		Recv: receivers,
		Name: ident,
		Body: body,
		Type: funcType,
	}
	if receivers == nil {
		declare(funcDecl, p.pkgScope, ast.Fun, ident)
	}
	return funcDecl
}

func (p *parser) parseFile(importsOnly bool) *ast.File {
	pos := p.pos
	p.expect("package", __func__)
	p.unresolved = nil
	ident := p.parseIdent()
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
			spec := p.parseValueSpec(p.tok)
			specs := []ast.Spec{spec}
			decl = &ast.GenDecl{
				Specs: specs,
			}
		case "func":
			decl = p.parseFuncDecl()
		case "type":
			spec := p.parserTypeSpec()
			specs := []ast.Spec{spec}
			decl = &ast.GenDecl{
				Specs: specs,
			}
		default:
			p.panic(__func__, "TBI:"+p.tok)
		}
		decls = append(decls, decl)
	}

	p.topScope = nil

	var unresolved []*ast.Ident
	for _, idnt := range p.unresolved {
		ps := p.pkgScope
		obj := ps.Lookup(idnt.Name)
		if obj != nil {
			idnt.Obj = obj
		} else {
			unresolved = append(unresolved, idnt)
		}
	}

	return &ast.File{
		Package:    pos,
		Name:       packageName,
		Scope:      p.pkgScope,
		Decls:      decls,
		Unresolved: unresolved,
		Imports:    p.imports,
	}
}

func readSource(filename string) ([]uint8, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func ParseFile(fset *token.FileSet, filename string, src interface{}, mode uint8) (*ast.File, *ParserError) {
	var importsOnly bool
	if mode == ImportsOnly {
		importsOnly = true
	}

	text, err := readSource(filename)
	if err != nil {
		msg := err.Error()
		e := &ParserError{msg: msg}
		return nil, e
	}
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

func (p *parser) panic(caller string, x string) {
	panic(caller + ": " + x + "\n\t" + p.fset.Position(p.pos).String())
}

type ParserError struct {
	msg string
}

func (err *ParserError) Error() string {
	return err.msg
}
