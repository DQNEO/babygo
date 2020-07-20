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

var __func__ string

func panic(x string) {
	var s = "panic: " + x+"\n\n"
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

func Itoa(ival int) string {
	if ival == 0 {
		return "0"
	}

	var __itoa_buf []uint8 = make([]uint8, 100,100)
	var __itoa_r []uint8 = make([]uint8, 100, 100)

	var next int
	var right int
	var ix = 0
	var minus bool
	minus = false
	for ix = 0; ival != 0; ix = ix + 1 {
		if ival < 0 {
			ival = -1 * ival
			minus = true
			__itoa_r[0] = '-'
		} else {
			next = ival / 10
			right = ival - next*10
			ival = next
			__itoa_buf[ix] = uint8('0' + right)
		}
	}

	var j int
	var c uint8
	for j = 0; j < ix; j = j + 1 {
		c = __itoa_buf[ix-j-1]
		if minus {
			__itoa_r[j+1] = c
		} else {
			__itoa_r[j] = c
		}
	}

	return string(__itoa_r[0:ix])
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
	fmtPrintf("# src len = %s\n", Itoa(len(scannerSrc)))
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
	//fmtPrintf("begin: scannerScanString\n")
	var offset = scannerOffset - 1
	for scannerCh != '"' {
		scannerNext()
	}
	scannerNext() // consume ending '""
	return string(scannerSrc[offset:scannerOffset])
}

func scannerScanChar() string {
	//fmtPrintf("begin: scannerScanString\n")
	var offset = scannerOffset - 1
	for scannerCh != '\'' {
		//fmtPrintf("in loop char:%s\n", string([]uint8{scannerCh}))
		scannerNext()
	}
	scannerNext() // consume ending '""
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
				tok = "&="
			case '&':
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
			panic2(__func__, "unknown char:" + string([]uint8{ch}) + ":" + Itoa(int(ch)))
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
type astImportSpec struct {
	Path string
}

// Pseudo interface for *ast.Decl
type astDecl struct {
	dtype string
	genDecl *astGenDecl
	funcDecl *astFuncDecl
}

type astFuncDecl struct {
	Name *astIdent
	Sig  *signature
	Body *astBlockStmt
}

type astField struct {
	Name *astIdent
	Type *astExpr
	Offset int
}

type astFieldList struct {
	List []*astField
}

type signature struct {
	params *astFieldList
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
	isRange bool
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
	Lhs *astExpr
	Tok string
	Rhs *astExpr
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
	Name *astIdent
	Type *astExpr
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
}

type astObject struct {
	Kind string
	Name string
	Decl *ObjDecl
	Variable *Variable
}

type astIdent struct {
	Name string
	Obj *astObject
}

type astArrayType struct {
	Len *astExpr
	Elt *astExpr
}

type astStructType struct {
	Fields     *astFieldList
}

type astBasicLit struct {
	Kind     string   // token.INT, token.CHAR, or token.STRING
	Value    string
}

type astSelectorExpr struct {
	X *astExpr
	Sel *astIdent
}

type astCallExpr struct {
	Fun      *astExpr      // function expression
	Args     []*astExpr    // function arguments; or nil
}

type astUnaryExpr struct {
	X *astExpr
	Op string
}

type astBinaryExpr struct {
	X *astExpr
	Y *astExpr
	Op string
}

type astSliceExpr struct {
	X *astExpr
	Low *astExpr
	High *astExpr
	Max *astExpr
	Slice3 bool
}

type astIndexExpr struct {
	X *astExpr
	Index *astExpr
}

type astParenExpr struct {
	X *astExpr
}

type astStarExpr struct {
	X *astExpr
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
	labelPost string
	labelExit string
}

type astReturnStmt struct {
	Results []*astExpr
}

type astSpec struct {
	dtype string
	valueSpec *astValueSpec
	typeSpec *astTypeSpec
}

type astScope struct {
	Outer *astScope
	Objects []*objectEntry
}

type astFile struct {
	Name string
	Decls []*astDecl
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
	for _, oe =  range s.Objects {
		if oe.name == name {
			return oe.obj
		}
	}
	var r *astObject
	return r
}

// --- parser ---
const O_READONLY int = 0
const FILE_SIZE int = 20000

func readFile(filename string) []uint8 {
	var fd int
	fd, _ = syscall.Open(filename, O_READONLY, 0)
	//fmtPrintf(itoa(fd))
	//fmtPrintf("\n")
	var buf []uint8 = make([]uint8, FILE_SIZE, FILE_SIZE)
	var n int
	n, _ = syscall.Read(fd, buf)
	//fmtPrintf(itoa(n))
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
	obj *astObject
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
	//fmtPrintf("parserNext\n")
	if ptok.tok == ";" {
		fmtPrintf("# [parser] pointing at : [%s] newline (%s)\n", ptok.tok , Itoa(scannerOffset))
	} else {
		fmtPrintf("# [parser] pointing at: [%s] %s (%s)\n", ptok.tok, ptok.lit, Itoa(scannerOffset))
	}

	if ptok.tok == "COMMENT" {
		for ptok.tok == "COMMENT" {
			parserConsumeComment()
		}
	}
	//fmtPrintf("current ptok: tok=%s, lit=%s\n", ptok.tok, ptok.lit)
}

func parserExpect(tok string, who string) {
	if ptok.tok != tok {
		var s = fmtSprintf("%s expected, but got %s", []string{tok, ptok.tok})
		panic2(who, s)
	}
	fmtPrintf("# [%s] consumed \"%s\"\n", who, ptok.tok)
	parserNext()
}

func parserExpectSemi(caller string) {
	//fmtPrintf("parserExpectSemi\n")
	if ptok.tok != ")" && ptok.tok != "}" {
		switch ptok.tok {
		case ";":
			fmtPrintf("# [%s] consumed semicolon %s\n", caller, ptok.tok)
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
		panic2(__func__, "IDENT expected, but got " +  ptok.tok)
	}
	fmtPrintf("# [%s] ident name = %s\n", __func__, name)
	var r = new(astIdent)
	r.Name = name
	return r
}

func parserParseImportDecl() *astImportSpec {
	//fmtPrintf("parserParseImportDecl\n")
	parserExpect("import", __func__)
	var path = ptok.lit
	parserNext()
	parserExpectSemi(__func__)
	var spec = new(astImportSpec)
	spec.Path = path
	return spec
}

func tryVarType() *astExpr {
	return tryIdentOrType()
}

func parseVarType() *astExpr {
	fmtPrintf("# [%s] begin\n", __func__)
	var e = tryIdentOrType()
	fmtPrintf("# [%s] end\n", __func__)
	return e
}

func tryType() *astExpr {
	fmtPrintf("# [%s] begin\n", __func__)
	var typ = tryIdentOrType()
	if typ != nil {
		parserResolve(typ)
	}
	fmtPrintf("# [%s] end\n", __func__)
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

	var varType = parseVarType()
	var typ = tryVarType()

	parserExpectSemi(__func__)

	var field = new(astField)
	field.Type = typ
	field.Name = varType.ident
	declareField(field, scope, "Var", varType.ident)
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

	var fields =  new(astFieldList)
	fields.List = list
	structType.Fields = fields
	var r = new(astExpr)
	r.dtype = "*astStructType"
	r.structType = structType
	return r
}

func parseTypeName() *astExpr {
	fmtPrintf("# [%s] begin\n", __func__)
	var ident = parseIdent()
	var typ = new(astExpr)
	typ.ident = ident
	typ.dtype = "*astIdent"
	fmtPrintf("# [%s] end\n", __func__)
	return typ
}

func tryIdentOrType() *astExpr {
	fmtPrintf("# [%s] begin\n", __func__)
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

func parseParameterList(scope *astScope) []*astField {
	fmtPrintf("# [%s] begin\n", __func__)
	var list []*astExpr
	for {
		var varType = parseVarType()
		list = append(list, varType)
		if ptok.tok != "," {
			break
		}
		parserNext()
		if ptok.tok == ")" {
			break
		}
	}
	fmtPrintf("# [%s] collected list n=%s\n", __func__, Itoa(len(list)))

	var params []*astField

	var typ = tryType()
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
			fmtPrintf("# [%s] ident.Name=%s\n", __func__, ident.Name)
			fmtPrintf("# [%s] typ=%s\n", __func__, typ.dtype)
			field.Name = ident
			field.Type = typ
			fmtPrintf("# [%s]: Field %s %s\n", __func__, field.Name.Name, field.Type.dtype)
			params = append(params, field)
			declareField(field, scope, "Var", ident)
			parserResolve(typ)
			if ptok.tok != "," {
				fmtPrintf("#  end %s\n", __func__)
				return params
			}
			parserNext()
			for ptok.tok != ")" && ptok.tok != "EOF" {
				ident = parseIdent()
				typ = parseVarType()
				field = new(astField)
				field.Name = ident
				field.Type = typ
				params = append(params, field)
				declareField(field, scope, "Var", ident)
				parserResolve(typ)
				if ptok.tok != "," {
					break
				}
				parserNext()
			}
		fmtPrintf("#  end %s\n", __func__)
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
	}
	fmtPrintf("#  end %s\n", __func__)
	return params
}

func parseParameters(scope *astScope) *astFieldList {
	fmtPrintf("# [%s] begin\n", __func__)
	var params []*astField
	parserExpect("(",  __func__)
	if ptok.tok != ")" {
		params = parseParameterList(scope)
	}
	parserExpect(")", __func__)
	var afl = new(astFieldList)
	afl.List = params
	fmtPrintf("# [%s] end\n", __func__)
	return afl
}

func parserResult(scope *astScope) *astFieldList {
	fmtPrintf("# [%s] begin\n", __func__)

	if ptok.tok == "(" {
		var r = parseParameters(scope)
		fmtPrintf("# [%s] end\n", __func__)
		return r
	}

	var r = new(astFieldList)
	if ptok.tok == "{" {
		r = nil
		fmtPrintf("# [%s] end\n", __func__)
		return r
	}
	var typ = tryType()
	var field = new(astField)
	field.Type = typ
	r.List = append(r.List, field)
	fmtPrintf("# [%s] end\n", __func__)
	return r
}

func parseSignature(scope *astScope) *signature {
	fmtPrintf("# [%s] begin\n", __func__)
	var params *astFieldList
	var results *astFieldList
	params = parseParameters(scope)
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
	fmtPrintf("# [declare] ident %s\n", ident.Name)

	var obj = new(astObject) //valSpec.Name.Obj
	obj.Decl = objDecl
	obj.Name = ident.Name
	obj.Kind = kind
	ident.Obj = obj

	// scope insert
	if ident.Name != "_" {
		scopeInsert(scope, obj)
	}
	fmtPrintf("# [declare] end\n")

}

func parserResolve(x *astExpr) {
	tryResolve(x, true)
}
func tryResolve(x *astExpr, collectUnresolved bool) {
	if x.dtype != "*astIdent" {
		return
	}
	var ident = x.ident
	if ident.Name == "-" {
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
		fmtPrintf("# appended unresolved ident %s\n", ident.Name)
	}
}

func parseOperand() *astExpr {
	switch ptok.tok {
	case "IDENT":
		var eIdent = new(astExpr)
		eIdent.dtype = "*astIdent"
		var ident  =  parseIdent()
		eIdent.ident = ident
		tryResolve(eIdent, true)
		return eIdent
	case "INT":
		var basicLit = new(astBasicLit)
		basicLit.Kind = "INT"
		basicLit.Value = ptok.lit
		var r = new(astExpr)
		r.dtype = "*astBasicLit"
		r.basicLit = basicLit
		parserNext()
		return r
	case "STRING":
		var basicLit = new(astBasicLit)
		basicLit.Kind = "STRING"
		basicLit.Value = ptok.lit
		var r = new(astExpr)
		r.dtype = "*astBasicLit"
		r.basicLit = basicLit
		parserNext()
		return r
	case "CHAR":
		var basicLit = new(astBasicLit)
		basicLit.Kind = "CHAR"
		basicLit.Value = ptok.lit
		var r = new(astExpr)
		r.dtype = "*astBasicLit"
		r.basicLit = basicLit
		parserNext()
		return r
	}

	var typ = tryIdentOrType()
	if typ == nil {
		panic2(__func__, "# typ should not be nil\n")
	}

	return typ
}

func parseCallExpr(fn *astExpr) *astExpr {
	parserExpect("(", __func__)
	var callExpr = new(astCallExpr)
	callExpr.Fun = fn
	fmtPrintf("# [parsePrimaryExpr] ptok.tok=%s\n", ptok.tok)
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

func parsePrimaryExpr() *astExpr {
	fmtPrintf("#   begin parsePrimaryExpr()\n")
	var x = parseOperand()

	for {
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
				fmtPrintf("# [parsePrimaryExpr] 741 ptok.tok=%s\n", ptok.tok)
			} else {
				fmtPrintf("#   end parsePrimaryExpr()\n")
				x = new(astExpr)
				x.dtype = "*astSelectorExpr"
				x.selectorExpr = sel
			}
		case "(":
			x = parseCallExpr(x)
		case "[":
			parserResolve(x)
			x = parseIndexOrSlice(x)
		default:
			fmtPrintf("#   end parsePrimaryExpr()\n")
			return x
		}
	}

	fmtPrintf("#   end parsePrimaryExpr()\n")
	return x
}

func parseIndexOrSlice(x *astExpr) *astExpr {
	parserExpect("[", __func__)
	var index []*astExpr = make([]*astExpr,3,3)
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
	fmtPrintf("#   begin parseUnaryExpr()\n")
	switch ptok.tok {
	case "+","-","!","&":
		var tok = ptok.tok
		parserNext()
		var x = parseUnaryExpr()
		r = new(astExpr)
		r.dtype = "*astUnaryExpr"
		r.unaryExpr = new(astUnaryExpr)
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
	r  = parsePrimaryExpr()
	fmtPrintf("#   end parseUnaryExpr()\n")
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
	fmtPrintf("#   begin parseBinaryExpr() prec1=%s\n", Itoa(prec1))
	var x = parseUnaryExpr()
	var oprec int
	for {
		var op = ptok.tok
		oprec  = precedence(op)
		fmtPrintf("# oprec %s\n", Itoa(oprec))
		fmtPrintf("# precedence \"%s\" %s < %s\n", op, Itoa(oprec) , Itoa(prec1))
		if oprec < prec1 {
			fmtPrintf("#   end parseBinaryExpr() (NonBinary)\n")
			return x
		}
		parserExpect(op, __func__)
		var y = parseBinaryExpr(oprec+1)
		var binaryExpr = new(astBinaryExpr)
		binaryExpr.X = x
		binaryExpr.Y = y
		binaryExpr.Op = op
		var r = new(astExpr)
		r.dtype = "*astBinaryExpr"
		r.binaryExpr = binaryExpr
		x = r
	}
	fmtPrintf("#   end parseBinaryExpr()\n")
	return x
}

func parseExpr() *astExpr {
	fmtPrintf("#   begin parseExpr()\n")
	var e = parseBinaryExpr(1)
	fmtPrintf("#   end parseExpr()\n")
	return e
}

func parseRhs() *astExpr {
	var x = parseExpr()
	return x
}

func makeExpr(s *astStmt) *astExpr {
	if s.dtype != "*astExprStmt" {
		panic2(__func__, "unexpected dtype=" + s.dtype)
	}
	if s.exprStmt == nil {
		panic2(__func__, "exprStmt is nil")
	}
	return s.exprStmt.X
}

func parseForStmt() *astStmt {
	fmtPrintf("# begin %s\n", __func__)
	parserExpect("for", __func__)
	openScope()

	var s1 *astStmt
	var s2 *astStmt
	var s3 *astStmt
	var isRange bool
	if ptok.tok != "{" {
		if ptok.tok != ";" {
			s2 = parseSimpleStmt(true)
			isRange = s2.isRange
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

	var body = parseBlockStmt()
	parserExpectSemi(__func__)
	var forStmt = new(astForStmt)
	forStmt.Init = s1
	forStmt.Cond = makeExpr(s2)
	forStmt.Post = s3
	forStmt.Body = body
	var r = new(astStmt)
	r.dtype = "*astForStmt"
	r.forStmt = forStmt
	closeScope()
	fmtPrintf("# end %s\n", __func__)
	return r
}

func parserIfStmt() *astStmt {
	parserExpect("if", __func__)
	var condStmt *astStmt = parseSimpleStmt(false)
	if condStmt.dtype != "*astExprStmt" {
		panic2(__func__, "unexpected dtype=" + condStmt.dtype)
	}
	var cond = condStmt.exprStmt.X
	var body = parseBlockStmt()
	var else_ *astStmt
	if ptok.tok == "else" {
		parserNext()
		var elseblock = parseBlockStmt()
		parserExpectSemi(__func__)
		else_ = new(astStmt)
		else_.dtype = "*astBlockStmt"
		else_.blockStmt = elseblock
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

func parseLhsList() []*astExpr {
	var x = parseExpr()
	var r []*astExpr
	r = append(r, x)
	return r
}

func parseSimpleStmt(isRangeOK bool) *astStmt {
	fmtPrintf("# begin %s\n", __func__)
	var s = new(astStmt)
	var x = parseLhsList()
	var stok = ptok.tok
	switch stok {
	case "=":
		parserNext() // consume =
		var y = parseExpr() // rhs
		var as = new(astAssignStmt)
		as.Tok = "="
		as.Lhs = x[0]
		as.Rhs = y
		s.dtype = "*astAssignStmt"
		s.assignStmt = as
		fmtPrintf("# end %s\n", __func__)
		return s
	case ";":
		s.dtype = "*astExprStmt"
		var exprStmt = new(astExprStmt)
		exprStmt.X = x[0]
		s.exprStmt = exprStmt
		fmtPrintf("# end %s\n", __func__)
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
	fmtPrintf("# end %s\n", __func__)
	return r
}

func parseStmt() *astStmt {
	fmtPrintf("# = begin %s\n", __func__)
	var s = new(astStmt)
	switch ptok.tok {
	case "var":
		var genDecl = parseDecl("var")
		s.dtype = "*astDeclStmt"
		s.DeclStmt = new(astDeclStmt)
		var decl = new(astDecl)
		decl.dtype = "*astGenDecl"
		decl.genDecl = genDecl
		s.DeclStmt.Decl = decl
		fmtPrintf("# = end parseStmt()\n")
		return s
	case "IDENT","*":
		var s = parseSimpleStmt(false)
		parserExpectSemi(__func__)
		return s
	case "return":
		var s = parseReturnStmt()
		return s
	case "if":
		var s = parserIfStmt()
		return s
	case "for":
		var s = parseForStmt()
		return s
	default:
		panic2(__func__, "TBI 3:" +  ptok.tok)
	}
	fmtPrintf("# = end parseStmt()\n")
	return s
}

func parseExprList() []*astExpr {
	var list []*astExpr
	var e = parseExpr()
	list = append(list, e)
	for ptok.tok == "," {
		parserNext() // consume ","
		e = parseExpr()
		list = append(list, e)
	}

	return list
}

func parseRhsList() []*astExpr {
	var list = parseExprList()
	return list
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
	for ptok.tok != "}" {
		if ptok.tok == "EOF" {
			panic2(__func__, "unexpected EOF\n")
		}
		var stmt = parseStmt()
		list = append(list, stmt)
	}
	return list
}

func parseBody(scope *astScope) *astBlockStmt {
	parserExpect("{", __func__)
	parserTopScope = scope
	fmtPrintf("# begin parseStmtList()\n")
	var list = parseStmtList()
	fmtPrintf("# end parseStmtList()\n")

	closeScope()
	parserExpect("}", __func__)
	var r = new(astBlockStmt)
	r.List = list
	return r
}

func parseBlockStmt() *astBlockStmt {
	parserExpect("{", __func__)
	openScope()
	fmtPrintf("# begin parseStmtList()\n")
	var list = parseStmtList()
	fmtPrintf("# end parseStmtList()\n")
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
		declare(objDecl, parserTopScope, "Var", ident)
		r = new(astGenDecl)
		r.Spec = spec
		return r
	default:
		panic2(__func__, "TBI\n")
	}
	return r
}

func parserParseTypeSpec() *astSpec {
	fmtPrintf("# [%s] start\n", __func__)
	parserExpect("type", __func__)
	var ident = parseIdent()
	fmtPrintf("# decl type %s\n", ident.Name)

	var spec = new(astTypeSpec)
	spec.Name = ident
	var objDecl = new(ObjDecl)
	objDecl.dtype = "*astTypeSpec"
	objDecl.typeSpec = spec
	declare(objDecl, parserTopScope, "Typ", ident)
	var typ = parseType()
	parserExpectSemi(__func__)
	spec.Type = typ
	var r = new(astSpec)
	r.dtype = "*astTypeSpec"
	r.typeSpec = spec
	return r
}

func parserParseValueSpec() *astSpec {
	fmtPrintf("# [parserParseValueSpec] start\n")
	parserExpect("var", __func__)
	var ident = parseIdent()
	fmtPrintf("# var = %s\n", ident.Name)
	var typ = parseType()
	parserExpectSemi(__func__)
	var spec = new(astValueSpec)
	spec.Name = ident
	spec.Type = typ
	var r = new(astSpec)
	r.dtype = "*astValueSpec"
	r.valueSpec = spec
	var objDecl = new(ObjDecl)
	objDecl.dtype = "*astValueSpec"
	objDecl.valueSpec = spec
	declare(objDecl, parserTopScope, "Var", ident)
	fmtPrintf("# [parserParseValueSpec] end\n")
	return r
}

func parserParseFuncDecl() *astDecl {
	parserExpect("func", __func__)
	var scope = astNewScope(parserTopScope) // function scope

	var ident = parseIdent()
	var sig = parseSignature(scope)
	if sig.results == nil {
		fmtPrintf("# [parserParseFuncDecl] %s sig.results is nil\n", ident.Name)
	} else {
		fmtPrintf("# [parserParseFuncDecl] %s sig.results.List = %s\n", ident.Name, Itoa(len(sig.results.List)))
	}
	var body *astBlockStmt
	if ptok.tok == "{" {
		fmtPrintf("# begin parseBody()\n")
		body = parseBody(scope)
		fmtPrintf("# end parseBody()\n")
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
	declare(objDecl, parserPkgScope, "Fun", ident)
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

	fmtPrintf("#\n")
	fmtPrintf("# [parser] Parsing Top level decls\n")
	var decls []*astDecl
	var decl *astDecl

	for ptok.tok != "EOF" {
		switch ptok.tok {
		case "var":
			var spec = parserParseValueSpec()
			fmtPrintf("# [parserParseFile] debug 1\n")
			var genDecl = new(astGenDecl)
			fmtPrintf("# [parserParseFile] debug 2\n")
			genDecl.Spec = spec
			fmtPrintf("# [parserParseFile] debug 3\n")
			decl = new(astDecl)
			decl.dtype = "*astGenDecl"
			decl.genDecl = genDecl
			fmtPrintf("# [parserParseFile] debug 9\n")
		case "func":
			fmtPrintf("\n\n")
			decl = parserParseFuncDecl()
			fmtPrintf("# func decl parsed:%s\n", decl.funcDecl.Name.Name)
		case "type":
			var spec = parserParseTypeSpec()
			var genDecl = new(astGenDecl)
			genDecl.Spec = spec
			decl = new(astDecl)
			decl.dtype = "*astGenDecl"
			decl.genDecl = genDecl
			fmtPrintf("# type parsed:%s\n", "")
		default:
			panic2(__func__, "TBI:" + ptok.tok)
		}
		decls = append(decls, decl)
	}

	parserTopScope = nil

	var unresolved []*astIdent
	var idnt *astIdent
	fmtPrintf("# [parserParseFile] resolving parserUnresolved (n=%s)\n", Itoa(len(parserUnresolved)))
	for _, idnt = range parserUnresolved {
		fmtPrintf("# [parserParseFile] resolving ident %s \n", idnt.Name)
		var obj *astObject = scopeLookup(parserPkgScope ,idnt.Name)
		if obj != nil {
			fmtPrintf("# [parserParseFile] obj matched %s \n", obj.Name)
			idnt.Obj = obj
		} else {
			unresolved = append(unresolved, idnt)
		}
	}
	fmtPrintf("# [parserParseFile] Unresolved (n=%s)\n", Itoa(len(unresolved)))

	var f = new(astFile)
	f.Name = packageName
	f.Decls = decls
	f.Unresolved = unresolved
	fmtPrintf("# [%s] end\n", __func__)
	return f
}

func parseFile(filename string) *astFile {
	var text = readSource(filename)
	parserInit(text)
	return parserParseFile()
}

// --- codegen ---
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
	fmtPrintf("  # Add const: %s\n", comment)
	fmtPrintf("  popq %%rax\n")
	fmtPrintf("  addq $%s, %%rax\n", Itoa(addValue))
	fmtPrintf("  pushq %%rax\n")
}

func emitLoad(t *Type) {
	if (t == nil) {
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

	default:
		panic2(__func__, "TBI:kind=" + kind(t))
	}
}

func emitVariableAddr(variable *Variable) {
	fmtPrintf("  # variable %s\n", variable.name)

	var addr string
	if variable.isGlobal {
		addr = fmtSprintf("%s(%%rip)", []string{variable.globalSymbol})
	} else {
		addr = fmtSprintf("%d(%%rbp)", []string{Itoa(variable.localOffset)})
	}

	fmtPrintf("  leaq %s, %%rax # addr\n", addr)
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
		panic2(__func__, "[emitListHeadAddr] kind=" + kind(getTypeOfExpr(list)))
	}
}

func emitAddr(expr *astExpr) {
	fmtPrintf("  # [emitAddr] %s\n", expr.dtype)
	switch expr.dtype {
	case "*astIdent":
		fmtPrintf("  # %s\n", expr.ident.Name)
		if expr.ident.Obj.Kind == "Var" {
			fmtPrintf("  # is Var\n")
			assert(expr.ident.Obj.Variable != nil,
				"ERROR: Variable is nil for name : " + expr.ident.Obj.Name, __func__)
			emitVariableAddr(expr.ident.Obj.Variable)
		} else {
			panic2(__func__, "Unexpected Kind " + expr.ident.Obj.Kind)
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
			panic2(__func__, "TBI:" + kind(typeOfX))
		}
		var field = lookupStructField(getStructTypeSpec(structType), expr.selectorExpr.Sel.Name)
		var offset = getStructFieldOffset(field)
		emitAddConst(offset, "struct head address + struct.field offset")
	default:
		panic2(__func__, "TBI " + expr.dtype)
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
			panic2(__func__, " unresolved ident:" + expr.ident.Name)
		}
		fmtPrintf("# [isType][DEBUG] expr.ident.Name = %s\n", expr.ident.Name)
		fmtPrintf("# [isType][DEBUG] expr.ident.Obj = %s,%s\n",
			expr.ident.Obj.Name, expr.ident.Obj.Kind )
		return expr.ident.Obj.Kind == "Typ"
	case "*astParenExpr":
		return isType(expr.parenExpr.X)
	case "*astStarExpr":
		return isType(expr.starExpr.X)
	default:
		fmtPrintf("# [isType][%s] is not considered a type\n", expr.dtype)
	}

	return false

}

func emitConversion(tp *Type, arg0 *astExpr) {
	fmtPrintf("# [emitConversion]\n")
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
			fmtPrintf("# [emitConversion] to int \n")
			emitExpr(arg0, nil)
		default:
			panic2(__func__, "[*astIdent] TBI : " + typeExpr.ident.Obj.Name)
		}
	case "*astArrayType": // Conversion to slice
		var arrayType = typeExpr.arrayType
		if arrayType.Len != nil {
			panic2(__func__, "internal error")
		}
		if (kind(getTypeOfExpr(arg0))) != T_STRING {
			panic2(__func__, "source type should be string")
		}
		fmtPrintf("  # Conversion of string => slice \n")
		emitExpr(arg0, nil)
		emitPopString()
		fmtPrintf("  pushq %%rcx # cap\n")
		fmtPrintf("  pushq %%rcx # len\n")
		fmtPrintf("  pushq %%rax # ptr\n")
	case "*astParenExpr":
		emitConversion(e2t(typeExpr.parenExpr.X), arg0)
	case "*astStarExpr": // (*T)(e)
		fmtPrintf("# [emitConversion] to pointer \n")
		emitExpr(arg0, nil)
	default:
		panic2(__func__, "TBI :" + typeExpr.dtype)
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
		panic2(__func__, "TBI:" + kind(t))
	}
}

func emitLen(arg *astExpr) {
	fmtPrintf("# [%s] begin\n", __func__)
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
	fmtPrintf("# [%s] end\n", __func__)
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

func emitInvertBoolValue() {
	emitPopBool("")
	fmtPrintf("  xor $1, %%rax\n")
	fmtPrintf("  pushq %%rax\n")
}

func emitTrue() {
	fmtPrintf("  pushq $1 # true\n")
}

func emitFalse() {
	fmtPrintf("  pushq $0 # true\n")
}

type Arg struct {
	e      *astExpr
	t      *Type // expected type
	offset int
}

func emitArgs(args []*Arg) int {
	var totalPushedSize int = 0
	//var arg *astExpr
	var i int
	for i=len(args)-1;i>=0;i-- {
		var arg = args[i]
		var t *Type
		if arg.t != nil {
			t = arg.t
		} else {
			t = getTypeOfExpr(arg.e)
		}
		var size = getSizeOfType(t)
		totalPushedSize = totalPushedSize + size
		emitExpr(arg.e, nil) // @TODO forceType should be fetched func decl
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
	fmtPrintf("# [%s] %s\n", __func__, symbol)
	var totalPushedSize = emitArgs(args)
	fmtPrintf("  callq %s\n", symbol)
	emitRevertStackPointer(totalPushedSize)
}

func emitFuncall(fun *astExpr, eArgs []*astExpr) {
	switch fun.dtype {
	case "*astIdent":
		fmtPrintf("# [%s][*astCallExpr][*astIdent]\n", __func__)
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
		}

		var ident = fun.ident
		if ident.Name == "print" {
			emitExpr(eArgs[0],nil)
			fmtPrintf("  callq runtime.printstring\n")
			fmtPrintf("  addq $%s, %%rsp # revert \n", Itoa(16))
			return
		}

		// general function call
		fmtPrintf("# [%s][*astCallExpr][default] start\n", __func__)

		var args []*Arg
		var _arg *Arg
		var i int
		for i=0;i<len(eArgs);i++ {
			_arg = new(Arg)
			_arg.e = eArgs[i]
			_arg.t = getTypeOfExpr(eArgs[i])
			args = append(args, _arg)
		}

		var symbol = pkgName + "." + ident.Name
		emitCall(symbol, args)

		// push results
		var obj = ident.Obj
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
		var results = fndecl.Sig.results
		if fndecl.Sig.results == nil {
			fmtPrintf("# [emitExpr] %s sig.results is nil\n", ident.Name)
		} else {
			fmtPrintf("# [emitExpr] %s sig.results.List = %s\n", ident.Name, Itoa(len(fndecl.Sig.results.List)))
		}

		if results != nil && len(results.List) == 1 {
			var retval0 = fndecl.Sig.results.List[0]
			var knd = kind(e2t(retval0.Type))
			switch knd {
			case T_STRING:
				fmtPrintf("  # fn.Obj=%s\n", obj.Name)
				fmtPrintf("  pushq %%rdi # str len\n")
				fmtPrintf("  pushq %%rax # str ptr\n")
			case T_BOOL, T_INT, T_UINTPTR, T_POINTER:
				fmtPrintf("  # fn.Obj=%s\n", obj.Name)
				fmtPrintf("  pushq %%rax\n")
			default:
				panic2(__func__, "Unexpected kind="+knd)
			}
		} else {
			fmtPrintf("   # No results\n")
		}
		return
	case "*astSelectorExpr":
		var selectorExpr = fun.selectorExpr
		if selectorExpr.X.dtype != "*astIdent" {
			panic2(__func__, "TBI selectorExpr.X.dtype=" + selectorExpr.X.dtype)
		}
		var symbol string = selectorExpr.X.ident.Name + "." + selectorExpr.Sel.Name
		switch symbol {
		case "os.Exit":
			emitCallNonDecl(symbol, eArgs)
		case "syscall.Write":
			emitCallNonDecl(symbol, eArgs)
		case "syscall.Syscall":
			fmtPrintf("   # calling syscall.Syscall\n")
			emitCallNonDecl(symbol, eArgs)
			fmtPrintf("  pushq %%rax # ret\n")
		case "unsafe.Pointer":
			emitExpr(eArgs[0], nil)
		default:
			fmtPrintf("  callq %s.%s\n", selectorExpr.X.ident.Name, selectorExpr.Sel.Name)
			panic2(__func__, "[*astSelectorExpr] Unsupported call to " + symbol)
		}
	case "*astParenExpr":
		panic2(__func__, "[emitExpr][*astCallExpr][*astParenExpr] TBI ")
	default:
		panic2(__func__, "[emitExpr][*astCallExpr] TBI fun.dtype=" + fun.dtype)
	}
}

func emitExpr(e *astExpr, forceType *Type) {
	fmtPrintf("# [emitExpr] dtype=%s\n", e.dtype)
	switch e.dtype {
	case "*astIdent":
		var ident = e.ident
		if ident.Obj == nil {
			panic2(__func__, "ident unresolved:" + ident.Name)
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
				panic2(__func__, "Unexpected kind=" +  kind(forceType))
			}
			return
		}
		switch ident.Obj.Kind {
		case "Var":
			emitAddr(e)
			var t = getTypeOfExpr(e)
			emitLoad(t)
		case "Typ":
			panic2(__func__, "[*astIdent] Kind Typ should not come here")
		default:
			panic2(__func__, "[*astIdent] unknown Kind=" + ident.Obj.Kind + " Name=" + ident.Obj.Name)
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
		fmtPrintf("# basicLit.Kind = %s \n", e.basicLit.Kind)
		switch e.basicLit.Kind {
		case "INT":
			fmtPrintf("  pushq $%s # \n", e.basicLit.Value)
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
			panic2(__func__, "[*astBasicLit] TBI : " +  e.basicLit.Kind)
		}
	case "*astCallExpr":
		var fun = e.callExpr.Fun
		fmtPrintf("# [%s][*astCallExpr]\n", __func__)
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
			panic2(__func__, "Unknown kind=" + kind(listType))
		}

		emitExpr(e.sliceExpr.Low, nil)
		var elmType = getElementTypeOfListType(listType)
		emitListElementAddr(list, elmType)
	case "*astUnaryExpr":
		switch e.unaryExpr.Op {
		case "-":
			emitExpr(e.unaryExpr.X, nil)
			fmtPrintf("  popq %%rax # e.X\n")
			fmtPrintf("  imulq $-1, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		case "&":
			emitAddr(e.unaryExpr.X)
		default:
			panic2(__func__, "TBI:astUnaryExpr:" + e.unaryExpr.Op)
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
		var t = getTypeOfExpr(e.binaryExpr.X)
		emitExpr(e.binaryExpr.X, nil) // left
		emitExpr(e.binaryExpr.Y, nil) // right
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
			panic2(__func__, "# TBI: binary operation for " + e.binaryExpr.Op)
		}
	default:
		panic2(__func__, "[emitExpr] `TBI:" + e.dtype)
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
	default:
		panic2(__func__, "Unexpected kind=" + kind(t))
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
	fmtPrintf("  # emitStore(%s)\n", kind(t))
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
	default:
		panic2(__func__, "TBI:" + kind(t))
	}
}

func emitAssign(lhs *astExpr, rhs *astExpr) {
	fmtPrintf("  # Assignment: emitAddr(lhs:%s)\n", lhs.dtype)
	emitAddr(lhs)
	fmtPrintf("  # Assignment: emitExpr(rhs)\n")
	emitExpr(rhs, getTypeOfExpr(lhs))
	emitStore(getTypeOfExpr(lhs))
}

func emitStmt(stmt *astStmt) {
	fmtPrintf("# [emitStmt] dtype=%s\n", stmt.dtype)
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
			fmtPrintf("  # lhs addresss\n")
			emitAddr(lhs)
			fmtPrintf("  # emitZeroValue for %s\n", t.e.dtype)
			emitZeroValue(t)
			fmtPrintf("  # Assignment: zero value\n")
			emitStore(t)
		} else {
			rhs = valSpec.Value
			emitAssign(lhs, rhs)
		}
		/*
		var valueSpec *astValueSpec = genDecl.Specs[0]
		var obj *astObject = valueSpec.Name.Obj
		var typ *astExpr = valueSpec.Type

		fmtPrintf("[emitStmt] TBI declSpec:%s\n", valueSpec.Name.Name)
		os.Exit(1)

		 */
	case "*astAssignStmt":
		switch stmt.assignStmt.Tok {
		case "=":
			var lhs = stmt.assignStmt.Lhs
			var rhs = stmt.assignStmt.Rhs
			emitAssign(lhs, rhs)
		default:
			panic2(__func__, "TBI: assignment of " + stmt.assignStmt.Tok)
		}
	case "*astReturnStmt":
		if len(stmt.returnStmt.Results) == 0 {
			fmtPrintf("  leave\n")
			fmtPrintf("  ret\n")
		} else if len(stmt.returnStmt.Results) == 1 {
			emitExpr(stmt.returnStmt.Results[0],nil) // @TODO forceType should be fetched from func decl
			var knd = kind(getTypeOfExpr(stmt.returnStmt.Results[0]))
			switch knd {
			case T_BOOL, T_INT, T_UINTPTR, T_POINTER:
				fmtPrintf("  popq %%rax # return 64bit\n")
			case T_STRING:
				fmtPrintf("  popq %%rax # return string (ptr)\n")
				fmtPrintf("  popq %%rdi # return string (len)\n")
			default:
				panic2(__func__, "[*astReturnStmt] TBI:" + knd)
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
		fmtPrintf("  # if\n")

		labelid++
		var labelEndif = ".L.endif." +  Itoa(labelid)
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
		fmtPrintf("  # end if\n")
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
	case "*astIncDecStmt":
		var addValue int
		switch stmt.incDecStmt.Tok {
		case "++":
			addValue = 1
		case "--":
			addValue = -1
		default:
			panic2(__func__, "Unexpected Tok=" + stmt.incDecStmt.Tok)
		}
		emitAddr(stmt.incDecStmt.X)
		emitExpr(stmt.incDecStmt.X, nil)
		emitAddConst(addValue, "rhs ++ or --")
		emitStore(getTypeOfExpr(stmt.incDecStmt.X))
	default:
		panic2(__func__, "TBI:" + stmt.dtype)
	}
}

func blockStmt2Stmt(block *astBlockStmt) *astStmt {
	var stmt = new(astStmt)
	stmt.dtype = "*astBlockStmt"
	stmt.blockStmt = block
	return stmt
}

var labelid int

func emitFuncDecl(pkgPrefix string, fnc *Func) {
	var localarea = fnc.localarea
	fmtPrintf("\n")
	var fname = fnc.name
	fmtPrintf(pkgPrefix + "." + fname + ":\n")

	fmtPrintf("  pushq %%rbp\n")
	fmtPrintf("  movq %%rsp, %%rbp\n")
	if localarea != 0 {
		fmtPrintf("  subq $%d, %%rsp # local area\n", Itoa(-localarea))
	}

	fmtPrintf("  # func body\n")
	emitStmt(blockStmt2Stmt(fnc.Body))

	fmtPrintf("  leave\n")
	fmtPrintf("  ret\n")
}

func emitGlobalVariable(name *astIdent, t *Type, val *astExpr) {
	var typeKind = kind(t)
	fmtPrintf("%s: # T %s \n", name.Name, typeKind)
	switch typeKind {
	case T_STRING:
		fmtPrintf("  .quad 0\n")
		fmtPrintf("  .quad 0\n")
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
			panic2(__func__, "Unexpected type:" + t.e.dtype)
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
		var v = basicLit.Value[0]
		var length = int(v - '0')
		fmtPrintf("# [emitGlobalVariable] array length uint8=%s\n" , Itoa(length))
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
			panic2(__func__, "Unexpected kind:" + kind)
		}

		fmtPrintf("# -- debug 8 \n")
		var i int
		for i = 0; i< length ; i++ {
			fmtPrintf(zeroValue)
		}
		fmtPrintf("# -- debug 9 \n")
	default:
		panic2(__func__, "TBI:kind=" + typeKind)
	}
}

func emitData(pkgName string) {
	fmtPrintf(".data\n")
	fmtPrintf("# string literals len = %s\n", Itoa(len(stringLiterals)))
	var con *stringLiteralsContainer
	for _, con = range stringLiterals {
		fmtPrintf("# string literals\n")
		fmtPrintf("%s:\n", con.sl.label)
		fmtPrintf("  .string %s\n", con.sl.value)
	}

	fmtPrintf("# ===== Global Variables =====\n")

	var spec *astValueSpec
	for _, spec = range globalVars {
		var t *Type
		if spec.Type != nil {
			t = e2t(spec.Type)
		}
		emitGlobalVariable(spec.Name, t, spec.Value)
	}

	fmtPrintf("# ==============================\n")
}

func emitText(pkgName string) {
	fmtPrintf(".text\n")
	var i int
	for i = 0; i < len(globalFuncs); i++ {
		var fnc = globalFuncs[i]
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

var T_STRING string
var T_SLICE string
var T_BOOL string
var T_INT string
var T_UINT8 string
var T_UINT16 string
var T_UINTPTR string
var T_ARRAY string
var T_STRUCT string
var T_POINTER string

var tInt *Type
var tUint8 *Type
var tUint16 *Type
var tUintptr *Type
var tString *Type
var tBool *Type

func getTypeOfExpr(expr *astExpr) *Type {
	fmtPrintf("# [%s] start\n", __func__)
	switch expr.dtype {
	case "*astIdent":
		switch expr.ident.Obj.Kind {
		case "Var":
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
		case "Con":
			switch expr.ident.Obj {
			case gTrue, gFalse:
				return tBool
			default:
				panic2(__func__, "1:Obj.Kind=" + expr.ident.Obj.Kind)
			}
		default:
			panic2(__func__, "2:Obj.Kind=" + expr.ident.Obj.Kind)
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
			panic2(__func__, "TBI:" + expr.basicLit.Kind)
		}
	case "*astIndexExpr":
		var list = expr.indexExpr.X
		return getElementTypeOfListType(getTypeOfExpr(list))
	case "*astUnaryExpr":
		switch expr.unaryExpr.Op {
		case "-":
			return getTypeOfExpr(expr.unaryExpr.X)
		default:
			panic2(__func__, "TBI: Op=" + expr.unaryExpr.Op)
		}
	case "*astCallExpr":
		fmtPrintf("# [%s] *astCallExpr\n", __func__)
		var fun = expr.callExpr.Fun
		switch fun.dtype {
		case "*astIdent":
			var fn = fun.ident
			if fn.Obj == nil {
				panic2(__func__, "[astCallExpr] nil Obj is not allowed")
			}
			switch fn.Obj.Kind {
			case "Typ":
				return e2t(fun)
			case "Fun":
				switch fn.Obj {
				case gLen, gCap:
					return tInt
				}
				var decl = fn.Obj.Decl
				if decl == nil {
					panic2(__func__, "decl is nil")
				}
				switch decl.dtype {
				case "*astFuncDecl":
					var resultList = decl.funcDecl.Sig.results.List
					if len(resultList) != 1 {
						panic2(__func__, "[astCallExpr] len results.List is not 1")
					}
					return e2t(decl.funcDecl.Sig.results.List[0].Type)
				default:
					panic2(__func__, "[astCallExpr] decl.dtype=" + decl.dtype)
				}
				panic2(__func__, "[astCallExpr] Fun ident " + fn.Name)
			}
		default:
			panic2(__func__, "[astCallExpr] dtype=" + expr.callExpr.Fun.dtype)
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
	default:
		panic2(__func__, "TBI:dtype=" +  expr.dtype)
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
				panic2(__func__, "unsupported decl :" + decl.dtype)
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
	default:
		panic2(__func__, "Unkown dtype:" + t.e.dtype)
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
		var arrayType =  t.e.arrayType
		if arrayType == nil {
			panic2(__func__, "should not be nil")
		}
		return e2t(arrayType.Elt)
	case T_STRING:
		return tUint8
	default:
		panic2(__func__, "TBI kind=" +  kind(t))
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
	case T_INT, T_UINTPTR, T_POINTER:
		return 8
	case T_UINT8:
		return 1
	case T_BOOL:
		return 8
	case T_STRUCT:
		return calcStructSizeAndSetFieldOffset(getStructTypeSpec(t))
	default:
		panic2(__func__, "TBI:" + knd)
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
		panic2(__func__ , "not T_STRUCT")
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
	value  string // raw value
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
	Body *astBlockStmt
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

func getStringLiteral(lit *astBasicLit) *sliteral {
	var container *stringLiteralsContainer
	for _, container = range stringLiterals {
		if container.lit == lit {
			return container.sl
		}
	}

	panic2(__func__, "string literal not found:"  + lit.Value)
	var r *sliteral
	return r
}

func registerStringLiteral(lit *astBasicLit) {
	fmtPrintf("# [registerStringLiteral] begin\n")

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
	fmtPrintf("# [registerStringLiteral] label=%s, strlen=%s\n", sl.label, Itoa(sl.strlen))
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
	fmtPrintf("# [%s] begin dtype=%s\n", __func__, stmt.dtype)
	switch stmt.dtype {
	case "*astDeclStmt":
		fmtPrintf("# [%s] *ast.DeclStmt\n", __func__)
		if stmt.DeclStmt == nil {
			panic2(__func__ , "nil pointer exception\n")
		}
		var declStmt = stmt.DeclStmt
		if declStmt.Decl == nil {
			panic2(__func__ , "ERROR\n")
		}
		var dcl = declStmt.Decl
		if dcl.dtype != "*astGenDecl" {
			panic2(__func__, "[dcl.dtype] internal error")
		}
		var valSpec = dcl.genDecl.Spec.valueSpec
		var typ = valSpec.Type // ident "int"
		fmtPrintf("# [walkStmt] valSpec Name=%s, Type=%s\n",
			valSpec.Name.Name, typ.dtype)
		var t = e2t(typ)
		var sizeOfType = getSizeOfType(t)
		localoffset = localoffset - sizeOfType

		valSpec.Name.Obj.Variable = newLocalVariable(valSpec.Name.Name, localoffset)
		fmtPrintf("# var %s offset = %d\n", valSpec.Name.Obj.Name,
			Itoa(valSpec.Name.Obj.Variable.localOffset))
		if valSpec.Value != nil {
			walkExpr(valSpec.Value)
		}
	case "*astAssignStmt":
		var rhs = stmt.assignStmt.Rhs
		walkExpr(rhs)
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
	case "*astIncDecStmt":
		walkExpr(stmt.incDecStmt.X)
	case "*astBlockStmt":
		var s *astStmt
		for _, s = range stmt.blockStmt.List {
			walkStmt(s)
		}
	default:
		panic2(__func__, "TBI: stmt.dtype=" + stmt.dtype)
	}
}

func walkExpr(expr *astExpr) {
	fmtPrintf("# [walkExpr] dtype=%s\n", expr.dtype)
	switch expr.dtype {
	case "*astIdent":
		// what to do ?
	case "*astCallExpr":
		var arg *astExpr
		walkExpr(expr.callExpr.Fun)
		for _, arg = range expr.callExpr.Args {
			walkExpr(arg)
		}
	case "*astBasicLit":
		switch expr.basicLit.Kind {
		case "STRING":
			registerStringLiteral(expr.basicLit)
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
		panic2(__func__, "TBI:" + expr.dtype)
	}
}

func walk(file *astFile) string {
	var decl *astDecl
	for _, decl = range file.Decls {
		switch decl.dtype {
		case "*astGenDecl":
			var genDecl = decl.genDecl
			if genDecl.Spec.dtype != "*astValueSpec" {
				continue
			}
			var valSpec = genDecl.Spec.valueSpec
			var nameIdent = valSpec.Name
			nameIdent.Obj.Variable = newGlobalVariable(nameIdent.Obj.Name)
			globalVars = append(globalVars, valSpec)
		case "*astFuncDecl":
			var funcDecl = decl.funcDecl
			fmtPrintf("# [sema] == astFuncDecl %s ==\n", funcDecl.Name.Name)
			localoffset  = 0
			var paramoffset = 16
			var field *astField
			for _, field = range funcDecl.Sig.params.List {
				var obj = field.Name.Obj
				obj.Variable = newLocalVariable(obj.Name, paramoffset)
				var varSize = getSizeOfType(e2t(field.Type))
				paramoffset = paramoffset + varSize
				fmtPrintf("# field.Name.Obj.Name=%s\n", obj.Name)
				//fmtPrintf("#   field.Type=%#v\n", field.Type)
			}
			var stmt *astStmt
			for _, stmt = range funcDecl.Body.List {
				walkStmt(stmt)
			}
			var fnc = new(Func)
			fnc.name = funcDecl.Name.Name
			fnc.Body = funcDecl.Body
			fnc.localarea = localoffset
			globalFuncs = append(globalFuncs, fnc)
		default:
			panic2(__func__, "TBI: " + decl.dtype)
		}
	}

	if len(stringLiterals) == 0 {
		panic2(__func__, "stringLiterals is empty\n")
	}

	return pkgName
}

// --- universe ---
var gNil *astObject
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
	scopeInsert(universe, gLen)
	scopeInsert(universe, gCap)

	fmtPrintf("# [%s] scope insertion of predefined identifiers complete\n", __func__)

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
	fmtPrintf("# [%s] scope insertion complete\n", __func__)
	return universe
}

func resolveUniverse(file *astFile, universe *astScope) {
	fmtPrintf("# [%s] start\n", __func__)
	// create universe scope
	// inject predeclared identifers
	var unresolved []*astIdent
	var ident *astIdent
	fmtPrintf("# [SEMA] resolving file.Unresolved (n=%s)\n", Itoa(len(file.Unresolved)))
	for _, ident = range file.Unresolved {
		fmtPrintf("# [SEMA] resolving ident %s ... ", ident.Name)
		var obj *astObject = scopeLookup(universe, ident.Name)
		if obj != nil {
			fmtPrintf(" matched\n")
			ident.Obj = obj
		} else {
			panic2(__func__, "Unresolved : "+ident.Name)
			unresolved = append(unresolved, ident)
		}
	}
}

func initGlobals() {
	__func__ = "__func__"
	T_STRING  = "T_STRING"
	T_SLICE  = "T_SLICE"
	T_BOOL  = "T_BOOL"
	T_INT  = "T_INT"
	T_UINT8  = "T_UINT8"
	T_UINT16  = "T_UINT16"
	T_UINTPTR  = "T_UINTPTR"
	T_ARRAY  = "T_ARRAY"
	T_STRUCT  = "T_STRUCT"
	T_POINTER  = "T_POINTER"

	gNil  = new(astObject)
	gNil.Kind = "Con" // is it Con ?
	gNil.Name = "nil"

	gTrue = new(astObject)
	gTrue.Kind = "Con"
	gTrue.Name = "true"

	gFalse = new(astObject)
	gFalse.Kind = "Con"
	gFalse.Name = "false"

	gString = new(astObject)
	gString.Kind = "Typ"
	gString.Name = "string"
	tString = new(Type)
	tString.e = new(astExpr)
	tString.e.dtype = "*astIdent"
	tString.e.ident = new(astIdent)
	tString.e.ident.Name = "string"
	tString.e.ident.Obj = gString

	gInt = new(astObject)
	gInt.Kind = "Typ"
	gInt.Name = "int"
	tInt = new(Type)
	tInt.e = new(astExpr)
	tInt.e.dtype = "*astIdent"
	tInt.e.ident = new(astIdent)
	tInt.e.ident.Name = "int"
	tInt.e.ident.Obj = gInt

	gUint8 = new(astObject)
	gUint8.Kind = "Typ"
	gUint8.Name = "uint8"
	tUint8 = new(Type)
	tUint8.e = new(astExpr)
	tUint8.e.dtype = "*astIdent"
	tUint8.e.ident = new(astIdent)
	tUint8.e.ident.Name = "uint8"
	tUint8.e.ident.Obj = gUint8

	gUint16 = new(astObject)
	gUint16.Kind = "Typ"
	gUint16.Name = "uint16"
	tUint16 = new(Type)
	tUint16.e = new(astExpr)
	tUint16.e.dtype = "*astIdent"
	tUint16.e.ident = new(astIdent)
	tUint16.e.ident.Name = "uint16"
	tUint16.e.ident.Obj = gUint16

	gUintptr = new(astObject)
	gUintptr.Kind = "Typ"
	gUintptr.Name = "uintptr"
	tUintptr = new(Type)
	tUintptr.e = new(astExpr)
	tUintptr.e.dtype = "*astIdent"
	tUintptr.e.ident = new(astIdent)
	tUintptr.e.ident.Name = "uintptr"
	tUintptr.e.ident.Obj = gUintptr

	gBool = new(astObject)
	gBool.Kind = "Typ"
	gBool.Name = "bool"
	tBool = new(Type)
	tBool.e = new(astExpr)
	tBool.e.dtype = "*astIdent"
	tBool.e.ident = new(astIdent)
	tBool.e.ident.Name = "bool"
	tBool.e.ident.Obj = gBool

	gNew = new(astObject)
	gNew.Kind = "Fun"
	gNew.Name = "new"

	gMake = new(astObject)
	gMake.Kind = "Fun"
	gMake.Name = "make"

	gLen = new(astObject)
	gLen.Kind = "Fun"
	gLen.Name = "len"

	gCap = new(astObject)
	gCap.Kind = "Fun"
	gCap.Name = "cap"
}

var pkgName string

func main() {
	initGlobals()

	var sourceFiles = []string{"2gen/runtime.go", "2gen/sample.go"}
	var sourceFile string

	var universe = createUniverse()

	for _, sourceFile = range sourceFiles {
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
