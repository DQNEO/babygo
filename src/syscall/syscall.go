package syscall

import (
	"unsafe"

	"runtime"
)

// cheat sheet: https://chromium.googlesource.com/chromiumos/docs/+/HEAD/constants/syscalls.md#x86_64-64_bit
const SYS_READ uintptr = 0
const SYS_WRITE uintptr = 1
const SYS_OPEN uintptr = 2
const SYS_CLOSE uintptr = 3
const SYS_GETDENTS64 uintptr = 217
const SYS_EXIT uintptr = 60

func Getenv(key string) (string, bool) {
	for _, e := range runtime.Envs {
		if e.Key == key {
			return e.Value, true
		}
	}

	return "", false
}

func Environ() []string {
	return runtime.Envlines
}

func Read(fd int, buf []byte) (uintptr, error) {
	p := &buf[0]
	_cap := cap(buf)
	var ret uintptr
	ret, _, _ = Syscall(SYS_READ, uintptr(fd), uintptr(unsafe.Pointer(p)), uintptr(_cap))
	return ret, nil
}

func Open(path string, mode int, perm int) (uintptr, error) {
	buf := []byte(path)
	buf = append(buf, 0) // add null terminator
	p := &buf[0]
	var fd uintptr
	fd, _, _ = Syscall(SYS_OPEN, uintptr(unsafe.Pointer(p)), uintptr(mode), uintptr(perm))
	return fd, nil
}

func Close(fd int) error {
	Syscall(SYS_CLOSE, uintptr(fd), 0, 0)
	return nil
}

func Write(fd int, buf []byte) (uintptr, error) {
	p := &buf[0]
	_len := len(buf)
	var ret uintptr
	ret, _, _ = Syscall(SYS_WRITE, uintptr(fd), uintptr(unsafe.Pointer(p)), uintptr(_len))
	return ret, nil
}

func Getdents(fd int, buf []byte) (int, error) {
	var _p0 unsafe.Pointer
	_p0 = unsafe.Pointer(&buf[0])
	nread, _, _ := Syscall(SYS_GETDENTS64, uintptr(fd), uintptr(_p0), uintptr(len(buf)))
	return int(nread), nil
}

type Err struct {
	Msg string
}

func (err *Err) Error() string {
	return err.Msg
}

func StartProcess(path string, args []string, attr unsafe.Pointer) (uintptr, uintptr, error) {
	pid, err := forkExec(path, args, attr)
	return pid, 0, err
}

func forkExec(path string, args []string, attr unsafe.Pointer) (uintptr, error) {
	pid := fork()
	if pid < 0 {
		return 0, &Err{
			Msg: "[exec] fork failed",
		}
	}
	if pid == 0 {
		// child
		//		os.Stdout.Write([]byte("\nI am the child\n"))
		//		os.Stdout.Write([]byte("child: " + c.Name + " " + c.Arg + "\n"))
		execve(path, args)
		exit(0)
	}

	// parent
	return pid, nil
}

func exit(status int) {
	Syscall(SYS_EXIT, uintptr(status), 0, 0)
}

const CLONE_CHILD_CLEARTID uintptr = 0x00200000
const CLONE_CHILD_SETTID uintptr = 0x01000000
const SIGCHLD uintptr = 0x11

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
	flags := CLONE_CHILD_CLEARTID | CLONE_CHILD_SETTID | SIGCHLD
	// @TODO: use Syscall6 instead of Syscall
	r1, _, _ := Syscall(trap, flags, uintptr(0), uintptr(0))
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

	environ := Environ()
	var envp []uintptr
	for _, envLine := range environ {
		bufEnvLine := []byte(envLine)
		envp = append(envp, uintptr(unsafe.Pointer(&bufEnvLine[0])))
	}
	envp = append(envp, 0)
	envpAddr := uintptr(unsafe.Pointer(&envp[0]))
	Syscall(trap, pathname, argvAddr, envpAddr)
}

//go:linkname Syscall
func Syscall(trap uintptr, a1 uintptr, a2 uintptr, a3 uintptr) (uintptr, uintptr, uintptr)
