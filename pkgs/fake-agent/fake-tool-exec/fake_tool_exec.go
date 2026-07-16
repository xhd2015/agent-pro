package faketoolexec

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

type MockConfig = types.MockConfig

type FileChange = types.FileChange

// probeCmdTimeout bounds real tool probes so GenerateSession cannot hang CI
// on expensive accidental commands (e.g. repo-wide grep/go test).
const probeCmdTimeout = 2 * time.Second

func ExecuteBash(command string, workDir string, env []string) (stdout string, stderr string, exitCode int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), probeCmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	if workDir != "" {
		cmd.Dir = workDir
	}
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return
}

func ExecuteBashMock(command string, mock MockConfig) (stdout string, stderr string, exitCode int) {
	stdout = mock.Output
	stderr = mock.Stderr
	exitCode = 0
	if mock.ExitCode != nil {
		exitCode = *mock.ExitCode
	}
	return
}

func ExecuteRead(path string) (content string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ExecuteReadMock(mock MockConfig) (content string) {
	return mock.Content
}

func ExecuteWrite(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func ExecuteWriteMock() {
}

func ExecuteGrep(pattern string, searchPath string) (output string, exitCode int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), probeCmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "grep", "-rn", "--max-count=20", pattern, searchPath)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	output = stdoutBuf.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return
}

func ExecuteGrepMock(mock MockConfig) (output string, exitCode int) {
	output = mock.Output
	exitCode = 0
	if mock.ExitCode != nil {
		exitCode = *mock.ExitCode
	}
	return
}
