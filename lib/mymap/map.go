package mymap

type MapStrKey struct {
	first  *item
	last   *item
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
		itemKey := i.key.(string)
		if itemKey == k {
			return true
		} else {
			return false
		}
	default:
		panic("Not supported key type")
	}
	panic("Not supported key type")
}

func (mp *MapStrKey) Len() int {
	return mp.length
}

func (mp *MapStrKey) First() *item {
	return mp.first
}

func (mp *MapStrKey) Get(key interface{}) (interface{}, bool){
	for item:=mp.first; item!=nil; item=item.next {
		if item.match(key) {
			return item.Value, true
		}
	}
	return nil, false
}

func (mp *MapStrKey) Delete(key string) {
	if mp.first.match(key) {
		mp.first = mp.first.next
		mp.length -= 1
		if mp.last.match(key) {
			mp.last = nil
		}
		return
	}
	var last *item
	for item:=mp.first; item!=nil; item=item.next {
		if item.match(key) {
			last.next = item.next
			mp.length -= 1
			if mp.last.match(key) {
				mp.last = nil
			}
			return
		}
		last = item
	}

}

func (mp *MapStrKey) Set(key interface{}, value interface{}) {
	if mp.first == nil {
		mp.first = &item{
			key:   key,
			Value: value,
		}
		mp.last = mp.first
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
	mp.last = newItem
	mp.length += 1
}

