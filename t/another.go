package main

import "syscall"
import "github.com/DQNEO/babygo/lib/strconv"

func anotherFunc() string {
	return anotherVar
}

func nop()  {}
func nop1() {}
func nop2() {}

// --- utils ---
func write(x interface{}) {
	var s string
	switch xx := x.(type) {
	case int:
		s = strconv.Itoa(xx)
	case string:
		s = xx
	}
	var slc []uint8 = []uint8(s)
	syscall.Write(1, slc)
}

func writeln(s interface{}) {
	write(s)
	write("\n")
}
