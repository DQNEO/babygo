package types2

type Interface struct {
	Methods []*Func
}

var EmptyInterface = &Interface{}

func NewInterfaceType(methods []*Func) *Interface {
	i := &Interface{
		Methods: methods,
	}
	return i
}

func (t *Interface) Underlying() Type { return t }
func (t *Interface) String() string   { return "interface{" + "@TODO: list methods" + "}" }
