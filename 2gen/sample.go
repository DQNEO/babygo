package main

import "os"

func exit(x int) {
	os.Exit(x)
}

func sum(x int, y int) {
	os.Exit(x)
}

func main() {
	var s1 string
	s1 = "hello world\n"
	print(s1)
	var i int
	i = 7
	exit(i)
}
