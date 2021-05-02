package command

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
)

type Command interface {
	fmt.Stringer
	Create(args string) CommandExecution
	Name() string
}

type CommandExecution interface {
	GetStreamsOrNils() (io.ReadCloser, io.ReadCloser)
	Execute() error
}

func readerOrNil(pipe io.ReadCloser, err error) io.ReadCloser {
	if err != nil {
		log.Error("Can't get command output stream", err)
		return nil
	}
	return pipe
}
