package fmt

import (
	"io"
	"os"
	"reflect"

	"github.com/DQNEO/babygo/lib/strconv"
)

type buffer []byte

func (b *buffer) write(p []byte) {
	for _, c := range p {
		*b = append(*b, c)
	}
}

func (b *buffer) writeString(s string) {
	for _, c := range []byte(s) {
		*b = append(*b, c)
	}
}

func (b *buffer) writeByte(c byte) {
	*b = append(*b, c)
}

type pp struct {
	buf buffer
}

func newPrinter() *pp {
	return &pp{}
}

func Fprintf(w io.Writer, format string, a ...interface{}) (int, error) {
	p := newPrinter()
	p.doPrintf(format, a...)
	n, err := w.Write(p.buf)
	p.free()
	return n, err
}

func Printf(format string, a ...interface{}) (int, error) {
	return Fprintf(os.Stdout, format, a...)
}

func Sprintf(format string, a ...interface{}) string {
	p := newPrinter()
	p.doPrintf(format, a...)
	s := string(p.buf)
	p.free()
	return s
}

func Fprint(w io.Writer, a ...interface{}) (int, error) {
	p := newPrinter()
	p.doPrint(a)
	n, err := w.Write(p.buf)
	p.free()
	return n, err
}

func Print(a ...interface{}) (int, error) {
	return Fprint(os.Stdout, a...)
}

func Fprintln(w io.Writer, a ...interface{}) (int, error) {
	p := newPrinter()
	p.doPrintln(a)
	n, err := w.Write(p.buf)
	p.free()
	return n, err
}

func Println(a ...interface{}) (int, error) {
	return Fprintln(os.Stdout, a...)
}

func (p *pp) printArg(arg interface{}, verb byte) {
	switch f := arg.(type) {
	case string:
		bytes := []byte(f)
		for _, b := range bytes {
			p.buf.writeByte(b)
		}
	case int:
		str := strconv.Itoa(f)
		bytes := []byte(str)
		for _, b := range bytes {
			p.buf.writeByte(b)
		}
	default:
		panic("TBI:pp.printArg")
	}
}

func (p *pp) doPrintf(format string, a ...interface{}) {
	var inPercent bool
	var argIndex int

	for _, c := range []uint8(format) {
		if inPercent {
			if c == '%' { // "%%"
				p.buf.writeByte('%')
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
					p.buf.writeString(str)
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
					p.buf.writeString(str)
				case 'T':
					t := reflect.TypeOf(arg)
					if t == nil {
						// ?
					} else {
						str = t.String()
					}
					p.buf.writeString(str)
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
				p.buf.writeByte(c)
			}
		}
	}
}

func (p *pp) free() {
	p.buf = nil
}

func (p *pp) doPrint(a []interface{}) {
	for _, arg := range a {
		p.printArg(arg, 'v')
	}
}

func (p *pp) doPrintln(a []interface{}) {
	for argNum, arg := range a {
		if argNum > 0 {
			p.buf.writeByte(' ')
		}
		p.printArg(arg, 'v')
	}
	p.buf.writeByte('\n')
}
