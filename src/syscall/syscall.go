package syscall

import "unsafe"

const SYS_READ uintptr = 0
const SYS_WRITE uintptr = 1
const SYS_OPEN uintptr = 2
const SYS_CLOSE uintptr = 3
const SYS_GETDENTS64 uintptr = 217

func Read(fd int, buf []byte) (uintptr, int) {
	p := &buf[0]
	_cap := cap(buf)
	var ret uintptr
	ret = Syscall(SYS_READ, uintptr(fd), uintptr(unsafe.Pointer(p)), uintptr(_cap))
	return ret, 0
}

func Open(path string, mode int, perm int) (uintptr, int) {
	buf := []byte(path)
	buf = append(buf, 0) // add null terminator
	p := &buf[0]
	var fd uintptr
	fd = Syscall(SYS_OPEN, uintptr(unsafe.Pointer(p)), uintptr(mode), uintptr(perm))
	return fd, 0
}

func Close(fd int) uintptr {
	var e uintptr
	e = Syscall(SYS_CLOSE, uintptr(fd), 0, 0)
	return e
}

func Write(fd int, buf []byte) (uintptr, int) {
	p := &buf[0]
	_len := len(buf)
	var ret uintptr
	ret = Syscall(SYS_WRITE, uintptr(fd), uintptr(unsafe.Pointer(p)), uintptr(_len))
	return ret, 0
}

func Getdents(fd int, buf []byte) (int, int) {
	var _p0 unsafe.Pointer
	_p0 = unsafe.Pointer(&buf[0])
	nread := Syscall(SYS_GETDENTS64, uintptr(fd), uintptr(_p0), uintptr(len(buf)))
	return int(nread), 0
}

func Syscall(trap uintptr, a1 uintptr, a2 uintptr, a3 uintptr) uintptr
