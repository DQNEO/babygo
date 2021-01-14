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

var __func__ string = "__func__"

func panic(x string) {
	var s = "panic: " + x + "\n\n"
	syscall.Write(1, []uint8(s))
	os.Exit(1)
}

func panic2(caller string, x string) {
	panic("[" + caller + "] " + x)
}

// --- libs ---
func fmtSprintf(format string, a []string) string {
	var buf []uint8
	var inPercent bool
	var argIndex int
	var c uint8
	for _, c = range []uint8(format) {
		if inPercent {
			if c == '%' {
				buf = append(buf, c)
			} else {
				var arg = a[argIndex]
				argIndex++
				var s = arg // // p.printArg(arg, c)
				var _c uint8
				for _, _c = range []uint8(s) {
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
	var b uint8
	var n int

	var isMinus bool
	for _, b = range []uint8(gs) {
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
	var v string
	for _, v = range list {
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

func scannerNext(s *scanner) {
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

func scannerInit(s *scanner, src []uint8) {
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
	scannerNext(s)
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

func scannerScanIdentifier(s *scanner) string {
	var offset = s.offset
	for isLetter(s.ch) || isDecimal(s.ch) {
		scannerNext(s)
	}
	return string(s.src[offset:s.offset])
}

func scannerScanNumber(s *scanner) string {
	var offset = s.offset
	for isDecimal(s.ch) {
		scannerNext(s)
	}
	return string(s.src[offset:s.offset])
}

func scannerScanString(s *scanner) string {
	var offset = s.offset - 1
	var escaped bool
	for !escaped && s.ch != '"' {
		if s.ch == '\\' {
			escaped = true
			scannerNext(s)
			scannerNext(s)
			escaped = false
			continue
		}
		scannerNext(s)
	}
	scannerNext(s) // consume ending '""
	return string(s.src[offset:s.offset])
}

func scannerScanChar(s *scanner) string {
	// '\'' opening already consumed
	var offset = s.offset - 1
	var ch uint8
	for {
		ch = s.ch
		scannerNext(s)
		if ch == '\'' {
			break
		}
		if ch == '\\' {
			scannerNext(s)
		}
	}

	return string(s.src[offset:s.offset])
}

func scannerrScanComment(s *scanner) string {
	var offset = s.offset - 1
	for s.ch != '\n' {
		scannerNext(s)
	}
	return string(s.src[offset:s.offset])
}

type TokenContainer struct {
	pos int    // what's this ?
	tok string // token.Token
	lit string // raw data
}

// https://golang.org/ref/spec#Tokens
func scannerSkipWhitespace(s *scanner) {
	for s.ch == ' ' || s.ch == '\t' || (s.ch == '\n' && !s.insertSemi) || s.ch == '\r' {
		scannerNext(s)
	}
}

func scannerScan(s *scanner) *TokenContainer {
	scannerSkipWhitespace(s)
	var tc = &TokenContainer{}
	var lit string
	var tok string
	var insertSemi bool
	var ch = s.ch
	if isLetter(ch) {
		lit = scannerScanIdentifier(s)
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
		lit = scannerScanNumber(s)
		tok = "INT"
	} else {
		scannerNext(s)
		switch ch {
		case '\n':
			tok = ";"
			lit = "\n"
			insertSemi = false
		case '"': // double quote
			insertSemi = true
			lit = scannerScanString(s)
			tok = "STRING"
		case '\'': // single quote
			insertSemi = true
			lit = scannerScanChar(s)
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
				scannerNext(s)
				tok = ":="
			} else {
				tok = ":"
			}
		case '.': // ..., .
			var peekCh = s.src[s.nextOffset]
			if s.ch == '.' && peekCh == '.' {
				scannerNext(s)
				scannerNext(s)
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
				scannerNext(s)
				tok = "+="
			case '+':
				scannerNext(s)
				tok = "++"
				insertSemi = true
			default:
				tok = "+"
			}
		case '-': // -= --  -
			switch s.ch {
			case '-':
				scannerNext(s)
				tok = "--"
				insertSemi = true
			case '=':
				scannerNext(s)
				tok = "-="
			default:
				tok = "-"
			}
		case '*': // *=  *
			if s.ch == '=' {
				scannerNext(s)
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
				lit = scannerrScanComment(s)
				tok = "COMMENT"
			} else if s.ch == '=' {
				tok = "/="
			} else {
				tok = "/"
			}
		case '%': // %= %
			if s.ch == '=' {
				scannerNext(s)
				tok = "%="
			} else {
				tok = "%"
			}
		case '^': // ^= ^
			if s.ch == '=' {
				scannerNext(s)
				tok = "^="
			} else {
				tok = "^"
			}
		case '<': //  <= <- <<= <<
			switch s.ch {
			case '-':
				scannerNext(s)
				tok = "<-"
			case '=':
				scannerNext(s)
				tok = "<="
			case '<':
				var peekCh = s.src[s.nextOffset]
				if peekCh == '=' {
					scannerNext(s)
					scannerNext(s)
					tok = "<<="
				} else {
					scannerNext(s)
					tok = "<<"
				}
			default:
				tok = "<"
			}
		case '>': // >= >>= >> >
			switch s.ch {
			case '=':
				scannerNext(s)
				tok = ">="
			case '>':
				var peekCh = s.src[s.nextOffset]
				if peekCh == '=' {
					scannerNext(s)
					scannerNext(s)
					tok = ">>="
				} else {
					scannerNext(s)
					tok = ">>"
				}
			default:
				tok = ">"
			}
		case '=': // == =
			if s.ch == '=' {
				scannerNext(s)
				tok = "=="
			} else {
				tok = "="
			}
		case '!': // !=, !
			if s.ch == '=' {
				scannerNext(s)
				tok = "!="
			} else {
				tok = "!"
			}
		case '&': // & &= && &^ &^=
			switch s.ch {
			case '=':
				scannerNext(s)
				tok = "&="
			case '&':
				scannerNext(s)
				tok = "&&"
			case '^':
				var peekCh = s.src[s.nextOffset]
				if peekCh == '=' {
					scannerNext(s)
					scannerNext(s)
					tok = "&^="
				} else {
					scannerNext(s)
					tok = "&^"
				}
			default:
				tok = "&"
			}
		case '|': // |= || |
			switch s.ch {
			case '|':
				scannerNext(s)
				tok = "||"
			case '=':
				scannerNext(s)
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
}

type astObject struct {
	Kind     string
	Name     string
	Decl     *ObjDecl
	Variable *Variable
}

type astExpr struct {
	dtype        string
	ident        *astIdent
	arrayType    *astArrayType
	basicLit     *astBasicLit
	callExpr     *astCallExpr
	binaryExpr   *astBinaryExpr
	unaryExpr    *astUnaryExpr
	selectorExpr *astSelectorExpr
	indexExpr    *astIndexExpr
	sliceExpr    *astSliceExpr
	starExpr     *astStarExpr
	parenExpr    *astParenExpr
	structType   *astStructType
	compositeLit *astCompositeLit
	keyValueExpr *astKeyValueExpr
	ellipsis     *astEllipsis
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

// Type nodes
type astArrayType struct {
	Len *astExpr
	Elt *astExpr
}

type astStructType struct {
	Fields *astFieldList
}

type astFuncType struct {
	params  *astFieldList
	results *astFieldList
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
	var oe *objectEntry
	for _, oe = range s.Objects {
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

func parserInit(src []uint8) {
	scannerInit(p.scanner, src)
	parserNext()
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

var p *parser

func openScope() {
	p.topScope = astNewScope(p.topScope)
}

func closeScope() {
	p.topScope = p.topScope.Outer
}

func parserConsumeComment() {
	parserNext0()
}

func parserNext0() {
	p.tok = scannerScan(p.scanner)
}

func parserNext() {
	parserNext0()
	if p.tok.tok == ";" {
		logf(" [parser] pointing at : \"%s\" newline (%s)\n", p.tok.tok, Itoa(p.scanner.offset))
	} else if p.tok.tok == "IDENT" {
		logf(" [parser] pointing at: IDENT \"%s\" (%s)\n", p.tok.lit, Itoa(p.scanner.offset))
	} else {
		logf(" [parser] pointing at: \"%s\" %s (%s)\n", p.tok.tok, p.tok.lit, Itoa(p.scanner.offset))
	}

	if p.tok.tok == "COMMENT" {
		for p.tok.tok == "COMMENT" {
			parserConsumeComment()
		}
	}
}

func parserExpect(tok string, who string) {
	if p.tok.tok != tok {
		var s = fmtSprintf("%s expected, but got %s", []string{tok, p.tok.tok})
		panic2(who, s)
	}
	logf(" [%s] consumed \"%s\"\n", who, p.tok.tok)
	parserNext()
}

func parserExpectSemi(caller string) {
	if p.tok.tok != ")" && p.tok.tok != "}" {
		switch p.tok.tok {
		case ";":
			logf(" [%s] consumed semicolon %s\n", caller, p.tok.tok)
			parserNext()
		default:
			panic2(caller, "semicolon expected, but got token "+p.tok.tok)
		}
	}
}

func parseIdent() *astIdent {
	var name string
	if p.tok.tok == "IDENT" {
		name = p.tok.lit
		parserNext()
	} else {
		panic2(__func__, "IDENT expected, but got "+p.tok.tok)
	}
	logf(" [%s] ident name = %s\n", __func__, name)
	return &astIdent{
		Name: name,
	}
}

func parserImportDecl() *astImportSpec {
	parserExpect("import", __func__)
	var path = p.tok.lit
	parserNext()
	parserExpectSemi(__func__)

	return &astImportSpec{
		Path: path,
	}
}

func tryVarType(ellipsisOK bool) *astExpr {
	if ellipsisOK && p.tok.tok == "..." {
		parserNext() // consume "..."
		var typ = tryIdentOrType()
		if typ != nil {
			parserResolve(typ)
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
	return tryIdentOrType()
}

func parseVarType(ellipsisOK bool) *astExpr {
	logf(" [%s] begin\n", __func__)
	var typ = tryVarType(ellipsisOK)
	if typ == nil {
		panic2(__func__, "nil is not expected")
	}
	logf(" [%s] end\n", __func__)
	return typ
}

func tryType() *astExpr {
	logf(" [%s] begin\n", __func__)
	var typ = tryIdentOrType()
	if typ != nil {
		parserResolve(typ)
	}
	logf(" [%s] end\n", __func__)
	return typ
}

func parseType() *astExpr {
	var typ = tryType()
	return typ
}

func parsePointerType() *astExpr {
	parserExpect("*", __func__)
	var base = parseType()
	return &astExpr{
		dtype: "*astStarExpr",
		starExpr : &astStarExpr{
			X: base,
		},
	}
}

func parseArrayType() *astExpr {
	parserExpect("[", __func__)
	var ln *astExpr
	if p.tok.tok != "]" {
		ln = parseRhs()
	}
	parserExpect("]", __func__)
	var elt = parseType()

	var r = &astExpr{
		dtype : "*astArrayType",
		arrayType : &astArrayType{
			Elt : elt,
			Len : ln,
		},
	}
	return r
}

func parseFieldDecl(scope *astScope) *astField {

	var varType = parseVarType(false)
	var typ = tryVarType(false)

	parserExpectSemi(__func__)

	var field = &astField{
		Type : typ,
		Name : varType.ident,
	}
	declareField(field, scope, astVar, varType.ident)
	parserResolve(typ)
	return field
}

func parseStructType() *astExpr {
	parserExpect("struct", __func__)
	parserExpect("{", __func__)

	var _nil *astScope
	var scope = astNewScope(_nil)

	var list []*astField
	for p.tok.tok == "IDENT" || p.tok.tok == "*" {
		var field *astField = parseFieldDecl(scope)
		list = append(list, field)
	}
	parserExpect("}", __func__)

	return &astExpr{
		dtype : "*astStructType",
		structType : &astStructType{
			Fields: &astFieldList{
				List : list,
			},
		},
	}
}

func parseTypeName() *astExpr {
	logf(" [%s] begin\n", __func__)
	var ident = parseIdent()
	logf(" [%s] end\n", __func__)
	return &astExpr{
		ident: ident,
		dtype: "*astIdent",
	}
}

func tryIdentOrType() *astExpr {
	logf(" [%s] begin\n", __func__)
	switch p.tok.tok {
	case "IDENT":
		return parseTypeName()
	case "[":
		return parseArrayType()
	case "struct":
		return parseStructType()
	case "*":
		return parsePointerType()
	case "(":
		parserNext()
		var _typ = parseType()
		parserExpect(")", __func__)
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

func parseParameterList(scope *astScope, ellipsisOK bool) []*astField {
	logf(" [%s] begin\n", __func__)
	var list []*astExpr
	for {
		var varType = parseVarType(ellipsisOK)
		list = append(list, varType)
		if p.tok.tok != "," {
			break
		}
		parserNext()
		if p.tok.tok == ")" {
			break
		}
	}
	logf(" [%s] collected list n=%s\n", __func__, Itoa(len(list)))

	var params []*astField

	var typ = tryVarType(ellipsisOK)
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
		parserResolve(typ)
		if p.tok.tok != "," {
			logf("  end %s\n", __func__)
			return params
		}
		parserNext()
		for p.tok.tok != ")" && p.tok.tok != "EOF" {
			ident = parseIdent()
			typ = parseVarType(ellipsisOK)
			field = &astField{
				Name : ident,
				Type : typ,
			}
			params = append(params, field)
			declareField(field, scope, astVar, ident)
			parserResolve(typ)
			if p.tok.tok != "," {
				break
			}
			parserNext()
		}
		logf("  end %s\n", __func__)
		return params
	}

	// Type { "," Type } (anonymous parameters)
	params = make([]*astField, len(list), len(list))
	var i int
	for i, typ = range list {
		parserResolve(typ)
		params[i] = &astField{
			Type: typ,
		}
		logf(" [DEBUG] range i = %s\n", Itoa(i))
	}
	logf("  end %s\n", __func__)
	return params
}

func parseParameters(scope *astScope, ellipsisOk bool) *astFieldList {
	logf(" [%s] begin\n", __func__)
	var params []*astField
	parserExpect("(", __func__)
	if p.tok.tok != ")" {
		params = parseParameterList(scope, ellipsisOk)
	}
	parserExpect(")", __func__)
	logf(" [%s] end\n", __func__)
	return &astFieldList{
		List: params,
	}
}

func parserResult(scope *astScope) *astFieldList {
	logf(" [%s] begin\n", __func__)

	if p.tok.tok == "(" {
		var r = parseParameters(scope, false)
		logf(" [%s] end\n", __func__)
		return r
	}

	if p.tok.tok == "{" {
		logf(" [%s] end\n", __func__)
		var _r *astFieldList = nil
		return _r
	}
	var typ = tryType()
	var list []*astField
	list = append(list, &astField{
		Type: typ,
	})
	logf(" [%s] end\n", __func__)
	return &astFieldList{
		List: list,
	}
}

func parseSignature(scope *astScope) *signature {
	logf(" [%s] begin\n", __func__)
	var params *astFieldList
	var results *astFieldList
	params = parseParameters(scope, true)
	results = parserResult(scope)
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

func parserResolve(x *astExpr) {
	tryResolve(x, true)
}
func tryResolve(x *astExpr, collectUnresolved bool) {
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

func parseOperand() *astExpr {
	logf("   begin %s\n", __func__)
	switch p.tok.tok {
	case "IDENT":
		var ident = parseIdent()
		var eIdent = &astExpr{
			dtype : "*astIdent",
			ident : ident,
		}
		tryResolve(eIdent, true)
		logf("   end %s\n", __func__)
		return eIdent
	case "INT", "STRING", "CHAR":
		var basicLit = &astBasicLit{
			Kind : p.tok.tok,
			Value : p.tok.lit,
		}
		parserNext()
		logf("   end %s\n", __func__)
		return &astExpr{
			dtype:    "*astBasicLit",
			basicLit: basicLit,
		}
	case "(":
		parserNext() // consume "("
		parserExprLev++
		var x = parserRhsOrType()
		parserExprLev--
		parserExpect(")", __func__)
		return &astExpr{
			dtype: "*astParenExpr",
			parenExpr: &astParenExpr{
				X: x,
			},
		}
	}

	var typ = tryIdentOrType()
	if typ == nil {
		panic2(__func__, "# typ should not be nil\n")
	}
	logf("   end %s\n", __func__)

	return typ
}

func parserRhsOrType() *astExpr {
	var x = parseExpr()
	return x
}

func parseCallExpr(fn *astExpr) *astExpr {
	parserExpect("(", __func__)
	logf(" [parsePrimaryExpr] p.tok.tok=%s\n", p.tok.tok)
	var list []*astExpr
	for p.tok.tok != ")" {
		var arg = parseExpr()
		list = append(list, arg)
		if p.tok.tok == "," {
			parserNext()
		} else if p.tok.tok == ")" {
			break
		}
	}
	parserExpect(")", __func__)
	return &astExpr{
		dtype:    "*astCallExpr",
		callExpr: &astCallExpr{
			Fun:  fn,
			Args: list,
		},
	}
}

var parserExprLev int // < 0: in control clause, >= 0: in expression

func parsePrimaryExpr() *astExpr {
	logf("   begin %s\n", __func__)
	var x = parseOperand()

	var cnt int

	for {
		cnt++
		logf("    [%s] tok=%s\n", __func__, p.tok.tok)
		if cnt > 100 {
			panic2(__func__, "too many iteration")
		}

		switch p.tok.tok {
		case ".":
			parserNext() // consume "."
			if p.tok.tok != "IDENT" {
				panic2(__func__, "tok should be IDENT")
			}
			// Assume CallExpr
			var secondIdent = parseIdent()
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
				x = parseCallExpr(fn)
				logf(" [parsePrimaryExpr] 741 p.tok.tok=%s\n", p.tok.tok)
			} else {
				logf("   end parsePrimaryExpr()\n")
				x = &astExpr{
					dtype : "*astSelectorExpr",
					selectorExpr : sel,
				}
			}
		case "(":
			x = parseCallExpr(x)
		case "[":
			parserResolve(x)
			x = parseIndexOrSlice(x)
		case "{":
			if isLiteralType(x) && parserExprLev >= 0 {
				x = parseLiteralValue(x)
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

func parserElement() *astExpr {
	var x = parseExpr() // key or value
	var v *astExpr
	var kvExpr *astKeyValueExpr
	if p.tok.tok == ":" {
		parserNext() // skip ":"
		v = parseExpr()
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

func parserElementList() []*astExpr {
	var list []*astExpr
	var e *astExpr
	for p.tok.tok != "}" {
		e = parserElement()
		list = append(list, e)
		if p.tok.tok != "," {
			break
		}
		parserExpect(",", __func__)
	}
	return list
}

func parseLiteralValue(typ *astExpr) *astExpr {
	logf("   start %s\n", __func__)
	parserExpect("{", __func__)
	var elts []*astExpr
	if p.tok.tok != "}" {
		elts = parserElementList()
	}
	parserExpect("}", __func__)

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

func parseIndexOrSlice(x *astExpr) *astExpr {
	parserExpect("[", __func__)
	var index = make([]*astExpr, 3, 3)
	if p.tok.tok != ":" {
		index[0] = parseRhs()
	}
	var ncolons int
	for p.tok.tok == ":" && ncolons < 2 {
		ncolons++
		parserNext() // consume ":"
		if p.tok.tok != ":" && p.tok.tok != "]" {
			index[ncolons] = parseRhs()
		}
	}
	parserExpect("]", __func__)

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

func parseUnaryExpr() *astExpr {
	var r *astExpr
	logf("   begin parseUnaryExpr()\n")
	switch p.tok.tok {
	case "+", "-", "!", "&":
		var tok = p.tok.tok
		parserNext()
		var x = parseUnaryExpr()
		r = &astExpr{}
		r.dtype = "*astUnaryExpr"
		r.unaryExpr = &astUnaryExpr{}
		logf(" [DEBUG] unary op = %s\n", tok)
		r.unaryExpr.Op = tok
		r.unaryExpr.X = x
		return r
	case "*":
		parserNext() // consume "*"
		var x = parseUnaryExpr()
		r = &astExpr{}
		r.dtype = "*astStarExpr"
		r.starExpr = &astStarExpr{}
		r.starExpr.X = x
		return r
	}
	r = parsePrimaryExpr()
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

func parseBinaryExpr(prec1 int) *astExpr {
	logf("   begin parseBinaryExpr() prec1=%s\n", Itoa(prec1))
	var x = parseUnaryExpr()
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
		parserExpect(op, __func__)
		var y = parseBinaryExpr(oprec + 1)
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

func parseExpr() *astExpr {
	logf("   begin parseExpr()\n")
	var e = parseBinaryExpr(1)
	logf("   end parseExpr()\n")
	return e
}

func parseRhs() *astExpr {
	var x = parseExpr()
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

func parseForStmt() *astStmt {
	logf(" begin %s\n", __func__)
	parserExpect("for", __func__)
	openScope()

	var s1 *astStmt
	var s2 *astStmt
	var s3 *astStmt
	var isRange bool
	parserExprLev = -1
	if p.tok.tok != "{" {
		if p.tok.tok != ";" {
			s2 = parseSimpleStmt(true)
			isRange = s2.isRange
			logf(" [%s] isRange=true\n", __func__)
		}
		if !isRange && p.tok.tok == ";" {
			parserNext() // consume ";"
			s1 = s2
			s2 = nil
			if p.tok.tok != ";" {
				s2 = parseSimpleStmt(false)
			}
			parserExpectSemi(__func__)
			if p.tok.tok != "{" {
				s3 = parseSimpleStmt(false)
			}
		}
	}

	parserExprLev = 0
	var body = parseBlockStmt()
	parserExpectSemi(__func__)

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
		var r = &astStmt{}
		r.dtype = "*astRangeStmt"
		r.rangeStmt = rangeStmt
		closeScope()
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
	closeScope()
	logf(" end %s\n", __func__)
	return r
}

func parseIfStmt() *astStmt {
	parserExpect("if", __func__)
	parserExprLev = -1
	var condStmt *astStmt = parseSimpleStmt(false)
	if condStmt.dtype != "*astExprStmt" {
		panic2(__func__, "unexpected dtype="+condStmt.dtype)
	}
	var cond = condStmt.exprStmt.X
	parserExprLev = 0
	var body = parseBlockStmt()
	var else_ *astStmt
	if p.tok.tok == "else" {
		parserNext()
		if p.tok.tok == "if" {
			else_ = parseIfStmt()
		} else {
			var elseblock = parseBlockStmt()
			parserExpectSemi(__func__)
			else_ = &astStmt{}
			else_.dtype = "*astBlockStmt"
			else_.blockStmt = elseblock
		}
	} else {
		parserExpectSemi(__func__)
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

func parseCaseClause() *astCaseClause {
	logf(" [%s] start\n", __func__)
	var list []*astExpr
	if p.tok.tok == "case" {
		parserNext() // consume "case"
		list = parseRhsList()
	} else {
		parserExpect("default", __func__)
	}

	parserExpect(":", __func__)
	openScope()
	var body = parseStmtList()
	var r = &astCaseClause{}
	r.Body = body
	r.List = list
	closeScope()
	logf(" [%s] end\n", __func__)
	return r
}

func parseSwitchStmt() *astStmt {
	parserExpect("switch", __func__)
	openScope()

	var s2 *astStmt
	parserExprLev = -1
	s2 = parseSimpleStmt(false)
	parserExprLev = 0

	parserExpect("{", __func__)
	var list []*astStmt
	var cc *astCaseClause
	var ccs *astStmt
	for p.tok.tok == "case" || p.tok.tok == "default" {
		cc = parseCaseClause()
		ccs = &astStmt{}
		ccs.dtype = "*astCaseClause"
		ccs.caseClause = cc
		list = append(list, ccs)
	}
	parserExpect("}", __func__)
	parserExpectSemi(__func__)
	var body = &astBlockStmt{}
	body.List = list

	var switchStmt = &astSwitchStmt{}
	switchStmt.Body = body
	switchStmt.Tag = makeExpr(s2)
	var s = &astStmt{}
	s.dtype = "*astSwitchStmt"
	s.switchStmt = switchStmt
	closeScope()
	return s
}

func parseLhsList() []*astExpr {
	logf(" [%s] start\n", __func__)
	var list = parseExprList()
	logf(" end %s\n", __func__)
	return list
}

func parseSimpleStmt(isRangeOK bool) *astStmt {
	logf(" begin %s\n", __func__)
	var s = &astStmt{}
	var x = parseLhsList()
	var stok = p.tok.tok
	var isRange = false
	var y *astExpr
	var rangeX *astExpr
	var rangeUnary *astUnaryExpr
	switch stok {
	case "=":
		parserNext() // consume =
		if isRangeOK && p.tok.tok == "range" {
			parserNext() // consume "range"
			rangeX = parseRhs()
			rangeUnary = &astUnaryExpr{}
			rangeUnary.Op = "range"
			rangeUnary.X = rangeX
			y = &astExpr{}
			y.dtype = "*astUnaryExpr"
			y.unaryExpr = rangeUnary
			isRange = true
		} else {
			y = parseExpr() // rhs
		}
		var as = &astAssignStmt{}
		as.Tok = "="
		as.Lhs = x
		as.Rhs = make([]*astExpr, 1, 1)
		as.Rhs[0] = y
		s.dtype = "*astAssignStmt"
		s.assignStmt = as
		s.isRange = isRange
		logf(" end %s\n", __func__)
		return s
	case ";":
		s.dtype = "*astExprStmt"
		var exprStmt = &astExprStmt{}
		exprStmt.X = x[0]
		s.exprStmt = exprStmt
		logf(" end %s\n", __func__)
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
		parserNext() // consume "++" or "--"
		return s
	}
	var exprStmt = &astExprStmt{}
	exprStmt.X = x[0]
	var r = &astStmt{}
	r.dtype = "*astExprStmt"
	r.exprStmt = exprStmt
	logf(" end %s\n", __func__)
	return r
}

func parseStmt() *astStmt {
	logf("\n")
	logf(" = begin %s\n", __func__)
	var s *astStmt
	switch p.tok.tok {
	case "var":
		var genDecl = parseDecl("var")
		s = &astStmt{}
		s.dtype = "*astDeclStmt"
		s.DeclStmt = &astDeclStmt{}
		var decl = &astDecl{}
		decl.dtype = "*astGenDecl"
		decl.genDecl = genDecl
		s.DeclStmt.Decl = decl
		logf(" = end parseStmt()\n")
	case "IDENT", "*":
		s = parseSimpleStmt(false)
		parserExpectSemi(__func__)
	case "return":
		s = parseReturnStmt()
	case "break", "continue":
		s = parseBranchStmt(p.tok.tok)
	case "if":
		s = parseIfStmt()
	case "switch":
		s = parseSwitchStmt()
	case "for":
		s = parseForStmt()
	default:
		panic2(__func__, "TBI 3:"+p.tok.tok)
	}
	logf(" = end parseStmt()\n")
	return s
}

func parseExprList() []*astExpr {
	logf(" [%s] start\n", __func__)
	var list []*astExpr
	var e = parseExpr()
	list = append(list, e)
	for p.tok.tok == "," {
		parserNext() // consume ","
		e = parseExpr()
		list = append(list, e)
	}

	logf(" [%s] end\n", __func__)
	return list
}

func parseRhsList() []*astExpr {
	var list = parseExprList()
	return list
}

func parseBranchStmt(tok string) *astStmt {
	parserExpect(tok, __func__)

	parserExpectSemi(__func__)

	var branchStmt = &astBranchStmt{}
	branchStmt.Tok = tok
	var s = &astStmt{}
	s.dtype = "*astBranchStmt"
	s.branchStmt = branchStmt
	return s
}

func parseReturnStmt() *astStmt {
	parserExpect("return", __func__)
	var x []*astExpr
	if p.tok.tok != ";" && p.tok.tok != "}" {
		x = parseRhsList()
	}
	parserExpectSemi(__func__)
	var returnStmt = &astReturnStmt{}
	returnStmt.Results = x
	var r = &astStmt{}
	r.dtype = "*astReturnStmt"
	r.returnStmt = returnStmt
	return r
}

func parseStmtList() []*astStmt {
	var list []*astStmt
	for p.tok.tok != "}" && p.tok.tok != "EOF" && p.tok.tok != "case" && p.tok.tok != "default" {
		var stmt = parseStmt()
		list = append(list, stmt)
	}
	return list
}

func parseBody(scope *astScope) *astBlockStmt {
	parserExpect("{", __func__)
	p.topScope = scope
	logf(" begin parseStmtList()\n")
	var list = parseStmtList()
	logf(" end parseStmtList()\n")

	closeScope()
	parserExpect("}", __func__)
	var r = &astBlockStmt{}
	r.List = list
	return r
}

func parseBlockStmt() *astBlockStmt {
	parserExpect("{", __func__)
	openScope()
	logf(" begin parseStmtList()\n")
	var list = parseStmtList()
	logf(" end parseStmtList()\n")
	closeScope()
	parserExpect("}", __func__)
	var r = &astBlockStmt{}
	r.List = list
	return r
}

func parseDecl(keyword string) *astGenDecl {
	var r *astGenDecl
	switch p.tok.tok {
	case "var":
		parserExpect(keyword, __func__)
		var ident = parseIdent()
		var typ = parseType()
		var value *astExpr
		if p.tok.tok == "=" {
			parserNext()
			value = parseExpr()
		}
		parserExpectSemi(__func__)
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

func parserTypeSpec() *astSpec {
	logf(" [%s] start\n", __func__)
	parserExpect("type", __func__)
	var ident = parseIdent()
	logf(" decl type %s\n", ident.Name)

	var spec = &astTypeSpec{}
	spec.Name = ident
	var objDecl = &ObjDecl{}
	objDecl.dtype = "*astTypeSpec"
	objDecl.typeSpec = spec
	declare(objDecl, p.topScope, astTyp, ident)
	var typ = parseType()
	parserExpectSemi(__func__)
	spec.Type = typ
	var r = &astSpec{}
	r.dtype = "*astTypeSpec"
	r.typeSpec = spec
	return r
}

func parserValueSpec(keyword string) *astSpec {
	logf(" [parserValueSpec] start\n")
	parserExpect(keyword, __func__)
	var ident = parseIdent()
	logf(" var = %s\n", ident.Name)
	var typ = parseType()
	var value *astExpr
	if p.tok.tok == "=" {
		parserNext()
		value = parseExpr()
	}
	parserExpectSemi(__func__)
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

func parserFuncDecl() *astDecl {
	parserExpect("func", __func__)
	var scope = astNewScope(p.topScope) // function scope

	var ident = parseIdent()
	var sig = parseSignature(scope)
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
		body = parseBody(scope)
		logf(" end parseBody()\n")
		parserExpectSemi(__func__)
	} else {
		parserExpectSemi(__func__)
	}
	var decl = &astDecl{}
	decl.dtype = "*astFuncDecl"
	var funcDecl = &astFuncDecl{}
	decl.funcDecl = funcDecl
	decl.funcDecl.Name = ident
	decl.funcDecl.Type = &astFuncType{}
	decl.funcDecl.Type.params = params
	decl.funcDecl.Type.results = results
	decl.funcDecl.Body = body
	var objDecl = &ObjDecl{}
	objDecl.dtype = "*astFuncDecl"
	objDecl.funcDecl = funcDecl
	declare(objDecl, p.pkgScope, astFun, ident)
	return decl
}

func parserFile() *astFile {
	// expect "package" keyword
	parserExpect("package", __func__)
	p.unresolved = nil
	var ident = parseIdent()
	var packageName = ident.Name
	parserExpectSemi(__func__)

	p.topScope = &astScope{} // open scope
	p.pkgScope = p.topScope

	for p.tok.tok == "import" {
		parserImportDecl()
	}

	logf("\n")
	logf(" [parser] Parsing Top level decls\n")
	var decls []*astDecl
	var decl *astDecl

	for p.tok.tok != "EOF" {
		switch p.tok.tok {
		case "var", "const":
			var spec = parserValueSpec(p.tok.tok)
			var genDecl = &astGenDecl{}
			genDecl.Spec = spec
			decl = &astDecl{}
			decl.dtype = "*astGenDecl"
			decl.genDecl = genDecl
		case "func":
			logf("\n\n")
			decl = parserFuncDecl()
			logf(" func decl parsed:%s\n", decl.funcDecl.Name.Name)
		case "type":
			var spec = parserTypeSpec()
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
	var oe *objectEntry
	for _, oe = range p.pkgScope.Objects {
		logf("    object %s\n", oe.name)
	}

	var unresolved []*astIdent
	var idnt *astIdent
	logf(" [parserFile] resolving parser's unresolved (n=%s)\n", Itoa(len(p.unresolved)))
	for _, idnt = range p.unresolved {
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

	p = &parser{}
	p.scanner = &scanner{}
	parserInit(text)
	return parserFile()
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
		fmtPrintf("  movq %d(%%rax), %%rdx\n", Itoa(8))
		fmtPrintf("  movq %d(%%rax), %%rax\n", Itoa(0))
		fmtPrintf("  pushq %%rdx # len\n")
		fmtPrintf("  pushq %%rax # ptr\n")
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
		fmtPrintf("  leaq %s(%%rip), %%rax # global variable addr\n", variable.globalSymbol)
	} else {
		fmtPrintf("  leaq %d(%%rbp), %%rax # local variable addr\n", Itoa(variable.localOffset))
	}

	fmtPrintf("  pushq %%rax\n")
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

func emitAddr(expr *astExpr) {
	emitComment(2, "[emitAddr] %s\n", expr.dtype)
	switch expr.dtype {
	case "*astIdent":
		if expr.ident.Obj.Kind == astVar {
			assert(expr.ident.Obj.Variable != nil,
				"ERROR: Variable is nil for name : "+expr.ident.Obj.Name, __func__)
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
	case "*astSelectorExpr": // X.Sel
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
		emitComment(0, "[isType][DEBUG] expr.ident.Name = %s\n", expr.ident.Name)
		emitComment(0, "[isType][DEBUG] expr.ident.Obj = %s,%s\n",
			expr.ident.Obj.Name, expr.ident.Obj.Kind)
		return expr.ident.Obj.Kind == astTyp
	case "*astParenExpr":
		return isType(expr.parenExpr.X)
	case "*astStarExpr":
		return isType(expr.starExpr.X)
	default:
		emitComment(0, "[isType][%s] is not considered a type\n", expr.dtype)
	}

	return false

}

func emitConversion(tp *Type, arg0 *astExpr) {
	emitComment(0, "[emitConversion]\n")
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
			emitComment(0, "[emitConversion] to int \n")
			emitExpr(arg0, nil)
		default:
			panic2(__func__, "[*astIdent] TBI : "+typeExpr.ident.Obj.Name)
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
		emitComment(0, "[emitConversion] to pointer \n")
		emitExpr(arg0, nil)
	default:
		panic2(__func__, "TBI :"+typeExpr.dtype)
	}
}

func emitZeroValue(t *Type) {
	switch kind(t) {
	case T_SLICE:
		fmtPrintf("  pushq $0 # slice zero value\n")
		fmtPrintf("  pushq $0 # slice zero value\n")
		fmtPrintf("  pushq $0 # slice zero valuer\n")
	case T_STRING:
		fmtPrintf("  pushq $0 # string zero value\n")
		fmtPrintf("  pushq $0 # string zero value\n")
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
	emitComment(0, "[%s] begin\n", __func__)
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
	emitComment(0, "[%s] end\n", __func__)
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
	fmtPrintf("  callq runtime.malloc\n") // no need to invert args orders
	emitRevertStackPointer(intSize)
	fmtPrintf("  pushq %%rax # addr\n")
}

func emitStructLiteral(e *astCompositeLit) {
	// allocate heap area with zero value
	fmtPrintf("  # Struct literal\n")
	var structType = e2t(e.Type)
	emitZeroValue(structType) // push address of the new storage
	var kvExpr *astKeyValueExpr
	var i int
	var elm *astExpr
	for i, elm = range e.Elts {
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
		emitExpr(kvExpr.Value, fieldType)
		// assign
		emitStore(fieldType)
	}
}

func emitArrayLiteral(arrayType *astArrayType, arrayLen int, elts []*astExpr) {
	var elmType = e2t(arrayType.Elt)
	var elmSize = getSizeOfType(elmType)
	var memSize = elmSize * arrayLen
	emitCallMalloc(memSize) // push
	var i int
	var elm *astExpr
	for i, elm = range elts {
		// emit lhs
		emitPushStackTop(tUintptr, "malloced address")
		emitAddConst(elmSize*i, "malloced address + elmSize * index ("+Itoa(i)+")")
		emitExpr(elm, elmType)
		emitStore(elmType)
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
	var arg *Arg
	for _, arg = range args {
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
	for _, arg = range args {
		emitExpr(arg.e, arg.t)
	}
	fmtPrintf("  addq $%d, %%rsp # for args\n", Itoa(totalPushedSize))

	for _, arg = range args {
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
		case T_STRING:
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

func emitCallNonDecl(symbol string, eArgs []*astExpr) {
	var args []*Arg
	var eArg *astExpr
	for _, eArg = range eArgs {
		var arg = &Arg{}
		arg.e = eArg
		arg.t = nil
		args = append(args, arg)
	}
	emitCall(symbol, args)
}

func emitCall(symbol string, args []*Arg) {
	emitComment(0, "[%s] %s\n", __func__, symbol)
	var totalPushedSize = emitArgs(args)
	fmtPrintf("  callq %s\n", symbol)
	emitRevertStackPointer(totalPushedSize)
}

func emitFuncall(fun *astExpr, eArgs []*astExpr) {
	switch fun.dtype {
	case "*astIdent":
		emitComment(0, "[%s][*astIdent]\n", __func__)
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
				var args []*Arg
				var arg0 *Arg // elmSize
				var arg1 *Arg
				var arg2 *Arg
				arg0 = &Arg{}
				arg0.e = eNumLit
				arg0.t = tInt
				arg1 = &Arg{} // len
				arg1.e = eArgs[1]
				arg1.t = tInt
				arg2 = &Arg{} // cap
				arg2.e = eArgs[2]
				arg2.t = tInt
				args = append(args, arg0)
				args = append(args, arg1)
				args = append(args, arg2)

				emitCall("runtime.makeSlice", args)
				fmtPrintf("  pushq %%rsi # slice cap\n")
				fmtPrintf("  pushq %%rdi # slice len\n")
				fmtPrintf("  pushq %%rax # slice ptr\n")
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

			var args []*Arg
			var arg0 = &Arg{} // slice
			arg0.e = sliceArg
			args = append(args, arg0)

			var arg1 = &Arg{} // element
			arg1.e = elemArg
			arg1.t = elmType
			args = append(args, arg1)

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
			emitCall(symbol, args)
			fmtPrintf("  pushq %%rsi # slice cap\n")
			fmtPrintf("  pushq %%rdi # slice len\n")
			fmtPrintf("  pushq %%rax # slice ptr\n")
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
		var symbol = pkg.name + "." + fn.Name
		emitComment(0, "[%s][*astIdent][default] start\n", __func__)

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

		var params = fndecl.Type.params.List
		var variadicArgs []*astExpr
		var variadicElp *astEllipsis
		var args []*Arg
		var eArg *astExpr
		var param *astField
		var argIndex int
		var arg *Arg
		var lenParams = len(params)
		for argIndex, eArg = range eArgs {
			emitComment(0, "[%s][*astIdent][default] loop idx %s, len params %s\n", __func__, Itoa(argIndex), Itoa(lenParams))
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
			emitComment(0, "len(args)=%s, len(params)=%s\n", Itoa(len(args)), Itoa(len(params)))
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

		emitCall(symbol, args)

		// push results
		var results = fndecl.Type.results
		if fndecl.Type.results == nil {
			emitComment(0, "[emitExpr] %s sig.results is nil\n", fn.Name)
		} else {
			emitComment(0, "[emitExpr] %s sig.results.List = %s\n", fn.Name, Itoa(len(fndecl.Type.results.List)))
		}

		if results != nil && len(results.List) == 1 {
			var retval0 = fndecl.Type.results.List[0]
			var knd = kind(e2t(retval0.Type))
			switch knd {
			case T_STRING:
				emitComment(2, "fn.Obj=%s\n", obj.Name)
				fmtPrintf("  pushq %%rdi # str len\n")
				fmtPrintf("  pushq %%rax # str ptr\n")
			case T_BOOL, T_INT, T_UINTPTR, T_POINTER:
				emitComment(2, "fn.Obj=%s\n", obj.Name)
				fmtPrintf("  pushq %%rax\n")
			case T_SLICE:
				fmtPrintf("  pushq %%rsi # slice cap\n")
				fmtPrintf("  pushq %%rdi # slice len\n")
				fmtPrintf("  pushq %%rax # slice ptr\n")
			default:
				panic2(__func__, "Unexpected kind="+knd)
			}
		} else {
			emitComment(2, "No results\n")
		}
		return
	case "*astSelectorExpr":
		var selectorExpr = fun.selectorExpr
		if selectorExpr.X.dtype != "*astIdent" {
			panic2(__func__, "TBI selectorExpr.X.dtype="+selectorExpr.X.dtype)
		}
		var symbol string = selectorExpr.X.ident.Name + "." + selectorExpr.Sel.Name
		switch symbol {
		case "os.Exit":
			emitCallNonDecl(symbol, eArgs)
		case "syscall.Write":
			emitCallNonDecl(symbol, eArgs)
		case "syscall.Open":
			// func decl is in runtime
			emitCallNonDecl(symbol, eArgs)
			fmtPrintf("  pushq %%rax # fd\n")
		case "syscall.Read":
			// func decl is in runtime
			emitCallNonDecl(symbol, eArgs)
			fmtPrintf("  pushq %%rax # fd\n")
		case "syscall.Syscall":
			emitCallNonDecl(symbol, eArgs)
			fmtPrintf("  pushq %%rax # ret\n")
		case "unsafe.Pointer":
			emitExpr(eArgs[0], nil)
		default:
			fmtPrintf("  callq %s.%s\n", selectorExpr.X.ident.Name, selectorExpr.Sel.Name)
			panic2(__func__, "[*astSelectorExpr] Unsupported call to "+symbol)
		}
	case "*astParenExpr":
		panic2(__func__, "[astParenExpr] TBI ")
	default:
		panic2(__func__, "TBI fun.dtype="+fun.dtype)
	}
}

func emitExpr(e *astExpr, forceType *Type) {
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
			return
		case gFalse:
			emitFalse()
			return
		case gNil:
			if forceType == nil {
				panic2(__func__, "Type is required to emit nil")
			}
			switch kind(forceType) {
			case T_SLICE, T_POINTER:
				emitZeroValue(forceType)
			default:
				panic2(__func__, "Unexpected kind="+kind(forceType))
			}
			return
		}
		switch ident.Obj.Kind {
		case astVar:
			emitAddr(e)
			var t = getTypeOfExpr(e)
			emitLoad(t)
		case astCon:
			var valSpec = ident.Obj.Decl.valueSpec
			assert(valSpec != nil, "valSpec should not be nil", __func__)
			assert(valSpec.Value != nil, "valSpec should not be nil", __func__)
			assert(valSpec.Value.dtype == "*astBasicLit", "const value should be a literal", __func__)
			var t *Type
			if valSpec.Type != nil {
				t = e2t(valSpec.Type)
			} else {
				t = forceType
			}
			emitExpr(valSpec.Value, t)
		case astTyp:
			panic2(__func__, "[*astIdent] Kind Typ should not come here")
		default:
			panic2(__func__, "[*astIdent] unknown Kind="+ident.Obj.Kind+" Name="+ident.Obj.Name)
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
	case "*astCallExpr":
		var fun = e.callExpr.Fun
		emitComment(0, "[%s][*astCallExpr]\n", __func__)
		if isType(fun) {
			emitConversion(e2t(fun), e.callExpr.Args[0])
			return
		}
		emitFuncall(fun, e.callExpr.Args)
	case "*astParenExpr":
		emitExpr(e.parenExpr.X, nil)
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
		emitComment(0, "[DEBUG] unary op = %s\n", e.unaryExpr.Op)
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
				emitCall("runtime.catstrings", args)
				fmtPrintf("  pushq %%rdi # slice len\n")
				fmtPrintf("  pushq %%rax # slice ptr\n")
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
			return
		}

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
			return
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
			return
		}

		var t = getTypeOfExpr(e.binaryExpr.X)
		emitExpr(e.binaryExpr.X, nil) // left
		emitExpr(e.binaryExpr.Y, t)   // right
		switch e.binaryExpr.Op {
		case "+":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  addq %%rcx, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		case "-":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  subq %%rcx, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		case "*":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  imulq %%rcx, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		case "%":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  movq $0, %%rdx # init %%rdx\n")
			fmtPrintf("  divq %%rcx\n")
			fmtPrintf("  movq %%rdx, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		case "/":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  movq $0, %%rdx # init %%rdx\n")
			fmtPrintf("  divq %%rcx\n")
			fmtPrintf("  pushq %%rax\n")
		case "==":
			emitCompEq(t)
		case "!=":
			emitCompEq(t)
			emitInvertBoolValue()
		case "<":
			emitCompExpr("setl")
		case "<=":
			emitCompExpr("setle")
		case ">":
			emitCompExpr("setg")
		case ">=":
			emitCompExpr("setge")
		default:
			panic2(__func__, "# TBI: binary operation for "+e.binaryExpr.Op)
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
	default:
		panic2(__func__, "[emitExpr] `TBI:"+e.dtype)
	}
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
		fmtPrintf("  callq runtime.cmpstrings\n")
		emitRevertStackPointer(stringSize * 2)
		fmtPrintf("  pushq %%rax # cmp result (1 or 0)\n")
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

func emitStore(t *Type) {
	emitComment(2, "emitStore(%s)\n", kind(t))
	switch kind(t) {
	case T_SLICE:
		emitPopSlice()
		fmtPrintf("  popq %%rsi # lhs ptr addr\n")
		fmtPrintf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", Itoa(0))
		fmtPrintf("  movq %%rcx, %d(%%rsi) # len to len\n", Itoa(8))
		fmtPrintf("  movq %%rdx, %d(%%rsi) # cap to cap\n", Itoa(16))
	case T_STRING:
		emitPopString()
		fmtPrintf("  popq %%rsi # lhs ptr addr\n")
		fmtPrintf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", Itoa(0))
		fmtPrintf("  movq %%rcx, %d(%%rsi) # len to len\n", Itoa(8))
	case T_INT, T_BOOL, T_UINTPTR, T_POINTER:
		fmtPrintf("  popq %%rdi # rhs evaluated\n")
		fmtPrintf("  popq %%rax # lhs addr\n")
		fmtPrintf("  movq %%rdi, (%%rax) # assign\n")
	case T_UINT8:
		fmtPrintf("  popq %%rdi # rhs evaluated\n")
		fmtPrintf("  popq %%rax # lhs addr\n")
		fmtPrintf("  movb %%dil, (%%rax) # assign byte\n")
	case T_UINT16:
		fmtPrintf("  popq %%rdi # rhs evaluated\n")
		fmtPrintf("  popq %%rax # lhs addr\n")
		fmtPrintf("  movw %%di, (%%rax) # assign word\n")
	case T_STRUCT, T_ARRAY:
		fmtPrintf("  popq %%rdi # rhs: addr of data\n")
		fmtPrintf("  popq %%rax # lhs: addr to store\n")
		fmtPrintf("  pushq $%d # size\n", Itoa(getSizeOfType(t)))
		fmtPrintf("  pushq %%rax # dst lhs\n")
		fmtPrintf("  pushq %%rdi # src rhs\n")
		fmtPrintf("  callq runtime.memcopy\n")
		emitRevertStackPointer(ptrSize*2 + intSize)
	default:
		panic2(__func__, "TBI:"+kind(t))
	}
}

func emitAssign(lhs *astExpr, rhs *astExpr) {
	emitComment(2, "Assignment: emitAddr(lhs:%s)\n", lhs.dtype)
	emitAddr(lhs)
	emitComment(2, "Assignment: emitExpr(rhs)\n")
	emitExpr(rhs, getTypeOfExpr(lhs))
	emitStore(getTypeOfExpr(lhs))
}

func emitStmt(stmt *astStmt) {
	emitComment(2, "\n")
	emitComment(2, "== Stmt %s ==\n", stmt.dtype)
	switch stmt.dtype {
	case "*astBlockStmt":
		var stmt2 *astStmt
		for _, stmt2 = range stmt.blockStmt.List {
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
			emitStore(t)
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
			var lhs = stmt.assignStmt.Lhs
			var rhs = stmt.assignStmt.Rhs
			emitAssign(lhs[0], rhs[0])
		default:
			panic2(__func__, "TBI: assignment of "+stmt.assignStmt.Tok)
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
		emitStore(tInt)

		emitComment(2, "  assign 0 to indexvar\n")
		// indexvar = 0
		emitVariableAddr(stmt.rangeStmt.indexvar)
		emitZeroValue(tInt)
		emitStore(tInt)

		// init key variable with 0
		if stmt.rangeStmt.Key != nil {
			assert(stmt.rangeStmt.Key.dtype == "*astIdent", "key expr should be an ident", __func__)
			var keyIdent = stmt.rangeStmt.Key.ident
			if keyIdent.Name != "_" {
				emitAddr(stmt.rangeStmt.Key) // lhs
				emitZeroValue(tInt)
				emitStore(tInt)
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
		emitStore(elemType)

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
		emitStore(tInt)

		if stmt.rangeStmt.Key != nil {
			assert(stmt.rangeStmt.Key.dtype == "*astIdent", "key expr should be an ident", __func__)
			var keyIdent = stmt.rangeStmt.Key.ident
			if keyIdent.Name != "_" {
				emitAddr(stmt.rangeStmt.Key)              // lhs
				emitVariableAddr(stmt.rangeStmt.indexvar) // rhs
				emitLoad(tInt)
				emitStore(tInt)
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
		emitStore(getTypeOfExpr(stmt.incDecStmt.X))
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
		var i int
		var c *astStmt
		emitComment(2, "Start comparison with cases\n")
		for i, c = range cases {
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
			var e *astExpr
			for _, e = range cc.List {
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
		for i, c = range cases {
			assert(c.dtype == "*astCaseClause", "should be *astCaseClause", __func__)
			var cc = c.caseClause
			fmtPrintf("%s:\n", labels[i])
			var _s *astStmt
			for _, _s = range cc.Body {
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

func emitFuncDecl(pkgPrefix string, fnc *Func) {
	var localarea = fnc.localarea
	fmtPrintf("\n")
	fmtPrintf("%s.%s: # args %d, locals %d\n",
		pkgPrefix, fnc.name, Itoa(fnc.argsarea), Itoa(fnc.localarea))

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

func emitGlobalVariable(name *astIdent, t *Type, val *astExpr) {
	var typeKind = kind(t)
	fmtPrintf("%s: # T %s\n", name.Name, typeKind)
	switch typeKind {
	case T_STRING:
		if val != nil && val.dtype == "*astBasicLit" {
			var sl = getStringLiteral(val.basicLit)
			fmtPrintf("  .quad %s\n", sl.label)
			fmtPrintf("  .quad %d\n", Itoa(sl.strlen))
		} else {
			fmtPrintf("  .quad 0\n")
			fmtPrintf("  .quad 0\n")
		}
	case T_POINTER:
		fmtPrintf("  .quad 0 # pointer \n") // @TODO
	case T_UINTPTR:
		fmtPrintf("  .quad 0\n")
	case T_BOOL:
		if val != nil {
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
		} else {
			fmtPrintf("  .quad 0 # bool zero value\n")
		}
	case T_INT:
		fmtPrintf("  .quad 0\n")
	case T_UINT8:
		fmtPrintf("  .byte 0\n")
	case T_UINT16:
		fmtPrintf("  .word 0\n")
	case T_SLICE:
		fmtPrintf("  .quad 0 # ptr\n")
		fmtPrintf("  .quad 0 # len\n")
		fmtPrintf("  .quad 0 # cap\n")
	case T_ARRAY:
		if val != nil {
			panic2(__func__, "TBI")
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

func emitData(pkgName string, vars []*astValueSpec, sliterals []*stringLiteralsContainer) {
	fmtPrintf(".data\n")
	emitComment(0, "string literals len = %s\n", Itoa(len(sliterals)))
	var con *stringLiteralsContainer
	for _, con = range sliterals {
		emitComment(0, "string literals\n")
		fmtPrintf("%s:\n", con.sl.label)
		fmtPrintf("  .string %s\n", con.sl.value)
	}

	emitComment(0, "===== Global Variables =====\n")

	var spec *astValueSpec
	for _, spec = range vars {
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariable(spec.Name, t, spec.Value)
	}

	emitComment(0, "==============================\n")
}

func emitText(pkgName string, funcs []*Func) {
	fmtPrintf(".text\n")
	var fnc *Func
	for _, fnc = range funcs {
		emitFuncDecl(pkgName, fnc)
	}
}

func generateCode(pkgContainer *PkgContainer) {
	emitData(pkgContainer.name, pkgContainer.vars, stringLiterals)
	emitText(pkgContainer.name, pkgContainer.funcs)
}

// --- type ---
const sliceSize int = 24
const stringSize int = 16
const intSize int = 8
const ptrSize int = 8

type Type struct {
	//kind string
	e *astExpr
}

const T_STRING string = "T_STRING"
const T_SLICE string = "T_SLICE"
const T_BOOL string = "T_BOOL"
const T_INT string = "T_INT"
const T_UINT8 string = "T_UINT8"
const T_UINT16 string = "T_UINT16"
const T_UINTPTR string = "T_UINTPTR"
const T_ARRAY string = "T_ARRAY"
const T_STRUCT string = "T_STRUCT"
const T_POINTER string = "T_POINTER"

var tInt *Type
var tUint8 *Type
var tUint16 *Type
var tUintptr *Type
var tString *Type
var tBool *Type

func getTypeOfExpr(expr *astExpr) *Type {
	//emitComment(0, "[%s] start\n", __func__)
	switch expr.dtype {
	case "*astIdent":
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
			default:
				panic2(__func__, "ERROR 0\n")
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
		default:
			panic2(__func__, "TBI: Op="+expr.unaryExpr.Op)
		}
	case "*astCallExpr":
		emitComment(0, "[%s] *astCallExpr\n", __func__)
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
				}
				var decl = fn.Obj.Decl
				if decl == nil {
					panic2(__func__, "decl of function "+fn.Name+" is  nil")
				}
				switch decl.dtype {
				case "*astFuncDecl":
					var resultList = decl.funcDecl.Type.results.List
					if len(resultList) != 1 {
						panic2(__func__, "[astCallExpr] len results.List is not 1")
					}
					return e2t(decl.funcDecl.Type.results.List[0].Type)
				default:
					panic2(__func__, "[astCallExpr] decl.dtype="+decl.dtype)
				}
				panic2(__func__, "[astCallExpr] Fun ident "+fn.Name)
			}
		case "*astArrayType":
			return e2t(fun)
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
		var structType = getStructTypeOfX(expr.selectorExpr)
		var field = lookupStructField(getStructTypeSpec(structType), expr.selectorExpr.Sel.Name)
		return e2t(field.Type)
	case "*astCompositeLit":
		return e2t(expr.compositeLit.Type)
	case "*astParenExpr":
		return getTypeOfExpr(expr.parenExpr.X)
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

func kind(t *Type) string {
	if t == nil {
		panic2(__func__, "nil type is not expected\n")
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
	default:
		panic2(__func__, "Unkown dtype:"+t.e.dtype)
	}
	panic2(__func__, "error")
	return ""
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
	for _, field = range getStructFields(structTypeSpec) {
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
	var field *astField
	for _, field = range fields {
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
	localvars []*string
	localarea int
	argsarea  int
	name      string
	Body      *astBlockStmt
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
	var container *stringLiteralsContainer
	for _, container = range stringLiterals {
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
	var c uint8
	var vl = []uint8(lit.Value)
	for _, c = range vl {
		if c != '\\' {
			strlen++
		}
	}

	var label = fmtSprintf(".%s.S%d", []string{pkg.name, Itoa(stringIndex)})
	stringIndex++

	var sl = &sliteral{}
	sl.label = label
	sl.strlen = strlen - 2
	sl.value = lit.Value
	logf(" [registerStringLiteral] label=%s, strlen=%s\n", sl.label, Itoa(sl.strlen))
	var cont = &stringLiteralsContainer{}
	cont.sl = sl
	cont.lit = lit
	stringLiterals = append(stringLiterals, cont)
}

func newGlobalVariable(name string) *Variable {
	var vr = &Variable{}
	vr.name = name
	vr.isGlobal = true
	vr.globalSymbol = name
	return vr
}

func newLocalVariable(name string, localoffset int) *Variable {
	var vr = &Variable{}
	vr.name = name
	vr.isGlobal = false
	vr.localOffset = localoffset
	return vr
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
		var rhs = stmt.assignStmt.Rhs
		var rhsE *astExpr
		for _, rhsE = range rhs {
			walkExpr(rhsE)
		}
	case "*astExprStmt":
		walkExpr(stmt.exprStmt.X)
	case "*astReturnStmt":
		var rt *astExpr
		for _, rt = range stmt.returnStmt.Results {
			walkExpr(rt)
		}
	case "*astIfStmt":
		if stmt.ifStmt.Init != nil {
			walkStmt(stmt.ifStmt.Init)
		}
		walkExpr(stmt.ifStmt.Cond)
		var s *astStmt
		for _, s = range stmt.ifStmt.Body.List {
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
		stmt.rangeStmt.lenvar = lenvar
		stmt.rangeStmt.indexvar = indexvar
		currentFor = stmt.rangeStmt.Outer
	case "*astIncDecStmt":
		walkExpr(stmt.incDecStmt.X)
	case "*astBlockStmt":
		var s *astStmt
		for _, s = range stmt.blockStmt.List {
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
		var e_ *astExpr
		var s_ *astStmt
		for _, e_ = range stmt.caseClause.List {
			walkExpr(e_)
		}
		for _, s_ = range stmt.caseClause.Body {
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
		var arg *astExpr
		walkExpr(expr.callExpr.Fun)
		// Replace __func__ ident by a string literal
		var basicLit *astBasicLit
		var i int
		var newArg *astExpr
		for i, arg = range expr.callExpr.Args {
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
		var v *astExpr
		for _, v = range expr.compositeLit.Elts {
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
	default:
		panic2(__func__, "TBI:"+expr.dtype)
	}
}

func walk(pkgContainer *PkgContainer, file *astFile) {
	var decl *astDecl
	for _, decl = range file.Decls {
		switch decl.dtype {
		case "*astGenDecl":
			var genDecl = decl.genDecl
			switch genDecl.Spec.dtype {
			case "*astValueSpec":
				var valSpec = genDecl.Spec.valueSpec
				var nameIdent = valSpec.Name
				if nameIdent.Obj.Kind == astVar {
					nameIdent.Obj.Variable = newGlobalVariable(nameIdent.Obj.Name)
					pkgContainer.vars = append(pkgContainer.vars, valSpec)
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
			var field *astField
			for _, field = range funcDecl.Type.params.List {
				var obj = field.Name.Obj
				obj.Variable = newLocalVariable(obj.Name, paramoffset)
				var varSize = getSizeOfType(e2t(field.Type))
				paramoffset = paramoffset + varSize
				logf(" field.Name.Obj.Name=%s\n", obj.Name)
				//logf("   field.Type=%#v\n", field.Type)
			}
			if funcDecl.Body != nil {
				var stmt *astStmt
				for _, stmt = range funcDecl.Body.List {
					walkStmt(stmt)
				}
				var fnc = &Func{}
				fnc.name = funcDecl.Name.Name
				fnc.Body = funcDecl.Body
				fnc.localarea = localoffset
				fnc.argsarea = paramoffset

				pkgContainer.funcs = append(pkgContainer.funcs, fnc)
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
var gNil *astObject
var identNil *astIdent
var eNil *astExpr
var gTrue *astObject
var gFalse *astObject
var gString *astObject
var gInt *astObject
var gUint8 *astObject
var gUint16 *astObject
var gUintptr *astObject
var gBool *astObject
var gNew *astObject
var gMake *astObject
var gAppend *astObject
var gLen *astObject
var gCap *astObject

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

	logf(" [%s] scope insertion of predefined identifiers complete\n", __func__)

	// @FIXME package names should not be be in universe
	var pkgOs = &astObject{}
	pkgOs.Kind = "Pkg"
	pkgOs.Name = "os"
	scopeInsert(universe, pkgOs)

	var pkgSyscall = &astObject{}
	pkgSyscall.Kind = "Pkg"
	pkgSyscall.Name = "syscall"
	scopeInsert(universe, pkgSyscall)

	var pkgUnsafe = &astObject{}
	pkgUnsafe.Kind = "Pkg"
	pkgUnsafe.Name = "unsafe"
	scopeInsert(universe, pkgUnsafe)
	logf(" [%s] scope insertion complete\n", __func__)
	return universe
}

func resolveUniverse(file *astFile, universe *astScope) {
	logf(" [%s] start\n", __func__)
	// create universe scope
	// inject predeclared identifers
	var unresolved []*astIdent
	var ident *astIdent
	logf(" [SEMA] resolving file.Unresolved (n=%s)\n", Itoa(len(file.Unresolved)))
	for _, ident = range file.Unresolved {
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

func initGlobals() {
	gNil = &astObject{}
	gNil.Kind = astCon // is it Con ?
	gNil.Name = "nil"

	identNil = &astIdent{}
	identNil.Obj = gNil
	identNil.Name = "nil"
	eNil = &astExpr{}
	eNil.dtype = "*astIdent"
	eNil.ident = identNil

	gTrue = &astObject{}
	gTrue.Kind = astCon
	gTrue.Name = "true"

	gFalse = &astObject{}
	gFalse.Kind = astCon
	gFalse.Name = "false"

	gString = &astObject{}
	gString.Kind = astTyp
	gString.Name = "string"
	tString = &Type{}
	tString.e = &astExpr{}
	tString.e.dtype = "*astIdent"
	tString.e.ident = &astIdent{}
	tString.e.ident.Name = "string"
	tString.e.ident.Obj = gString

	gInt = &astObject{}
	gInt.Kind = astTyp
	gInt.Name = "int"
	tInt = &Type{}
	tInt.e = &astExpr{}
	tInt.e.dtype = "*astIdent"
	tInt.e.ident = &astIdent{}
	tInt.e.ident.Name = "int"
	tInt.e.ident.Obj = gInt

	gUint8 = &astObject{}
	gUint8.Kind = astTyp
	gUint8.Name = "uint8"
	tUint8 = &Type{}
	tUint8.e = &astExpr{}
	tUint8.e.dtype = "*astIdent"
	tUint8.e.ident = &astIdent{}
	tUint8.e.ident.Name = "uint8"
	tUint8.e.ident.Obj = gUint8

	gUint16 = &astObject{}
	gUint16.Kind = astTyp
	gUint16.Name = "uint16"
	tUint16 = &Type{}
	tUint16.e = &astExpr{}
	tUint16.e.dtype = "*astIdent"
	tUint16.e.ident = &astIdent{}
	tUint16.e.ident.Name = "uint16"
	tUint16.e.ident.Obj = gUint16

	gUintptr = &astObject{}
	gUintptr.Kind = astTyp
	gUintptr.Name = "uintptr"
	tUintptr = &Type{}
	tUintptr.e = &astExpr{}
	tUintptr.e.dtype = "*astIdent"
	tUintptr.e.ident = &astIdent{}
	tUintptr.e.ident.Name = "uintptr"
	tUintptr.e.ident.Obj = gUintptr

	gBool = &astObject{}
	gBool.Kind = astTyp
	gBool.Name = "bool"
	tBool = &Type{}
	tBool.e = &astExpr{}
	tBool.e.dtype = "*astIdent"
	tBool.e.ident = &astIdent{}
	tBool.e.ident.Name = "bool"
	tBool.e.ident.Obj = gBool

	gNew = &astObject{}
	gNew.Kind = astFun
	gNew.Name = "new"

	gMake = &astObject{}
	gMake.Kind = astFun
	gMake.Name = "make"

	gAppend = &astObject{}
	gAppend.Kind = astFun
	gAppend.Name = "append"

	gLen = &astObject{}
	gLen.Kind = astFun
	gLen.Name = "len"

	gCap = &astObject{}
	gCap.Kind = astFun
	gCap.Name = "cap"
}

var pkg *PkgContainer

type PkgContainer struct {
	name string
	vars []*astValueSpec
	funcs []*Func
}

func main() {
	initGlobals()
	var universe = createUniverse()

	var sourceFiles = []string{"runtime.go", "/dev/stdin"}

	var sourceFile string
	for _, sourceFile = range sourceFiles {
		fmtPrintf("# file: %s\n", sourceFile)
		stringIndex = 0
		stringLiterals = nil
		var f = parseFile(sourceFile)
		resolveUniverse(f, universe)
		pkg = &PkgContainer{
			name: f.Name,
		}
		walk(pkg, f)
		generateCode(pkg)
	}
}
