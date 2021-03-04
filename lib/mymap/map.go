package mymap

import "unsafe"

type Map struct {
	first  *item
	length int
}

type item struct {
	key   interface{}
	Value interface{}
	next  *item
}

func (i *item) Next() *item {
	return i.next
}

func (i *item) GetKeyAsString() string {
	return i.key.(string)
}

func (i *item) match(key interface{}) bool {
	switch k := key.(type) {
	case string:
		return i.key.(string) == k
	case unsafe.Pointer:
		return i.key.(unsafe.Pointer) == k
	default:
		panic("Not supported key type")
	}
	panic("Not supported key type")
}

func (mp *Map) Len() int {
	return mp.length
}

func (mp *Map) First() *item {
	return mp.first
}

func (mp *Map) Get(key interface{}) (interface{}, bool){
	for item:=mp.first; item!=nil; item=item.next {
		if item.match(key) {
			return item.Value, true
		}
	}
	return nil, false
}

func (mp *Map) Delete(key interface{}) {
	if mp.first == nil {
		return
	}
	if mp.first.match(key) {
		mp.first = mp.first.next
		mp.length -= 1
		return
	}
	var prev *item
	for item:=mp.first; item!=nil; item=item.next {
		if item.match(key) {
			prev.next = item.next
			mp.length -= 1
			return
		}
		prev = item
	}

}

func (mp *Map) Set(key interface{}, value interface{}) {
	if mp.first == nil {
		mp.first = &item{
			key:   key,
			Value: value,
		}
		mp.length += 1
		return
	}
	var last *item
	for item:=mp.first; item!=nil; item=item.next {
		if item.match(key) {
			item.Value = value
			return
		}
		last = item
	}
	newItem := &item{
		key:   key,
		Value: value,
	}
	last.next = newItem
	mp.length += 1
}

