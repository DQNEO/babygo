package main

import (
	"os"

	"github.com/DQNEO/babygo/internal/builder"
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/mylib"
	"github.com/DQNEO/babygo/lib/strconv"
)

const Version string = "0.1.0"

const ProgName string = "babygo"

func showHelp() {
	fmt.Printf("Usage:\n")
	fmt.Printf("    %s version:  show version\n", ProgName)
	fmt.Printf("    %s [-DG] filename\n", ProgName)
}

func showVersion() {
	fmt.Printf("babygo version %s  linux/amd64\n", Version)
}

func main() {
	if len(os.Args) == 1 {
		showHelp()
		return
	}
	var mode string
	var args []string
	switch os.Args[1] {
	case "version":
		showVersion()
		return
	case "help":
		showHelp()
		return
	case "compile":
		mode = "single"
		args = os.Args[2:]
	case "panic": // What's this for ?? I can't remember ...
		panicVersion := strconv.Itoa(mylib.Sum(1, 1))
		panic("I am panic version " + panicVersion)
	default:
		mode = "build"
		args = os.Args[1:]
	}

	workdir := os.Getenv("WORKDIR")
	if workdir == "" {
		workdir = "/tmp"
	}
	srcPath := os.Getenv("GOPATH") + "/src"                    // userland packages
	bbgRootSrcPath := srcPath + "/github.com/DQNEO/babygo/src" // std packages

	b := builder.Builder{
		SrcPath:        srcPath,
		BbgRootSrcPath: bbgRootSrcPath,
	}
	switch mode {
	case "single":
		b.Compile(workdir, args)
	case "build":
		b.Build(workdir, args)
	}
}
