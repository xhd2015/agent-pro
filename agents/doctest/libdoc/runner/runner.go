package runner

import (
	"fmt"
	"os"
	"os/exec"

	runnerbuild "github.com/xhd2015/agent-pro/agents/test-case-tree-runner/build"
	"github.com/xhd2015/agent-pro/agents/test-case-tree-runner/core"
)

func Build(dir string) error {
	return runnerbuild.Build(dir, core.Options{RemoveTemp: true})
}

func Test(dir string) error {
	cmd := exec.Command("go", "run", "./agents/test-case-tree-runner", "test", "--rm", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test failed: %w", err)
	}
	return nil
}
