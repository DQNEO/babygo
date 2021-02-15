package main

import (
	"os"
	"reflect"
	"syscall"

	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/path"
	"github.com/DQNEO/babygo/lib/strconv"
	"github.com/DQNEO/babygo/lib/strings"
)

func returnMultiValues(a uint8, b uint8, c uint8) (uint8, uint8, uint8) {
	return a, b, c
}

func testReturnMultiValues() {
	var a uint8
	var b uint8
	var c uint8
	a, b, c = returnMultiValues('l', 'm', 'n')
	fmt.Printf("abc=%s\n", string([]uint8{a,b,c}))
	os.Exit(0)
}

func receiveBytes(a uint8, b uint8, c uint8) uint8 {
	var r uint8 = c
	return r
}

func testPassBytes() {
	var a uint8 = 'a'
	var b uint8 = 'b'
	rc := receiveBytes(a, b, 'c')
	buf := []uint8{rc}
	s := string(buf)
	fmt.Printf("s=%s\n", s)
}

func testSprinfMore() {
	var s string
	s = fmt.Sprintf("hello")
	writeln(s)
	s = fmt.Sprintf("%%rax")
	writeln(s)

	var i int = 1234
	s = fmt.Sprintf("number %d", i)
	writeln(s)

	var str = "I am string"
	s = fmt.Sprintf("string %s", str)
	writeln(s)

	s = fmt.Sprintf("types are %T", str)
	writeln(s)

	s = fmt.Sprintf("types are %T", i)
	writeln(s)

	s = fmt.Sprintf("types are %T", &i)
	writeln(s)

	s = fmt.Sprintf("%d", "xyz")
	writeln(s)

	var a []interface{} = make([]interface{}, 3, 3)
	a[0] = 1234
	a[1] = "c"
	a[2] = "efg"
	s = fmt.Sprintf("%dab%sd%s", a...)
	writeln(s)

	// type mismatch cases
	s = fmt.Sprintf("string %d", str)
	writeln(s)

	s = fmt.Sprintf("%s", 123)
	writeln(s)

}

var anotherVar string = "Another Hello\n"

func testAnotherFile() {
	anotherFunc()
}


func testSortStrings() {
	ss := []string{
		// sample strings
		"github.com/DQNEO/babygo/lib/strings",
		"unsafe",
		"reflect",
		"github.com/DQNEO/babygo/lib/fmt",
		"github.com/DQNEO/babygo/lib/mylib2",
		"github.com/DQNEO/babygo/lib/strconv",
		"syscall",
		"github.com/DQNEO/babygo/lib/mylib",
		"github.com/DQNEO/babygo/lib/path",
		"os",
	}

	fmt.Printf("--------------------------------\n")
	for _, s := range ss {
		fmt.Printf("%s\n", s)
	}
	mylib.SortStrings(ss)
	fmt.Printf("--------------------------------\n")
	for _, s := range ss {
		fmt.Printf("%s\n", s)
	}
}

func testGetdents64() {
	dirents := mylib.GetDirents("t")
	var counter int
	for _, dirent := range dirents {
		if dirent == "." || dirent == ".." {
			continue
		}
		counter++
	}
	fmt.Printf("%d\n", counter)
}

func testEnv() {
	var gopath string = os.Getenv("FOO")
	fmt.Printf("env FOO=%s\n", gopath)
}

func testReflect() {
	var i int = 123
	rt := reflect.TypeOf(i)
	fmt.Printf("%s\n", rt.String()) // int

	rt = reflect.TypeOf(&i)
	fmt.Printf("%s\n", rt.String()) // *int

	rt = reflect.TypeOf("hello")
	fmt.Printf("%s\n", rt.String()) // string

	myStruct := MyStruct{}
	rt = reflect.TypeOf(myStruct)
	fmt.Printf("%s\n", rt.String()) // main.MyStruct

	myStructP := &MyStruct{}
	rt = reflect.TypeOf(myStructP)
	fmt.Printf("%s\n", rt.String()) // *main.MyStruct
}

func returnSlice() []string {
	r := []string{"aa", "bb", "cc"}
	return r
}

func testReturnSlice() {
	slice := returnSlice()
	write(slice[0])
	write(slice[1])
	writeln(slice[2])
}

func testStrings() {
	// Split
	s := strings.Split("foo/bar", "/")
	fmt.Printf("%d\n", len(s)+1) // 3
	fmt.Printf("%s\n", s[0])     // foo
	fmt.Printf("%s\n", s[1])     // bar

	target := "foo bar buz"
	if !strings.HasPrefix(target, "foo") {
		panic("error")
	}

	if strings.HasPrefix(target, " ") {
		panic("error")
	}

	if strings.HasPrefix(target, "buz") {
		panic("error")
	}

	s2 := "main.go"
	suffix := ".go"
	if strings.HasSuffix(s2, suffix) {
		fmt.Printf("1\n")
	} else {
		panic("ERROR")
	}

	if strings.Contains("foo/bar", "/") {
		fmt.Printf("ok\n")
	} else {
		panic("ERROR")
	}
}

// https://golang.org/ref/spec#Slice_expressions
func testSliceExpr() {
	a := [5]int{1, 2, 3, 4, 5}
	var s []int
	s = a[1:4]
	for _, elm := range s {
		write(elm)
	}
	write(len(s))
	write(cap(s))
	writeln("")

	s = a[:2]
	for _, elm := range s {
		write(elm)
	}
	write(len(s))
	write(cap(s))
	writeln("")

	s = a[1:]
	for _, elm := range s {
		write(elm)
	}
	write(len(s))
	write(cap(s))
	writeln("")

	s = a[:]
	for _, elm := range s {
		write(elm)
	}
	write(len(s))
	write(cap(s))
	writeln("")

	str := "12345"
	str2 := str[:]
	writeln(str2)
	writeln(str[1:])
	writeln(str[:4])
	writeln(str[1:3])
}

func testPath() {
	// Copied from https://golang.org/pkg/path/#Base
	writeln(path.Base("/a/b"))
	writeln(path.Base("/"))
	writeln(path.Base(""))

	// Copied from https://golang.org/pkg/mylib/#Dir
	writeln(path.Dir("/a/b/c"))
	writeln(path.Dir("a/b/c"))
	writeln(path.Dir("/a/"))
	writeln(path.Dir("a/"))
	writeln(path.Dir("/"))
	writeln(path.Dir(""))
}

func testByteType() {
	var b byte = uint8('x')
	var s string = string([]byte{b})
	writeln(s)
}

func testExtLib() {
	y := mylib.Sum2(3, 4)
	write("# testExtLib() => ")
	writeln(strconv.Itoa(y))
}

func passVargs(a ...interface{}) {
	takeInterfaceVaargs(a...)
}

func takeStringVaargs(a ...string) {
	writeln(len(a))
	for _, s := range a {
		writeln(s)
	}
}

func testExpandSlice() {
	var slicexxx = []interface{}{2, 4, 6}
	nop()
	passVargs(slicexxx...)

	takeStringVaargs("foo", "bar", "buz")
}

var gArrayForFullSlice [3]int

func testFullSlice() {
	gArrayForFullSlice = [3]int{
		2,
		4,
		6,
	}

	for _, i := range gArrayForFullSlice {
		writeln(i)
	}

	fullSlice := gArrayForFullSlice[1:2:3]
	writeln(cap(fullSlice))

	for _, i := range fullSlice {
		writeln(i)
	}
}

func takeInterfaceVaargs(a ...interface{}) {
	writeln(len(a))
	for _, ifc := range a {
		writeln(ifc)
	}
}

func testInterfaceVaargs() {
	var i = 1419
	var s = "s1419"
	takeInterfaceVaargs(i, s)
}

var gefacearray [3]interface{}

func returnInterface() interface{} {
	return 14
}

func testConvertToInterface() {
	// Explicit conversion
	var ifc interface{} = interface{}(7)
	var i int
	i, _ = ifc.(int)
	writeln(i)

	// Implicit conversion to variable
	var j int = 8
	ifc = j
	i, _ = ifc.(int)
	writeln(i)

	// Implicit conversion to struct field
	var k int = 9
	var strct MyStruct
	strct.ifc = k
	i, _ = strct.ifc.(int)
	writeln(i)

	// Implicit conversion to array element
	var l int = 10
	gefacearray[2] = l
	i, _ = gefacearray[2].(int)
	writeln(i)

	// Implicit conversion to struct literal
	k = 11
	strct = MyStruct{
		field1: 1,
		field2: 2,
		ifc:    k,
	}
	i, _ = strct.ifc.(int)
	writeln(i)

	// Implicit conversion to array literal
	gefacearray = [3]interface{}{
		0,
		0,
		12,
	}
	i, _ = gefacearray[2].(int)
	writeln(i)

	// Implicit conversion to function call
	l = 13
	takeInterface(l)

	// Implicit conversion to return
	ifc = returnInterface()
	i, _ = ifc.(int)
	writeln(i)

	// Implicit conversion to function call (vaargs)
}

func testTypeSwitch() {
	var ifc EmptyInterface
	var i int = 7
	ifc = &i
	switch ifc.(type) {
	case *int:
		x := ifc.(*int)
		writeln("type is *int")
		writeln(strconv.Itoa(*x))
	case string:
		x := ifc.(string)
		writeln("type is string")
		writeln(x)
	default:
		panic("ERROR")
	}

	var s string = "abcde"
	ifc = s
	switch xxx := ifc.(type) {
	case int:
		writeln("type is int")
		var zzzzz = xxx // test inference of xxx
		writeln(zzzzz)
	case string:
		writeln("type is string")
		writeln(xxx)
	default:
		panic("ERROR")
	}

	var srct MyStruct = MyStruct{
		field1: 111,
		field2: 222,
	}
	ifc = srct
	switch yyy := ifc.(type) {
	case MyStruct:
		writeln("type is MySruct")
		writeln(strconv.Itoa(yyy.field2))
	default:
		panic("ERROR")
	}
	ifc = true
	switch ifc.(type) {
	case int:
		panic("ERROR")
	case string:
		panic("ERROR")
	default:
		writeln("type is bool")
	}
}

func makeInterface() interface{} {
	var r interface{} = 1829
	return r
}

func testGetInterface() {
	var x interface{} = makeInterface()
	var i int
	var ok bool
	i, ok = x.(int)
	if ok {
		writeln(i)
	}
}

func takeInterface(ifc interface{}) {
	var s int
	var ok bool
	s, ok = ifc.(int)
	if ok {
		writeln(s)
	}
}

func testPassInterface() {
	var s int = 1537
	var ifc interface{} = s
	takeInterface(ifc)
}

type EmptyInterface interface {
}

func testInterfaceAssertion() {
	var ifc interface{}
	var ifc2 EmptyInterface

	var ok bool
	var i int = 20210124
	ifc = i
	writeln("aaaa")
	i, ok = ifc.(int)
	if ok {
		writeln(" type matched")
		write(strconv.Itoa(i))
	} else {
		panic("FAILED")
	}

	var j int
	j = ifc.(int)
	write(strconv.Itoa(j))
	_, ok = ifc.(int) // error!!!
	if ok {
		writeln("ok")
	} else {
		panic("FAILED")
	}

	i, _ = ifc.(int)
	write(strconv.Itoa(i))

	_, _ = ifc.(int)
	writeln("end of testInterfaceAssertion")
	var s string
	s, ok = ifc.(string)
	writeln(s)
	if ok {
		panic("FAILED")
	} else {
		writeln(s)
	}
	s = "I am string"
	ifc2 = s
	s, ok = ifc2.(string)
	writeln(s)
	if ok {
		writeln("ok")
	} else {
		panic("FAILED")
	}

	ifc2 = nil
	s, ok = ifc2.(string)
	if ok {
		panic("FAILED")
	} else {
		writeln(s)
	}
}

func testInterfaceimplicitConversion() {
	var eface interface{}
	var i int = 7
	eface = i
	geface = eface
	writeln("1111")
	if geface == eface {
		writeln("eface match")
	}
	writeln("22222")
	var x **[1][]*int
	writeln("3333")
	eface = x
	writeln("4444")
	if geface != eface {
		writeln("eface not match")
	}
	var m MyType
	eface = m
	if geface != eface {
		writeln("eface not match")
	}
}

var geface interface{}

func testInterfaceZeroValue() {
	var eface interface{}
	if eface == nil {
		writeln("eface is nil")
	}
	if geface == nil {
		writeln("geface is nil")
	}
	geface = eface
	if geface == nil {
		writeln("geface is nil")
	}
	eface = geface
	if eface == nil {
		writeln("eface is nil")
	}
}

func testForRangeShortDecl() {
	var ary []int = []int{3, 2, 1, 0}

	for _, w := range ary {
		write(strconv.Itoa(w))
	}
	writeln("")

	var i int = 20210122
	for i, www := range ary {
		write(strconv.Itoa(i))
		write(strconv.Itoa(www))
	}
	writeln("")

	writeln(i)
}

var gi = 123   // int
var gs = "abc" // string
var gstrctPtr = &MyStruct{
	field2: 456,
}

func testInferVarTypes() {
	// type by literal
	var gstrct = MyStruct{
		field1: 789,
	}

	var gslc = []int{1, 2, 3}

	writeln(gi)
	writeln(gs)
	writeln(strconv.Itoa(gslc[2]))
	writeln(strconv.Itoa(gstrctPtr.field2))
	writeln(strconv.Itoa(gstrct.field1))
}

var gInt int = 1010
var gBool bool = true
var gString string = "gString"
var gPointer *MyStruct = &MyStruct{
	field1: 11,
	field2: 22,
}

var gChar uint8 = 'A'

func testGlobalValues() {
	writeln(gInt)
	if gBool {
		writeln("gBool is true")
	}
	writeln(gString)
	writeln(strconv.Itoa(gPointer.field2))
	writeln(strconv.Itoa(int(gChar)))
}

func testShortVarDecl() {
	x := 123
	writeln(x)

	var p = &MyStruct{
		field1: 10,
	}

	f1 := p.getField1() // infer method return type
	writeln(f1)

	s := "infer string literal"
	writeln(s)

	i := 3 + 5
	j := i
	writeln(j)
}

func testStructPointerMethods() {
	var p = &MyStruct{
		field1: 10,
	}

	var f1 = p.getField1() // infer method return type
	writeln(f1)
	p.setField1(20)
	writeln(strconv.Itoa(p.getField1()))
}

func (p *MyStruct) getField1() int {
	return p.field1
}

func (p *MyStruct) setField1(x int) {
	p.field1 = x
}

type T int

//type MV interface {
//	mv(int)
//}
//
//type MP interface {
//	mp(int)
//}

func (v T) mv(a int) {
	v = T(a)
}

func (p *T) mp(a int) {
	*p = T(a)
}

//func testBasicMethodCalls() {
//	var v T = 1
//	writeln(mylib.Itoa(int(v)))
//	v.mv(2) // ordinary
//	writeln(mylib.Itoa(int(v)))
//	v.mp(3) // (&v).mp()
//	writeln(mylib.Itoa(int(v)))
//
//	var p *T = &v
//	p.mp(4) // ordinary
//	writeln(mylib.Itoa(int(v)))
//	p.mv(5) // (*p).mv()
//	writeln(mylib.Itoa(int(v)))
//}

func testPointerMethod() {
	var v T = 1
	v.mv(2)
	writeln(strconv.Itoa(int(v)))

	var p *T = &v
	p.mp(3)
	writeln(strconv.Itoa(int(v)))
}

type MyAnotherType int

func (x MyAnotherType) add10() int {
	return int(x) + 10
}

func testMethodAnother() {
	var x MyAnotherType = 10
	var y int = x.add10()
	writeln(y)
}

type MyType int

func add10(x MyType) int {
	return int(x) + 10
}

func (x MyType) add10() int {
	return int(x) + 10
}

func testMethodSimple() {
	var x MyType = 4
	writeln(strconv.Itoa(x.add10()))
	writeln(strconv.Itoa(add10(x)))
}

func testOsArgs() {
	if len(os.Args) > 1 {
		writeln(os.Args[1])
	}
}

func testStructLiteralWithContents() {
	var strct = MyStruct{
		field1: 10,
		field2: 20,
	}
	writeln(strconv.Itoa(strct.field1))
	writeln(strconv.Itoa(strct.field2))

	var strct2 = MyStruct{
		field2: 20,
	}
	writeln(strconv.Itoa(strct2.field1))
	writeln(strconv.Itoa(strct2.field2))

	var strctp = &MyStruct{
		field1: 30,
		field2: 40,
	}
	writeln(strconv.Itoa(strctp.field1))
	writeln(strconv.Itoa(strctp.field2))
}

func returnPointerOfStruct() *MyStruct {
	var strct *MyStruct = &MyStruct{}
	strct.field1 = 345
	strct.field2 = 678
	return strct
}

func testAddressOfStructLiteral() {
	var strct *MyStruct = returnPointerOfStruct()
	writeln(strconv.Itoa(strct.field1))
	writeln(strconv.Itoa(strct.field2))
}

func testStructCopy() {
	var strct MyStruct = MyStruct{}
	strct.field1 = 123
	strct.field2 = 456

	var strct2 MyStruct = MyStruct{}
	strct2 = strct

	writeln(strconv.Itoa(strct2.field1))
	writeln(strconv.Itoa(strct2.field2))

	// assert 2 struct does not share memory
	strct2.field1 = 789
	writeln(strconv.Itoa(strct.field1))
}

func testStructLiteral() {
	var strct MyStruct = MyStruct{}
	writeln(strconv.Itoa(strct.field1))
	writeln(strconv.Itoa(strct.field2))

	strct.field1 = 123
	strct.field2 = 456

	writeln(strconv.Itoa(strct.field1))
	writeln(strconv.Itoa(strct.field2))
}

func testStructZeroValue() {
	var strct MyStruct
	writeln(strconv.Itoa(strct.field1))
	writeln(strconv.Itoa(strct.field2))
}

func testAtoi() {
	writeln(strconv.Itoa(strconv.Atoi("")))  // "0"
	writeln(strconv.Itoa(strconv.Atoi("0"))) // "0"
	writeln(strconv.Itoa(strconv.Atoi("1")))
	writeln(strconv.Itoa(strconv.Atoi("12")))
	writeln(strconv.Itoa(strconv.Atoi("1234567890")))
	writeln(strconv.Itoa(strconv.Atoi("-1234567890")))
	writeln(strconv.Itoa(strconv.Atoi("-7")))
}

func isLetter_(ch uint8) bool {
	if ch == '_' {
		return true
	}
	return ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z')
}

func testIsLetter() {
	if isLetter_('A') {
		writeln("OK isLetter A")
	} else {
		writeln("ERROR isLetter")
	}

}

func funcVaarg1(f string, a ...interface{}) {
	write(fmt.Sprintf(f, a...))
}

func funcVaarg2(a int, b ...int) {
	if b == nil {
		write(strconv.Itoa(a))
		writeln(" nil vaargs ok")
	} else {
		writeln("ERROR")
	}
}

func testVaargs() {
	funcVaarg1("pass nil slice\n")
	funcVaarg1("%s %s %s\n", "a", "bc", "def")
	funcVaarg2(777)
}

const O_READONLY_ int = 0

func testOpenRead() {
	var fd int
	fd, _ = syscall.Open("t/text.txt", O_READONLY_, 0)
	writeln(fd) // should be 3
	var buf []uint8 = make([]uint8, 300, 300)
	var n int
	n, _ = syscall.Read(fd, buf)
	if n < 279 {
		panic("ERROR")
	}
	writeln(n) // should be 280
	var readbytes []uint8 = buf[0:n]
	writeln(string(readbytes))
}

func testInfer() {
	var s = "infer string literal"
	writeln(s)

	var i = 3 + 5
	var j = i
	writeln(j)
}

func testEscapedChar() {
	var chars []uint8 = []uint8{'\\', '\t', '\r', '\'', '\n'}
	writeln("start")
	write(string(chars))
	writeln("end")
}

func testSwitchString() {
	var testVar string = "foo"
	var caseVar string = "fo"

	switch testVar {
	case "dummy":
		writeln("ERROR")
	}

	switch testVar {
	case "x", caseVar + "o":
		writeln("swithc string 1 ok")
	case "", "y":
		writeln("ERROR")
	default:
		writeln("ERROR")
	}

	switch testVar {
	case "":
		writeln("ERROR")
	case "fo":
		writeln("ERROR")
	default:
		writeln("switch string default ok")
	case "fooo":
		writeln("ERROR")
	}
}

func testSwitchByte() {
	var testVar uint8 = 'c'
	var caseVar uint8 = 'a'
	switch testVar {
	case 'b':
		writeln("ERROR")
	case caseVar + 2:
		writeln("switch uint8 ok")
	default:
		writeln("ERROR")
	}

	switch testVar {
	case 0:
		writeln("ERROR")
	case 'b':
		writeln("ERROR")
	default:
		writeln("switch default ok")
	case 'd':
		writeln("ERROR")
	}
}

func testSwitchInt() {
	var testVar int = 7
	var caseVar int = 5
	switch testVar {
	case 1:
		writeln("ERROR")
	case caseVar + 2:
		writeln("switch int ok")
	default:
		writeln("ERROR")
	}

	switch testVar {
	case 0:
		writeln("ERROR")
	case 6:
		writeln("ERROR")
	default:
		writeln("switch default ok")
	case 8:
		writeln("ERROR")
	}
}

func testLogicalAndOr() {
	var t bool = true
	var f bool = false

	if t && t {
		writeln("true && true ok")
	} else {
		writeln("ERROR")
	}

	if t && f {
		writeln("ERROR")
	} else {
		writeln("true && false ok")
	}
	if f && t {
		writeln("ERROR")
	} else {
		writeln("false && true ok")
	}
	if f && f {
		writeln("ERROR")
	} else {
		writeln("false && false ok")
	}

	if t || t {
		writeln("true || true ok")
	} else {
		writeln("ERROR")
	}
	if t || f {
		writeln("true || false ok")
	} else {
		writeln("ERROR")
	}
	if f || t {
		writeln("false || true ok")
	} else {
		writeln("ERROR")
	}
	if f || f {
		writeln("ERROR")
	} else {
		writeln("false || false ok")
	}
}

const MY_CONST_INT_VALUE int = 24

func testConst() {
	writeln(strconv.Itoa(MY_CONST_INT_VALUE))
}

func testForOmissible() {
	var i int
	for {
		i++
		if i == 2 {
			break
		}
	}
	write(strconv.Itoa(i))

	i = 0
	for i < 3 {
		i++
	}
	write(strconv.Itoa(i))

	i = 0
	for i < 4 {
		i++
	}
	write(strconv.Itoa(i))

	write("\n")
}

func testForBreakContinue() {
	var i int
	for i = 0; i < 10; i = i + 1 {
		if i == 3 {
			break
		}
		write(strconv.Itoa(i))
	}
	write("exit")
	writeln(strconv.Itoa(i))

	for i = 0; i < 10; i = i + 1 {
		if i < 7 {
			continue
		}
		write(strconv.Itoa(i))
	}
	write("exit")
	writeln(strconv.Itoa(i))

	var ary []int = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, i = range ary {
		if i == 3 {
			break
		}
		write(strconv.Itoa(i))
	}
	write("exit")
	writeln(strconv.Itoa(i))
	for _, i = range ary {
		if i < 7 {
			continue
		}
		write(strconv.Itoa(i))
	}
	write("exit")
	writeln(strconv.Itoa(i))
}

func returnTrue1() bool {
	var bol bool
	bol = true
	return bol
}

func returnTrue2() bool {
	var bol bool
	return !bol
}

func returnFalse() bool {
	var bol bool = true
	return !bol
}

var globalbool1 bool = true
var globalbool2 bool = false
var globalbool3 bool

func testGlobalBool() {
	if globalbool1 {
		writeln("globalbool 1 ok")
	} else {
		writeln("ERROR")
	}

	if globalbool2 {
		writeln("ERROR")
	} else {
		writeln("globalbool 2 ok")
	}

	if globalbool3 {
		writeln("ERROR")
	} else {
		writeln("globalbool 3 ok")
	}
}

func testLocalBool() {
	var bol bool = returnTrue1()
	if bol {
		writeln("bool 1 ok")
	} else {
		writeln("ERROR")
	}

	if !bol {
		writeln("ERROR")
	} else {
		writeln("bool ! 1 ok")
	}

	if returnTrue2() {
		writeln("bool 2 ok")
	} else {
		writeln("ERROR")
	}

	if returnFalse() {
		writeln("ERROR")
	} else {
		writeln("bool 3 ok")
	}
}

func testNilComparison() {
	var p *MyStruct
	if p == nil {
		writeln("nil pointer 1 ok")
	} else {
		writeln("ERROR")
	}
	p = nil
	if p == nil {
		writeln("nil pointer 2 ok")
	} else {
		writeln("ERROR")
	}

	var slc []string
	if slc == nil {
		writeln("nil pointer 3 ok")
	} else {
		writeln("ERROR")
	}
	slc = nil
	if slc == nil {
		writeln("nil pointer 4 ok")
	} else {
		writeln("ERROR")
	}
}

func testSliceLiteral() {
	var slc []string = []string{"this is ", "slice literal"}
	writeln(slc[0] + slc[1])
}

func testArrayCopy() {
	var aInt [3]int = [3]int{1, 2, 3}
	var bInt [3]int = aInt
	aInt[1] = 20

	write(strconv.Itoa(aInt[1]))
	write(strconv.Itoa(bInt[1]))
	write("\n")
}

func testLocalArrayWithMoreTypes() {
	var aInt [3]int = [3]int{1, 2, 3}
	var i int
	for _, i = range aInt {
		writeln(strconv.Itoa(i))
	}

	var aString [3]string = [3]string{"a", "bb", "ccc"}
	var s string
	for _, s = range aString {
		write(s)
	}
	write("\n")

	var aByte [4]uint8 = [4]uint8{'x', 'y', 'z', 10}
	var buf []uint8 = aByte[0:4]
	write(string(buf))
}

func testLocalArray() {
	var aInt [3]int = [3]int{1, 2, 3}
	write(strconv.Itoa(aInt[0]))
	write(strconv.Itoa(aInt[1]))
	write(strconv.Itoa(aInt[2]))
	write("\n")
}

func testAppendSlice() {
	var slcslc [][]string
	var slc []string
	slc = append(slc, "aa")
	slc = append(slc, "bb")
	slcslc = append(slcslc, slc)
	slcslc = append(slcslc, slc)
	slcslc = append(slcslc, slc)
	var s string
	for _, slc = range slcslc {
		for _, s = range slc {
			write(s)
		}
		write("|")
	}
	write("\n")
}

func testAppendPtr() {
	var slc []*MyStruct
	var p *MyStruct
	var i int
	for i = 0; i < 10; i++ {
		p = new(MyStruct)
		p.field1 = i
		slc = append(slc, p)
	}

	for _, p = range slc {
		write(strconv.Itoa(p.field1)) // 123456789
	}
	write("\n")
}

func testAppendString() {
	var slc []string
	slc = append(slc, "a")
	slc = append(slc, "bcde")
	var elm string = "fghijklmn\n"
	slc = append(slc, elm)
	var s string
	for _, s = range slc {
		write(s)
	}
	writeln(strconv.Itoa(len(slc))) // 3
}

func testAppendInt() {
	var slc []int
	slc = append(slc, 1)
	var i int
	for i = 2; i < 10; i++ {
		slc = append(slc, i)
	}
	slc = append(slc, 10)

	for _, i = range slc {
		write(strconv.Itoa(i)) // 12345678910
	}
	write("\n")
}

func testAppendByte() {
	var slc []uint8
	var char uint8
	for char = 'a'; char <= 'z'; char++ {
		slc = append(slc, char)
	}
	slc = append(slc, 10) // '\n'
	write(string(slc))
	writeln(strconv.Itoa(len(slc))) // 27
}

func testSringIndex() {
	var s string = "abcde"
	var char uint8 = s[3]
	writeln(strconv.Itoa(int(char)))
}

func testSubstring() {
	var s string = "abcdefghi"
	var subs string = s[2:5] // "cde"
	writeln(subs)
}

func testSliceOfSlice() {
	var slc []uint8 = make([]uint8, 3, 3)
	slc[0] = 'a'
	slc[1] = 'b'
	slc[2] = 'c'
	writeln(string(slc))

	var slc1 []uint8 = slc[0:3]
	writeln(string(slc1))

	var slc2 []uint8 = slc[0:2]
	writeln(string(slc2))

	var slc3 []uint8 = slc[1:3]
	writeln(string(slc3))
}

func testForrangeKey() {
	var i int
	var slc []string
	var s string
	slc = make([]string, 3, 3)
	slc[0] = "a"
	slc[1] = "b"
	slc[2] = "c"
	for i, s = range slc {
		write(strconv.Itoa(i))
		writeln(s)
	}
}

func testForrange() {
	var slc []string
	var s string

	writeln("going to loop 0 times")
	for _, s = range slc {
		write(s)
		write("ERROR")
	}

	slc = make([]string, 2, 2)
	slc[0] = ""
	slc[1] = ""

	writeln("going to loop 2 times")
	for _, s = range slc {
		write(s)
		writeln(" in loop")
	}

	writeln("going to loop 4 times")
	var a int
	for _, a = range globalintarray {
		write(strconv.Itoa(a))
	}
	writeln("")

	slc = make([]string, 3, 3)
	slc[0] = "hello"
	slc[1] = "for"
	slc[2] = "range"
	for _, s = range slc {
		write(s)
	}
	writeln("")
}

func newStruct() *MyStruct {
	var strct *MyStruct = new(MyStruct)
	writeln(strconv.Itoa(strct.field2))
	strct.field2 = 2
	return strct
}

func testNewStruct() {
	var strct *MyStruct
	strct = newStruct()
	writeln(strconv.Itoa(strct.field1))
	writeln(strconv.Itoa(strct.field2))
}

var nilSlice []*MyStruct

func testNilSlice() {
	nilSlice = make([]*MyStruct, 2, 2)
	writeln(strconv.Itoa(len(nilSlice)))
	writeln(strconv.Itoa(cap(nilSlice)))

	nilSlice = nil
	writeln(strconv.Itoa(len(nilSlice)))
	writeln(strconv.Itoa(cap(nilSlice)))
}

func testZeroValues() {
	writeln("-- testZeroValues()")
	var s string
	write(s)

	var s2 string = ""
	write(s2)
	var h int = 1
	var i int
	var j int = 2
	writeln(strconv.Itoa(h))
	writeln(strconv.Itoa(i))
	writeln(strconv.Itoa(j))

	if i == 0 {
		writeln("int zero ok")
	} else {
		writeln("ERROR")
	}
}

func testIncrDecr() {
	var i int = 0
	i++
	writeln(strconv.Itoa(i))

	i--
	i--
	writeln(strconv.Itoa(i))
}

type MyStruct struct {
	field1 int
	field2 int
	ifc    interface{}
}

var globalstrings1 [2]string
var globalstrings2 [2]string
var __slice []string

func testGlobalStrings() {
	globalstrings1[0] = "aaa,"
	globalstrings1[1] = "bbb,"
	globalstrings2[0] = "ccc,"
	globalstrings2[1] = "ddd,"
	__slice = make([]string, 1, 1)
	write(globalstrings1[0])
	write(globalstrings1[1])
	write(globalstrings1[0])
	write(globalstrings1[1])
}

var globalstrings [2]string

func testSliceOfStrings() {
	var s1 string = "hello"
	var s2 string = " strings\n"
	var strings []string = make([]string, 2, 2)
	var i int
	strings[0] = s1
	strings[1] = s2
	for i = 0; i < 2; i = i + 1 {
		write(strings[i])
	}

	globalstrings[0] = s1
	globalstrings[1] = " globalstrings\n"
	for i = 0; i < 2; i = i + 1 {
		write(globalstrings[i])
	}
}

var structPointers []*MyStruct

func testSliceOfPointers() {
	var strct1 MyStruct
	var strct2 MyStruct
	var p1 *MyStruct = &strct1
	var p2 *MyStruct = &strct2

	strct1.field2 = 11
	strct2.field2 = 22
	structPointers = make([]*MyStruct, 2, 2)
	structPointers[0] = p1
	structPointers[1] = p2

	var i int
	var x int
	for i = 0; i < 2; i = i + 1 {
		x = structPointers[i].field2
		writeln(strconv.Itoa(x))
	}
}

func testStructPointer() {
	var _strct MyStruct
	var strct *MyStruct
	strct = &_strct
	strct.field1 = 123

	var i int
	i = strct.field1
	writeln(strconv.Itoa(i))

	strct.field2 = 456
	writeln(strconv.Itoa(_strct.field2))

	strct.field1 = 777
	strct.field2 = strct.field1
	writeln(strconv.Itoa(strct.field2))
}

func testStruct() {
	var strct MyStruct
	strct.field1 = 123

	var i int
	i = strct.field1
	writeln(strconv.Itoa(i))

	strct.field2 = 456
	writeln(strconv.Itoa(strct.field2))

	strct.field1 = 777
	strct.field2 = strct.field1
	writeln(strconv.Itoa(strct.field2))
}

func testPointer() {
	var i int = 12
	var j int
	var p *int
	p = &i
	j = *p
	writeln(strconv.Itoa(j))
	*p = 11
	writeln(strconv.Itoa(i))

	var c uint8 = 'A'
	var pc *uint8
	pc = &c
	*pc = 'B'
	var slc []uint8
	slc = make([]uint8, 1, 1)
	slc[0] = c
	writeln(string(slc))
}

func testDeclValue() {
	var i int = 123
	writeln(strconv.Itoa(i))
}

func testStringComparison() {
	var s string
	if s == "" {
		writeln("string cmp 1 ok")
	} else {
		writeln("ERROR")
	}
	var s2 string = ""
	if s2 == s {
		writeln("string cmp 2 ok")
	} else {
		writeln("ERROR")
	}

	var s3 string = "abc"
	s3 = s3 + "def"
	var s4 string = "1abcdef1"
	var s5 string = s4[1:7]
	if s3 == s5 {
		writeln("string cmp 3 ok")
	} else {
		writeln("ERROR")
	}

	if "abcdef" == s5 {
		writeln("string cmp 4 ok")
	} else {
		writeln("ERROR")
	}

	if s3 != s5 {
		writeln(s3)
		writeln(s5)
		writeln("ERROR")
		return
	} else {
		writeln("string cmp not 1 ok")
	}

	if s4 != s3 {
		writeln("string cmp not 2 ok")
	} else {
		writeln("ERROR")
	}
}

func testConcateStrings() {
	var concatenated string = "foo" + "bar" + "1234"
	writeln(concatenated)
}

func testLenCap() {
	var x []uint8
	x = make([]uint8, 0, 0)
	writeln(strconv.Itoa(len(x)))

	writeln(strconv.Itoa(cap(x)))

	x = make([]uint8, 12, 24)
	writeln(strconv.Itoa(len(x)))

	writeln(strconv.Itoa(cap(x)))

	writeln(strconv.Itoa(len(globalintarray)))

	writeln(strconv.Itoa(cap(globalintarray)))

	var s string
	s = "hello\n"
	writeln(strconv.Itoa(len(s))) // 6
}

func testMakeSlice() {
	var x []uint8 = make([]uint8, 3, 20)
	x[0] = 'A'
	x[1] = 'B'
	x[2] = 'C'
	writeln(string(x))
}

func testNew() {
	var i *int
	i = new(int)
	writeln(strconv.Itoa(*i)) // 0
}

func testItoa() {
	writeln(strconv.Itoa(0))
	writeln(strconv.Itoa(1))
	writeln(strconv.Itoa(12))
	writeln(strconv.Itoa(123))
	writeln(strconv.Itoa(12345))
	writeln(strconv.Itoa(12345678))
	writeln(strconv.Itoa(1234567890))
	writeln(strconv.Itoa(54321))
	writeln(strconv.Itoa(-1))
	writeln(strconv.Itoa(-54321))
	writeln(strconv.Itoa(-7654321))
	writeln(strconv.Itoa(-1234567890))
}

var globalintarray [4]int

func testIndexExprOfArray() {
	globalintarray[0] = 11
	globalintarray[1] = 22
	globalintarray[2] = globalintarray[1]
	globalintarray[3] = 44
	write("\n")
}

func testIndexExprOfSlice() {
	var intslice []int = globalintarray[0:4]
	intslice[0] = 66
	intslice[1] = 77
	intslice[2] = intslice[1]
	intslice[3] = 88

	var i int
	for i = 0; i < 4; i = i + 1 {
		write(strconv.Itoa(intslice[i]))
	}
	write("\n")

	for i = 0; i < 4; i = i + 1 {
		write(strconv.Itoa(globalintarray[i]))
	}
	write("\n")
}

func testFor() {
	var i int
	for i = 0; i < 3; i = i + 1 {
		write("A")
	}
	write("\n")
}

func testCmpUint8() {
	var localuint8 uint8 = 1
	if localuint8 == 1 {
		writeln("uint8 cmp == ok")
	}
	if localuint8 != 1 {
		writeln("ERROR")
	} else {
		writeln("uint8 cmp != ok")
	}
	if localuint8 > 0 {
		writeln("uint8 cmp > ok")
	}
	if localuint8 < 0 {
		writeln("ERROR")
	} else {
		writeln("uint8 cmp < ok")
	}

	if localuint8 >= 1 {
		writeln("uint8 cmp >= ok")
	}
	if localuint8 <= 1 {
		writeln("uint8 cmp <= ok")
	}

	localuint8 = 101
	if localuint8 == 'A' {
		writeln("uint8 cmp == A ok")
	}
}

func testCmpInt() {
	var a int = 1
	if a == 1 {
		writeln("int cmp == ok")
	}
	if a != 1 {
		writeln("ERROR")
	} else {
		writeln("int cmp != ok")
	}
	if a > 0 {
		writeln("int cmp > ok")
	}
	if a < 0 {
		writeln("ERROR")
	} else {
		writeln("int cmp < ok")
	}

	if a >= 1 {
		writeln("int cmp >= ok")
	}
	if a <= 1 {
		writeln("int cmp <= ok")
	}
	a = 101
	if a == 'A' {
		writeln("int cmp == A ok")
	}
}

func testElseIf() {
	if false {
		writeln("ERROR")
	} else if true {
		writeln("ok else if")
	} else {
		writeln("ERROR")
	}

	if false {
		writeln("ERROR")
	} else if false {
		writeln("ERROR")
	} else {
		writeln("ok else if else")
	}
}

func testIf() {
	var tr bool = true
	var fls bool = false
	if tr {
		writeln("ok true")
	}
	if fls {
		writeln("ERROR")
	}
	writeln("ok false")
}

func testElse() {
	if true {
		writeln("ok true")
	} else {
		writeln("ERROR")
	}

	if false {
		writeln("ERROR")
	} else {
		writeln("ok false")
	}
}

var globalint int
var globalint2 int
var globaluint8 uint8
var globaluint16 uint16

var globalstring string
var globalarray [9]uint8
var globalslice []uint8
var globaluintptr uintptr

func assignGlobal() {
	globalint = 22
	globaluint8 = 1
	globaluint16 = 5
	globaluintptr = 7
	globalstring = "globalstring changed\n"
}

func add1(x int) int {
	return x + 1
}

func sum(x int, y int) int {
	return x + y
}

func print1(a string) {
	write(a)
	return
}

func print2(a string, b string) {
	write(a)
	write(b)
}

func returnstring() string {
	return "i am a local 1\n"
}

func testGlobalCharArray() {
	globalarray[0] = 'A'
	globalarray[1] = 'B'
	globalarray[2] = globalarray[0]
	globalarray[3] = 100 / 10 // '\n'
	globalarray[1] = 'B'
	var chars []uint8 = globalarray[0:4]
	write(string(chars))
	globalslice = chars
	write(string(globalarray[0:4]))
}

func testString() {
	write(globalstring)
	assignGlobal()

	print1("hello string literal\n")

	var s string = "hello string"
	writeln(s)

	var localstring1 string = returnstring()
	var localstring2 string
	localstring2 = "i m local2\n"
	print1(localstring1)
	print2(localstring1, localstring2)
	write(globalstring)
}

func testArgAssign(x int) int {
	x = 13
	return x
}

func testMinus() int {
	var x int = -1
	x = x * -5
	return x
}

func testMisc() {
	var i13 int = 0
	i13 = testArgAssign(i13)
	var i5 int = testMinus()
	globalint2 = sum(1, i13%i5)

	var locali3 int
	var tmp int
	tmp = int(uint8('3' - '1'))
	tmp = tmp + int(globaluint16)
	tmp = tmp + int(globaluint8)
	tmp = tmp + int(globaluintptr)
	locali3 = add1(tmp)
	var i42 int
	i42 = sum(globalint, globalint2) + locali3

	writeln(strconv.Itoa(i42))
}

func main() {
	testReturnMultiValues()
	testPassBytes()
	testSprinfMore()
	testAnotherFile()
	testSortStrings()
	testGetdents64()
	testEnv()
	testReflect()
	testReturnSlice()
	testStrings()
	testSliceExpr()
	testPath()
	testByteType()
	testExtLib()
	testExpandSlice()
	testFullSlice()
	testInterfaceVaargs()
	testConvertToInterface()
	testTypeSwitch()
	testGetInterface()
	testPassInterface()
	testInterfaceAssertion()
	testInterfaceimplicitConversion()
	testInterfaceZeroValue()
	testForRangeShortDecl()
	testInferVarTypes()
	testGlobalValues()
	testShortVarDecl()
	testStructPointerMethods()
	//testBasicMethodCalls()
	testPointerMethod()
	testMethodAnother()
	testMethodSimple()
	testOsArgs()
	testStructLiteralWithContents()
	testAddressOfStructLiteral()
	testStructCopy()
	testStructLiteral()
	testStructZeroValue()
	testAtoi()
	testIsLetter()
	testVaargs()
	testOpenRead()
	testInfer()
	testEscapedChar()
	testSwitchString()
	testSwitchByte()
	testSwitchInt()

	testLogicalAndOr()
	testConst()
	testForOmissible()
	testForBreakContinue()
	testGlobalBool()
	testLocalBool()
	testNilComparison()
	testSliceLiteral()
	testArrayCopy()
	testLocalArrayWithMoreTypes()
	testLocalArray()
	testAppendSlice()
	testAppendPtr()
	testAppendString()
	testAppendInt()
	testAppendByte()
	testSringIndex()
	testSubstring()
	testSliceOfSlice()
	testForrangeKey()
	testForrange()
	testNewStruct()
	testNilSlice()
	testZeroValues()
	testIncrDecr()
	testGlobalStrings()
	testSliceOfStrings()
	testSliceOfPointers()
	testStructPointer()
	testStruct()
	testPointer()
	testDeclValue()
	testStringComparison()
	testConcateStrings()
	testLenCap()
	testMakeSlice()
	testNew()

	testItoa()
	testIndexExprOfArray()
	testIndexExprOfSlice()
	testString()
	testFor()
	testCmpUint8()
	testCmpInt()
	testElseIf()
	testElse()
	testIf()
	testGlobalCharArray()

	testMisc()
	os.Exit(0)
}
