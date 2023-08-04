package util

import (
	"os"

	"github.com/DQNEO/babygo/lib/fmt"
)

// Logf writes General debug log to stderr
func Logf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}
