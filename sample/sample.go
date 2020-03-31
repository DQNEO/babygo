// How to compile:
// go tool compile -N -S sample.go
package main

func min() {

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
	arg1(1)
	arg1ret1(2)
	sum(2, 3)
	concate("hello", " world")
}
