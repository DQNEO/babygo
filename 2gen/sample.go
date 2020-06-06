package main

import "os"
import "syscall"

func main() {

	var s string
	s = "hello world\n"
	//var buf []uint8
	//buf = []uint8(s)
	//syscall.Write(1, buf)


	os.Exit(42)
}
