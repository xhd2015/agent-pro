package sudosetup

import (
	"os"
	"strings"
	"testing"
)

type guardTestFS struct {
	cacheDir string
}

func (f guardTestFS) UserCacheDir() (string, error) { return f.cacheDir, nil }
func (f guardTestFS) Stat(name string) (os.FileInfo, error) {
	return nil, os.ErrNotExist
}
func (f guardTestFS) ReadFile(name string) ([]byte, error)   { return nil, os.ErrNotExist }
func (f guardTestFS) WriteFile(string, []byte, os.FileMode) error { return nil }
func (f guardTestFS) Remove(name string) error             { return os.ErrNotExist }
func (f guardTestFS) MkdirAll(string, os.FileMode) error     { return nil }
func (f guardTestFS) CreateTemp(string, string) (TempFile, error) {
	return nil, os.ErrInvalid
}

func TestManagerPanicsWithoutInjectedFSInTests(t *testing.T) {
	m := &Manager{
		Config: Config{CacheDirName: "test", SudoersName: "test"},
		Rule:   Rule{Command: "/bin/true"},
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when Manager.FS is nil under go test")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("panic value type %T, want string", r)
		}
		if !strings.Contains(msg, "Manager.FS must be injected") {
			t.Fatalf("panic = %q, want FS injection hint", msg)
		}
	}()
	_, _ = m.IsInstalled()
}

func TestManagerPanicsWithoutInjectedRunnerInTests(t *testing.T) {
	m := &Manager{
		Config: Config{CacheDirName: "test", SudoersName: "test", Username: "testuser"},
		Rule:   Rule{Command: "/bin/true"},
		FS:     guardTestFS{cacheDir: t.TempDir()},
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when Manager.Runner is nil under go test")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("panic value type %T, want string", r)
		}
		if !strings.Contains(msg, "Manager.Runner must be injected") {
			t.Fatalf("panic = %q, want Runner injection hint", msg)
		}
	}()
	_ = m.Detect()
}