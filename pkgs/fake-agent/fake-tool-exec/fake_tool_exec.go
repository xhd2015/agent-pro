package faketoolexec

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
)

type MockConfig struct {
	Output   string       `json:"output,omitempty"`
	Stderr   string       `json:"stderr,omitempty"`
	ExitCode *int         `json:"exit_code,omitempty"`
	Content  string       `json:"content,omitempty"`
	Changes  []FileChange `json:"changes,omitempty"`
}

type FileChange struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

func ExecuteBash(command string, workDir string, env []string) (stdout string, stderr string, exitCode int, err error) {
	cmd := exec.Command("bash", "-c", command)
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
	cmd := exec.Command("grep", "-rn", pattern, searchPath)
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
