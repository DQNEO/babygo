package main

import (
	"os"
	"syscall"

	"github.com/DQNEO/babygo/lib/myfmt"
	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/path"
	"github.com/DQNEO/babygo/lib/strconv"
)

// --- foundation ---
var __func__ = "__func__"

func assert(bol bool, msg string, caller string) {
	if !bol {
		panic2(caller, msg)
	}
}

func throw(s string) {
	panic(s)
}

func panic2(caller string, x string) {
	panic("[" + caller + "] " + x)
}

var debugFrontEnd bool

func logf(format string, a ...interface{}) {
	if !debugFrontEnd {
		return
	}
	var f = "# " + format
	var s = myfmt.Sprintf(f, a...)
	syscall.Write(1, []uint8(s))
}

// --- libs ---
func inArray(x string, list []string) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}

// --- scanner ---
type scanner struct {
	src        []uint8
	ch         uint8
	offset     int
	nextOffset int
	insertSemi bool
}

func (s *scanner) next() {
	if s.nextOffset < len(s.src) {
		s.offset = s.nextOffset
		s.ch = s.src[s.offset]
		s.nextOffset++
	} else {
		s.offset = len(s.src)
		s.ch = 1 //EOF
	}
}

var keywords []string

func (s *scanner) Init(src []uint8) {
	// https://golang.org/ref/spec#Keywords
	keywords = []string{
		"break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var",
	}
	s.src = src
	s.offset = 0
	s.ch = ' '
	s.nextOffset = 0
	s.insertSemi = false
	logf("src len = %s\n", strconv.Itoa(len(s.src)))
	s.next()
}

func isLetter(ch uint8) bool {
	if ch == '_' {
		return true
	}
	return ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z')
}

func isDecimal(ch uint8) bool {
	return '0' <= ch && ch <= '9'
}

func (s *scanner) scanIdentifier() string {
	var offset = s.offset
	for isLetter(s.ch) || isDecimal(s.ch) {
		s.next()
	}
	return string(s.src[offset:s.offset])
}

func (s *scanner) scanNumber() string {
	var offset = s.offset
	for isDecimal(s.ch) {
		s.next()
	}
	return string(s.src[offset:s.offset])
}

func (s *scanner) scanString() string {
	var offset = s.offset - 1
	var escaped bool
	for !escaped && s.ch != '"' {
		if s.ch == '\\' {
			escaped = true
			s.next()
			s.next()
			escaped = false
			continue
		}
		s.next()
	}
	s.next() // consume ending '""
	return string(s.src[offset:s.offset])
}

func (s *scanner) scanChar() string {
	// '\'' opening already consumed
	var offset = s.offset - 1
	var ch uint8
	for {
		ch = s.ch
		s.next()
		if ch == '\'' {
			break
		}
		if ch == '\\' {
			s.next()
		}
	}

	return string(s.src[offset:s.offset])
}

func (s *scanner) scanComment() string {
	var offset = s.offset - 1
	for s.ch != '\n' {
		s.next()
	}
	return string(s.src[offset:s.offset])
}

type TokenContainer struct {
	pos int    // what's this ?
	tok string // token.Token
	lit string // raw data
}

// https://golang.org/ref/spec#Tokens
func (s *scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || (s.ch == '\n' && !s.insertSemi) || s.ch == '\r' {
		s.next()
	}
}

func (s *scanner) Scan() *TokenContainer {
	s.skipWhitespace()
	var tc = &TokenContainer{}
	var lit string
	var tok string
	var insertSemi bool
	var ch = s.ch
	if isLetter(ch) {
		lit = s.scanIdentifier()
		if inArray(lit, keywords) {
			tok = lit
			switch tok {
			case "break", "continue", "fallthrough", "return":
				insertSemi = true
			}
		} else {
			insertSemi = true
			tok = "IDENT"
		}
	} else if isDecimal(ch) {
		insertSemi = true
		lit = s.scanNumber()
		tok = "INT"
	} else {
		s.next()
		switch ch {
		case '\n':
			tok = ";"
			lit = "\n"
			insertSemi = false
		case '"': // double quote
			insertSemi = true
			lit = s.scanString()
			tok = "STRING"
		case '\'': // single quote
			insertSemi = true
			lit = s.scanChar()
			tok = "CHAR"
		// https://golang.org/ref/spec#Operators_and_punctuation
		//	+    &     +=    &=     &&    ==    !=    (    )
		//	-    |     -=    |=     ||    <     <=    [    ]
		//  *    ^     *=    ^=     <-    >     >=    {    }
		//	/    <<    /=    <<=    ++    =     :=    ,    ;
		//	%    >>    %=    >>=    --    !     ...   .    :
		//	&^          &^=
		case ':': // :=, :
			if s.ch == '=' {
				s.next()
				tok = ":="
			} else {
				tok = ":"
			}
		case '.': // ..., .
			var peekCh = s.src[s.nextOffset]
			if s.ch == '.' && peekCh == '.' {
				s.next()
				s.next()
				tok = "..."
			} else {
				tok = "."
			}
		case ',':
			tok = ","
		case ';':
			tok = ";"
			lit = ";"
		case '(':
			tok = "("
		case ')':
			insertSemi = true
			tok = ")"
		case '[':
			tok = "["
		case ']':
			insertSemi = true
			tok = "]"
		case '{':
			tok = "{"
		case '}':
			insertSemi = true
			tok = "}"
		case '+': // +=, ++, +
			switch s.ch {
			case '=':
				s.next()
				tok = "+="
			case '+':
				s.next()
				tok = "++"
				insertSemi = true
			default:
				tok = "+"
			}
		case '-': // -= --  -
			switch s.ch {
			case '-':
				s.next()
				tok = "--"
				insertSemi = true
			case '=':
				s.next()
				tok = "-="
			default:
				tok = "-"
			}
		case '*': // *=  *
			if s.ch == '=' {
				s.next()
				tok = "*="
			} else {
				tok = "*"
			}
		case '/':
			if s.ch == '/' {
				// comment
				// @TODO block comment
				if s.insertSemi {
					s.ch = '/'
					s.offset = s.offset - 1
					s.nextOffset = s.offset + 1
					tc.lit = "\n"
					tc.tok = ";"
					s.insertSemi = false
					return tc
				}
				lit = s.scanComment()
				tok = "COMMENT"
			} else if s.ch == '=' {
				tok = "/="
			} else {
				tok = "/"
			}
		case '%': // %= %
			if s.ch == '=' {
				s.next()
				tok = "%="
			} else {
				tok = "%"
			}
		case '^': // ^= ^
			if s.ch == '=' {
				s.next()
				tok = "^="
			} else {
				tok = "^"
			}
		case '<': //  <= <- <<= <<
			switch s.ch {
			case '-':
				s.next()
				tok = "<-"
			case '=':
				s.next()
				tok = "<="
			case '<':
				var peekCh = s.src[s.nextOffset]
				if peekCh == '=' {
					s.next()
					s.next()
					tok = "<<="
				} else {
					s.next()
					tok = "<<"
				}
			default:
				tok = "<"
			}
		case '>': // >= >>= >> >
			switch s.ch {
			case '=':
				s.next()
				tok = ">="
			case '>':
				var peekCh = s.src[s.nextOffset]
				if peekCh == '=' {
					s.next()
					s.next()
					tok = ">>="
				} else {
					s.next()
					tok = ">>"
				}
			default:
				tok = ">"
			}
		case '=': // == =
			if s.ch == '=' {
				s.next()
				tok = "=="
			} else {
				tok = "="
			}
		case '!': // !=, !
			if s.ch == '=' {
				s.next()
				tok = "!="
			} else {
				tok = "!"
			}
		case '&': // & &= && &^ &^=
			switch s.ch {
			case '=':
				s.next()
				tok = "&="
			case '&':
				s.next()
				tok = "&&"
			case '^':
				var peekCh = s.src[s.nextOffset]
				if peekCh == '=' {
					s.next()
					s.next()
					tok = "&^="
				} else {
					s.next()
					tok = "&^"
				}
			default:
				tok = "&"
			}
		case '|': // |= || |
			switch s.ch {
			case '|':
				s.next()
				tok = "||"
			case '=':
				s.next()
				tok = "|="
			default:
				tok = "|"
			}
		case 1:
			tok = "EOF"
		default:
			panic2(__func__, "unknown char:"+string([]uint8{ch})+":"+strconv.Itoa(int(ch)))
			tok = "UNKNOWN"
		}
	}
	tc.lit = lit
	tc.pos = 0
	tc.tok = tok
	s.insertSemi = insertSemi
	return tc
}

// --- ast ---
var astCon string = "Con"
var astTyp string = "Typ"
var astVar string = "Var"
var astFun string = "Fun"
var astPkg string = "Pkg"

type signature struct {
	params  *astFieldList
	results *astFieldList
}

type astObject struct {
	Kind     string
	Name     string
	Decl     interface{} // *astValueSpec|*astFuncDecl|*astTypeSpec|*astField|*astAssignStmt
	Variable *Variable
}

type astExpr struct {
	dtype         string
	ifc           interface{}
	ident         *astIdent
	arrayType     *astArrayType
	basicLit      *astBasicLit
	callExpr      *astCallExpr
	binaryExpr    *astBinaryExpr
	unaryExpr     *astUnaryExpr
	selectorExpr  *astSelectorExpr
	indexExpr     *astIndexExpr
	sliceExpr     *astSliceExpr
	starExpr      *astStarExpr
	parenExpr     *astParenExpr
	structType    *astStructType
	compositeLit  *astCompositeLit
	keyValueExpr  *astKeyValueExpr
	ellipsis      *astEllipsis
	interfaceType *astInterfaceType
	typeAssertExpr    *astTypeAssertExpr
}

type astField struct {
	Name   *astIdent
	Type   *astExpr
	Offset int
}

type astFieldList struct {
	List []*astField
}

type astIdent struct {
	Name string
	Obj  *astObject
}

type astEllipsis struct {
	Elt *astExpr
}

type astBasicLit struct {
	Kind  string // token.INT, token.CHAR, or token.STRING
	Value string
}

type astCompositeLit struct {
	Type *astExpr
	Elts []*astExpr
}

type astKeyValueExpr struct {
	Key *astExpr
	Value *astExpr
}

type astParenExpr struct {
	X *astExpr
}

type astSelectorExpr struct {
	X   *astExpr
	Sel *astIdent
}

type astIndexExpr struct {
	X     *astExpr
	Index *astExpr
}

type astSliceExpr struct {
	X      *astExpr
	Low    *astExpr
	High   *astExpr
	Max    *astExpr
	Slice3 bool
}

type astCallExpr struct {
	Fun      *astExpr   // function expression
	Args     []*astExpr // function arguments; or nil
	Ellipsis bool
}

type astStarExpr struct {
	X *astExpr
}

type astUnaryExpr struct {
	X  *astExpr
	Op string
}

type astBinaryExpr struct {
	X  *astExpr
	Y  *astExpr
	Op string
}

type astTypeAssertExpr struct {
	X *astExpr
	Type *astExpr // asserted type; nil means type switch X.(type)
}

// Type nodes
type astArrayType struct {
	Len *astExpr
	Elt *astExpr
}

type astStructType struct {
	Fields *astFieldList
}

type astInterfaceType struct {
	methods []string
}

type astFuncType struct {
	Params  *astFieldList
	Results *astFieldList
}

type astStmt struct {
	dtype          string
	DeclStmt       *astDeclStmt
	exprStmt       *astExprStmt
	blockStmt      *astBlockStmt
	assignStmt     *astAssignStmt
	returnStmt     *astReturnStmt
	ifStmt         *astIfStmt
	forStmt        *astForStmt
	incDecStmt     *astIncDecStmt
	isRange        bool
	rangeStmt      *astRangeStmt
	branchStmt     *astBranchStmt
	switchStmt     *astSwitchStmt
	typeSwitchStmt *astTypeSwitchStmt
	caseClause     *astCaseClause
}

type astDeclStmt struct {
	Decl astDecl
}

type astExprStmt struct {
	X *astExpr
}

type astIncDecStmt struct {
	X   *astExpr
	Tok string
}

type astAssignStmt struct {
	Lhs []*astExpr
	Tok string
	Rhs []*astExpr
}

type astReturnStmt struct {
	Results []*astExpr
	node    *nodeReturnStmt
}

type astBranchStmt struct {
	Tok        string
	Label      string
	currentFor *astStmt
}

type astBlockStmt struct {
	List []*astStmt
}

type astIfStmt struct {
	Init *astStmt
	Cond *astExpr
	Body *astBlockStmt
	Else *astStmt
}

type astCaseClause struct {
	List []*astExpr
	Body []*astStmt
}

type astSwitchStmt struct {
	Tag  *astExpr
	Body *astBlockStmt
	// lableExit string
}

type astTypeSwitchStmt struct {
	Assign *astStmt
	Body *astBlockStmt
	node *nodeTypeSwitchStmt
}

type nodeReturnStmt struct {
	fnc *Func
}

type nodeTypeSwitchStmt struct {
	subject         *astExpr
	subjectVariable *Variable
	assignIdent     *astIdent
	cases           []*TypeSwitchCaseClose
}

type TypeSwitchCaseClose struct {
	variable     *Variable
	variableType *Type
	orig         *astCaseClause
}

type astForStmt struct {
	Init      *astStmt
	Cond      *astExpr
	Post      *astStmt
	Body      *astBlockStmt
	Outer     *astStmt // outer loop
	labelPost string
	labelExit string
}


type astRangeStmt struct {
	Key       *astExpr
	Value     *astExpr
	X         *astExpr
	Body      *astBlockStmt
	Outer     *astStmt // outer loop
	labelPost string
	labelExit string
	lenvar    *Variable
	indexvar  *Variable
	Tok       string
}

type astImportSpec struct {
	Path string
}

type astValueSpec struct {
	Name  *astIdent
	Type  *astExpr
	Value *astExpr
}

type astTypeSpec struct {
	Name *astIdent
	Type *astExpr
}

// Pseudo interface for *ast.Decl
// *astGenDecl | *astFuncDecl
type astDecl interface {
}

type astSpec interface{}

type astGenDecl struct {
	Spec astSpec // *astValueSpec | *TypeSpec
}

type astFuncDecl struct {
	Recv *astFieldList
	Name *astIdent
	Type *astFuncType
	Body *astBlockStmt
}

type astFile struct {
	Name       string
	Imports []*astImportSpec
	Decls      []astDecl
	Unresolved []*astIdent
}

type astScope struct {
	Outer   *astScope
	Objects []*objectEntry
}

func astNewScope(outer *astScope) *astScope {
	return &astScope{
		Outer: outer,
	}
}

func scopeInsert(s *astScope, obj *astObject) {
	if s == nil {
		panic2(__func__, "s sholud not be nil\n")
	}

	s.Objects = append(s.Objects, &objectEntry{
		name: obj.Name,
		obj:  obj,
	})
}

func scopeLookup(s *astScope, name string) *astObject {
	for _, oe := range s.Objects {
		if oe.name == name {
			return oe.obj
		}
	}
	var r *astObject
	return r
}

// --- parser ---
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
	var readbytes = buf[0:n]
	return readbytes
}

func readSource(filename string) []uint8 {
	return readFile(filename)
}

func (p *parser) init(src []uint8) {
	var s = p.scanner
	s.Init(src)
	p.next()
}

type objectEntry struct {
	name string
	obj  *astObject
}

type parser struct {
	tok        *TokenContainer
	unresolved []*astIdent
	topScope   *astScope
	pkgScope   *astScope
	scanner    *scanner
	imports    []*astImportSpec
}

func (p *parser) openScope() {
	p.topScope = astNewScope(p.topScope)
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
		var s = myfmt.Sprintf("%s expected, but got %s", tok, p.tok.tok)
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

func (p *parser) parseIdent() *astIdent {
	var name string
	if p.tok.tok == "IDENT" {
		name = p.tok.lit
		p.next()
	} else {
		panic2(__func__, "IDENT expected, but got "+p.tok.tok)
	}
	logf(" [%s] ident name = %s\n", __func__, name)
	return &astIdent{
		Name: name,
	}
}

func (p *parser) parseImportSpec() *astImportSpec {
	var pth = p.tok.lit
	p.next()
	spec := &astImportSpec{
		Path: pth,
	}
	p.imports = append(p.imports, spec)
	return spec
}

func (p *parser) tryVarType(ellipsisOK bool) *astExpr {
	if ellipsisOK && p.tok.tok == "..." {
		p.next() // consume "..."
		var typ = p.tryIdentOrType()
		if typ != nil {
			p.resolve(typ)
		} else {
			panic2(__func__, "Syntax error")
		}

		return newExpr(&astEllipsis{
			Elt: typ,
		})
	}
	return p.tryIdentOrType()
}

func (p *parser) parseVarType(ellipsisOK bool) *astExpr {
	logf(" [%s] begin\n", __func__)
	var typ = p.tryVarType(ellipsisOK)
	if typ == nil {
		panic2(__func__, "nil is not expected")
	}
	logf(" [%s] end\n", __func__)
	return typ
}

func (p *parser) tryType() *astExpr {
	logf(" [%s] begin\n", __func__)
	var typ = p.tryIdentOrType()
	if typ != nil {
		p.resolve(typ)
	}
	logf(" [%s] end\n", __func__)
	return typ
}

func (p *parser) parseType() *astExpr {
	var typ = p.tryType()
	return typ
}

func (p *parser) parsePointerType() *astExpr {
	p.expect("*", __func__)
	var base = p.parseType()
	return newExpr(&astStarExpr{
		X: base,
	})
}

func (p *parser) parseArrayType() *astExpr {
	p.expect("[", __func__)
	var ln *astExpr
	if p.tok.tok != "]" {
		ln = p.parseRhs()
	}
	p.expect("]", __func__)
	var elt = p.parseType()

	return newExpr(&astArrayType{
		Elt : elt,
		Len : ln,
	})
}

func (p *parser) parseFieldDecl(scope *astScope) *astField {

	var varType = p.parseVarType(false)
	var typ = p.tryVarType(false)

	p.expectSemi(__func__)

	var field = &astField{
		Type : typ,
		Name : varType.ident,
	}
	declareField(field, scope, astVar, varType.ident)
	p.resolve(typ)
	return field
}

func (p *parser) parseStructType() *astExpr {
	p.expect("struct", __func__)
	p.expect("{", __func__)

	var _nil *astScope
	var scope = astNewScope(_nil)

	var list []*astField
	for p.tok.tok == "IDENT" || p.tok.tok == "*" {
		var field *astField = p.parseFieldDecl(scope)
		list = append(list, field)
	}
	p.expect("}", __func__)

	return newExpr(&astStructType{
		Fields: &astFieldList{
			List : list,
		},
	})
}

func (p *parser) parseTypeName() *astExpr {
	logf(" [%s] begin\n", __func__)
	var ident = p.parseIdent()
	if p.tok.tok == "." {
		// ident is a package name
		p.next() // consume "."
		eIdent := newExpr(ident)
		//p.resolve(eIdent)
		sel := p.parseIdent()
		selectorExpr := &astSelectorExpr{
			X:   eIdent,
			Sel: sel,
		}
		return newExpr(selectorExpr)
	}
	logf(" [%s] end\n", __func__)
	return newExpr(ident)
}

func (p *parser) tryIdentOrType() *astExpr {
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
		return newExpr(&astInterfaceType{
			methods: nil,
		})
	case "(":
		p.next()
		var _typ = p.parseType()
		p.expect(")", __func__)
		return newExpr(&astParenExpr{
			X: _typ,
		})
	case "type":
		p.next()
		return nil
	}

	return nil
}

func (p *parser) parseParameterList(scope *astScope, ellipsisOK bool) []*astField {
	logf(" [%s] begin\n", __func__)
	var list []*astExpr
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

	var params []*astField

	var typ = p.tryVarType(ellipsisOK)
	if typ != nil {
		if len(list) > 1 {
			panic2(__func__, "Ident list is not supported")
		}
		var eIdent = list[0]
		if eIdent.dtype != "*astIdent" {
			panic2(__func__, "Unexpected dtype")
		}
		var ident = eIdent.ident
		if ident == nil {
			panic2(__func__, "Ident should not be nil")
		}
		logf(" [%s] ident.Name=%s\n", __func__, ident.Name)
		logf(" [%s] typ=%s\n", __func__, typ.dtype)
		var field = &astField{}
		field.Name = ident
		field.Type = typ
		logf(" [%s]: Field %s %s\n", __func__, field.Name.Name, field.Type.dtype)
		params = append(params, field)
		declareField(field, scope, astVar, ident)
		p.resolve(typ)
		if p.tok.tok != "," {
			logf("  end %s\n", __func__)
			return params
		}
		p.next()
		for p.tok.tok != ")" && p.tok.tok != "EOF" {
			ident = p.parseIdent()
			typ = p.parseVarType(ellipsisOK)
			field = &astField{
				Name : ident,
				Type : typ,
			}
			params = append(params, field)
			declareField(field, scope, astVar, ident)
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
	params = make([]*astField, len(list), len(list))

	for i, typ := range list {
		p.resolve(typ)
		params[i] = &astField{
			Type: typ,
		}
		logf(" [DEBUG] range i = %s\n", strconv.Itoa(i))
	}
	logf("  end %s\n", __func__)
	return params
}

func (p *parser) parseParameters(scope *astScope, ellipsisOk bool) *astFieldList {
	logf(" [%s] begin\n", __func__)
	var params []*astField
	p.expect("(", __func__)
	if p.tok.tok != ")" {
		params = p.parseParameterList(scope, ellipsisOk)
	}
	p.expect(")", __func__)
	logf(" [%s] end\n", __func__)
	return &astFieldList{
		List: params,
	}
}

func (p *parser) parseResult(scope *astScope) *astFieldList {
	logf(" [%s] begin\n", __func__)

	if p.tok.tok == "(" {
		var r = p.parseParameters(scope, false)
		logf(" [%s] end\n", __func__)
		return r
	}

	if p.tok.tok == "{" {
		logf(" [%s] end\n", __func__)
		var _r *astFieldList = nil
		return _r
	}
	var typ = p.tryType()
	var list []*astField
	if typ != nil {
		list = append(list, &astField{
			Type: typ,
		})
	}
	logf(" [%s] end\n", __func__)
	return &astFieldList{
		List: list,
	}
}

func (p *parser) parseSignature(scope *astScope) *signature {
	logf(" [%s] begin\n", __func__)
	var params *astFieldList
	var results *astFieldList
	params = p.parseParameters(scope, true)
	results = p.parseResult(scope)
	return &signature{
		params:  params,
		results: results,
	}
}

func declareField(decl *astField, scope *astScope, kind string, ident *astIdent) {
	// declare
	var obj = &astObject{
		Decl : decl,
		Name : ident.Name,
		Kind : kind,
	}

	ident.Obj = obj

	// scope insert
	if ident.Name != "_" {
		scopeInsert(scope, obj)
	}
}

func declare(decl interface{}, scope *astScope, kind string, ident *astIdent) {
	logf(" [declare] ident %s\n", ident.Name)

	//valSpec.Name.Obj
	var obj = &astObject{
		Decl : decl,
		Name : ident.Name,
		Kind : kind,
	}
	ident.Obj = obj

	// scope insert
	if ident.Name != "_" {
		scopeInsert(scope, obj)
	}
	logf(" [declare] end\n")

}

func (p *parser) resolve(x *astExpr) {
	p.tryResolve(x, true)
}
func (p *parser) tryResolve(x *astExpr, collectUnresolved bool) {
	if x.dtype != "*astIdent" {
		return
	}
	var ident = x.ident
	if ident.Name == "_" {
		return
	}

	var s *astScope
	for s = p.topScope; s != nil; s = s.Outer {
		var obj = scopeLookup(s, ident.Name)
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

func (p *parser) parseOperand() *astExpr {
	logf("   begin %s\n", __func__)
	switch p.tok.tok {
	case "IDENT":
		var ident = p.parseIdent()
		var eIdent = newExpr(ident)
		p.tryResolve(eIdent, true)
		logf("   end %s\n", __func__)
		return eIdent
	case "INT", "STRING", "CHAR":
		var basicLit = &astBasicLit{
			Kind : p.tok.tok,
			Value : p.tok.lit,
		}
		p.next()
		logf("   end %s\n", __func__)
		return newExpr(basicLit)
	case "(":
		p.next() // consume "("
		parserExprLev++
		var x = p.parseRhsOrType()
		parserExprLev--
		p.expect(")", __func__)
		return newExpr(&astParenExpr{
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

func (p *parser) parseRhsOrType() *astExpr {
	var x = p.parseExpr()
	return x
}

func (p *parser) parseCallExpr(fn *astExpr) *astExpr {
	p.expect("(", __func__)
	logf(" [parsePrimaryExpr] p.tok.tok=%s\n", p.tok.tok)
	var list []*astExpr
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
	return newExpr(&astCallExpr{
		Fun:  fn,
		Args: list,
		Ellipsis: ellipsis,
	})
}

var parserExprLev int // < 0: in control clause, >= 0: in expression

func (p *parser) parsePrimaryExpr() *astExpr {
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
				var sel = &astSelectorExpr{
					X : x,
					Sel : secondIdent,
				}
				if p.tok.tok == "(" {
					var fn = newExpr(sel)
					// string = x.ident.Name + "." + secondIdent
					x = p.parseCallExpr(fn)
					logf(" [parsePrimaryExpr] 741 p.tok.tok=%s\n", p.tok.tok)
				} else {
					logf("   end parsePrimaryExpr()\n")
					x = newExpr(sel)
				}
			case "(": // type assertion
				x = p.parseTypeAssertion(x)
			default:
				panic2(__func__, "Unexpected token:" + p.tok.tok)
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

func (p *parser) parseTypeAssertion(x *astExpr) *astExpr {
	p.expect("(", __func__)
	typ := p.parseType()
	p.expect(")", __func__)
	return newExpr(&astTypeAssertExpr{
		X:    x,
		Type: typ,
	})
}

func (p *parser) parseElement() *astExpr {
	var x = p.parseExpr() // key or value
	var v *astExpr
	var kvExpr *astKeyValueExpr
	if p.tok.tok == ":" {
		p.next() // skip ":"
		v = p.parseExpr()
		kvExpr = &astKeyValueExpr{
			Key : x,
			Value : v,
		}
		x = newExpr(kvExpr)
	}
	return x
}

func (p *parser) parseElementList() []*astExpr {
	var list []*astExpr
	var e *astExpr
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

func (p *parser) parseLiteralValue(typ *astExpr) *astExpr {
	logf("   start %s\n", __func__)
	p.expect("{", __func__)
	var elts []*astExpr
	if p.tok.tok != "}" {
		elts = p.parseElementList()
	}
	p.expect("}", __func__)

	logf("   end %s\n", __func__)
	return  newExpr(&astCompositeLit{
			Type: typ,
			Elts: elts,
	})
}

func dtypeOf(x interface{}) string {
	switch xx := x.(type) {
	case *astExpr:
		return xx.dtype
	case *astStmt:
		return xx.dtype
	}

	panic("Unexpected")
}

const pastIdent string = "*astIdent"

func isLiteralType(x *astExpr) bool {
	switch dtypeOf(x) {
	case pastIdent:
	case "*astSelectorExpr":
		return x.selectorExpr.X.dtype == "*astIdent"
	case "*astArrayType":
	case "*astStructType":
	case "*astMapType":
	default:
		return false
	}

	return true
}

func (p *parser) parseIndexOrSlice(x *astExpr) *astExpr {
	p.expect("[", __func__)
	var index = make([]*astExpr, 3, 3)
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
		var sliceExpr = &astSliceExpr{
			Slice3 : false,
			X : x,
			Low : index[0],
			High : index[1],
		}
		if ncolons == 2 {
			sliceExpr.Max = index[2]
		}
		return &astExpr{
			dtype:     "*astSliceExpr",
			sliceExpr: sliceExpr,
		}
	}

	var indexExpr = &astIndexExpr{}
	indexExpr.X = x
	indexExpr.Index = index[0]
	var r = &astExpr{}
	r.dtype = "*astIndexExpr"
	r.indexExpr = indexExpr
	return r
}

func (p *parser) parseUnaryExpr() *astExpr {
	var r *astExpr
	logf("   begin parseUnaryExpr()\n")
	switch p.tok.tok {
	case "+", "-", "!", "&":
		var tok = p.tok.tok
		p.next()
		var x = p.parseUnaryExpr()
		r = &astExpr{}
		r.dtype = "*astUnaryExpr"
		r.unaryExpr = &astUnaryExpr{}
		logf(" [DEBUG] unary op = %s\n", tok)
		r.unaryExpr.Op = tok
		r.unaryExpr.X = x
		return r
	case "*":
		p.next() // consume "*"
		var x = p.parseUnaryExpr()
		r = &astExpr{}
		r.dtype = "*astStarExpr"
		r.starExpr = &astStarExpr{}
		r.starExpr.X = x
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

func (p *parser) parseBinaryExpr(prec1 int) *astExpr {
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
		var binaryExpr = &astBinaryExpr{}
		binaryExpr.X = x
		binaryExpr.Y = y
		binaryExpr.Op = op
		var r = &astExpr{}
		r.dtype = "*astBinaryExpr"
		r.binaryExpr = binaryExpr
		x = r
	}
	logf("   end parseBinaryExpr()\n")
	return x
}

func (p *parser) parseExpr() *astExpr {
	logf("   begin p.parseExpr()\n")
	var e = p.parseBinaryExpr(1)
	logf("   end p.parseExpr()\n")
	return e
}

func (p *parser) parseRhs() *astExpr {
	var x = p.parseExpr()
	return x
}

// Extract Expr from ExprStmt. Returns nil if input is nil
func makeExpr(s *astStmt) *astExpr {
	logf(" begin %s\n", __func__)
	if s == nil {
		var r *astExpr
		return r
	}
	if s.dtype != "*astExprStmt" {
		panic2(__func__, "unexpected dtype="+s.dtype)
	}
	if s.exprStmt == nil {
		panic2(__func__, "exprStmt is nil")
	}
	return s.exprStmt.X
}

func (p *parser) parseForStmt() *astStmt {
	logf(" begin %s\n", __func__)
	p.expect("for", __func__)
	p.openScope()

	var s1 *astStmt
	var s2 *astStmt
	var s3 *astStmt
	var isRange bool
	parserExprLev = -1
	if p.tok.tok != "{" {
		if p.tok.tok != ";" {
			s2 = p.parseSimpleStmt(true)
			isRange = s2.isRange
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

	var as *astAssignStmt
	var rangeX *astExpr
	if isRange {
		assert(s2.dtype == "*astAssignStmt", "type mismatch", __func__)
		as = s2.assignStmt
		logf(" [DEBUG] range as len lhs=%s\n", strconv.Itoa(len(as.Lhs)))
		var key *astExpr
		var value *astExpr
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

		rangeX = as.Rhs[0].unaryExpr.X
		var rangeStmt = &astRangeStmt{}
		rangeStmt.Key = key
		rangeStmt.Value = value
		rangeStmt.X = rangeX
		rangeStmt.Body = body
		rangeStmt.Tok = as.Tok
		var r = &astStmt{}
		r.dtype = "*astRangeStmt"
		r.rangeStmt = rangeStmt
		p.closeScope()
		logf(" end %s\n", __func__)
		return r
	}
	var forStmt = &astForStmt{}
	forStmt.Init = s1
	forStmt.Cond = makeExpr(s2)
	forStmt.Post = s3
	forStmt.Body = body
	var r = &astStmt{}
	r.dtype = "*astForStmt"
	r.forStmt = forStmt
	p.closeScope()
	logf(" end %s\n", __func__)
	return r
}

func (p *parser) parseIfStmt() *astStmt {
	p.expect("if", __func__)
	parserExprLev = -1
	var condStmt *astStmt = p.parseSimpleStmt(false)
	if condStmt.dtype != "*astExprStmt" {
		panic2(__func__, "unexpected dtype="+condStmt.dtype)
	}
	var cond = condStmt.exprStmt.X
	parserExprLev = 0
	var body = p.parseBlockStmt()
	var else_ *astStmt
	if p.tok.tok == "else" {
		p.next()
		if p.tok.tok == "if" {
			else_ = p.parseIfStmt()
		} else {
			var elseblock = p.parseBlockStmt()
			p.expectSemi(__func__)
			else_ = &astStmt{}
			else_.dtype = "*astBlockStmt"
			else_.blockStmt = elseblock
		}
	} else {
		p.expectSemi(__func__)
	}
	var ifStmt = &astIfStmt{}
	ifStmt.Cond = cond
	ifStmt.Body = body
	ifStmt.Else = else_

	var r = &astStmt{}
	r.dtype = "*astIfStmt"
	r.ifStmt = ifStmt
	return r
}

func (p *parser) parseCaseClause() *astCaseClause {
	logf(" [%s] start\n", __func__)
	var list []*astExpr
	if p.tok.tok == "case" {
		p.next() // consume "case"
		list = p.parseRhsList()
	} else {
		p.expect("default", __func__)
	}
	p.expect(":", __func__)
	p.openScope()
	var body = p.parseStmtList()
	var r = &astCaseClause{}
	r.Body = body
	r.List = list
	p.closeScope()
	logf(" [%s] end\n", __func__)
	return r
}

func isTypeSwitchAssert(x *astExpr) bool {
	return x.dtype == "*astTypeAssertExpr" && x.typeAssertExpr.Type == nil
}

func isTypeSwitchGuard(s *astStmt) bool {
	switch dtypeOf(s) {
	case "*astExprStmt":
		if isTypeSwitchAssert(s.exprStmt.X) {
			return true
		}
	case "*astAssignStmt":
		if len(s.assignStmt.Lhs) == 1 && len(s.assignStmt.Rhs) == 1 && isTypeSwitchAssert(s.assignStmt.Rhs[0]) {
			return true
		}
	}
	return false
}

func (p *parser) parseSwitchStmt() *astStmt {
	p.expect("switch", __func__)
	p.openScope()

	var s2 *astStmt
	parserExprLev = -1
	s2 = p.parseSimpleStmt(false)
	parserExprLev = 0

	p.expect("{", __func__)
	var list []*astStmt
	var cc *astCaseClause
	var ccs *astStmt
	for p.tok.tok == "case" || p.tok.tok == "default" {
		cc = p.parseCaseClause()
		ccs = &astStmt{}
		ccs.dtype = "*astCaseClause"
		ccs.caseClause = cc
		list = append(list, ccs)
	}
	p.expect("}", __func__)
	p.expectSemi(__func__)
	var body = &astBlockStmt{}
	body.List = list

	typeSwitch := isTypeSwitchGuard(s2)

	p.closeScope()
	var s = &astStmt{}
	if typeSwitch {
		s.dtype = "*astTypeSwitchStmt"
		s.typeSwitchStmt = &astTypeSwitchStmt{
			Assign: s2,
			Body:   body,
		}
		return s
	} else {
		s.dtype = "*astSwitchStmt"
		s.switchStmt = &astSwitchStmt{
			Body: body,
			Tag:  makeExpr(s2),
		}
		return s
	}
}

func (p *parser) parseLhsList() []*astExpr {
	logf(" [%s] start\n", __func__)
	var list = p.parseExprList()
	logf(" end %s\n", __func__)
	return list
}

func (p *parser) parseSimpleStmt(isRangeOK bool) *astStmt {
	logf(" begin %s\n", __func__)
	var s = &astStmt{}
	var x = p.parseLhsList()
	var stok = p.tok.tok
	var isRange = false
	var y *astExpr
	var rangeX *astExpr
	var rangeUnary *astUnaryExpr
	switch stok {
	case ":=", "=":
		var assignToken = stok
		p.next() // consume =
		if isRangeOK && p.tok.tok == "range" {
			p.next() // consume "range"
			rangeX = p.parseRhs()
			rangeUnary = &astUnaryExpr{}
			rangeUnary.Op = "range"
			rangeUnary.X = rangeX
			y = &astExpr{}
			y.dtype = "*astUnaryExpr"
			y.unaryExpr = rangeUnary
			isRange = true
		} else {
			y = p.parseExpr() // rhs
		}
		var as = &astAssignStmt{}
		as.Tok = assignToken
		as.Lhs = x
		as.Rhs = make([]*astExpr, 1, 1)
		as.Rhs[0] = y
		s.dtype = "*astAssignStmt"
		s.assignStmt = as
		s.isRange = isRange
		if as.Tok == ":=" {
			lhss := x
			for _, lhs := range lhss {
				assert(lhs.dtype == "*astIdent", "should be ident", __func__)
				declare(as, p.topScope, astVar, lhs.ident)
			}
		}
		logf(" parseSimpleStmt end =, := %s\n", __func__)
		return s
	case ";":
		s.dtype = "*astExprStmt"
		var exprStmt = &astExprStmt{}
		exprStmt.X = x[0]
		s.exprStmt = exprStmt
		logf(" parseSimpleStmt end ; %s\n", __func__)
		return s
	}

	switch stok {
	case "++", "--":
		var s = &astStmt{}
		var sInc = &astIncDecStmt{}
		sInc.X = x[0]
		sInc.Tok = stok
		s.dtype = "*astIncDecStmt"
		s.incDecStmt = sInc
		p.next() // consume "++" or "--"
		return s
	}
	var exprStmt = &astExprStmt{}
	exprStmt.X = x[0]
	var r = &astStmt{}
	r.dtype = "*astExprStmt"
	r.exprStmt = exprStmt
	logf(" parseSimpleStmt end (final) %s\n", __func__)
	return r
}

func (p *parser) parseStmt() *astStmt {
	logf("\n")
	logf(" = begin %s\n", __func__)
	var s *astStmt
	switch p.tok.tok {
	case "var":
		var genDecl = p.parseDecl("var")
		s = &astStmt{}
		s.dtype = "*astDeclStmt"
		s.DeclStmt = &astDeclStmt{}
		s.DeclStmt.Decl = genDecl
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

func (p *parser) parseExprList() []*astExpr {
	logf(" [%s] start\n", __func__)
	var list []*astExpr
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

func (p *parser) parseRhsList() []*astExpr {
	var list = p.parseExprList()
	return list
}

func (p *parser) parseBranchStmt(tok string) *astStmt {
	p.expect(tok, __func__)

	p.expectSemi(__func__)

	var branchStmt = &astBranchStmt{}
	branchStmt.Tok = tok
	var s = &astStmt{}
	s.dtype = "*astBranchStmt"
	s.branchStmt = branchStmt
	return s
}

func (p *parser) parseReturnStmt() *astStmt {
	p.expect("return", __func__)
	var x []*astExpr
	if p.tok.tok != ";" && p.tok.tok != "}" {
		x = p.parseRhsList()
	}
	p.expectSemi(__func__)
	var returnStmt = &astReturnStmt{}
	returnStmt.Results = x
	var r = &astStmt{}
	r.dtype = "*astReturnStmt"
	r.returnStmt = returnStmt
	return r
}

func (p *parser) parseStmtList () []*astStmt {
	var list []*astStmt
	for p.tok.tok != "}" && p.tok.tok != "EOF" && p.tok.tok != "case" && p.tok.tok != "default" {
		var stmt = p.parseStmt()
		list = append(list, stmt)
	}
	return list
}

func (p *parser) parseBody(scope *astScope) *astBlockStmt {
	p.expect("{", __func__)
	p.topScope = scope
	logf(" begin parseStmtList()\n")
	var list = p.parseStmtList()
	logf(" end parseStmtList()\n")

	p.closeScope()
	p.expect("}", __func__)
	var r = &astBlockStmt{}
	r.List = list
	return r
}

func (p *parser) parseBlockStmt() *astBlockStmt {
	p.expect("{", __func__)
	p.openScope()
	logf(" begin parseStmtList()\n")
	var list = p.parseStmtList()
	logf(" end parseStmtList()\n")
	p.closeScope()
	p.expect("}", __func__)
	var r = &astBlockStmt{}
	r.List = list
	return r
}

func (p *parser) parseDecl(keyword string) *astGenDecl {
	var r *astGenDecl
	switch p.tok.tok {
	case "var":
		p.expect(keyword, __func__)
		var ident = p.parseIdent()
		var typ = p.parseType()
		var value *astExpr
		if p.tok.tok == "=" {
			p.next()
			value = p.parseExpr()
		}
		p.expectSemi(__func__)
		var valSpec = &astValueSpec{}
		valSpec.Name = ident
		valSpec.Type = typ
		valSpec.Value = value
		declare(valSpec, p.topScope, astVar, ident)
		r = &astGenDecl{}
		r.Spec = valSpec
		return r
	default:
		panic2(__func__, "TBI\n")
	}
	return r
}

func (p *parser) parserTypeSpec() *astTypeSpec {
	logf(" [%s] start\n", __func__)
	p.expect("type", __func__)
	var ident = p.parseIdent()
	logf(" decl type %s\n", ident.Name)

	var spec = &astTypeSpec{}
	spec.Name = ident
	declare(spec, p.topScope, astTyp, ident)
	var typ = p.parseType()
	p.expectSemi(__func__)
	spec.Type = typ
	return spec
}

func (p *parser) parseValueSpec(keyword string) *astValueSpec {
	logf(" [parserValueSpec] start\n")
	p.expect(keyword, __func__)
	var ident = p.parseIdent()
	logf(" var = %s\n", ident.Name)
	var typ = p.parseType()
	var value *astExpr
	if p.tok.tok == "=" {
		p.next()
		value = p.parseExpr()
	}
	p.expectSemi(__func__)
	var spec = &astValueSpec{}
	spec.Name = ident
	spec.Type = typ
	spec.Value = value
	var kind = astCon
	if keyword == "var" {
		kind = astVar
	}
	declare(spec, p.topScope, kind, ident)
	logf(" [parserValueSpec] end\n")
	return spec
}

func (p *parser) parseFuncDecl() astDecl {
	p.expect("func", __func__)
	var scope = astNewScope(p.topScope) // function scope
	var receivers *astFieldList
	if p.tok.tok == "(" {
		logf("  [parserFuncDecl] parsing method")
		receivers = p.parseParameters(scope, false)
	} else {
		logf("  [parserFuncDecl] parsing function")
	}
	var ident = p.parseIdent() // func name
	var sig = p.parseSignature(scope)
	var params = sig.params
	var results = sig.results
	if results == nil {
		logf(" [parserFuncDecl] %s sig.results is nil\n", ident.Name)
	} else {
		logf(" [parserFuncDecl] %s sig.results.List = %s\n", ident.Name, strconv.Itoa(len(sig.results.List)))
	}
	var body *astBlockStmt
	if p.tok.tok == "{" {
		logf(" begin parseBody()\n")
		body = p.parseBody(scope)
		logf(" end parseBody()\n")
		p.expectSemi(__func__)
	} else {
		logf(" no function body\n")
		p.expectSemi(__func__)
	}
	var decl astDecl

	var funcDecl = &astFuncDecl{}
	funcDecl.Recv = receivers
	funcDecl.Name = ident
	funcDecl.Type = &astFuncType{}
	funcDecl.Type.Params = params
	funcDecl.Type.Results = results
	funcDecl.Body = body
	decl = funcDecl
	if receivers == nil {
		declare(funcDecl, p.pkgScope, astFun, ident)
	}
	return decl
}

func (p *parser) parseFile(importsOnly bool) *astFile {
	// expect "package" keyword
	p.expect("package", __func__)
	p.unresolved = nil
	var ident = p.parseIdent()
	var packageName = ident.Name
	p.expectSemi(__func__)

	p.topScope = &astScope{} // open scope
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
	var decls []astDecl
	var decl astDecl

	for !importsOnly && p.tok.tok != "EOF" {
		switch p.tok.tok {
		case "var", "const":
			var spec = p.parseValueSpec(p.tok.tok)
			var genDecl = &astGenDecl{}
			genDecl.Spec = spec
			decl = genDecl
		case "func":
			logf("\n\n")
			decl = p.parseFuncDecl()
			//logf(" func decl parsed:%s\n", decl.funcDecl.Name.Name)
		case "type":
			var spec = p.parserTypeSpec()
			var genDecl = &astGenDecl{}
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
		logf("    object %s\n", oe.name)
	}

	var unresolved []*astIdent
	logf(" [parserFile] resolving parser's unresolved (n=%s)\n", strconv.Itoa(len(p.unresolved)))
	for _, idnt := range p.unresolved {
		logf(" [parserFile] resolving ident %s ...\n", idnt.Name)
		var obj *astObject = scopeLookup(p.pkgScope, idnt.Name)
		if obj != nil {
			logf(" resolved \n")
			idnt.Obj = obj
		} else {
			logf(" unresolved \n")
			unresolved = append(unresolved, idnt)
		}
	}
	logf(" [parserFile] Unresolved (n=%s)\n", strconv.Itoa(len(unresolved)))

	var f = &astFile{}
	f.Name = packageName
	f.Decls = decls
	f.Unresolved = unresolved
	f.Imports = p.imports
	logf(" [%s] end\n", __func__)
	return f
}

func parseImports(filename string) *astFile {
	return parseFile(filename, true)
}

func parseFile(filename string, importsOnly bool) *astFile {
	var text = readSource(filename)

	var p = &parser{}
	p.scanner = &scanner{}
	p.init(text)
	return p.parseFile(importsOnly)
}

// --- codegen ---
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
	var format2 = string(spaces) + "# " + format
	var s = myfmt.Sprintf(format2, a...)
	syscall.Write(1, []uint8(s))
}

func evalInt(expr *astExpr) int {
	switch dtypeOf(expr) {
	case "*astBasicLit":
		return strconv.Atoi(expr.basicLit.Value)
	}
	return 0
}

func emitPopPrimitive(comment string) {
	myfmt.Printf("  popq %%rax # result of %s\n", comment)
}

func emitPopBool(comment string) {
	myfmt.Printf("  popq %%rax # result of %s\n", comment)
}

func emitPopAddress(comment string) {
	myfmt.Printf("  popq %%rax # address of %s\n", comment)
}

func emitPopString() {
	myfmt.Printf("  popq %%rax # string.ptr\n")
	myfmt.Printf("  popq %%rcx # string.len\n")
}

func emitPopInterFace() {
	myfmt.Printf("  popq %%rax # eface.dtype\n")
	myfmt.Printf("  popq %%rcx # eface.data\n")
}

func emitPopSlice() {
	myfmt.Printf("  popq %%rax # slice.ptr\n")
	myfmt.Printf("  popq %%rcx # slice.len\n")
	myfmt.Printf("  popq %%rdx # slice.cap\n")
}

func emitPushStackTop(condType *Type, comment string) {
	switch kind(condType) {
	case T_STRING:
		myfmt.Printf("  movq 8(%%rsp), %%rcx # copy str.len from stack top (%s)\n", comment)
		myfmt.Printf("  movq 0(%%rsp), %%rax # copy str.ptr from stack top (%s)\n", comment)
		myfmt.Printf("  pushq %%rcx # str.len\n")
		myfmt.Printf("  pushq %%rax # str.ptr\n")
	case T_POINTER, T_UINTPTR, T_BOOL, T_INT, T_UINT8, T_UINT16:
		myfmt.Printf("  movq (%%rsp), %%rax # copy stack top value (%s) \n", comment)
		myfmt.Printf("  pushq %%rax\n")
	default:
		throw(kind(condType))
	}
}

func emitRevertStackPointer(size int) {
	myfmt.Printf("  addq $%s, %%rsp # revert stack pointer\n", strconv.Itoa(size))
}

func emitAddConst(addValue int, comment string) {
	emitComment(2, "Add const: %s\n", comment)
	myfmt.Printf("  popq %%rax\n")
	myfmt.Printf("  addq $%s, %%rax\n", strconv.Itoa(addValue))
	myfmt.Printf("  pushq %%rax\n")
}

func emitLoad(t *Type) {
	if t == nil {
		panic2(__func__, "nil type error\n")
	}
	emitPopAddress(kind(t))
	switch kind(t) {
	case T_SLICE:
		myfmt.Printf("  movq %d(%%rax), %%rdx\n", strconv.Itoa(16))
		myfmt.Printf("  movq %d(%%rax), %%rcx\n", strconv.Itoa(8))
		myfmt.Printf("  movq %d(%%rax), %%rax\n", strconv.Itoa(0))
		myfmt.Printf("  pushq %%rdx # cap\n")
		myfmt.Printf("  pushq %%rcx # len\n")
		myfmt.Printf("  pushq %%rax # ptr\n")
	case T_STRING:
		myfmt.Printf("  movq %d(%%rax), %%rdx # len\n", strconv.Itoa(8))
		myfmt.Printf("  movq %d(%%rax), %%rax # ptr\n", strconv.Itoa(0))
		myfmt.Printf("  pushq %%rdx # len\n")
		myfmt.Printf("  pushq %%rax # ptr\n")
	case T_INTERFACE:
		myfmt.Printf("  movq %d(%%rax), %%rdx # data\n", strconv.Itoa(8))
		myfmt.Printf("  movq %d(%%rax), %%rax # dtype\n", strconv.Itoa(0))
		myfmt.Printf("  pushq %%rdx # data\n")
		myfmt.Printf("  pushq %%rax # dtype\n")
	case T_UINT8:
		myfmt.Printf("  movzbq %d(%%rax), %%rax # load uint8\n", strconv.Itoa(0))
		myfmt.Printf("  pushq %%rax\n")
	case T_UINT16:
		myfmt.Printf("  movzwq %d(%%rax), %%rax # load uint16\n", strconv.Itoa(0))
		myfmt.Printf("  pushq %%rax\n")
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		myfmt.Printf("  movq %d(%%rax), %%rax # load int\n", strconv.Itoa(0))
		myfmt.Printf("  pushq %%rax\n")
	case T_ARRAY, T_STRUCT:
		// pure proxy
		myfmt.Printf("  pushq %%rax\n")
	default:
		panic2(__func__, "TBI:kind="+kind(t))
	}
}

func emitVariableAddr(variable *Variable) {
	emitComment(2, "emit Addr of variable \"%s\" \n", variable.name)

	if variable.isGlobal {
		myfmt.Printf("  leaq %s(%%rip), %%rax # global variable \"%s\"\n", variable.globalSymbol,  variable.name)
	} else {
		myfmt.Printf("  leaq %d(%%rbp), %%rax # local variable \"%s\"\n", strconv.Itoa(variable.localOffset),  variable.name)
	}

	myfmt.Printf("  pushq %%rax # variable address\n")
}

func emitListHeadAddr(list *astExpr) {
	var t = getTypeOfExpr(list)
	switch kind(t) {
	case T_ARRAY:
		emitAddr(list) // array head
	case T_SLICE:
		emitExpr(list, nil)
		emitPopSlice()
		myfmt.Printf("  pushq %%rax # slice.ptr\n")
	case T_STRING:
		emitExpr(list, nil)
		emitPopString()
		myfmt.Printf("  pushq %%rax # string.ptr\n")
	default:
		panic2(__func__, "kind="+kind(getTypeOfExpr(list)))
	}
}

func emitAddr(expr *astExpr) {
	emitComment(2, "[emitAddr] %s\n", expr.dtype)
	switch dtypeOf(expr) {
	case "*astIdent":
		if expr.ident.Name == "_" {
			panic(" \"_\" has no address")
		}
		if expr.ident.Obj == nil {
			throw("expr.ident.Obj is nil: " + expr.ident.Name)
		}
		if expr.ident.Obj.Kind == astVar {
			assert(expr.ident.Obj.Variable != nil,
				"ERROR: Obj.Variable is not set for ident : "+expr.ident.Obj.Name, __func__)
			emitVariableAddr(expr.ident.Obj.Variable)
		} else {
			panic2(__func__, "Unexpected Kind "+expr.ident.Obj.Kind)
		}
	case "*astIndexExpr":
		emitExpr(expr.indexExpr.Index, nil) // index number
		var list = expr.indexExpr.X
		var elmType = getTypeOfExpr(expr)
		emitListElementAddr(list, elmType)
	case "*astStarExpr":
		emitExpr(expr.starExpr.X, nil)
	case "*astSelectorExpr": // (X).Sel
		var typeOfX = getTypeOfExpr(expr.selectorExpr.X)
		var structType *Type
		switch kind(typeOfX) {
		case T_STRUCT:
			// strct.field
			structType = typeOfX
			emitAddr(expr.selectorExpr.X)
		case T_POINTER:
			// ptr.field
			assert(typeOfX.e.dtype == "*astStarExpr", "should be *astStarExpr", __func__)
			var ptrType = typeOfX.e.starExpr
			structType = e2t(ptrType.X)
			emitExpr(expr.selectorExpr.X, nil)
		default:
			panic2(__func__, "TBI:"+kind(typeOfX))
		}
		var field = lookupStructField(getStructTypeSpec(structType), expr.selectorExpr.Sel.Name)
		var offset = getStructFieldOffset(field)
		emitAddConst(offset, "struct head address + struct.field offset")
	case "*astCompositeLit":
		var knd = kind(getTypeOfExpr(expr))
		switch knd {
		case T_STRUCT:
			// result of evaluation of a struct literal is its address
			emitExpr(expr, nil)
		default:
			panic2(__func__, "TBI "+ knd)
		}
	default:
		panic2(__func__, "TBI "+expr.dtype)
	}
}

func isType(expr *astExpr) bool {
	switch dtypeOf(expr) {
	case "*astArrayType":
		return true
	case "*astIdent":
		if expr.ident == nil {
			panic2(__func__, "ident should not be nil")
		}
		if expr.ident.Obj == nil {
			panic2(__func__, " unresolved ident:"+expr.ident.Name)
		}
		emitComment(2, "[isType][DEBUG] expr.ident.Name = %s\n", expr.ident.Name)
		emitComment(2, "[isType][DEBUG] expr.ident.Obj = %s,%s\n",
			expr.ident.Obj.Name, expr.ident.Obj.Kind)
		return expr.ident.Obj.Kind == astTyp
	case "*astParenExpr":
		return isType(expr.parenExpr.X)
	case "*astStarExpr":
		return isType(expr.starExpr.X)
	case "*astInterfaceType":
		return true
	default:
		emitComment(2, "[isType][%s] is not considered a type\n", expr.dtype)
	}

	return false

}

// explicit conversion T(e)
func emitConversion(toType *Type, arg0 *astExpr) {
	emitComment(2, "[emitConversion]\n")
	var to = toType.e
	switch dtypeOf(to) {
	case "*astIdent":
		switch to.ident.Obj {
		case gString: // string(e)
			switch kind(getTypeOfExpr(arg0)) {
			case T_SLICE: // string(slice)
				emitExpr(arg0, nil) // slice
				emitPopSlice()
				myfmt.Printf("  pushq %%rcx # str len\n")
				myfmt.Printf("  pushq %%rax # str ptr\n")
			}
		case gInt, gUint8, gUint16, gUintptr: // int(e)
			emitComment(2, "[emitConversion] to int \n")
			emitExpr(arg0, nil)
		default:
			if to.ident.Obj.Kind == astTyp {
				var isTypeSpec bool
				_, isTypeSpec = to.ident.Obj.Decl.(*astTypeSpec)
				if !isTypeSpec {
					panic2(__func__, "Something is wrong")
				}
				//e2t(to.ident.Obj.Decl.typeSpec.Type))
				emitExpr(arg0, nil)
			} else{
				panic2(__func__, "[*astIdent] TBI : "+to.ident.Obj.Name)
			}
		}
	case "*astArrayType": // Conversion to slice
		var arrayType = to.arrayType
		if arrayType.Len != nil {
			panic2(__func__, "internal error")
		}
		if (kind(getTypeOfExpr(arg0))) != T_STRING {
			panic2(__func__, "source type should be string")
		}
		emitComment(2, "Conversion of string => slice \n")
		emitExpr(arg0, nil)
		emitPopString()
		myfmt.Printf("  pushq %%rcx # cap\n")
		myfmt.Printf("  pushq %%rcx # len\n")
		myfmt.Printf("  pushq %%rax # ptr\n")
	case "*astParenExpr":
		emitConversion(e2t(to.parenExpr.X), arg0)
	case "*astStarExpr": // (*T)(e)
		emitComment(2, "[emitConversion] to pointer \n")
		emitExpr(arg0, nil)
	case "*astInterfaceType":
		emitExpr(arg0, nil)
		if isInterface(getTypeOfExpr(arg0))  {
			// do nothing
		} else {
			// Convert dynamic value to interface
			emitConvertToInterface(getTypeOfExpr(arg0))
		}
	default:
		panic2(__func__, "TBI :"+to.dtype)
	}
}

func emitZeroValue(t *Type) {
	switch kind(t) {
	case T_SLICE:
		myfmt.Printf("  pushq $0 # slice cap\n")
		myfmt.Printf("  pushq $0 # slice len\n")
		myfmt.Printf("  pushq $0 # slice ptr\n")
	case T_STRING:
		myfmt.Printf("  pushq $0 # string len\n")
		myfmt.Printf("  pushq $0 # string ptr\n")
	case T_INTERFACE:
		myfmt.Printf("  pushq $0 # interface data\n")
		myfmt.Printf("  pushq $0 # interface dtype\n")
	case T_INT, T_UINTPTR, T_UINT8, T_POINTER, T_BOOL:
		myfmt.Printf("  pushq $0 # %s zero value\n", kind(t))
	case T_STRUCT:
		var structSize = getSizeOfType(t)
		emitComment(2, "zero value of a struct. size=%s (allocating on heap)\n", strconv.Itoa(structSize))
		emitCallMalloc(structSize)
	default:
		panic2(__func__, "TBI:"+kind(t))
	}
}

func emitLen(arg *astExpr) {
	emitComment(2, "[%s] begin\n", __func__)
	switch kind(getTypeOfExpr(arg)) {
	case T_ARRAY:
		var typ = getTypeOfExpr(arg)
		var arrayType = typ.e.arrayType
		emitExpr(arrayType.Len, nil)
	case T_SLICE:
		emitExpr(arg, nil)
		emitPopSlice()
		myfmt.Printf("  pushq %%rcx # len\n")
	case T_STRING:
		emitExpr(arg, nil)
		emitPopString()
		myfmt.Printf("  pushq %%rcx # len\n")
	default:
		throw(kind(getTypeOfExpr(arg)))
	}
	emitComment(2, "[%s] end\n", __func__)
}

func emitCap(arg *astExpr) {
	switch kind(getTypeOfExpr(arg)) {
	case T_ARRAY:
		var typ = getTypeOfExpr(arg)
		var arrayType = typ.e.arrayType
		emitExpr(arrayType.Len, nil)
	case T_SLICE:
		emitExpr(arg, nil)
		emitPopSlice()
		myfmt.Printf("  pushq %%rdx # cap\n")
	case T_STRING:
		panic("cap() cannot accept string type")
	default:
		throw(kind(getTypeOfExpr(arg)))
	}
}

func emitCallMalloc(size int) {
	myfmt.Printf("  pushq $%s\n", strconv.Itoa(size))
	// call malloc and return pointer
	var resultList = []*astField{
		&astField{
			Type:    tUintptr.e,
		},
	}
	myfmt.Printf("  callq runtime.malloc\n") // no need to invert args orders
	emitRevertStackPointer(intSize)
	emitReturnedValue(resultList)
}

func emitStructLiteral(e *astCompositeLit) {
	// allocate heap area with zero value
	myfmt.Printf("  # Struct literal\n")
	var structType = e2t(e.Type)
	emitZeroValue(structType) // push address of the new storage
	var kvExpr *astKeyValueExpr
	for i, elm := range e.Elts {
		assert(elm.dtype == "*astKeyValueExpr", "wrong dtype 1:" + elm.dtype, __func__)
		kvExpr = elm.keyValueExpr
		assert(kvExpr.Key.dtype == "*astIdent", "wrong dtype 2:" + elm.dtype, __func__)
		var fieldName = kvExpr.Key.ident
		emitComment(2,"  - [%s] : key=%s, value=%s\n", strconv.Itoa(i), fieldName.Name, kvExpr.Value.dtype)
		var field = lookupStructField(getStructTypeSpec(structType), fieldName.Name)
		var fieldType = e2t(field.Type)
		var fieldOffset = getStructFieldOffset(field)
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

func emitArrayLiteral(arrayType *astArrayType, arrayLen int, elts []*astExpr) {
	var elmType = e2t(arrayType.Elt)
	var elmSize = getSizeOfType(elmType)
	var memSize = elmSize * arrayLen
	emitCallMalloc(memSize) // push
	for i, elm := range elts {
		// emit lhs
		emitPushStackTop(tUintptr, "malloced address")
		emitAddConst(elmSize*i, "malloced address + elmSize * index ("+strconv.Itoa(i)+")")
		ctx := &evalContext{
			_type: elmType,
		}
		emitExprIfc(elm, ctx)
		emitStore(elmType, true, false)
	}
}

func emitInvertBoolValue() {
	emitPopBool("")
	myfmt.Printf("  xor $1, %%rax\n")
	myfmt.Printf("  pushq %%rax\n")
}

func emitTrue() {
	myfmt.Printf("  pushq $1 # true\n")
}

func emitFalse() {
	myfmt.Printf("  pushq $0 # false\n")
}

type Arg struct {
	e      *astExpr
	t      *Type // expected type
	offset int
}

func emitArgs(args []*Arg) int {
	emitComment(2, "emitArgs len=%s\n", strconv.Itoa(len(args)))
	var totalPushedSize int
	//var arg *astExpr
	for _, arg := range args {
		var t *Type
		if arg.t != nil {
			t = arg.t
		} else {
			t = getTypeOfExpr(arg.e)
		}
		arg.offset = totalPushedSize
		totalPushedSize = totalPushedSize + getPushSizeOfType(t)
	}
	myfmt.Printf("  subq $%d, %%rsp # for args\n", strconv.Itoa(totalPushedSize))
	for _, arg := range args {
		ctx := &evalContext{
			_type: arg.t,
		}
		emitExprIfc(arg.e, ctx)
	}
	myfmt.Printf("  addq $%d, %%rsp # for args\n", strconv.Itoa(totalPushedSize))

	for _, arg := range args {
		var t *Type
		if arg.t != nil {
			t = arg.t
		} else {
			t = getTypeOfExpr(arg.e)
		}
		switch kind(t) {
		case T_BOOL, T_INT, T_UINT8, T_POINTER, T_UINTPTR:
			myfmt.Printf("  movq %d-8(%%rsp) , %%rax # load\n", strconv.Itoa(-arg.offset))
			myfmt.Printf("  movq %%rax, %d(%%rsp) # store\n", strconv.Itoa(+arg.offset))
		case T_STRING, T_INTERFACE:
			myfmt.Printf("  movq %d-16(%%rsp), %%rax\n", strconv.Itoa(-arg.offset))
			myfmt.Printf("  movq %d-8(%%rsp), %%rcx\n", strconv.Itoa(-arg.offset))
			myfmt.Printf("  movq %%rax, %d(%%rsp)\n", strconv.Itoa(+arg.offset))
			myfmt.Printf("  movq %%rcx, %d+8(%%rsp)\n", strconv.Itoa(+arg.offset))
		case T_SLICE:
			myfmt.Printf("  movq %d-24(%%rsp), %%rax\n", strconv.Itoa(-arg.offset)) // arg1: slc.ptr
			myfmt.Printf("  movq %d-16(%%rsp), %%rcx\n", strconv.Itoa(-arg.offset)) // arg1: slc.len
			myfmt.Printf("  movq %d-8(%%rsp), %%rdx\n", strconv.Itoa(-arg.offset))  // arg1: slc.cap
			myfmt.Printf("  movq %%rax, %d+0(%%rsp)\n", strconv.Itoa(+arg.offset))  // arg1: slc.ptr
			myfmt.Printf("  movq %%rcx, %d+8(%%rsp)\n", strconv.Itoa(+arg.offset))  // arg1: slc.len
			myfmt.Printf("  movq %%rdx, %d+16(%%rsp)\n", strconv.Itoa(+arg.offset)) // arg1: slc.cap
		default:
			throw(kind(t))
		}
	}

	return totalPushedSize
}

func prepareArgs(funcType *astFuncType, receiver *astExpr, eArgs []*astExpr, expandElipsis bool) []*Arg {
	if funcType == nil {
		panic("no funcType")
	}
	var params = funcType.Params.List
	var variadicArgs []*astExpr
	var variadicElmType *astExpr
	var args []*Arg
	var param *astField
	var arg *Arg
	lenParams := len(params)
	for argIndex, eArg := range eArgs {
		emitComment(2, "[%s][*astIdent][default] loop idx %s, len params %s\n", __func__, strconv.Itoa(argIndex), strconv.Itoa(lenParams))
		if argIndex < lenParams {
			param = params[argIndex]
			if param.Type.dtype == "*astEllipsis" {
				variadicElmType = param.Type.ellipsis.Elt
				variadicArgs = make([]*astExpr, 0, 20)
			}
		}
		if variadicElmType != nil && !expandElipsis {
			variadicArgs = append(variadicArgs, eArg)
			continue
		}

		var paramType = e2t(param.Type)
		arg = &Arg{}
		arg.e = eArg
		arg.t = paramType
		args = append(args, arg)
	}

	if variadicElmType != nil && !expandElipsis {
		// collect args as a slice
		var sliceType = &astArrayType{}
		sliceType.Elt = variadicElmType
		var eSliceType = newExpr(sliceType)
		var sliceLiteral = &astCompositeLit{}
		sliceLiteral.Type = eSliceType
		sliceLiteral.Elts = variadicArgs
		var _arg = &Arg{
			e : newExpr(sliceLiteral),
			t : e2t(eSliceType),
		}
		args = append(args, _arg)
	} else if len(args) < len(params) {
		// Add nil as a variadic arg
		emitComment(2, "len(args)=%s, len(params)=%s\n", strconv.Itoa(len(args)), strconv.Itoa(len(params)))
		var param = params[len(args)]
		if param == nil {
			panic2(__func__, "param should not be nil")
		}
		if param.Type == nil {
			panic2(__func__, "param.Type should not be nil")
		}
		assert(param.Type.dtype == "*astEllipsis", "internal error", __func__)

		var _arg = &Arg{}
		_arg.e = eNil
		_arg.t = e2t(param.Type)
		args = append(args, _arg)
	}

	if receiver != nil {
		var receiverAndArgs []*Arg = []*Arg{
			&Arg{
				e: receiver,
				t: getTypeOfExpr(receiver),
			},
		}

		for _, a := range args {
			receiverAndArgs = append(receiverAndArgs, a)
		}
		return receiverAndArgs
	}

	return args
}

func emitCall(symbol string, args []*Arg, results []*astField) {
	emitComment(2, "[%s] %s\n", __func__, symbol)
	var totalPushedSize = emitArgs(args)
	myfmt.Printf("  callq %s\n", symbol)
	emitRevertStackPointer(totalPushedSize)
	emitReturnedValue(results)
}

func emitReturnedValue(resultList []*astField) {
	switch len(resultList) {
	case 0:
		// do nothing
	case 1:
		var retval0 = resultList[0]
		var knd = kind(e2t(retval0.Type))
		switch knd {
		case T_STRING:
			myfmt.Printf("  pushq %%rdi # str len\n")
			myfmt.Printf("  pushq %%rax # str ptr\n")
		case T_INTERFACE:
			myfmt.Printf("  pushq %%rdi # ifc data\n")
			myfmt.Printf("  pushq %%rax # ifc dtype\n")
		case T_BOOL, T_UINT8, T_INT, T_UINTPTR, T_POINTER:
			myfmt.Printf("  pushq %%rax\n")
		case T_SLICE:
			myfmt.Printf("  pushq %%rsi # slice cap\n")
			myfmt.Printf("  pushq %%rdi # slice len\n")
			myfmt.Printf("  pushq %%rax # slice ptr\n")
		default:
			panic2(__func__, "Unexpected kind="+knd)
		}
	default:
		panic2(__func__,"multipul returned values is not supported ")
	}
}

func emitFuncall(fun *astExpr, eArgs []*astExpr, hasEllissis bool) {
	var symbol string
	var receiver *astExpr
	var funcType *astFuncType
	switch dtypeOf(fun) {
	case "*astIdent":
		emitComment(2, "[%s][*astIdent]\n", __func__)
		var fnIdent = fun.ident
		switch fnIdent.Obj {
		case gLen:
			var arg = eArgs[0]
			emitLen(arg)
			return
		case gCap:
			var arg = eArgs[0]
			emitCap(arg)
			return
		case gNew:
			var typeArg = e2t(eArgs[0])
			var size = getSizeOfType(typeArg)
			emitCallMalloc(size)
			return
		case gMake:
			var typeArg = e2t(eArgs[0])
			switch kind(typeArg) {
			case T_SLICE:
				// make([]T, ...)
				var arrayType = typeArg.e.arrayType
				//assert(ok, "should be *ast.ArrayType")
				var elmSize = getSizeOfType(e2t(arrayType.Elt))
				var numlit = newNumberLiteral(elmSize)
				var eNumLit = &astExpr{}
				eNumLit.dtype = "*astBasicLit"
				eNumLit.basicLit = numlit

				var args []*Arg = []*Arg{
					// elmSize
					&Arg{
						e: eNumLit,
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


				var resultList = []*astField{
					&astField{
						Type: generalSlice,
					},
				}
				emitCall("runtime.makeSlice", args, resultList)
				return
			default:
				panic2(__func__, "TBI")
			}

			return
		case gAppend:
			var sliceArg = eArgs[0]
			var elemArg = eArgs[1]
			var elmType = getElementTypeOfListType(getTypeOfExpr(sliceArg))
			var elmSize = getSizeOfType(elmType)

			var args []*Arg = []*Arg{
				// slice
				&Arg{
					e: sliceArg,
				},
				// element
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
				panic2(__func__, "Unexpected elmSize")
			}
			var resultList = []*astField{
				&astField{
					Type: generalSlice,
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

		var fn = fun.ident
		if fn.Name == "print" {
			emitExpr(eArgs[0], nil)
			myfmt.Printf("  callq runtime.printstring\n")
			myfmt.Printf("  addq $%s, %%rsp # revert \n", strconv.Itoa(16))
			return
		}

		if fn.Name == "makeSlice1" || fn.Name == "makeSlice8" || fn.Name == "makeSlice16" || fn.Name == "makeSlice24" {
			fn.Name = "makeSlice"
		}
		// general function call
		symbol = getPackageSymbol(pkg.name, fn.Name)
		emitComment(2, "[%s][*astIdent][default] start\n", __func__)
		if pkg.name == "os" && fn.Name == "runtime_args" {
			symbol = "runtime.runtime_args"
		} else if pkg.name == "os" && fn.Name == "runtime_getenv" {
			symbol = "runtime.runtime_getenv"
		}

		var obj = fn.Obj
		if obj.Decl == nil {
			panic2(__func__, "[*astCallExpr] decl is nil")
		}

		var fndecl *astFuncDecl // = decl.funcDecl
		var isFuncDecl bool
		fndecl, isFuncDecl = obj.Decl.(*astFuncDecl)
		if !isFuncDecl || fndecl == nil {
			panic2(__func__, "[*astCallExpr] fndecl is nil")
		}
		if fndecl.Type == nil {
			panic2(__func__, "[*astCallExpr] fndecl.Type is nil")
		}
		funcType = fndecl.Type
	case "*astSelectorExpr":
		var selectorExpr = fun.selectorExpr
		if selectorExpr.X.dtype != "*astIdent" {
			panic2(__func__, "TBI selectorExpr.X.dtype="+selectorExpr.X.dtype)
		}
		symbol = selectorExpr.X.ident.Name + "." + selectorExpr.Sel.Name
		switch symbol {
		case "unsafe.Pointer":
			emitExpr(eArgs[0], nil)
			return
		default:
			// Assume method call
			fn := selectorExpr
			xIdent := fn.X.ident
			if xIdent.Obj == nil {
				throw("xIdent.Obj  should not be nil:" + xIdent.Name)
			}
			if xIdent.Obj.Kind == astPkg {
				// pkg.Sel()
				funcdecl := lookupForeignFunc(xIdent.Name, fn.Sel.Name)
				funcType = funcdecl.Type
			} else {
				receiver = selectorExpr.X
				var receiverType = getTypeOfExpr(receiver)
				var method = lookupMethod(receiverType, selectorExpr.Sel)
				funcType = method.funcType
				symbol = getMethodSymbol(method)
			}
		}
	case "*astParenExpr":
		panic2(__func__, "[astParenExpr] TBI ")
	default:
		panic2(__func__, "TBI fun.dtype="+fun.dtype)
	}

	var args = prepareArgs(funcType, receiver, eArgs, hasEllissis)
	var resultList []*astField
	if funcType.Results != nil {
		resultList = funcType.Results.List
	}
	emitCall(symbol, args, resultList)
}

func emitNil(targetType *Type) {
	if targetType == nil {
		panic2(__func__, "Type is required to emit nil")
	}
	switch kind(targetType) {
	case T_SLICE, T_POINTER, T_INTERFACE:
		emitZeroValue(targetType)
	default:
		panic2(__func__, "Unexpected kind="+kind(targetType))
	}
}

func emitNamedConst(ident *astIdent, ctx *evalContext) {
	var valSpec *astValueSpec
	var ok bool
	valSpec, ok = ident.Obj.Decl.(*astValueSpec)
	assert(ok, "valSpec should not be nil", __func__)
	assert(valSpec != nil, "valSpec should not be nil", __func__)
	assert(valSpec.Value != nil, "valSpec should not be nil", __func__)
	assert(valSpec.Value.dtype == "*astBasicLit", "const value should be a literal", __func__)
	emitExprIfc(valSpec.Value, ctx)
}

type okContext struct {
	needMain bool
	needOk   bool
}

type evalContext struct {
	okContext *okContext
	_type     *Type
}

func emitExpr(e *astExpr, ctx *evalContext) bool {
	var isNilObject bool
	emitComment(2, "[emitExpr] dtype=%s\n", e.dtype)
	switch dtypeOf(e) {
	case "*astIdent":
		var ident = e.ident
		if ident.Obj == nil {
			panic2(__func__, "ident unresolved:"+ident.Name)
		}
		switch e.ident.Obj {
		case gTrue:
			emitTrue()
		case gFalse:
			emitFalse()
		case gNil:
			if ctx._type == nil {
				panic2(__func__, "context of nil is not passed")
			}
			emitNil(ctx._type)
			isNilObject = true
		default:
			switch ident.Obj.Kind {
			case astVar:
				emitAddr(e)
				var t = getTypeOfExpr(e)
				emitLoad(t)
			case astCon:
				emitNamedConst(ident, ctx)
			case astTyp:
				panic2(__func__, "[*astIdent] Kind Typ should not come here")
			default:
				panic2(__func__, "[*astIdent] unknown Kind="+ident.Obj.Kind+" Name="+ident.Obj.Name)
			}
		}
	case "*astIndexExpr":
		emitAddr(e)
		emitLoad(getTypeOfExpr(e))
	case "*astStarExpr":
		emitAddr(e)
		emitLoad(getTypeOfExpr(e))
	case "*astSelectorExpr":
		if e.selectorExpr.X.dtype == "*astIdent" && e.selectorExpr.X.ident.Obj.Kind == astPkg {
			ident := lookupForeignVar(e.selectorExpr.X.ident.Name, e.selectorExpr.Sel.Name)
			e := newExpr(ident)
			emitExpr(e, ctx)
		} else {
			emitAddr(e)
			emitLoad(getTypeOfExpr(e))
		}
	case "*astCallExpr":
		var fun = e.callExpr.Fun
		emitComment(2, "[%s][*astCallExpr]\n", __func__)
		if isType(fun) {
			emitConversion(e2t(fun), e.callExpr.Args[0])
		} else {
			emitFuncall(fun, e.callExpr.Args, e.callExpr.Ellipsis)
		}
	case "*astParenExpr":
		emitExpr(e.parenExpr.X, ctx)
	case "*astBasicLit":
		//		emitComment(0, "basicLit.Kind = %s \n", e.basicLit.Kind)
		switch e.basicLit.Kind {
		case "INT":
			var ival = strconv.Atoi(e.basicLit.Value)
			myfmt.Printf("  pushq $%d # number literal\n", strconv.Itoa(ival))
		case "STRING":
			var sl = getStringLiteral(e.basicLit)
			if sl.strlen == 0 {
				// zero value
				emitZeroValue(tString)
			} else {
				myfmt.Printf("  pushq $%d # str len\n", strconv.Itoa(sl.strlen))
				myfmt.Printf("  leaq %s, %%rax # str ptr\n", sl.label)
				myfmt.Printf("  pushq %%rax # str ptr\n")
			}
		case "CHAR":
			var val = e.basicLit.Value
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
			myfmt.Printf("  pushq $%d # convert char literal to int\n", strconv.Itoa(int(char)))
		default:
			panic2(__func__, "[*astBasicLit] TBI : "+e.basicLit.Kind)
		}
	case "*astSliceExpr":
		var list = e.sliceExpr.X
		var listType = getTypeOfExpr(list)

		// For convenience, any of the indices may be omitted.
		// A missing low index defaults to zero;
		var low *astExpr
		if e.sliceExpr.Low != nil {
			low = e.sliceExpr.Low
		} else {
			low = eZeroInt
		}

		// a missing high index defaults to the length of the sliced operand:
		// @TODO

		switch kind(listType) {
		case T_SLICE, T_ARRAY:
			if e.sliceExpr.Max == nil {
				// new cap = cap(operand) - low
				emitCap(e.sliceExpr.X)
				emitExpr(low, nil)
				myfmt.Printf("  popq %%rcx # low\n")
				myfmt.Printf("  popq %%rax # orig_cap\n")
				myfmt.Printf("  subq %%rcx, %%rax # orig_cap - low\n")
				myfmt.Printf("  pushq %%rax # new cap\n")

				// new len = high - low
				if e.sliceExpr.High != nil {
					emitExpr(e.sliceExpr.High, nil)
				} else {
					// high = len(orig)
					emitLen(e.sliceExpr.X)
				}
				emitExpr(low, nil)
				myfmt.Printf("  popq %%rcx # low\n")
				myfmt.Printf("  popq %%rax # high\n")
				myfmt.Printf("  subq %%rcx, %%rax # high - low\n")
				myfmt.Printf("  pushq %%rax # new len\n")
			} else {
				// new cap = max - low
				emitExpr(e.sliceExpr.Max, nil)
				emitExpr(low, nil)
				myfmt.Printf("  popq %%rcx # low\n")
				myfmt.Printf("  popq %%rax # max\n")
				myfmt.Printf("  subq %%rcx, %%rax # new cap = max - low\n")
				myfmt.Printf("  pushq %%rax # new cap\n")
				// new len = high - low
				emitExpr(e.sliceExpr.High, nil)
				emitExpr(low, nil)
				myfmt.Printf("  popq %%rcx # low\n")
				myfmt.Printf("  popq %%rax # high\n")
				myfmt.Printf("  subq %%rcx, %%rax # new len = high - low\n")
				myfmt.Printf("  pushq %%rax # new len\n")
			}
		case T_STRING:
			// new len = high - low
			if e.sliceExpr.High != nil {
				emitExpr(e.sliceExpr.High, nil)
			} else {
				emitLen(e.sliceExpr.X)
			}
			emitExpr(low, nil)
			myfmt.Printf("  popq %%rcx # low\n")
			myfmt.Printf("  popq %%rax # high\n")
			myfmt.Printf("  subq %%rcx, %%rax # high - low\n")
			myfmt.Printf("  pushq %%rax # len\n")
			// no cap
		default:
			panic2(__func__, "Unknown kind="+kind(listType))
		}

		emitExpr(low, nil)
		var elmType = getElementTypeOfListType(listType)
		emitListElementAddr(list, elmType)
	case "*astUnaryExpr":
		emitComment(2, "[DEBUG] unary op = %s\n", e.unaryExpr.Op)
		switch e.unaryExpr.Op {
		case "+":
			emitExpr(e.unaryExpr.X, nil)
		case "-":
			emitExpr(e.unaryExpr.X, nil)
			myfmt.Printf("  popq %%rax # e.X\n")
			myfmt.Printf("  imulq $-1, %%rax\n")
			myfmt.Printf("  pushq %%rax\n")
		case "&":
			emitAddr(e.unaryExpr.X)
		case "!":
			emitExpr(e.unaryExpr.X, nil)
			emitInvertBoolValue()
		default:
			panic2(__func__, "TBI:astUnaryExpr:"+e.unaryExpr.Op)
		}
	case "*astBinaryExpr":
		if kind(getTypeOfExpr(e.binaryExpr.X)) == T_STRING {
			var args []*Arg
			var argX = &Arg{}
			var argY = &Arg{}
			argX.e = e.binaryExpr.X
			argY.e = e.binaryExpr.Y
			args = append(args, argX)
			args = append(args, argY)
			switch e.binaryExpr.Op {
			case "+":
				var resultList = []*astField{
					&astField{
						Type:    tString.e,
					},
				}
				emitCall("runtime.catstrings", args, resultList)
			case "==":
				emitArgs(args)
				emitCompEq(getTypeOfExpr(e.binaryExpr.X))
			case "!=":
				emitArgs(args)
				emitCompEq(getTypeOfExpr(e.binaryExpr.X))
				emitInvertBoolValue()
			default:
				panic2(__func__, "[emitExpr][*astBinaryExpr] string : TBI T_STRING")
			}
		} else {
			switch e.binaryExpr.Op {
			case "&&":
				labelid++
				var labelExitWithFalse = myfmt.Sprintf(".L.%s.false", strconv.Itoa(labelid))
				var labelExit = myfmt.Sprintf(".L.%d.exit", strconv.Itoa(labelid))
				emitExpr(e.binaryExpr.X, nil) // left
				emitPopBool("left")
				myfmt.Printf("  cmpq $1, %%rax\n")
				// exit with false if left is false
				myfmt.Printf("  jne %s\n", labelExitWithFalse)

				// if left is true, then eval right and exit
				emitExpr(e.binaryExpr.Y, nil) // right
				myfmt.Printf("  jmp %s\n", labelExit)

				myfmt.Printf("  %s:\n", labelExitWithFalse)
				emitFalse()
				myfmt.Printf("  %s:\n", labelExit)
			case "||":
				labelid++
				var labelExitWithTrue = myfmt.Sprintf(".L.%d.true", strconv.Itoa(labelid))
				var labelExit = myfmt.Sprintf(".L.%d.exit", strconv.Itoa(labelid))
				emitExpr(e.binaryExpr.X, nil) // left
				emitPopBool("left")
				myfmt.Printf("  cmpq $1, %%rax\n")
				// exit with true if left is true
				myfmt.Printf("  je %s\n", labelExitWithTrue)

				// if left is false, then eval right and exit
				emitExpr(e.binaryExpr.Y, nil) // right
				myfmt.Printf("  jmp %s\n", labelExit)

				myfmt.Printf("  %s:\n", labelExitWithTrue)
				emitTrue()
				myfmt.Printf("  %s:\n", labelExit)
			case "+":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				myfmt.Printf("  popq %%rcx # right\n")
				myfmt.Printf("  popq %%rax # left\n")
				myfmt.Printf("  addq %%rcx, %%rax\n")
				myfmt.Printf("  pushq %%rax\n")
			case "-":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				myfmt.Printf("  popq %%rcx # right\n")
				myfmt.Printf("  popq %%rax # left\n")
				myfmt.Printf("  subq %%rcx, %%rax\n")
				myfmt.Printf("  pushq %%rax\n")
			case "*":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				myfmt.Printf("  popq %%rcx # right\n")
				myfmt.Printf("  popq %%rax # left\n")
				myfmt.Printf("  imulq %%rcx, %%rax\n")
				myfmt.Printf("  pushq %%rax\n")
			case "%":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				myfmt.Printf("  popq %%rcx # right\n")
				myfmt.Printf("  popq %%rax # left\n")
				myfmt.Printf("  movq $0, %%rdx # init %%rdx\n")
				myfmt.Printf("  divq %%rcx\n")
				myfmt.Printf("  movq %%rdx, %%rax\n")
				myfmt.Printf("  pushq %%rax\n")
			case "/":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				myfmt.Printf("  popq %%rcx # right\n")
				myfmt.Printf("  popq %%rax # left\n")
				myfmt.Printf("  movq $0, %%rdx # init %%rdx\n")
				myfmt.Printf("  divq %%rcx\n")
				myfmt.Printf("  pushq %%rax\n")
			case "==":
				var t = getTypeOfExpr(e.binaryExpr.X)
				emitExpr(e.binaryExpr.X, nil) // left
				ctx := &evalContext{_type: t}
				emitExpr(e.binaryExpr.Y, ctx)   // right
				emitCompEq(t)
			case "!=":
				var t = getTypeOfExpr(e.binaryExpr.X)
				emitExpr(e.binaryExpr.X, nil) // left
				ctx := &evalContext{_type: t}
				emitExpr(e.binaryExpr.Y, ctx)   // right
				emitCompEq(t)
				emitInvertBoolValue()
			case "<":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				emitCompExpr("setl")
			case "<=":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				emitCompExpr("setle")
			case ">":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				emitCompExpr("setg")
			case ">=":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				emitCompExpr("setge")
			default:
				panic2(__func__, "# TBI: binary operation for "+e.binaryExpr.Op)
			}
		}
	case "*astCompositeLit":
		// slice , array, map or struct
		var k = kind(e2t(e.compositeLit.Type))
		switch k {
		case T_STRUCT:
			emitStructLiteral(e.compositeLit)
		case T_ARRAY:
			assert(e.compositeLit.Type.dtype == "*astArrayType", "expect *ast.ArrayType", __func__)
			var arrayType = e.compositeLit.Type.arrayType
			var arrayLen = evalInt(arrayType.Len)
			emitArrayLiteral(arrayType, arrayLen, e.compositeLit.Elts)
		case T_SLICE:
			assert(e.compositeLit.Type.dtype == "*astArrayType", "expect *ast.ArrayType", __func__)
			var arrayType = e.compositeLit.Type.arrayType
			var length = len(e.compositeLit.Elts)
			emitArrayLiteral(arrayType, length, e.compositeLit.Elts)
			emitPopAddress("malloc")
			myfmt.Printf("  pushq $%d # slice.cap\n", strconv.Itoa(length))
			myfmt.Printf("  pushq $%d # slice.len\n", strconv.Itoa(length))
			myfmt.Printf("  pushq %%rax # slice.ptr\n")
		default:
			panic2(__func__, "Unexpected kind="+k)
		}
	case "*astTypeAssertExpr":
		emitExpr(e.typeAssertExpr.X, nil)
		myfmt.Printf("  popq  %%rax # ifc.dtype\n")
		myfmt.Printf("  popq  %%rcx # ifc.data\n")
		myfmt.Printf("  pushq %%rax # ifc.data\n")

		typ := e2t(e.typeAssertExpr.Type)
		sType := serializeType(typ)
		_id := getTypeId(sType)
		typeSymbol := typeIdToSymbol(_id)
		// check if type matches
		myfmt.Printf("  leaq %s(%%rip), %%rax # ifc.dtype\n", typeSymbol)
		myfmt.Printf("  pushq %%rax           # ifc.dtype\n")

		emitCompExpr("sete") // this pushes 1 or 0 in the end
		emitPopBool("type assertion ok value")
		myfmt.Printf("  cmpq $1, %%rax\n")

		labelid++
		labelTypeAssertionEnd := myfmt.Sprintf(".L.end_type_assertion.%d", strconv.Itoa(labelid))
		labelElse := myfmt.Sprintf(".L.unmatch.%d", strconv.Itoa(labelid))
		myfmt.Printf("  jne %s # jmp if false\n", labelElse)

		// if matched
		if ctx.okContext != nil {
			emitComment(2, " double value context\n")
			if ctx.okContext.needMain {
				emitExpr(e.typeAssertExpr.X, nil)
				myfmt.Printf("  popq %%rax # garbage\n")
				emitLoad(e2t(e.typeAssertExpr.Type)) // load dynamic data
			}
			if ctx.okContext.needOk {
				myfmt.Printf("  pushq $1 # ok = true\n")
			}
		} else {
			emitComment(2, " single value context\n")
			emitExpr(e.typeAssertExpr.X, nil)
			myfmt.Printf("  popq %%rax # garbage\n")
			emitLoad(e2t(e.typeAssertExpr.Type)) // load dynamic data
		}

		// exit
		myfmt.Printf("  jmp %s\n", labelTypeAssertionEnd)

		// if not matched
		myfmt.Printf("  %s:\n", labelElse)
		if ctx.okContext != nil {
			emitComment(2, " double value context\n")
			if ctx.okContext.needMain {
				emitZeroValue(typ)
			}
			if ctx.okContext.needOk {
				myfmt.Printf("  pushq $0 # ok = false\n")
			}
		} else {
			emitComment(2, " single value context\n")
			emitZeroValue(typ)
		}

		myfmt.Printf("  %s:\n", labelTypeAssertionEnd)
	default:
		panic2(__func__, "[emitExpr] `TBI:"+e.dtype)
	}

	return isNilObject
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

func emitExprIfc(expr *astExpr, ctx *evalContext) {
	isNilObj := emitExpr(expr, ctx)
	if !isNilObj && ctx != nil && ctx._type != nil && isInterface(ctx._type) && !isInterface(getTypeOfExpr(expr)) {
		emitConvertToInterface(getTypeOfExpr(expr))
	}
}

var typeMap []*typeEntry
type typeEntry struct {
	serialized string
	id int
}

var typeId int = 1

func typeIdToSymbol(id int) string {
	return "dtype." + strconv.Itoa(id)
}

func getTypeId(serialized string) int {
	for _, te := range typeMap{
		if te.serialized == serialized {
			return te.id
		}
	}
	r := typeId
	te := &typeEntry{
		serialized: serialized,
		id:         typeId,
	}
	typeMap = append(typeMap, te)
	typeId++
	return r
}

func emitDtypeSymbol(t *Type) {
	str := serializeType(t)
	typeId := getTypeId(str)
	typeSymbol := typeIdToSymbol(typeId)
	myfmt.Printf("  leaq %s(%%rip), %%rax # type symbol \"%s\"\n", typeSymbol, str)
	myfmt.Printf("  pushq %%rax           # type symbol\n")
}

func newNumberLiteral(x int) *astBasicLit {
	var r = &astBasicLit{}
	r.Kind = "INT"
	r.Value = strconv.Itoa(x)
	return r
}

func emitListElementAddr(list *astExpr, elmType *Type) {
	emitListHeadAddr(list)
	emitPopAddress("list head")
	myfmt.Printf("  popq %%rcx # index id\n")
	myfmt.Printf("  movq $%s, %%rdx # elm size\n", strconv.Itoa(getSizeOfType(elmType)))
	myfmt.Printf("  imulq %%rdx, %%rcx\n")
	myfmt.Printf("  addq %%rcx, %%rax\n")
	myfmt.Printf("  pushq %%rax # addr of element\n")
}

func emitCompEq(t *Type) {
	switch kind(t) {
	case T_STRING:
		var resultList = []*astField{
			&astField{
				Type:    tBool.e,
			},
		}
		myfmt.Printf("  callq runtime.cmpstrings\n")
		emitRevertStackPointer(stringSize * 2)
		emitReturnedValue(resultList)
	case T_INTERFACE:
		var resultList = []*astField{
			&astField{
				Type:    tBool.e,
			},
		}
		myfmt.Printf("  callq runtime.cmpinterface\n")
		emitRevertStackPointer(interfaceSize * 2)
		emitReturnedValue(resultList)
	case T_INT, T_UINT8, T_UINT16, T_UINTPTR, T_POINTER:
		emitCompExpr("sete")
	case T_SLICE:
		emitCompExpr("sete") // @FIXME this is not correct
	default:
		panic2(__func__, "Unexpected kind="+kind(t))
	}
}

//@TODO handle larger types than int
func emitCompExpr(inst string) {
	myfmt.Printf("  popq %%rcx # right\n")
	myfmt.Printf("  popq %%rax # left\n")
	myfmt.Printf("  cmpq %%rcx, %%rax\n")
	myfmt.Printf("  %s %%al\n", inst)
	myfmt.Printf("  movzbq %%al, %%rax\n") // true:1, false:0
	myfmt.Printf("  pushq %%rax\n")
}

func emitPop(knd string) {
	switch knd {
	case T_SLICE:
		emitPopSlice()
	case T_STRING:
		emitPopString()
	case T_INTERFACE:
		emitPopInterFace()
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		emitPopPrimitive(knd)
	case T_UINT16:
		emitPopPrimitive(knd)
	case T_UINT8:
		emitPopPrimitive(knd)
	case T_STRUCT, T_ARRAY:
		emitPopPrimitive(knd)
	default:
		panic("TBI:" + knd)
	}
}

func emitStore(t *Type, rhsTop bool, pushLhs bool) {
	knd := kind(t)
	emitComment(2, "emitStore(%s)\n", knd)
	if rhsTop {
		emitPop(knd) // rhs
		myfmt.Printf("  popq %%rsi # lhs addr\n")
	} else {
		myfmt.Printf("  popq %%rsi # lhs addr\n")
		emitPop(knd) // rhs
	}
	if pushLhs {
		myfmt.Printf("  pushq %%rsi # lhs addr\n")
	}
	switch knd {
	case T_SLICE:
		myfmt.Printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", strconv.Itoa(0))
		myfmt.Printf("  movq %%rcx, %d(%%rsi) # len to len\n", strconv.Itoa(8))
		myfmt.Printf("  movq %%rdx, %d(%%rsi) # cap to cap\n", strconv.Itoa(16))
	case T_STRING:
		myfmt.Printf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", strconv.Itoa(0))
		myfmt.Printf("  movq %%rcx, %d(%%rsi) # len to len\n", strconv.Itoa(8))
	case T_INTERFACE:
		myfmt.Printf("  movq %%rax, %d(%%rsi) # store dtype\n", strconv.Itoa(0))
		myfmt.Printf("  movq %%rcx, %d(%%rsi) # store data\n", strconv.Itoa(8))
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		myfmt.Printf("  movq %%rax, (%%rsi) # assign\n")
	case T_UINT16:
		myfmt.Printf("  movw %%ax, (%%rsi) # assign word\n")
	case T_UINT8:
		myfmt.Printf("  movb %%al, (%%rsi) # assign byte\n")
	case T_STRUCT, T_ARRAY:
		myfmt.Printf("  pushq $%d # size\n", strconv.Itoa(getSizeOfType(t)))
		myfmt.Printf("  pushq %%rsi # dst lhs\n")
		myfmt.Printf("  pushq %%rax # src rhs\n")
		myfmt.Printf("  callq runtime.memcopy\n")
		emitRevertStackPointer(ptrSize*2 + intSize)
	default:
		panic2(__func__, "TBI:"+kind(t))
	}
}

func isBlankIdentifier(e *astExpr) bool {
	if e.dtype != "*astIdent" {
		return false
	}
	return e.ident.Name == "_"
}

func emitAssignWithOK(lhss []*astExpr, rhs *astExpr) {
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
	}
	emitExprIfc(rhs, ctx) // {push data}, {push bool}
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

func emitAssign(lhs *astExpr, rhs *astExpr) {
	emitComment(2, "Assignment: emitAddr(lhs:%s)\n", lhs.dtype)
	emitAddr(lhs)
	emitComment(2, "Assignment: emitExpr(rhs)\n")
	ctx := &evalContext{
		_type: getTypeOfExpr(lhs),
	}
	emitExprIfc(rhs, ctx)
	emitStore(getTypeOfExpr(lhs), true, false)
}

func emitStmt(stmt *astStmt) {
	emitComment(2, "== Statement %s ==\n", stmt.dtype)
	switch dtypeOf(stmt) {
	case "*astBlockStmt":
		for _, stmt2 := range stmt.blockStmt.List {
			emitStmt(stmt2)
		}
	case "*astExprStmt":
		emitExpr(stmt.exprStmt.X, nil)
	case "*astDeclStmt":
		var decl astDecl = stmt.DeclStmt.Decl
		var genDecl *astGenDecl
		var isGenDecl bool
		genDecl, isGenDecl = decl.(*astGenDecl)
		if !isGenDecl {
			panic2(__func__, "[*astDeclStmt] internal error")
		}

		var valSpec *astValueSpec
		var ok bool
		valSpec, ok = genDecl.Spec.(*astValueSpec)
		assert(ok, "should be ok", __func__)
		var t = e2t(valSpec.Type)
		var ident = valSpec.Name
		var lhs = &astExpr{}
		lhs.dtype = "*astIdent"
		lhs.ident = ident
		var rhs *astExpr
		if valSpec.Value == nil {
			emitComment(2, "lhs addresss\n")
			emitAddr(lhs)
			emitComment(2, "emitZeroValue for %s\n", t.e.dtype)
			emitZeroValue(t)
			emitComment(2, "Assignment: zero value\n")
			emitStore(t, true, false)
		} else {
			rhs = valSpec.Value
			emitAssign(lhs, rhs)
		}

		//var valueSpec *astValueSpec = genDecl.Specs[0]
		//var obj *astObject = valueSpec.Name.Obj
		//var typ *astExpr = valueSpec.Type
		//mylib.Printf("[emitStmt] TBI declSpec:%s\n", valueSpec.Name.Name)
		//os.Exit(1)

	case "*astAssignStmt":
		switch stmt.assignStmt.Tok {
		case "=":
		case ":=":
		default:
		}
		var lhs = stmt.assignStmt.Lhs[0]
		var rhs = stmt.assignStmt.Rhs[0]
		if len(stmt.assignStmt.Lhs) == 2 && rhs.dtype == "*astTypeAssertExpr" {
			emitAssignWithOK(stmt.assignStmt.Lhs, rhs)
		} else {
			emitAssign(lhs, rhs)
		}
	case "*astReturnStmt":
		node := stmt.returnStmt.node
		funcType := node.fnc.funcType
		if len(stmt.returnStmt.Results) == 0 {
			myfmt.Printf("  leave\n")
			myfmt.Printf("  ret\n")
		} else if len(stmt.returnStmt.Results) == 1 {
			targetType := e2t(funcType.Results.List[0].Type)
			ctx := &evalContext{
				_type:     targetType,
			}
			emitExprIfc(stmt.returnStmt.Results[0], ctx)
			var knd = kind(targetType)
			switch knd {
			case T_BOOL, T_UINT8, T_INT, T_UINTPTR, T_POINTER:
				myfmt.Printf("  popq %%rax # return 64bit\n")
			case T_STRING, T_INTERFACE:
				myfmt.Printf("  popq %%rax # return string (head)\n")
				myfmt.Printf("  popq %%rdi # return string (tail)\n")
			case T_SLICE:
				myfmt.Printf("  popq %%rax # return string (head)\n")
				myfmt.Printf("  popq %%rdi # return string (body)\n")
				myfmt.Printf("  popq %%rsi # return string (tail)\n")
			default:
				panic2(__func__, "[*astReturnStmt] TBI:"+knd)
			}
			myfmt.Printf("  leave\n")
			myfmt.Printf("  ret\n")
		} else if len(stmt.returnStmt.Results) == 3 {
			// Special treatment to return a slice
			emitExpr(stmt.returnStmt.Results[2], nil) // @FIXME
			emitExpr(stmt.returnStmt.Results[1], nil) // @FIXME
			emitExpr(stmt.returnStmt.Results[0], nil) // @FIXME
			myfmt.Printf("  popq %%rax # return 64bit\n")
			myfmt.Printf("  popq %%rdi # return 64bit\n")
			myfmt.Printf("  popq %%rsi # return 64bit\n")
		} else {
			panic2(__func__, "[*astReturnStmt] TBI\n")
		}
	case "*astIfStmt":
		emitComment(2, "if\n")

		labelid++
		var labelEndif = ".L.endif." + strconv.Itoa(labelid)
		var labelElse = ".L.else." + strconv.Itoa(labelid)

		emitExpr(stmt.ifStmt.Cond, nil)
		emitPopBool("if condition")
		myfmt.Printf("  cmpq $1, %%rax\n")
		var bodyStmt = &astStmt{}
		bodyStmt.dtype = "*astBlockStmt"
		bodyStmt.blockStmt = stmt.ifStmt.Body
		if stmt.ifStmt.Else != nil {
			myfmt.Printf("  jne %s # jmp if false\n", labelElse)
			emitStmt(bodyStmt) // then
			myfmt.Printf("  jmp %s\n", labelEndif)
			myfmt.Printf("  %s:\n", labelElse)
			emitStmt(stmt.ifStmt.Else) // then
		} else {
			myfmt.Printf("  jne %s # jmp if false\n", labelEndif)
			emitStmt(bodyStmt) // then
		}
		myfmt.Printf("  %s:\n", labelEndif)
		emitComment(2, "end if\n")
	case "*astForStmt":
		labelid++
		var labelCond = ".L.for.cond." + strconv.Itoa(labelid)
		var labelPost = ".L.for.post." + strconv.Itoa(labelid)
		var labelExit = ".L.for.exit." + strconv.Itoa(labelid)
		//forStmt, ok := mapForNodeToFor[s]
		//assert(ok, "map value should exist")
		stmt.forStmt.labelPost = labelPost
		stmt.forStmt.labelExit = labelExit

		if stmt.forStmt.Init != nil {
			emitStmt(stmt.forStmt.Init)
		}

		myfmt.Printf("  %s:\n", labelCond)
		if stmt.forStmt.Cond != nil {
			emitExpr(stmt.forStmt.Cond, nil)
			emitPopBool("for condition")
			myfmt.Printf("  cmpq $1, %%rax\n")
			myfmt.Printf("  jne %s # jmp if false\n", labelExit)
		}
		emitStmt(blockStmt2Stmt(stmt.forStmt.Body))
		myfmt.Printf("  %s:\n", labelPost) // used for "continue"
		if stmt.forStmt.Post != nil {
			emitStmt(stmt.forStmt.Post)
		}
		myfmt.Printf("  jmp %s\n", labelCond)
		myfmt.Printf("  %s:\n", labelExit)
	case "*astRangeStmt": // only for array and slice
		labelid++
		var labelCond = ".L.range.cond." + strconv.Itoa(labelid)
		var labelPost = ".L.range.post." + strconv.Itoa(labelid)
		var labelExit = ".L.range.exit." + strconv.Itoa(labelid)

		stmt.rangeStmt.labelPost = labelPost
		stmt.rangeStmt.labelExit = labelExit
		// initialization: store len(rangeexpr)
		emitComment(2, "ForRange Initialization\n")

		emitComment(2, "  assign length to lenvar\n")
		// lenvar = len(s.X)
		emitVariableAddr(stmt.rangeStmt.lenvar)
		emitLen(stmt.rangeStmt.X)
		emitStore(tInt, true, false)

		emitComment(2, "  assign 0 to indexvar\n")
		// indexvar = 0
		emitVariableAddr(stmt.rangeStmt.indexvar)
		emitZeroValue(tInt)
		emitStore(tInt, true, false)

		// init key variable with 0
		if stmt.rangeStmt.Key != nil {
			assert(stmt.rangeStmt.Key.dtype == "*astIdent", "key expr should be an ident", __func__)
			var keyIdent = stmt.rangeStmt.Key.ident
			if keyIdent.Name != "_" {
				emitAddr(stmt.rangeStmt.Key) // lhs
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
		myfmt.Printf("  %s:\n", labelCond)

		emitVariableAddr(stmt.rangeStmt.indexvar)
		emitLoad(tInt)
		emitVariableAddr(stmt.rangeStmt.lenvar)
		emitLoad(tInt)
		emitCompExpr("setl")
		emitPopBool(" indexvar < lenvar")
		myfmt.Printf("  cmpq $1, %%rax\n")
		myfmt.Printf("  jne %s # jmp if false\n", labelExit)

		emitComment(2, "assign list[indexvar] value variables\n")
		var elemType = getTypeOfExpr(stmt.rangeStmt.Value)
		emitAddr(stmt.rangeStmt.Value) // lhs

		emitVariableAddr(stmt.rangeStmt.indexvar)
		emitLoad(tInt) // index value
		emitListElementAddr(stmt.rangeStmt.X, elemType)

		emitLoad(elemType)
		emitStore(elemType, true, false)

		// Body
		emitComment(2, "ForRange Body\n")
		emitStmt(blockStmt2Stmt(stmt.rangeStmt.Body))

		// Post statement: Increment indexvar and go next
		emitComment(2, "ForRange Post statement\n")
		myfmt.Printf("  %s:\n", labelPost)        // used for "continue"
		emitVariableAddr(stmt.rangeStmt.indexvar) // lhs
		emitVariableAddr(stmt.rangeStmt.indexvar) // rhs
		emitLoad(tInt)
		emitAddConst(1, "indexvar value ++")
		emitStore(tInt, true, false)

		if stmt.rangeStmt.Key != nil {
			assert(stmt.rangeStmt.Key.dtype == "*astIdent", "key expr should be an ident", __func__)
			var keyIdent = stmt.rangeStmt.Key.ident
			if keyIdent.Name != "_" {
				emitAddr(stmt.rangeStmt.Key)              // lhs
				emitVariableAddr(stmt.rangeStmt.indexvar) // rhs
				emitLoad(tInt)
				emitStore(tInt, true, false)
			}
		}

		myfmt.Printf("  jmp %s\n", labelCond)

		myfmt.Printf("  %s:\n", labelExit)

	case "*astIncDecStmt":
		var addValue int
		switch stmt.incDecStmt.Tok {
		case "++":
			addValue = 1
		case "--":
			addValue = -1
		default:
			panic2(__func__, "Unexpected Tok="+stmt.incDecStmt.Tok)
		}
		emitAddr(stmt.incDecStmt.X)
		emitExpr(stmt.incDecStmt.X, nil)
		emitAddConst(addValue, "rhs ++ or --")
		emitStore(getTypeOfExpr(stmt.incDecStmt.X), true, false)
	case "*astSwitchStmt":
		labelid++
		var labelEnd = myfmt.Sprintf(".L.switch.%s.exit", strconv.Itoa(labelid))
		if stmt.switchStmt.Tag == nil {
			panic2(__func__, "Omitted tag is not supported yet")
		}
		emitExpr(stmt.switchStmt.Tag, nil)
		var condType = getTypeOfExpr(stmt.switchStmt.Tag)
		var cases = stmt.switchStmt.Body.List
		emitComment(2, "[DEBUG] cases len=%s\n", strconv.Itoa(len(cases)))
		var labels = make([]string, len(cases), len(cases))
		var defaultLabel string
		emitComment(2, "Start comparison with cases\n")
		for i, c := range cases {
			emitComment(2, "CASES idx=%s\n", strconv.Itoa(i))
			assert(c.dtype == "*astCaseClause", "should be *astCaseClause", __func__)
			var cc = c.caseClause
			labelid++
			var labelCase = ".L.case." + strconv.Itoa(labelid)
			labels[i] = labelCase
			if len(cc.List) == 0 { // @TODO implement slice nil comparison
				defaultLabel = labelCase
				continue
			}
			for _, e := range cc.List {
				assert(getSizeOfType(condType) <= 8 || kind(condType) == T_STRING, "should be one register size or string", __func__)
				emitPushStackTop(condType, "switch expr")
				emitExpr(e, nil)
				emitCompEq(condType)
				emitPopBool(" of switch-case comparison")
				myfmt.Printf("  cmpq $1, %%rax\n")
				myfmt.Printf("  je %s # jump if match\n", labelCase)
			}
		}
		emitComment(2, "End comparison with cases\n")

		// if no case matches, then jump to
		if defaultLabel != "" {
			// default
			myfmt.Printf("  jmp %s\n", defaultLabel)
		} else {
			// exit
			myfmt.Printf("  jmp %s\n", labelEnd)
		}

		emitRevertStackTop(condType)
		for i, c := range cases {
			assert(c.dtype == "*astCaseClause", "should be *astCaseClause", __func__)
			var cc = c.caseClause
			myfmt.Printf("%s:\n", labels[i])
			for _, _s := range cc.Body {
				emitStmt(_s)
			}
			myfmt.Printf("  jmp %s\n", labelEnd)
		}
		myfmt.Printf("%s:\n", labelEnd)
	case "*astTypeSwitchStmt":
		s := stmt.typeSwitchStmt
		typeSwitch := stmt.typeSwitchStmt.node
//		assert(ok, "should exist")
		labelid++
		labelEnd := myfmt.Sprintf(".L.typeswitch.%d.exit", strconv.Itoa(labelid))

		// subjectVariable = subject
		emitVariableAddr(typeSwitch.subjectVariable)
		emitExpr(typeSwitch.subject, nil)
		emitStore(tEface, true, false)

		cases := s.Body.List
		var labels = make([]string, len(cases), len(cases))
		var defaultLabel string
		emitComment(2, "Start comparison with cases\n")
		for i, c := range cases {
			cc := c.caseClause
			//assert(ok, "should be *ast.CaseClause")
			labelid++
			labelCase := ".L.case." + strconv.Itoa(labelid)
			labels[i] = labelCase
			if len(cc.List) == 0 { // @TODO implement slice nil comparison
				defaultLabel = labelCase
				continue
			}
			for _, e := range cc.List {
				emitVariableAddr(typeSwitch.subjectVariable)
				emitLoad(tEface)

				emitDtypeSymbol(e2t(e))
				emitCompExpr("sete") // this pushes 1 or 0 in the end

				emitPopBool(" of switch-case comparison")
				myfmt.Printf("  cmpq $1, %%rax\n")
				myfmt.Printf("  je %s # jump if match\n", labelCase)
			}
		}
		emitComment(2, "End comparison with cases\n")

		// if no case matches, then jump to
		if defaultLabel != "" {
			// default
			myfmt.Printf("  jmp %s\n", defaultLabel)
		} else {
			// exit
			myfmt.Printf("  jmp %s\n", labelEnd)
		}

		for i, typeSwitchCaseClose := range typeSwitch.cases {
			if typeSwitchCaseClose.variable != nil {
				typeSwitch.assignIdent.Obj.Variable = typeSwitchCaseClose.variable
			}
			myfmt.Printf("%s:\n", labels[i])

			for _, _s := range typeSwitchCaseClose.orig.Body {
				if typeSwitchCaseClose.variable != nil {
					// do assignment
					expr := &astExpr{
						dtype: "*astIdent",
						ident: typeSwitch.assignIdent,
					}
					emitAddr(expr)
					emitVariableAddr(typeSwitch.subjectVariable)
					emitLoad(tEface)
					myfmt.Printf("  popq %%rax # ifc.dtype\n")
					myfmt.Printf("  popq %%rcx # ifc.data\n")
					myfmt.Printf("  push %%rcx # ifc.data\n")
					emitLoad(typeSwitchCaseClose.variableType)

					emitStore(typeSwitchCaseClose.variableType, true, false)
				}

				emitStmt(_s)
			}
			myfmt.Printf("  jmp %s\n", labelEnd)
		}
		myfmt.Printf("%s:\n", labelEnd)

	case "*astBranchStmt":
		var containerFor = stmt.branchStmt.currentFor
		var labelToGo string
		switch stmt.branchStmt.Tok {
		case "continue":
			switch dtypeOf(containerFor) {
			case "*astForStmt":
				labelToGo = containerFor.forStmt.labelPost
			case "*astRangeStmt":
				labelToGo = containerFor.rangeStmt.labelPost
			default:
				panic2(__func__, "unexpected container dtype="+containerFor.dtype)
			}
			myfmt.Printf("jmp %s # continue\n", labelToGo)
		case "break":
			switch dtypeOf(containerFor) {
			case "*astForStmt":
				labelToGo = containerFor.forStmt.labelExit
			case "*astRangeStmt":
				labelToGo = containerFor.rangeStmt.labelExit
			default:
				panic2(__func__, "unexpected container dtype="+containerFor.dtype)
			}
			myfmt.Printf("jmp %s # break\n", labelToGo)
		default:
			panic2(__func__, "unexpected tok="+stmt.branchStmt.Tok)
		}
	default:
		panic2(__func__, "TBI:"+stmt.dtype)
	}
}

func blockStmt2Stmt(block *astBlockStmt) *astStmt {
	var stmt = &astStmt{}
	stmt.dtype = "*astBlockStmt"
	stmt.blockStmt = block
	return stmt
}

func emitRevertStackTop(t *Type) {
	myfmt.Printf("  addq $%s, %%rsp # revert stack top\n", strconv.Itoa(getSizeOfType(t)))
}

var labelid int

func getMethodSymbol(method *Method) string {
	var rcvTypeName = method.rcvNamedType
	var subsymbol string
	if method.isPtrMethod {
		subsymbol =  "$" + rcvTypeName.Name + "." + method.name // pointer
	} else {
		subsymbol =  rcvTypeName.Name + "." + method.name // value
	}

	return getPackageSymbol(method.pkgName, subsymbol)
}

func getPackageSymbol(pkgName string, subsymbol string) string {
	return pkgName + "." + subsymbol
}

func emitFuncDecl(pkgName string, fnc *Func) {
	myfmt.Printf("# emitFuncDecl\n")
	var localarea = fnc.localarea
	var symbol string
	if fnc.method != nil {
		symbol = getMethodSymbol(fnc.method)
	} else {
		symbol = getPackageSymbol(pkgName, fnc.name)
	}
	myfmt.Printf("%s: # args %d, locals %d\n",
		symbol, strconv.Itoa(int(fnc.argsarea)), strconv.Itoa(int(fnc.localarea)))

	myfmt.Printf("  pushq %%rbp\n")
	myfmt.Printf("  movq %%rsp, %%rbp\n")
	if localarea != 0 {
		myfmt.Printf("  subq $%d, %%rsp # local area\n", strconv.Itoa(-localarea))
	}

	if fnc.Body != nil {
		emitStmt(blockStmt2Stmt(fnc.Body))
	}

	myfmt.Printf("  leave\n")
	myfmt.Printf("  ret\n")
}

func emitGlobalVariableComplex(name *astIdent, t *Type, val *astExpr) {
	typeKind := kind(t)
	switch typeKind {
	case T_POINTER:
		myfmt.Printf("# init global %s: # T %s\n", name.Name, typeKind)
		lhs := &astExpr{
			ident: name,
			dtype: "*astIdent",
		}
		emitAssign(lhs, val)
	}
}

func emitGlobalVariable(pkg *PkgContainer, name *astIdent, t *Type, val *astExpr) {
	typeKind := kind(t)
	myfmt.Printf("%s.%s: # T %s\n", pkg.name, name.Name, typeKind)
	switch typeKind {
	case T_STRING:
		if val == nil {
			myfmt.Printf("  .quad 0\n")
			myfmt.Printf("  .quad 0\n")
			return
		}
		switch dtypeOf(val) {
		case "*astBasicLit":
			var sl = getStringLiteral(val.basicLit)
			myfmt.Printf("  .quad %s\n", sl.label)
			myfmt.Printf("  .quad %d\n", strconv.Itoa(sl.strlen))
		default:
			panic("Unsupported global string value")
		}
	case T_INTERFACE:
		// only zero value
		myfmt.Printf("  .quad 0 # dtype\n")
		myfmt.Printf("  .quad 0 # data\n")
	case T_BOOL:
		if val == nil {
			myfmt.Printf("  .quad 0 # bool zero value\n")
			return
		}
		switch dtypeOf(val) {
		case "*astIdent":
			switch val.ident.Obj {
			case gTrue:
				myfmt.Printf("  .quad 1 # bool true\n")
			case gFalse:
				myfmt.Printf("  .quad 0 # bool false\n")
			default:
				panic2(__func__, "")
			}
		default:
			panic2(__func__, "")
		}
	case T_INT:
		if val == nil {
			myfmt.Printf("  .quad 0\n")
			return
		}
		switch dtypeOf(val) {
		case "*astBasicLit":
			myfmt.Printf("  .quad %s\n", val.basicLit.Value)
		default:
			panic("Unsupported global value")
		}
	case T_UINT8:
		if val == nil {
			myfmt.Printf("  .byte 0\n")
			return
		}
		switch dtypeOf(val) {
		case "*astBasicLit":
			myfmt.Printf("  .byte %s\n", val.basicLit.Value)
		default:
			panic("Unsupported global value")
		}
	case T_UINT16:
		if val == nil {
			myfmt.Printf("  .word 0\n")
			return
		}
		switch dtypeOf(val) {
		case "*astBasicLit":
			myfmt.Printf("  .word %s\n", val.basicLit.Value)
		default:
			panic("Unsupported global value")
		}
	case T_POINTER:
		// will be set in the initGlobal func
		myfmt.Printf("  .quad 0\n")
	case T_UINTPTR:
		if val != nil {
			panic("Unsupported global value")
		}
		// only zero value
		myfmt.Printf("  .quad 0\n")
	case T_SLICE:
		if val != nil {
			panic("Unsupported global value")
		}
		// only zero value
		myfmt.Printf("  .quad 0 # ptr\n")
		myfmt.Printf("  .quad 0 # len\n")
		myfmt.Printf("  .quad 0 # cap\n")
	case T_ARRAY:
		// only zero value
		if val != nil {
			panic("Unsupported global value")
		}
		if t.e.dtype != "*astArrayType" {
			panic2(__func__, "Unexpected type:"+t.e.dtype)
		}
		var arrayType = t.e.arrayType
		if arrayType.Len == nil {
			panic2(__func__, "global slice is not supported")
		}
		if arrayType.Len.dtype != "*astBasicLit" {
			panic2(__func__, "shoulbe basic literal")
		}
		var length = evalInt(arrayType.Len)
		emitComment(0, "[emitGlobalVariable] array length uint8=%s\n", strconv.Itoa(length))
		var zeroValue string
		var kind string = kind(e2t(arrayType.Elt))
		switch kind {
		case T_INT:
			zeroValue = "  .quad 0 # int zero value\n"
		case T_UINT8:
			zeroValue = "  .byte 0 # uint8 zero value\n"
		case T_STRING:
			zeroValue = "  .quad 0 # string zero value (ptr)\n"
			zeroValue = zeroValue + "  .quad 0 # string zero value (len)\n"
		case T_INTERFACE:
			zeroValue = "  .quad 0 # eface zero value (dtype)\n"
			zeroValue = zeroValue + "  .quad 0 # eface zero value (data)\n"
		default:
			panic2(__func__, "Unexpected kind:"+kind)
		}

		var i int
		for i = 0; i < length; i++ {
			myfmt.Printf(zeroValue)
		}
	default:
		panic2(__func__, "TBI:kind="+typeKind)
	}
}

func generateCode(pkg *PkgContainer) {
	myfmt.Printf("#===================== generateCode %s =====================\n", pkg.name)
	myfmt.Printf(".data\n")
	emitComment(0, "string literals len = %s\n", strconv.Itoa(len(stringLiterals)))
	for _, con := range stringLiterals {
		emitComment(0, "string literals\n")
		myfmt.Printf("%s:\n", con.sl.label)
		myfmt.Printf("  .string %s\n", con.sl.value)
	}

	for _, spec := range pkg.vars {
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariable(pkg, spec.Name, t, spec.Value)
	}

	myfmt.Printf("\n")
	myfmt.Printf(".text\n")
	myfmt.Printf("%s.__initGlobals:\n", pkg.name)
	for _, spec := range pkg.vars {
		if spec.Value == nil{
			continue
		}
		val := spec.Value
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariableComplex(spec.Name, t, val)
	}
	myfmt.Printf("  ret\n")

	for _, fnc := range pkg.funcs {
		emitFuncDecl(pkg.name, fnc)
	}

	myfmt.Printf("\n")
}

// --- type ---
const sliceSize int = 24
const stringSize int = 16
const intSize int = 8
const ptrSize int = 8
const interfaceSize int = 16

type Type struct {
	//kind string
	e *astExpr
}

const T_STRING string = "T_STRING"
const T_INTERFACE string = "T_INTERFACE"
const T_SLICE string = "T_SLICE"
const T_BOOL string = "T_BOOL"
const T_INT string = "T_INT"
const T_INT32 string = "T_INT32"
const T_UINT8 string = "T_UINT8"
const T_UINT16 string = "T_UINT16"
const T_UINTPTR string = "T_UINTPTR"
const T_ARRAY string = "T_ARRAY"
const T_STRUCT string = "T_STRUCT"
const T_POINTER string = "T_POINTER"

func getTypeOfExpr(expr *astExpr) *Type {
	//emitComment(0, "[%s] start\n", __func__)
	switch dtypeOf(expr) {
	case "*astIdent":
		if expr.ident.Obj == nil {
			panic(expr.ident.Name)
		}
		switch expr.ident.Obj.Kind {
		case astVar:
			// injected type is the 1st priority
			// this use case happens in type switch with short decl var
			// switch ident := x.(type) {
			// case T:
			//    y := ident // <= type of ident cannot be associated directly with ident
			//
			if expr.ident.Obj.Variable != nil {
				return expr.ident.Obj.Variable.typ
			}
			switch decl := expr.ident.Obj.Decl.(type) {
			case *astValueSpec:
				var t = &Type{}
				t.e = decl.Type
				return t
			case *astField:
				var t = &Type{}
				t.e = decl.Type
				return t
			case *astAssignStmt: // lhs := rhs
				return getTypeOfExpr(decl.Rhs[0])
			default:
				panic2(__func__, "unkown dtype ")
			}
		case astCon:
			switch expr.ident.Obj {
			case gTrue, gFalse:
				return tBool
			}
			switch decl2 := expr.ident.Obj.Decl.(type) {
			case *astValueSpec:
				return e2t(decl2.Type)
			default:
				panic2(__func__, "cannot decide type of cont ="+expr.ident.Obj.Name)
			}
		default:
			panic2(__func__, "2:Obj=" + expr.ident.Obj.Name + expr.ident.Obj.Kind)
		}
	case "*astBasicLit":
		switch expr.basicLit.Kind {
		case "STRING":
			return tString
		case "INT":
			return tInt
		case "CHAR":
			return tInt32
		default:
			panic2(__func__, "TBI:"+expr.basicLit.Kind)
		}
	case "*astIndexExpr":
		var list = expr.indexExpr.X
		return getElementTypeOfListType(getTypeOfExpr(list))
	case "*astUnaryExpr":
		switch expr.unaryExpr.Op {
		case "+":
			return getTypeOfExpr(expr.unaryExpr.X)
		case "-":
			return getTypeOfExpr(expr.unaryExpr.X)
		case "!":
			return tBool
		case "&":
			var starExpr = &astStarExpr{}
			var t = getTypeOfExpr(expr.unaryExpr.X)
			starExpr.X = t.e
			var eStarExpr = &astExpr{}
			eStarExpr.dtype = "*astStarExpr"
			eStarExpr.starExpr = starExpr
			return e2t(eStarExpr)
		case "range":
			listType := getTypeOfExpr(expr.unaryExpr.X)
			elmType := getElementTypeOfListType(listType)
			return elmType
		default:
			panic2(__func__, "TBI: Op="+expr.unaryExpr.Op)
		}
	case "*astCallExpr":
		emitComment(2, "[%s] *astCallExpr\n", __func__)
		var fun = expr.callExpr.Fun
		switch dtypeOf(fun) {
		case "*astIdent":
			var fn = fun.ident
			if fn.Obj == nil {
				panic2(__func__, "[astCallExpr] nil Obj is not allowed")
			}
			switch fn.Obj.Kind {
			case astTyp:
				return e2t(fun)
			case astFun:
				switch fn.Obj {
				case gLen, gCap:
					return tInt
				case gNew:
					var starExpr = &astStarExpr{}
					starExpr.X = expr.callExpr.Args[0]
					var eStarExpr = &astExpr{}
					eStarExpr.dtype = "*astStarExpr"
					eStarExpr.starExpr = starExpr
					return e2t(eStarExpr)
				case gMake:
					return e2t(expr.callExpr.Args[0])
				case gAppend:
					return e2t(expr.callExpr.Args[0])
				}
				var decl = fn.Obj.Decl
				if decl == nil {
					panic2(__func__, "decl of function "+fn.Name+" is  nil")
				}
				switch dcl := decl.(type) {
				case *astFuncDecl:
					var resultList = dcl.Type.Results.List
					if len(resultList) != 1 {
						panic2(__func__, "[astCallExpr] len results.List is not 1")
					}
					return e2t(dcl.Type.Results.List[0].Type)
				default:
					panic2(__func__, "[astCallExpr] unknown dtype")
				}
				panic2(__func__, "[astCallExpr] Fun ident "+fn.Name)
			}
		case "*astParenExpr": // (X)(e) funcall or conversion
			if isType(fun.parenExpr.X) {
				return e2t(fun.parenExpr.X)
			} else {
				panic("TBI: what should we do ?")
			}
		case "*astArrayType":
			return e2t(fun)
		case "*astSelectorExpr": // (X).Sel()
			assert(fun.selectorExpr.X.dtype == "*astIdent", "want ident, but got " + fun.selectorExpr.X.dtype, __func__)
			xIdent :=  fun.selectorExpr.X.ident
			symbol := xIdent.Name + "." + fun.selectorExpr.Sel.Name
			switch symbol {
			case "unsafe.Pointer":
				// unsafe.Pointer(x)
				return tUintptr
			default:
				fn := fun.selectorExpr
				xIdent := fn.X.ident
				if xIdent.Obj == nil {
					panic2(__func__,  "xIdent.Obj should not be nil")
				}
				if xIdent.Obj.Kind == astPkg {
					funcdecl := lookupForeignFunc(xIdent.Name, fn.Sel.Name)
					return e2t(funcdecl.Type.Results.List[0].Type)
				} else {
					var xType = getTypeOfExpr(fun.selectorExpr.X)
					var method = lookupMethod(xType, fun.selectorExpr.Sel)
					assert(len(method.funcType.Results.List) == 1, "func is expected to return a single value", __func__)
					return e2t(method.funcType.Results.List[0].Type)
				}
			}
		case "*astInterfaceType":
			return tEface
		default:
			panic2(__func__, "[astCallExpr] dtype="+expr.callExpr.Fun.dtype)
		}
	case "*astSliceExpr":
		var underlyingCollectionType = getTypeOfExpr(expr.sliceExpr.X)
		if kind(underlyingCollectionType) == T_STRING {
			// str2 = str1[n:m]
			return tString
		}
		var elementTyp *astExpr
		switch underlyingCollectionType.e.dtype {
		case "*astArrayType":
			elementTyp = underlyingCollectionType.e.arrayType.Elt
		}
		var t = &astArrayType{}
		t.Len = nil
		t.Elt = elementTyp
		var e = &astExpr{}
		e.dtype = "*astArrayType"
		e.arrayType = t
		return e2t(e)
	case "*astStarExpr":
		var t = getTypeOfExpr(expr.starExpr.X)
		var ptrType = t.e.starExpr
		if ptrType == nil {
			panic2(__func__, "starExpr shoud not be nil")
		}
		return e2t(ptrType.X)
	case "*astBinaryExpr":
		switch expr.binaryExpr.Op {
		case "==", "!=", "<", ">", "<=", ">=":
			return tBool
		default:
			return getTypeOfExpr(expr.binaryExpr.X)
		}
	case "*astSelectorExpr":
		if expr.selectorExpr.X.dtype == "*astIdent" && expr.selectorExpr.X.ident.Obj.Kind == astPkg {
			ident := lookupForeignVar(expr.selectorExpr.X.ident.Name, expr.selectorExpr.Sel.Name)
			return getTypeOfExpr(newExpr(ident))
		} else {
			var structType = getStructTypeOfX(expr.selectorExpr)
			var field = lookupStructField(getStructTypeSpec(structType), expr.selectorExpr.Sel.Name)
			return e2t(field.Type)
		}
	case "*astCompositeLit":
		return e2t(expr.compositeLit.Type)
	case "*astParenExpr":
		return getTypeOfExpr(expr.parenExpr.X)
	case "*astTypeAssertExpr":
		return e2t(expr.typeAssertExpr.Type)
	case "*astInterfaceType":
		return tEface
	default:
		panic2(__func__, "TBI:dtype="+expr.dtype)
	}

	panic2(__func__, "nil type is not allowed\n")
	var r *Type
	return r
}

func e2t(typeExpr *astExpr) *Type {
	if typeExpr == nil {
		panic2(__func__, "nil is not allowed")
	}
	var r = &Type{}
	r.e = typeExpr
	return r
}

func serializeType(t *Type) string {
	if t == nil {
		panic("nil type is not expected")
	}
	if t.e == generalSlice {
		panic("TBD: generalSlice")
	}

	switch t.e.dtype {
	case "*astIdent":
		e := t.e.ident
		if t.e.ident.Obj == nil {
			panic("Unresolved identifier:" + t.e.ident.Name)
		}
		if e.Obj.Kind == astVar {
			throw("bug?")
		} else if e.Obj.Kind == astTyp {
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
				var typeSpec *astTypeSpec
				var ok bool
				typeSpec, ok = decl.(*astTypeSpec)
				if !ok {
					panic("unexpected dtype")
				}
				return "main." + typeSpec.Name.Name
			}
		}
	case "*astStructType":
		return "struct"
	case "*astArrayType":
		e := t.e.arrayType
		if e.Len == nil {
			if e.Elt == nil {
				panic(e)
			}
			return "[]" + serializeType(e2t(e.Elt))
		} else {
			return "[" + strconv.Itoa(evalInt(e.Len)) + "]" + serializeType(e2t(e.Elt))
		}
	case "*astStarExpr":
		e := t.e.starExpr
		return "*" + serializeType(e2t(e.X))
	case "*astEllipsis": // x ...T
		panic("TBD: Ellipsis")
	case "*astInterfaceType":
		return "interface"
	default:
		throw(t.e.dtype)
	}
	return ""
}


func kind(t *Type) string {
	if t == nil {
		panic2(__func__, "nil type is not expected\n")
	}
	if t.e == generalSlice {
		return T_SLICE
	}

	var e = t.e
	switch t.e.dtype {
	case "*astIdent":
		var ident = t.e.ident
		switch ident.Name {
		case "uintptr":
			return T_UINTPTR
		case "int":
			return T_INT
		case "int32":
			return T_INT32
		case "string":
			return T_STRING
		case "uint8", "byte":
			return T_UINT8
		case "uint16":
			return T_UINT16
		case "bool":
			return T_BOOL
		default:
			// named type
			var decl = ident.Obj.Decl
			var typeSpec *astTypeSpec
			var ok bool
			typeSpec , ok = decl.(*astTypeSpec)
			if !ok {
				panic2(__func__, "unsupported decl :")
			}
			return kind(e2t(typeSpec.Type))
		}
	case "*astStructType":
		return T_STRUCT
	case "*astArrayType":
		if e.arrayType.Len == nil {
			return T_SLICE
		} else {
			return T_ARRAY
		}
	case "*astStarExpr":
		return T_POINTER
	case "*astEllipsis": // x ...T
		return T_SLICE // @TODO is this right ?
	case "*astInterfaceType":
		return T_INTERFACE
	case "*astParenExpr":
		panic(t.e.parenExpr)
	case "*astSelectorExpr":
		full := t.e.selectorExpr.X.ident.Name + "." + t.e.selectorExpr.Sel.Name
		if full == "unsafe.Pointer" {
			return T_POINTER
		} else {
			panic2(__func__, "Unkown type")
		}
	default:
		panic2(__func__, "Unkown dtype:"+t.e.dtype)
	}
	panic2(__func__, "error")
	return ""
}

func isInterface(t *Type) bool {
	return kind(t) == T_INTERFACE
}

func getStructTypeOfX(e *astSelectorExpr) *Type {
	var typeOfX = getTypeOfExpr(e.X)
	var structType *Type
	switch kind(typeOfX) {
	case T_STRUCT:
		// strct.field => e.X . e.Sel
		structType = typeOfX
	case T_POINTER:
		// ptr.field => e.X . e.Sel
		assert(typeOfX.e.dtype == "*astStarExpr", "should be astStarExpr", __func__)
		var ptrType = typeOfX.e.starExpr
		structType = e2t(ptrType.X)
	default:
		panic2(__func__, "TBI")
	}
	return structType
}

func getElementTypeOfListType(t *Type) *Type {
	switch kind(t) {
	case T_SLICE, T_ARRAY:
		switch t.e.dtype {
	case "*astArrayType":
		return e2t(t.e.arrayType.Elt)
	case "*astEllipsis":
		return e2t(t.e.ellipsis.Elt)
	default:
		throw(t.e.dtype)
	}
	case T_STRING:
		return tUint8
	default:
		panic2(__func__, "TBI kind="+kind(t))
	}
	var r *Type
	return r
}

func getSizeOfType(t *Type) int {
	var knd = kind(t)
	switch kind(t) {
	case T_SLICE:
		return sliceSize
	case T_STRING:
		return 16
	case T_ARRAY:
		var elemSize = getSizeOfType(e2t(t.e.arrayType.Elt))
		return elemSize * evalInt(t.e.arrayType.Len)
	case T_INT, T_UINTPTR, T_POINTER:
		return 8
	case T_UINT8:
		return 1
	case T_UINT16:
		return 2
	case T_BOOL:
		return 8
	case T_STRUCT:
		return calcStructSizeAndSetFieldOffset(getStructTypeSpec(t))
	case T_INTERFACE:
		return interfaceSize
	default:
		panic2(__func__, "TBI:"+knd)
	}
	return 0
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
		throw(kind(t))
	}
	throw(kind(t))
	return 0
}

func getStructFieldOffset(field *astField) int {
	var offset = field.Offset
	return offset
}

func setStructFieldOffset(field *astField, offset int) {
	field.Offset = offset
}

func getStructFields(structTypeSpec *astTypeSpec) []*astField {
	if structTypeSpec.Type.dtype != "*astStructType" {
		panic2(__func__, "Unexpected dtype")
	}
	var structType = structTypeSpec.Type.structType
	return structType.Fields.List
}

func getStructTypeSpec(namedStructType *Type) *astTypeSpec {
	if kind(namedStructType) != T_STRUCT {
		panic2(__func__, "not T_STRUCT")
	}
	if namedStructType.e.dtype != "*astIdent" {
		panic2(__func__, "astIdent expected got " + namedStructType.e.dtype)
	}

	var ident = namedStructType.e.ident
	var typeSpec *astTypeSpec
	var ok bool
	typeSpec , ok = ident.Obj.Decl.(*astTypeSpec)
	if !ok {
		panic2(__func__, "not *astTypeSpec")
	}

	return typeSpec
}

func lookupStructField(structTypeSpec *astTypeSpec, selName string) *astField {
	var field *astField
	for _, field := range getStructFields(structTypeSpec) {
		if field.Name.Name == selName {
			return field
		}
	}
	panic("Unexpected flow: struct field not found:" + selName)
	return field
}

func calcStructSizeAndSetFieldOffset(structTypeSpec *astTypeSpec) int {
	var offset int = 0

	var fields = getStructFields(structTypeSpec)
	for _, field := range fields {
		setStructFieldOffset(field, offset)
		var size = getSizeOfType(e2t(field.Type))
		offset = offset + size
	}
	return offset
}

// --- walk ---
type sliteral struct {
	label  string
	strlen int
	value  string // raw value/pre/precompiler.go:2150
}

type stringLiteralsContainer struct {
	lit *astBasicLit
	sl  *sliteral
}

type Func struct {
	localvars   []*string
	localarea   int
	argsarea    int
	funcType    *astFuncType
	rcvType     *astExpr
	name        string
	Body        *astBlockStmt
	method *Method
}

type Method struct {
	pkgName string
	rcvNamedType *astIdent
	isPtrMethod  bool
	name string
	funcType *astFuncType
}

type Variable struct {
	name         string
	isGlobal     bool
	globalSymbol string
	localOffset  int
	typ          *Type
}

//type localoffsetint int //@TODO

func (fnc *Func) registerParamVariable(name string, t *Type) *Variable {
	vr := newLocalVariable(name, fnc.argsarea, t)
	fnc.argsarea = fnc.argsarea + getSizeOfType(t)
	return vr
}

func (fnc *Func) registerLocalVariable(name string, t *Type) *Variable {
	assert(t != nil && t.e != nil, "type of local var should not be nil", __func__)
	fnc.localarea = fnc.localarea - getSizeOfType(t)
	return newLocalVariable(name, currentFunc.localarea, t)
}

var stringLiterals []*stringLiteralsContainer
var stringIndex int

var currentFunc *Func

func getStringLiteral(lit *astBasicLit) *sliteral {
	for _, container := range stringLiterals {
		if container.lit == lit {
			return container.sl
		}
	}

	panic2(__func__, "string literal not found:"+lit.Value)
	var r *sliteral
	return r
}

func registerStringLiteral(lit *astBasicLit) {
	logf(" [registerStringLiteral] begin\n")

	if pkg.name == "" {
		panic2(__func__, "no pkgName")
	}

	var strlen int
	var vl = []uint8(lit.Value)
	for _, c := range vl {
		if c != '\\' {
			strlen++
		}
	}

	label := myfmt.Sprintf(".%s.S%d", pkg.name, strconv.Itoa(stringIndex))
	stringIndex++

	sl := &sliteral{}
	sl.label = label
	sl.strlen = strlen - 2
	sl.value = lit.Value
	logf(" [registerStringLiteral] label=%s, strlen=%s %s\n", sl.label, strconv.Itoa(sl.strlen), sl.value)
	cont := &stringLiteralsContainer{}
	cont.sl = sl
	cont.lit = lit
	stringLiterals = append(stringLiterals, cont)
}

func newGlobalVariable(pkgName string, name string, t *Type) *Variable {
	vr := &Variable{
		name:         name,
		isGlobal:     true,
		globalSymbol: pkgName + "." + name,
		typ: t,
	}
	return vr
}

func newLocalVariable(name string, localoffset int, t *Type) *Variable {
	vr := &Variable{
		name : name,
		isGlobal : false,
		localOffset : localoffset,
		typ: t,
	}
	return vr
}


type methodEntry struct {
	name string
	method *Method
}

type namedTypeEntry struct {
	name    string
	methods []*methodEntry
}

var typesWithMethods []*namedTypeEntry

func findNamedType(typeName string) *namedTypeEntry {
	var typ *namedTypeEntry
	for _, typ := range typesWithMethods {
		if typ.name == typeName {
			return typ
		}
	}
	typ = nil
	return typ
}

func newMethod(pkgName string, funcDecl *astFuncDecl) *Method {
	var rcvType = funcDecl.Recv.List[0].Type
	var isPtr bool
	if rcvType.dtype == "*astStarExpr" {
		isPtr = true
		rcvType = rcvType.starExpr.X
	}
	assert(rcvType.dtype == "*astIdent", "receiver type should be ident", __func__)
	var rcvNamedType = rcvType.ident

	var method = &Method{
		pkgName: pkgName,
		rcvNamedType: rcvNamedType,
		isPtrMethod : isPtr,
		name: funcDecl.Name.Name,
		funcType: funcDecl.Type,
	}
	return method
}

func registerMethod(method *Method) {
	var nt = findNamedType(method.rcvNamedType.Name)
	if nt == nil {
		nt = &namedTypeEntry{
			name:    method.rcvNamedType.Name,
			methods: nil,
		}
		typesWithMethods = append(typesWithMethods, nt)
	}

	var me *methodEntry = &methodEntry{
		name: method.name,
		method: method,
	}
	nt.methods = append(nt.methods, me)
}

func lookupMethod(rcvT *Type, methodName *astIdent) *Method {
	var rcvType = rcvT.e
	if rcvType.dtype == "*astStarExpr" {
		rcvType = rcvType.starExpr.X
	}
	assert(rcvType.dtype == "*astIdent", "receiver type should be ident", __func__)
	var rcvTypeName = rcvType.ident
	var nt = findNamedType(rcvTypeName.Name)
	if nt == nil {
		panic(rcvTypeName.Name + " has no moethodeiverTypeName:")
	}

	for _, me := range nt.methods {
		if me.name == methodName.Name {
			return me.method
		}
	}

	panic("method not found: " + methodName.Name)
	var r *Method
	return r
}

func walkStmt(stmt *astStmt) {
	logf(" [%s] begin dtype=%s\n", __func__, stmt.dtype)
	switch dtypeOf(stmt) {
	case "*astDeclStmt":
		logf(" [%s] *ast.DeclStmt\n", __func__)
		if stmt.DeclStmt == nil {
			panic2(__func__, "nil pointer exception\n")
		}
		var declStmt = stmt.DeclStmt
		if declStmt.Decl == nil {
			panic2(__func__, "ERROR\n")
		}
		var dcl = declStmt.Decl
		var genDecl *astGenDecl
		var ok bool
		genDecl, ok = dcl.(*astGenDecl)
		if !ok {
			panic2(__func__, "[dcl.dtype] internal error")
		}
		var valSpec *astValueSpec
		valSpec, ok = genDecl.Spec.(*astValueSpec)
		if valSpec.Type == nil {
			if valSpec.Value == nil {
				panic2(__func__, "type inference requires a value")
			}
			var _typ = getTypeOfExpr(valSpec.Value)
			if _typ != nil && _typ.e != nil {
				valSpec.Type = _typ.e
			} else {
				panic2(__func__, "type inference failed")
			}
		}
		var typ = valSpec.Type // Type can be nil
		logf(" [walkStmt] valSpec Name=%s, Type=%s\n",
			valSpec.Name.Name, typ.dtype)

		t := e2t(typ)
		valSpec.Name.Obj.Variable = currentFunc.registerLocalVariable(valSpec.Name.Name, t)
		logf(" var %s offset = %d\n", valSpec.Name.Obj.Name,
			strconv.Itoa(valSpec.Name.Obj.Variable.localOffset))
		if valSpec.Value != nil {
			walkExpr(valSpec.Value)
		}
	case "*astAssignStmt":
		var lhs = stmt.assignStmt.Lhs[0]
		var rhs = stmt.assignStmt.Rhs[0]
		if stmt.assignStmt.Tok == ":=" {
			assert(lhs.dtype == "*astIdent", "should be ident", __func__)
			var obj = lhs.ident.Obj
			assert(obj.Kind == astVar, "should be ast.Var", __func__)
			walkExpr(rhs)
			// infer type
			var typ = getTypeOfExpr(rhs)
			if typ != nil && typ.e != nil {
			} else {
				panic("type inference is not supported: " + obj.Name)
			}
			logf("infered type of %s is %s, rhs=%s\n", obj.Name, typ.e.dtype, rhs.dtype)
			obj.Variable = currentFunc.registerLocalVariable(obj.Name, typ)
		} else {
			walkExpr(rhs)
		}
	case "*astExprStmt":
		walkExpr(stmt.exprStmt.X)
	case "*astReturnStmt":
		stmt.returnStmt.node = &nodeReturnStmt{
			fnc: currentFunc,
		}
		for _, rt := range stmt.returnStmt.Results {
			walkExpr(rt)
		}
	case "*astIfStmt":
		if stmt.ifStmt.Init != nil {
			walkStmt(stmt.ifStmt.Init)
		}
		walkExpr(stmt.ifStmt.Cond)
		for _, s := range stmt.ifStmt.Body.List {
			walkStmt(s)
		}
		if stmt.ifStmt.Else != nil {
			walkStmt(stmt.ifStmt.Else)
		}
	case "*astForStmt":
		stmt.forStmt.Outer = currentFor
		currentFor = stmt
		if stmt.forStmt.Init != nil {
			walkStmt(stmt.forStmt.Init)
		}
		if stmt.forStmt.Cond != nil {
			walkExpr(stmt.forStmt.Cond)
		}
		if stmt.forStmt.Post != nil {
			walkStmt(stmt.forStmt.Post)
		}
		var _s = &astStmt{}
		_s.dtype = "*astBlockStmt"
		_s.blockStmt = stmt.forStmt.Body
		walkStmt(_s)
		currentFor = stmt.forStmt.Outer
	case "*astRangeStmt":
		walkExpr(stmt.rangeStmt.X)
		stmt.rangeStmt.Outer = currentFor
		currentFor = stmt
		var _s = blockStmt2Stmt(stmt.rangeStmt.Body)
		walkStmt(_s)
		var lenvar = currentFunc.registerLocalVariable(".range.len", tInt)
		var indexvar = currentFunc.registerLocalVariable(".range.index", tInt)

		if stmt.rangeStmt.Tok == ":=" {
			listType := getTypeOfExpr(stmt.rangeStmt.X)

			keyIdent := stmt.rangeStmt.Key.ident
			//@TODO map key can be any type
			//keyType := getKeyTypeOfListType(listType)
			var keyType *Type = tInt
			keyIdent.Obj.Variable =  currentFunc.registerLocalVariable(keyIdent.Name, keyType)

			// determine type of Value
			elmType := getElementTypeOfListType(listType)
			valueIdent := stmt.rangeStmt.Value.ident
			valueIdent.Obj.Variable = currentFunc.registerLocalVariable(valueIdent.Name, elmType)
		}
		stmt.rangeStmt.lenvar = lenvar
		stmt.rangeStmt.indexvar = indexvar
		currentFor = stmt.rangeStmt.Outer
	case "*astIncDecStmt":
		walkExpr(stmt.incDecStmt.X)
	case "*astBlockStmt":
		for _, s := range stmt.blockStmt.List {
			walkStmt(s)
		}
	case "*astBranchStmt":
		stmt.branchStmt.currentFor = currentFor
	case "*astSwitchStmt":
		if stmt.switchStmt.Tag != nil {
			walkExpr(stmt.switchStmt.Tag)
		}
		walkStmt(blockStmt2Stmt(stmt.switchStmt.Body))
	case "*astTypeSwitchStmt":
		s := stmt.typeSwitchStmt
		typeSwitch := &nodeTypeSwitchStmt{}
		stmt.typeSwitchStmt.node = typeSwitch
		var assignIdent *astIdent
		switch s.Assign.dtype {
		case "*astExprStmt":
			typeAssertExpr := s.Assign.exprStmt.X.typeAssertExpr
			//assert(ok, "should be *ast.TypeAssertExpr")
			typeSwitch.subject = typeAssertExpr.X
			walkExpr(typeAssertExpr.X)
		case "*astAssignStmt":
			lhs := s.Assign.assignStmt.Lhs[0]
			//var ok bool
			assignIdent = lhs.ident
			//assert(ok, "lhs should be ident")
			typeSwitch.assignIdent = assignIdent
			// ident will be a new local variable in each case clause
			typeAssertExpr := s.Assign.assignStmt.Rhs[0].typeAssertExpr
			//assert(ok, "should be *ast.TypeAssertExpr")
			typeSwitch.subject = typeAssertExpr.X
			walkExpr(typeAssertExpr.X)
		default:
			throw(s.Assign.dtype)
		}

		typeSwitch.subjectVariable = currentFunc.registerLocalVariable(".switch_expr", tEface)
		for _, _case := range s.Body.List {
			cc := _case.caseClause
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
				assignIdent.Obj.Variable = vr
			}

			for _, s_ := range cc.Body {
				walkStmt(s_)
			}
			if assignIdent != nil {
				assignIdent.Obj.Variable = nil
			}
		}
	case "*astCaseClause":
		for _, e_ := range stmt.caseClause.List {
			walkExpr(e_)
		}
		for _, s_ := range stmt.caseClause.Body {
			walkStmt(s_)
		}
	default:
		panic2(__func__, "TBI: stmt.dtype="+stmt.dtype)
	}
}

var currentFor *astStmt

func walkExpr(expr *astExpr) {
	logf(" [walkExpr] dtype=%s\n", expr.dtype)
	switch dtypeOf(expr) {
	case "*astIdent":
		// what to do ?
	case "*astCallExpr":
		walkExpr(expr.callExpr.Fun)
		// Replace __func__ ident by a string literal
		var basicLit *astBasicLit
		var newArg *astExpr
		for i, arg := range expr.callExpr.Args {
			if arg.dtype == "*astIdent" {
				var ident = arg.ident
				if ident.Name == "__func__" && ident.Obj.Kind == astVar {
					basicLit = &astBasicLit{}
					basicLit.Kind = "STRING"
					basicLit.Value = "\"" + currentFunc.name + "\""
					newArg = &astExpr{}
					newArg.dtype = "*astBasicLit"
					newArg.basicLit = basicLit
					expr.callExpr.Args[i] = newArg
					arg = newArg
				}
			}
			walkExpr(arg)
		}
	case "*astBasicLit":
		switch expr.basicLit.Kind {
		case "STRING":
			registerStringLiteral(expr.basicLit)
		}
	case "*astCompositeLit":
		for _, v := range expr.compositeLit.Elts {
			walkExpr(v)
		}
	case "*astUnaryExpr":
		walkExpr(expr.unaryExpr.X)
	case "*astBinaryExpr":
		walkExpr(expr.binaryExpr.X)
		walkExpr(expr.binaryExpr.Y)
	case "*astIndexExpr":
		walkExpr(expr.indexExpr.Index)
		walkExpr(expr.indexExpr.X)
	case "*astSliceExpr":
		if expr.sliceExpr.Low != nil {
			walkExpr(expr.sliceExpr.Low)
		}
		if expr.sliceExpr.High != nil {
			walkExpr(expr.sliceExpr.High)
		}
		if expr.sliceExpr.Max != nil {
			walkExpr(expr.sliceExpr.Max)
		}
		walkExpr(expr.sliceExpr.X)
	case "*astStarExpr":
		walkExpr(expr.starExpr.X)
	case "*astSelectorExpr":
		walkExpr(expr.selectorExpr.X)
	case "*astArrayType": // []T(e)
		// do nothing ?
	case "*astParenExpr":
		walkExpr(expr.parenExpr.X)
	case "*astKeyValueExpr":
		walkExpr(expr.keyValueExpr.Key)
		walkExpr(expr.keyValueExpr.Value)
	case "*astInterfaceType":
		// interface{}(e)  conversion. Nothing to do.
	case "*astTypeAssertExpr":
		walkExpr(expr.typeAssertExpr.X)
	default:
		panic2(__func__, "TBI:"+expr.dtype)
	}
}

var ExportedQualifiedIdents []*exportEntry

type exportEntry struct {
	qi string
	any interface{} // *astFuncDecl|*astIdent(variable)
}

func walk(pkg *PkgContainer, file *astFile) {
	var typeSpecs []*astTypeSpec
	var funcDecls []*astFuncDecl
	var varSpecs []*astValueSpec
	var constSpecs []*astValueSpec

	for _, decl := range file.Decls {
		switch dcl := decl.(type) {
		case *astGenDecl:
			switch spec := dcl.Spec.(type) {
			case *astTypeSpec:
				typeSpecs = append(typeSpecs, spec)
			case *astValueSpec:
				if spec.Name.Obj.Kind == astVar {
					varSpecs = append(varSpecs, spec)
				} else if spec.Name.Obj.Kind == astCon {
					constSpecs = append(constSpecs, spec)
				} else {
					panic("Unexpected")
				}
			}
		case *astFuncDecl:
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
		exportEntry := &exportEntry{
			qi: pkg.name + "." + funcDecl.Name.Name,
			any: funcDecl,
		}
		ExportedQualifiedIdents = append(ExportedQualifiedIdents, exportEntry)
		if funcDecl.Body != nil {
			if funcDecl.Recv != nil { // is Method
				var method = newMethod(pkg.name, funcDecl)
				registerMethod(method)
			}
		}
	}

	for _, constSpec := range constSpecs {
		walkExpr(constSpec.Value)
	}

	for _, valSpec := range varSpecs {
		var nameIdent = valSpec.Name
		assert(nameIdent.Obj.Kind == astVar, "should be Var", __func__)
		if valSpec.Type == nil {
			var val = valSpec.Value
			var t = getTypeOfExpr(val)
			valSpec.Type = t.e
		}
		nameIdent.Obj.Variable = newGlobalVariable(pkg.name, nameIdent.Obj.Name, e2t(valSpec.Type))
		pkg.vars = append(pkg.vars, valSpec)
		exportEntry := &exportEntry{
			qi: pkg.name + "." + nameIdent.Name,
			any: nameIdent,
		}
		ExportedQualifiedIdents = append(ExportedQualifiedIdents, exportEntry)
		if valSpec.Value != nil {
			walkExpr(valSpec.Value)
		}
	}

		for _, funcDecl := range funcDecls {
			fnc := &Func{
				name:      funcDecl.Name.Name,
				funcType:  funcDecl.Type,
				localarea: 0,
				argsarea:  16,
			}
			currentFunc = fnc
			logf(" [sema] == astFuncDecl %s ==\n", funcDecl.Name.Name)
			//var paramoffset = 16
			var paramFields []*astField

			if funcDecl.Recv != nil { // Method
				paramFields = append(paramFields, funcDecl.Recv.List[0])
			}
			for _, field := range funcDecl.Type.Params.List {
				paramFields = append(paramFields, field)
			}

			for _, field := range paramFields {
				obj := field.Name.Obj
				obj.Variable = fnc.registerParamVariable(obj.Name, e2t(field.Type))
			}

			if funcDecl.Body != nil {
				for _, stmt := range funcDecl.Body.List {
					walkStmt(stmt)
				}
				fnc.Body = funcDecl.Body

				if funcDecl.Recv != nil { // Method
					fnc.method = newMethod(pkg.name, funcDecl)
				}
				pkg.funcs = append(pkg.funcs, fnc)
			}
		}
}

// --- universe ---
var gNil = &astObject{
	Kind : astCon, // is it Con ?
	Name : "nil",
}

var eNil = &astExpr{
	dtype : "*astIdent",
	ident : &astIdent{
		Obj:  gNil,
		Name: "nil",
	},
}

var eZeroInt = &astExpr{
	dtype: "*astBasicLit",
	basicLit: &astBasicLit{
		Value: "0",
		Kind: "INT",
	},
}


var gTrue = &astObject{
	Kind: astCon,
	Name: "true",
}
var gFalse = &astObject{
	Kind: astCon,
	Name: "false",
}

var gString = &astObject{
	Kind: astTyp,
	Name: "string",
}

var gInt = &astObject{
	Kind: astTyp,
	Name: "int",
}

var gInt32 = &astObject{
	Kind: astTyp,
	Name: "int32",
}


var gUint8 = &astObject{
	Kind: astTyp,
	Name: "uint8",
}

var gUint16 = &astObject{
	Kind: astTyp,
	Name: "uint16",
}
var gUintptr = &astObject{
	Kind: astTyp,
	Name: "uintptr",
}
var gBool = &astObject{
	Kind: astTyp,
	Name: "bool",
}

var gNew = &astObject{
	Kind: astFun,
	Name: "new",
}

var gMake = &astObject{
	Kind: astFun,
	Name: "make",
}
var gAppend = &astObject{
	Kind: astFun,
	Name: "append",
}

var gLen = &astObject{
	Kind: astFun,
	Name: "len",
}

var gCap = &astObject{
	Kind: astFun,
	Name: "cap",
}
var gPanic = &astObject{
	Kind: astFun,
	Name: "panic",
}

var tInt = &Type{
	e: &astExpr{
		dtype: "*astIdent",
		ident : &astIdent{
			Name: "int",
			Obj: gInt,
		},
	},
}

// Rune
var tInt32 = &Type{
	e: &astExpr{
		dtype: "*astIdent",
		ident : &astIdent{
			Name: "int32",
			Obj: gInt32,
		},
	},
}

var tUint8 = &Type{
	e: &astExpr{
		dtype: "*astIdent",
		ident: &astIdent{
			Name: "uint8",
			Obj:  gUint8,
		},
	},
}

var tUint16 = &Type{
	e: &astExpr{
		dtype: "*astIdent",
		ident: &astIdent{
			Name: "uint16",
			Obj:  gUint16,
		},
	},
}
var tUintptr = &Type{
	e: &astExpr{
		dtype: "*astIdent",
		ident: &astIdent{
			Name: "uintptr",
			Obj:  gUintptr,
		},
	},
}

var tString = &Type{
	e : &astExpr{
		dtype : "*astIdent",
		ident : &astIdent{
			Name : "string",
			Obj : gString,
		},
	},
}

var tEface *Type = &Type{
	e: &astExpr{
		dtype: "*astInterfaceType",
		interfaceType: &astInterfaceType{},
	},
}


var tBool = &Type{
	e : &astExpr{
		dtype : "*astIdent",
		ident : &astIdent{
			Name : "bool",
			Obj : gBool,
		},
	},
}

var tSliceOfString = &Type{
	e: &astExpr{
		dtype: "*astArrayType",
		arrayType: &astArrayType{
			Len: nil,
			Elt: &astExpr{
				dtype: "*astIdent",
				ident: &astIdent{
					Name: "string",
					Obj: gString,
				},
			},
		},
	},
}

var generalSlice = &astExpr{
	dtype: "*astIdent",
	ident: &astIdent{},
}

func createUniverse() *astScope {
	var universe = new(astScope)

	scopeInsert(universe, gInt)
	scopeInsert(universe, gUint8)

	universe.Objects = append(universe.Objects, &objectEntry{
		name: "byte",
		obj:  gUint8,
	})

	scopeInsert(universe, gUint16)
	scopeInsert(universe, gUintptr)
	scopeInsert(universe, gString)
	scopeInsert(universe, gBool)
	scopeInsert(universe, gNil)
	scopeInsert(universe, gTrue)
	scopeInsert(universe, gFalse)
	scopeInsert(universe, gNew)
	scopeInsert(universe, gMake)
	scopeInsert(universe, gAppend)
	scopeInsert(universe, gLen)
	scopeInsert(universe, gCap)
	scopeInsert(universe, gPanic)

	logf(" [%s] scope insertion of predefined identifiers complete\n", __func__)

	// @FIXME package names should not be be in universe

	scopeInsert(universe, &astObject{
		Kind: "Pkg",
		Name: "os",
	})

	scopeInsert(universe, &astObject{
		Kind: "Pkg",
		Name: "syscall",
	})

	scopeInsert(universe, &astObject{
		Kind: "Pkg",
		Name: "unsafe",
	})
	logf(" [%s] scope insertion complete\n", __func__)
	return universe
}

func resolveUniverse(file *astFile, universe *astScope) {
	logf("[%s] start\n", __func__)

	var mapImports []string
	for _, imprt := range file.Imports {
		// unwrap double quote "..."
		rawPath := imprt.Path[1:(len(imprt.Path) - 1)]
		base := path.Base(rawPath)
		mapImports = append(mapImports, base)
	}

	// inject predeclared identifers
	var unresolved []*astIdent
	logf(" [SEMA] resolving file.Unresolved (n=%s)\n", strconv.Itoa(len(file.Unresolved)))
	for _, ident := range file.Unresolved {
		logf(" [SEMA] resolving ident %s ... \n", ident.Name)
		var obj *astObject = scopeLookup(universe, ident.Name)
		if obj != nil {
			logf(" matched\n")
			ident.Obj = obj
		} else {
			if inArray(ident.Name, mapImports) {
				ident.Obj = &astObject{
					Kind: astPkg,
					Name: ident.Name,
				}
				logf("# resolved: %s\n", ident.Name)
			} else {
				// we should allow unresolved for now.
				// e.g foo in X{foo:bar,}
				logf("Unresolved (maybe struct field name in composite literal): "+ident.Name)
				unresolved = append(unresolved, ident)
			}
		}
	}
}

func lookupForeignVar(pkg string, identifier string) *astIdent {
	key := pkg + "." + identifier
	logf("lookupForeignVar... %s\n", key)
	for _, entry := range ExportedQualifiedIdents {
		logf("  looking into %s\n", entry.qi)
		if entry.qi == key {
			var ident *astIdent
			var ok bool
			ident, ok = entry.any.(*astIdent)
			if !ok  {
				panic("not ident")
			}
			return ident
		}
	}
	return nil
}

func lookupForeignFunc(pkg string, identifier string) *astFuncDecl {
	key := pkg + "." + identifier
	logf("lookupForeignFunc... %s\n", key)
	for _, entry := range ExportedQualifiedIdents {
		logf("  looking into %s\n", entry.qi)
		if entry.qi == key {
			var fdecl *astFuncDecl
			var ok bool
			fdecl, ok = entry.any.(*astFuncDecl)
			if !ok  {
				panic("not fdecl")
			}
			return fdecl
		}
	}
	panic("function not found: " + key)
	return nil
}


var pkg *PkgContainer

type PkgContainer struct {
	name string
	vars []*astValueSpec
	funcs []*Func
}

func showHelp() {
	myfmt.Printf("Usage:\n")
	myfmt.Printf("    babygo version:  show version\n")
	myfmt.Printf("    babygo [-DF] [-DG] filename\n")
}

const GOPATH string = "/root/go"

// "foo/bar" => "bar.go"
// "some/dir" => []string{"a.go", "b.go"}
func findFilesInDir(dir string) []string {
	//fname := path2.Base(dir) + ".go"
	//return []string{fname}
	dirents := mylib.GetDirents(dir)
	var r []string
	for _, dirent := range dirents {
		if dirent == "." || dirent == ".." {
			continue
		}
		r = append(r, dirent)
	}
	return r
}

func isStdLib(pth string) bool {
	for _, b := range []uint8(pth) {
		if b == '/' {
			return false
		}
	}
	return true
}

func getImportPathsFromFile(file string) []string {
	astFile0 := parseImports(file)
	var importPaths []string
	for _, importSpec := range astFile0.Imports {
		rawValue := importSpec.Path
		logf("import %s\n", rawValue)
		pth :=  rawValue[1:len(rawValue)-1]
		importPaths = append(importPaths, pth)
	}
	return 	importPaths
}

func isInTree(tree []*depEntry, pth string) bool {
	for _, entry := range tree {
		if entry.path == pth {
			return true
		}
	}
	return false
}

func getPackageDir(importPath string) string {
	if isStdLib(importPath) {
		return srcPath + "/github.com/DQNEO/babygo/src/" + importPath
	} else {
		return srcPath + "/" + importPath
	}
}

func collectDependency(tree []*depEntry, paths []string) []*depEntry {
	logf(" collectDependency\n")
	for _, pkgPath := range paths {
		if isInTree(tree, pkgPath) {
			continue
		}
		logf("   in pkgPath=%s\n", pkgPath)
		packageDir := getPackageDir(pkgPath)
		fnames := findFilesInDir(packageDir)
		var children []string
		for _, fname := range fnames {
			_paths := getImportPathsFromFile(packageDir + "/" + fname)
			for _, p := range _paths {
				children = append(children, p)
			}
		}

		newEntry := &depEntry{
			path:     pkgPath,
			children: children,
		}
		tree = append(tree, newEntry)
		tree = collectDependency(tree, children)
	}
	return tree
}

func removeLeafNode(tree []*depEntry, sortedPaths []string) []*depEntry {
	// remove leaf node
	var newTree []*depEntry
	for _, entry := range tree {
		if inArray(entry.path, sortedPaths) {
			continue
		}
		de := &depEntry{
			path:     entry.path,
			children: nil,
		}
		for _, child := range entry.children {
			if inArray(child, sortedPaths) {
				continue
			}
			de.children = append(de.children, child)
		}
		newTree = append(newTree, de)
	}
	return newTree
}

func collectLeafNode(sortedPaths []string, tree []*depEntry) []string {
	for _, entry := range tree {
		if len(entry.children) == 0 {
			// leaf node
			logf("Found leaf node: %s\n", entry.path)
			logf("  num children: %d\n", len(entry.children))
			sortedPaths = append(sortedPaths, entry.path)
		}
	}
	return sortedPaths
}

func sortDepTree(tree []*depEntry) []string {
	var sortedPaths []string

	var keys []string
	for _, entry := range tree {
		keys = append(keys, entry.path)
	}
	mylib.SortStrings(keys)
	var newTree []*depEntry
	for _, key := range keys {
		for _, entry := range tree {
			if entry.path == key {
				newTree = append(newTree, entry)
			}
		}
	}
	tree = newTree
	logf("====TREE====\n")
	for {
		if len(tree) == 0 {
			break
		}
		sortedPaths = collectLeafNode(sortedPaths, tree)
		tree = removeLeafNode(tree, sortedPaths)
	}
	return sortedPaths
}

var srcPath string

func main() {
	if len(os.Args) == 1 {
		showHelp()
		return
	}

	if os.Args[1] == "version" {
		myfmt.Printf("babygo version 0.1.0  linux/amd64\n")
		return
	} else if os.Args[1] == "help" {
		showHelp()
		return
	} else if os.Args[1] == "panic" {
		panicVersion := strconv.Itoa(mylib.Sum(1 , 1))
		panic("I am panic version " + panicVersion)
	}

	logf("Build start\n")
	srcPath = os.Getenv("GOPATH") + "/src"

	var universe = createUniverse()
	var arg string

	for _, arg = range os.Args {
		switch arg {
		case "-DF":
			debugFrontEnd = true
		case "-DG":
			debugCodeGen = true
		}
	}

	var mainFile = arg // last arg
	logf("input file: \"%s\"\n", mainFile)
	logf("Parsing imports\n")

	importPaths := getImportPathsFromFile(mainFile)
	var tree []*depEntry
	tree = collectDependency(tree, importPaths)
	logf("====TREE====\n")
	for _, _pkg := range tree {
		logf("pkg: %s\n", _pkg.path)
		for _, child := range _pkg.children {
			logf("  %s\n", child)
		}
	}


	var stdPackagesUsed []string
	var extPackagesUsed []string

	sortedPaths := sortDepTree(tree)
	for _, pth := range sortedPaths {
		if isStdLib(pth) {
			stdPackagesUsed = append(stdPackagesUsed, pth)
		} else {
			extPackagesUsed = append(extPackagesUsed, pth)
		}
	}

	myfmt.Printf("# === sorted stdPackagesUsed ===\n")
	for _, pth := range stdPackagesUsed {
		myfmt.Printf("#  %s\n", pth)
	}

	myfmt.Printf("# === sorted extPackagesUsed ===\n")
	for _, pth := range extPackagesUsed {
		myfmt.Printf("#  %s\n", pth)
	}

	var packagesToBuild = []string{"runtime.go"}
	for _, p := range stdPackagesUsed {
		logf("std package: %s\n", p)
		pkgDir := srcPath + "/github.com/DQNEO/babygo/src/" + p
		logf("srcPath=%s\n", srcPath)
		fnames := findFilesInDir(pkgDir)
		srcFile := pkgDir + "/" + fnames[0]
		logf("# internal file: %s\n", srcFile)
		packagesToBuild = append(packagesToBuild, srcFile)
	}
	for _, p := range extPackagesUsed {
		logf("ext package: %s\n", p)
		pkgDir := srcPath + "/" + p
		logf("srcPath=%s\n", srcPath)
		fnames := findFilesInDir(pkgDir)
		extfile := pkgDir + "/" + fnames[0]
		logf("# external file: %s\n", extfile)
		packagesToBuild = append(packagesToBuild, extfile)
	}

	packagesToBuild = append(packagesToBuild, mainFile)

	for _, sourceFile := range packagesToBuild {
		logf("# package %s ============================================\n", sourceFile)
		stringIndex = 0
		stringLiterals = nil
		astFile := parseFile(sourceFile, false)
		resolveUniverse(astFile, universe)
		pkg = &PkgContainer{
			name: astFile.Name,
		}
		walk(pkg, astFile)
		generateCode(pkg)
	}

	// emitting dynamic types
	myfmt.Printf("# ------- Dynamic Types ------\n")
	myfmt.Printf(".data\n")
	for _, te := range typeMap {
		id := te.id
		name := te.serialized
		symbol := typeIdToSymbol(id)
		myfmt.Printf("%s: # %s\n", symbol, name)
		myfmt.Printf("  .quad %s\n", strconv.Itoa(id))
		myfmt.Printf("  .quad .S.dtype.%s\n", strconv.Itoa(id))
		myfmt.Printf("  .quad %d\n", strconv.Itoa(len(name)))
		myfmt.Printf(".S.dtype.%s:\n", strconv.Itoa(id))
		myfmt.Printf("  .string \"%s\"\n", name)
	}
	myfmt.Printf("\n")

}

type depEntry struct {
	path string
	children []string
}

func newExpr(x interface{}) *astExpr {
	r := &astExpr{
		ifc: x,
	}

	switch xx := x.(type) {
	case *astIdent:
		r.dtype = "*astIdent"
		r.ident = xx
	case *astArrayType:
		r.dtype = "*astArrayType"
		r.arrayType = xx
	case *astBasicLit:
		r.dtype = "*astBasicLit"
		r.basicLit = xx
	case *astCallExpr:
		r.dtype = "*astCallExpr"
		r.callExpr = xx
	case *astBinaryExpr:
		r.dtype = "*astBinaryExpr"
		r.binaryExpr = xx
	case *astUnaryExpr:
		r.dtype = "*astUnaryExpr"
		r.unaryExpr = xx
	case *astSelectorExpr:
		r.dtype = "*astSelectorExpr"
		r.selectorExpr = xx
	case *astIndexExpr:
		r.dtype = "*astIndexExpr"
		r.indexExpr = xx
	case *astSliceExpr:
		r.dtype = "*astSliceExpr"
		r.sliceExpr = xx
	case *astStarExpr:
		r.dtype = "*astStarExpr"
		r.starExpr = xx
	case *astParenExpr:
		r.dtype = "*astParenExpr"
		r.parenExpr = xx
	case *astStructType:
		r.dtype = "*astStructType"
		r.structType = xx
	case *astCompositeLit:
		r.dtype = "*astCompositeLit"
		r.compositeLit = xx
	case *astKeyValueExpr:
		r.dtype = "*astKeyValueExpr"
		r.keyValueExpr = xx
	case *astEllipsis:
		r.dtype = "*astEllipsis"
		r.ellipsis = xx
	case *astInterfaceType:
		r.dtype = "*astInterfaceType"
		r.interfaceType = xx
	case *astTypeAssertExpr:
		r.dtype = "*astTypeAssertExpr"
		r.typeAssertExpr = xx
	}
	return r
}

