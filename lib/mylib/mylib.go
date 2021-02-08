package mylib

import "unsafe"
import "syscall"
import "github.com/DQNEO/babygo/lib/mylib2"

func InArray(x string, list []string) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}

func Sum(a int, b int) int {
	return a + b
}

func Sum2(a int, b int) int {
	return mylib2.Sum2(a, b)
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

var _buf [1024]byte
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

func print_dirp(dirp *linux_dirent) {
	//var reclen int = int(dirp.d_reclen1)

	//fmt.Printf("%p  ", uintptr(dirp))
	//fmt.Printf("%d\t", dirp.d_ino)
	//fmt.Printf("%d\t", dirp.d_off)
	//fmt.Printf("%d\t", dirp.d_type)
	//fmt.Printf("%d\t", reclen)
	//reclen := int(dirp.d_reclen1)
	//fmt.Printf("%d  ", dirp.d_type)
	//p := unsafe.Pointer(&dirp.d_name)
	//var bp *byte = (*byte)(p)
	//var s string = Cstring2string(bp)
	//return
	//myfmt.Printf("%s", s)
	//myfmt.Printf("\n")
}

const O_READONLY_ int = 0

func GetDirents(dir string) []string {
	var entries []string
	var fd int
	fd, _ = syscall.Open(dir, O_READONLY_, 0)

	var buf []byte = _buf[:]
	var counter int
	for {
		nread, _ := syscall.Getdents(int(fd), buf)
		if nread == -1 {
			panic("getdents failed")
		}
		if nread == 0 {
			break
		}

		//fmt.Printf("--------------- nread=%d ---------------\n", nread)
		//fmt.Printf("inode   d_off   d_type  d_reclen    d_name\n")
		var bpos int
		for ; bpos < nread; {
			var dirp *linux_dirent
			p := uintptr(unsafe.Pointer(&buf[0])) + uintptr(bpos)
			dirp = (*linux_dirent)(unsafe.Pointer(p))
			//print_dirp(dirp)
			pp := unsafe.Pointer(&dirp.d_name)
			var bp *byte = (*byte)(pp)
			var s string = Cstring2string(bp)
			entries = append(entries, s)
			bpos = bpos + int(dirp.d_reclen1) // 24 is wrong
			counter++
		}
	}

	return entries
}

func needSwap(a string, b string) bool {
	//myfmt.Printf("# comparing %s <-> %s\n", a, b)
	if len(a) == 0 {
		return false
	}
	var i int
	for i=0;i<len(a);i++ {
		if i == len(b) {
			//myfmt.Printf("#    loose. right is shorter.\n")
			return true
		}
		var aa int = int(a[i])
		var bb int = int(b[i])
		//myfmt.Printf("#    comparing byte at i=%d %s <-> %s\n", i, aa, bb)
		if aa < bb {
			return false
		} else if aa > bb {
			//myfmt.Printf("#    loose at i=%d %s > %s\n", i, aa, bb)
			return true
		}
	}
	return false
}

func SortStrings(ss []string) {
	var i int
	var j int
	for i=0;i<len(ss);i++ {
		for j=0;j<len(ss)-1;j++ {
			a := ss[j]
			b := ss[j+1]
			if needSwap(a, b) {
				//myfmt.Printf("#    loose\n")
				ss[j] = b
				ss[j+1] = a
			} else {
				//myfmt.Printf("#     won\n")
			}
		}
	}
}
