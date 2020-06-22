package main

import "os"

var gvar1 int
var gvar2 int
var gstring string
func exit(x int) {
	os.Exit(x)
}

func main() {
	var s1 string
	s1 = "hello world\n"
	gstring = s1
	print(gstring)
	var i int
	gvar1 = 10
	gvar1 = gvar1 + 10
	gvar2 = gvar1 + 1
	i = gvar2 + 1
	exit(i)
}
