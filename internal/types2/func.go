package types2

type Func struct {
	Typ  *Signature
	Name string
}

func NewFunc(sig *Signature) *Func {
	return &Func{
		Typ: sig,
	}
}

func (t *Func) Underlying() Type { return t.Typ }
func (t *Func) String() string   { return "func " + t.Typ.String() }
