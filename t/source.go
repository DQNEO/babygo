package main

import "os"

var myInt int = 30
var myInt2 int = 12
var myString string = "to stderr\n"

func main() {
	print("hello world\n")
	print(myString)
	os.Exit(myInt + myInt2)
}
