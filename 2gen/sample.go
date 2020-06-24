package main

import "os"

func exit(x int) {
	os.Exit(x)
}

func main() {
	var i int
	var x uint8
	x = 10
	i = int(x)
	exit(i)
}
