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
	return 5
}

func testMisc() {
	var i13 int = 12
	i13 = testArgAssign(i13)
	var i5 int = testMinus()
	exit(i5)
}

func test() {
	testMisc()
}

func main() {
	test()
}
