package main

import (
	"os"
)
import "syscall"

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
				var arg string = a[argIndex]
				argIndex++
				var s string = arg // // p.printArg(arg, c)
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
	var s string = fmtSprintf(format, a)
	syscall.Write(1, []uint8(s))
}


func Itoa(ival int) string {
	var __itoa_buf []uint8 = make([]uint8, 100,100)
	var __itoa_r []uint8 = make([]uint8, 100, 100)

	var next int
	var right int
	var ix int = 0
	if ival == 0 {
		return "0"
	}
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
	var offset int = scannerOffset
	for isLetter(scannerCh) || isDecimal(scannerCh) {
		scannerNext()
	}
	return string(scannerSrc[offset:scannerOffset])
}

func scannerScanNumber() string {
	var offset int = scannerOffset
	for isDecimal(scannerCh) {
		scannerNext()
	}
	return string(scannerSrc[offset:scannerOffset])
}

func scannerScanString() string {
	//fmtPrintf("begin: scannerScanString\n")
	var offset int = scannerOffset - 1
	for scannerCh != '"' {
		scannerNext()
	}
	scannerNext() // consume ending '""
	return string(scannerSrc[offset:scannerOffset])
}

func scannerScanChar() string {
	//fmtPrintf("begin: scannerScanString\n")
	var offset int = scannerOffset - 1
	for scannerCh != '\'' {
		//fmtPrintf("in loop char:%s\n", string([]uint8{scannerCh}))
		scannerNext()
	}
	scannerNext() // consume ending '""
	return string(scannerSrc[offset:scannerOffset])
}

func scannerrScanComment() string {
	var offset int = scannerOffset - 1
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
	var tc *TokenContainer
	tc = new(TokenContainer)
	var lit string
	var tok string
	var insertSemi bool
	var ch uint8 = scannerCh
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
			var peekCh uint8 = scannerSrc[scannerNextOffset]
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
			// @TODO block comment
			if scannerCh == '/' {
				lit = scannerrScanComment()
				tok = "COMMENT"
				scannerInsertSemi = false
				return scannerScan()
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
				var peekCh uint8 = scannerSrc[scannerNextOffset]
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
				var peekCh uint8 = scannerSrc[scannerNextOffset]
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
				var peekCh uint8 = scannerSrc[scannerNextOffset]
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
			fmtPrintf("unknown char:%s:%s\n", string([]uint8{ch}), Itoa(int(ch)))
			os.Exit(1)
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

type astDecl struct {
	Name *astIdent
	Sig  *signature
	Body *astBlockStmt
}

type astField struct {
	Name *astIdent
	Type *astExpr
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
}

type astDeclStmt struct {
	GenDecl *astGenDecl
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

type astGenDecl struct {
	Spec *astValueSpec
}

type ObjDecl struct {
	dtype string
	valueSpec *astValueSpec
	field *astField
}

type astValueSpec struct {
	Name *astIdent
	Type *astExpr
	Value *astExpr
}

type astExpr struct {
	dtype           string
	ident           *astIdent
	arrayType       *astArrayType
	basicLit        *astBasicLit
	callExpr        *astCallExpr
	binaryExpr      *astBinaryExpr
	unaryExpr       *astUnaryExpr
	selectorExpr *astSelectorExpr
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

// --- parser ---

const O_READONLY int = 0
const FILE_SIZE int = 20000

func readFile(filename string) []uint8 {
	var fd int
	fd, _ = syscall.Open(filename, O_READONLY, 0)
	//fmtPrintf(Itoa(fd))
	//fmtPrintf("\n")
	var buf []uint8 = make([]uint8, FILE_SIZE, FILE_SIZE)
	var n int
	n, _ = syscall.Read(fd, buf)
	//fmtPrintf(Itoa(n))
	var readbytes []uint8 = buf[0:n]
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

type astScope struct {
	Outer *astScope
	Objects []*objectEntry
}

var ptok *TokenContainer
var topScope *astScope

func parserNext() {
	//fmtPrintf("parserNext\n")
	ptok = scannerScan()
	if ptok.tok == ";" {
		fmtPrintf("# [parser] looking at : [%s] newline (%s)\n", ptok.tok , Itoa(scannerOffset))
	} else {
		fmtPrintf("# [parser] looking at: [%s] %s (%s)\n", ptok.tok, ptok.lit, Itoa(scannerOffset))
	}
	//fmtPrintf("current ptok: tok=%s, lit=%s\n", ptok.tok, ptok.lit)
}

func parserExpect(tok string, who string) {
	if ptok.tok != tok {
		fmtPrintf("# [%s] %s expected, but got %s\n", who, tok, ptok.tok)
		os.Exit(1)
	}
	fmtPrintf("# [%s] got expected tok %s\n", who, ptok.tok)
	parserNext()
}

func parserExpectSemi(caller string) {
	//fmtPrintf("parserExpectSemi\n")
	if ptok.tok != ")" && ptok.tok != "}" {
		switch ptok.tok {
		case ";":
			parserNext()
		default:
			fmtPrintf("[%s] ; expected, but got %s(%s)\n", caller, ptok.tok, ptok.lit)
			os.Exit(1)
		}
	}
}

func parseIdent() *astIdent {
	var name string
	if ptok.tok == "IDENT" {
		name = ptok.lit
		parserNext()
	} else {
		fmtPrintf("IDENT expected, but got %s\n", ptok.tok)
		os.Exit(1)
	}

	var r *astIdent = new(astIdent)
	r.Name = name
	return r
}

func parserParseImportDecl() *astImportSpec {
	//fmtPrintf("parserParseImportDecl\n")
	parserExpect("import", "parserParseImportDecl")
	var path string = ptok.lit
	parserNext()
	parserExpectSemi("parserParseImportDecl")
	var spec *astImportSpec = new(astImportSpec)
	spec.Path = path
	return spec
}

func parseVarType() *astExpr {
	var e *astExpr = tryIdentOrType()
	return e
}

func parseType() *astExpr {
	var typ *astExpr = tryIdentOrType()
	return typ
}

func eFromArrayType(t *astArrayType) *astExpr {
	var r *astExpr = new(astExpr)
	r.dtype = "*astArrayType"
	r.arrayType = t
	return r
}

func parseArrayType() *astExpr {
	parserExpect("[", "parseArrayType")
	if ptok.tok != "]" {
		fmtPrintf("[parseArrayType] TBI:%s\n",ptok.tok)
		os.Exit(1)
	}
	parserExpect("]", "parseArrayTypes")
	var elt *astExpr = parseType()
	var arrayType *astArrayType = new(astArrayType)
	arrayType.Elt = elt
	return eFromArrayType(arrayType)
}

func tryIdentOrType() *astExpr {
	var typ *astExpr = new(astExpr)
	fmtPrintf("# debug 1-1\n")
	switch ptok.tok {
	case "IDENT":
		var ident *astIdent = parseIdent()
		typ.ident = ident
		typ.dtype = "*astIdent"
	case "[":
		fmtPrintf("debug 1-2\n")
		typ = parseArrayType()
		fmtPrintf("debug 1-3\n")
		return typ
	default:
		fmtPrintf("TBI\n")
		os.Exit(1)
	}
	return typ
}


func parseParameterList() []*astField {
	var params []*astField
	//var list []*astExpr
	for ptok.tok != ")" {
		var field *astField = new(astField)
		var ident *astIdent = parseIdent()
//		fmtPrintf("debug 1\n")
		var typ *astExpr = parseVarType()
//		fmtPrintf("debug 2\n")
		field.Name = ident
		field.Type = typ
//		fmtPrintf("debug 3\n")
		fmtPrintf("# [parser] parseParameterList: Field %s %s\n", field.Name.Name, field.Type.dtype)
//		fmtPrintf("debug 4\n")
		params = append(params, field)

		declareField(field, topScope, "Var", ident)
		if ptok.tok == "," {
			parserNext()
			continue
		} else if ptok.tok == ")" {
			break
		} else {
			fmtPrintf("[parseParameterList] Syntax Error\n")
			os.Exit(1)
		}
	}

	return params
}

func parseParameters() *astFieldList {
	var params []*astField
	parserExpect("(", "parseParameters")
	if ptok.tok != ")" {
		params = parseParameterList()
	}
	parserExpect(")", "parseParameters")
	var afl *astFieldList = new(astFieldList)
	afl.List = params
	return afl
}

func parserResult() *astFieldList {
	var r *astFieldList = new(astFieldList)
	if ptok.tok == "{" {
		r = nil
		return r
	}
	var typ *astExpr = parseType()
	var field *astField = new(astField)
	field.Type = typ
	r.List = append(r.List, field)
	return r
}

func parseSignature() *signature {
	var params *astFieldList
	var results *astFieldList
	params = parseParameters()
	results = parserResult()
	var sig *signature = new(signature)
	sig.params = params
	sig.results = results
	return sig
}

func scopeInsert(s *astScope, obj *astObject) {
	var oe *objectEntry = new(objectEntry)
	oe.name = obj.Name
	oe.obj = obj
	topScope.Objects = append(topScope.Objects, oe)
}

func declareField(decl *astField, scope *astScope, kind string, ident *astIdent) {
	// delcare
	var obj *astObject = new(astObject)
	var objDecl *ObjDecl = new(ObjDecl)
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

func declare(decl *astValueSpec, scope *astScope, kind string, ident *astIdent) {
	// delcare
	var obj *astObject = new(astObject) //valSpec.Name.Obj
	var objDecl *ObjDecl = new(ObjDecl)
	objDecl.dtype = "*astValueSpec"
	objDecl.valueSpec = decl
	obj.Decl = objDecl
	obj.Name = ident.Name
	obj.Kind = kind
	ident.Obj = obj

	// scope insert
	if ident.Name != "_" {
		scopeInsert(scope, obj)
	}
}

func scopeLookup(s *astScope, name string) *astObject {
	var oe *objectEntry
	for _, oe =  range s.Objects {
		if oe.name == name {
			return oe.obj
		}
	}
	var r *astObject = nil
	return r
}

func parserResolve(ident *astIdent) {
	if ident.Name == "-" {
		return
	}

	if topScope != nil {
		var obj *astObject
		obj = scopeLookup(topScope, ident.Name)
		if obj != nil {
			ident.Obj = obj
		}

		return
	}
	fmtPrintf("cannot Resovle\n")
	os.Exit(1)
}

func parseOperand() *astExpr {
	switch ptok.tok {
	case "IDENT":
		var eIdent *astExpr = new(astExpr)
		eIdent.dtype = "*astIdent"
		var ident *astIdent =  parseIdent()
		eIdent.ident = ident
		parserResolve(ident)
		return eIdent
	case "INT":
		var basicLit *astBasicLit = new(astBasicLit)
		basicLit.Kind = "INT"
		basicLit.Value = ptok.lit
		var r *astExpr = new(astExpr)
		r.dtype = "*astBasicLit"
		r.basicLit = basicLit
		parserNext()
		return r
	case "STRING":
		var basicLit *astBasicLit = new(astBasicLit)
		basicLit.Kind = "STRING"
		basicLit.Value = ptok.lit
		var r *astExpr = new(astExpr)
		r.dtype = "*astBasicLit"
		r.basicLit = basicLit
		parserNext()
		return r
	}
	fmtPrintf("# [parsePrimaryExpr] TBI: ptok.tok=%s\n", ptok.tok)
	os.Exit(1)
	var r *astExpr
	return r
}

func parseCallExpr(fn *astExpr) *astCallExpr {
	parserExpect("(", "parseCallExpr")
	var callExpr *astCallExpr = new(astCallExpr)
	callExpr.Fun = fn
	fmtPrintf("# [parsePrimaryExpr] ptok.tok=%s\n", ptok.tok)
	var list []*astExpr
	for ptok.tok != ")" {
		var arg *astExpr = parseExpr()
		list = append(list, arg)
		if ptok.tok == "," {
			parserNext()
		} else if ptok.tok == ")" {
			break
		}
	}
	parserExpect(")", "parsePrimaryExpr") // consume ")"
	callExpr.Args = list
	return callExpr
}

func parsePrimaryExpr() *astExpr {
	fmtPrintf("#   begin parsePrimaryExpr()\n")
	var x *astExpr = parseOperand()
	var r *astExpr = new(astExpr)
	switch ptok.tok {
	case ".":
		parserNext() // consume "."
		if ptok.tok != "IDENT" {
			fmtPrintf("tok should be IDENT")
			os.Exit(1)
		}
		// Assume CallExpr
		var secondIdent *astIdent = parseIdent()
		if ptok.tok == "(" {
			var fn *astExpr = new(astExpr)
			fn.dtype = "*astSelectorExpr"
			fn.selectorExpr = new(astSelectorExpr)
			fn.selectorExpr.X = x
			fn.selectorExpr.Sel = secondIdent
			// string = x.ident.Name + "." + secondIdent
			r.dtype = "*astCallExpr"
			r.callExpr = parseCallExpr(fn)
			fmtPrintf("# [parsePrimaryExpr] 741 ptok.tok=%s\n", ptok.tok)
		} else {
			fmtPrintf("#   end parsePrimaryExpr()\n")
			return x
		}
	case "(":
		r.dtype = "*astCallExpr"
		r.callExpr = parseCallExpr(x)
		fmtPrintf("# [parsePrimaryExpr] 741 ptok.tok=%s\n", ptok.tok)
	default:
		fmtPrintf("#   end parsePrimaryExpr()\n")
		return x
	}
	fmtPrintf("#   end parsePrimaryExpr()\n")
	return r

}

func parseUnaryExpr() *astExpr {
	var r *astExpr
	fmtPrintf("#   begin parseUnaryExpr()\n")
	switch ptok.tok {
	case "+","-", "!":
		var tok string = ptok.tok
		parserNext()
		var x *astExpr = parseUnaryExpr()
		r = new(astExpr)
		r.dtype = "*astUnaryExpr"
		r.unaryExpr = new(astUnaryExpr)
		r.unaryExpr.Op = tok
		r.unaryExpr.X = x
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
	var x *astExpr = parseUnaryExpr()
	var oprec int
	for {
		var op string = ptok.tok
		oprec  = precedence(op)
		fmtPrintf("# oprec %s\n", Itoa(oprec))
		fmtPrintf("# precedence \"%s\" %s < %s\n", op, Itoa(oprec) , Itoa(prec1))
		if oprec < prec1 {
			fmtPrintf("#   end parseBinaryExpr() (NonBinary)\n")
			return x
		}
		parserExpect(op, "parseBinaryExpr")
		var y *astExpr = parseBinaryExpr(oprec+1)
		var binaryExpr *astBinaryExpr = new(astBinaryExpr)
		binaryExpr.X = x
		binaryExpr.Y = y
		binaryExpr.Op = op
		var r *astExpr = new(astExpr)
		r.dtype = "*astBinaryExpr"
		r.binaryExpr = binaryExpr
		x = r
	}
	fmtPrintf("#   end parseBinaryExpr()\n")
	return x
}

func parseExpr() *astExpr {
	fmtPrintf("#   begin parseExpr()\n")
	var e *astExpr = parseBinaryExpr(1)
	fmtPrintf("#   end parseExpr()\n")
	return e
}

func nop() {
}

func nop1() {
}

func parseStmt() *astStmt {
	fmtPrintf("# = begin parseStmt()\n")
	var s *astStmt
	s = new(astStmt)
	switch ptok.tok {
	case "var":
		var decl *astGenDecl = parseDecl("var")
		s.dtype = "*astDeclStmt"
		s.DeclStmt = new(astDeclStmt)
		s.DeclStmt.GenDecl = decl
		fmtPrintf("# = end parseStmt()\n")
		return s
	case "IDENT":
		fmtPrintf("# [parseStmt] is IDENT:%s\n",ptok.lit)
		var x *astExpr = parseExpr()
		var stok string = ptok.tok
		switch stok {
		case "=":
			parserNext() // consume =
			//fmtPrintf("# [parseStmt] ERROR:%s\n",stok)
			//os.Exit(1)
			var y *astExpr = parseExpr() // rhs
			var as *astAssignStmt = new(astAssignStmt)
			as.Tok = "="
			as.Lhs = x
			as.Rhs = y
			s.dtype = "*astAssignStmt"
			s.assignStmt = as
			parserExpectSemi("parseStmt:IDENT")
			fmtPrintf("# = end parseStmt()\n")
			return s
		case ";":
			s.dtype = "*astExprStmt"
			var exprStmt *astExprStmt = new(astExprStmt)
			exprStmt.X = x
			s.exprStmt = exprStmt
			parserExpectSemi("parseStmt:,")
			fmtPrintf("# = end parseStmt()\n")
			return s
		default:
			fmtPrintf("parseStmt:TBI 2:%s\n", ptok.tok)
			os.Exit(1)
		}
	default:
		fmtPrintf("parseStmt:TBI 3:%s\n", ptok.tok)
		os.Exit(1)
	}
	fmtPrintf("# = end parseStmt()\n")
	return s
}

func parseStmtList() []*astStmt {
	var list []*astStmt
	for ptok.tok != "}" {
		if ptok.tok == "EOF" {
			fmtPrintf("#[parseStmtList] unexpected EOF\n")
			os.Exit(1)
		}
		var stmt *astStmt = parseStmt()
		list = append(list, stmt)
	}
	return list
}

func parseBody() *astBlockStmt {
	parserExpect("{", "parseBody")

	var list  []*astStmt
	fmtPrintf("# begin parseStmtList()\n")
	list = parseStmtList()
	fmtPrintf("# end parseStmtList()\n")
	parserExpect("}", "parseBody")

	var r *astBlockStmt = new(astBlockStmt)
	r.List = list
	return r
}

func parseDecl(keyword string) *astGenDecl {
	var r *astGenDecl
	switch ptok.tok {
	case "var":
		parserExpect(keyword, "parseDecl:var:1")
		var ident *astIdent = parseIdent()
		var typ *astExpr= parseType()
		var value *astExpr
		if ptok.tok == "=" {
			parserNext()
			value = parseExpr()
		}
		parserExpectSemi("parseDecl:var:2")
		var spec *astValueSpec = new(astValueSpec)
		spec.Name = ident
		spec.Type = typ
		spec.Value = value
		declare(spec, topScope, "Var", ident)
		r = new(astGenDecl)
		r.Spec = spec
		return r
	default:
		fmtPrintf("[parseDecl] TBI\n")
		os.Exit(1)
	}
	return r
}

func parserParseFuncDecl() *astDecl {
	parserExpect("func", "parserParseFuncDecl")

	topScope = new(astScope)

	var ident *astIdent = parseIdent()
	var sig *signature = parseSignature()
	var body *astBlockStmt
	if ptok.tok == "{" {
		fmtPrintf("# begin parseBody()\n")
		body = parseBody()
		fmtPrintf("# end parseBody()\n")
		parserExpectSemi("parserParseFuncDecl")
	}
	var decl *astDecl = new(astDecl)
	decl.Name = ident
	decl.Sig = sig
	decl.Body = body
	topScope = nil
	return decl
}

func parserParseFile() *astFile {
	// expect "package" keyword
	parserExpect("package", "parserParseFile")

	var ident *astIdent = parseIdent()
	fmtPrintf("# parser: package name = %s\n", ident.Name)
	parserExpectSemi("parserParseFile")

	for ptok.tok == "import" {
		parserParseImportDecl()
	}

	var decls []*astDecl
	var decl *astDecl
	for ptok.tok != "EOF" {
		switch ptok.tok {
		case "func":
			decl  = parserParseFuncDecl()
			fmtPrintf("# func decl parsed:%s\n", decl.Name.Name)
		default:
			fmtPrintf("# parserParseFile:TBI:%s\n", ptok.tok)
			os.Exit(1)
		}
		decls = append(decls, decl)
	}

	var f *astFile = new(astFile)
	f.Name = ident.Name
	f.Decls = decls
	return f
}


func parseFile(filename string) *astFile {
	var text []uint8 = readSource(filename)
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

const sliceSize int = 24
const stringSize int = 16
const intSize int = 8
const ptrSize int = 8

func throw(s string) {
	syscall.Write(2, []uint8(s))
}

func kind(t *Type) string {
	if t == nil {
		fmtPrintf("nil type is not expected\n")
		os.Exit(1)
	}
	switch t.e.dtype {
	case "*astIdent":
		var ident *astIdent = t.e.ident
		switch ident.Name {
		case "int":
			return T_INT
		case "string":
			return T_STRING
		default:
			fmtPrintf("[kind] unsupported type %s\n", ident.Name)
			os.Exit(1)
		}
	default:
		fmtPrintf("error")
		os.Exit(1)
	}
	fmtPrintf("error")
	os.Exit(1)
	return ""
}

//type localoffsetint int //@TODO
var localoffset int = 0

type Variable struct {
	name         string
	isGlobal     bool
	globalSymbol string
	localOffset  int
}

func walkStmt(stmt *astStmt) {
	fmtPrintf("# [walkStmt] begin dtype=%s\n", stmt.dtype)
	switch stmt.dtype {
	case "*astDeclStmt":
		fmtPrintf("# [walkStmt] *ast.DeclStmt\n")
		var declStmt *astDeclStmt = stmt.DeclStmt
		if declStmt.GenDecl == nil {
			fmtPrintf("[walkStmt] ERROR\n")
			os.Exit(1)
		}
		var dcl *astGenDecl = declStmt.GenDecl
		var valSpec *astValueSpec = dcl.Spec
		var typ *astExpr = valSpec.Type // ident "int"
		fmtPrintf("# [walkStmt] valSpec Name=%s, Type=%s\n",
			valSpec.Name.Name, typ.dtype)
		var sizeOfType int = 8
		localoffset = localoffset - sizeOfType

		valSpec.Name.Obj.Variable = newLocalVariable(valSpec.Name.Name, localoffset)
		fmtPrintf("# var %s offset = %d\n", valSpec.Name.Obj.Name,
			Itoa(valSpec.Name.Obj.Variable.localOffset))
	case "*astAssignStmt":
		var rhs *astExpr = stmt.assignStmt.Rhs
		walkExpr(rhs)
	default:

	}
}

func panic(x string) {
	fmtPrintf(x+"\n")
	os.Exit(1)
}

var stringLiterals []*stringLiteralsContainer

type stringLiteralsContainer struct {
	lit *astBasicLit
	sl  *sliteral
}

var stringIndex int


func getStringLiteral(lit *astBasicLit) *sliteral {
	var container *stringLiteralsContainer
	for _, container = range stringLiterals {
		if container.lit == lit {
			return container.sl
		}
	}

	fmtPrintf("error: %s\n" , lit.Value)
	os.Exit(1)
	var r *sliteral
	return r
}

func registerStringLiteral(lit *astBasicLit) {
	var pkgName string = "main"
	if pkgName == "" {
		panic("no pkgName")
	}

	var strlen int
	var c uint8
	var vl []uint8 = []uint8(lit.Value)
	for _, c = range vl {
		if c != '\\' {
			strlen++
		}
	}

	var label string = fmtSprintf(".%s.S%d", []string{pkgName, Itoa(stringIndex)})
	stringIndex++

	var sl *sliteral = new(sliteral)
	sl.label = label
	sl.strlen = strlen - 2
	sl.value = lit.Value
	fmtPrintf("# [registerStringLiteral] label=%s, strlen=%s\n", sl.label, Itoa(sl.strlen))
	var cont *stringLiteralsContainer = new(stringLiteralsContainer)
	cont.sl = sl
	cont.lit = lit
	stringLiterals = append(stringLiterals, cont)
}

func walkExpr(expr *astExpr) {
	fmtPrintf("# [walkExpr] dtype=%s\n", expr.dtype)
	switch expr.dtype {
	case "*astBasicLit":
		switch expr.basicLit.Kind {
		case "STRING":
			registerStringLiteral(expr.basicLit)
		}
	default:
	}
}


func newLocalVariable(name string, localoffset int) *Variable {
	var vr *Variable = new(Variable)
	vr.name = name
	vr.isGlobal = false
	vr.localOffset = localoffset
	return vr
}

func semanticAnalyze(file *astFile) string {
	var funcDecl *astDecl

	for _, funcDecl = range file.Decls {
		fmtPrintf("# [sema] funcdef %s\n", funcDecl.Name.Name)
		localoffset  = 0
		var paramoffset int = 16
		var field *astField
		for _, field = range funcDecl.Sig.params.List {
			var obj *astObject = field.Name.Obj
			obj.Variable = newLocalVariable(obj.Name, paramoffset)
			paramoffset = paramoffset + 8 // @FIXME
			fmtPrintf("# field.Name.Obj.Name=%s\n", obj.Name)
			//fmtPrintf("#   field.Type=%#v\n", field.Type)
		}
		var stmt *astStmt
		for _, stmt = range funcDecl.Body.List {
			walkStmt(stmt)
		}
		var fnc *Func = new(Func)
		fnc.name = funcDecl.Name.Name
		fnc.Body = funcDecl.Body
		fnc.localarea = localoffset
		globalFuncs = append(globalFuncs, fnc)
	}
	return file.Name
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



func emitGlobalVariable(name string, t *Type, val string) {
	var typeKind string
	if t != nil {
		typeKind = T_INT
	}
	fmtPrintf("%s: # T %s \n", name, typeKind)
	switch typeKind {
	case T_STRING:
		fmtPrintf("  .quad 0\n")
		fmtPrintf("  .quad 0\n")
	case T_INT:
		fmtPrintf("  .quad %s\n", val)
	default:
		fmtPrintf("ERROR\n")
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

	/*
	var spec *astValueSpec
	for _, spec = range globalVars {
		//emitGlobalVariable(spec.Name.Name, spec.Type, spec.Value)
		//emitGlobalVariable(spec.Name.Name, nil, "")
	}
*/
	fmtPrintf("# ==============================\n")
}

func emitExpr(e *astExpr) {
	switch e.dtype {
	case "*astIdent":
		var ident *astIdent = e.ident
		if ident.Obj == nil {
			fmtPrintf("[emitExpr] ident %s is unresolved\n", ident.Name)
			os.Exit(1)
		}
		switch ident.Obj.Kind {
		case "Var":
			emitAddr(e)
			var t *Type = getTypeOfExpr(e)
			emitLoad(t)
		default:
			fmtPrintf("unknown Kind=%s\n", ident.Obj.Kind)
			os.Exit(1)
		}
	case "*astBasicLit":
		fmtPrintf("# basicLit.Kind = %s \n", e.basicLit.Kind)
		switch e.basicLit.Kind {
		case "INT":
			fmtPrintf("  pushq $%s # \n", e.basicLit.Value)
		case "STRING":
			var sl *sliteral = getStringLiteral(e.basicLit)
			if sl.strlen == 0 {
				// zero value
				//emitZeroValue(tString)
				fmtPrintf("[emitExpr] TBI: empty string literal\n")
				os.Exit(1)
			} else {
				fmtPrintf("  pushq $%d # str len\n", Itoa(sl.strlen))
				fmtPrintf("  leaq %s, %%rax # str ptr\n", sl.label)
				fmtPrintf("  pushq %%rax # str ptr\n")
			}
		default:
			fmtPrintf("[emitExpr] TBI\n")
			os.Exit(1)
		}
	case "*astCallExpr":
		var fun *astExpr = e.callExpr.Fun
		emitExpr(e.callExpr.Args[0])
		switch fun.dtype {
		case "*astSelectorExpr":
			var selector *astSelectorExpr = fun.selectorExpr
			if selector.X.dtype != "*astIdent" {
				fmtPrintf("[emitExpr] TBI selector.X.dtype=%s\n", selector.X.dtype)
				os.Exit(1)
			}
			fmtPrintf("  callq %s.%s\n", selector.X.ident.Name, selector.Sel.Name)
		case "*astIdent":
			var ident *astIdent = fun.ident
			if ident.Name == "print" {
				fmtPrintf("  callq runtime.printstring\n")
			} else {
				fmtPrintf("  callq main.%s\n", ident.Name)
			}
		default:
			fmtPrintf("[emitExpr] TBI fun.dtype=%s\n", fun.dtype)
			os.Exit(1)
		}
	case "*astUnaryExpr":
		switch e.unaryExpr.Op {
		case "-":
			emitExpr(e.unaryExpr.X)
			fmtPrintf("  popq %%rax # e.X\n")
			fmtPrintf("  imulq $-1, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		default:
			fmtPrintf("[emitExpr] `TBI:astUnaryExpr:%s\n", e.dtype)
			os.Exit(1)
		}
	case "*astBinaryExpr":
		emitExpr(e.binaryExpr.X) // left
		emitExpr(e.binaryExpr.Y) // right
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
		default:
			fmtPrintf("# TBI: binary operation for '%s'", e.binaryExpr.Op)
			os.Exit(1)
		}
	default:
		fmtPrintf("[emitExpr] `TBI:%s\n", e.dtype)
		os.Exit(1)
	}
}


func emitVariableAddr(variable *Variable) {
	fmtPrintf("  # variable %#v\n", variable.name)

	var addr string
	if variable.isGlobal {
		addr = fmtSprintf("%s(%%rip)", []string{variable.globalSymbol})
	} else {
		addr = fmtSprintf("%d(%%rbp)", []string{Itoa(variable.localOffset)})
	}

	fmtPrintf("  leaq %s, %%rax # addr\n", addr)
	fmtPrintf("  pushq %%rax\n")
}

func emitAddr(expr *astExpr) {
	fmtPrintf("  # emitAddr %s\n", expr.dtype)
	switch expr.dtype {
	case "*astIdent":
		fmtPrintf("  # %s\n", expr.ident.Name)
		if expr.ident.Obj.Kind == "Var" {
			fmtPrintf("  # is Var\n")
			if expr.ident.Obj.Variable == nil {
				fmtPrintf("ERROR: Variable is nil for name %s\n", expr.ident.Obj.Name)
				os.Exit(1)
			}
			emitVariableAddr(expr.ident.Obj.Variable)
		} else {
			fmtPrintf("[emitAddr] Unexpected Kind %s \n" ,expr.ident.Obj.Kind)
		}
	default:
		fmtPrintf("[emitAddr] TBI %s \n" ,expr.dtype)
		os.Exit(1)
	}
}

func emitLoad(t *Type) {
	if (t == nil) {
		fmtPrintf("[emitLoad] nil type error\n")
		os.Exit(1)
	}
	emitPopAddress(kind(t))
	switch kind(t) {
	case T_INT:
		fmtPrintf("  movq %d(%%rax), %%rax # load int\n", Itoa(0))
		fmtPrintf("  pushq %%rax\n")
	case T_STRING:
		fmtPrintf("  movq %d(%%rax), %%rdx\n", Itoa(8))
		fmtPrintf("  movq %d(%%rax), %%rax\n", Itoa(0))
		fmtPrintf("  pushq %%rdx # len\n")
		fmtPrintf("  pushq %%rax # ptr\n")

	default:
		fmtPrintf("[emitLoad] TBI\n")
		os.Exit(1)
	}
}

func emitStore(t *Type) {
	fmtPrintf("  # emitStore(%s)\n", kind(t))
	switch kind(t) {
	case T_STRING:
		emitPopString()
		fmtPrintf("  popq %%rsi # lhs ptr addr\n")
		fmtPrintf("  movq %%rax, %d(%%rsi) # ptr to ptr\n", Itoa(0))
		fmtPrintf("  movq %%rcx, %d(%%rsi) # len to len\n", Itoa(8))
	case T_INT:
		fmtPrintf("  popq %%rdi # rhs evaluated\n")
		fmtPrintf("  popq %%rax # lhs addr\n")
		fmtPrintf("  movq %%rdi, (%%rax) # assign\n")
	default:
		fmtPrintf("TBI:" + kind(t))
		os.Exit(1)
	}
}

func emitAssign(lhs *astExpr, rhs *astExpr) {
	fmtPrintf("  # Assignment: emitAddr(lhs)\n")
	emitAddr(lhs)
	fmtPrintf("  # Assignment: emitExpr(rhs)\n")
	emitExpr(rhs)
	emitStore(getTypeOfExpr(lhs))
}


func emitStmt(stmt *astStmt) {
	switch stmt.dtype {
	case "*astBlockStmt":
		var stmt2 *astStmt
		for _, stmt2 = range stmt.blockStmt.List {
			emitStmt(stmt2)
		}
	case "*astExprStmt":
		emitExpr(stmt.exprStmt.X)
	case "*astDeclStmt":
		/*
		var genDecl *astGenDecl = stmt.DeclStmt.GenDecl
		var valueSpec *astValueSpec = genDecl.Specs[0]
		var obj *astObject = valueSpec.Name.Obj
		var typ *astExpr = valueSpec.Type

		fmtPrintf("[emitStmt] TBI declSpec:%s\n", valueSpec.Name.Name)
		os.Exit(1)

		 */
	case "*astAssignStmt":
		switch stmt.assignStmt.Tok {
		case "=":
			var lhs *astExpr = stmt.assignStmt.Lhs
			var rhs *astExpr = stmt.assignStmt.Rhs
			emitAssign(lhs, rhs)
		default:
			fmtPrintf("TBI: assignment of " + stmt.assignStmt.Tok)
			os.Exit(1)
		}
	default:
		fmtPrintf("[emitStmt] TBI:%s\n", stmt.dtype)
		os.Exit(1)
	}
}

func emitFuncDecl(pkgPrefix string, fnc *Func) {
	var localarea int = fnc.localarea
	fmtPrintf("\n")
	var fname string = fnc.name
	fmtPrintf(pkgPrefix + "." + fname + ":\n")

	fmtPrintf("  pushq %%rbp\n")
	fmtPrintf("  movq %%rsp, %%rbp\n")
	if localarea != 0 {
		fmtPrintf("  subq $%d, %%rsp # local area\n", Itoa(-localarea))
	}

	fmtPrintf("  # func body\n")
	var stmt *astStmt = new(astStmt)
	stmt.dtype = "*astBlockStmt"
	stmt.blockStmt = fnc.Body
	emitStmt(stmt)

	fmtPrintf("  leave\n")
	fmtPrintf("  ret\n")
}

func emitText(pkgName string) {
	fmtPrintf(".text\n")
	var i int
	for i = 0; i < len(globalFuncs); i++ {
		var fnc *Func = globalFuncs[i]
		emitFuncDecl(pkgName, fnc)
	}
}

type Type struct {
	//kind string
	e *astExpr
}

func getTypeOfExpr(expr *astExpr) *Type {
	switch expr.dtype {
	case "*astIdent":
		switch expr.ident.Obj.Kind {
		case "Var":
			switch expr.ident.Obj.Decl.dtype {
			case "*astValueSpec":
				var decl *astValueSpec = expr.ident.Obj.Decl.valueSpec
				var t *Type = new(Type)
				t.e = decl.Type
				return t
			case "*astField":
				var decl *astField = expr.ident.Obj.Decl.field
				var t *Type = new(Type)
				t.e = decl.Type
				return t
			default:
				fmtPrintf("[getTypeOfExpr] ERROR 0\n")
				os.Exit(1)

			}
		default:
			fmtPrintf("[getTypeOfExpr] ERROR 1\n")
			os.Exit(1)
		}
		default:
		fmtPrintf("[getTypeOfExpr] ERROR 2\n")
		os.Exit(1)
	}
	fmtPrintf("[getTypeOfExpr] nil type is not allowed\n")
	os.Exit(1)
	var r *Type
	return r
}
func generateCode(pkgName string) {
	fmtPrintf("")
	fmtPrintf("runtime.heapInit:\n")
	fmtPrintf("ret\n")

	emitData(pkgName)
	emitText(pkgName)
}

type Func struct {
	localvars []*string
	localarea int
	argsarea  int
	name      string
	Body *astBlockStmt
}

type sliteral struct {
	label  string
	strlen int
	value  string // raw value
}

var globalVars []*astValueSpec
var globalFuncs []*Func

var _garbage string

type astFile struct {
	Name string
	Decls []*astDecl
}

func main() {
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
	
	var sourceFiles []string = []string{"2gen/sample.go"}
	var sourceFile string
	for _, sourceFile = range sourceFiles {
		globalVars = nil
		globalFuncs = nil
		stringLiterals = nil
		stringIndex = 0
		var f *astFile = parseFile(sourceFile)
		semanticAnalyze(f)
		generateCode(f.Name)
	}

}
