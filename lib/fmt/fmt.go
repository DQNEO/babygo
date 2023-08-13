package fmt

import (
	"io"
	"os"
	"reflect"

	"github.com/DQNEO/babygo/lib/strconv"
)

func Sprintf(format string, a ...interface{}) string {
	var r []uint8
	var inPercent bool
	var argIndex int

	for _, c := range []uint8(format) {
		if inPercent {
			if c == '%' { // "%%"
				r = append(r, '%')
			} else {
				arg := a[argIndex]
				var sign uint8 = c
				var str string
				switch sign {
				case '#':
					// skip for now
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
				case 'd', 'p': // %d
					switch _arg := arg.(type) {
					case string: // ("%d", "xyz")
						str = "%!d(string=" + _arg + ")" // %!d(string=xyz)
					case int: // ("%d", 123)
						str = strconv.Itoa(_arg)
					case uintptr: // ("%d", 123)
						intVal := int(_arg)
						str = strconv.Itoa(intVal)
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
					panic("Sprintf: Unknown format:" + string([]uint8{uint8(sign)}))
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

func Fprint(w io.Writer, a string) {
	b := []byte(a)
	w.Write(b)
}

func Fprintln(w io.Writer, a string) {
	b := []byte(a)
	b = append(b, '\n')
	w.Write(b)
}

func Print(a string) {
	Fprint(os.Stdout, a)
}

func Println(a string) {
	Fprintln(os.Stdout, a)
}

func Printf(format string, a ...interface{}) {
	Fprintf(os.Stdout, format, a...)
}

func Fprintf(w io.Writer, format string, a ...interface{}) {
	var s = Sprintf(format, a...)
	w.Write([]uint8(s))
}
