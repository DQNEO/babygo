package main

import "os"

func exit(x int) {
	os.Exit(x)
}

func testArgAssign(x int) int {
	x = 13
	return x
}

func testMinus() int {
	var x int = -1
	x = x * -5
	return x
}

func sum(x int, y int) int {
	return x + y
}

var globalint2 int

func testMisc() {
	var i13 int = 0
	i13 = testArgAssign(i13)
	var i5 int = testMinus()
	globalint2 = sum(1, i13 * i5)
	exit(globalint2)
}

func test() {
	testMisc()
}

func main() {
	test()
}
