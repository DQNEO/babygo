package mylib

import "syscall"

const DUMMY string = "dummy"

func Sum(a int, b int) int {
	return a + b
}

func Itoa(ival int) string {
	if ival == 0 {
		return "0"
	}

	var buf = make([]uint8, 100, 100)
	var r = make([]uint8, 100, 100)

	var next int
	var right int
	var ix = 0
	var minus bool
	minus = false
	for ix = 0; ival != 0; ix = ix + 1 {
		if ival < 0 {
			ival = -1 * ival
			minus = true
			r[0] = '-'
		} else {
			next = ival / 10
			right = ival - next*10
			ival = next
			buf[ix] = uint8('0' + right)
		}
	}

	var j int
	var c uint8
	for j = 0; j < ix; j = j + 1 {
		c = buf[ix-j-1]
		if minus {
			r[j+1] = c
		} else {
			r[j] = c
		}
	}

	return string(r[0:ix])
}

// search index of the specified char from backward
func LastIndexByte(s string, c uint8) int {
	for i:=len(s)-1;i>=0;i-- {
		if s[i] == c {
			return i
		}
	}
	// not found
	return -1
}

// "foo/bar/buz" => "buz"
func Base(path string) string {
	if len(path) == 0 {
		return "."
	}

	if path == "/" {
		return "/"
	}

	if path[len(path) - 1] == '/' {
		path = path[0:len(path) - 1]
	}

	found := LastIndexByte(path, '/')
	if found == -1 {
		// not found
		return path
	}

	_len := len(path)
	r := path[found+1:_len]
	return r
}

func Atoi(gs string) int {
	if len(gs) == 0 {
		return 0
	}
	var n int

	var isMinus bool
	for _, b := range []uint8(gs) {
		if b == '.' {
			return -999 // @FIXME all no number should return error
		}
		if b == '-' {
			isMinus = true
			continue
		}
		var x uint8 = b - uint8('0')
		n = n * 10
		n = n + int(x)
	}
	if isMinus {
		n = -n
	}

	return n
}

func Printf(format string, a ...interface{}) {
	var s = Sprintf(format, a...)
	syscall.Write(1, []uint8(s))
}

func Sprintf(format string, a...interface{}) string {
	var buf []uint8
	var inPercent bool
	var argIndex int
	for _, c := range []uint8(format) {
		if inPercent {
			if c == '%' {
				buf = append(buf, c)
			} else {
				arg := a[argIndex]
				var str string
				switch _arg := arg.(type) {
				case string:
					str = _arg
				case int:
					str = Itoa(_arg)
				}
				argIndex++
				for _, _c := range []uint8(str) {
					buf = append(buf, _c)
				}
			}
			inPercent = false
		} else {
			if c == '%' {
				inPercent = true
			} else {
				buf = append(buf, c)
			}
		}
	}

	return string(buf)
}
