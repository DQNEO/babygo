package types2

// A Map represents a map type.
type Map struct {
	key  Type
	elem Type
}

// NewMap returns a new map for the given key and element types.
func NewMap(key Type, elem Type) *Map {
	return &Map{key: key, elem: elem}
}

// Key returns the key type of map m.
func (m *Map) Key() Type { return m.key }

// Elem returns the element type of map m.
func (m *Map) Elem() Type { return m.elem }

func (t *Map) Underlying() Type { return t }
func (t *Map) String() string   { return "map[" + t.key.String() + "]" + t.Elem().String() }
