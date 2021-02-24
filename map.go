package main

type mapStringInt []*mapStringIntEntry
type mapStringIntEntry struct {
	key   string
	value int
}

func (mp *mapStringInt) get(key string) (int, bool) {
	for _, te := range *mp {
		if te.key == key {
			return te.value, true
		}
	}
	return 0, false
}

func (mp *mapStringInt) set(key string, value int)  {
	for _, te := range *mp {
		if te.key == key {
			te.value = value
		}
	}
	te := &mapStringIntEntry{
		key:   key,
		value: value,
	}
	*mp = append(*mp, te)
}

type mapStringIfc []*mapStringIfcEntry

type mapStringIfcEntry struct {
	key   string
	value interface{}
}

func (mp *mapStringIfc) get(key string) (interface{}, bool) {
	for _, te := range *mp {
		if te.key == key {
			return te.value, true
		}
	}
	return 0, false
}

func (mp *mapStringIfc) set(key string, value interface{})  {
	for _, te := range *mp {
		if te.key == key {
			te.value = value
		}
	}
	te := &mapStringIfcEntry{
		key:   key,
		value: value,
	}
	*mp = append(*mp, te)
}
