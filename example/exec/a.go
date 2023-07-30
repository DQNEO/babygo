package main

import "os/exec"

func main() {
	err := exec.Command("/usr/bin/mkdir", "/tmp/0043").Run()
	if err != nil {
		panic(err)
	}
}
