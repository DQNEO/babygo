package main

import "os"
import "syscall"


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

func add1(x int) int {
	return x + 1
}

var globalint int
var globalint2 int
var globaluint8 uint8
var globaluint16 uint16
var globaluintptr uintptr

func write(s string) {
	var slc []uint8 = []uint8(s)
	syscall.Write(1, slc)
}

func testString() {
	var s string = "hello string\n"
	write(s)
}

func testMisc() {
	var i13 int = 0
	i13 = testArgAssign(i13)
	var i5 int = testMinus()
	globalint2 = sum(1, i13 * i5)
	var locali3 int
	var tmp int
	tmp = int(uint8('3' - '1'))
	tmp = tmp + int(globaluint16)
	tmp = tmp + int(globaluint8)
	tmp = tmp + int(globaluintptr)
	locali3 = add1(tmp)
	var i42 int
	i42 = sum(globalint, globalint2) + locali3
	exit(i42)
}

func test() {
	testString()
	testMisc()
}

func main() {
	test()
}
