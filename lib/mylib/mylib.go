package mylib

import "github.com/DQNEO/babygo/lib/mylib2"

func Sum(a int, b int) int {
	return a + b
}

func Sum2(a int, b int) int {
	return mylib2.Sum2(a, b)
}

