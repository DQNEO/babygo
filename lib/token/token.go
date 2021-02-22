package token

type Token string

func (tok Token) String() string {
	return string(tok)
}

