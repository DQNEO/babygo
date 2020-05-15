package main

import "syscall"

func fmtSprintf(format string, a []string) string {
	var buf []uint8
	var inPercent bool
	var argIndex int
	var c uint8
	for _, c = range []uint8(format) {
		if inPercent {
			if c == '%' {
				buf = append(buf, c)
			} else {
				var arg string = a[argIndex]
				argIndex++
				var s string = arg // // p.printArg(arg, c)
				var _c uint8
				for _, _c = range []uint8(s) {
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

func fmtPrintf(format string, a ...string) {
	var s string = fmtSprintf(format, a)
	syscall.Write(1, []uint8(s))
}

var __itoa_buf [100]uint8
var __itoa_r [100]uint8

func Itoa(ival int) string {
	var next int
	var right int
	var ix int = 0
	if ival == 0 {
		return "0"
	}
	var minus bool
	minus = false
	for ix = 0; ival != 0; ix = ix + 1 {
		if ival < 0 {
			ival = -1 * ival
			minus = true
			__itoa_r[0] = '-'
		} else {
			next = ival / 10
			right = ival - next*10
			ival = next
			__itoa_buf[ix] = uint8('0' + right)
		}
	}

	var j int
	var c uint8
	for j = 0; j < ix; j = j + 1 {
		c = __itoa_buf[ix-j-1]
		if minus {
			__itoa_r[j+1] = c
		} else {
			__itoa_r[j] = c
		}
	}

	return string(__itoa_r[0:ix])
}

func emitPopBool(comment string) {
	fmtPrintf("  popq %%rax # result of %s\n", comment)
}

func emitPopAddress(comment string) {
	fmtPrintf("  popq %%rax # address of %s\n", comment)
}

func emitPopString() {
	fmtPrintf("  popq %%rax # string.ptr\n")
	fmtPrintf("  popq %%rcx # string.len\n")
}

func emitPopSlice() {
	fmtPrintf("  popq %%rax # slice.ptr\n")
	fmtPrintf("  popq %%rcx # slice.len\n")
	fmtPrintf("  popq %%rdx # slice.cap\n")
}

func emitPushStackTop(condType *Type, comment string) {
	switch kind(condType) {
	case T_STRING:
		fmtPrintf("  movq 8(%%rsp), %%rcx # copy str.len from stack top (%s)\n", comment)
		fmtPrintf("  movq 0(%%rsp), %%rax # copy str.ptr from stack top (%s)\n", comment)
		fmtPrintf("  pushq %%rcx # str.len\n")
		fmtPrintf("  pushq %%rax # str.ptr\n")
	case T_POINTER, T_UINTPTR, T_BOOL, T_INT, T_UINT8, T_UINT16:
		fmtPrintf("  movq (%%rsp), %%rax # copy stack top value (%s) \n", comment)
		fmtPrintf("  pushq %%rax\n")
	default:
		throw(kind(condType))
	}
}

func throw(s string) {
	syscall.Write(2, []uint8(s))
}
func kind(t *Type) string {
	return t.kind
}

func semanticAnalyze(file *astFile) string {
	return fakeSemanticAnalyze(file)
}

const T_STRING string = "T_STRING"
const T_SLICE string = "T_SLICE"
const T_BOOL string = "T_BOOL"
const T_INT string = "T_INT"
const T_UINT8 string = "T_UINT8"
const T_UINT16 string = "T_UINT16"
const T_UINTPTR string = "T_UINTPTR"
const T_ARRAY string = "T_ARRAY"
const T_STRUCT string = "T_STRUCT"
const T_POINTER string = "T_POINTER"

func emitGlobalVariable(name string, t *Type, val string) {
	var typeKind string
	if t != nil {
		typeKind = t.kind
	}
	fmtPrintf("%s: # T %s \n", name, typeKind)
	switch typeKind {
	case T_STRING:
		fmtPrintf("  .quad 0\n")
		fmtPrintf("  .quad 0\n")
	case T_INT:
		fmtPrintf("  .quad %s\n", val)
	default:
		fmtPrintf("ERROR\n")
	}
}

func emitData(pkgName string) {
	fmtPrintf(".data\n")
	var sl *sliteral
	for _, sl = range stringLiterals {
		fmtPrintf("# string literals\n")
		fmtPrintf("%s:\n", sl.label)
		fmtPrintf("  .string %s\n", sl.value)
	}

	fmtPrintf("# ===== Global Variables =====\n")
	var spec *astValueSpec
	for _, spec = range globalVars {
		emitGlobalVariable(spec.name, spec.t, spec.value)
	}

	fmtPrintf("# ==============================\n")
}

func emitFuncDecl(pkgPrefix string, fnc *Func) {
	fmtPrintf("\n")
	var fname string = fnc.name
	fmtPrintf(pkgPrefix + "." + fname + ":\n")
	if len(fnc.localvars) > 0 {
		var slocalarea string = Itoa(fnc.localarea)
		fmtPrintf("  subq $" + slocalarea + ", %rsp # local area\n")
	}

	fmtPrintf("  leave\n")
	fmtPrintf("  ret\n")
}

func emitText(pkgName string) {
	fmtPrintf(".text\n")
	var i int
	for i = 0; i < len(globalFuncs); i++ {
		var fnc *Func = globalFuncs[i]
		emitFuncDecl(pkgName, fnc)
	}
}

func generateCode(pkgName string) {
	emitData(pkgName)
	emitText(pkgName)
}

type Type struct {
	kind string
}

type astValueSpec struct {
	name  string
	t     *Type
	value string
}

type Func struct {
	localvars []*string
	localarea int
	argsarea  int
	name      string
}

type sliteral struct {
	label  string
	strlen int
	value  string // raw value
}

var stringLiterals []*sliteral
var stringIndex int
var globalVars []*astValueSpec
var globalFuncs []*Func

var _garbage string

type astFile struct {
	Name string
}

func readSource(filename string) []uint8 {
	return []uint8("main")
}

var gtext []uint8

func parserInit(text []uint8) {
	gtext = text
}

func parserParseFile() *astFile {
	var f *astFile = new(astFile)
	f.Name = string(gtext)
	return f
}

func parseFile(filename string) *astFile {
	var text []uint8 = readSource(filename)
	parserInit(text)
	var f *astFile = parserParseFile()
	return f
}

func main() {
	var sourceFiles []string = []string{"main"}
	var sourceFile string
	for _, sourceFile = range sourceFiles {
		globalVars = nil
		globalFuncs = nil
		stringLiterals = nil
		stringIndex = 0
		var f *astFile = parseFile(sourceFile)
		var pkgName string = semanticAnalyze(f)
		generateCode(pkgName)
	}

	test()
}

func test() {
	// Test funcs
	emitPopBool("comment")
	emitPopAddress("comment")
	emitPopString()
	emitPopSlice()
	var t1 *Type = new(Type)
	t1.kind = T_INT
	emitPushStackTop(t1, "comment")
}

func fakeSemanticAnalyze(file *astFile) string {
	globalFuncs = make([]*Func, 2, 2)
	var fnc *Func = new(Func)
	fnc.name = file.Name
	globalFuncs[0] = fnc
	var fnc2 *Func = new(Func)
	fnc2.name = "foo"
	globalFuncs[1] = fnc2

	stringLiterals = make([]*sliteral, 2, 2)
	var s1 *sliteral = new(sliteral)
	s1.value = "hello0"
	s1.label = ".main.S0"
	stringLiterals[0] = s1

	var s2 *sliteral = new(sliteral)
	s2.value = "hello1"
	s2.label = ".main.S1"
	stringLiterals[1] = s2

	var globalVar0 *astValueSpec = new(astValueSpec)
	globalVar0.name = "_gvar0"
	globalVar0.value = "10"
	globalVar0.t = new(Type)
	globalVar0.t.kind = "T_INT"
	globalVars = append(globalVars, globalVar0)

	var globalVar1 *astValueSpec = new(astValueSpec)
	globalVar1.name = "_gvar1"
	globalVar1.value = "11"
	globalVar1.t = new(Type)
	globalVar1.t.kind = "T_INT"
	globalVars = append(globalVars, globalVar1)

	var globalVar2 *astValueSpec = new(astValueSpec)
	globalVar2.name = "_gvar2"
	globalVar2.value = "foo"
	globalVar2.t = new(Type)
	globalVar2.t.kind = "T_STRING"
	globalVars = append(globalVars, globalVar2)

	return file.Name
}
