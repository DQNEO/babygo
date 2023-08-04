package testing

import (
	"os"

	"github.com/DQNEO/babygo/lib/fmt"
)

type T struct {
	failed bool
}

func (t *T) Failed() bool {
	return t.failed
}

func (t *T) Error(args ...interface{}) {
	for _, arg := range args {
		fmt.Fprintf(os.Stderr, "%s ", arg)
	}
	fmt.Fprintf(os.Stderr, "\n")
	t.failed = true
}

func TestMain(funcs []func(*T)) {
	t := &T{}

	for _, fn := range funcs {
		fn(t)
	}

	if t.Failed() {
		os.Stderr.Write([]byte("FAILED\n"))
		os.Exit(1)
	}
}
