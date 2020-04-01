package main

import "os"

var globalstring string = "hello stderr\n"

var globalint int = 30
var globalint2 int = 0

func add1(x int) int {
	return x + 1
}

func sum(x int, y int) int {
	return x + y
}

func print1(a string) {
	print(a)
	return
}

func print2(a string, b string) {
	print(a)
	print(b)
}

func returnstring() string {
	return "i am a local 1\n"
}

var globalchars [10]uint8

func main() {
	globalint2 = sum(1, 2)

	print(globalstring)

	var localstring1 string
	var localstring2 string

	print1("hello string literal\n")
	localstring1 = returnstring()
	localstring2 = "i m local2\n"
	print2(localstring1, localstring2)
	globalstring = "globalstring changed\n"
	print(globalstring)
	var locali3 int
	var eight int
	eight = '9' - '1'
	locali3 = add1(eight)
	os.Exit( sum(globalint , globalint2) + locali3)
}
