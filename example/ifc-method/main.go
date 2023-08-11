package main

import "os"

func main() {
	var msg = []uint8("hello world!\n")
	os.Stdout.Write(msg)
}
