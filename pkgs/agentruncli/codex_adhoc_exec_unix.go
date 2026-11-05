//go:build !windows

package agentruncli

import (
	"errors"
	"os/exec"
	"syscall"
)

func replaceCodexProcess(argv, environ []string) error {
	if len(argv) == 0 {
		return errors.New("missing Codex executable")
	}
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return errors.New("cannot resolve Codex executable")
	}
	return syscall.Exec(path, argv, environ)
}
