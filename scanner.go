package main

import (
	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/strconv"
)

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
		if mylib.InArray(lit, keywords) {
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
