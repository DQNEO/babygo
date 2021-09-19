package os

import "syscall"
import "unsafe"

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

func Open(name string) (*File, error) {
	var fd int
	fd, _ = syscall.Open(name, O_READONLY, 438)
	if fd < 0 {
		panic("unable to create file " + name)
	}

	f := new(File)
	f.fd = fd
	return f, nil
}

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
			panic("getdents failed")
		}
		if nread == 0 {
			break
		}

		var bpos int
		for bpos < nread {
			var dirp *linux_dirent
			p := uintptr(unsafe.Pointer(&buf[0])) + uintptr(bpos)
			dirp = (*linux_dirent)(unsafe.Pointer(p))
			pp := unsafe.Pointer(&dirp.d_name)
			var bp *byte = (*byte)(pp)
			var s string = Cstring2string(bp)
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
