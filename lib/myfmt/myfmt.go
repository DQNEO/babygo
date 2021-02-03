package myfmt

import "github.com/DQNEO/babygo/lib/strconv"
import "syscall"


func Sprintf(format string, a ...interface{}) string {
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
					str = strconv.Itoa(_arg)
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

// package fmt
func Printf(format string, a ...interface{}) {
	var s = Sprintf(format, a...)
	syscall.Write(1, []uint8(s))
}

