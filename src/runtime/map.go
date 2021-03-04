package runtime

import "unsafe"

type Map struct {
	first  *item
	length int
}

type item struct {
	key   interface{}
	Value string // interface{}
	next  *item
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

func getAddrForMapSet(mp *Map, key string) unsafe.Pointer {
	if mp.first == nil {
		// alloc new item
		mp.first = &item{
			key: key,
		}
		mp.length += 1
		return unsafe.Pointer(&mp.first.Value)
	}
	var last *item
	for item:=mp.first; item!=nil; item=item.next {
		if item.match(key) {
			return unsafe.Pointer(&item.Value)
		}
		last = item
	}
	newItem := &item{
		key:   key,
	}
	last.next = newItem
	mp.length += 1

	return unsafe.Pointer(&newItem.Value)
	//return unsafe.Pointer(mp)
}

var emptyString string  // = "not found"
func getAddrForMapGet(mp *Map, key string) unsafe.Pointer {
	for item:=mp.first; item!=nil; item=item.next {
		if item.match(key) {
			return unsafe.Pointer(&item.Value)
		}
	}
	return unsafe.Pointer(&emptyString)
}


