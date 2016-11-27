package main

import (
	"io"
	"os"
)

type Files interface {
	Get(id string) (io.ReadCloser, error)
}

type DummyFiles struct {

}

func (this DummyFiles) Get(string) (io.ReadCloser, error) {
	return os.Open("sample.kml")
}
