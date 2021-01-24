package main

import "syscall"
import "os"

// --- foundation ---
func assert(bol bool, msg string, caller string) {
	if !bol {
		panic2(caller, msg)
	}
}

func throw(s string) {
	panic(s)
}

var __func__ = "__func__"

func panic2(caller string, x string) {
	panic("[" + caller + "] " + x)
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
				var arg = a[argIndex]
				argIndex++
				var s = arg // // p.printArg(arg, c)
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

func inArray(x string, list []string) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}

var debugFrontEnd bool

func logf(format string, a ...string) {
	if !debugFrontEnd {
		return
	}
	var f = "# " + format
	var s = fmtSprintf(f, a)
	syscall.Write(1, []uint8(s))
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
	logf("src len = %s\n", Itoa(len(s.src)))
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
			panic2(__func__, "unknown char:"+string([]uint8{ch})+":"+Itoa(int(ch)))
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

type signature struct {
	params  *astFieldList
	results *astFieldList
}

type ObjDecl struct {
	dtype     string
	valueSpec *astValueSpec
	funcDecl  *astFuncDecl
	typeSpec  *astTypeSpec
	field     *astField
	assignment *astAssignStmt
}

type astObject struct {
	Kind     string
	Name     string
	Decl     *ObjDecl
	Variable *Variable
}

type astExpr struct {
	dtype         string
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
	typeAssert    *astTypeAssertExpr
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
	Fun  *astExpr   // function expression
	Args []*astExpr // function arguments; or nil
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
	Type *astExpr
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
	dtype      string
	DeclStmt   *astDeclStmt
	exprStmt   *astExprStmt
	blockStmt  *astBlockStmt
	assignStmt *astAssignStmt
	returnStmt *astReturnStmt
	ifStmt     *astIfStmt
	forStmt    *astForStmt
	incDecStmt *astIncDecStmt
	isRange    bool
	rangeStmt  *astRangeStmt
	branchStmt *astBranchStmt
	switchStmt *astSwitchStmt
	caseClause *astCaseClause
}

type astDeclStmt struct {
	Decl *astDecl
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

// Declarations
type astSpec struct {
	dtype     string
	valueSpec *astValueSpec
	typeSpec  *astTypeSpec
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
type astDecl struct {
	dtype    string
	genDecl  *astGenDecl
	funcDecl *astFuncDecl
}

type astGenDecl struct {
	Spec *astSpec
}

type astFuncDecl struct {
	Recv *astFieldList
	Name *astIdent
	Type *astFuncType
	Body *astBlockStmt
}

type astFile struct {
	Name       string
	Decls      []*astDecl
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
	// @TODO check error
	fd, _ = syscall.Open(filename, O_READONLY, 0)
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
		logf(" [parser] pointing at : \"%s\" newline (%s)\n", p.tok.tok, Itoa(p.scanner.offset))
	} else if p.tok.tok == "IDENT" {
		logf(" [parser] pointing at: IDENT \"%s\" (%s)\n", p.tok.lit, Itoa(p.scanner.offset))
	} else {
		logf(" [parser] pointing at: \"%s\" %s (%s)\n", p.tok.tok, p.tok.lit, Itoa(p.scanner.offset))
	}

	if p.tok.tok == "COMMENT" {
		for p.tok.tok == "COMMENT" {
			p.consumeComment()
		}
	}
}

func (p *parser) expect(tok string, who string) {
	if p.tok.tok != tok {
		var s = fmtSprintf("%s expected, but got %s", []string{tok, p.tok.tok})
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

func (p *parser) parseImportDecl() *astImportSpec {
	p.expect("import", __func__)
	var path = p.tok.lit
	p.next()
	p.expectSemi(__func__)

	return &astImportSpec{
		Path: path,
	}
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

		return &astExpr{
			dtype: "*astEllipsis",
			ellipsis: &astEllipsis{
				Elt: typ,
			},
		}
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
	return &astExpr{
		dtype: "*astStarExpr",
		starExpr : &astStarExpr{
			X: base,
		},
	}
}

func (p *parser) parseArrayType() *astExpr {
	p.expect("[", __func__)
	var ln *astExpr
	if p.tok.tok != "]" {
		ln = p.parseRhs()
	}
	p.expect("]", __func__)
	var elt = p.parseType()

	var r = &astExpr{
		dtype : "*astArrayType",
		arrayType : &astArrayType{
			Elt : elt,
			Len : ln,
		},
	}
	return r
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

	return &astExpr{
		dtype : "*astStructType",
		structType : &astStructType{
			Fields: &astFieldList{
				List : list,
			},
		},
	}
}

func (p *parser) parseTypeName() *astExpr {
	logf(" [%s] begin\n", __func__)
	var ident = p.parseIdent()
	logf(" [%s] end\n", __func__)
	return &astExpr{
		ident: ident,
		dtype: "*astIdent",
	}
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
		return &astExpr{
			dtype : "*astInterfaceType",
			interfaceType: &astInterfaceType{
				methods: nil,
			},
		}
	case "(":
		p.next()
		var _typ = p.parseType()
		p.expect(")", __func__)
		return &astExpr{
			dtype: "*astParenExpr",
			parenExpr: &astParenExpr{
				X: _typ,
			},
		}
	}
	var _nil *astExpr
	return _nil
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
	logf(" [%s] collected list n=%s\n", __func__, Itoa(len(list)))

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
		logf(" [DEBUG] range i = %s\n", Itoa(i))
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
	list = append(list, &astField{
		Type: typ,
	})
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
		Decl : &ObjDecl{
			dtype: "*astField",
			field: decl,
		},
		Name : ident.Name,
		Kind : kind,
	}

	ident.Obj = obj

	// scope insert
	if ident.Name != "_" {
		scopeInsert(scope, obj)
	}
}

func declare(objDecl *ObjDecl, scope *astScope, kind string, ident *astIdent) {
	logf(" [declare] ident %s\n", ident.Name)

	//valSpec.Name.Obj
	var obj = &astObject{
		Decl : objDecl,
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
		var eIdent = &astExpr{
			dtype : "*astIdent",
			ident : ident,
		}
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
		return &astExpr{
			dtype:    "*astBasicLit",
			basicLit: basicLit,
		}
	case "(":
		p.next() // consume "("
		parserExprLev++
		var x = p.parseRhsOrType()
		parserExprLev--
		p.expect(")", __func__)
		return &astExpr{
			dtype: "*astParenExpr",
			parenExpr: &astParenExpr{
				X: x,
			},
		}
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
	for p.tok.tok != ")" {
		var arg = p.parseExpr()
		list = append(list, arg)
		if p.tok.tok == "," {
			p.next()
		} else if p.tok.tok == ")" {
			break
		}
	}
	p.expect(")", __func__)
	return &astExpr{
		dtype:    "*astCallExpr",
		callExpr: &astCallExpr{
			Fun:  fn,
			Args: list,
		},
	}
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
					var fn = &astExpr{
						dtype : "*astSelectorExpr",
						selectorExpr : sel,
					}
					// string = x.ident.Name + "." + secondIdent
					x = p.parseCallExpr(fn)
					logf(" [parsePrimaryExpr] 741 p.tok.tok=%s\n", p.tok.tok)
				} else {
					logf("   end parsePrimaryExpr()\n")
					x = &astExpr{
						dtype : "*astSelectorExpr",
						selectorExpr : sel,
					}
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
	return &astExpr{
		dtype: "*astTypeAssertExpr",
		typeAssert: &astTypeAssertExpr{
			X:    x,
			Type: typ,
		},
	}
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
		x = &astExpr{
			dtype : "*astKeyValueExpr",
			keyValueExpr : kvExpr,
		}
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
	return &astExpr{
		dtype:        "*astCompositeLit",
		compositeLit: &astCompositeLit{
			Type: typ,
			Elts: elts,
		},
	}
}

func isLiteralType(x *astExpr) bool {
	switch x.dtype {
	case "*astIdent":
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
		if ncolons == 2 {
			panic2(__func__, "TBI: ncolons=2")
		}
		var sliceExpr = &astSliceExpr{
			Slice3 : false,
			X : x,
			Low : index[0],
			High : index[1],
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
	logf("   begin parseBinaryExpr() prec1=%s\n", Itoa(prec1))
	var x = p.parseUnaryExpr()
	var oprec int
	for {
		var op = p.tok.tok
		oprec = precedence(op)
		logf(" oprec %s\n", Itoa(oprec))
		logf(" precedence \"%s\" %s < %s\n", op, Itoa(oprec), Itoa(prec1))
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
		logf(" [DEBUG] range as len lhs=%s\n", Itoa(len(as.Lhs)))
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

	var switchStmt = &astSwitchStmt{}
	switchStmt.Body = body
	switchStmt.Tag = makeExpr(s2)
	var s = &astStmt{}
	s.dtype = "*astSwitchStmt"
	s.switchStmt = switchStmt
	p.closeScope()
	return s
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
				var objDecl = &ObjDecl{
					dtype: "*astAssignStmt",
					assignment: as,
				}
				assert(lhs.dtype == "*astIdent", "should be ident", __func__)
				declare(objDecl, p.topScope, astVar, lhs.ident)
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
		var decl = &astDecl{}
		decl.dtype = "*astGenDecl"
		decl.genDecl = genDecl
		s.DeclStmt.Decl = decl
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
		var spec = &astSpec{}
		spec.dtype = "*astValueSpec"
		spec.valueSpec = valSpec
		var objDecl = &ObjDecl{}
		objDecl.dtype = "*astValueSpec"
		objDecl.valueSpec = valSpec
		declare(objDecl, p.topScope, astVar, ident)
		r = &astGenDecl{}
		r.Spec = spec
		return r
	default:
		panic2(__func__, "TBI\n")
	}
	return r
}

func (p *parser) parserTypeSpec() *astSpec {
	logf(" [%s] start\n", __func__)
	p.expect("type", __func__)
	var ident = p.parseIdent()
	logf(" decl type %s\n", ident.Name)

	var spec = &astTypeSpec{}
	spec.Name = ident
	var objDecl = &ObjDecl{}
	objDecl.dtype = "*astTypeSpec"
	objDecl.typeSpec = spec
	declare(objDecl, p.topScope, astTyp, ident)
	var typ = p.parseType()
	p.expectSemi(__func__)
	spec.Type = typ
	var r = &astSpec{}
	r.dtype = "*astTypeSpec"
	r.typeSpec = spec
	return r
}

func (p *parser) parseValueSpec(keyword string) *astSpec {
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
	var r = &astSpec{}
	r.dtype = "*astValueSpec"
	r.valueSpec = spec
	var objDecl = &ObjDecl{}
	objDecl.dtype = "*astValueSpec"
	objDecl.valueSpec = spec
	var kind = astCon
	if keyword == "var" {
		kind = astVar
	}
	declare(objDecl, p.topScope, kind, ident)
	logf(" [parserValueSpec] end\n")
	return r
}

func (p *parser) parseFuncDecl() *astDecl {
	p.expect("func", __func__)
	var scope = astNewScope(p.topScope) // function scope
	var receivers *astFieldList
	if p.tok.tok == "(" {
		logf("  [parserFuncDecl] parsing method")
		receivers = p.parseParameters(scope, false)
	}
	var ident = p.parseIdent() // func name
	var sig = p.parseSignature(scope)
	var params = sig.params
	var results = sig.results
	if results == nil {
		logf(" [parserFuncDecl] %s sig.results is nil\n", ident.Name)
	} else {
		logf(" [parserFuncDecl] %s sig.results.List = %s\n", ident.Name, Itoa(len(sig.results.List)))
	}
	var body *astBlockStmt
	if p.tok.tok == "{" {
		logf(" begin parseBody()\n")
		body = p.parseBody(scope)
		logf(" end parseBody()\n")
		p.expectSemi(__func__)
	} else {
		p.expectSemi(__func__)
	}
	var decl = &astDecl{}
	decl.dtype = "*astFuncDecl"
	var funcDecl = &astFuncDecl{}
	decl.funcDecl = funcDecl
	decl.funcDecl.Recv = receivers
	decl.funcDecl.Name = ident
	decl.funcDecl.Type = &astFuncType{}
	decl.funcDecl.Type.Params = params
	decl.funcDecl.Type.Results = results
	decl.funcDecl.Body = body
	if receivers == nil {
		var objDecl = &ObjDecl{}
		objDecl.dtype = "*astFuncDecl"
		objDecl.funcDecl = funcDecl
		declare(objDecl, p.pkgScope, astFun, ident)
	}
	return decl
}

func (p *parser) parseFile() *astFile {
	// expect "package" keyword
	p.expect("package", __func__)
	p.unresolved = nil
	var ident = p.parseIdent()
	var packageName = ident.Name
	p.expectSemi(__func__)

	p.topScope = &astScope{} // open scope
	p.pkgScope = p.topScope

	for p.tok.tok == "import" {
		p.parseImportDecl()
	}

	logf("\n")
	logf(" [parser] Parsing Top level decls\n")
	var decls []*astDecl
	var decl *astDecl

	for p.tok.tok != "EOF" {
		switch p.tok.tok {
		case "var", "const":
			var spec = p.parseValueSpec(p.tok.tok)
			var genDecl = &astGenDecl{}
			genDecl.Spec = spec
			decl = &astDecl{}
			decl.dtype = "*astGenDecl"
			decl.genDecl = genDecl
		case "func":
			logf("\n\n")
			decl = p.parseFuncDecl()
			logf(" func decl parsed:%s\n", decl.funcDecl.Name.Name)
		case "type":
			var spec = p.parserTypeSpec()
			var genDecl = &astGenDecl{}
			genDecl.Spec = spec
			decl = &astDecl{}
			decl.dtype = "*astGenDecl"
			decl.genDecl = genDecl
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
	logf(" [parserFile] resolving parser's unresolved (n=%s)\n", Itoa(len(p.unresolved)))
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
	logf(" [parserFile] Unresolved (n=%s)\n", Itoa(len(unresolved)))

	var f = &astFile{}
	f.Name = packageName
	f.Decls = decls
	f.Unresolved = unresolved
	logf(" [%s] end\n", __func__)
	return f
}

func parseFile(filename string) *astFile {
	var text = readSource(filename)

	var p = &parser{}
	p.scanner = &scanner{}
	p.init(text)
	return p.parseFile()
}

// --- codegen ---
var debugCodeGen bool

func emitComment(indent int, format string, a ...string) {
	if !debugCodeGen {
		return
	}
	var spaces []uint8
	var i int
	for i = 0; i < indent; i++ {
		spaces = append(spaces, ' ')
	}
	var format2 = string(spaces) + "# " + format
	var s = fmtSprintf(format2, a)
	syscall.Write(1, []uint8(s))
}

func evalInt(expr *astExpr) int {
	switch expr.dtype {
	case "*astBasicLit":
		return Atoi(expr.basicLit.Value)
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

func emitLoad(t *Type) {
	if t == nil {
		panic2(__func__, "nil type error\n")
	}
	emitPopAddress(kind(t))
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
		panic2(__func__, "TBI:kind="+kind(t))
	}
}

func emitVariableAddr(variable *Variable) {
	emitComment(2, "emit Addr of variable \"%s\" \n", variable.name)

	if variable.isGlobal {
		fmtPrintf("  leaq %s(%%rip), %%rax # global variable addr \"%s\"\n", variable.globalSymbol,  variable.name)
	} else {
		fmtPrintf("  leaq %d(%%rbp), %%rax # local variable addr \"%s\"\n", Itoa(variable.localOffset),  variable.name)
	}

	fmtPrintf("  pushq %%rax # variable address\n")
}

func emitListHeadAddr(list *astExpr) {
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
		panic2(__func__, "kind="+kind(getTypeOfExpr(list)))
	}
}

func isOsArgs(e *astSelectorExpr) bool {
	return e.X.dtype == "*astIdent" && e.X.ident.Name == "os" && e.Sel.Name == "Args"
}

func emitAddr(expr *astExpr) {
	emitComment(2, "[emitAddr] %s\n", expr.dtype)
	switch expr.dtype {
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
		if isOsArgs(expr.selectorExpr) {
			fmtPrintf("  leaq %s(%%rip), %%rax # hack for os.Args\n", "runtime.__args__")
			fmtPrintf("  pushq %%rax\n")
			return
		}
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
	switch expr.dtype {
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

func emitConversion(tp *Type, arg0 *astExpr) {
	emitComment(2, "[emitConversion]\n")
	var typeExpr = tp.e
	switch typeExpr.dtype {
	case "*astIdent":
		switch typeExpr.ident.Obj {
		case gString: // string(e)
			switch kind(getTypeOfExpr(arg0)) {
			case T_SLICE: // string(slice)
				emitExpr(arg0, nil) // slice
				emitPopSlice()
				fmtPrintf("  pushq %%rcx # str len\n")
				fmtPrintf("  pushq %%rax # str ptr\n")
			}
		case gInt, gUint8, gUint16, gUintptr: // int(e)
			emitComment(2, "[emitConversion] to int \n")
			emitExpr(arg0, nil)
		default:
			if typeExpr.ident.Obj.Kind == astTyp {
				if typeExpr.ident.Obj.Decl.dtype != "*astTypeSpec" {
					panic2(__func__, "Something is wrong")
				}
				//e2t(typeExpr.ident.Obj.Decl.typeSpec.Type))
				emitExpr(arg0, nil)
			} else{
				panic2(__func__, "[*astIdent] TBI : "+typeExpr.ident.Obj.Name)
			}
		}
	case "*astArrayType": // Conversion to slice
		var arrayType = typeExpr.arrayType
		if arrayType.Len != nil {
			panic2(__func__, "internal error")
		}
		if (kind(getTypeOfExpr(arg0))) != T_STRING {
			panic2(__func__, "source type should be string")
		}
		emitComment(2, "Conversion of string => slice \n")
		emitExpr(arg0, nil)
		emitPopString()
		fmtPrintf("  pushq %%rcx # cap\n")
		fmtPrintf("  pushq %%rcx # len\n")
		fmtPrintf("  pushq %%rax # ptr\n")
	case "*astParenExpr":
		emitConversion(e2t(typeExpr.parenExpr.X), arg0)
	case "*astStarExpr": // (*T)(e)
		emitComment(2, "[emitConversion] to pointer \n")
		emitExpr(arg0, nil)
	default:
		panic2(__func__, "TBI :"+typeExpr.dtype)
	}
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
		fmtPrintf("  pushq $0 # %s zero value\n", kind(t))
	case T_STRUCT:
		var structSize = getSizeOfType(t)
		fmtPrintf("# zero value of a struct. size=%s (allocating on heap)\n", Itoa(structSize))
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
		fmtPrintf("  pushq %%rcx # len\n")
	case T_STRING:
		emitExpr(arg, nil)
		emitPopString()
		fmtPrintf("  pushq %%rcx # len\n")
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
		fmtPrintf("  pushq %%rdx # cap\n")
	case T_STRING:
		panic("cap() cannot accept string type")
	default:
		throw(kind(getTypeOfExpr(arg)))
	}
}

func emitCallMalloc(size int) {
	fmtPrintf("  pushq $%s\n", Itoa(size))
	// call malloc and return pointer
	var resultList = []*astField{
		&astField{
			Type:    tUintptr.e,
		},
	}
	fmtPrintf("  callq runtime.malloc\n") // no need to invert args orders
	emitRevertStackPointer(intSize)
	emitReturnedValue(resultList)
}

func emitStructLiteral(e *astCompositeLit) {
	// allocate heap area with zero value
	fmtPrintf("  # Struct literal\n")
	var structType = e2t(e.Type)
	emitZeroValue(structType) // push address of the new storage
	var kvExpr *astKeyValueExpr
	for i, elm := range e.Elts {
		assert(elm.dtype == "*astKeyValueExpr", "wrong dtype 1:" + elm.dtype, __func__)
		kvExpr = elm.keyValueExpr
		assert(kvExpr.Key.dtype == "*astIdent", "wrong dtype 2:" + elm.dtype, __func__)
		var fieldName = kvExpr.Key.ident
		fmtPrintf("  #  - [%s] : key=%s, value=%s\n", Itoa(i), fieldName.Name, kvExpr.Value.dtype)
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
		emitAddConst(elmSize*i, "malloced address + elmSize * index ("+Itoa(i)+")")
		ctx := &evalContext{
			_type: elmType,
		}
		emitExprIfc(elm, ctx)
		emitStore(elmType, true, false)
	}
}

func emitInvertBoolValue() {
	emitPopBool("")
	fmtPrintf("  xor $1, %%rax\n")
	fmtPrintf("  pushq %%rax\n")
}

func emitTrue() {
	fmtPrintf("  pushq $1 # true\n")
}

func emitFalse() {
	fmtPrintf("  pushq $0 # false\n")
}

type Arg struct {
	e      *astExpr
	t      *Type // expected type
	offset int
}

func emitArgs(args []*Arg) int {
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

func prepareArgs(funcType *astFuncType, receiver *astExpr, eArgs []*astExpr) []*Arg {
	if funcType == nil {
		panic("no funcType")
	}
	var params = funcType.Params.List
	var variadicArgs []*astExpr
	var variadicElp *astEllipsis
	var args []*Arg
	var param *astField
	var arg *Arg
	var lenParams = len(params)
	for argIndex, eArg := range eArgs {
		emitComment(2, "[%s][*astIdent][default] loop idx %s, len params %s\n", __func__, Itoa(argIndex), Itoa(lenParams))
		if argIndex < lenParams {
			param = params[argIndex]
			if param.Type.dtype == "*astEllipsis" {
				variadicElp = param.Type.ellipsis
				variadicArgs = make([]*astExpr, 0, 20)
			}
		}
		if variadicElp != nil {
			variadicArgs = append(variadicArgs, eArg)
			continue
		}

		var paramType = e2t(param.Type)
		arg = &Arg{}
		arg.e = eArg
		arg.t = paramType
		args = append(args, arg)
	}

	if variadicElp != nil {
		// collect args as a slice
		var sliceType = &astArrayType{}
		sliceType.Elt = variadicElp.Elt
		var eSliceType = &astExpr{}
		eSliceType.dtype = "*astArrayType"
		eSliceType.arrayType = sliceType
		var sliceLiteral = &astCompositeLit{}
		sliceLiteral.Type = eSliceType
		sliceLiteral.Elts = variadicArgs
		var eSliceLiteral = &astExpr{}
		eSliceLiteral.compositeLit = sliceLiteral
		eSliceLiteral.dtype = "*astCompositeLit"
		var _arg = &Arg{}
		_arg.e = eSliceLiteral
		_arg.t = e2t(eSliceType)
		args = append(args, _arg)
	} else if len(args) < len(params) {
		// Add nil as a variadic arg
		emitComment(2, "len(args)=%s, len(params)=%s\n", Itoa(len(args)), Itoa(len(params)))
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
	fmtPrintf("  callq %s\n", symbol)
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
			fmtPrintf("  pushq %%rdi # str len\n")
			fmtPrintf("  pushq %%rax # str ptr\n")
		case T_INTERFACE:
			fmtPrintf("  pushq %%rdi # ifc data\n")
			fmtPrintf("  pushq %%rax # ifc dtype\n")
		case T_BOOL, T_INT, T_UINTPTR, T_POINTER:
			fmtPrintf("  pushq %%rax\n")
		case T_SLICE:
			fmtPrintf("  pushq %%rsi # slice cap\n")
			fmtPrintf("  pushq %%rdi # slice len\n")
			fmtPrintf("  pushq %%rax # slice ptr\n")
		default:
			panic2(__func__, "Unexpected kind="+knd)
		}
	default:
		panic2(__func__,"multipul returned values is not supported ")
	}
}

func emitFuncall(fun *astExpr, eArgs []*astExpr) {
	var symbol string
	var receiver *astExpr
	var funcType *astFuncType
	switch fun.dtype {
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
			fmtPrintf("  callq runtime.printstring\n")
			fmtPrintf("  addq $%s, %%rsp # revert \n", Itoa(16))
			return
		}

		if fn.Name == "makeSlice1" || fn.Name == "makeSlice8" || fn.Name == "makeSlice16" || fn.Name == "makeSlice24" {
			fn.Name = "makeSlice"
		}
		// general function call
		symbol = getFuncSymbol(pkg.name, fn.Name)
		emitComment(2, "[%s][*astIdent][default] start\n", __func__)

		var obj = fn.Obj
		var decl = obj.Decl
		if decl == nil {
			panic2(__func__, "[*astCallExpr] decl is nil")
		}
		if decl.dtype != "*astFuncDecl" {
			panic2(__func__, "[*astCallExpr] decl.dtype is invalid")
		}
		var fndecl = decl.funcDecl
		if fndecl == nil {
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
		case "os.Exit":
			funcType = funcTypeOsExit
		case "syscall.Open":
			funcType = funcTypeSyscallOpen
		case "syscall.Read":
			funcType = funcTypeSyscallRead
		case "syscall.Write":
			funcType = funcTypeSyscallWrite
		case "syscall.Syscall":
			funcType = funcTypeSyscallSyscall
		default:
			// Assume method call
			receiver = selectorExpr.X
			var receiverType = getTypeOfExpr(receiver)
			var method = lookupMethod(receiverType, selectorExpr.Sel)
			funcType = method.funcType
			var subsymbol = getMethodSymbol(method)
			symbol = getFuncSymbol(pkg.name, subsymbol)
		}
	case "*astParenExpr":
		panic2(__func__, "[astParenExpr] TBI ")
	default:
		panic2(__func__, "TBI fun.dtype="+fun.dtype)
	}

	var args = prepareArgs(funcType, receiver, eArgs)
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
	var valSpec = ident.Obj.Decl.valueSpec
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
	switch e.dtype {
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
		emitAddr(e)
		emitLoad(getTypeOfExpr(e))
	case "*astCallExpr":
		var fun = e.callExpr.Fun
		emitComment(2, "[%s][*astCallExpr]\n", __func__)
		if isType(fun) {
			emitConversion(e2t(fun), e.callExpr.Args[0])
		} else {
			emitFuncall(fun, e.callExpr.Args)
		}
	case "*astParenExpr":
		emitExpr(e.parenExpr.X, ctx)
	case "*astBasicLit":
		//		emitComment(0, "basicLit.Kind = %s \n", e.basicLit.Kind)
		switch e.basicLit.Kind {
		case "INT":
			var ival = Atoi(e.basicLit.Value)
			fmtPrintf("  pushq $%d # number literal\n", Itoa(ival))
		case "STRING":
			var sl = getStringLiteral(e.basicLit)
			if sl.strlen == 0 {
				// zero value
				emitZeroValue(tString)
			} else {
				fmtPrintf("  pushq $%d # str len\n", Itoa(sl.strlen))
				fmtPrintf("  leaq %s, %%rax # str ptr\n", sl.label)
				fmtPrintf("  pushq %%rax # str ptr\n")
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
			fmtPrintf("  pushq $%d # convert char literal to int\n", Itoa(int(char)))
		default:
			panic2(__func__, "[*astBasicLit] TBI : "+e.basicLit.Kind)
		}
	case "*astSliceExpr":
		var list = e.sliceExpr.X
		var listType = getTypeOfExpr(list)
		emitExpr(e.sliceExpr.High, nil)
		emitExpr(e.sliceExpr.Low, nil)
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
			panic2(__func__, "Unknown kind="+kind(listType))
		}

		emitExpr(e.sliceExpr.Low, nil)
		var elmType = getElementTypeOfListType(listType)
		emitListElementAddr(list, elmType)
	case "*astUnaryExpr":
		emitComment(2, "[DEBUG] unary op = %s\n", e.unaryExpr.Op)
		switch e.unaryExpr.Op {
		case "+":
			emitExpr(e.unaryExpr.X, nil)
		case "-":
			emitExpr(e.unaryExpr.X, nil)
			fmtPrintf("  popq %%rax # e.X\n")
			fmtPrintf("  imulq $-1, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
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
				var labelExitWithFalse = fmtSprintf(".L.%s.false", []string{Itoa(labelid)})
				var labelExit = fmtSprintf(".L.%d.exit", []string{Itoa(labelid)})
				emitExpr(e.binaryExpr.X, nil) // left
				emitPopBool("left")
				fmtPrintf("  cmpq $1, %%rax\n")
				// exit with false if left is false
				fmtPrintf("  jne %s\n", labelExitWithFalse)

				// if left is true, then eval right and exit
				emitExpr(e.binaryExpr.Y, nil) // right
				fmtPrintf("  jmp %s\n", labelExit)

				fmtPrintf("  %s:\n", labelExitWithFalse)
				emitFalse()
				fmtPrintf("  %s:\n", labelExit)
			case "||":
				labelid++
				var labelExitWithTrue = fmtSprintf(".L.%d.true", []string{Itoa(labelid)})
				var labelExit = fmtSprintf(".L.%d.exit", []string{Itoa(labelid)})
				emitExpr(e.binaryExpr.X, nil) // left
				emitPopBool("left")
				fmtPrintf("  cmpq $1, %%rax\n")
				// exit with true if left is true
				fmtPrintf("  je %s\n", labelExitWithTrue)

				// if left is false, then eval right and exit
				emitExpr(e.binaryExpr.Y, nil) // right
				fmtPrintf("  jmp %s\n", labelExit)

				fmtPrintf("  %s:\n", labelExitWithTrue)
				emitTrue()
				fmtPrintf("  %s:\n", labelExit)
			case "+":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				fmtPrintf("  popq %%rcx # right\n")
				fmtPrintf("  popq %%rax # left\n")
				fmtPrintf("  addq %%rcx, %%rax\n")
				fmtPrintf("  pushq %%rax\n")
			case "-":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				fmtPrintf("  popq %%rcx # right\n")
				fmtPrintf("  popq %%rax # left\n")
				fmtPrintf("  subq %%rcx, %%rax\n")
				fmtPrintf("  pushq %%rax\n")
			case "*":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				fmtPrintf("  popq %%rcx # right\n")
				fmtPrintf("  popq %%rax # left\n")
				fmtPrintf("  imulq %%rcx, %%rax\n")
				fmtPrintf("  pushq %%rax\n")
			case "%":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				fmtPrintf("  popq %%rcx # right\n")
				fmtPrintf("  popq %%rax # left\n")
				fmtPrintf("  movq $0, %%rdx # init %%rdx\n")
				fmtPrintf("  divq %%rcx\n")
				fmtPrintf("  movq %%rdx, %%rax\n")
				fmtPrintf("  pushq %%rax\n")
			case "/":
				emitExpr(e.binaryExpr.X, nil) // left
				emitExpr(e.binaryExpr.Y, nil)   // right
				fmtPrintf("  popq %%rcx # right\n")
				fmtPrintf("  popq %%rax # left\n")
				fmtPrintf("  movq $0, %%rdx # init %%rdx\n")
				fmtPrintf("  divq %%rcx\n")
				fmtPrintf("  pushq %%rax\n")
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
			fmtPrintf("  pushq $%d # slice.cap\n", Itoa(length))
			fmtPrintf("  pushq $%d # slice.len\n", Itoa(length))
			fmtPrintf("  pushq %%rax # slice.ptr\n")
		default:
			panic2(__func__, "Unexpected kind="+k)
		}
	case "*astTypeAssertExpr":
		emitExpr(e.typeAssert.X, nil)
		fmtPrintf("  popq %%rax # type id\n")
		fmtPrintf("  popq %%rcx # data\n")
		fmtPrintf("  pushq %%rax # type id\n")
		fmtPrintf("  callq main.nop\n")
		typ := e2t(e.typeAssert.Type)
		sType := serializeType(typ)
		_id := getTypeId(sType)
		typeSymbol := typeIdToSymbol(_id)
		// check if type matches
		fmtPrintf("  leaq %s(%%rip), %%rax # typeid\n", typeSymbol)
		fmtPrintf("  pushq %%rax # type id\n")
		emitCompExpr("sete") // this pushes 1 or 0 in the end
		emitPopBool("type assertion ok value")
		fmtPrintf("  cmpq $1, %%rax\n")

		labelid++
		labelTypeAssertionEnd := fmtSprintf(".L.end_type_assertion.%d", []string{Itoa(labelid)})
		labelElse := fmtSprintf(".L.unmatch.%d", []string{Itoa(labelid)})
		fmtPrintf("  jne %s # jmp if false\n", labelElse)

		// if matched
		if ctx.okContext != nil {
			emitComment(2, " double value context\n")
			if ctx.okContext.needMain {
				emitExpr(e.typeAssert.X, nil)
				fmtPrintf("  popq %%rax # garbage\n")
				emitLoad(e2t(e.typeAssert.Type)) // load dynamic data
			}
			if ctx.okContext.needOk {
				fmtPrintf("  pushq $1 # ok = true\n")
			}
		} else {
			emitComment(2, " single value context\n")
			emitExpr(e.typeAssert.X, nil)
			fmtPrintf("  popq %%rax # garbage\n")
			emitLoad(e2t(e.typeAssert.Type)) // load dynamic data
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
		panic2(__func__, "[emitExpr] `TBI:"+e.dtype)
	}

	return isNilObject
}

func emitExprIfc(expr *astExpr, ctx *evalContext) {
	isNilObj := emitExpr(expr, ctx)
	if !isNilObj && ctx != nil {
		sourceType := getTypeOfExpr(expr)
		if ctx._type != nil{
			if isInterface(ctx._type) && !isInterface(sourceType) {
				emitComment(2, "ConversionToInterface\n")
				memSize := getSizeOfType(sourceType)
				// copy data to heap
				emitCallMalloc(memSize)
				emitStore(sourceType, false, true) // heap addr pushed
				// push type id
				emitTypeId(sourceType)
			}
		}
	}
}

var typeMap []*typeEntry
type typeEntry struct {
	serialized string
	id int
}

var typeId int = 1

func typeIdToSymbol(id int) string {
	return "dtype." + Itoa(id)
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

func emitTypeId(t *Type) {
	str := serializeType(t)
	typeId := getTypeId(str)
	typeSymbol := typeIdToSymbol(typeId)
	fmtPrintf("  leaq %s(%%rip), %%rax # typeid \"%s\"\n", typeSymbol, str)
	fmtPrintf("  pushq %%rax # type symbol %s\n", Itoa(typeId), typeSymbol)
}

func newNumberLiteral(x int) *astBasicLit {
	var r = &astBasicLit{}
	r.Kind = "INT"
	r.Value = Itoa(x)
	return r
}

func emitListElementAddr(list *astExpr, elmType *Type) {
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
		var resultList = []*astField{
			&astField{
				Type:    tBool.e,
			},
		}
		fmtPrintf("  callq runtime.cmpstrings\n")
		emitRevertStackPointer(stringSize * 2)
		emitReturnedValue(resultList)
	case T_INTERFACE:
		var resultList = []*astField{
			&astField{
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
		panic2(__func__, "Unexpected kind="+kind(t))
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
	fmtPrintf("\n")
	emitComment(2, "Statement %s\n", stmt.dtype)
	switch stmt.dtype {
	case "*astBlockStmt":
		for _, stmt2 := range stmt.blockStmt.List {
			emitStmt(stmt2)
		}
	case "*astExprStmt":
		emitExpr(stmt.exprStmt.X, nil)
	case "*astDeclStmt":
		var decl *astDecl = stmt.DeclStmt.Decl
		if decl.dtype != "*astGenDecl" {
			panic2(__func__, "[*astDeclStmt] internal error")
		}
		var genDecl = decl.genDecl
		var valSpec = genDecl.Spec.valueSpec
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
		//fmtPrintf("[emitStmt] TBI declSpec:%s\n", valueSpec.Name.Name)
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
		if len(stmt.returnStmt.Results) == 0 {
			fmtPrintf("  leave\n")
			fmtPrintf("  ret\n")
		} else if len(stmt.returnStmt.Results) == 1 {
			emitExpr(stmt.returnStmt.Results[0], nil) // @TODO forceType should be fetched from func decl
			var knd = kind(getTypeOfExpr(stmt.returnStmt.Results[0]))
			switch knd {
			case T_BOOL, T_INT, T_UINTPTR, T_POINTER:
				fmtPrintf("  popq %%rax # return 64bit\n")
			case T_STRING:
				fmtPrintf("  popq %%rax # return string (ptr)\n")
				fmtPrintf("  popq %%rdi # return string (len)\n")
			case T_SLICE:
				fmtPrintf("  popq %%rax # return string (ptr)\n")
				fmtPrintf("  popq %%rdi # return string (len)\n")
				fmtPrintf("  popq %%rsi # return string (cap)\n")
			default:
				panic2(__func__, "[*astReturnStmt] TBI:"+knd)
			}
			fmtPrintf("  leave\n")
			fmtPrintf("  ret\n")
		} else if len(stmt.returnStmt.Results) == 3 {
			// Special treatment to return a slice
			emitExpr(stmt.returnStmt.Results[2], nil) // @FIXME
			emitExpr(stmt.returnStmt.Results[1], nil) // @FIXME
			emitExpr(stmt.returnStmt.Results[0], nil) // @FIXME
			fmtPrintf("  popq %%rax # return 64bit\n")
			fmtPrintf("  popq %%rdi # return 64bit\n")
			fmtPrintf("  popq %%rsi # return 64bit\n")
		} else {
			panic2(__func__, "[*astReturnStmt] TBI\n")
		}
	case "*astIfStmt":
		emitComment(2, "if\n")

		labelid++
		var labelEndif = ".L.endif." + Itoa(labelid)
		var labelElse = ".L.else." + Itoa(labelid)

		emitExpr(stmt.ifStmt.Cond, nil)
		emitPopBool("if condition")
		fmtPrintf("  cmpq $1, %%rax\n")
		var bodyStmt = &astStmt{}
		bodyStmt.dtype = "*astBlockStmt"
		bodyStmt.blockStmt = stmt.ifStmt.Body
		if stmt.ifStmt.Else != nil {
			fmtPrintf("  jne %s # jmp if false\n", labelElse)
			emitStmt(bodyStmt) // then
			fmtPrintf("  jmp %s\n", labelEndif)
			fmtPrintf("  %s:\n", labelElse)
			emitStmt(stmt.ifStmt.Else) // then
		} else {
			fmtPrintf("  jne %s # jmp if false\n", labelEndif)
			emitStmt(bodyStmt) // then
		}
		fmtPrintf("  %s:\n", labelEndif)
		emitComment(2, "end if\n")
	case "*astForStmt":
		labelid++
		var labelCond = ".L.for.cond." + Itoa(labelid)
		var labelPost = ".L.for.post." + Itoa(labelid)
		var labelExit = ".L.for.exit." + Itoa(labelid)
		//forStmt, ok := mapForNodeToFor[s]
		//assert(ok, "map value should exist")
		stmt.forStmt.labelPost = labelPost
		stmt.forStmt.labelExit = labelExit

		if stmt.forStmt.Init != nil {
			emitStmt(stmt.forStmt.Init)
		}

		fmtPrintf("  %s:\n", labelCond)
		if stmt.forStmt.Cond != nil {
			emitExpr(stmt.forStmt.Cond, nil)
			emitPopBool("for condition")
			fmtPrintf("  cmpq $1, %%rax\n")
			fmtPrintf("  jne %s # jmp if false\n", labelExit)
		}
		emitStmt(blockStmt2Stmt(stmt.forStmt.Body))
		fmtPrintf("  %s:\n", labelPost) // used for "continue"
		if stmt.forStmt.Post != nil {
			emitStmt(stmt.forStmt.Post)
		}
		fmtPrintf("  jmp %s\n", labelCond)
		fmtPrintf("  %s:\n", labelExit)
	case "*astRangeStmt": // only for array and slice
		labelid++
		var labelCond = ".L.range.cond." + Itoa(labelid)
		var labelPost = ".L.range.post." + Itoa(labelid)
		var labelExit = ".L.range.exit." + Itoa(labelid)

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
		fmtPrintf("  %s:\n", labelCond)

		emitVariableAddr(stmt.rangeStmt.indexvar)
		emitLoad(tInt)
		emitVariableAddr(stmt.rangeStmt.lenvar)
		emitLoad(tInt)
		emitCompExpr("setl")
		emitPopBool(" indexvar < lenvar")
		fmtPrintf("  cmpq $1, %%rax\n")
		fmtPrintf("  jne %s # jmp if false\n", labelExit)

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
		fmtPrintf("  %s:\n", labelPost)           // used for "continue"
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

		fmtPrintf("  jmp %s\n", labelCond)

		fmtPrintf("  %s:\n", labelExit)

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
		var labelEnd = fmtSprintf(".L.switch.%s.exit", []string{Itoa(labelid)})
		if stmt.switchStmt.Tag == nil {
			panic2(__func__, "Omitted tag is not supported yet")
		}
		emitExpr(stmt.switchStmt.Tag, nil)
		var condType = getTypeOfExpr(stmt.switchStmt.Tag)
		var cases = stmt.switchStmt.Body.List
		emitComment(2, "[DEBUG] cases len=%s\n", Itoa(len(cases)))
		var labels = make([]string, len(cases), len(cases))
		var defaultLabel string
		emitComment(2, "Start comparison with cases\n")
		for i, c := range cases {
			emitComment(2, "CASES idx=%s\n", Itoa(i))
			assert(c.dtype == "*astCaseClause", "should be *astCaseClause", __func__)
			var cc = c.caseClause
			labelid++
			var labelCase = ".L.case." + Itoa(labelid)
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

		emitRevertStackTop(condType)
		for i, c := range cases {
			assert(c.dtype == "*astCaseClause", "should be *astCaseClause", __func__)
			var cc = c.caseClause
			fmtPrintf("%s:\n", labels[i])
			for _, _s := range cc.Body {
				emitStmt(_s)
			}
			fmtPrintf("  jmp %s\n", labelEnd)
		}
		fmtPrintf("%s:\n", labelEnd)
	case "*astBranchStmt":
		var containerFor = stmt.branchStmt.currentFor
		var labelToGo string
		switch stmt.branchStmt.Tok {
		case "continue":
			switch containerFor.dtype {
			case "*astForStmt":
				labelToGo = containerFor.forStmt.labelPost
			case "*astRangeStmt":
				labelToGo = containerFor.rangeStmt.labelPost
			default:
				panic2(__func__, "unexpected container dtype="+containerFor.dtype)
			}
			fmtPrintf("jmp %s # continue\n", labelToGo)
		case "break":
			switch containerFor.dtype {
			case "*astForStmt":
				labelToGo = containerFor.forStmt.labelExit
			case "*astRangeStmt":
				labelToGo = containerFor.rangeStmt.labelExit
			default:
				panic2(__func__, "unexpected container dtype="+containerFor.dtype)
			}
			fmtPrintf("jmp %s # break\n", labelToGo)
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
	fmtPrintf("  addq $%s, %%rsp # revert stack top\n", Itoa(getSizeOfType(t)))
}

var labelid int

func getMethodSymbol(method *Method) string {
	var rcvTypeName = method.rcvNamedType
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

func getFuncSymbol(pkgName string, subsymbol string) string {
	return pkgName + "." + subsymbol
}

func emitFuncDecl(pkgName string, fnc *Func) {
	var localarea = fnc.localarea
	fmtPrintf("\n")
	var subsymbol = getFuncSubSymbol(fnc)
	var symbol = getFuncSymbol(pkgName, subsymbol)
	fmtPrintf("%s: # args %d, locals %d\n",
		symbol, Itoa(int(fnc.argsarea)), Itoa(int(fnc.localarea)))

	fmtPrintf("  pushq %%rbp\n")
	fmtPrintf("  movq %%rsp, %%rbp\n")
	if localarea != 0 {
		fmtPrintf("  subq $%d, %%rsp # local area\n", Itoa(-localarea))
	}

	if fnc.Body != nil {
		emitStmt(blockStmt2Stmt(fnc.Body))
	}

	fmtPrintf("  leave\n")
	fmtPrintf("  ret\n")
}

func emitGlobalVariableComplex(name *astIdent, t *Type, val *astExpr) {
	typeKind := kind(t)
	switch typeKind {
	case T_POINTER:
		fmtPrintf("# init global %s: # T %s\n", name.Name, typeKind)
		lhs := &astExpr{
			ident: name,
			dtype: "*astIdent",
		}
		emitAssign(lhs, val)
	}
}

func emitGlobalVariable(pkg *PkgContainer, name *astIdent, t *Type, val *astExpr) {
	typeKind := kind(t)
	fmtPrintf("%s.%s: # T %s\n", pkg.name, name.Name, typeKind)
	switch typeKind {
	case T_STRING:
		if val == nil {
			fmtPrintf("  .quad 0\n")
			fmtPrintf("  .quad 0\n")
			return
		}
		switch val.dtype {
		case "*astBasicLit":
			var sl = getStringLiteral(val.basicLit)
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
		if val == nil {
			fmtPrintf("  .quad 0 # bool zero value\n")
			return
		}
		switch val.dtype {
		case "*astIdent":
			switch val.ident.Obj {
			case gTrue:
				fmtPrintf("  .quad 1 # bool true\n")
			case gFalse:
				fmtPrintf("  .quad 0 # bool false\n")
			default:
				panic2(__func__, "")
			}
		default:
			panic2(__func__, "")
		}
	case T_INT:
		if val == nil {
			fmtPrintf("  .quad 0\n")
			return
		}
		switch val.dtype {
		case "*astBasicLit":
			fmtPrintf("  .quad %s\n", val.basicLit.Value)
		default:
			panic("Unsupported global value")
		}
	case T_UINT8:
		if val == nil {
			fmtPrintf("  .byte 0\n")
			return
		}
		switch val.dtype {
		case "*astBasicLit":
			fmtPrintf("  .byte %s\n", val.basicLit.Value)
		default:
			panic("Unsupported global value")
		}
	case T_UINT16:
		if val == nil {
			fmtPrintf("  .word 0\n")
			return
		}
		switch val.dtype {
		case "*astBasicLit":
			fmtPrintf("  .word %s\n", val.basicLit.Value)
		default:
			panic("Unsupported global value")
		}
	case T_POINTER:
		// will be set in the initGlobal func
		fmtPrintf("  .quad 0\n")
	case T_UINTPTR:
		if val != nil {
			panic("Unsupported global value")
		}
		// only zero value
		fmtPrintf("  .quad 0\n")
	case T_SLICE:
		if val != nil {
			panic("Unsupported global value")
		}
		// only zero value
		fmtPrintf("  .quad 0 # ptr\n")
		fmtPrintf("  .quad 0 # len\n")
		fmtPrintf("  .quad 0 # cap\n")
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
		var basicLit = arrayType.Len.basicLit
		if len(basicLit.Value) > 1 {
			panic2(__func__, "array length >= 10 is not supported yet.")
		}
		var length = evalInt(arrayType.Len)
		emitComment(0, "[emitGlobalVariable] array length uint8=%s\n", Itoa(length))
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
			fmtPrintf(zeroValue)
		}
	default:
		panic2(__func__, "TBI:kind="+typeKind)
	}
}

func generateCode(pkg *PkgContainer) {
	fmtPrintf(".data\n")
	emitComment(0, "string literals len = %s\n", Itoa(len(stringLiterals)))
	for _, con := range stringLiterals {
		emitComment(0, "string literals\n")
		fmtPrintf("%s:\n", con.sl.label)
		fmtPrintf("  .string %s\n", con.sl.value)
	}

	for _, spec := range pkg.vars {
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariable(pkg, spec.Name, t, spec.Value)
	}

	fmtPrintf("\n")
	fmtPrintf(".text\n")
	fmtPrintf("%s.__initGlobals:\n", pkg.name)
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
	fmtPrintf("  ret\n")

	for _, fnc := range pkg.funcs {
		emitFuncDecl(pkg.name, fnc)
	}
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
const T_UINT8 string = "T_UINT8"
const T_UINT16 string = "T_UINT16"
const T_UINTPTR string = "T_UINTPTR"
const T_ARRAY string = "T_ARRAY"
const T_STRUCT string = "T_STRUCT"
const T_POINTER string = "T_POINTER"

func getTypeOfExpr(expr *astExpr) *Type {
	//emitComment(0, "[%s] start\n", __func__)
	switch expr.dtype {
	case "*astIdent":
		if expr.ident.Obj == nil {
			panic(expr.ident.Name)
		}
		switch expr.ident.Obj.Kind {
		case astVar:
			switch expr.ident.Obj.Decl.dtype {
			case "*astValueSpec":
				var decl = expr.ident.Obj.Decl.valueSpec
				var t = &Type{}
				t.e = decl.Type
				return t
			case "*astField":
				var decl = expr.ident.Obj.Decl.field
				var t = &Type{}
				t.e = decl.Type
				return t
			case "*astAssignStmt": // lhs := rhs
				return getTypeOfExpr(expr.ident.Obj.Decl.assignment.Rhs[0])
			default:
				panic2(__func__, "dtype:" + expr.ident.Obj.Decl.dtype )
			}
		case astCon:
			switch expr.ident.Obj {
			case gTrue, gFalse:
				return tBool
			}
			switch expr.ident.Obj.Decl.dtype {
			case "*astValueSpec":
				return e2t(expr.ident.Obj.Decl.valueSpec.Type)
			default:
				panic2(__func__, "cannot decide type of cont ="+expr.ident.Obj.Name)
			}
		default:
			panic2(__func__, "2:Obj.Kind="+expr.ident.Obj.Kind)
		}
	case "*astBasicLit":
		switch expr.basicLit.Kind {
		case "STRING":
			return tString
		case "INT":
			return tInt
		case "CHAR":
			return tInt
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
		switch fun.dtype {
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
				switch decl.dtype {
				case "*astFuncDecl":
					var resultList = decl.funcDecl.Type.Results.List
					if len(resultList) != 1 {
						panic2(__func__, "[astCallExpr] len results.List is not 1")
					}
					return e2t(decl.funcDecl.Type.Results.List[0].Type)
				default:
					panic2(__func__, "[astCallExpr] decl.dtype="+decl.dtype)
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
			var funcType *astFuncType
			symbol := xIdent.Name + "." + fun.selectorExpr.Sel.Name
			switch symbol {
			case "unsafe.Pointer":
				// unsafe.Pointer(x)
				return tUintptr
			case "os.Exit":
				funcType = funcTypeOsExit
				var r *Type
				return r
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
				var xType = getTypeOfExpr(fun.selectorExpr.X)
				var method = lookupMethod(xType, fun.selectorExpr.Sel)
				assert(len(method.funcType.Results.List) == 1, "func is expected to return a single value", __func__)
				return e2t(method.funcType.Results.List[0].Type)
			}
		default:
			panic2(__func__, "[astCallExpr] dtype="+expr.callExpr.Fun.dtype)
		}
	case "*astSliceExpr":
		var underlyingCollectionType = getTypeOfExpr(expr.sliceExpr.X)
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
		if isOsArgs(expr.selectorExpr) {
			// os.Args
			return tSliceOfString
		}
		var structType = getStructTypeOfX(expr.selectorExpr)
		var field = lookupStructField(getStructTypeSpec(structType), expr.selectorExpr.Sel.Name)
		return e2t(field.Type)
	case "*astCompositeLit":
		return e2t(expr.compositeLit.Type)
	case "*astParenExpr":
		return getTypeOfExpr(expr.parenExpr.X)
	case "*astTypeAssertExpr":
		return e2t(expr.typeAssert.Type)
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
				if decl.dtype != "*astTypeSpec" {
					panic("unexpected dtype" + decl.dtype)
				}
				return "main." + decl.typeSpec.Name.Name
			}
		}
	case "*astStructType":
		return "struct"
	case "*astArrayType":
		e := t.e.arrayType
		if e.Len == nil {
			return "[]" + serializeType(e2t(e.Elt))
		} else {
			return "[" + Itoa(evalInt(e.Len)) + "]" + serializeType(e2t(e.Elt))
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
		case "string":
			return T_STRING
		case "uint8":
			return T_UINT8
		case "uint16":
			return T_UINT16
		case "bool":
			return T_BOOL
		default:
			// named type
			var decl = ident.Obj.Decl
			if decl.dtype != "*astTypeSpec" {
				panic2(__func__, "unsupported decl :"+decl.dtype)
			}
			var typeSpec = decl.typeSpec
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
		var arrayType = t.e.arrayType
		if arrayType == nil {
			panic2(__func__, "should not be nil")
		}
		return e2t(arrayType.Elt)
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

	if ident.Obj.Decl.dtype != "*astTypeSpec" {
		panic2(__func__, "not *astTypeSpec")
	}

	var typeSpec = ident.Obj.Decl.typeSpec
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
}

//type localoffsetint int //@TODO

var stringLiterals []*stringLiteralsContainer
var stringIndex int
var localoffset int
var currentFuncDecl *astFuncDecl

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

	label := fmtSprintf(".%s.S%d", []string{pkg.name, Itoa(stringIndex)})
	stringIndex++

	sl := &sliteral{}
	sl.label = label
	sl.strlen = strlen - 2
	sl.value = lit.Value
	logf(" [registerStringLiteral] label=%s, strlen=%s\n", sl.label, Itoa(sl.strlen))
	cont := &stringLiteralsContainer{}
	cont.sl = sl
	cont.lit = lit
	stringLiterals = append(stringLiterals, cont)
}

func newGlobalVariable(pkgName string, name string) *Variable {
	vr := &Variable{
		name:         name,
		isGlobal:     true,
		globalSymbol: pkgName + "." + name,
	}
	return vr
}

func newLocalVariable(name string, localoffset int) *Variable {
	vr := &Variable{
		name : name,
		isGlobal : false,
		localOffset : localoffset,
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

func newMethod(funcDecl *astFuncDecl) *Method {
	var rcvType = funcDecl.Recv.List[0].Type
	var isPtr bool
	if rcvType.dtype == "*astStarExpr" {
		isPtr = true
		rcvType = rcvType.starExpr.X
	}
	assert(rcvType.dtype == "*astIdent", "receiver type should be ident", __func__)
	var rcvNamedType = rcvType.ident

	var method = &Method{
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
	switch stmt.dtype {
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
		if dcl.dtype != "*astGenDecl" {
			panic2(__func__, "[dcl.dtype] internal error")
		}
		var valSpec = dcl.genDecl.Spec.valueSpec
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

		var t = e2t(typ)
		var sizeOfType = getSizeOfType(t)
		localoffset = localoffset - sizeOfType

		valSpec.Name.Obj.Variable = newLocalVariable(valSpec.Name.Name, localoffset)
		logf(" var %s offset = %d\n", valSpec.Name.Obj.Name,
			Itoa(valSpec.Name.Obj.Variable.localOffset))
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
			localoffset = localoffset - getSizeOfType(typ)
			obj.Variable = newLocalVariable(obj.Name, localoffset)
		} else {
			walkExpr(rhs)
		}
	case "*astExprStmt":
		walkExpr(stmt.exprStmt.X)
	case "*astReturnStmt":
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
		localoffset = localoffset - intSize
		var lenvar = newLocalVariable(".range.len", localoffset)
		localoffset = localoffset - intSize
		var indexvar = newLocalVariable(".range.index", localoffset)

		if stmt.rangeStmt.Tok == ":=" {
			listType := getTypeOfExpr(stmt.rangeStmt.X)

			keyIdent := stmt.rangeStmt.Key.ident
			//@TODO map key can be any type
			//keyType := getKeyTypeOfListType(listType)
			var keyType *Type = tInt
			localoffset = localoffset - getSizeOfType(keyType)
			keyIdent.Obj.Variable = newLocalVariable(keyIdent.Name, localoffset)

			// determine type of Value
			elmType := getElementTypeOfListType(listType)
			valueIdent := stmt.rangeStmt.Value.ident
			localoffset = localoffset - getSizeOfType(elmType)
			valueIdent.Obj.Variable = newLocalVariable(valueIdent.Name, localoffset)
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
	switch expr.dtype {
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
					basicLit.Value = "\"" + currentFuncDecl.Name.Name + "\""
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
	case "*astTypeAssertExpr":
		walkExpr(expr.typeAssert.X)
	default:
		panic2(__func__, "TBI:"+expr.dtype)
	}
}

func walk(pkg *PkgContainer, file *astFile) {
	for _, decl := range file.Decls {
		switch decl.dtype {
		case "*astFuncDecl":
			var funcDecl = decl.funcDecl
			if funcDecl.Body != nil {
				if funcDecl.Recv != nil { // is Method
					var method = newMethod(funcDecl)
					registerMethod(method)
				}
			}
		}
	}
	for _, decl := range file.Decls {
		switch decl.dtype {
		case "*astGenDecl":
			var genDecl = decl.genDecl
			switch genDecl.Spec.dtype {
			case "*astValueSpec":
				var valSpec = genDecl.Spec.valueSpec
				var nameIdent = valSpec.Name
				if nameIdent.Obj.Kind == astVar {
					if valSpec.Type == nil {
						var val = valSpec.Value
						var t = getTypeOfExpr(val)
						valSpec.Type = t.e
					}
					nameIdent.Obj.Variable = newGlobalVariable(pkg.name, nameIdent.Obj.Name)
					pkg.vars = append(pkg.vars, valSpec)
				}
				if valSpec.Value != nil {
					walkExpr(valSpec.Value)
				}
			case "*astTypeSpec":
				// do nothing
				var typeSpec = genDecl.Spec.typeSpec
				switch kind(e2t(typeSpec.Type)) {
				case T_STRUCT:
					calcStructSizeAndSetFieldOffset(typeSpec)
				default:
					// do nothing
				}
			default:
				panic2(__func__, "Unexpected dtype="+genDecl.Spec.dtype)
			}
		case "*astFuncDecl":
			var funcDecl = decl.funcDecl
			currentFuncDecl = funcDecl
			logf(" [sema] == astFuncDecl %s ==\n", funcDecl.Name.Name)
			localoffset = 0
			var paramoffset = 16
			var paramFields []*astField

			if funcDecl.Recv != nil { // Method
				paramFields = append(paramFields, funcDecl.Recv.List[0])
			}
			for _, field := range funcDecl.Type.Params.List {
				paramFields = append(paramFields, field)
			}

			for _, field := range paramFields {
				var obj = field.Name.Obj
				obj.Variable = newLocalVariable(obj.Name, paramoffset)
				var varSize = getSizeOfType(e2t(field.Type))
				paramoffset = paramoffset + varSize
				logf(" field.Name.Obj.Name=%s\n", obj.Name)
				//logf("   field.Type=%#v\n", field.Type)
			}
			if funcDecl.Body != nil {
				for _, stmt := range funcDecl.Body.List {
					walkStmt(stmt)
				}
				var fnc = &Func{}
				fnc.name = funcDecl.Name.Name
				fnc.funcType =  funcDecl.Type
				fnc.Body = funcDecl.Body
				fnc.localarea = localoffset
				fnc.argsarea = paramoffset

				if funcDecl.Recv != nil { // Method
					fnc.method = newMethod(funcDecl)
				}
				pkg.funcs = append(pkg.funcs, fnc)
			}
		default:
			panic2(__func__, "TBI: "+decl.dtype)
		}
	}

	if len(stringLiterals) == 0 {
		panic2(__func__, "stringLiterals is empty\n")
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

// func type of runtime functions
var funcTypeOsExit = &astFuncType{
	Params: &astFieldList{
		List: []*astField{
			&astField{
				Type:  tInt.e,
			},
		},
	},
	Results: nil,
}
var funcTypeSyscallOpen = &astFuncType{
	Params: &astFieldList{
		List: []*astField{
			&astField{
				Type:  tString.e,
			},
			&astField{
				Type:  tInt.e,
			},
			&astField{
				Type:  tInt.e,
			},
		},
	},
	Results: &astFieldList{
		List: []*astField{
			&astField{
				Type:  tInt.e,
			},
		},
	},
}
var funcTypeSyscallRead = &astFuncType{
	Params: &astFieldList{
		List: []*astField{
			&astField{
				Type:  tInt.e,
			},
			&astField{
				Type: generalSlice,
			},
		},
	},
	Results: &astFieldList{
		List: []*astField{
			&astField{
				Type:  tInt.e,
			},
		},
	},
}
var funcTypeSyscallWrite = &astFuncType{
	Params: &astFieldList{
		List: []*astField{
			&astField{
				Type:  tInt.e,
			},
			&astField{
				Type: generalSlice,
			},
		},
	},
	Results: &astFieldList{
		List: []*astField{
			&astField{
				Type:  tInt.e,
			},
		},
	},
}
var funcTypeSyscallSyscall = &astFuncType{
	Params: &astFieldList{
		List: []*astField{
			&astField{
				Type:  tUintptr.e,
			},
			&astField{
				Type:  tUintptr.e,
			},
			&astField{
				Type:  tUintptr.e,
			},
			&astField{
				Type:  tUintptr.e,
			},
		},
	},
	Results: &astFieldList{
		List: []*astField{
			&astField{
				Type:  tUintptr.e,
			},
		},
	},
}

func createUniverse() *astScope {
	var universe = new(astScope)

	scopeInsert(universe, gInt)
	scopeInsert(universe, gUint8)
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
	logf(" [%s] start\n", __func__)
	// create universe scope
	// inject predeclared identifers
	var unresolved []*astIdent
	logf(" [SEMA] resolving file.Unresolved (n=%s)\n", Itoa(len(file.Unresolved)))
	for _, ident := range file.Unresolved {
		logf(" [SEMA] resolving ident %s ... \n", ident.Name)
		var obj *astObject = scopeLookup(universe, ident.Name)
		if obj != nil {
			logf(" matched\n")
			ident.Obj = obj
		} else {
			// we should allow unresolved for now.
			// e.g foo in X{foo:bar,}
			logf("Unresolved (maybe struct field name in composite literal): "+ident.Name)
			unresolved = append(unresolved, ident)
		}
	}
}

var pkg *PkgContainer

type PkgContainer struct {
	name string
	vars []*astValueSpec
	funcs []*Func
}

func showHelp() {
	fmtPrintf("Usage:\n")
	fmtPrintf("    babygo version:  show version\n")
	fmtPrintf("    babygo [-DF] [-DG] filename\n")
}

func main() {
	var universe = createUniverse()
	if len(os.Args) == 1 {
		showHelp()
		return
	}

	if os.Args[1] == "version" {
		fmtPrintf("babygo version 0.1.0  linux/amd64\n")
		return
	} else if os.Args[1] == "help" {
		showHelp()
		return
	} else if os.Args[1] == "panic" {
		panic("I am panic")
	}

	var arg string
	for _, arg = range os.Args {
		switch arg {
		case "-DF":
			debugFrontEnd = true
		case "-DG":
			debugCodeGen = true
		}
	}
	var inputFile = arg
	var sourceFiles = []string{"runtime.go", inputFile}

	for _, sourceFile := range sourceFiles {
		fmtPrintf("# file: %s\n", sourceFile)
		stringIndex = 0
		stringLiterals = nil
		astFile := parseFile(sourceFile)
		resolveUniverse(astFile, universe)
		pkg = &PkgContainer{
			name: astFile.Name,
		}
		walk(pkg, astFile)
		generateCode(pkg)
	}

	// emitting dynamic types
	fmtPrintf("# ------- Dynamic Types ------\n")
	fmtPrintf(".data\n")
	for _, te := range typeMap {
		id := te.id
		name := te.serialized
		symbol := typeIdToSymbol(id)
		fmtPrintf("%s: # %s\n", symbol, name)
		fmtPrintf("  .quad %s\n", Itoa(id))
		fmtPrintf("  .quad .S.dtype.%s\n", Itoa(id))
		fmtPrintf(".S.dtype.%s:\n", Itoa(id))
		fmtPrintf("  .string \"%s\"\n", name)
	}
	fmtPrintf("\n")
}
