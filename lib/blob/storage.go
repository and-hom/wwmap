package blob

import (
	"errors"
	"github.com/and-hom/wwmap/lib/util"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	Mkdirs  bool
	// 0 - means up to storage root
	// 1 - delete only element
	// 2 - delete element and it's directory if empty
	// 3 - delete element, it's directory if empty and upper directory if empty
	// ....
	DeleteRecursivelyMaxDepth int
}

func (this BasicFsStorage) Store(id string, r io.Reader) error {
	path := this.path(id)
	if this.Mkdirs {
		os.MkdirAll(filepath.Dir(path), os.ModeDir|0755)
	}
	f, err := os.Create(path)
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
	p := strings.Split(id, string(os.PathSeparator))
	if len(p) == 0 {
		return errors.New("Empty path: " + id)
	}

	if err := os.Remove(this.path(id)); err != nil {
		return err
	}

	topLevelOfEmptyDirDetection := 0
	if this.DeleteRecursivelyMaxDepth > 0 {
		topLevelOfEmptyDirDetection = util.Max(len(p)-this.DeleteRecursivelyMaxDepth, 0)
	}
	for i := len(p) - 1; i > topLevelOfEmptyDirDetection; i-- {
		path := this.path(p[:i]...)
		empty, err := IsEmpty(path)
		if err != nil {
			return err
		}
		if !empty {
			return nil
		}
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func (this BasicFsStorage) ListIds() ([]string, error) {
	return this.findFiles(this.BaseDir, "")
}

func (this BasicFsStorage) findFiles(dir string, baseName string) ([]string, error) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return []string{}, err
	}
	result := make([]string, 0, len(infos))
	for i := 0; i < len(infos); i++ {
		if infos[i].IsDir() {
			subEntries, err := this.findFiles(
				filepath.Join(dir, infos[i].Name()),
				filepath.Join(baseName, infos[i].Name()))

			if err != nil {
				return []string{}, err
			}
			result = append(result, subEntries...)
		} else {
			result = append(result, filepath.Join(baseName, infos[i].Name()))
		}
	}
	return result, nil
}

func (this BasicFsStorage) path(id ...string) string {
	return filepath.Join(this.BaseDir, filepath.Join(id...))
}
