// How to compile:
// go tool compile -N -S -l sample.go > sample.s
package main

func multiVars(a uint8, b uint8, c uint8) (uint8, uint8, uint8) {
	return a, b, c
}

func receiveBytes(a uint8, b uint8, c uint8) uint8 {
	var r uint8 = a
	return r
}

func testPassBytes() {
	var a uint8 = 'a'
	var b uint8 = 'b'
	var c uint8 = 'c'
	d := receiveBytes(a, b, c)
	println(d)
}

func char(a byte) byte {
	b := a
	b = 'A'
	return b
}

func slice(a []byte) []byte {
	var b []byte = a
	return b
}

func arg1(x int) {

}

func arg1ret1(x int) int {
	return x
}

func sumAndMul(x int, y int) (int, int) {
	return x + y, x * y
}

func sumAndMulWithNamedReturn(x int, y int) (sum int, mul int) {
	sum = x + y
	mul = x * y
	return
}

func sum(x int, y int) int {
	return x + y
}

func concate(x string, y string) string {
	return x + y
}

func main() {
	multiVars(1,2, 3)
	slice(nil)
	slice([]byte{'a', 'b', 'c'})
	arg1(1)
	arg1ret1(2)
	sumAndMul(5, 7)
	sumAndMulWithNamedReturn(5, 7)
	sum(2, 3)
	concate("hello", " world")
}
