package fmt

import (
	"github.com/DQNEO/babygo/lib/strconv"
	"reflect"
	"os"
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
				case 's': // %s
					switch _arg := arg.(type) {
					case string: // ("%s", "xyz")
						str = _arg
					case int: // ("%s", 123)
						strNumber := strconv.Itoa(_arg)
						str = "%!s(int=" + strNumber + ")" // %!s(int=123)
					default:
						str = "unknown type"
					}
					for _, _c := range []uint8(str) {
						r = append(r, _c)
					}
				case 'd': // %d
					switch _arg := arg.(type) {
					case string: // ("%d", "xyz")
						str = "%!d(string=" + _arg + ")" // %!d(string=xyz)
					case int: // ("%d", 123)
						str = strconv.Itoa(_arg)
					default:
						str = "unknown type"
					}
					for _, _c := range []uint8(str) {
						r = append(r, _c)
					}
				case 'T':
					t := reflect.TypeOf(arg)
					if t == nil {
						// ?
					} else {
						str = t.String()
					}
					for _, _c := range []uint8(str) {
						r = append(r, _c)
					}
				default:
					panic("Sprintf: Unknown format:" + strconv.Itoa(int(sign)))
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

func Printf(format string, a ...interface{}) {
	var s = Sprintf(format, a...)
	syscall.Write(1, []uint8(s))
}

func Fprintf(w *os.File, format string, a ...interface{}) {
	var s = Sprintf(format, a...)
	w.Write([]uint8(s))
}

