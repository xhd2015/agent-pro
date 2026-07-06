package sudosetup

import "os"

// Runner runs subprocesses; tests inject fakes.
type Runner interface {
	Run(name string, args ...string) error
	CombinedOutput(name string, args ...string) ([]byte, error)
}

// TempFile is a temporary file used during sudoers install.
type TempFile interface {
	Name() string
	Write(p []byte) (int, error)
	Close() error
}

// FS abstracts filesystem access for tests and production.
type FS interface {
	UserCacheDir() (string, error)
	Stat(name string) (os.FileInfo, error)
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	Remove(name string) error
	MkdirAll(path string, perm os.FileMode) error
	CreateTemp(dir, pattern string) (TempFile, error)
}