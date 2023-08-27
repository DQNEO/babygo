package types2

type Slice struct {
	elem   Type
	IsElps bool
}

var GeneralSliceType = &Slice{}

// NewSlice returns a new slice type for the given element type.
func NewSlice(elem Type) *Slice { return &Slice{elem: elem} }

// Elem returns the element Type of slice s.
func (s *Slice) Elem() Type { return s.elem }

func (t *Slice) Underlying() Type { return t }
func (t *Slice) String() string   { return "[]" + t.Elem().String() }
