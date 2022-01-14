package blob_test

import (
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

const BASE = "/tmp/fs-storage-test"

func TestBasicFsStorageRecursivelyRemoveDepth1(t *testing.T) {
	os.RemoveAll(BASE)
	storage := blob.BasicFsStorage{
		BaseDir:                   BASE,
		Mkdirs:                    false,
		DeleteRecursivelyMaxDepth: 1,
	}

	err := os.MkdirAll("/tmp/fs-storage-test/a/b/", os.ModeDir|0755)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Create("/tmp/fs-storage-test/a/b/c")
	if err != nil {
		log.Fatal(err)
	}
	if _, err = os.Stat("/tmp/fs-storage-test/a/b/c"); os.IsNotExist(err) {
		log.Fatal(err)
	}

	err = storage.Remove("a/b/c")
	assert.Nil(t, err)

	assertNotExists(t, "/tmp/fs-storage-test/a/b/c")
	assertExists(t, "/tmp/fs-storage-test/a/b")
}

func TestBasicFsStorageRecursivelyRemoveDepth2(t *testing.T) {
	os.RemoveAll(BASE)
	storage := blob.BasicFsStorage{
		BaseDir:                   BASE,
		Mkdirs:                    false,
		DeleteRecursivelyMaxDepth: 2,
	}

	err := os.MkdirAll("/tmp/fs-storage-test/a/b/", os.ModeDir|0755)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Create("/tmp/fs-storage-test/a/b/c")
	if err != nil {
		log.Fatal(err)
	}
	if _, err = os.Stat("/tmp/fs-storage-test/a/b/c"); os.IsNotExist(err) {
		log.Fatal(err)
	}

	err = storage.Remove("a/b/c")
	assert.Nil(t, err)

	assertNotExists(t, "/tmp/fs-storage-test/a/b")
	assertExists(t, "/tmp/fs-storage-test/a")
}

func TestBasicFsStorageRecursivelyRemoveDepthUnlimited(t *testing.T) {
	os.RemoveAll(BASE)
	storage := blob.BasicFsStorage{
		BaseDir: BASE,
		Mkdirs:  false,
	}

	err := os.MkdirAll("/tmp/fs-storage-test/a/b/", os.ModeDir|0755)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Create("/tmp/fs-storage-test/a/b/c")
	if err != nil {
		log.Fatal(err)
	}
	if _, err = os.Stat("/tmp/fs-storage-test/a/b/c"); os.IsNotExist(err) {
		log.Fatal(err)
	}

	err = storage.Remove("a/b/c")
	assert.Nil(t, err)

	assertNotExists(t, "/tmp/fs-storage-test/a")
	assertExists(t, "/tmp/fs-storage-test")
}

func assertNotExists(t *testing.T, path string) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		assert.Fail(t, path+" exists!")
	}
}

func assertExists(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		assert.Fail(t, path+" not exists!")
	}
}

func TestExists(t *testing.T) {
	os.RemoveAll(BASE)
	os.MkdirAll(BASE, os.ModeDir|0755)

	storage := blob.BasicFsStorage{
		BaseDir: BASE,
		Mkdirs:  true,
	}

	_, err := os.Create(BASE + "/a")
	if err != nil {
		log.Fatal(err)
	}
	if _, err = os.Stat("/tmp/fs-storage-test/a"); os.IsNotExist(err) {
		log.Fatal(err)
	}

	exists, err := storage.Exists("a")
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestNotExists(t *testing.T) {
	os.RemoveAll(BASE)
	os.MkdirAll(BASE, os.ModeDir|0755)

	storage := blob.BasicFsStorage{
		BaseDir: BASE,
		Mkdirs:  true,
	}

	exists, err := storage.Exists("a")
	assert.Nil(t, err)
	assert.False(t, exists)
}
