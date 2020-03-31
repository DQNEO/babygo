// How to compile:
// go tool compile -N -S sample.go
package main

func min() {
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

func sum(x int, y int ) int {
	return x + y
}

func concate(x string, y string) string {
	return x + y
}

func main() {
	min()
	slice(nil)
	slice([]byte{'a','b','c'})
	arg1(1)
	arg1ret1(2)
	sum(2, 3)
	concate("hello", " world")
}
