package os

import "syscall"

const SYS_EXIT int = 60

var Args []string

func runtime_args() []string

func init() {
	Args = runtime_args()
}

func Exit(status int) {
	syscall.Syscall(uintptr(SYS_EXIT), uintptr(status), 0 , 0)
}
