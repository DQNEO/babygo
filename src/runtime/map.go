package runtime

import "unsafe"

type Map struct {
	first  *item
	length int
	valueSize uintptr
}

type item struct {
	key   interface{}
	value uintptr
	next  *item
}

func (i *item) valueAddr() unsafe.Pointer {
	return unsafe.Pointer(i.value)
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
	var mp = &Map{
		valueSize: 24,
	}
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
	var last *item
	for item:=mp.first; item!=nil; item=item.next {
		if item.match(key) {
			return item.valueAddr()
		}
		last = item
	}
	newItem := &item{
		key:   key,
		value: malloc(mp.valueSize),
	}
	if mp.first == nil {
		mp.first = newItem
	} else {
		last.next = newItem
	}
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


