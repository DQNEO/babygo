package exec

import (
	"os"
	"syscall"
	"unsafe"
)

type Cmd struct {
	Name     string
	Args     []string
	childPid uintptr
}

func Command(name string, arg ...string) *Cmd {
	return &Cmd{
		Name: name,
		Args: arg,
	}
}

type ExitError struct {
	Status int
}

func (err *ExitError) Error() string {
	return "[ExitError] Command failed"
}

func (c *Cmd) Run() error {
	err := c.Start()
	if err != nil {
		return err
	}
	return c.Wait()
}

func (c *Cmd) Start() error {
	pid, err := os.StartProcess(c.Name, c.Args, nil)
	if err != nil {
		return err
	}
	c.childPid = pid
	return nil
}

func (c *Cmd) Wait() error {
	status := wait4(c.childPid)
	if status != 0 {
		return &ExitError{Status: status}
	}
	return nil
}

func wait4(pid uintptr) int {
	trap := uintptr(61)
	var status int
	stat_addr := uintptr(unsafe.Pointer(&status))
	r1, _, _ := syscall.Syscall(trap, pid, stat_addr, uintptr(0))
	_ = r1
	return status
}
