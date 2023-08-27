package types2

// An Array represents an array type.
type Array struct {
	len  int
	elem Type
}

// NewArray returns a new array type for the given element type and length.
// A negative length indicates an unknown length.
func NewArray(elem Type, len int) *Array { return &Array{len: len, elem: elem} }

// Len returns the length of array a.
// A negative result indicates an unknown length.
func (a *Array) Len() int { return a.len }

// Elem returns element type of array a.
func (a *Array) Elem() Type { return a.elem }

func (t *Array) Underlying() Type { return t }
func (t *Array) String() string   { return "@TBI" }
