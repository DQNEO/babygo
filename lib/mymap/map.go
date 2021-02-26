package mymap

type MapStrKey struct {
	first *mapLinkedListStrKey
	last *mapLinkedListStrKey
	length int
}

type mapLinkedListStrKey struct {
	key string
	value interface{}
	next *mapLinkedListStrKey
}

func (mp *MapStrKey) Len() int {
	return mp.length
}

func (mp *MapStrKey) Get(key string) (interface{}, bool){
	for item:=mp.first; item!=nil; item=item.next {
		if item.key == key {
			return item.value, true
		}
	}
	return nil, false
}

func (mp *MapStrKey) Delete(key string) {
	if mp.first.key == key {
		mp.first = mp.first.next
		mp.length -= 1
		if mp.last.key == key {
			mp.last = nil
		}
		return
	}
	var last *mapLinkedListStrKey
	for item:=mp.first; item!=nil; item=item.next {
		if item.key == key {
			last.next = item.next
			mp.length -= 1
			if mp.last.key == key {
				mp.last = nil
			}
			return
		}
		last = item
	}

}

func (mp *MapStrKey) Set(key string, value interface{}) {
	if mp.first == nil {
		mp.first = &mapLinkedListStrKey{
			key:   key,
			value: value,
		}
		mp.last = mp.first
		mp.length += 1
		return
	}
	var last *mapLinkedListStrKey
	for item:=mp.first; item!=nil; item=item.next {
		if item.key == key {
			item.value = value
			return
		}
		last = item
	}
	newItem := &mapLinkedListStrKey{
		key: key,
		value: value,
	}
	last.next = newItem
	mp.last = newItem
	mp.length += 1
}

