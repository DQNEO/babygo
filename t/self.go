package main

import "syscall"

func write(s string) {
	var slc []uint8 = []uint8(s)
	syscall.Write(1, slc)
}

func semanticAnalyze() {
	globalFuncsArray[0] = "main"
}

func emitData() {

}

func emitFuncDecl(pkgPrefix string, fn string) {
	write(".global main.main\n")
	write(pkgPrefix)
	write(".")
	write(fn)
	write(":\n")
	write("  ret\n")
}

var globalFuncsArray [1]string

func emitText() {
	write(".text\n")
	var i int
	for i = 0; i<1; i = i + 1 {
		emitFuncDecl("main", globalFuncsArray[i])
	}
}

func generateCode() {
	emitData()
	emitText()
}

func main() {
	semanticAnalyze()
	generateCode()
}

