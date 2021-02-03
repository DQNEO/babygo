package syscall

import "unsafe"

const SYS_READ int = 0

func Read(fd int, buf []byte) uintptr {
	p := &buf[0]
	_cap := cap(buf)
	var ret uintptr
	ret = syscall.Syscall(uintptr(SYS_READ), uintptr(fd), uintptr(unsafe.Pointer(p)), uintptr(_cap))
	return ret
}

