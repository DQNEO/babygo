// go build -o ifc-call -gcflags='-N -l' example/ifc-method/main.go
package main

import (
	"io"
	"os"
)

var msg = []uint8("hello world!\n")
var w io.Writer = os.Stdout

func main() {
	w.Write(msg)
}
