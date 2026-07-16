# Scenario

**Feature**: gen-commit-msg `--agent-runner=commandcode` via llm-mock-run-commandcode

```
# commandcode runner path (not opencode NDJSON)
staged diff -> gen-commit-msg --agent-runner commandcode --agent-runner-binary <llm-mock-run-commandcode>
  -> binary -p <prompt> --skip-onboarding --yolo --max-turns 1 [-m MODEL]
  -> full stdout agent text -> SanitizeOrError -> format message

# deterministic mock (no live Command Code API)
LLM_MOCK_RUN_COMMANDCODE_COMMAND hook prints fixed {"title","description"} JSON
  (still spawns llm-mock-run-commandcode so argv/binary path is exercised)
```

## Preconditions
- Root harness initialized `TempDir` / `RepoRoot` and built `fake-opencode` (default binary).
- This subtree builds `llm-mock-run-commandcode` and defaults runner to `commandcode`.
- Classic TDD: commandcode support is not implemented yet → leaves RED until implementer.

## Steps
1. Build `./agent/llm/llm-mock/llm-mock-run-commandcode` into TempDir.
2. Set `req.AgentRunner = "commandcode"` and `req.AgentRunnerBinary` to the mock binary.
3. Install default `CommandCodeHook` that prints fixed commit JSON to stdout.
4. Leaves override DryRun / Commit / Help as needed.

## Context
- Default LookPath name for production is `cmd`; tests always override via `--agent-runner-binary`.
- Help leaves use `cmd/gen-commit-msg -h` subprocess (library `RunGenCommitMsg(-h)` would `os.Exit`).
- Generate leaves do not require `FAKE_OPENCODE_MOCK_CONFIG`.

```go
import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// DefaultCommandCodeJSON is the fixed agent stdout for mock generate leaves.
const DefaultCommandCodeJSON = `{"title":"feat: via commandcode","description":"from commandcode mock"}`

// DefaultCommandCodeHook prints DefaultCommandCodeJSON then exits 0.
const DefaultCommandCodeHook = `printf '%s\n' '{"title":"feat: via commandcode","description":"from commandcode mock"}'`

func Setup(t *testing.T, req *Request) error {
	_ = DefaultCommandCodeJSON
	_ = DefaultCommandCodeHook
	_ = StageCommandCodeRepo
	_ = GitHEADSubjectCmd
	if req.TempDir == "" {
		return fmt.Errorf("commandcode subtree requires initialized TempDir from root Setup")
	}
	if req.RepoRoot == "" {
		return fmt.Errorf("commandcode subtree requires RepoRoot from root Setup")
	}

	req.MockCommandCode = filepath.Join(req.TempDir, "llm-mock-run-commandcode")
	build := exec.Command("go", "build", "-o", req.MockCommandCode, "./agent/llm/llm-mock/llm-mock-run-commandcode")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock-run-commandcode: %w\n%s", err, string(out))
	}

	req.GenCommitMsgBin = filepath.Join(req.TempDir, "gen-commit-msg")
	buildCLI := exec.Command("go", "build", "-o", req.GenCommitMsgBin, "./cmd/gen-commit-msg")
	buildCLI.Dir = req.RepoRoot
	if out, err := buildCLI.CombinedOutput(); err != nil {
		return fmt.Errorf("build gen-commit-msg: %w\n%s", err, string(out))
	}

	req.AgentRunner = "commandcode"
	req.AgentRunnerBinary = req.MockCommandCode
	if req.CommandCodeHook == "" {
		req.CommandCodeHook = DefaultCommandCodeHook
	}
	return nil
}

// StageCommandCodeRepo initializes a git repo and stages one text file.
func StageCommandCodeRepo(t *testing.T, req *Request) {
	t.Helper()
	if req.GitDir == "" {
		req.GitDir = filepath.Join(req.TempDir, "repo")
	}
	InitGitRepo(t, req.GitDir)
	WriteFile(t, filepath.Join(req.GitDir, "feature.go"), "package main\n// commandcode feature\n")
	RunGit(t, req.GitDir, "add", "feature.go")
}

func GitHEADSubjectCmd(t *testing.T, gitDir string) string {
	t.Helper()
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = gitDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log -1 --format=%%s: %v\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out))
}
```
