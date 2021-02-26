package main

import "unsafe"

type mapPtrIfc []*mapPtrIfcEntry

type mapPtrIfcEntry struct {
	key   unsafe.Pointer
	value interface{}
}

func (mp *mapPtrIfc) get(key unsafe.Pointer) (interface{}, bool) {
	for _, te := range *mp {
		if te.key == key {
			if te.value == nil {
				panic("no value")
			}
			return te.value, true
		}
	}
	return nil, false
}

func (mp *mapPtrIfc) set(key unsafe.Pointer, value interface{}) {
	for _, te := range *mp {
		if te.key == key {
			te.value = value
		}
	}
	te := &mapPtrIfcEntry{
		key:   key,
		value: value,
	}
	*mp = append(*mp, te)
}
