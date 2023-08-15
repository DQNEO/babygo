package exec

import (
	"os"
)

type Cmd struct {
	Name    string
	Args    []string
	Process *os.Process
}

func Command(name string, arg ...string) *Cmd {
	return &Cmd{
		Name: name,
		Args: arg,
	}
}

func (c *Cmd) CombinedOutput() ([]byte, error) {
	err := c.Run()
	var buf = []byte("<NONE>") // @TODO Implement me
	return buf, err
}

func (c *Cmd) Run() error {
	err := c.Start()
	if err != nil {
		return err
	}
	return c.Wait()
}

func (c *Cmd) Start() error {
	p, err := os.StartProcess(c.Name, c.Args, nil)
	if err != nil {
		return err
	}
	c.Process = p
	return nil
}

func (c *Cmd) Wait() error {
	state, err := c.Process.Wait()
	_ = state
	return err
}
