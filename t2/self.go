package main

import "syscall"

func write(s string) {
	var slc []uint8 = []uint8(s)
	syscall.Write(1, slc)
}

func semanticAnalyze() string {
	globalFuncsArray[0] = "main"

	stringLiterals = make([]string, 1, 1)
	stringLiterals[0] = "hello"

	return "main"
}

func emitData(pkgName string) {
	write(".data\n")

	var i int = 0
	for i=0;i<1;i++ {
		write("." + pkgName + ".S0" + ":\n")
		write("  .string " + stringLiterals[i] + "\n")
	}
}

func emitFuncDecl(pkgPrefix string, fn string) {
	write(".global " + pkgPrefix + ".main\n")
	write(pkgPrefix + "." + fn + ":\n")
	write("  ret\n")
}

var globalFuncsArray [1]string
var stringLiterals []string

func emitText(pkgName string) {
	write(".text\n")
	var i int
	for i = 0; i<1; i++ {
		emitFuncDecl(pkgName, globalFuncsArray[i])
	}
}

func generateCode(pkgName string) {
	emitData(pkgName)
	emitText(pkgName)
}

var sourceFiles [1]string
var _garbage string
func main() {
	sourceFiles[0] = "./t/self.go"
	var i int
	for i=0;i<len(sourceFiles); i++ {
		_garbage = sourceFiles[i]
		var pkgName string = semanticAnalyze()
		generateCode(pkgName)
	}
}

