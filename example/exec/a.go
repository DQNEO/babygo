package main

import (
	"fmt"
	"os/exec"
	"syscall"
)

func main() {
	const CLONE_VM uintptr = 256      // 0x100
	const CLONE_VFORK uintptr = 16384 // 0x00004000 //
	const SIGCHLD uintptr = 3

	var flags uintptr = CLONE_VM | CLONE_VFORK | SIGCHLD
	r1, r2, err := syscall.Syscall(uintptr(syscall.SYS_CLONE), 0, flags, 0)
	fmt.Println(r1, r2, err)
}

func execCommand() {
	err := exec.Command("/usr/bin/mkdir", "/tmp/0043").Run()
	if err != nil {
		panic(err)
	}
}
