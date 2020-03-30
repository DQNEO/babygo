package main

import "os"

var myInt int = 40
var myString string = "to stderr\n"

func main() {
	print("hello world\n")
	print(myString)
	os.Exit(myInt + 2)
}
