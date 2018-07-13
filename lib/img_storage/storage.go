package img_storage

import (
	"io"
	"path/filepath"
	"os"
)

type ImgStorage interface {
	Store(id string, r io.Reader) error
	Read(id string) (io.ReadCloser, error)
}

type BasicFsStorage struct {
	BaseDir string
}

func (this BasicFsStorage) Store(id string, r io.Reader) error {
	f, err := os.Create(this.path(id))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (this BasicFsStorage) Read(id string) (io.ReadCloser, error) {
	return  os.Open(this.path(id))
}

func (this BasicFsStorage) path(id string) string {
	return filepath.Join(this.BaseDir, id)
}
