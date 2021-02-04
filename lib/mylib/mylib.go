package mylib

import "unsafe"
import "github.com/DQNEO/babygo/lib/mylib2"

func Sum(a int, b int) int {
	return a + b
}

func Sum2(a int, b int) int {
	return mylib2.Sum2(a, b)
}

func Cstring2string(b *byte) string {
	var bs []byte
	for {
		if b == nil || *b == 0 {
			break
		}
		bs = append(bs, *b)
		p := uintptr(unsafe.Pointer(b)) + 1
		b = (*byte)(unsafe.Pointer(p))
	}
	return string(bs)
}

