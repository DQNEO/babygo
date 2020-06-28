package main

import "os"

func exit(x int) {
	os.Exit(x)
}

func testArgAssign(x int) int {
	x = 13
	return x
}

func testVoid() {
	return
}

func testMisc() {
	//testVoid()
	var i int = 12
	//i = testArgAssign(i)
	exit(i)
}

func test() {
	testMisc()
}

func main() {
//	test()
	var i int = 14
	var j int
	exit(i + j)
}
