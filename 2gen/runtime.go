package runtime

import "syscall"
import "unsafe"

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
	heapCurrent = heapHead
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

func panic(s string) {
	var buf []uint8 = []uint8(s)
	syscall.Write(2, buf)
	var arg0 uintptr = uintptr(60) // sys exit
	var arg1 uintptr = 1 // status
	var arg2 uintptr = uintptr(0)
	var arg3 uintptr = uintptr(0)
	syscall.Syscall(arg0, arg1, arg2, arg3)
}

func memzeropad(addr uintptr, size uintptr) {
	var p *uint8 = (*uint8)(unsafe.Pointer(addr))
	var isize int = int(size)
	var i int
	var up uintptr
	for i = 0; i < isize; i = i+1 {
		*p = 0
		up = uintptr(unsafe.Pointer(p)) + 1
		p = (*uint8)(unsafe.Pointer(up))
	}
}

func malloc(size uintptr) uintptr {
	if heapCurrent+size > heapTail {
		panic("malloc exceeds heap capacity")
		return 0
	}
	var r uintptr
	r = heapCurrent
	heapCurrent = heapCurrent + size
	memzeropad(r, size)
	return r
}
