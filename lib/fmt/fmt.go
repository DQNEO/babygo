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
	bt := []byte(s)
	for _, c := range bt {
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
	p.doPrintf(format, a)
	n, err := w.Write(p.buf)
	p.free()
	return n, err
}

func Printf(format string, a ...interface{}) (int, error) {
	return Fprintf(os.Stdout, format, a...)
}

func Sprintf(format string, a ...interface{}) string {
	p := newPrinter()
	p.doPrintf(format, a)
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

func (p *pp) fmtString(v string, verb byte) {
	switch verb {
	case 'v', 's':
		p.buf.writeString(v)
	case 'd', 'p':
		str := "%!d(string=" + v + ")" // %!d(string=xyz)
		p.buf.writeString(str)
	default:
		panic("*pp.fmtString: TBI")
	}
}

func (p *pp) fmtInteger(v int, verb byte) {
	switch verb {
	case 's':
		strNumber := strconv.Itoa(v)
		str := "%!s(int=" + strNumber + ")" // %!s(int=123)
		p.buf.writeString(str)
	case 'v', 'd', 'p':
		str := strconv.Itoa(v)
		p.buf.writeString(str)
	default:
		panic("*pp.fmtInteger: TBI")
	}
}

func (p *pp) printArg(arg interface{}, verb byte) {
	switch verb {
	case 'T':
		typ := reflect.TypeOf(arg)
		var str string
		if typ == nil {
			str = "nil"
		} else {
			str = typ.String()
		}
		p.buf.writeString(str)
		return
	}

	// Some types can be done without reflection.
	switch f := arg.(type) {
	case int:
		p.fmtInteger(f, verb)
	case uintptr:
		p.fmtInteger(int(f), verb)
	case string:
		p.fmtString(f, verb)
	default:
		p.buf.writeString("unknown type")
	}
}

func (p *pp) doPrintf(format string, a []interface{}) {
	var inPercent bool
	var argNum int

	for _, c := range []uint8(format) {
		if inPercent {
			if c == '%' { // "%%"
				p.buf.writeByte('%')
			} else {
				p.printArg(a[argNum], c)
				argNum++
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
