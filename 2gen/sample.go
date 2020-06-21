package main

import "os"

func exit(x int) {
	os.Exit(x)
}

func sum(x int, y int) {
	os.Exit(x)
}

func main() {
	exit(3)
	//var i1 int
	//i1 = 7
	//sum(i1, 5)
}
