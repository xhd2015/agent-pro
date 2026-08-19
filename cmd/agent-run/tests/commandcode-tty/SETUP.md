# Scenario

Builds `agent-run` and `llm-mock-run-commandcode` into temp dir. Sets `AGENT_RUN_HOME`.

## Steps

1. Resolve repo root from `d.DOCTEST_ROOT/../../../..`.
2. Build `agent-run` binary.
3. Build `llm-mock-run-commandcode` binary (unless `SkipMockBuild`).
4. Set `AGENT_RUN_HOME` to temp directory.

```go
import (
	"runtime"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	repoRoot, err := findAgentProRoot(d.DOCTEST_ROOT)
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot

	req.TempDir = t.TempDir()
	req.Home = filepath.Join(req.TempDir, ".agent-pro")
	req.Env = append(req.Env, "AGENT_RUN_HOME="+req.Home)

	req.AgentRun = filepath.Join(req.TempDir, "bin", "agent-run")
	if err := os.MkdirAll(filepath.Dir(req.AgentRun), 0755); err != nil {
		return fmt.Errorf("mkdir bin: %w", err)
	}
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", req.AgentRun, "./agent-run")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build agent-run: %v\n%s", err, out)
	}

	if !req.SkipMockBuild {
		req.MockBinary = filepath.Join(req.TempDir, "bin", "llm-mock-run-commandcode")
		build2 := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", req.MockBinary,
			"./agent/llm/llm-mock/llm-mock-run-commandcode")
		build2.Dir = req.RepoRoot
		if out2, err2 := build2.CombinedOutput(); err2 != nil {
			t.Fatalf("build mock: %v\n%s", err2, out2)
		}
	}

	return nil
}

// fakeCommandcodeKeepAliveTUI prints a banner then echoes each injected line.
// Used by --open leaves so real `cmd` one-shot exit cannot race attach.
func fakeCommandcodeKeepAliveTUI() string {
	return `sh -c 'printf "COMMANDCODE_TTY_BANNER\n"; while IFS= read -r line; do printf "%s\n" "$line"; done'`
}

// withCommandcodeOpenTestEnv installs keep-alive fake TUI + instant attach for --open e2e.
func withCommandcodeOpenTestEnv(req *Request) {
	req.Env = append(req.Env,
		"AGENT_RUN_COMMANDCODE_TTY_COMMAND="+fakeCommandcodeKeepAliveTUI(),
		"AGENT_RUN_OPEN_ATTACH_INSTANT=1",
	)
}

func findAgentProRoot(start string) (string, error) {
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = wd
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(line) == "module github.com/xhd2015/agent-pro" {
					return dir, nil
				}
			}
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("could not find agent-pro module root above %s", start)
		}
	}
}

```
