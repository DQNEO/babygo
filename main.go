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
var scannerSrc []uint8
var scannerCh uint8
var scannerOffset int
var scannerNextOffset int
var scannerInsertSemi bool

func scannerNext() {
	if scannerNextOffset < len(scannerSrc) {
		scannerOffset = scannerNextOffset
		scannerCh = scannerSrc[scannerOffset]
		scannerNextOffset++
	} else {
		scannerOffset = len(scannerSrc)
		scannerCh = 1 //EOF
	}
}

var keywords []string

func scannerInit(src []uint8) {
	// https://golang.org/ref/spec#Keywords
	keywords = []string{
		"break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var",
	}
	scannerSrc = src
	scannerOffset = 0
	scannerNextOffset = 0
	scannerInsertSemi = false
	scannerCh = ' '
	logf("src len = %s\n", Itoa(len(scannerSrc)))
	scannerNext()
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

func scannerScanIdentifier() string {
	var offset = scannerOffset
	for isLetter(scannerCh) || isDecimal(scannerCh) {
		scannerNext()
	}
	return string(scannerSrc[offset:scannerOffset])
}

func scannerScanNumber() string {
	var offset = scannerOffset
	for isDecimal(scannerCh) {
		scannerNext()
	}
	return string(scannerSrc[offset:scannerOffset])
}

func scannerScanString() string {
	var offset = scannerOffset - 1
	var escaped bool
	for !escaped && scannerCh != '"' {
		if scannerCh == '\\' {
			escaped = true
			scannerNext()
			scannerNext()
			escaped = false
			continue
		}
		scannerNext()
	}
	scannerNext() // consume ending '""
	return string(scannerSrc[offset:scannerOffset])
}

func scannerScanChar() string {
	// '\'' opening already consumed
	var offset = scannerOffset - 1
	var ch uint8
	for {
		ch = scannerCh
		scannerNext()
		if ch == '\'' {
			break
		}
		if ch == '\\' {
			scannerNext()
		}
	}

	return string(scannerSrc[offset:scannerOffset])
}

func scannerrScanComment() string {
	var offset = scannerOffset - 1
	for scannerCh != '\n' {
		scannerNext()
	}
	return string(scannerSrc[offset:scannerOffset])
}

type TokenContainer struct {
	pos int    // what's this ?
	tok string // token.Token
	lit string // raw data
}

// https://golang.org/ref/spec#Tokens
func scannerSkipWhitespace() {
	for scannerCh == ' ' || scannerCh == '\t' || (scannerCh == '\n' && !scannerInsertSemi) || scannerCh == '\r' {
		scannerNext()
	}
}

func scannerScan() *TokenContainer {
	scannerSkipWhitespace()
	var tc = new(TokenContainer)
	var lit string
	var tok string
	var insertSemi bool
	var ch = scannerCh
	if isLetter(ch) {
		lit = scannerScanIdentifier()
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
		lit = scannerScanNumber()
		tok = "INT"
	} else {
		scannerNext()
		switch ch {
		case '\n':
			tok = ";"
			lit = "\n"
			insertSemi = false
		case '"': // double quote
			insertSemi = true
			lit = scannerScanString()
			tok = "STRING"
		case '\'': // single quote
			insertSemi = true
			lit = scannerScanChar()
			tok = "CHAR"
		// https://golang.org/ref/spec#Operators_and_punctuation
		//	+    &     +=    &=     &&    ==    !=    (    )
		//	-    |     -=    |=     ||    <     <=    [    ]
		//  *    ^     *=    ^=     <-    >     >=    {    }
		//	/    <<    /=    <<=    ++    =     :=    ,    ;
		//	%    >>    %=    >>=    --    !     ...   .    :
		//	&^          &^=
		case ':': // :=, :
			if scannerCh == '=' {
				scannerNext()
				tok = ":="
			} else {
				tok = ":"
			}
		case '.': // ..., .
			var peekCh = scannerSrc[scannerNextOffset]
			if scannerCh == '.' && peekCh == '.' {
				scannerNext()
				scannerNext()
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
			switch scannerCh {
			case '=':
				scannerNext()
				tok = "+="
			case '+':
				scannerNext()
				tok = "++"
				insertSemi = true
			default:
				tok = "+"
			}
		case '-': // -= --  -
			switch scannerCh {
			case '-':
				scannerNext()
				tok = "--"
				insertSemi = true
			case '=':
				scannerNext()
				tok = "-="
			default:
				tok = "-"
			}
		case '*': // *=  *
			if scannerCh == '=' {
				scannerNext()
				tok = "*="
			} else {
				tok = "*"
			}
		case '/':
			if scannerCh == '/' {
				// comment
				// @TODO block comment
				if scannerInsertSemi {
					scannerCh = '/'
					scannerOffset = scannerOffset - 1
					scannerNextOffset = scannerOffset + 1
					tc.lit = "\n"
					tc.tok = ";"
					scannerInsertSemi = false
					return tc
				}
				lit = scannerrScanComment()
				tok = "COMMENT"
			} else if scannerCh == '=' {
				tok = "/="
			} else {
				tok = "/"
			}
		case '%': // %= %
			if scannerCh == '=' {
				scannerNext()
				tok = "%="
			} else {
				tok = "%"
			}
		case '^': // ^= ^
			if scannerCh == '=' {
				scannerNext()
				tok = "^="
			} else {
				tok = "^"
			}
		case '<': //  <= <- <<= <<
			switch scannerCh {
			case '-':
				scannerNext()
				tok = "<-"
			case '=':
				scannerNext()
				tok = "<="
			case '<':
				var peekCh = scannerSrc[scannerNextOffset]
				if peekCh == '=' {
					scannerNext()
					scannerNext()
					tok = "<<="
				} else {
					scannerNext()
					tok = "<<"
				}
			default:
				tok = "<"
			}
		case '>': // >= >>= >> >
			switch scannerCh {
			case '=':
				scannerNext()
				tok = ">="
			case '>':
				var peekCh = scannerSrc[scannerNextOffset]
				if peekCh == '=' {
					scannerNext()
					scannerNext()
					tok = ">>="
				} else {
					scannerNext()
					tok = ">>"
				}
			default:
				tok = ">"
			}
		case '=': // == =
			if scannerCh == '=' {
				scannerNext()
				tok = "=="
			} else {
				tok = "="
			}
		case '!': // !=, !
			if scannerCh == '=' {
				scannerNext()
				tok = "!="
			} else {
				tok = "!"
			}
		case '&': // & &= && &^ &^=
			switch scannerCh {
			case '=':
				scannerNext()
				tok = "&="
			case '&':
				scannerNext()
				tok = "&&"
			case '^':
				var peekCh = scannerSrc[scannerNextOffset]
				if peekCh == '=' {
					scannerNext()
					scannerNext()
					tok = "&^="
				} else {
					scannerNext()
					tok = "&^"
				}
			default:
				tok = "&"
			}
		case '|': // |= || |
			switch scannerCh {
			case '|':
				scannerNext()
				tok = "||"
			case '=':
				scannerNext()
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
	scannerInsertSemi = insertSemi
	return tc
}

// --- ast ---
var astCon string = "Con"
var astTyp string = "Typ"
var astVar string = "Var"
var astFun string = "Fun"

type astImportSpec struct {
	Path string
}

// Pseudo interface for *ast.Decl
type astDecl struct {
	dtype    string
	genDecl  *astGenDecl
	funcDecl *astFuncDecl
}

type astFuncDecl struct {
	Name *astIdent
	Sig  *signature
	Body *astBlockStmt
}

type astField struct {
	Name   *astIdent
	Type   *astExpr
	Offset int
}

type astFieldList struct {
	List []*astField
}

type signature struct {
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

type astBlockStmt struct {
	List []*astStmt
}

type astAssignStmt struct {
	Lhs []*astExpr
	Tok string
	Rhs []*astExpr
}

type astIncDecStmt struct {
	X   *astExpr
	Tok string
}

type astGenDecl struct {
	Spec *astSpec
}

type ObjDecl struct {
	dtype     string
	valueSpec *astValueSpec
	funcDecl  *astFuncDecl
	typeSpec  *astTypeSpec
	field     *astField
}

type astTypeSpec struct {
	Name *astIdent
	Type *astExpr
}

type astValueSpec struct {
	Name  *astIdent
	Type  *astExpr
	Value *astExpr
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
	ellipsis     *astEllipsis
}

type astObject struct {
	Kind     string
	Name     string
	Decl     *ObjDecl
	Variable *Variable
}

type astIdent struct {
	Name string
	Obj  *astObject
}

type astArrayType struct {
	Len *astExpr
	Elt *astExpr
}

type astStructType struct {
	Fields *astFieldList
}

type astBasicLit struct {
	Kind  string // token.INT, token.CHAR, or token.STRING
	Value string
}

type astSelectorExpr struct {
	X   *astExpr
	Sel *astIdent
}

type astCallExpr struct {
	Fun  *astExpr   // function expression
	Args []*astExpr // function arguments; or nil
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

type astSliceExpr struct {
	X      *astExpr
	Low    *astExpr
	High   *astExpr
	Max    *astExpr
	Slice3 bool
}

type astIndexExpr struct {
	X     *astExpr
	Index *astExpr
}

type astParenExpr struct {
	X *astExpr
}

type astStarExpr struct {
	X *astExpr
}

type astEllipsis struct {
	Elt *astExpr
}

type astCompositeLit struct {
	Type *astExpr
	Elts []*astExpr
}

type astIfStmt struct {
	Init *astStmt
	Cond *astExpr
	Body *astBlockStmt
	Else *astStmt
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

type astCaseClause struct {
	List []*astExpr
	Body []*astStmt
}

type astSwitchStmt struct {
	Tag  *astExpr
	Body *astBlockStmt
	// lableExit string
}

type astReturnStmt struct {
	Results []*astExpr
}

type astBranchStmt struct {
	Tok        string
	Label      string
	currentFor *astStmt
}

type astSpec struct {
	dtype     string
	valueSpec *astValueSpec
	typeSpec  *astTypeSpec
}

type astScope struct {
	Outer   *astScope
	Objects []*objectEntry
}

type astFile struct {
	Name       string
	Decls      []*astDecl
	Unresolved []*astIdent
}

func astNewScope(outer *astScope) *astScope {
	var r = new(astScope)
	r.Outer = outer
	return r
}

func scopeInsert(s *astScope, obj *astObject) {
	if s == nil {
		panic2(__func__, "s sholud not be nil\n")
	}
	var oe = new(objectEntry)
	oe.name = obj.Name
	oe.obj = obj
	s.Objects = append(s.Objects, oe)
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
	scannerInit(src)
	parserNext()
}

type objectEntry struct {
	name string
	obj  *astObject
}

var ptok *TokenContainer

var parserUnresolved []*astIdent

var parserTopScope *astScope
var parserPkgScope *astScope

func openScope() {
	parserTopScope = astNewScope(parserTopScope)
}

func closeScope() {
	parserTopScope = parserTopScope.Outer
}

func parserConsumeComment() {
	parserNext0()
}

func parserNext0() {
	ptok = scannerScan()
}

func parserNext() {
	parserNext0()
	if ptok.tok == ";" {
		logf(" [parser] pointing at : \"%s\" newline (%s)\n", ptok.tok, Itoa(scannerOffset))
	} else if ptok.tok == "IDENT" {
		logf(" [parser] pointing at: IDENT \"%s\" (%s)\n", ptok.lit, Itoa(scannerOffset))
	} else {
		logf(" [parser] pointing at: \"%s\" %s (%s)\n", ptok.tok, ptok.lit, Itoa(scannerOffset))
	}

	if ptok.tok == "COMMENT" {
		for ptok.tok == "COMMENT" {
			parserConsumeComment()
		}
	}
}

func parserExpect(tok string, who string) {
	if ptok.tok != tok {
		var s = fmtSprintf("%s expected, but got %s", []string{tok, ptok.tok})
		panic2(who, s)
	}
	logf(" [%s] consumed \"%s\"\n", who, ptok.tok)
	parserNext()
}

func parserExpectSemi(caller string) {
	if ptok.tok != ")" && ptok.tok != "}" {
		switch ptok.tok {
		case ";":
			logf(" [%s] consumed semicolon %s\n", caller, ptok.tok)
			parserNext()
		default:
			panic2(caller, "semicolon expected, but got token "+ptok.tok)
		}
	}
}

func parseIdent() *astIdent {
	var name string
	if ptok.tok == "IDENT" {
		name = ptok.lit
		parserNext()
	} else {
		panic2(__func__, "IDENT expected, but got "+ptok.tok)
	}
	logf(" [%s] ident name = %s\n", __func__, name)
	var r = new(astIdent)
	r.Name = name
	return r
}

func parserParseImportDecl() *astImportSpec {
	parserExpect("import", __func__)
	var path = ptok.lit
	parserNext()
	parserExpectSemi(__func__)
	var spec = new(astImportSpec)
	spec.Path = path
	return spec
}

func tryVarType(ellipsisOK bool) *astExpr {
	if ellipsisOK && ptok.tok == "..." {
		parserNext() // consume "..."
		var typ = tryIdentOrType()
		if typ != nil {
			parserResolve(typ)
		} else {
			panic2(__func__, "Syntax error")
		}
		var elps = new(astEllipsis)
		elps.Elt = typ
		var r = new(astExpr)
		r.dtype = "*astEllipsis"
		r.ellipsis = elps
		return r
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
	var starExpr = new(astStarExpr)
	starExpr.X = base
	var r = new(astExpr)
	r.dtype = "*astStarExpr"
	r.starExpr = starExpr
	return r
}

func parseArrayType() *astExpr {
	parserExpect("[", __func__)
	var ln *astExpr
	if ptok.tok != "]" {
		ln = parseRhs()
	}
	parserExpect("]", __func__)
	var elt = parseType()
	var arrayType = new(astArrayType)
	arrayType.Elt = elt
	arrayType.Len = ln
	var r = new(astExpr)
	r.dtype = "*astArrayType"
	r.arrayType = arrayType
	return r
}

func parseFieldDecl(scope *astScope) *astField {

	var varType = parseVarType(false)
	var typ = tryVarType(false)

	parserExpectSemi(__func__)

	var field = new(astField)
	field.Type = typ
	field.Name = varType.ident
	declareField(field, scope, astVar, varType.ident)
	parserResolve(typ)
	return field
}

func parseStructType() *astExpr {
	parserExpect("struct", __func__)
	parserExpect("{", __func__)

	var _nil *astScope
	var scope = astNewScope(_nil)

	var structType = new(astStructType)
	var list []*astField
	for ptok.tok == "IDENT" || ptok.tok == "*" {
		var field *astField = parseFieldDecl(scope)
		list = append(list, field)
	}
	parserExpect("}", __func__)

	var fields = new(astFieldList)
	fields.List = list
	structType.Fields = fields
	var r = new(astExpr)
	r.dtype = "*astStructType"
	r.structType = structType
	return r
}

func parseTypeName() *astExpr {
	logf(" [%s] begin\n", __func__)
	var ident = parseIdent()
	var typ = new(astExpr)
	typ.ident = ident
	typ.dtype = "*astIdent"
	logf(" [%s] end\n", __func__)
	return typ
}

func tryIdentOrType() *astExpr {
	logf(" [%s] begin\n", __func__)
	switch ptok.tok {
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
		var parenExpr = new(astParenExpr)
		parenExpr.X = _typ
		var typ = new(astExpr)
		typ.dtype = "*astParenExpr"
		typ.parenExpr = parenExpr
		return typ
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
		if ptok.tok != "," {
			break
		}
		parserNext()
		if ptok.tok == ")" {
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
		var field = new(astField)
		if ident == nil {
			panic2(__func__, "Ident should not be nil")
		}
		logf(" [%s] ident.Name=%s\n", __func__, ident.Name)
		logf(" [%s] typ=%s\n", __func__, typ.dtype)
		field.Name = ident
		field.Type = typ
		logf(" [%s]: Field %s %s\n", __func__, field.Name.Name, field.Type.dtype)
		params = append(params, field)
		declareField(field, scope, astVar, ident)
		parserResolve(typ)
		if ptok.tok != "," {
			logf("  end %s\n", __func__)
			return params
		}
		parserNext()
		for ptok.tok != ")" && ptok.tok != "EOF" {
			ident = parseIdent()
			typ = parseVarType(ellipsisOK)
			field = new(astField)
			field.Name = ident
			field.Type = typ
			params = append(params, field)
			declareField(field, scope, astVar, ident)
			parserResolve(typ)
			if ptok.tok != "," {
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
		var field = new(astField)
		field.Type = typ
		params[i] = field
		logf(" [DEBUG] range i = %s\n", Itoa(i))
	}
	logf("  end %s\n", __func__)
	return params
}

func parseParameters(scope *astScope, ellipsisOk bool) *astFieldList {
	logf(" [%s] begin\n", __func__)
	var params []*astField
	parserExpect("(", __func__)
	if ptok.tok != ")" {
		params = parseParameterList(scope, ellipsisOk)
	}
	parserExpect(")", __func__)
	var afl = new(astFieldList)
	afl.List = params
	logf(" [%s] end\n", __func__)
	return afl
}

func parserResult(scope *astScope) *astFieldList {
	logf(" [%s] begin\n", __func__)

	if ptok.tok == "(" {
		var r = parseParameters(scope, false)
		logf(" [%s] end\n", __func__)
		return r
	}

	var r = new(astFieldList)
	if ptok.tok == "{" {
		r = nil
		logf(" [%s] end\n", __func__)
		return r
	}
	var typ = tryType()
	var field = new(astField)
	field.Type = typ
	r.List = append(r.List, field)
	logf(" [%s] end\n", __func__)
	return r
}

func parseSignature(scope *astScope) *signature {
	logf(" [%s] begin\n", __func__)
	var params *astFieldList
	var results *astFieldList
	params = parseParameters(scope, true)
	results = parserResult(scope)
	var sig = new(signature)
	sig.params = params
	sig.results = results
	return sig
}

func declareField(decl *astField, scope *astScope, kind string, ident *astIdent) {
	// delcare
	var obj = new(astObject)
	var objDecl = new(ObjDecl)
	objDecl.dtype = "*astField"
	objDecl.field = decl
	obj.Decl = objDecl
	obj.Name = ident.Name
	obj.Kind = kind
	ident.Obj = obj

	// scope insert
	if ident.Name != "_" {
		scopeInsert(scope, obj)
	}
}

func declare(objDecl *ObjDecl, scope *astScope, kind string, ident *astIdent) {
	logf(" [declare] ident %s\n", ident.Name)

	var obj = new(astObject) //valSpec.Name.Obj
	obj.Decl = objDecl
	obj.Name = ident.Name
	obj.Kind = kind
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
	for s = parserTopScope; s != nil; s = s.Outer {
		var obj = scopeLookup(s, ident.Name)
		if obj != nil {
			ident.Obj = obj
			return
		}
	}

	if collectUnresolved {
		parserUnresolved = append(parserUnresolved, ident)
		logf(" appended unresolved ident %s\n", ident.Name)
	}
}

func parseOperand() *astExpr {
	logf("   begin %s\n", __func__)
	switch ptok.tok {
	case "IDENT":
		var eIdent = new(astExpr)
		eIdent.dtype = "*astIdent"
		var ident = parseIdent()
		eIdent.ident = ident
		tryResolve(eIdent, true)
		logf("   end %s\n", __func__)
		return eIdent
	case "INT", "STRING", "CHAR":
		var basicLit = new(astBasicLit)
		basicLit.Kind = ptok.tok
		basicLit.Value = ptok.lit
		var r = new(astExpr)
		r.dtype = "*astBasicLit"
		r.basicLit = basicLit
		parserNext()
		logf("   end %s\n", __func__)
		return r
	case "(":
		parserNext() // consume "("
		parserExprLev++
		var x = parserRhsOrType()
		parserExprLev--
		parserExpect(")", __func__)
		var p = new(astParenExpr)
		p.X = x
		var r = new(astExpr)
		r.dtype = "*astParenExpr"
		r.parenExpr = p
		return r
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
	var callExpr = new(astCallExpr)
	callExpr.Fun = fn
	logf(" [parsePrimaryExpr] ptok.tok=%s\n", ptok.tok)
	var list []*astExpr
	for ptok.tok != ")" {
		var arg = parseExpr()
		list = append(list, arg)
		if ptok.tok == "," {
			parserNext()
		} else if ptok.tok == ")" {
			break
		}
	}
	parserExpect(")", __func__)
	callExpr.Args = list
	var r = new(astExpr)
	r.dtype = "*astCallExpr"
	r.callExpr = callExpr
	return r
}

var parserExprLev int // < 0: in control clause, >= 0: in expression

func parsePrimaryExpr() *astExpr {
	logf("   begin %s\n", __func__)
	var x = parseOperand()

	var cnt int

	for {
		cnt++
		logf("    [%s] tok=%s\n", __func__, ptok.tok)
		if cnt > 100 {
			panic2(__func__, "too many iteration")
		}

		switch ptok.tok {
		case ".":
			parserNext() // consume "."
			if ptok.tok != "IDENT" {
				panic2(__func__, "tok should be IDENT")
			}
			// Assume CallExpr
			var secondIdent = parseIdent()
			var sel = new(astSelectorExpr)
			sel.X = x
			sel.Sel = secondIdent
			if ptok.tok == "(" {
				var fn = new(astExpr)
				fn.dtype = "*astSelectorExpr"
				fn.selectorExpr = sel
				// string = x.ident.Name + "." + secondIdent
				x = parseCallExpr(fn)
				logf(" [parsePrimaryExpr] 741 ptok.tok=%s\n", ptok.tok)
			} else {
				logf("   end parsePrimaryExpr()\n")
				x = new(astExpr)
				x.dtype = "*astSelectorExpr"
				x.selectorExpr = sel
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

func parseLiteralValue(x *astExpr) *astExpr {
	logf("   start %s\n", __func__)
	parserExpect("{", __func__)
	var elts []*astExpr
	var e *astExpr
	for ptok.tok != "}" {
		e = parseExpr()
		elts = append(elts, e)
		if ptok.tok == "}" {
			break
		} else {
			parserExpect(",", __func__)
		}
	}
	parserExpect("}", __func__)

	var compositeLit = new(astCompositeLit)
	compositeLit.Type = x
	compositeLit.Elts = elts
	var r = new(astExpr)
	r.dtype = "*astCompositeLit"
	r.compositeLit = compositeLit
	logf("   end %s\n", __func__)
	return r
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
	if ptok.tok != ":" {
		index[0] = parseRhs()
	}
	var ncolons int
	for ptok.tok == ":" && ncolons < 2 {
		ncolons++
		parserNext() // consume ":"
		if ptok.tok != ":" && ptok.tok != "]" {
			index[ncolons] = parseRhs()
		}
	}
	parserExpect("]", __func__)

	if ncolons > 0 {
		// slice expression
		if ncolons == 2 {
			panic2(__func__, "TBI: ncolons=2")
		}
		var sliceExpr = new(astSliceExpr)
		sliceExpr.Slice3 = false
		sliceExpr.X = x
		sliceExpr.Low = index[0]
		sliceExpr.High = index[1]
		var r = new(astExpr)
		r.dtype = "*astSliceExpr"
		r.sliceExpr = sliceExpr
		return r
	}

	var indexExpr = new(astIndexExpr)
	indexExpr.X = x
	indexExpr.Index = index[0]
	var r = new(astExpr)
	r.dtype = "*astIndexExpr"
	r.indexExpr = indexExpr
	return r
}

func parseUnaryExpr() *astExpr {
	var r *astExpr
	logf("   begin parseUnaryExpr()\n")
	switch ptok.tok {
	case "+", "-", "!", "&":
		var tok = ptok.tok
		parserNext()
		var x = parseUnaryExpr()
		r = new(astExpr)
		r.dtype = "*astUnaryExpr"
		r.unaryExpr = new(astUnaryExpr)
		logf(" [DEBUG] unary op = %s\n", tok)
		r.unaryExpr.Op = tok
		r.unaryExpr.X = x
		return r
	case "*":
		parserNext() // consume "*"
		var x = parseUnaryExpr()
		r = new(astExpr)
		r.dtype = "*astStarExpr"
		r.starExpr = new(astStarExpr)
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
		var op = ptok.tok
		oprec = precedence(op)
		logf(" oprec %s\n", Itoa(oprec))
		logf(" precedence \"%s\" %s < %s\n", op, Itoa(oprec), Itoa(prec1))
		if oprec < prec1 {
			logf("   end parseBinaryExpr() (NonBinary)\n")
			return x
		}
		parserExpect(op, __func__)
		var y = parseBinaryExpr(oprec + 1)
		var binaryExpr = new(astBinaryExpr)
		binaryExpr.X = x
		binaryExpr.Y = y
		binaryExpr.Op = op
		var r = new(astExpr)
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
	if ptok.tok != "{" {
		if ptok.tok != ";" {
			s2 = parseSimpleStmt(true)
			isRange = s2.isRange
			logf(" [%s] isRange=true\n", __func__)
		}
		if !isRange && ptok.tok == ";" {
			parserNext() // consume ";"
			s1 = s2
			s2 = nil
			if ptok.tok != ";" {
				s2 = parseSimpleStmt(false)
			}
			parserExpectSemi(__func__)
			if ptok.tok != "{" {
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
		var rangeStmt = new(astRangeStmt)
		rangeStmt.Key = key
		rangeStmt.Value = value
		rangeStmt.X = rangeX
		rangeStmt.Body = body
		var r = new(astStmt)
		r.dtype = "*astRangeStmt"
		r.rangeStmt = rangeStmt
		closeScope()
		logf(" end %s\n", __func__)
		return r
	}
	var forStmt = new(astForStmt)
	forStmt.Init = s1
	forStmt.Cond = makeExpr(s2)
	forStmt.Post = s3
	forStmt.Body = body
	var r = new(astStmt)
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
	if ptok.tok == "else" {
		parserNext()
		if ptok.tok == "if" {
			else_ = parseIfStmt()
		} else {
			var elseblock = parseBlockStmt()
			parserExpectSemi(__func__)
			else_ = new(astStmt)
			else_.dtype = "*astBlockStmt"
			else_.blockStmt = elseblock
		}
	} else {
		parserExpectSemi(__func__)
	}
	var ifStmt = new(astIfStmt)
	ifStmt.Cond = cond
	ifStmt.Body = body
	ifStmt.Else = else_

	var r = new(astStmt)
	r.dtype = "*astIfStmt"
	r.ifStmt = ifStmt
	return r
}

func parseCaseClause() *astCaseClause {
	logf(" [%s] start\n", __func__)
	var list []*astExpr
	if ptok.tok == "case" {
		parserNext() // consume "case"
		list = parseRhsList()
	} else {
		parserExpect("default", __func__)
	}

	parserExpect(":", __func__)
	openScope()
	var body = parseStmtList()
	var r = new(astCaseClause)
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
	for ptok.tok == "case" || ptok.tok == "default" {
		cc = parseCaseClause()
		ccs = new(astStmt)
		ccs.dtype = "*astCaseClause"
		ccs.caseClause = cc
		list = append(list, ccs)
	}
	parserExpect("}", __func__)
	parserExpectSemi(__func__)
	var body = new(astBlockStmt)
	body.List = list

	var switchStmt = new(astSwitchStmt)
	switchStmt.Body = body
	switchStmt.Tag = makeExpr(s2)
	var s = new(astStmt)
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
	var s = new(astStmt)
	var x = parseLhsList()
	var stok = ptok.tok
	var isRange = false
	var y *astExpr
	var rangeX *astExpr
	var rangeUnary *astUnaryExpr
	switch stok {
	case "=":
		parserNext() // consume =
		if isRangeOK && ptok.tok == "range" {
			parserNext() // consume "range"
			rangeX = parseRhs()
			rangeUnary = new(astUnaryExpr)
			rangeUnary.Op = "range"
			rangeUnary.X = rangeX
			y = new(astExpr)
			y.dtype = "*astUnaryExpr"
			y.unaryExpr = rangeUnary
			isRange = true
		} else {
			y = parseExpr() // rhs
		}
		var as = new(astAssignStmt)
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
		var exprStmt = new(astExprStmt)
		exprStmt.X = x[0]
		s.exprStmt = exprStmt
		logf(" end %s\n", __func__)
		return s
	}

	switch stok {
	case "++", "--":
		var s = new(astStmt)
		var sInc = new(astIncDecStmt)
		sInc.X = x[0]
		sInc.Tok = stok
		s.dtype = "*astIncDecStmt"
		s.incDecStmt = sInc
		parserNext() // consume "++" or "--"
		return s
	}
	var exprStmt = new(astExprStmt)
	exprStmt.X = x[0]
	var r = new(astStmt)
	r.dtype = "*astExprStmt"
	r.exprStmt = exprStmt
	logf(" end %s\n", __func__)
	return r
}

func parseStmt() *astStmt {
	logf("\n")
	logf(" = begin %s\n", __func__)
	var s *astStmt
	switch ptok.tok {
	case "var":
		var genDecl = parseDecl("var")
		s = new(astStmt)
		s.dtype = "*astDeclStmt"
		s.DeclStmt = new(astDeclStmt)
		var decl = new(astDecl)
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
		s = parseBranchStmt(ptok.tok)
	case "if":
		s = parseIfStmt()
	case "switch":
		s = parseSwitchStmt()
	case "for":
		s = parseForStmt()
	default:
		panic2(__func__, "TBI 3:"+ptok.tok)
	}
	logf(" = end parseStmt()\n")
	return s
}

func parseExprList() []*astExpr {
	logf(" [%s] start\n", __func__)
	var list []*astExpr
	var e = parseExpr()
	list = append(list, e)
	for ptok.tok == "," {
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

	var branchStmt = new(astBranchStmt)
	branchStmt.Tok = tok
	var s = new(astStmt)
	s.dtype = "*astBranchStmt"
	s.branchStmt = branchStmt
	return s
}

func parseReturnStmt() *astStmt {
	parserExpect("return", __func__)
	var x []*astExpr
	if ptok.tok != ";" && ptok.tok != "}" {
		x = parseRhsList()
	}
	parserExpectSemi(__func__)
	var returnStmt = new(astReturnStmt)
	returnStmt.Results = x
	var r = new(astStmt)
	r.dtype = "*astReturnStmt"
	r.returnStmt = returnStmt
	return r
}

func parseStmtList() []*astStmt {
	var list []*astStmt
	for ptok.tok != "}" && ptok.tok != "EOF" && ptok.tok != "case" && ptok.tok != "default" {
		var stmt = parseStmt()
		list = append(list, stmt)
	}
	return list
}

func parseBody(scope *astScope) *astBlockStmt {
	parserExpect("{", __func__)
	parserTopScope = scope
	logf(" begin parseStmtList()\n")
	var list = parseStmtList()
	logf(" end parseStmtList()\n")

	closeScope()
	parserExpect("}", __func__)
	var r = new(astBlockStmt)
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
	var r = new(astBlockStmt)
	r.List = list
	return r
}

func parseDecl(keyword string) *astGenDecl {
	var r *astGenDecl
	switch ptok.tok {
	case "var":
		parserExpect(keyword, __func__)
		var ident = parseIdent()
		var typ = parseType()
		var value *astExpr
		if ptok.tok == "=" {
			parserNext()
			value = parseExpr()
		}
		parserExpectSemi(__func__)
		var valSpec = new(astValueSpec)
		valSpec.Name = ident
		valSpec.Type = typ
		valSpec.Value = value
		var spec = new(astSpec)
		spec.dtype = "*astValueSpec"
		spec.valueSpec = valSpec
		var objDecl = new(ObjDecl)
		objDecl.dtype = "*astValueSpec"
		objDecl.valueSpec = valSpec
		declare(objDecl, parserTopScope, astVar, ident)
		r = new(astGenDecl)
		r.Spec = spec
		return r
	default:
		panic2(__func__, "TBI\n")
	}
	return r
}

func parserParseTypeSpec() *astSpec {
	logf(" [%s] start\n", __func__)
	parserExpect("type", __func__)
	var ident = parseIdent()
	logf(" decl type %s\n", ident.Name)

	var spec = new(astTypeSpec)
	spec.Name = ident
	var objDecl = new(ObjDecl)
	objDecl.dtype = "*astTypeSpec"
	objDecl.typeSpec = spec
	declare(objDecl, parserTopScope, astTyp, ident)
	var typ = parseType()
	parserExpectSemi(__func__)
	spec.Type = typ
	var r = new(astSpec)
	r.dtype = "*astTypeSpec"
	r.typeSpec = spec
	return r
}

func parserParseValueSpec(keyword string) *astSpec {
	logf(" [parserParseValueSpec] start\n")
	parserExpect(keyword, __func__)
	var ident = parseIdent()
	logf(" var = %s\n", ident.Name)
	var typ = parseType()
	var value *astExpr
	if ptok.tok == "=" {
		parserNext()
		value = parseExpr()
	}
	parserExpectSemi(__func__)
	var spec = new(astValueSpec)
	spec.Name = ident
	spec.Type = typ
	spec.Value = value
	var r = new(astSpec)
	r.dtype = "*astValueSpec"
	r.valueSpec = spec
	var objDecl = new(ObjDecl)
	objDecl.dtype = "*astValueSpec"
	objDecl.valueSpec = spec
	var kind = astCon
	if keyword == "var" {
		kind = astVar
	}
	declare(objDecl, parserTopScope, kind, ident)
	logf(" [parserParseValueSpec] end\n")
	return r
}

func parserParseFuncDecl() *astDecl {
	parserExpect("func", __func__)
	var scope = astNewScope(parserTopScope) // function scope

	var ident = parseIdent()
	var sig = parseSignature(scope)
	if sig.results == nil {
		logf(" [parserParseFuncDecl] %s sig.results is nil\n", ident.Name)
	} else {
		logf(" [parserParseFuncDecl] %s sig.results.List = %s\n", ident.Name, Itoa(len(sig.results.List)))
	}
	var body *astBlockStmt
	if ptok.tok == "{" {
		logf(" begin parseBody()\n")
		body = parseBody(scope)
		logf(" end parseBody()\n")
		parserExpectSemi(__func__)
	} else {
		parserExpectSemi(__func__)
	}
	var decl = new(astDecl)
	decl.dtype = "*astFuncDecl"
	var funcDecl = new(astFuncDecl)
	decl.funcDecl = funcDecl
	decl.funcDecl.Name = ident
	decl.funcDecl.Sig = sig
	decl.funcDecl.Body = body
	var objDecl = new(ObjDecl)
	objDecl.dtype = "*astFuncDecl"
	objDecl.funcDecl = funcDecl
	declare(objDecl, parserPkgScope, astFun, ident)
	return decl
}

func parserParseFile() *astFile {
	// expect "package" keyword
	parserExpect("package", __func__)
	parserUnresolved = nil
	var ident = parseIdent()
	var packageName = ident.Name
	parserExpectSemi(__func__)

	parserTopScope = new(astScope) // open scope
	parserPkgScope = parserTopScope

	for ptok.tok == "import" {
		parserParseImportDecl()
	}

	logf("\n")
	logf(" [parser] Parsing Top level decls\n")
	var decls []*astDecl
	var decl *astDecl

	for ptok.tok != "EOF" {
		switch ptok.tok {
		case "var", "const":
			var spec = parserParseValueSpec(ptok.tok)
			var genDecl = new(astGenDecl)
			genDecl.Spec = spec
			decl = new(astDecl)
			decl.dtype = "*astGenDecl"
			decl.genDecl = genDecl
		case "func":
			logf("\n\n")
			decl = parserParseFuncDecl()
			logf(" func decl parsed:%s\n", decl.funcDecl.Name.Name)
		case "type":
			var spec = parserParseTypeSpec()
			var genDecl = new(astGenDecl)
			genDecl.Spec = spec
			decl = new(astDecl)
			decl.dtype = "*astGenDecl"
			decl.genDecl = genDecl
			logf(" type parsed:%s\n", "")
		default:
			panic2(__func__, "TBI:"+ptok.tok)
		}
		decls = append(decls, decl)
	}

	parserTopScope = nil

	// dump parserPkgScope
	logf("[DEBUG] Dump objects in the package scope\n")
	var oe *objectEntry
	for _, oe = range parserPkgScope.Objects {
		logf("    object %s\n", oe.name)
	}

	var unresolved []*astIdent
	var idnt *astIdent
	logf(" [parserParseFile] resolving parserUnresolved (n=%s)\n", Itoa(len(parserUnresolved)))
	for _, idnt = range parserUnresolved {
		logf(" [parserParseFile] resolving ident %s ...\n", idnt.Name)
		var obj *astObject = scopeLookup(parserPkgScope, idnt.Name)
		if obj != nil {
			logf(" resolved \n")
			idnt.Obj = obj
		} else {
			logf(" unresolved \n")
			unresolved = append(unresolved, idnt)
		}
	}
	logf(" [parserParseFile] Unresolved (n=%s)\n", Itoa(len(unresolved)))

	var f = new(astFile)
	f.Name = packageName
	f.Decls = decls
	f.Unresolved = unresolved
	logf(" [%s] end\n", __func__)
	return f
}

func parseFile(filename string) *astFile {
	var text = readSource(filename)
	parserInit(text)
	return parserParseFile()
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
	case T_ARRAY:
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
		//@FIXME
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
		var arg = new(Arg)
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
				var eNumLit = new(astExpr)
				eNumLit.dtype = "*astBasicLit"
				eNumLit.basicLit = numlit
				var args []*Arg
				var arg0 *Arg // elmSize
				var arg1 *Arg
				var arg2 *Arg
				arg0 = new(Arg)
				arg0.e = eNumLit
				arg0.t = tInt
				arg1 = new(Arg) // len
				arg1.e = eArgs[1]
				arg1.t = tInt
				arg2 = new(Arg) // cap
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
			var arg0 = new(Arg) // slice
			arg0.e = sliceArg
			args = append(args, arg0)

			var arg1 = new(Arg) // element
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
		var symbol = pkgName + "." + fn.Name
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
		if fndecl.Sig == nil {
			panic2(__func__, "[*astCallExpr] fndecl.Sig is nil")
		}

		var params = fndecl.Sig.params.List
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
			arg = new(Arg)
			arg.e = eArg
			arg.t = paramType
			args = append(args, arg)
		}

		if variadicElp != nil {
			// collect args as a slice
			var sliceType = new(astArrayType)
			sliceType.Elt = variadicElp.Elt
			var eSliceType = new(astExpr)
			eSliceType.dtype = "*astArrayType"
			eSliceType.arrayType = sliceType
			var sliceLiteral = new(astCompositeLit)
			sliceLiteral.Type = eSliceType
			sliceLiteral.Elts = variadicArgs
			var eSliceLiteral = new(astExpr)
			eSliceLiteral.compositeLit = sliceLiteral
			eSliceLiteral.dtype = "*astCompositeLit"
			var _arg = new(Arg)
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

			var _arg = new(Arg)
			_arg.e = eNil
			_arg.t = e2t(param.Type)
			args = append(args, _arg)
		}

		emitCall(symbol, args)

		// push results
		var results = fndecl.Sig.results
		if fndecl.Sig.results == nil {
			emitComment(0, "[emitExpr] %s sig.results is nil\n", fn.Name)
		} else {
			emitComment(0, "[emitExpr] %s sig.results.List = %s\n", fn.Name, Itoa(len(fndecl.Sig.results.List)))
		}

		if results != nil && len(results.List) == 1 {
			var retval0 = fndecl.Sig.results.List[0]
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
			var argX = new(Arg)
			var argY = new(Arg)
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
	var r = new(astBasicLit)
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
	case T_STRUCT:
		// @FXIME
	case T_ARRAY:
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
		var lhs = new(astExpr)
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
		var bodyStmt = new(astStmt)
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
	var stmt = new(astStmt)
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
		fmtPrintf("  .quad 0 # bool zero value\n") // @TODO
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

func emitData(pkgName string) {
	fmtPrintf(".data\n")
	emitComment(0, "string literals len = %s\n", Itoa(len(stringLiterals)))
	var con *stringLiteralsContainer
	for _, con = range stringLiterals {
		emitComment(0, "string literals\n")
		fmtPrintf("%s:\n", con.sl.label)
		fmtPrintf("  .string %s\n", con.sl.value)
	}

	emitComment(0, "===== Global Variables =====\n")

	var spec *astValueSpec
	for _, spec = range globalVars {
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariable(spec.Name, t, spec.Value)
	}

	emitComment(0, "==============================\n")
}

func emitText(pkgName string) {
	fmtPrintf(".text\n")
	var fnc *Func
	for _, fnc = range globalFuncs {
		emitFuncDecl(pkgName, fnc)
	}
}

func generateCode(pkgName string) {
	emitData(pkgName)
	emitText(pkgName)
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
				var t = new(Type)
				t.e = decl.Type
				return t
			case "*astField":
				var decl = expr.ident.Obj.Decl.field
				var t = new(Type)
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
					var starExpr = new(astStarExpr)
					starExpr.X = expr.callExpr.Args[0]
					var eStarExpr = new(astExpr)
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
					var resultList = decl.funcDecl.Sig.results.List
					if len(resultList) != 1 {
						panic2(__func__, "[astCallExpr] len results.List is not 1")
					}
					return e2t(decl.funcDecl.Sig.results.List[0].Type)
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
		var t = new(astArrayType)
		t.Len = nil
		t.Elt = elementTyp
		var e = new(astExpr)
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
	var r = new(Type)
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
		panic2(__func__, "not ident")
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
var globalVars []*astValueSpec
var globalFuncs []*Func
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

	if pkgName == "" {
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

	var label = fmtSprintf(".%s.S%d", []string{pkgName, Itoa(stringIndex)})
	stringIndex++

	var sl = new(sliteral)
	sl.label = label
	sl.strlen = strlen - 2
	sl.value = lit.Value
	logf(" [registerStringLiteral] label=%s, strlen=%s\n", sl.label, Itoa(sl.strlen))
	var cont = new(stringLiteralsContainer)
	cont.sl = sl
	cont.lit = lit
	stringLiterals = append(stringLiterals, cont)
}

func newGlobalVariable(name string) *Variable {
	var vr = new(Variable)
	vr.name = name
	vr.isGlobal = true
	vr.globalSymbol = name
	return vr
}

func newLocalVariable(name string, localoffset int) *Variable {
	var vr = new(Variable)
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
		var _s = new(astStmt)
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
					basicLit = new(astBasicLit)
					basicLit.Kind = "STRING"
					basicLit.Value = "\"" + currentFuncDecl.Name.Name + "\""
					newArg = new(astExpr)
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
	default:
		panic2(__func__, "TBI:"+expr.dtype)
	}
}

func walk(file *astFile) string {
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
					globalVars = append(globalVars, valSpec)
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
			for _, field = range funcDecl.Sig.params.List {
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
				var fnc = new(Func)
				fnc.name = funcDecl.Name.Name
				fnc.Body = funcDecl.Body
				fnc.localarea = localoffset
				fnc.argsarea = paramoffset

				globalFuncs = append(globalFuncs, fnc)
			}
		default:
			panic2(__func__, "TBI: "+decl.dtype)
		}
	}

	if len(stringLiterals) == 0 {
		panic2(__func__, "stringLiterals is empty\n")
	}

	return pkgName
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

	// @FIXME package names should be be in universe
	var pkgOs = new(astObject)
	pkgOs.Kind = "Pkg"
	pkgOs.Name = "os"
	scopeInsert(universe, pkgOs)

	var pkgSyscall = new(astObject)
	pkgSyscall.Kind = "Pkg"
	pkgSyscall.Name = "syscall"
	scopeInsert(universe, pkgSyscall)

	var pkgUnsafe = new(astObject)
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
			panic2(__func__, "Unresolved : "+ident.Name)
			unresolved = append(unresolved, ident)
		}
	}
}

func initGlobals() {
	gNil = new(astObject)
	gNil.Kind = astCon // is it Con ?
	gNil.Name = "nil"

	identNil = new(astIdent)
	identNil.Obj = gNil
	identNil.Name = "nil"
	eNil = new(astExpr)
	eNil.dtype = "*astIdent"
	eNil.ident = identNil

	gTrue = new(astObject)
	gTrue.Kind = astCon
	gTrue.Name = "true"

	gFalse = new(astObject)
	gFalse.Kind = astCon
	gFalse.Name = "false"

	gString = new(astObject)
	gString.Kind = astTyp
	gString.Name = "string"
	tString = new(Type)
	tString.e = new(astExpr)
	tString.e.dtype = "*astIdent"
	tString.e.ident = new(astIdent)
	tString.e.ident.Name = "string"
	tString.e.ident.Obj = gString

	gInt = new(astObject)
	gInt.Kind = astTyp
	gInt.Name = "int"
	tInt = new(Type)
	tInt.e = new(astExpr)
	tInt.e.dtype = "*astIdent"
	tInt.e.ident = new(astIdent)
	tInt.e.ident.Name = "int"
	tInt.e.ident.Obj = gInt

	gUint8 = new(astObject)
	gUint8.Kind = astTyp
	gUint8.Name = "uint8"
	tUint8 = new(Type)
	tUint8.e = new(astExpr)
	tUint8.e.dtype = "*astIdent"
	tUint8.e.ident = new(astIdent)
	tUint8.e.ident.Name = "uint8"
	tUint8.e.ident.Obj = gUint8

	gUint16 = new(astObject)
	gUint16.Kind = astTyp
	gUint16.Name = "uint16"
	tUint16 = new(Type)
	tUint16.e = new(astExpr)
	tUint16.e.dtype = "*astIdent"
	tUint16.e.ident = new(astIdent)
	tUint16.e.ident.Name = "uint16"
	tUint16.e.ident.Obj = gUint16

	gUintptr = new(astObject)
	gUintptr.Kind = astTyp
	gUintptr.Name = "uintptr"
	tUintptr = new(Type)
	tUintptr.e = new(astExpr)
	tUintptr.e.dtype = "*astIdent"
	tUintptr.e.ident = new(astIdent)
	tUintptr.e.ident.Name = "uintptr"
	tUintptr.e.ident.Obj = gUintptr

	gBool = new(astObject)
	gBool.Kind = astTyp
	gBool.Name = "bool"
	tBool = new(Type)
	tBool.e = new(astExpr)
	tBool.e.dtype = "*astIdent"
	tBool.e.ident = new(astIdent)
	tBool.e.ident.Name = "bool"
	tBool.e.ident.Obj = gBool

	gNew = new(astObject)
	gNew.Kind = astFun
	gNew.Name = "new"

	gMake = new(astObject)
	gMake.Kind = astFun
	gMake.Name = "make"

	gAppend = new(astObject)
	gAppend.Kind = astFun
	gAppend.Name = "append"

	gLen = new(astObject)
	gLen.Kind = astFun
	gLen.Name = "len"

	gCap = new(astObject)
	gCap.Kind = astFun
	gCap.Name = "cap"
}

var pkgName string

func main() {
	initGlobals()

	var sourceFiles = []string{"runtime.go", "/dev/stdin"}
	var sourceFile string

	var universe = createUniverse()

	for _, sourceFile = range sourceFiles {
		fmtPrintf("# file: %s\n", sourceFile)
		globalVars = nil
		globalFuncs = nil
		stringIndex = 0
		stringLiterals = nil
		var f = parseFile(sourceFile)
		resolveUniverse(f, universe)
		pkgName = f.Name
		walk(f)
		generateCode(pkgName)
	}
}
