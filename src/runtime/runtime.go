// runtime for 2nd generation compiler
package runtime

import "unsafe"

const heapSize uintptr = 620205360

var heapHead uintptr
var heapCurrent uintptr
var heapTail uintptr

const SYS_BRK int = 12
const SYS_EXIT int = 60

var argc int
var argv **uint8

func argv_index(argv **uint8, i int) *uint8 {
	return *(**uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(argv)) + uintptr(i)*8))
}

func args(c int, v **uint8) {
	argc = c
	argv = v
}

func goargs() {
	argslice = make([]string, argc, argc)
	for i := 0; i < argc; i++ {
		argslice[i] = cstring2string(argv_index(argv, i))
	}
}

var argslice []string

func schedinit() {
	heapInit()
	goargs()
	envInit()
}

var mainStarted bool

var main_main func() // = main.main

func main() {
	mainStarted = true
	var fn = main_main
	fn()
	exit(0)
}

type p struct {
	runq func()
}

var p0 p

func newproc(size int, fn *func()) {
	p0.runq = *fn
}

func mstart1() {
	Write(2, []byte("hello, I am a cloned thread in mstart1\n"))
	exitThread()
}

const CloneFlags int = 331520
func newosproc() {
	var fn func() = mstart1
	clone(CloneFlags - 256 , 0, fn)
}

func mstart0() {
	// clone : create OS thread

	var g func() = p0.runq
	g()
}

// Environment variables
var envp uintptr
var envlines []string // []{"FOO=BAR\0", "HOME=/home/...\0", ..}

type envEntry struct {
	key   string
	value string
}

var Envs []*envEntry

func heapInit() {
	heapHead = brk(0)
	heapTail = brk(heapHead + heapSize)
	heapCurrent = heapHead
}

// Inital stack layout is illustrated in this page
// http://asm.sourceforge.net/articles/startup.html#st
func envInit() {
	var p uintptr // **byte

	for p = envp; true; p = p + 8 {
		var bpp **byte = (**byte)(unsafe.Pointer(p))
		if *bpp == nil {
			break
		}
		envlines = append(envlines, cstring2string(*bpp))
	}

	for _, envline := range envlines {
		var i int
		var c byte
		for i, c = range []byte(envline) {
			if c == '=' {
				break
			}
		}
		key := envline[:i]
		value := envline[i+1:]

		entry := &envEntry{
			key:   key,
			value: value,
		}
		Envs = append(Envs, entry)

	}
}

func runtime_getenv(key string) string {
	for _, e := range Envs {
		if e.key == key {
			return e.value
		}
	}

	return ""
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


// This func has an alias in os package
func runtime_args() []string {
	return argslice
}

func brk(addr uintptr) uintptr {
	var ret uintptr
	ret = Syscall(uintptr(SYS_BRK), addr, uintptr(0), uintptr(0))
	return ret
}

func panic(ifc interface{}) {
	switch x := ifc.(type) {
	case string:
		var s = "panic: " + x + "\n\n"
		Write(2, []uint8(s))
		Syscall(uintptr(SYS_EXIT), 1, uintptr(0), uintptr(0))
	default:
		var s = "panic: " + "Unknown type" + "\n\n"
		Write(2, []uint8(s))
		Syscall(uintptr(SYS_EXIT), 1, uintptr(0), uintptr(0))
	}
}

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
		Write(2, []uint8("malloc exceeded heap max"))
		Syscall(uintptr(SYS_EXIT), 1, uintptr(0), uintptr(0))
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
	if a == c && b == d {
		return true
	}
	return false
}

func Write(fd int, p []byte) int
func Syscall(trap uintptr, a1 uintptr, a2 uintptr, a3 uintptr) uintptr
func exit(c int)
func exitThread()
func clone(flags int, b uintptr, fn func())

// Actually this is an alias to makeSlice
func makeSlice1(elmSize int, slen int, scap int) []uint8
func makeSlice8(elmSize int, slen int, scap int) []int
func makeSlice16(elmSize int, slen int, scap int) []string
func makeSlice24(elmSize int, slen int, scap int) [][]int
