# Scenario

**Feature**: gen-commit-msg generates commit messages via fake-opencode and optionally commits

```
git repo with staged diff -> gen-commit-msg -> fake-opencode mock events -> parsed message
optional --commit -> git commit (must not race with concurrent git from agent)
```

## Preconditions
- The repository contains `cmd/fake-opencode` and `agent/commit_msg`.
- Tests build `fake-opencode` and call `commit_msg.RunGenCommitMsg` or `commit_msg.Generate`.

## Steps
1. Build `fake-opencode` into the temp directory.
2. Initialize an isolated git repository (optionally with a worktree).
3. Write a leaf-specific mock config when needed.
4. Run gen-commit-msg with `--agent-runner=opencode` and `--agent-runner-binary`.

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/agent/commit_msg"
)

func Setup(t *testing.T, req *Request) error {
	_ = InitGitRepo
	_ = InitGitRepoWithWorktree
	_ = WriteMockConfig
	_ = WriteFile
	_ = captureRunGenCommitMsg
	req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "../../.."))
	if _, err := os.Stat(filepath.Join(req.RepoRoot, "go.mod")); err != nil {
		return fmt.Errorf("repo root not found: %w", err)
	}
	req.TempDir = t.TempDir()
	req.FakeOpencode = filepath.Join(req.TempDir, "fake-opencode")
	build := exec.Command("go", "build", "-o", req.FakeOpencode, "./cmd/fake-opencode")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build fake-opencode: %w\n%s", err, string(out))
	}
	if req.AgentRunner == "" {
		req.AgentRunner = "opencode"
	}
	if req.AgentRunnerBinary == "" {
		req.AgentRunnerBinary = req.FakeOpencode
	}
	if req.Model == "" {
		req.Model = "openai/gpt-5"
	}
	return nil
}

func RunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s in %s failed: %s: %v", strings.Join(args, " "), dir, string(out), err)
	}
}

func WriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func InitGitRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir git dir: %v", err)
	}
	RunGit(t, dir, "init", "--template=")
	RunGit(t, dir, "config", "user.email", "test@example.com")
	RunGit(t, dir, "config", "user.name", "Test User")
	RunGit(t, dir, "config", "core.hooksPath", "/dev/null")
	WriteFile(t, filepath.Join(dir, "README.md"), "initial\n")
	RunGit(t, dir, "add", "README.md")
	RunGit(t, dir, "commit", "-m", "initial")
}

func InitGitRepoWithWorktree(t *testing.T, mainDir, worktreeDir string) {
	t.Helper()
	InitGitRepo(t, mainDir)
	if err := os.MkdirAll(filepath.Dir(worktreeDir), 0755); err != nil {
		t.Fatalf("mkdir worktree parent: %v", err)
	}
	RunGit(t, mainDir, "worktree", "add", worktreeDir, "-b", "feature/worktree-test")
}

func WriteMockConfig(t *testing.T, req *Request, body string) {
	t.Helper()
	req.MockConfigPath = filepath.Join(req.TempDir, "mock.json")
	WriteFile(t, req.MockConfigPath, body)
}

func captureRunGenCommitMsg(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	gitDir := req.GitDir
	if gitDir == "" {
		gitDir = filepath.Join(req.TempDir, "repo")
		req.GitDir = gitDir
	}

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = stdoutW
	os.Stderr = stderrW

	opencodeConfigDir := filepath.Join(req.TempDir, "opencode-config")
	_ = os.MkdirAll(opencodeConfigDir, 0755)
	oldMockConfig := os.Getenv("FAKE_OPENCODE_MOCK_CONFIG")
	oldOpencodeConfigDir := os.Getenv("OPENCODE_CONFIG_DIR")
	if req.MockConfigPath != "" {
		os.Setenv("FAKE_OPENCODE_MOCK_CONFIG", req.MockConfigPath)
	}
	os.Setenv("OPENCODE_CONFIG_DIR", opencodeConfigDir)
	defer func() {
		if oldMockConfig == "" {
			os.Unsetenv("FAKE_OPENCODE_MOCK_CONFIG")
		} else {
			os.Setenv("FAKE_OPENCODE_MOCK_CONFIG", oldMockConfig)
		}
		if oldOpencodeConfigDir == "" {
			os.Unsetenv("OPENCODE_CONFIG_DIR")
		} else {
			os.Setenv("OPENCODE_CONFIG_DIR", oldOpencodeConfigDir)
		}
	}()

	args := []string{
		"--dir", gitDir,
		"--model", req.Model,
		"--agent-runner", req.AgentRunner,
		"--agent-runner-binary", req.AgentRunnerBinary,
	}
	if req.Commit {
		args = append(args, "--commit")
	}

	runErr := commit_msg.RunGenCommitMsg(args)

	stdoutW.Close()
	stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var stdoutBuf, stderrBuf bytes.Buffer
	_, _ = stdoutBuf.ReadFrom(stdoutR)
	_, _ = stderrBuf.ReadFrom(stderrR)

	resp := &Response{
		Stdout: stdoutBuf.String(),
		Stderr: stderrBuf.String(),
		Err:    runErr,
	}
	if runErr != nil {
		resp.ExitCode = 1
	}
	lines := strings.Split(strings.TrimSpace(resp.Stdout), "\n")
	if len(lines) > 0 && lines[0] != "" {
		resp.Message = lines[0]
	}
	return resp, nil
}


```