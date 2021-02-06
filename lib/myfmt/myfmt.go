package myfmt

import (
	"github.com/DQNEO/babygo/lib/strconv"
	"reflect"
)
import "syscall"

func Sprintf(format string, a ...interface{}) string {
	var r []uint8
	var inPercent bool
	var argIndex int

	//syscall.Write(1, []uint8("# @@@ Sprintf start. format=" + format + "\n"))

	for _, c := range []uint8(format) {
		//syscall.Write(1, []uint8{c})
		if inPercent {
			//syscall.Write(1, []uint8("@inPercent@"))
			if c == '%' { // "%%"
				r = append(r, '%')
			} else {
				arg := a[argIndex]
				var sign uint8 = c
				var str string
				switch sign {
				case 's':
					switch _arg := arg.(type) {
					case string:
						str = _arg
					case int:
						str = strconv.Itoa(_arg) // %!s(int=123)
					}
					for _, _c := range []uint8(str) {
						r = append(r, _c)
					}
				case 'd':
					switch _arg := arg.(type) {
					case string:
						str = _arg // "%!d(string=" + _arg + ")"
					case int:
						str = strconv.Itoa(_arg)
					}
					for _, _c := range []uint8(str) {
						r = append(r, _c)
					}
				case 'T':
					t := reflect.TypeOf(arg)
					str = t.String()
					for _, _c := range []uint8(str) {
						r = append(r, _c)
					}
				default:
					panic("Unknown format:" + strconv.Itoa(int(sign)))
				}
				argIndex++
			}
			inPercent = false
		} else {
			if c == '%' {
				inPercent = true
			} else {
				r = append(r, c)
			}
		}
	}

	return string(r)
}

// package fmt
func Printf(format string, a ...interface{}) {
	var s = Sprintf(format, a...)
	syscall.Write(1, []uint8(s))
}

