package main

import (
	"io"

	"os"
)

var msg = []uint8("hello world!\n")
var ifc io.WriteCloser

func f1() {
	ifc = os.Stdout
}

func f2() {
	ifc.Write(msg)
}
func main() {
	f1()
	f2()
}
