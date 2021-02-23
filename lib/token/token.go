package token

type Token string
type Pos int

// Kind
var INT Token = "INT"
var STRING Token = "STRING"

var NoPos Pos = 0

// Token
var ADD Token = "+"
var SUB Token = "-"

func (tok Token) String() string {
	return string(tok)
}
