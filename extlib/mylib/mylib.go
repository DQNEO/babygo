package mylib

import "syscall"

func Sum(a int, b int) int {
	return a + b
}

// package strconv
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

// package strings

// "foo/bar", "/" => []string{"foo", "bar"}
func Split(s string, ssep string) []string {
	if len(ssep) > 1 {
		panic("no supported")
	}
	sepchar := ssep[0]
	var buf []byte
	var r []string
	for _, b := range []byte(s) {
		if b == sepchar {
			r = append(r, string(buf))
			buf = nil
		} else {
			buf = append(buf, b)
		}
	}
	r = append(r, string(buf))

	return r
}

func HasPrefix(s string, prefix string) bool {
	for i, bp := range []byte(prefix) {
		if bp != s[i] {
			return false
		}
	}
	return true
}

func HasSuffix(s string, suffix string) bool {
	if len(s) >= len(suffix) {
		var low int = len(s) - len(suffix)
		var lensb int = len(s)
		var suf []byte
		sb := []byte(s)
		suf = sb[low:lensb] // lensb is required
		return eq2(suf, []byte(suffix))
	}
	return false
}

func eq2(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Contains reports whether substr is within s.
func Contains(s string, substr string) bool {
	return Index(s, substr) >= 0
}

func Index(s string, substr string) int {
	var in bool
	var subIndex int
	var r int = -1 // not found
	for i, b := range []byte(s) {
		if !in && b == substr[0] {
			in = true
			r = i
			subIndex = 0
		}

		if in {
			if b == substr[subIndex] {
				if subIndex == len(substr)-1 {
					return r
				}
			} else {
				in = false
				r = -1
				subIndex = 0
			}
		}
	}

	return -1
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

// package path
// "foo/bar/buz" => "foo/bar"
func Dir(path string) string {
	if len(path) == 0 {
		return "."
	}

	if path == "/" {
		return "/"
	}

	found := LastIndexByte(path, '/')
	if found == -1 {
		// not found
		return path
	}

	return path[:found]
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

// package fmt
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
