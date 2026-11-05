//go:build windows

package agentruncli

import (
	"errors"
	"os"
	"os/exec"
)

func replaceCodexProcess(argv, environ []string) error {
	if len(argv) == 0 {
		return errors.New("missing Codex executable")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = environ
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
