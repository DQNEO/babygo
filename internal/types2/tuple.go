package types2

type Tuple struct {
	Types []Type
}

func (t *Tuple) Underlying() Type { return t }
func (t *Tuple) String() string   { return "@TBI" }
