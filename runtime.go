// runtime
package runtime

import "unsafe"

var heap [4096]uint8

var heapHead uintptr
var heapCurrent uintptr
var heapTail uintptr

func heapInit() {
	heapHead = uintptr(unsafe.Pointer(&heap[0])) // brk(0)
	heapTail = heapHead + 4096   // brk(heapHead + heapSize)
	heapCurrent = heapHead
}

func malloc(size uintptr) uintptr {
	if heapCurrent+size > heapTail {
		panic("malloc exceeds heap capacity")
		return 0
	}
	var r uintptr
	r = heapCurrent
	heapCurrent = heapCurrent + size
	return r
}

func panic(s string) {
	print(s)
	//exit(1)
}

func catstrings(a string, b string) string {
	var totallen int
	var r []uint8
	totallen = len(a) + len(b)
	r = make([]uint8, totallen, totallen)
	var i int
	for i=0;i<len(a);i=i+1 {
		r[i] = a[i]
	}
	var j int
	for j=0;j<len(b);j=j+1 {
		r[i+j] = b[j]
	}
	return string(r)
}
