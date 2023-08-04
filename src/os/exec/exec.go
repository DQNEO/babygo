package exec

import (
	"os"
	"syscall"

	"runtime"
)

type Cmd struct {
	Name string
	Arg  string
}

func Command(name string, arg string) *Cmd {
	return &Cmd{
		Name: name,
		Arg:  arg,
	}
}

const CLONE_CHILD_CLEARTID uintptr = 2097152 //  0x00200000 // 2097152
const CLONE_CHILD_SETTID uintptr = 16777216  // 0x01000000 // 16777216
const SIGCHLD uintptr = 17

func (c *Cmd) Run() error {

	pid := fork()

	if pid == 0 {
		// child
		os.Stdout.Write([]byte("\nI am the child\n"))
		os.Stdout.Write([]byte("child: " + c.Name + " " + c.Arg + "\n"))
		os.Exit(0)
	} else {
		os.Stdout.Write([]byte("I am the parent\n"))
		//_ = r2
		//_ = err
		spid := runtime.Itoa(pid)
		os.Stdout.Write([]byte("parent: child pid=" + spid + "\n"))
	}
	return nil
}

func fork() int {
	//   runtime.clone:
	//     movq 24(%rsp), %r12
	//     movl $56, %eax # sys_clone
	//     movq 8(%rsp), %rdi # flags
	//     movq 16(%rsp), %rsi # stack
	//     movq $0, %rdx # ptid
	//     movq $0, %r10 # chtid
	//     movq $0, %r8 # tls
	//     syscall
	//
	//clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|CLONE_CHILD_SETTID|SIGCHLD, child_tidptr=0x7feb85d62a10) = 42420
	//
	// The raw system call interface on x86-64 and some other
	//       architectures (including sh, tile, and alpha) is:
	//
	// long clone(unsigned long flags, void *stack,
	//                      int *parent_tid, int *child_tid,
	//                      unsigned long tls);

	trap := uintptr(56) // sys_clone
	var flags uintptr = CLONE_CHILD_CLEARTID | CLONE_CHILD_SETTID | SIGCHLD
	//var childTid int
	//var r1 uintptr
	//var r2 uintptr
	//var err error
	r1 := syscall.Syscall(trap, flags, uintptr(0), uintptr(0))

	return int(r1)
}
