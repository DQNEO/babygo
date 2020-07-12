package runtime

var heapHead uintptr
var heapCurrent uintptr
var heapTail uintptr

var s string

func heapInit() {
	heapHead = brk(0)
	s = "runtime"
}

func brk(addr uintptr) uintptr {
	var ret uintptr
	ret = 0
	return ret
}
