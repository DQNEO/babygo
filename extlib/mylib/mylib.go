package mylib

import "github.com/DQNEO/babygo/extlib/strings"

func Sum(a int, b int) int {
	return a + b
}

func Atoi(gs string) int {
	if len(gs) == 0 {
		return 0
	}
	var n int

	var isMinus bool
	for _, b := range []uint8(gs) {
		if b == '.' {
			return -999 // @FIXME all no number should return error
		}
		if b == '-' {
			isMinus = true
			continue
		}
		var x uint8 = b - uint8('0')
		n = n * 10
		n = n + int(x)
	}
	if isMinus {
		n = -n
	}

	return n
}

// package path
// "foo/bar/buz" => "foo/bar"
func Dir(path string) string {
	if len(path) == 0 {
		return "."
	}

	if path == "/" {
		return "/"
	}

	found := strings.LastIndexByte(path, '/')
	if found == -1 {
		// not found
		return path
	}

	return path[:found]
}

// "foo/bar/buz" => "buz"
func Base(path string) string {
	if len(path) == 0 {
		return "."
	}

	if path == "/" {
		return "/"
	}

	if path[len(path) - 1] == '/' {
		path = path[0:len(path) - 1]
	}

	found := strings.LastIndexByte(path, '/')
	if found == -1 {
		// not found
		return path
	}

	_len := len(path)
	r := path[found+1:_len]
	return r
}

