package main

import "syscall"

var buf [100]uint8
var r [100]uint8

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

func fmtPrintf(s string) {
	var slc []uint8 = []uint8(s)
	syscall.Write(1, slc)
}

func semanticAnalyze(name string) string {
	globalFuncs = make([]*Func, 2, 2)
	var fnc *Func = new(Func)
	fnc.name = "main"
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

	return name
}

func emitData(pkgName string) {
	fmtPrintf(".data\n")
	var sl *sliteral
	for _, sl = range stringLiterals {
		fmtPrintf("# string literals\n")
		fmtPrintf(sl.label + ":\n")
		fmtPrintf("  .string " + sl.value + "\n")
	}
	fmtPrintf("# ===== Global Variables =====\n")
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

type astValueSpec struct {
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

var sourceFiles [1]string
var _garbage string

func main() {
	sourceFiles[0] = "main"
	var sourceFile string
	for _, sourceFile = range sourceFiles {
		globalVars = nil
		globalFuncs = nil
		stringLiterals = nil
		stringIndex = 0
		var pkgName string = semanticAnalyze(sourceFile)
		generateCode(pkgName)
	}
}
