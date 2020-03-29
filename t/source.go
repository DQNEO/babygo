package main

import "os"

func main() {
	print("hello world\n")
	print("to stderr\n")
	os.Exit((20 + 1) * 2)
}
