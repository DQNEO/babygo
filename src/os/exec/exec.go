package exec

import (
	"os"
	"syscall"
	"unsafe"

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
		execve(c.Name, c.Arg)
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
	// strace:
	// clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|CLONE_CHILD_SETTID|SIGCHLD, child_tidptr=0x7feb85d62a10) = 42420
	//
	// https://man7.org/linux/man-pages/man2/clone.2.html
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

func execve(cmds string, argv1s string) {
	//  int execve(const char *pathname, char *const _Nullable argv[],
	//                  char *const _Nullable envp[]);
	trap := uintptr(59)
	cmd := []byte(cmds)
	cmd = append(cmd, 0)
	pathname := uintptr(unsafe.Pointer(&cmd[0]))
	var argv []uintptr
	argv0 := pathname
	argv1 := []byte(argv1s)
	argv1 = append(argv1, 0)
	argv = append(argv, argv0)
	argv = append(argv, uintptr(unsafe.Pointer(&argv1[0])))
	argv = append(argv, 0)
	argvAddr := uintptr(unsafe.Pointer(&argv[0]))
	syscall.Syscall(trap, pathname, argvAddr, uintptr(0))
}
