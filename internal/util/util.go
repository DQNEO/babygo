package util

import (
	"os"

	"github.com/DQNEO/babygo/lib/fmt"
)

const ThrowFormat string = "%T"

func Throw(x interface{}) {
	panic(fmt.Sprintf(ThrowFormat, x))
}

// Logf writes General debug log to stderr
func Logf(format string, a ...interface{}) {
	f := "# " + format
	fmt.Fprintf(os.Stderr, f, a...)
}

var CurrentPkgName string

func Assert(bol bool, msg string, caller string) {
	if !bol {
		panic(CurrentPkgName + ":" + caller + ": " + msg)
	}
}
