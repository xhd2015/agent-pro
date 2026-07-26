# Scenario

**Feature**: git_runner classifies transient index write/lock errors and retries commits

```
# classify
doctest -> IsTransientIndexError(output) -> true|false

# CommitWithRetry recovery
interference (stale lock | uchg-then-clear | hook fail)
  -> RemoveStaleIndexLock + git commit (retry loop)
doctest <- success subject or non-transient error
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/agent/git_runner` exports
  `IsTransientIndexError`, `CommitWithRetry`, `RemoveStaleIndexLock`,
  `IndexLockPath`, and `Commit`.
- `git` is available on PATH for commit-with-retry leaves.
- Leaves set `req.Mode` to `classify` or `commit-with-retry`.
- Darwin-only leaf `unable-to-write-then-clear` uses `chflags uchg` on `.git/index`.

## Steps

1. Root `Setup` records module root for path resolution if needed.
2. Leaf `Setup` fills classify strings or initializes a temp git repo + interference.
3. `Run` dispatches on `Mode` and returns classifier result or commit outcome.
4. Leaf `Assert` checks transient flag or commit subject / fail-fast error.

## Context

- Mirrors inspect script `script/debug/git-index-write-retry`.
- Unit tests in `gitrunner_test.go` cover the same classifier table; doctests lock
  the production error string and recovery paths at the doc-style layer.
- Repos use `core.hooksPath=/dev/null` unless the leaf installs a failing hook.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Keep shared helpers referenced so classify-only packages still compile
	// the full helper set from root SETUP.
	_ = initGitRepo
	_ = initGitRepoWithFailingHook
	_ = stageNewFile
	_ = runGit
	_ = ensureGitAvailable
	if req.MaxAttempts == 0 {
		req.MaxAttempts = 5
	}
	return nil
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_CONFIG_NOSYSTEM=1",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %s: %v", strings.Join(args, " "), string(out), err)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	runGit(t, dir, "init", "--template=", "-b", "main")
	runGit(t, dir, "config", "user.email", "doctest@example.com")
	runGit(t, dir, "config", "user.name", "Doctest")
	runGit(t, dir, "config", "core.hooksPath", "/dev/null")
	seed := filepath.Join(dir, "seed.txt")
	if err := os.WriteFile(seed, []byte("seed\n"), 0644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	runGit(t, dir, "add", "seed.txt")
	runGit(t, dir, "commit", "-m", "seed")
}

func initGitRepoWithFailingHook(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	runGit(t, dir, "init", "--template=", "-b", "main")
	runGit(t, dir, "config", "user.email", "doctest@example.com")
	runGit(t, dir, "config", "user.name", "Doctest")
	// Do not null out hooksPath — leaf needs a real pre-commit failure.
	seed := filepath.Join(dir, "seed.txt")
	if err := os.WriteFile(seed, []byte("seed\n"), 0644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	runGit(t, dir, "add", "seed.txt")
	runGit(t, dir, "commit", "-m", "seed")
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if err := os.MkdirAll(filepath.Dir(hookPath), 0755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho 'husky - pre-commit hook exited with code 1' >&2\nexit 1\n"), 0755); err != nil {
		t.Fatalf("write pre-commit: %v", err)
	}
}

func stageNewFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	runGit(t, dir, "add", name)
}

// ensureGitAvailable fails early with a clear message if git is missing.
func ensureGitAvailable(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Fatalf("git not on PATH: %v", err)
	}
}
```
