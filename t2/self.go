package main

import "syscall"

func fmtSprintf(format string, a []string) string {
	if len(a) == 0 {
		return format
	}
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

func fmtPrint(s string) {
	var slc []uint8 = []uint8(s)
	syscall.Write(1, slc)
}

func fmtPrintf(format string, a []string) {
	var s string = fmtSprintf(format, a)
	syscall.Write(1, []uint8(s))
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

	return name
}

func emitGlobalVariable(name string, t *Type, val string) {
	var typeKind string
	if t != nil {
		typeKind = t.kind
	}
	fmtPrintf("%s: # T %s \n", []string{name, typeKind})
	switch typeKind {
	case "T_STRING":
		fmtPrint("  .quad 0\n")
		fmtPrint("  .quad 0\n")
	case "T_INT":
		fmtPrintf("  .quad %s\n", []string{val})
	default:
		fmtPrint("ERROR\n")
	}
}

func emitData(pkgName string) {
	fmtPrint(".data\n")
	var sl *sliteral
	for _, sl = range stringLiterals {
		fmtPrint("# string literals\n")
		fmtPrintf("%s:\n", []string{sl.label})
		fmtPrintf("  .string %s\n", []string{sl.value})
	}
	fmtPrint("# ===== Global Variables =====\n")
	var spec *astValueSpec
	for _, spec = range globalVars {
		emitGlobalVariable(spec.name, spec.t, spec.value)
	}

	fmtPrint("# ==============================\n")
}

func emitFuncDecl(pkgPrefix string, fnc *Func) {
	fmtPrint("\n")
	var fname string = fnc.name
	fmtPrint(pkgPrefix + "." + fname + ":\n")
	if len(fnc.localvars) > 0 {
		var slocalarea string = Itoa(fnc.localarea)
		fmtPrint("  subq $" + slocalarea + ", %rsp # local area\n")
	}

	fmtPrint("  leave\n")
	fmtPrint("  ret\n")
}

func emitText(pkgName string) {
	fmtPrint(".text\n")
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

func main() {
	var sourceFiles []string = []string{"main"}
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
