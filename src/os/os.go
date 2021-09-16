package os

import "syscall"

const SYS_EXIT int = 60
var Args []string

var Stdin *File
var Stdout *File
var Stderr *File

type File struct {
	fd int
}

const O_READONLY int = 0
const FILE_SIZE int = 2000000
const O_CREATE_WRITE int = 524866 // O_RDWR|O_CREAT|O_TRUNC|O_CLOEXEC

func Create(name string) (*File, error) {
	var fd int
	fd, _ = syscall.Open(name, O_CREATE_WRITE, 438)
	if fd < 0 {
		panic("unable to create file " + name)
	}

	f := new(File)
	f.fd = fd
	return f, nil
}

func (f *File) Close() error {
	err := syscall.Close(f.fd)
	return err
}

func (f *File) Write(p []byte) (int, error) {
	syscall.Write(f.fd, p)
	return 0, nil
}

func ReadFile(filename string) ([]uint8, error) {
	// @TODO check error
	var fd int
	fd, _ = syscall.Open(filename, O_READONLY, 0)
	if fd < 0 {
		panic("syscall.Open failed: " + filename)
	}
	var buf = make([]uint8, FILE_SIZE, FILE_SIZE)
	var n int
	n, _ = syscall.Read(fd, buf)
	syscall.Close(fd)
	var readbytes = buf[0:n]
	return readbytes, nil
}

func init() {
	Args = runtime_args()
	Stdin = &File{
		fd: 0,
	}
	Stdout = &File{
		fd: 1,
	}
	Stderr = &File{
		fd: 2,
	}
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
