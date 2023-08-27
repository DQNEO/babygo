package types2

type Signature struct {
	Params  *Tuple
	Results *Tuple
}

func (t *Signature) Underlying() Type { return t }
func (t *Signature) String() string   { return "@TBI" }
