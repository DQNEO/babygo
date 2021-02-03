package syscall

import "unsafe"

const SYS_READ int = 0
const SYS_OPEN int = 2

func Read(fd int, buf []byte) uintptr {
	p := &buf[0]
	_cap := cap(buf)
	var ret uintptr
	ret  = Syscall(uintptr(SYS_READ), uintptr(fd), uintptr(unsafe.Pointer(p)), uintptr(_cap))
	return ret
}

func Open(path string, mode int, perm int) uintptr {
	buf := []byte(path)
	buf = append(buf, 0) // add null terminator
	p := &buf[0]
	var ret uintptr
	ret = Syscall(uintptr(SYS_OPEN), uintptr(unsafe.Pointer(p)), uintptr(mode), uintptr(perm))
	return ret
}

func Syscall(trap uintptr, a1 uintptr, a2 uintptr, a3 uintptr) uintptr
