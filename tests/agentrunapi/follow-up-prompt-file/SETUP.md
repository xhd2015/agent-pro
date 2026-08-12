# Scenario

**Feature**: `BuildFollowUpCommand` auto-spills long prompts to `--prompt-file`

```
FollowUpOpts{Prompt, PromptFile, PromptSpillDir, Open, …}
  -> BuildFollowUpCommand
  -> shell line with inline `--` prompt  OR  --prompt-file=… + spill file
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentrunapi` exports
  `BuildFollowUpCommand` and `FollowUpOpts` with **`PromptFile`** and
  **`PromptSpillDir`** (RED until implementer adds fields + emission).
- Nested root: `d.DOCTEST_ROOT` is `tests/agentrunapi/follow-up-prompt-file`.
- No real agent-run binary / iTerm / PATH LookPath for production spill defaults.
- Parallel-safe: leaves inject `PromptSpillDir` under `d.DOCTEST_CASE` only;
  **no** `t.Setenv` / `t.Chdir` / `os.Setenv` / process stdio reassignment.
- Locked: flag `--prompt-file`; threshold **600 runes** after `TrimSpace(Prompt)`.

## Steps

1. Root seeds default session / open profile when unset.
2. Leaf sets Prompt / PromptFile / PromptSpillDir fixtures under `d.DOCTEST_CASE`.
3. `Run` calls `BuildFollowUpCommand`; records follow-up line + spill dir entries.
4. Leaf `Assert` checks tokens, spill content, and abs path preferences.

## Context

- Default: `SessionID=sess-fu-prompt-file`, `AgentRunner=grok-tty`, `Open=true`.
- Short fixture prompt: `hello` (`fixtureShortPrompt`).
- Threshold constant: `promptFileSpillMinRunes = 600`.
- Flag token: `--prompt-file` (`flagPromptFile`).

```go
import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if req == nil {
		return fmt.Errorf("nil Request")
	}
	_ = d
	if req.SessionID == "" {
		req.SessionID = "sess-fu-prompt-file"
	}
	if req.AgentRunner == "" {
		req.AgentRunner = "grok-tty"
	}
	if !req.Open && !req.Detach {
		req.Open = true
	}
	return nil
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertNoAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp != nil && resp.ErrString != "" {
		t.Fatalf("unexpected API error: %s", resp.ErrString)
	}
}

func assertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %#v, want %#v", field, got, want)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in %q", want, got)
	}
}

func assertNotContains(t *testing.T, got, forbidden string) {
	t.Helper()
	if strings.Contains(got, forbidden) {
		t.Fatalf("unexpected %q in %q", forbidden, got)
	}
}

func assertNoPromptFileFlag(t *testing.T, line string) {
	t.Helper()
	if hasPromptFileFlag(line) {
		t.Fatalf("FollowUp must not contain %s; got %q", flagPromptFile, line)
	}
}

func assertHasPromptFileFlag(t *testing.T, line string) {
	t.Helper()
	if !hasPromptFileFlag(line) {
		t.Fatalf("FollowUp must contain %s; got %q", flagPromptFile, line)
	}
}

func assertSpillDirEmpty(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.SpillDirEntries) != 0 {
		t.Fatalf("PromptSpillDir must have no auto-spill files; got %v", resp.SpillDirEntries)
	}
}

func assertAbsPath(t *testing.T, path string) {
	t.Helper()
	if path == "" {
		t.Fatal("prompt-file path is empty")
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("prompt-file path must be absolute; got %q", path)
	}
}

func assertOpenProfile(t *testing.T, line, sessionID string) {
	t.Helper()
	assertContains(t, line, sessionID)
	assertContains(t, line, "--open")
	assertContains(t, line, "--auto-send-or-resume")
	assertNotContains(t, line, "--new-terminal")
}
```
