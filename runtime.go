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

func memzeropad(addr uintptr, size uintptr) {
	var p *uint8 = (*uint8)(unsafe.Pointer(addr))
	var isize int = int(size)
	var i int
	var up uintptr
	for i=0;i<isize;i++ {
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

func makeSlice(elmSize int, slen int, scap int) (uintptr, int, int) {
	var size uintptr = uintptr(elmSize * scap)
	var addr uintptr = malloc(size)
	return addr, slen, scap
}

func copySlice8(src []uint8, dst []uint8) {
	var i int
	for i = 0; i < len(src); i++ {
		dst[i] = src[i]
	}
}

// Actually this is an alias to makeSlice
func _makeSlice8(elmSize int, slen int, scap int) []uint8

func append8(x []uint8, elm uint8) []uint8 {
	var z []uint8
	var xlen int = len(x)
	var zlen int = xlen + 1

	if cap(x) >= zlen {
		z = x[0:zlen]
		nop1()
	} else {
		var newcap int
		if xlen == 0 {
			newcap = 1
		} else {
			newcap = xlen * 2
		}
		z = _makeSlice8(1, zlen, newcap)
		nop()
		copySlice8(x,z)
	}

	z[xlen] = elm
	return z
}

func panic(s string) {
	print(s)
	//exit(1)
}

func nop() {
}

func nop1() {
}

func nop2() {
}

func nop3() {
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
