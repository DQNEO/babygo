package types2

type Struct struct {
	Fields       []*Var // Fields != nil indicates the struct is set up (possibly with len(Fields) == 0)
	IsCalculated bool   // the offsets of fields are calculated or not
}

func NewStruct(fields []*Var) *Struct {
	return &Struct{
		Fields: fields,
	}
}

func (t *Struct) Underlying() Type { return t }
func (t *Struct) String() string   { return "struct{" + "@TODO:list fields" + "}" }
