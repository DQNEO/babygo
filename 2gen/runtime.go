package runtime

import "syscall"

var SYS_BRK int

var heapHead uintptr
var heapCurrent uintptr
var heapTail uintptr
var heapSize uintptr

var s string

func heapInit() {
	SYS_BRK = 12
	heapSize = 10205360
	heapHead = brk(0)
	heapTail = brk(heapHead + heapSize)
	s = "runtime"
}

func brk(addr uintptr) uintptr {
	var ret uintptr = 0
	var arg0 uintptr = uintptr(SYS_BRK)
	var arg1 uintptr = addr
	var arg2 uintptr = uintptr(0)
	var arg3 uintptr = uintptr(0)
	ret = syscall.Syscall(arg0, arg1, arg2, arg3)
	return ret
}
