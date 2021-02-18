package main

import (
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/strconv"
	"syscall"
)

const O_READONLY int = 0
const FILE_SIZE int = 2000000

func readFile(filename string) []uint8 {
	var fd int
	logf("Opening %s\n", filename)
	fd, _ = syscall.Open(filename, O_READONLY, 0)
	if fd < 0 {
		panic("syscall.Open failed: " + filename)
	}
	var buf = make([]uint8, FILE_SIZE, FILE_SIZE)
	var n int
	// @TODO check error
	n, _ = syscall.Read(fd, buf)
	logf("syscall.Read len=%s\n", strconv.Itoa(n))
	var readbytes = buf[0:n]
	return readbytes
}

func (p *parser) init(src []uint8) {
	var s = p.scanner
	s.Init(src)
	p.next()
}

type parser struct {
	tok        *TokenContainer
	unresolved []*ast.Ident
	topScope   *ast.Scope
	pkgScope   *ast.Scope
	scanner    *scanner
	imports    []*ast.ImportSpec
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
	p.tok = s.Scan()
}

func (p *parser) next() {
	p.next0()
	if p.tok.tok == ";" {
		logf(" [parser] pointing at : \"%s\" newline (%s)\n", p.tok.tok, strconv.Itoa(p.scanner.offset))
	} else if p.tok.tok == "IDENT" {
		logf(" [parser] pointing at: IDENT \"%s\" (%s)\n", p.tok.lit, strconv.Itoa(p.scanner.offset))
	} else {
		logf(" [parser] pointing at: \"%s\" %s (%s)\n", p.tok.tok, p.tok.lit, strconv.Itoa(p.scanner.offset))
	}

	if p.tok.tok == "COMMENT" {
		for p.tok.tok == "COMMENT" {
			p.consumeComment()
		}
	}
}

func (p *parser) expect(tok string, who string) {
	if p.tok.tok != tok {
		var s = fmt.Sprintf("%s expected, but got %s", tok, p.tok.tok)
		panic2(who, s)
	}
	logf(" [%s] consumed \"%s\"\n", who, p.tok.tok)
	p.next()
}

func (p *parser) expectSemi(caller string) {
	if p.tok.tok != ")" && p.tok.tok != "}" {
		switch p.tok.tok {
		case ";":
			logf(" [%s] consumed semicolon %s\n", caller, p.tok.tok)
			p.next()
		default:
			panic2(caller, "semicolon expected, but got token "+p.tok.tok)
		}
	}
}

func (p *parser) parseIdent() *ast.Ident {
	var name string
	if p.tok.tok == "IDENT" {
		name = p.tok.lit
		p.next()
	} else {
		panic2(__func__, "IDENT expected, but got "+p.tok.tok)
	}
	logf(" [%s] ident name = %s\n", __func__, name)
	return &ast.Ident{
		Name: name,
	}
}

func (p *parser) parseImportSpec() *ast.ImportSpec {
	var pth = p.tok.lit
	p.next()
	spec := &ast.ImportSpec{
		Path: pth,
	}
	p.imports = append(p.imports, spec)
	return spec
}

func (p *parser) tryVarType(ellipsisOK bool) ast.Expr {
	if ellipsisOK && p.tok.tok == "..." {
		p.next() // consume "..."
		var typ = p.tryIdentOrType()
		if typ != nil {
			p.resolve(typ)
		} else {
			panic2(__func__, "Syntax error")
		}

		return (&ast.Ellipsis{
			Elt: typ,
		})
	}
	return p.tryIdentOrType()
}

func (p *parser) parseVarType(ellipsisOK bool) ast.Expr {
	logf(" [%s] begin\n", __func__)
	var typ = p.tryVarType(ellipsisOK)
	if typ == nil {
		panic2(__func__, "nil is not expected")
	}
	logf(" [%s] end\n", __func__)
	return typ
}

func (p *parser) tryType() ast.Expr {
	logf(" [%s] begin\n", __func__)
	var typ = p.tryIdentOrType()
	if typ != nil {
		p.resolve(typ)
	}
	logf(" [%s] end\n", __func__)
	return typ
}

func (p *parser) parseType() ast.Expr {
	var typ = p.tryType()
	return typ
}

func (p *parser) parsePointerType() ast.Expr {
	p.expect("*", __func__)
	var base = p.parseType()
	return (&ast.StarExpr{
		X: base,
	})
}

func (p *parser) parseArrayType() ast.Expr {
	p.expect("[", __func__)
	var ln ast.Expr
	if p.tok.tok != "]" {
		ln = p.parseRhs()
	}
	p.expect("]", __func__)
	var elt = p.parseType()

	return (&ast.ArrayType{
		Elt: elt,
		Len: ln,
	})
}

func (p *parser) parseFieldDecl(scope *ast.Scope) *ast.Field {

	var varType = p.parseVarType(false)
	var typ = p.tryVarType(false)

	p.expectSemi(__func__)
	ident := expr2Ident(varType)
	var field = &ast.Field{
		Type: typ,
		Name: ident,
	}
	declareField(field, scope, ast.Var, ident)
	p.resolve(typ)
	return field
}

func (p *parser) parseStructType() ast.Expr {
	p.expect("struct", __func__)
	p.expect("{", __func__)

	var _nil *ast.Scope
	var scope = ast.NewScope(_nil)

	var list []*ast.Field
	for p.tok.tok == "IDENT" || p.tok.tok == "*" {
		var field *ast.Field = p.parseFieldDecl(scope)
		list = append(list, field)
	}
	p.expect("}", __func__)

	return (&ast.StructType{
		Fields: &ast.FieldList{
			List: list,
		},
	})
}

func (p *parser) parseTypeName() ast.Expr {
	logf(" [%s] begin\n", __func__)
	var ident = p.parseIdent()
	if p.tok.tok == "." {
		// ident is a package name
		p.next() // consume "."
		eIdent := (ident)
		p.resolve(eIdent)
		sel := p.parseIdent()
		selectorExpr := &ast.SelectorExpr{
			X:   eIdent,
			Sel: sel,
		}
		return (selectorExpr)
	}
	logf(" [%s] end\n", __func__)
	return (ident)
}

func (p *parser) tryIdentOrType() ast.Expr {
	logf(" [%s] begin\n", __func__)
	switch p.tok.tok {
	case "IDENT":
		return p.parseTypeName()
	case "[":
		return p.parseArrayType()
	case "struct":
		return p.parseStructType()
	case "*":
		return p.parsePointerType()
	case "interface":
		p.next()
		p.expect("{", __func__)
		// @TODO parser method sets
		p.expect("}", __func__)
		return (&ast.InterfaceType{
			Methods: nil,
		})
	case "(":
		p.next()
		var _typ = p.parseType()
		p.expect(")", __func__)
		return (&ast.ParenExpr{
			X: _typ,
		})
	case "type":
		p.next()
		return nil
	}

	return nil
}

func (p *parser) parseParameterList(scope *ast.Scope, ellipsisOK bool) []*ast.Field {
	logf(" [%s] begin\n", __func__)
	var list []ast.Expr
	for {
		var varType = p.parseVarType(ellipsisOK)
		list = append(list, varType)
		if p.tok.tok != "," {
			break
		}
		p.next()
		if p.tok.tok == ")" {
			break
		}
	}
	logf(" [%s] collected list n=%s\n", __func__, strconv.Itoa(len(list)))

	var params []*ast.Field

	var typ = p.tryVarType(ellipsisOK)
	if typ != nil {
		if len(list) > 1 {
			panic2(__func__, "Ident list is not supported")
		}
		var eIdent = list[0]
		ident := expr2Ident(eIdent)
		logf(" [%s] ident.Name=%s\n", __func__, ident.Name)
		var field = &ast.Field{}
		field.Name = ident
		field.Type = typ
		params = append(params, field)
		declareField(field, scope, ast.Var, ident)
		p.resolve(typ)
		if p.tok.tok != "," {
			logf("  end %s\n", __func__)
			return params
		}
		p.next()
		for p.tok.tok != ")" && p.tok.tok != "EOF" {
			ident = p.parseIdent()
			typ = p.parseVarType(ellipsisOK)
			field = &ast.Field{
				Name: ident,
				Type: typ,
			}
			params = append(params, field)
			declareField(field, scope, ast.Var, ident)
			p.resolve(typ)
			if p.tok.tok != "," {
				break
			}
			p.next()
		}
		logf("  end %s\n", __func__)
		return params
	}

	// Type { "," Type } (anonymous parameters)
	params = make([]*ast.Field, len(list), len(list))

	for i, typ := range list {
		p.resolve(typ)
		params[i] = &ast.Field{
			Type: typ,
		}
		logf(" [DEBUG] range i = %s\n", strconv.Itoa(i))
	}
	logf("  end %s\n", __func__)
	return params
}

func (p *parser) parseParameters(scope *ast.Scope, ellipsisOk bool) *ast.FieldList {
	logf(" [%s] begin\n", __func__)
	var params []*ast.Field
	p.expect("(", __func__)
	if p.tok.tok != ")" {
		params = p.parseParameterList(scope, ellipsisOk)
	}
	p.expect(")", __func__)
	logf(" [%s] end\n", __func__)
	return &ast.FieldList{
		List: params,
	}
}

func (p *parser) parseResult(scope *ast.Scope) *ast.FieldList {
	logf(" [%s] begin\n", __func__)

	if p.tok.tok == "(" {
		var r = p.parseParameters(scope, false)
		logf(" [%s] end\n", __func__)
		return r
	}

	if p.tok.tok == "{" {
		logf(" [%s] end\n", __func__)
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
	logf(" [%s] end\n", __func__)
	return &ast.FieldList{
		List: list,
	}
}

func (p *parser) parseSignature(scope *ast.Scope) *ast.Signature {
	logf(" [%s] begin\n", __func__)
	var params *ast.FieldList
	var results *ast.FieldList
	params = p.parseParameters(scope, true)
	results = p.parseResult(scope)
	return &ast.Signature{
		Params:  params,
		Results: results,
	}
}

func declareField(decl *ast.Field, scope *ast.Scope, kind string, ident *ast.Ident) {
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

func declare(decl interface{}, scope *ast.Scope, kind string, ident *ast.Ident) {
	logf(" [declare] ident %s\n", ident.Name)

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
	logf(" [declare] end\n")

}

func (p *parser) resolve(x ast.Expr) {
	p.tryResolve(x, true)
}
func (p *parser) tryResolve(x ast.Expr, collectUnresolved bool) {
	if !isExprIdent(x) {
		return
	}
	ident := expr2Ident(x)
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
		logf(" appended unresolved ident %s\n", ident.Name)
	}
}

func (p *parser) parseOperand() ast.Expr {
	logf("   begin %s\n", __func__)
	switch p.tok.tok {
	case "IDENT":
		var ident = p.parseIdent()
		var eIdent = (ident)
		p.tryResolve(eIdent, true)
		logf("   end %s\n", __func__)
		return eIdent
	case "INT", "STRING", "CHAR":
		var basicLit = &ast.BasicLit{
			Kind:  p.tok.tok,
			Value: p.tok.lit,
		}
		p.next()
		logf("   end %s\n", __func__)
		return (basicLit)
	case "(":
		p.next() // consume "("
		parserExprLev++
		var x = p.parseRhsOrType()
		parserExprLev--
		p.expect(")", __func__)
		return (&ast.ParenExpr{
			X: x,
		})
	}

	var typ = p.tryIdentOrType()
	if typ == nil {
		panic2(__func__, "# typ should not be nil\n")
	}
	logf("   end %s\n", __func__)

	return typ
}

func (p *parser) parseRhsOrType() ast.Expr {
	var x = p.parseExpr()
	return x
}

func (p *parser) parseCallExpr(fn ast.Expr) ast.Expr {
	p.expect("(", __func__)
	logf(" [parsePrimaryExpr] p.tok.tok=%s\n", p.tok.tok)
	var list []ast.Expr
	var ellipsis bool
	for p.tok.tok != ")" {
		var arg = p.parseExpr()
		list = append(list, arg)
		if p.tok.tok == "," {
			p.next()
		} else if p.tok.tok == ")" {
			break
		} else if p.tok.tok == "..." {
			// f(a, b, c...)
			//          ^ this
			break
		}
	}

	if p.tok.tok == "..." {
		p.next()
		ellipsis = true
	}

	p.expect(")", __func__)
	return (&ast.CallExpr{
		Fun:      fn,
		Args:     list,
		Ellipsis: ellipsis,
	})
}

var parserExprLev int // < 0: in control clause, >= 0: in expression

func (p *parser) parsePrimaryExpr() ast.Expr {
	logf("   begin %s\n", __func__)
	var x = p.parseOperand()

	var cnt int

	for {
		cnt++
		logf("    [%s] tok=%s\n", __func__, p.tok.tok)
		if cnt > 100 {
			panic2(__func__, "too many iteration")
		}

		switch p.tok.tok {
		case ".":
			p.next() // consume "."

			switch p.tok.tok {
			case "IDENT":
				// Assume CallExpr
				var secondIdent = p.parseIdent()
				var sel = &ast.SelectorExpr{
					X:   x,
					Sel: secondIdent,
				}
				if p.tok.tok == "(" {
					var fn = (sel)
					// string = x.ident.Name + "." + secondIdent
					x = p.parseCallExpr(fn)
					logf(" [parsePrimaryExpr] 741 p.tok.tok=%s\n", p.tok.tok)
				} else {
					logf("   end parsePrimaryExpr()\n")
					x = (sel)
				}
			case "(": // type assertion
				x = p.parseTypeAssertion(x)
			default:
				panic2(__func__, "Unexpected token:"+p.tok.tok)
			}
		case "(":
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
			logf("   end %s\n", __func__)
			return x
		}
	}

	logf("   end %s\n", __func__)
	return x
}

func (p *parser) parseTypeAssertion(x ast.Expr) ast.Expr {
	p.expect("(", __func__)
	typ := p.parseType()
	p.expect(")", __func__)
	return (&ast.TypeAssertExpr{
		X:    x,
		Type: typ,
	})
}

func (p *parser) parseElement() ast.Expr {
	var x = p.parseExpr() // key or value
	var v ast.Expr
	var kvExpr *ast.KeyValueExpr
	if p.tok.tok == ":" {
		p.next() // skip ":"
		v = p.parseExpr()
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
	for p.tok.tok != "}" {
		e = p.parseElement()
		list = append(list, e)
		if p.tok.tok != "," {
			break
		}
		p.expect(",", __func__)
	}
	return list
}

func (p *parser) parseLiteralValue(typ ast.Expr) ast.Expr {
	logf("   start %s\n", __func__)
	p.expect("{", __func__)
	var elts []ast.Expr
	if p.tok.tok != "}" {
		elts = p.parseElementList()
	}
	p.expect("}", __func__)

	logf("   end %s\n", __func__)
	return (&ast.CompositeLit{
		Type: typ,
		Elts: elts,
	})
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
	if p.tok.tok != ":" {
		index[0] = p.parseRhs()
	}
	var ncolons int
	for p.tok.tok == ":" && ncolons < 2 {
		ncolons++
		p.next() // consume ":"
		if p.tok.tok != ":" && p.tok.tok != "]" {
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

func (p *parser) parseUnaryExpr() ast.Expr {
	var r ast.Expr
	logf("   begin parseUnaryExpr()\n")
	switch p.tok.tok {
	case "+", "-", "!", "&":
		var tok = p.tok.tok
		p.next()
		var x = p.parseUnaryExpr()
		logf(" [DEBUG] unary op = %s\n", tok)
		r = (&ast.UnaryExpr{
			X:  x,
			Op: tok,
		})
		return r
	case "*":
		p.next() // consume "*"
		var x = p.parseUnaryExpr()
		r = (&ast.StarExpr{
			X: x,
		})
		return r
	}
	r = p.parsePrimaryExpr()
	logf("   end parseUnaryExpr()\n")
	return r
}

const LowestPrec int = 0

func precedence(op string) int {
	switch op {
	case "||":
		return 1
	case "&&":
		return 2
	case "==", "!=", "<", "<=", ">", ">=":
		return 3
	case "+", "-":
		return 4
	case "*", "/", "%":
		return 5
	default:
		return 0
	}
	return 0
}

func (p *parser) parseBinaryExpr(prec1 int) ast.Expr {
	logf("   begin parseBinaryExpr() prec1=%s\n", strconv.Itoa(prec1))
	var x = p.parseUnaryExpr()
	var oprec int
	for {
		var op = p.tok.tok
		oprec = precedence(op)
		logf(" oprec %s\n", strconv.Itoa(oprec))
		logf(" precedence \"%s\" %s < %s\n", op, strconv.Itoa(oprec), strconv.Itoa(prec1))
		if oprec < prec1 {
			logf("   end parseBinaryExpr() (NonBinary)\n")
			return x
		}
		p.expect(op, __func__)
		var y = p.parseBinaryExpr(oprec + 1)
		var binaryExpr = &ast.BinaryExpr{}
		binaryExpr.X = x
		binaryExpr.Y = y
		binaryExpr.Op = op
		var r = (binaryExpr)
		x = r
	}
	logf("   end parseBinaryExpr()\n")
	return x
}

func (p *parser) parseExpr() ast.Expr {
	logf("   begin p.parseExpr()\n")
	var e = p.parseBinaryExpr(1)
	logf("   end p.parseExpr()\n")
	return e
}

func (p *parser) parseRhs() ast.Expr {
	var x = p.parseExpr()
	return x
}

// Extract ast.Expr from ExprStmt. Returns nil if input is nil
func makeExpr(s ast.Stmt) ast.Expr {
	logf(" begin %s\n", __func__)
	if s == nil {
		var r ast.Expr
		return r
	}
	return stmt2ExprStmt(s).X
}

func (p *parser) parseForStmt() ast.Stmt {
	logf(" begin %s\n", __func__)
	p.expect("for", __func__)
	p.openScope()

	var s1 ast.Stmt
	var s2 ast.Stmt
	var s3 ast.Stmt
	var isRange bool
	parserExprLev = -1
	if p.tok.tok != "{" {
		if p.tok.tok != ";" {
			s2 = p.parseSimpleStmt(true)
			var isAssign bool
			var assign *ast.AssignStmt
			assign, isAssign = s2.(*ast.AssignStmt)
			isRange = isAssign && assign.IsRange
			logf(" [%s] isRange=true\n", __func__)
		}
		if !isRange && p.tok.tok == ";" {
			p.next() // consume ";"
			s1 = s2
			s2 = nil
			if p.tok.tok != ";" {
				s2 = p.parseSimpleStmt(false)
			}
			p.expectSemi(__func__)
			if p.tok.tok != "{" {
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
		assert(isStmtAssignStmt(s2), "type mismatch:"+dtypeOf(s2), __func__)
		as = stmt2AssignStmt(s2)
		logf(" [DEBUG] range as len lhs=%s\n", strconv.Itoa(len(as.Lhs)))
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

		rangeX = expr2UnaryExpr(as.Rhs[0]).X
		var rangeStmt = &ast.RangeStmt{}
		rangeStmt.Key = key
		rangeStmt.Value = value
		rangeStmt.X = rangeX
		rangeStmt.Body = body
		rangeStmt.Tok = as.Tok
		p.closeScope()
		logf(" end %s\n", __func__)
		return newStmt(rangeStmt)
	}
	var forStmt = &ast.ForStmt{}
	forStmt.Init = s1
	forStmt.Cond = makeExpr(s2)
	forStmt.Post = s3
	forStmt.Body = body
	p.closeScope()
	logf(" end %s\n", __func__)
	return newStmt(forStmt)
}

func (p *parser) parseIfStmt() ast.Stmt {
	p.expect("if", __func__)
	parserExprLev = -1
	var condStmt ast.Stmt = p.parseSimpleStmt(false)
	exprStmt := stmt2ExprStmt(condStmt)
	var cond = exprStmt.X
	parserExprLev = 0
	var body = p.parseBlockStmt()
	var else_ ast.Stmt
	if p.tok.tok == "else" {
		p.next()
		if p.tok.tok == "if" {
			else_ = p.parseIfStmt()
		} else {
			var elseblock = p.parseBlockStmt()
			p.expectSemi(__func__)
			else_ = newStmt(elseblock)
		}
	} else {
		p.expectSemi(__func__)
	}
	var ifStmt = &ast.IfStmt{}
	ifStmt.Cond = cond
	ifStmt.Body = body
	ifStmt.Else = else_

	return newStmt(ifStmt)
}

func (p *parser) parseCaseClause() *ast.CaseClause {
	logf(" [%s] start\n", __func__)
	var list []ast.Expr
	if p.tok.tok == "case" {
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
	logf(" [%s] end\n", __func__)
	return r
}

func isTypeSwitchAssert(x ast.Expr) bool {
	return isExprTypeAssertExpr(x) && expr2TypeAssertExpr(x).Type == nil
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
	for p.tok.tok == "case" || p.tok.tok == "default" {
		cc = p.parseCaseClause()
		ccs = newStmt(cc)
		list = append(list, ccs)
	}
	p.expect("}", __func__)
	p.expectSemi(__func__)
	var body = &ast.BlockStmt{}
	body.List = list

	typeSwitch := isTypeSwitchGuard(s2)

	p.closeScope()
	if typeSwitch {
		return newStmt(&ast.TypeSwitchStmt{
			Assign: s2,
			Body:   body,
		})
	} else {
		return newStmt(&ast.SwitchStmt{
			Body: body,
			Tag:  makeExpr(s2),
		})
	}
}

func (p *parser) parseLhsList() []ast.Expr {
	logf(" [%s] start\n", __func__)
	var list = p.parseExprList()
	logf(" end %s\n", __func__)
	return list
}

func (p *parser) parseSimpleStmt(isRangeOK bool) ast.Stmt {
	logf(" begin %s\n", __func__)
	var x = p.parseLhsList()
	var stok = p.tok.tok
	var isRange = false
	var y ast.Expr
	var rangeX ast.Expr
	var rangeUnary *ast.UnaryExpr
	switch stok {
	case ":=", "=":
		var assignToken = stok
		p.next() // consume =
		if isRangeOK && p.tok.tok == "range" {
			p.next() // consume "range"
			rangeX = p.parseRhs()
			rangeUnary = &ast.UnaryExpr{}
			rangeUnary.Op = "range"
			rangeUnary.X = rangeX
			y = (rangeUnary)
			isRange = true
		} else {
			y = p.parseExpr() // rhs
		}
		var as = &ast.AssignStmt{}
		as.Tok = assignToken
		as.Lhs = x
		as.Rhs = make([]ast.Expr, 1, 1)
		as.Rhs[0] = y
		as.IsRange = isRange
		s := newStmt(as)
		if as.Tok == ":=" {
			lhss := x
			for _, lhs := range lhss {
				assert(isExprIdent(lhs), "should be ident", __func__)
				declare(as, p.topScope, ast.Var, expr2Ident(lhs))
			}
		}
		logf(" parseSimpleStmt end =, := %s\n", __func__)
		return s
	case ";":
		var exprStmt = &ast.ExprStmt{}
		exprStmt.X = x[0]
		logf(" parseSimpleStmt end ; %s\n", __func__)
		return newStmt(exprStmt)
	}

	switch stok {
	case "++", "--":
		var sInc = &ast.IncDecStmt{}
		sInc.X = x[0]
		sInc.Tok = stok
		p.next() // consume "++" or "--"
		return newStmt(sInc)
	}
	var exprStmt = &ast.ExprStmt{}
	exprStmt.X = x[0]
	logf(" parseSimpleStmt end (final) %s\n", __func__)
	return newStmt(exprStmt)
}

func (p *parser) parseStmt() ast.Stmt {
	logf("\n")
	logf(" = begin %s\n", __func__)
	var s ast.Stmt
	switch p.tok.tok {
	case "var":
		var genDecl = p.parseDecl("var")
		s = newStmt(&ast.DeclStmt{
			Decl: genDecl,
		})
		logf(" = end parseStmt()\n")
	case "IDENT", "*":
		s = p.parseSimpleStmt(false)
		p.expectSemi(__func__)
	case "return":
		s = p.parseReturnStmt()
	case "break", "continue":
		s = p.parseBranchStmt(p.tok.tok)
	case "if":
		s = p.parseIfStmt()
	case "switch":
		s = p.parseSwitchStmt()
	case "for":
		s = p.parseForStmt()
	default:
		panic2(__func__, "TBI 3:"+p.tok.tok)
	}
	logf(" = end parseStmt()\n")
	return s
}

func (p *parser) parseExprList() []ast.Expr {
	logf(" [%s] start\n", __func__)
	var list []ast.Expr
	var e = p.parseExpr()
	list = append(list, e)
	for p.tok.tok == "," {
		p.next() // consume ","
		e = p.parseExpr()
		list = append(list, e)
	}

	logf(" [%s] end\n", __func__)
	return list
}

func (p *parser) parseRhsList() []ast.Expr {
	var list = p.parseExprList()
	return list
}

func (p *parser) parseBranchStmt(tok string) ast.Stmt {
	p.expect(tok, __func__)

	p.expectSemi(__func__)

	var branchStmt = &ast.BranchStmt{}
	branchStmt.Tok = tok
	return newStmt(branchStmt)
}

func (p *parser) parseReturnStmt() ast.Stmt {
	p.expect("return", __func__)
	var x []ast.Expr
	if p.tok.tok != ";" && p.tok.tok != "}" {
		x = p.parseRhsList()
	}
	p.expectSemi(__func__)
	var returnStmt = &ast.ReturnStmt{}
	returnStmt.Results = x
	return newStmt(returnStmt)
}

func (p *parser) parseStmtList() []ast.Stmt {
	var list []ast.Stmt
	for p.tok.tok != "}" && p.tok.tok != "EOF" && p.tok.tok != "case" && p.tok.tok != "default" {
		var stmt = p.parseStmt()
		list = append(list, stmt)
	}
	return list
}

func (p *parser) parseBody(scope *ast.Scope) *ast.BlockStmt {
	p.expect("{", __func__)
	p.topScope = scope
	logf(" begin parseStmtList()\n")
	var list = p.parseStmtList()
	logf(" end parseStmtList()\n")

	p.closeScope()
	p.expect("}", __func__)
	var r = &ast.BlockStmt{}
	r.List = list
	return r
}

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	p.expect("{", __func__)
	p.openScope()
	logf(" begin parseStmtList()\n")
	var list = p.parseStmtList()
	logf(" end parseStmtList()\n")
	p.closeScope()
	p.expect("}", __func__)
	var r = &ast.BlockStmt{}
	r.List = list
	return r
}

func (p *parser) parseDecl(keyword string) *ast.GenDecl {
	var r *ast.GenDecl
	switch p.tok.tok {
	case "var":
		p.expect(keyword, __func__)
		var ident = p.parseIdent()
		var typ = p.parseType()
		var value ast.Expr
		if p.tok.tok == "=" {
			p.next()
			value = p.parseExpr()
		}
		p.expectSemi(__func__)
		var valSpec = &ast.ValueSpec{}
		valSpec.Name = ident
		valSpec.Type = typ
		valSpec.Value = value
		declare(valSpec, p.topScope, ast.Var, ident)
		r = &ast.GenDecl{}
		r.Spec = valSpec
		return r
	default:
		panic2(__func__, "TBI\n")
	}
	return r
}

func (p *parser) parserTypeSpec() *ast.TypeSpec {
	logf(" [%s] start\n", __func__)
	p.expect("type", __func__)
	var ident = p.parseIdent()
	logf(" decl type %s\n", ident.Name)

	var spec = &ast.TypeSpec{}
	spec.Name = ident
	declare(spec, p.topScope, ast.Typ, ident)
	var typ = p.parseType()
	p.expectSemi(__func__)
	spec.Type = typ
	return spec
}

func (p *parser) parseValueSpec(keyword string) *ast.ValueSpec {
	logf(" [parserValueSpec] start\n")
	p.expect(keyword, __func__)
	var ident = p.parseIdent()
	logf(" var = %s\n", ident.Name)
	var typ = p.parseType()
	var value ast.Expr
	if p.tok.tok == "=" {
		p.next()
		value = p.parseExpr()
	}
	p.expectSemi(__func__)
	var spec = &ast.ValueSpec{}
	spec.Name = ident
	spec.Type = typ
	spec.Value = value
	var kind = ast.Con
	if keyword == "var" {
		kind = ast.Var
	}
	declare(spec, p.topScope, kind, ident)
	logf(" [parserValueSpec] end\n")
	return spec
}

func (p *parser) parseFuncDecl() ast.Decl {
	p.expect("func", __func__)
	var scope = ast.NewScope(p.topScope) // function scope
	var receivers *ast.FieldList
	if p.tok.tok == "(" {
		logf("  [parserFuncDecl] parsing method")
		receivers = p.parseParameters(scope, false)
	} else {
		logf("  [parserFuncDecl] parsing function")
	}
	var ident = p.parseIdent() // func name
	var sig = p.parseSignature(scope)
	var params = sig.Params
	var results = sig.Results
	if results == nil {
		logf(" [parserFuncDecl] %s sig.results is nil\n", ident.Name)
	} else {
		logf(" [parserFuncDecl] %s sig.results.List = %s\n", ident.Name, strconv.Itoa(len(sig.Results.List)))
	}
	var body *ast.BlockStmt
	if p.tok.tok == "{" {
		logf(" begin parseBody()\n")
		body = p.parseBody(scope)
		logf(" end parseBody()\n")
		p.expectSemi(__func__)
	} else {
		logf(" no function body\n")
		p.expectSemi(__func__)
	}
	var decl ast.Decl

	var funcDecl = &ast.FuncDecl{}
	funcDecl.Recv = receivers
	funcDecl.Name = ident
	funcDecl.Type = &ast.FuncType{}
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
	p.expect("package", __func__)
	p.unresolved = nil
	var ident = p.parseIdent()
	var packageName = ident.Name
	p.expectSemi(__func__)

	p.topScope = &ast.Scope{} // open scope
	p.pkgScope = p.topScope

	for p.tok.tok == "import" {
		p.expect("import", __func__)
		if p.tok.tok == "(" {
			p.next()
			for p.tok.tok != ")" {
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

	logf("\n")
	logf(" [parser] Parsing Top level decls\n")
	var decls []ast.Decl
	var decl ast.Decl

	for !importsOnly && p.tok.tok != "EOF" {
		switch p.tok.tok {
		case "var", "const":
			var spec = p.parseValueSpec(p.tok.tok)
			var genDecl = &ast.GenDecl{}
			genDecl.Spec = spec
			decl = genDecl
		case "func":
			logf("\n\n")
			decl = p.parseFuncDecl()
			//logf(" func decl parsed:%s\n", decl.funcDecl.Name.Name)
		case "type":
			var spec = p.parserTypeSpec()
			var genDecl = &ast.GenDecl{}
			genDecl.Spec = spec
			decl = genDecl
			logf(" type parsed:%s\n", "")
		default:
			panic2(__func__, "TBI:"+p.tok.tok)
		}
		decls = append(decls, decl)
	}

	p.topScope = nil

	// dump p.pkgScope
	logf("[DEBUG] Dump objects in the package scope\n")
	for _, oe := range p.pkgScope.Objects {
		logf("    object %s\n", oe.Name)
	}

	var unresolved []*ast.Ident
	logf(" [parserFile] resolving parser's unresolved (n=%s)\n", strconv.Itoa(len(p.unresolved)))
	for _, idnt := range p.unresolved {
		logf(" [parserFile] resolving ident %s ...\n", idnt.Name)
		ps := p.pkgScope
		var obj *ast.Object = ps.Lookup(idnt.Name)
		if obj != nil {
			logf(" resolved \n")
			idnt.Obj = obj
		} else {
			logf(" unresolved \n")
			unresolved = append(unresolved, idnt)
		}
	}
	logf(" [parserFile] Unresolved (n=%s)\n", strconv.Itoa(len(unresolved)))

	var f = &ast.File{}
	f.Name = packageName
	f.Scope = p.pkgScope
	f.Decls = decls
	f.Unresolved = unresolved
	f.Imports = p.imports
	logf(" [%s] end\n", __func__)
	return f
}

func parseImports(filename string) *ast.File {
	return parseFile(filename, true)
}

func readSource(filename string) []uint8 {
	return readFile(filename)
}

func parseFile(filename string, importsOnly bool) *ast.File {
	var text = readSource(filename)

	var p = &parser{}
	p.scanner = &scanner{}
	p.init(text)
	return p.parseFile(importsOnly)
}
