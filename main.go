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
	fmt.Printf("    version:  show version\n")
	fmt.Printf("    compile -o <tmpbasename> <pkgpath> e.g. compile -o /tmp/os os\n")
	fmt.Printf("    list -deps <pkgpath> e.g. list -deps ./t/\n")
}

func showVersion() {
	fmt.Printf("babygo version %s  linux/amd64\n", Version)
}

func main() {
	if len(os.Args) == 1 {
		showHelp()
		return
	}

	switch os.Args[1] {
	case "version":
		showVersion()
		return
	case "help":
		showHelp()
		return
	case "panic": // What's this for ?? I can't remember ...
		panicVersion := strconv.Itoa(mylib.Sum(1, 1))
		panic("I am panic version " + panicVersion)
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		panic("GOPATH is not set")
	}
	workdir := os.Getenv("WORKDIR")
	if workdir == "" {
		panic("WORKDIR is not set")
	}
	srcPath := gopath + "/src"                                 // userland packages
	bbgRootSrcPath := srcPath + "/github.com/DQNEO/babygo/src" // std packages

	b := builder.Builder{
		SrcPath:        srcPath,
		BbgRootSrcPath: bbgRootSrcPath,
	}

	switch os.Args[1] {
	case "compile": // e.g. WORKDIR=/tmpfs/bbg/bbg-test.d babygo compile -o /tmp/os os
		outputBaseName := os.Args[3]
		pkgPath := os.Args[4]
		b.BuildOne(workdir, outputBaseName, pkgPath)
	case "list": // e.g. WORKDIR=/tmpfs/bbg/bbg-test.d babygo list -depth ./t/
		pkgPath := os.Args[3]
		b.ListDepth(workdir, pkgPath, os.Stdout)
	case "link": // e.g. babygo link -o $out_file_abs a.o b.o
		outFilePath := os.Args[3]
		objFileNames := os.Args[4:]
		b.Link(outFilePath, objFileNames)
	case "build": // e.g. babygo build -o outfile ./t
		outFilePath := os.Args[3]
		pkgPath := os.Args[4]
		b.Build(os.Args[0], workdir, outFilePath, pkgPath)
	default:
		showHelp()
		return
	}
}
