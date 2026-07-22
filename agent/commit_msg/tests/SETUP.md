# Scenario

**Feature**: gen-commit-msg generates commit messages via agent runners and optionally commits

```
# opencode path (default)
git repo with staged diff -> gen-commit-msg --agent-runner opencode -> fake-opencode mock events -> parsed message

# commandcode path
git repo with staged diff -> gen-commit-msg --agent-runner commandcode -> llm-mock-run-commandcode (+ hook) -> full stdout JSON

optional --commit -> git commit (must not race with concurrent git from agent)
```

## Preconditions
- The repository contains `cmd/fake-opencode`, `cmd/gen-commit-msg`, `agent/commit_msg`,
  and `agent/llm/llm-mock/llm-mock-run-commandcode`.
- Tests build agent mocks and call `commit_msg.RunGenCommitMsg` (or CLI `-h` subprocess).

## Steps
1. Build `fake-opencode` into the temp directory (default agent binary).
2. Initialize an isolated git repository (optionally with a worktree).
3. Write a leaf-specific mock config / commandcode hook when needed.
4. Run gen-commit-msg with `--agent-runner` and optional `--agent-runner-binary`.

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
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = InitGitRepo
	_ = InitGitRepoWithWorktree
	_ = WriteMockConfig
	_ = WriteFile
	_ = captureRunGenCommitMsg
	_ = captureGenCommitMsgHelp
	req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../.."))
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

func InitGitRepoWithFailingPreCommitHook(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir git dir: %v", err)
	}
	RunGit(t, dir, "init", "--template=")
	RunGit(t, dir, "config", "user.email", "test@example.com")
	RunGit(t, dir, "config", "user.name", "Test User")
	WriteFile(t, filepath.Join(dir, "README.md"), "initial\n")
	RunGit(t, dir, "add", "README.md")
	RunGit(t, dir, "commit", "-m", "initial")
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if err := os.MkdirAll(filepath.Dir(hookPath), 0755); err != nil {
		t.Fatalf("mkdir hooks dir: %v", err)
	}
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatalf("write pre-commit hook: %v", err)
	}
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

	// -h/--help calls os.Exit(0) in less-flags unless HelpNoExit; run the CLI binary
	// in a subprocess so the test process is not terminated.
	if req.Help {
		return captureGenCommitMsgHelp(t, req)
	}

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
	oldCommandCodeHook := os.Getenv("LLM_MOCK_RUN_COMMANDCODE_COMMAND")
	if req.MockConfigPath != "" {
		os.Setenv("FAKE_OPENCODE_MOCK_CONFIG", req.MockConfigPath)
	}
	os.Setenv("OPENCODE_CONFIG_DIR", opencodeConfigDir)
	if req.CommandCodeHook != "" {
		os.Setenv("LLM_MOCK_RUN_COMMANDCODE_COMMAND", req.CommandCodeHook)
	}
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
		if oldCommandCodeHook == "" {
			os.Unsetenv("LLM_MOCK_RUN_COMMANDCODE_COMMAND")
		} else {
			os.Setenv("LLM_MOCK_RUN_COMMANDCODE_COMMAND", oldCommandCodeHook)
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
	if req.NoVerify {
		args = append(args, "--no-verify")
	}
	if req.DryRun {
		args = append(args, "--dry-run")
	}
	if req.AddAll {
		args = append(args, "--add-all")
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

// captureGenCommitMsgHelp runs cmd/gen-commit-msg -h in a subprocess.
func captureGenCommitMsgHelp(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	bin := req.GenCommitMsgBin
	if bin == "" {
		if req.TempDir == "" {
			return nil, fmt.Errorf("help capture requires TempDir or GenCommitMsgBin")
		}
		bin = filepath.Join(req.TempDir, "gen-commit-msg")
		build := exec.Command("go", "build", "-o", bin, "./cmd/gen-commit-msg")
		build.Dir = req.RepoRoot
		if out, err := build.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("build gen-commit-msg: %w\n%s", err, string(out))
		}
		req.GenCommitMsgBin = bin
	}
	cmd := exec.Command(bin, "-h")
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	runErr := cmd.Run()
	resp := &Response{
		Stdout: stdoutBuf.String(),
		Stderr: stderrBuf.String(),
		Err:    runErr,
	}
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = ee.ExitCode()
		} else {
			resp.ExitCode = 1
		}
	}
	return resp, nil
}


```