package blob

import (
	"io"
	"path/filepath"
	"os"
	"io/ioutil"
)

type BlobStorage interface {
	Store(id string, r io.Reader) error
	Read(id string) (io.ReadCloser, error)
	Length(id string) (int64, error)
	Remove(id string) error
	ListIds() ([]string, error)
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
	return os.Open(this.path(id))
}

func (this BasicFsStorage) Length(id string) (int64, error) {
	stat, err := os.Stat(this.path(id))
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

func (this BasicFsStorage) Remove(id string) error {
	return os.Remove(this.path(id))
}

func (this BasicFsStorage) ListIds() ([]string, error) {
	infos, err := ioutil.ReadDir(this.BaseDir)
	if err != nil {
		return []string{}, err
	}
	result := make([]string, 0, len(infos))
	for i := 0; i < len(infos); i++ {
		result = append(result, infos[i].Name())
	}
	return result, nil
}

func (this BasicFsStorage) path(id string) string {
	return filepath.Join(this.BaseDir, id)
}
