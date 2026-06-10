package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	runnerbuild "github.com/xhd2015/agent-pro/agents/test-case-tree-runner/build"
	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/core"
)

func Build(dir string) error {
	return runnerbuild.Build(dir, core.Options{RemoveTemp: true})
}

func BuildArgs(args []string) error {
	return runTestCaseTreeRunner("build", args)
}

func Test(args []string) error {
	return runTestCaseTreeRunner("test", args)
}

func runTestCaseTreeRunner(command string, args []string) error {
	root, err := findRepoRoot()
	if err != nil {
		return err
	}
	runnerDir := filepath.Join(root, "agents", "test-case-tree-runner")
	cmdArgs := append([]string{"run", runnerDir, command, "--rm"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", command, err)
	}
	return nil
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		if isAgentProRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("could not locate agent-pro repository root from %s", wd)
}

func isAgentProRoot(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return false
	}
	if !bytes.Contains(data, []byte("module github.com/xhd2015/agent-pro")) {
		return false
	}
	if _, err := os.Stat(filepath.Join(dir, "agents", "test-case-tree-runner")); err != nil {
		return false
	}
	return true
}
