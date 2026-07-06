package sudosetup

import (
	"strings"
	"testing"
)

func TestEnsureInstalledErrorsWithoutInteractiveStdin(t *testing.T) {
	mgr := &Manager{
		Config: Config{CacheDirName: "test", SudoersName: "test", Username: "testuser"},
		Rule:   Rule{Command: "/bin/true"},
		FS:     guardTestFS{cacheDir: t.TempDir()},
		StdinIsTerminal: func() bool {
			return false
		},
	}
	err := mgr.EnsureInstalled()
	if err == nil {
		t.Fatal("expected error when stdin is not a TTY")
	}
	if !strings.Contains(err.Error(), "TTY") {
		t.Fatalf("error = %q, want TTY hint", err)
	}
}