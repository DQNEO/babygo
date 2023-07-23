package main

import (
	"os"

	"github.com/DQNEO/babygo/internal/sema"
	"github.com/DQNEO/babygo/internal/types"
	"github.com/DQNEO/babygo/internal/universe"
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/strconv"
	//gofmt "fmt"
)

const Version string = "0.0.7"

var ProgName string = "babygo"

var __func__ = "__func__"

func assert(bol bool, msg string, caller string) {
	if !bol {
		panic(sema.CurrentPkg.Name + ":" + caller + ": " + msg)
	}
}

const ThrowFormat string = "%T"

func throw(x interface{}) {
	panic(fmt.Sprintf(ThrowFormat, x))
}

var fout *os.File

func printf(format string, a ...interface{}) {
	fmt.Fprintf(fout, format, a...)
}

// General debug log
func logf(format string, a ...interface{}) {
	f := "# " + format
	fmt.Fprintf(os.Stderr, f, a...)
}

// --- main ---
func showHelp() {
	fmt.Printf("Usage:\n")
	fmt.Printf("    %s version:  show version\n", ProgName)
	fmt.Printf("    %s [-DG] filename\n", ProgName)
}

func main() {
	// Check object addresses
	tIdent := types.Int.E.(*ast.Ident)
	if tIdent.Obj != universe.Int {
		panic("object mismatch")
	}

	srcPath = os.Getenv("GOPATH") + "/src"
	prjSrcPath = srcPath + "/github.com/DQNEO/babygo/src"
	if len(os.Args) == 1 {
		showHelp()
		return
	}

	if os.Args[1] == "version" {
		fmt.Printf("babygo version %s  linux/amd64\n", Version)
		return
	} else if os.Args[1] == "help" {
		showHelp()
		return
	} else if os.Args[1] == "panic" {
		panicVersion := strconv.Itoa(mylib.Sum(1, 1))
		panic("I am panic version " + panicVersion)
	}

	buildAll(os.Args[1:])
}
