package os

import "syscall"
import "unsafe"

const SYS_EXIT int = 60

var Args []string

var Stdin *File = &File{
	fd: 0,
}
var Stdout *File = &File{
	fd: 1,
}
var Stderr *File = &File{
	fd: 2,
}

type File struct {
	fd int
}

const FILE_SIZE int = 2000000

const O_READONLY int = 0
const O_RDWR int = 0x2
const O_CREATE int = 0x40
const O_TRUNC int = 0x200
const O_CLOSEXEC int = 0x80000

type PathError struct {
	Err string
}

func (e *PathError) Error() string {
	return e.Err
}

func Open(name string) (*File, error) {
	fd, _ := syscall.Open(name, O_READONLY, 438)
	if fd < 0 {
		e := &PathError{Err: "open " + name + ": no such file or directory"}
		return nil, e
	}

	f := &File{
		fd: fd,
	}
	return f, nil
}

func Create(name string) (*File, error) {
	fd, _ := syscall.Open(name, O_RDWR|O_CREATE|O_TRUNC|O_CLOSEXEC, 438)
	if fd < 0 {
		e := &PathError{Err: "open " + name + ": no such file or directory"}
		return nil, e
	}

	f := &File{
		fd: fd,
	}
	return f, nil
}

func (f *File) Fd() uintptr {
	return uintptr(f.fd)
}

func Cstring2string(b *byte) string {
	var bs []byte
	for {
		if b == nil || *b == 0 {
			break
		}
		bs = append(bs, *b)
		p := uintptr(unsafe.Pointer(b)) + 1
		b = (*byte)(unsafe.Pointer(p))
	}
	return string(bs)
}

// Translation of http://man7.org/linux/man-pages/man2/getdents64.2.html#top_of_page

//struct linux_dirent64 {
//	ino64_t        d_ino;    // 8 bytes: 64-bit inode number
//	off64_t        d_off;    // 8 bytes: 64-bit offset to next structure
//	unsigned short d_reclen; // 2 bytes: Size of this dirent
//	unsigned char  d_type;   // 1 byte: File type
//	char           d_name[]; // Filename (null-terminated)
//};

type linux_dirent struct {
	d_ino     int
	d_off     int
	d_reclen1 uint16
	d_type    byte
	d_name    byte
}

func (f *File) Readdirnames(n int) ([]string, error) {
	var fd uintptr = f.Fd()
	var buf []byte = make([]byte, 1024, 1024)
	var counter int
	var entries []string
	for {
		nread, _ := syscall.Getdents(int(fd), buf)
		if nread == -1 {
			e := &PathError{Err: "Getdents failed"}
			return nil, e
		}
		if nread == 0 {
			break
		}

		var bpos int
		for bpos < nread {
			var dirp *linux_dirent
			p := uintptr(unsafe.Pointer(&buf[0])) + uintptr(bpos)
			dirp = (*linux_dirent)(unsafe.Pointer(p))
			var bytes *byte = &dirp.d_name
			var s string = Cstring2string(bytes)
			bpos = bpos + int(dirp.d_reclen1)
			counter++
			if s == "." || s == ".." {
				continue
			}
			entries = append(entries, s)
		}
	}
	f.Close()
	return entries, nil
}

func (f *File) Close() error {
	syscall.Close(f.fd) // @TODO return error
	return nil
}

func (f *File) Write(p []byte) (int, error) {
	n, _ := syscall.Write(f.fd, p)
	if n < 0 {
		e := &PathError{Err: "Write failed"} // @TODO use appropriate error type
		return 0, e
	}
	return n, nil
}

func ReadFile(filename string) ([]uint8, error) {
	f, err := Open(filename)
	if err != nil {
		return nil, err
	}
	var buf = make([]uint8, FILE_SIZE, FILE_SIZE)
	var n int
	n, _ = syscall.Read(f.fd, buf)
	if n < 0 {
		e := &PathError{Err: "Read failed"} // @TODO use appropriate error type
		return nil, e
	}
	f.Close()
	read := buf[0:n]
	return read, nil
}

func init() {
	Args = runtime_args()
}

func Getenv(key string) string {
	v, _ := syscall.Getenv(key)
	return v
}

func Environ() []string {
	return syscall.Environ()
}

type Process struct {
	Pid int
}

func (p *Process) Wait() (int, error) {
	return p.wait()
}

func (p *Process) wait() (int, error) {
	var st int
	status, _ := syscall.Wait4(p.Pid, &st, 0, 0)
	if status != 0 {
		return 0, &ExitError{Status: status}
	}
	return 0, nil
}

type ExitError struct {
	Status int
}

func (err *ExitError) Error() string {
	return "[ExitError] Command failed"
}

func startProcess(path string, args []string, attr unsafe.Pointer) (*Process, error) {
	pid, _, err := syscall.StartProcess(path, args, attr)
	p := &Process{Pid: pid}
	return p, err
}

func StartProcess(path string, args []string, attr unsafe.Pointer) (*Process, error) {
	p, err := startProcess(path, args, attr)
	return p, err
}

func Exit(status int) {
	syscall.Syscall(uintptr(SYS_EXIT), uintptr(status), 0, 0)
}

//go:linkname runtime_args runtime.runtime_args
func runtime_args() []string
