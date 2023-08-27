package types2

// A Pointer represents a pointer type.
type Pointer struct {
	base Type // element Type
}

// NewPointer returns a new pointer Type for the given element (base) Type.
func NewPointer(elem Type) *Pointer { return &Pointer{base: elem} }

// Elem returns the element Type for the given pointer p.
func (p *Pointer) Elem() Type { return p.base }

func (t *Pointer) Underlying() Type { return t }
func (t *Pointer) String() string   { return "*" + t.Elem().String() }
