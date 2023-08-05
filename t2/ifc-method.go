package main

import (
	"os"
)

type WriteCloser interface {
	Write(p []byte) (n int, err error)
	Close() error
}

var ifc WriteCloser

func f1() {
	ifc = os.Stdout
}

func f2() {
	var msg = []uint8("hello world!\n")
	ifc.Write(msg)
}
func main() {
	f1()
	f2()
}
