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

type printer struct {
	buf []byte
}

func newPrinter() *printer {
	return &printer{}
}

func (p *printer) doPrint(a []interface{}) {
	for _, i := range a {
		s, ok := i.(string)
		if !ok {
			panic("only string is supported")
		}
		bytes := []byte(s)
		for _, b := range bytes {
			p.buf = append(p.buf, b)
		}
	}
}

func (p *printer) doPrintln(a []interface{}) {
	for _, i := range a {
		s, ok := i.(string)
		if !ok {
			panic("only string is supported")
		}
		bytes := []byte(s)
		for _, b := range bytes {
			p.buf = append(p.buf, b)
		}
	}
	p.buf = append(p.buf, '\n')
}

func (p *printer) free() {
	p.buf = nil
}

func Fprint(w io.Writer, a ...interface{}) (int, error) {
	p := newPrinter()
	p.doPrint(a)
	n, err := w.Write(p.buf)
	p.free()
	return n, err
}

func Fprintln(w io.Writer, a ...interface{}) (int, error) {
	p := newPrinter()
	p.doPrintln(a)
	n, err := w.Write(p.buf)
	p.free()
	return n, err
}

func Print(a ...interface{}) (int, error) {
	n, err := Fprint(os.Stdout, a...)
	return n, err
}

func Println(a ...interface{}) (int, error) {
	n, err := Fprintln(os.Stdout, a...)
	return n, err
}

func Printf(format string, a ...interface{}) (int, error) {
	n, err := Fprintf(os.Stdout, format, a...)
	return n, err
}

func Fprintf(w io.Writer, format string, a ...interface{}) (int, error) {
	var s = Sprintf(format, a...)
	n, err := w.Write([]uint8(s))
	return n, err
}
