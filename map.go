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
	te := &mapStringIntEntry{
		key:   key,
		value: value,
	}
	*mp = append(*mp, te)
}

