# Scenario

**Feature**: pure-plan `--dry-run` inspects staged set without agent or mutation

```
# dry-run pure plan: inspect staged files, print mock B, no LLM / no index mutation
staged set -> gen-commit-msg --dry-run -> stdout mock B (N files)
  # binaries would unstage on stderr only
  # --commit would: git commit on stderr only

# validation still applies under dry-run
--agent-runner unknown -> unsupported runner error (before mock success)
empty index -> no staged changes error
```

## Preconditions
- Root harness from `agent/commit_msg/tests/SETUP.md` has initialized `req.TempDir`.
- Pure dry-run success paths do not require a real agent binary (leaves may point
  `--agent-runner-binary` at a non-existent path to prove the agent is not invoked).
- `--dry-run` is not implemented yet (classic TDD: leaves assert intended behavior → RED).

## Steps
1. Inherit harness from root SETUP.md.
2. Leaf stages a known git state, sets `req.DryRun = true`, and optional Commit/NoVerify/Model/AgentRunner.
3. Assert mock stdout, would-lines on stderr, and absence of index/HEAD mutation.

## Context
- Mock message B (stdout, exact): `dry-run: would generate commit message for N staged file(s)\n`
- N counts staged files before unstage (same inspection basis as normal generate).
- Would-unstage and would-commit lines go to stderr only.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	_ = StageDryRunRepo
	_ = StageDryRunRepoWithBinary
	_ = NonExistentAgentBinary
	_ = GitHEADSubject
	_ = GitStagedNames
	_ = AssertNoAgentInvoked
	_ = AssertMockMessageB
	if req.TempDir == "" {
		return fmt.Errorf("dry-run subtree requires initialized TempDir from root Setup")
	}
	return nil
}

// NonExistentAgentBinary returns a path that must not be invoked under pure dry-run.
func NonExistentAgentBinary(req *Request) string {
	return filepath.Join(req.TempDir, "must-not-invoke-agent")
}

// StageDryRunRepo initializes a repo and stages nText text files named change_1.go …
func StageDryRunRepo(t *testing.T, req *Request, nText int) {
	t.Helper()
	if req.GitDir == "" {
		req.GitDir = filepath.Join(req.TempDir, "repo")
	}
	InitGitRepo(t, req.GitDir)
	for i := 1; i <= nText; i++ {
		name := fmt.Sprintf("change_%d.go", i)
		WriteFile(t, filepath.Join(req.GitDir, name), fmt.Sprintf("package p%d\n", i))
		RunGit(t, req.GitDir, "add", name)
	}
}

// StageDryRunRepoWithBinary stages one text file and one ELF-like binary.
// Returns the binary path relative to the repo root (as git reports it).
func StageDryRunRepoWithBinary(t *testing.T, req *Request) (binaryRel string) {
	t.Helper()
	if req.GitDir == "" {
		req.GitDir = filepath.Join(req.TempDir, "repo")
	}
	InitGitRepo(t, req.GitDir)
	WriteFile(t, filepath.Join(req.GitDir, "app.go"), "package main\n")
	RunGit(t, req.GitDir, "add", "app.go")

	binaryRel = "blob.bin"
	binPath := filepath.Join(req.GitDir, binaryRel)
	// Minimal ELF magic so detect.DetectFileType treats it as binary.
	if err := os.WriteFile(binPath, []byte{0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01}, 0755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	RunGit(t, req.GitDir, "add", binaryRel)
	return binaryRel
}

func GitHEADSubject(t *testing.T, gitDir string) string {
	t.Helper()
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = gitDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git log -1 --format=%%s: %v\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func GitStagedNames(t *testing.T, gitDir string) []string {
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

func AssertNoAgentInvoked(t *testing.T, resp *Response) {
	t.Helper()
	for _, marker := range []string{"Passing diff to agent", "Running agent", "agent failed"} {
		if strings.Contains(resp.Stderr, marker) {
			t.Fatalf("agent must not run under --dry-run, found %q in stderr:\n%s", marker, resp.Stderr)
		}
	}
}

// AssertMockMessageB checks stdout is exactly mock message B for N staged files.
func AssertMockMessageB(t *testing.T, stdout string, n int) {
	t.Helper()
	want := fmt.Sprintf("dry-run: would generate commit message for %d staged file(s)\n", n)
	if stdout != want {
		t.Fatalf("stdout mock B mismatch\n got: %q\nwant: %q", stdout, want)
	}
}
```
