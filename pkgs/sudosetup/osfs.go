package sudosetup

import (
	"io"
	"os"
)

type osFS struct{}

func (osFS) UserCacheDir() (string, error) {
	return os.UserCacheDir()
}

func (osFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (osFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (osFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (osFS) Remove(name string) error {
	return os.Remove(name)
}

func (osFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (osFS) CreateTemp(dir, pattern string) (TempFile, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, err
	}
	return &osTempFile{f}, nil
}

type osTempFile struct {
	*os.File
}

func (f *osTempFile) Write(p []byte) (int, error) {
	return f.File.Write(p)
}

func (f *osTempFile) Close() error {
	return f.File.Close()
}

var _ io.Closer = (*osTempFile)(nil)