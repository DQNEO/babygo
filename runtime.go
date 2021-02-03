// runtime for 2nd generation compiler
// +build ignore

package runtime

import "unsafe"

const heapSize uintptr = 320205360

var heapHead uintptr
var heapCurrent uintptr
var heapTail uintptr

const SYS_BRK int = 12
const SYS_EXIT int = 60

var __argv__ []*uint8 // C argv

func heapInit() {
	heapHead = brk(0)
	heapTail = brk(heapHead + heapSize)
	heapCurrent = heapHead
}

func cstring2string(b *uint8) string {
	var buf []uint8
	for {
		if b == nil || *b == 0 {
			break
		}
		buf = append(buf, *b)
		var p uintptr = uintptr(unsafe.Pointer(b)) + 1
		b = (*uint8)(unsafe.Pointer(p))
	}
	return string(buf)
}

// create os.Args
func runtime_args() []string {
	var a *uint8
	var r []string
	for _, a = range __argv__ {
		var s string = cstring2string(a)
		r = append(r, s)
	}

	return r
}

func brk(addr uintptr) uintptr {
	var ret uintptr
	ret = Syscall(uintptr(SYS_BRK), addr, uintptr(0), uintptr(0))
	return ret
}

func panic(x string) {
	var s = "panic: " + x + "\n\n"
	Write(2, []uint8(s))
	Syscall(uintptr(SYS_EXIT), 1, uintptr(0), uintptr(0))
}

func Write(fd int, p []byte) int
func Syscall(trap uintptr, a1 uintptr, a2 uintptr, a3 uintptr) uintptr

func memzeropad(addr1 uintptr, size uintptr) {
	var p *uint8 = (*uint8)(unsafe.Pointer(addr1))
	var isize int = int(size)
	var i int
	var up uintptr
	for i = 0; i < isize; i++ {
		*p = 0
		up = uintptr(unsafe.Pointer(p)) + 1
		p = (*uint8)(unsafe.Pointer(up))
	}
}

func memcopy(src uintptr, dst uintptr, length int) {
	var i int
	var srcp *uint8
	var dstp *uint8
	for i = 0; i < length; i++ {
		srcp = (*uint8)(unsafe.Pointer(src + uintptr(i)))
		dstp = (*uint8)(unsafe.Pointer(dst + uintptr(i)))
		*dstp = *srcp
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

// Actually this is an alias to makeSlice
func makeSlice1(elmSize int, slen int, scap int) []uint8
func makeSlice8(elmSize int, slen int, scap int) []int
func makeSlice16(elmSize int, slen int, scap int) []string
func makeSlice24(elmSize int, slen int, scap int) [][]int

func append1(old []uint8, elm uint8) (uintptr, int, int) {
	var new_ []uint8
	var elmSize int = 1

	var oldlen int = len(old)
	var newlen int = oldlen + 1

	if cap(old) >= newlen {
		new_ = old[0:newlen]
	} else {
		var newcap int
		if oldlen == 0 {
			newcap = 1
		} else {
			newcap = oldlen * 2
		}
		new_ = makeSlice1(elmSize, newlen, newcap)
		var oldSize int = oldlen * elmSize
		if oldlen > 0 {
			memcopy(uintptr(unsafe.Pointer(&old[0])), uintptr(unsafe.Pointer(&new_[0])), oldSize)
		}
	}

	new_[oldlen] = elm
	return uintptr(unsafe.Pointer(&new_[0])), newlen, cap(new_)
}

func append8(old []int, elm int) (uintptr, int, int) {
	var new_ []int
	var elmSize int = 8

	var oldlen int = len(old)
	var newlen int = oldlen + 1

	if cap(old) >= newlen {
		new_ = old[0:newlen]
	} else {
		var newcap int
		if oldlen == 0 {
			newcap = 1
		} else {
			newcap = oldlen * 2
		}
		new_ = makeSlice8(elmSize, newlen, newcap)
		var oldSize int = oldlen * elmSize
		if oldlen > 0 {
			memcopy(uintptr(unsafe.Pointer(&old[0])), uintptr(unsafe.Pointer(&new_[0])), oldSize)
		}
	}

	new_[oldlen] = elm
	return uintptr(unsafe.Pointer(&new_[0])), newlen, cap(new_)
}

func append16(old []string, elm string) (uintptr, int, int) {
	var new_ []string
	var elmSize int = 16

	var oldlen int = len(old)
	var newlen int = oldlen + 1

	if cap(old) >= newlen {
		new_ = old[0:newlen]
	} else {
		var newcap int
		if oldlen == 0 {
			newcap = 1
		} else {
			newcap = oldlen * 2
		}
		new_ = makeSlice16(elmSize, newlen, newcap)
		var oldSize int = oldlen * elmSize
		if oldlen > 0 {
			memcopy(uintptr(unsafe.Pointer(&old[0])), uintptr(unsafe.Pointer(&new_[0])), oldSize)
		}
	}

	new_[oldlen] = elm
	return uintptr(unsafe.Pointer(&new_[0])), newlen, cap(new_)
}

func append24(old [][]int, elm []int) (uintptr, int, int) {
	var new_ [][]int
	var elmSize int = 24

	var oldlen int = len(old)
	var newlen int = oldlen + 1

	if cap(old) >= newlen {
		new_ = old[0:newlen]
	} else {
		var newcap int
		if oldlen == 0 {
			newcap = 1
		} else {
			newcap = oldlen * 2
		}
		new_ = makeSlice24(elmSize, newlen, newcap)
		var oldSize int = oldlen * elmSize
		if oldlen > 0 {
			memcopy(uintptr(unsafe.Pointer(&old[0])), uintptr(unsafe.Pointer(&new_[0])), oldSize)
		}
	}

	new_[oldlen] = elm
	return uintptr(unsafe.Pointer(&new_[0])), newlen, cap(new_)
}

func catstrings(a string, b string) string {
	var totallen = len(a) + len(b)
	var r = make([]uint8, totallen, totallen+1) // +1 is a workaround for syscall.Open. see runtime.s
	var i int
	for i = 0; i < len(a); i = i + 1 {
		r[i] = a[i]
	}
	var j int
	for j = 0; j < len(b); j = j + 1 {
		r[i+j] = b[j]
	}
	return string(r)
}

func cmpstrings(a string, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var i int
	for i = 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Two interface values are equal if they have identical dynamic types and equal dynamic values or if both have value nil.
func cmpinterface(a uintptr, b uintptr, c uintptr, d uintptr) bool {
	if a == c && b ==d {
		return true
	}
	return false
}
