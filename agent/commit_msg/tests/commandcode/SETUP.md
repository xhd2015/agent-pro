# Scenario

**Feature**: gen-commit-msg `--agent-runner=commandcode` via llm-mock-run-commandcode

```
# commandcode runner path (not opencode NDJSON)
staged diff -> gen-commit-msg --agent-runner commandcode --agent-runner-binary <llm-mock-run-commandcode|argv-recorder>
  -> write full commit prompt to temp file (shared delivery path)
  -> binary -p <short: read ABS_PATH …> --skip-onboarding --yolo --max-turns 16 [-m MODEL]
  -> full stdout agent text -> SanitizeOrError -> format message
  -> delete temp prompt file after agent completes

# deterministic mock (no live Command Code API)
LLM_MOCK_RUN_COMMANDCODE_COMMAND hook prints fixed {"title","description"} JSON
  (still spawns llm-mock-run-commandcode so argv/binary path is exercised)

# argv recorder leaves (prompt file delivery)
shell recorder writes NUL-separated argv + optional prompt-file copy; prints fixed JSON
```

## Preconditions
- Root harness initialized `TempDir` / `RepoRoot` and built `fake-opencode` (default binary).
- This subtree builds `llm-mock-run-commandcode` and defaults runner to `commandcode`.
- Shared file-based prompt delivery is classic TDD: `short-p-argv` stays RED until implementer.

## Steps
1. Build `./agent/llm/llm-mock/llm-mock-run-commandcode` into TempDir.
2. Set `req.AgentRunner = "commandcode"` and `req.AgentRunnerBinary` to the mock binary.
3. Install default `CommandCodeHook` that prints fixed commit JSON to stdout.
4. Leaves override DryRun / Commit / Help / argv recorder as needed.

## Context
- Default LookPath name for production is `cmd`; tests always override via `--agent-runner-binary`.
- Help leaves use `cmd/gen-commit-msg -h` subprocess (library `RunGenCommitMsg(-h)` would `os.Exit`).
- Generate leaves do not require `FAKE_OPENCODE_MOCK_CONFIG`.
- `-p` must stay short (no embedded multi-line `diff --git` body); full prompt lives in a temp file.

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

// DefaultCommandCodeJSON is the fixed agent stdout for mock generate leaves.
const DefaultCommandCodeJSON = `{"title":"feat: via commandcode","description":"from commandcode mock"}`

// DefaultCommandCodeHook prints DefaultCommandCodeJSON then exits 0.
const DefaultCommandCodeHook = `printf '%s\n' '{"title":"feat: via commandcode","description":"from commandcode mock"}'`

// CommandCodeArgvRecorderJSON is the fixed JSON printed by the argv recorder binary.
const CommandCodeArgvRecorderJSON = `{"title":"feat: short-p argv","description":"prompt via temp file"}`

// MaxCommandCodePromptArgBytes is the tight upper bound for the `-p` value once
// the full commit prompt is delivered via temp file instead of argv.
const MaxCommandCodePromptArgBytes = 2048

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = DefaultCommandCodeJSON
	_ = DefaultCommandCodeHook
	_ = CommandCodeArgvRecorderJSON
	_ = MaxCommandCodePromptArgBytes
	_ = StageCommandCodeRepo
	_ = StageCommandCodeRepoLargeDiff
	_ = GitHEADSubjectCmd
	_ = InstallCommandCodeArgvRecorder
	_ = ParseNULSeparatedArgs
	_ = ArgvValueAfter
	if req.TempDir == "" {
		return fmt.Errorf("commandcode subtree requires initialized TempDir from root Setup")
	}
	if req.RepoRoot == "" {
		return fmt.Errorf("commandcode subtree requires RepoRoot from root Setup")
	}

	req.MockCommandCode = filepath.Join(req.TempDir, "llm-mock-run-commandcode")
	build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-o", req.MockCommandCode, "./agent/llm/llm-mock/llm-mock-run-commandcode")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build llm-mock-run-commandcode: %w\n%s", err, string(out))
	}

	req.GenCommitMsgBin = filepath.Join(req.TempDir, "gen-commit-msg")
	buildCLI := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", req.GenCommitMsgBin, "./gen-commit-msg")
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

// StageCommandCodeRepoLargeDiff stages a multi-line file so the unified staged
// diff (and thus the full commit prompt) clearly contains `diff --git` bodies.
func StageCommandCodeRepoLargeDiff(t *testing.T, req *Request) {
	t.Helper()
	if req.GitDir == "" {
		req.GitDir = filepath.Join(req.TempDir, "repo")
	}
	InitGitRepo(t, req.GitDir)
	var b strings.Builder
	b.WriteString("package main\n\n// LARGE_STAGED_DIFF_MARKER for prompt-file delivery tests\n")
	for i := 0; i < 80; i++ {
		fmt.Fprintf(&b, "// line %03d: padded content to enlarge unified diff body\n", i)
	}
	b.WriteString("func main() {}\n")
	WriteFile(t, filepath.Join(req.GitDir, "large_feature.go"), b.String())
	RunGit(t, req.GitDir, "add", "large_feature.go")
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

// InstallCommandCodeArgvRecorder writes a shell recorder as AgentRunnerBinary.
// The recorder NUL-separates argv to req.CommandCodeArgvPath, copies a prompt
// file referenced by `-p` (path arg or path token in the short instruction) to
// req.CommandCodePromptCopyPath when present, and prints CommandCodeArgvRecorderJSON.
func InstallCommandCodeArgvRecorder(t *testing.T, req *Request) {
	t.Helper()
	if req.TempDir == "" {
		t.Fatal("InstallCommandCodeArgvRecorder requires TempDir")
	}
	if req.CommandCodeArgvPath == "" {
		req.CommandCodeArgvPath = filepath.Join(req.TempDir, "cmd-argv.nul")
	}
	if req.CommandCodePromptCopyPath == "" {
		req.CommandCodePromptCopyPath = filepath.Join(req.TempDir, "prompt-file-copy.txt")
	}
	recorder := filepath.Join(req.TempDir, "cmd-argv-recorder")
	// Shell recorder: capture argv; best-effort copy of prompt file while agent runs
	// (production deletes the temp prompt after the process exits).
	script := fmt.Sprintf(`#!/bin/sh
set -eu
ARGS_FILE=%q
PROMPT_COPY=%q
: > "$PROMPT_COPY"
printf '%%s\0' "$@" > "$ARGS_FILE"
prev=
for a in "$@"; do
  if [ "$prev" = "-p" ]; then
    if [ -f "$a" ]; then
      cp "$a" "$PROMPT_COPY" || true
    else
      # shell-split the short instruction and copy any existing absolute path token
      old_ifs=$IFS
      IFS=' 	
'
      for tok in $a; do
        case "$tok" in
          /*)
            if [ -f "$tok" ]; then
              cp "$tok" "$PROMPT_COPY" || true
              break
            fi
            ;;
        esac
      done
      IFS=$old_ifs
    fi
  fi
  prev=$a
done
printf '%%s\n' %q
`, req.CommandCodeArgvPath, req.CommandCodePromptCopyPath, CommandCodeArgvRecorderJSON)
	if err := os.WriteFile(recorder, []byte(script), 0755); err != nil {
		t.Fatalf("write argv recorder: %v", err)
	}
	req.AgentRunner = "commandcode"
	req.AgentRunnerBinary = recorder
	// Recorder does not use the llm-mock hook env; clear so harness does not require it.
	req.CommandCodeHook = ""
}

// ParseNULSeparatedArgs splits recorder argv capture (trailing empty dropped).
func ParseNULSeparatedArgs(raw []byte) []string {
	parts := strings.Split(string(raw), "\x00")
	var args []string
	for _, p := range parts {
		if p != "" {
			args = append(args, p)
		}
	}
	return args
}

// ArgvValueAfter returns the argument immediately following flag, or "".
func ArgvValueAfter(args []string, flag string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag {
			return args[i+1]
		}
	}
	return ""
}
```
