package command

import (
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

type FileCommand struct {
	name     string
	fullPath string
}

func (this FileCommand) String() string {
	return this.name + ": " + this.fullPath
}

func (this FileCommand) Create(args string) CommandExecution {
	return fileCommandExecution{
		cmd: exec.Command(this.fullPath, args),
	}
}

func (this FileCommand) Name() string {
	return this.name
}

type fileCommandExecution struct {
	cmd *exec.Cmd
}

func (this fileCommandExecution) GetStreamsOrNils() (io.ReadCloser, io.ReadCloser) {
	return readerOrNil(this.cmd.StdoutPipe()), readerOrNil(this.cmd.StderrPipe())
}

func (this fileCommandExecution) Execute() error {
	if err := this.cmd.Start(); err != nil {
		return err
	}

	if err := this.cmd.Wait(); err != nil {
		return err
	}

	waitStatus, ok := this.cmd.ProcessState.Sys().(syscall.WaitStatus)
	if ok && waitStatus.ExitStatus() > 0 {
		return fmt.Errorf("Command \"%v\" exited with status %d", this.cmd.Args, waitStatus.ExitStatus())
	}
	return nil
}
