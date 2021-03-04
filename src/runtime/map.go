package runtime

import "unsafe"

type Map struct {
	first  *item
	length int
}

type item struct {
	key   interface{}
	value string // interface{}
	next  *item
}

func (i *item) valueAddr() unsafe.Pointer {
	return unsafe.Pointer(&i.value)
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
	return false
}

func makeMap(size uintptr) uintptr {
	var mp = &Map{}
	var addr uintptr = uintptr(unsafe.Pointer(mp))
	return addr
}

func lenMap(mp *Map) int {
	return mp.length
}

func deleteFromMap(mp *Map, key interface{}) {
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

func getAddrForMapSet(mp *Map, key interface{}) unsafe.Pointer {
	if mp.first == nil {
		// alloc new item
		mp.first = &item{
			key: key,
		}
		mp.length += 1
		return mp.first.valueAddr()
	}
	var last *item
	for item:=mp.first; item!=nil; item=item.next {
		if item.match(key) {
			return item.valueAddr()
		}
		last = item
	}
	newItem := &item{
		key:   key,
	}
	last.next = newItem
	mp.length += 1

	return newItem.valueAddr()
}

var emptyString string  // = "not found"
func getAddrForMapGet(mp *Map, key interface{}) unsafe.Pointer {
	for item:=mp.first; item!=nil; item=item.next {
		if item.match(key) {
			return item.valueAddr()
		}
	}
	return unsafe.Pointer(&emptyString)
}


