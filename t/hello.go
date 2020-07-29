package main

import "syscall"

func main() {
	var msg = []uint8("hello world!\n")
	syscall.Write(1, msg)
}
