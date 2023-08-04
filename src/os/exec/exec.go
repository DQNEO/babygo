package exec

import (
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

type Cmd struct {
	Name string
	Args []string
}

func Command(name string, arg ...string) *Cmd {
	return &Cmd{
		Name: name,
		Args: arg,
	}
}

const CLONE_CHILD_CLEARTID uintptr = 2097152 //  0x00200000 // 2097152
const CLONE_CHILD_SETTID uintptr = 16777216  // 0x01000000 // 16777216
const SIGCHLD uintptr = 17

type ExitError struct {
	Status int
}

func (err *ExitError) Error() string {
	sStatus := runtime.Itoa(err.Status)
	return "Command failed: status=" + sStatus
}

func (c *Cmd) Run() error {

	pid := fork()

	if pid == 0 {
		// child
		//		os.Stdout.Write([]byte("\nI am the child\n"))
		//		os.Stdout.Write([]byte("child: " + c.Name + " " + c.Arg + "\n"))
		execve(c.Name, c.Args)
		os.Exit(0)
	} else {
		// parent
		//		os.Stdout.Write([]byte("I am the parent\n"))
		//		spid := runtime.Itoa(pid)
		//		os.Stdout.Write([]byte("parent: child pid=" + spid + "\n"))
		status := wait4(pid)
		if status != 0 {
			return &ExitError{Status: status}
		}
	}
	return nil
}

func fork() uintptr {
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
	return r1
}

func execve(cmds string, args []string) {
	//  int execve(const char *pathname, char *const _Nullable argv[],
	//                  char *const _Nullable envp[]);
	trap := uintptr(59)
	cmd := []byte(cmds)
	cmd = append(cmd, 0)
	pathname := uintptr(unsafe.Pointer(&cmd[0]))
	var argv []uintptr
	argv0 := pathname
	argv = append(argv, argv0)
	for _, arg := range args {
		a := []byte(arg)
		a = append(a, 0)
		argv = append(argv, uintptr(unsafe.Pointer(&a[0])))
	}
	argv = append(argv, 0)
	argvAddr := uintptr(unsafe.Pointer(&argv[0]))
	syscall.Syscall(trap, pathname, argvAddr, uintptr(0))
}

func wait4(pid uintptr) int {
	trap := uintptr(61)
	var status int
	stat_addr := uintptr(unsafe.Pointer(&status))
	r1 := syscall.Syscall(trap, pid, stat_addr, uintptr(0))
	_ = r1
	return status
}
