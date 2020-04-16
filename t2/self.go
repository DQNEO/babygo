package main

import "syscall"

func write(s string) {
	var slc []uint8 = []uint8(s)
	syscall.Write(1, slc)
}

func semanticAnalyze() {
	globalFuncsArray[0] = "main"
}

func emitData(pkgName string) {

}

func emitFuncDecl(pkgPrefix string, fn string) {
	write(".global " + pkgPrefix + ".main\n")
	write(pkgPrefix + "." + fn + ":\n")
	write("  ret\n")
}

var globalFuncsArray [1]string

func emitText(pkgName string) {
	write(".text\n")
	var i int
	for i = 0; i<1; i = i + 1 {
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
	for i=0;i<len(sourceFiles); i=i+1 {
		_garbage = sourceFiles[i]
		semanticAnalyze()
		var pkgName string = "main"
		generateCode(pkgName)
	}
}

