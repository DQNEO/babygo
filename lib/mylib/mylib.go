package mylib

import "os"
import "github.com/DQNEO/babygo/lib/mylib2"

type Type struct {
	Field int
}

func (mt *Type) Method() int {
	return mt.Field
}

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

func Readdirnames(dir string) ([]string, error) {
	f, _ := os.Open(dir)
	if f.Fd() < 0 {
		panic("cannot open " + dir)
	}

	names, _ := f.Readdirnames(-1)
	f.Close()
	return names, nil
}

func needSwap(a string, b string) bool {
	//fmt.Printf("# comparing %s <-> %s\n", a, b)
	if len(a) == 0 {
		return false
	}
	var i int
	for i = 0; i < len(a); i++ {
		if i == len(b) {
			//fmt.Printf("#    loose. right is shorter.\n")
			return true
		}
		var aa int = int(a[i])
		var bb int = int(b[i])
		//fmt.Printf("#    comparing byte at i=%d %s <-> %s\n", i, aa, bb)
		if aa < bb {
			return false
		} else if aa > bb {
			//fmt.Printf("#    loose at i=%d %s > %s\n", i, aa, bb)
			return true
		}
	}
	return false
}

func SortStrings(ss []string) {
	var i int
	var j int
	for i = 0; i < len(ss); i++ {
		for j = 0; j < len(ss)-1; j++ {
			a := ss[j]
			b := ss[j+1]
			if needSwap(a, b) {
				//fmt.Printf("#    loose\n")
				ss[j] = b
				ss[j+1] = a
			} else {
				//fmt.Printf("#     won\n")
			}
		}
	}
}
