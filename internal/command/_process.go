package command

import (
	"net"
	"os"
	"os/exec"
)

type Process struct {
	cmd *exec.Cmd
	ln  net.Listener // could be nil
}

func (c *Process) Close() error {
	p := c.cmd.Process
	if p != nil {
		if err := p.Signal(os.Interrupt); err != nil {
			p.Kill()
		}
	}
	if err := c.cmd.Wait(); err != nil {
		return err
	}
	if err := c.ln.Close(); err != nil {
		return err
	}
	return nil
}

func (p *Process) Wait() error {
	return p.cmd.Wait()
}