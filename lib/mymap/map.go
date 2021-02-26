package mymap

type MapStrKey struct {
	first  *mapLinkedListStrKey
	last   *mapLinkedListStrKey
	length int
}

type mapLinkedListStrKey struct {
	Key   string
	Value interface{}
	next  *mapLinkedListStrKey
}

func (i *mapLinkedListStrKey) Next() *mapLinkedListStrKey {
	return i.next
}

func (mp *MapStrKey) Len() int {
	return mp.length
}

func (mp *MapStrKey) First() *mapLinkedListStrKey {
	return mp.first
}

func (mp *MapStrKey) Get(key string) (interface{}, bool){
	for item:=mp.first; item!=nil; item=item.next {
		if item.Key == key {
			return item.Value, true
		}
	}
	return nil, false
}

func (mp *MapStrKey) Delete(key string) {
	if mp.first.Key == key {
		mp.first = mp.first.next
		mp.length -= 1
		if mp.last.Key == key {
			mp.last = nil
		}
		return
	}
	var last *mapLinkedListStrKey
	for item:=mp.first; item!=nil; item=item.next {
		if item.Key == key {
			last.next = item.next
			mp.length -= 1
			if mp.last.Key == key {
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
			Key:   key,
			Value: value,
		}
		mp.last = mp.first
		mp.length += 1
		return
	}
	var last *mapLinkedListStrKey
	for item:=mp.first; item!=nil; item=item.next {
		if item.Key == key {
			item.Value = value
			return
		}
		last = item
	}
	newItem := &mapLinkedListStrKey{
		Key:   key,
		Value: value,
	}
	last.next = newItem
	mp.last = newItem
	mp.length += 1
}

