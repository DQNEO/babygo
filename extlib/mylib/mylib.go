package mylib

import "github.com/DQNEO/babygo/extlib/mylib2"
import "github.com/DQNEO/babygo/extlib/strings"


func Sum(a int, b int) int {
	return a + b
}

func Sum2(a int, b int) int {
	return mylib2.Sum2(a, b)
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

