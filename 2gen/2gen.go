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

var __itoa_buf [100]uint8
var __itoa_r [100]uint8

func Itoa(ival int) string {
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
	Decl    *astDecl
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
	Rhs *astExpr
}

type astGenDecl struct {
	Specs []*astValueSpec
}

type astValueSpec struct {
	Name *astIdent
	Type *astExpr
	Value *astExpr
}

type astExpr struct {
	dtype      string
	ident      *astIdent
	arrayType  *astArrayType
	basicLit   *astBasicLit
	callExpr   *astCallExpr
	binaryExpr *astBinaryExpr
}

type astIdent struct {
	Name string
}

type astArrayType struct {
	Len *astExpr
	Elt *astExpr
}

type astBasicLit struct {
	Kind     string   // token.INT, token.CHAR, or token.STRING
	Value    string
}

type astCallExpr struct {
	Fun      string      // function expression
	Args     []*astExpr    // function arguments; or nil
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

var ptok *TokenContainer

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

func parserExpectSemi() {
	//fmtPrintf("parserExpectSemi\n")
	if ptok.tok != ")" && ptok.tok != "}" {
		switch ptok.tok {
		case ";":
			parserNext()
		default:
			fmtPrintf("; expected, but got %s(%s)\n", ptok.tok, ptok.lit)
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
	parserExpectSemi()
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
		fmtPrintf("[parser] parseParameterList: Field %s %s\n", field.Name.Name, field.Type.dtype)
//		fmtPrintf("debug 4\n")
		params = append(params, field)
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

func parsePrimaryExpr() *astExpr {
	fmtPrintf("#   begin parsePrimaryExpr()\n")
	var r *astExpr = new(astExpr)
	switch ptok.tok {
	case "IDENT":
		var identToken *TokenContainer = ptok
		var eIdent *astExpr = new(astExpr)
		eIdent.dtype = "*astIdent"
		eIdent.ident = new(astIdent)
		eIdent.ident.Name = identToken.lit
		parserNext() // consume IDENT
		if ptok.tok == "." {
			parserNext() // consume "."
			if ptok.tok != "IDENT" {
				fmtPrintf("tok should be IDENT")
				os.Exit(1)
			}
			var secondIdent string = ptok.lit
			parserNext() // consume IDENT
			var callExpr *astCallExpr = new(astCallExpr)
			callExpr.Fun = eIdent.ident.Name + "." + secondIdent
			fmtPrintf("# [parsePrimaryExpr] ptok.tok=%s\n", ptok.tok)
			parserNext() // consume "("
			var arg *astExpr= parseExpr()
			fmtPrintf("#   debug\n")
			parserExpect(")", "parsePrimaryExpr") // consume ")"
			callExpr.Args = append(callExpr.Args, arg)
			r.dtype = "*astCallExpr"
			r.callExpr = callExpr
			fmtPrintf("# [parsePrimaryExpr] 741 ptok.tok=%s\n", ptok.tok)
		} else if ptok.tok == "(" {
			parserNext() // consume "("
			var callExpr *astCallExpr = new(astCallExpr)
			callExpr.Fun = identToken.lit
			var arg *astExpr = parseExpr()
			parserNext() // consume ")"
			callExpr.Args = append(callExpr.Args, arg)
			r.dtype = "*astCallExpr"
			r.callExpr = callExpr
		} else  {
			fmtPrintf("ERROR: ptok.tok=%s\n", ptok.tok)
			os.Exit(1)
			parserNext() // consume ")"
			fmtPrintf("#   end parsePrimaryExpr()\n")
			return eIdent
		}
	case "INT":
		var basicLit *astBasicLit = new(astBasicLit)
		basicLit.Kind = "INT"
		basicLit.Value = ptok.lit
		r.dtype = "*astBasicLit"
		r.basicLit = basicLit
		parserNext()
	default:
		fmtPrintf("# [parsePrimaryExpr] TBI: ptok.tok=%s\n", ptok.tok)
		os.Exit(1)
	}
	fmtPrintf("#   end parsePrimaryExpr()\n")
	return r

}

func parseUnaryExpr() *astExpr {
	fmtPrintf("#   begin parseUnaryExpr()\n")
	var r *astExpr = parsePrimaryExpr()
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
		s.dtype = "*astGenDecl"
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
			fmtPrintf("# [parseStmt] ERROR:%s\n",stok)
			os.Exit(1)
			parserNext() // consume =
			var y *astExpr = parseExpr()
			var as *astAssignStmt = new(astAssignStmt)
			as.Lhs = x
			as.Rhs = y
			s.dtype = "*astAssignStmt"
			s.assignStmt = as
			parserExpectSemi()
			fmtPrintf("# = end parseStmt()\n")
			return s
		case ";":
			s.dtype = "*astExprStmt"
			var exprStmt *astExprStmt = new(astExprStmt)
			exprStmt.X = x
			s.exprStmt = exprStmt
			parserExpectSemi()
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

/*
func parseVarDecl() {

}

func parseRhs() *astExpr {
	return nil
}

func parserVarDecl() *astStmt {
	return nil
}
*/

func parseDecl(keyword string) *astGenDecl {
	var r *astGenDecl
	switch ptok.tok {
	case "var":
		parserExpect(keyword, "parseDecl")
		var ident *astIdent = parseIdent()
		var typ *astExpr= parseType()
		var value *astExpr
		if ptok.tok == "=" {
			parserNext()
			//value =parseRhs()
			value = nil
		}
		parserExpectSemi()

		var spec *astValueSpec = new(astValueSpec)
		spec.Name = ident
		spec.Type = typ
		spec.Value = value
		r = new(astGenDecl)
		r.Specs = append(r.Specs, spec)
		return r
	default:
		fmtPrintf("[parseDecl] TBI\n")
		os.Exit(1)
	}
	return r
}

func parserParseFuncDecl() *astDecl {
	parserExpect("func", "parserParseFuncDecl")
	var ident *astIdent = parseIdent()
	var sig *signature = parseSignature()
	var body *astBlockStmt
	if ptok.tok == "{" {
		fmtPrintf("# begin parseBody()\n")
		body = parseBody()
		fmtPrintf("# end parseBody()\n")
		parserExpectSemi()
	}
	var decl *astDecl = new(astDecl)
	decl.Name = ident
	decl.Sig = sig
	decl.Body = body
	return decl
}

func parserParseFile() *astFile {
	// expect "package" keyword
	parserExpect("package", "parserParseFile")

	var ident *astIdent = parseIdent()
	fmtPrintf("# parser: package name = %s\n", ident.Name)
	parserExpectSemi()

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

type Type struct {
	kind string
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
	return t.kind
}

//type localoffsetint int //@TODO

func semanticAnalyze(file *astFile) string {
	var decl *astDecl

	for _, decl = range file.Decls {
		fmtPrintf("# [sema] decl func = %s\n", decl.Name.Name)
		var fnc *Func = new(Func)
		fnc.name = decl.Name.Name
		fnc.Body = decl.Body
		globalFuncs = append(globalFuncs, fnc)
	}
	return fakeSemanticAnalyze(file)
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

func emitGlobalVariable(name string, t *Type, val string) {
	var typeKind string
	if t != nil {
		typeKind = t.kind
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
	var sl *sliteral
	for _, sl = range stringLiterals {
		fmtPrintf("# string literals\n")
		fmtPrintf("%s:\n", sl.label)
		fmtPrintf("  .string \"%s\"\n", sl.value)
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
	case "*astBasicLit":
		fmtPrintf("  pushq $%s # status\n", e.basicLit.Value)
	case "*astCallExpr":
		var callExpr *astCallExpr = e.callExpr
		emitExpr(callExpr.Args[0])
		fmtPrintf("  callq %s\n", callExpr.Fun)
	case "*astBinaryExpr":
		emitExpr(e.binaryExpr.X) // left
		emitExpr(e.binaryExpr.Y) // right
		switch e.binaryExpr.Op {
		case "+":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  addq %%rcx, %%rax\n")
			fmtPrintf("  pushq %%rax\n")
		case "*":
			fmtPrintf("  popq %%rcx # right\n")
			fmtPrintf("  popq %%rax # left\n")
			fmtPrintf("  imulq %%rcx, %%rax\n")
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

func emitStmt(stmt *astStmt) {
	switch stmt.dtype {
	case "*astBlockStmt":
		var stmt2 *astStmt
		for _, stmt2 = range stmt.blockStmt.List {
			emitStmt(stmt2)
		}
	case "*astExprStmt":
		emitExpr(stmt.exprStmt.X)
	default:
		fmtPrintf("[emitStmt] TBI:%s\n", stmt.dtype)
		os.Exit(1)
	}
}

func emitFuncDecl(pkgPrefix string, fnc *Func) {
	fmtPrintf("\n")
	var fname string = fnc.name
	fmtPrintf(pkgPrefix + "." + fname + ":\n")
	if len(fnc.localvars) > 0 {
		var slocalarea string = Itoa(fnc.localarea)
		fmtPrintf("  subq $" + slocalarea + ", %rsp # local area\n")
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

var stringLiterals []*sliteral
var stringIndex int
var globalVars []*astValueSpec
var globalFuncs []*Func

var _garbage string

type astFile struct {
	Name string
	Decls []*astDecl
}

func main() {
	var sourceFiles []string = []string{"2gen/sample.go"}
	var sourceFile string
	for _, sourceFile = range sourceFiles {
		globalVars = nil
		globalFuncs = nil
		stringLiterals = nil
		stringIndex = 0
		var f *astFile = parseFile(sourceFile)
		var pkgName string = semanticAnalyze(f)
		generateCode(pkgName)
	}

	//test()
}

func test() {
	// Test funcs
	emitPopBool("comment")
	emitPopAddress("comment")
	emitPopString()
	emitPopSlice()
	var t1 *Type = new(Type)
	t1.kind = T_INT
	emitPushStackTop(t1, "comment")
	emitRevertStackPointer(24)
	emitAddConst(42, "comment")
	getPushSizeOfType(t1)
}

func fakeSemanticAnalyze(file *astFile) string {

	stringLiterals = make([]*sliteral, 2, 2)
	var s1 *sliteral = new(sliteral)
	s1.value = "hello0"
	s1.label = ".main.S0"
	stringLiterals[0] = s1

	var s2 *sliteral = new(sliteral)
	s2.value = "hello1"
	s2.label = ".main.S1"
	stringLiterals[1] = s2


	/*
	var globalVar0 *astValueSpec = new(astValueSpec)
	var ident0 *astIdent = new(astIdent)
	ident0.Name = "_gvar0"
	globalVar0.Name = ident0
	globalVar0.value = "10"
	globalVar0.t = new(Type)
	globalVar0.t.kind = "T_INT"
	globalVars = append(globalVars, globalVar0)

	var globalVar1 *astValueSpec = new(astValueSpec)
	globalVar1.name = "_gvar1"
	globalVar1.value = "11"
	globalVar1.t = new(Type)
	globalVar1.t.kind = "T_INT"
	globalVars = append(globalVars, globalVar1)

	var globalVar2 *astValueSpec = new(astValueSpec)
	globalVar2.name = "_gvar2"
	globalVar2.value = "foo"
	globalVar2.t = new(Type)
	globalVar2.t.kind = "T_STRING"
	globalVars = append(globalVars, globalVar2)
*/
	return file.Name
}
