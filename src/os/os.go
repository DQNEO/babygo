package os

import "syscall"

const SYS_EXIT int = 60

var Args []string

type File struct {
	fd int
}


func Create(name string) (*File, interface{}) {
	var O_CREATE_WRITE int = 524866 // O_RDWR|O_CREAT|O_TRUNC|O_CLOEXEC
	var fd int
	fd, _ = syscall.Open(name, O_CREATE_WRITE, 438)
	if fd < 0 {
		panic("unable to create file " + name)
	}

	var f *File = new(File)
	f.fd = fd
	return f, 0
}

func (f *File) Close() {
	syscall.Close(f.fd)
}

func (f *File) Write(p []byte) (int, int) {
	syscall.Write(f.fd, p)
	return 0, 0
}

func init() {
	Args = runtime_args()
}

func Getenv(key string) string {
	v := runtime_getenv(key)
	return v
}

func Exit(status int) {
	syscall.Syscall(uintptr(SYS_EXIT), uintptr(status), 0, 0)
}

func runtime_args() []string
func runtime_getenv(key string) string
