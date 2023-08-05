package syscall

import (
	"unsafe"

	"runtime"
)

// cheat sheet: https://chromium.googlesource.com/chromiumos/docs/+/HEAD/constants/syscalls.md#x86_64-64_bit
const SYS_READ uintptr = 0
const SYS_WRITE uintptr = 1
const SYS_OPEN uintptr = 2
const SYS_CLOSE uintptr = 3
const SYS_GETDENTS64 uintptr = 217

func Getenv(key string) (string, bool) {
	for _, e := range runtime.Envs {
		if e.Key == key {
			return e.Value, true
		}
	}

	return "", false
}

func Environ() []string {
	return runtime.Envlines
}

func Read(fd int, buf []byte) (uintptr, error) {
	p := &buf[0]
	_cap := cap(buf)
	var ret uintptr
	ret, _, _ = Syscall(SYS_READ, uintptr(fd), uintptr(unsafe.Pointer(p)), uintptr(_cap))
	return ret, nil
}

func Open(path string, mode int, perm int) (uintptr, error) {
	buf := []byte(path)
	buf = append(buf, 0) // add null terminator
	p := &buf[0]
	var fd uintptr
	fd, _, _ = Syscall(SYS_OPEN, uintptr(unsafe.Pointer(p)), uintptr(mode), uintptr(perm))
	return fd, nil
}

func Close(fd int) error {
	Syscall(SYS_CLOSE, uintptr(fd), 0, 0)
	return nil
}

func Write(fd int, buf []byte) (uintptr, error) {
	p := &buf[0]
	_len := len(buf)
	var ret uintptr
	ret, _, _ = Syscall(SYS_WRITE, uintptr(fd), uintptr(unsafe.Pointer(p)), uintptr(_len))
	return ret, nil
}

func Getdents(fd int, buf []byte) (int, error) {
	var _p0 unsafe.Pointer
	_p0 = unsafe.Pointer(&buf[0])
	nread, _, _ := Syscall(SYS_GETDENTS64, uintptr(fd), uintptr(_p0), uintptr(len(buf)))
	return int(nread), nil
}

//go:linkname Syscall
func Syscall(trap uintptr, a1 uintptr, a2 uintptr, a3 uintptr) (uintptr, uintptr, uintptr)
