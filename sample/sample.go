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

func main() {
	min()
	arg1(1)
	arg1ret1(2)
}
