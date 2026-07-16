# Scenario

**Feature**: gen-commit-msg `--add-all` stages like `git add -A` before generate

```
# real path: log then stage, then existing pipeline
untracked/unstaged -> gen-commit-msg --add-all
  -> stderr: $ git add -A
  -> gitwrite.AddAll(dir)
  -> auto-unstage / generate / optional --commit

# dry-run: plan only, no index mutation; count from current index
untracked present, empty index -> gen-commit-msg --add-all --dry-run
  -> stderr: would: git add -A
  -> index unchanged
  -> no staged changes error OK (honest dry-run)

# does not require --commit
```

## Preconditions
- Root harness from `agent/commit_msg/tests/SETUP.md` has initialized `TempDir` /
  `RepoRoot` and built `fake-opencode`.
- Classic TDD: `--add-all` is not implemented yet → leaves RED until implementer.

## Steps
1. Inherit root Setup (fake-opencode binary defaults).
2. Leaf writes untracked/unstaged state and sets `req.AddAll` (+ optional DryRun/Commit).
3. Assert would-line or `$ git add -A`, index/HEAD side effects, and help text.

## Context
- Real log line (stderr): `$ git add -A`
- Dry-run plan line (stderr): `would: git add -A`
- Staging uses `gitwrite.AddAll` (product); tests observe git index outcomes only.

```go
import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// AddAllUntrackedName is the default untracked file used by add-all leaves.
const AddAllUntrackedName = "untracked.go"

func Setup(t *testing.T, req *Request) error {
	_ = AddAllUntrackedName
	_ = InitAddAllRepoWithUntracked
	_ = GitHEADSubjectAddAll
	_ = GitStagedNamesAddAll
	_ = GitHEADFilesAddAll
	if req.TempDir == "" {
		return fmt.Errorf("add-all subtree requires initialized TempDir from root Setup")
	}
	return nil
}

// InitAddAllRepoWithUntracked initializes a git repo with one untracked text file
// and an empty index (nothing newly staged beyond the initial commit).
func InitAddAllRepoWithUntracked(t *testing.T, req *Request) string {
	t.Helper()
	if req.GitDir == "" {
		req.GitDir = filepath.Join(req.TempDir, "repo")
	}
	InitGitRepo(t, req.GitDir)
	name := AddAllUntrackedName
	WriteFile(t, filepath.Join(req.GitDir, name), "package main\n// untracked for --add-all\n")
	return name
}

func GitHEADSubjectAddAll(t *testing.T, gitDir string) string {
	t.Helper()
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = gitDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log -1 --format=%%s: %v\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func GitStagedNamesAddAll(t *testing.T, gitDir string) []string {
	t.Helper()
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = gitDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git diff --cached --name-only: %v\n%s", err, string(out))
	}
	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			names = append(names, line)
		}
	}
	return names
}

// GitHEADFilesAddAll returns paths listed in the latest commit (name-only).
func GitHEADFilesAddAll(t *testing.T, gitDir string) []string {
	t.Helper()
	cmd := exec.Command("git", "show", "--pretty=format:", "--name-only", "HEAD")
	cmd.Dir = gitDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git show --name-only HEAD: %v\n%s", err, string(out))
	}
	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			names = append(names, line)
		}
	}
	return names
}
```
