package scanner

import (
	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/strconv"
	"github.com/DQNEO/babygo/lib/token"
)

type Scanner struct {
	src        []uint8
	ch         uint8
	offset     int
	nextOffset int
	insertSemi bool
	File       *token.File
}

func (s *Scanner) next() {
	if s.nextOffset < len(s.src) {
		s.offset = s.nextOffset
		s.ch = s.src[s.offset]
		s.nextOffset++
		if s.ch == '\n' {
			s.File.Lines = append(s.File.Lines, token.Pos(s.File.Base+s.nextOffset))
		}
	} else {
		s.offset = len(s.src)
		s.ch = 1 //EOF
	}
}

var keywords []string

func (s *Scanner) Init(f *token.File, src []uint8) {
	// https://golang.org/ref/spec#Keywords
	keywords = []string{
		"break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var",
	}
	s.File = f
	s.src = src
	s.File.Lines = []token.Pos{token.Pos(f.Base)}
	s.offset = 0
	s.ch = ' '
	s.nextOffset = 0
	s.insertSemi = false
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

func isNumber(ch uint8) bool {
	switch ch {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'x', '_':
		return true
	default:
		return false
	}
}

func (s *Scanner) scanIdentifier() string {
	var offset = s.offset
	for isLetter(s.ch) || isDecimal(s.ch) {
		s.next()
	}
	return string(s.src[offset:s.offset])
}

func (s *Scanner) scanNumber() string {
	var offset = s.offset
	for isNumber(s.ch) {
		s.next()
	}
	return string(s.src[offset:s.offset])
}

func (s *Scanner) scanString() string {
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

func (s *Scanner) scanChar() string {
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

func (s *Scanner) scanComment() string {
	var offset = s.offset - 1
	for s.ch != '\n' {
		s.next()
	}
	return string(s.src[offset:s.offset])
}

// https://golang.org/ref/spec#Tokens
func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || (s.ch == '\n' && !s.insertSemi) || s.ch == '\r' {
		s.next()
	}
}

func (s *Scanner) Scan() (string, string, token.Pos) {
	s.skipWhitespace()
	pos := token.Pos(s.File.Base + s.offset)

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
		case '\'': // Single quote
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
					s.insertSemi = false
					return "\n", ";", pos
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
			//			logf("[Scanner] EOF @ file=%s line=%d final_offset=%d, Pos=%d\n", s.File.Name, len(s.File.Lines)-1, s.offset, Pos)
		default:
			panic("unknown char:" + string([]uint8{ch}) + ":" + strconv.Itoa(int(ch)))
			tok = "UNKNOWN"
		}
	}
	s.insertSemi = insertSemi
	return lit, tok, pos
}
