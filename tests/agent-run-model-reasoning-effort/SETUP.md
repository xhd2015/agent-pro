# Scenario

**Feature**: opt-in model + reasoning-effort pass-through (CLI, FollowUp, library Opts); empty invents nothing

```
# pure FollowUp (L2)
FollowUpOpts{Model, ModelReasoningEffort} -> BuildFollowUpCommand
  -> shell line with/without --model / --model-reasoning-effort

# library keep (L2)
Opts{Model, ModelReasoningEffort} -> AutoSendOrResume(RunSession hook)
  -> captured fields; empty stays empty

# CLI surface (L2; pure file read — no process stdio swap)
read pkgs/agentruncli/run_cmd.go -> runHelp documents --model-reasoning-effort
scan pkgs/agentruncli -> flag literal registered

# ForceNew wire
scan pkgs/agentruncli BuildFollowUpCommand site
  -> FollowUpOpts.Model + ModelReasoningEffort assigned
```

## Preconditions

- Package roots under agent-pro module:
  - `github.com/xhd2015/agent-pro/pkgs/agentrunapi`
  - `github.com/xhd2015/agent-pro/pkgs/agentruncli`
- Classic TDD: `FollowUpOpts.Model` / `ModelReasoningEffort` may be missing
  (compile RED) until implementer adds them; emission RED until flags appear.
- Network-free; no real agent-run binary spawn, iTerm, codex TTY, or PATH LookPath
  for unit leaves.
- Parallel-safe: no `t.Setenv` / `os.Setenv` / `t.Chdir` / process-global env
  mutation; **no process stdio reassignment**. Library leaves use
  `t.TempDir()` store homes only; inject `io.Discard` on Opts writers.
- CLI help is a pure read of `pkgs/agentruncli/run_cmd.go` (holds `runHelp`);
  no `Handle` / process stdio capture.
- `d.DOCTEST_ROOT` is `tests/agent-run-model-reasoning-effort`; module root is
  found by walking for `module github.com/xhd2015/agent-pro`.

## Steps

1. Root `Setup` validates Request; defaults Mode when empty only if leaf sets it.
2. Grouping `Setup` sets `req.Mode` and surface defaults.
3. Leaf fills Model / Effort / SourceWireTarget.
4. `Run` builds follow-up, captures library opts, reads help source, or scans sources.
5. Leaf `Assert` checks outcomes.

## Context

- Fixture model: `o3` (`fixtureModel`).
- Fixture effort (both-set): `high` (`fixtureEffort`).
- Fixture effort (effort-only / library): `max` (`fixtureEffortMax`).
- Forbidden invent model when empty: `gpt-5.6-luna`.
- Default open-profile for follow-up leaves: `Open=true`, runner `grok-tty`.

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
	req.Mode = strings.TrimSpace(req.Mode)
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Prompt = strings.TrimSpace(req.Prompt)
	req.AgentRunner = strings.TrimSpace(req.AgentRunner)
	req.Model = strings.TrimSpace(req.Model)
	req.ModelReasoningEffort = strings.TrimSpace(req.ModelReasoningEffort)
	req.SourceWireTarget = strings.TrimSpace(req.SourceWireTarget)
	req.Home = strings.TrimSpace(req.Home)

	// Mode-specific soft defaults (leaves override).
	switch req.Mode {
	case "follow_up":
		if req.SessionID == "" {
			req.SessionID = "sess-fu-model-effort"
		}
		if req.Prompt == "" {
			req.Prompt = "follow-up model effort"
		}
		if req.AgentRunner == "" {
			req.AgentRunner = "grok-tty"
		}
		if !req.Open && !req.Detach {
			req.Open = true
		}
	case "library_opts":
		if req.Home == "" {
			req.Home = filepath.Join(t.TempDir(), ".agent-run")
		}
		if req.SessionID == "" {
			req.SessionID = "sess-lib-opts"
		}
		if req.Prompt == "" {
			req.Prompt = "library opts prompt"
		}
	case "cli_help":
		// Pure source read of run_cmd.go in Run; no Args needed.
	case "source_wire":
		// SourceWireTarget set by grouping/leaf.
	case "":
		// Leaf/grouping will set Mode.
	default:
		// Unknown modes fail in Run.
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

// shellTokens approximates shell field split for pure follow-up lines
// (quoted tokens from ShellQuote rarely contain spaces for our fixtures).
func shellTokens(line string) []string {
	return strings.Fields(line)
}

// hasModelFlag is prefix-safe: --model-reasoning-effort is NOT a model flag.
func hasModelFlag(line string) bool {
	toks := shellTokens(line)
	for i := 0; i < len(toks); i++ {
		raw := strings.Trim(toks[i], `"'`)
		if raw == "--model" {
			return true
		}
		if strings.HasPrefix(raw, "--model=") {
			return true
		}
	}
	return false
}

func modelFlagValue(line string) (string, bool) {
	toks := shellTokens(line)
	for i := 0; i < len(toks); i++ {
		raw := strings.Trim(toks[i], `"'`)
		if raw == "--model" {
			if i+1 < len(toks) {
				return strings.Trim(toks[i+1], `"'`), true
			}
			return "", true
		}
		if strings.HasPrefix(raw, "--model=") {
			return strings.TrimPrefix(raw, "--model="), true
		}
	}
	return "", false
}

func hasReasoningEffortFlag(line string) bool {
	toks := shellTokens(line)
	for i := 0; i < len(toks); i++ {
		raw := strings.Trim(toks[i], `"'`)
		if raw == "--model-reasoning-effort" {
			return true
		}
		if strings.HasPrefix(raw, "--model-reasoning-effort=") {
			return true
		}
	}
	return false
}

func reasoningEffortFlagValue(line string) (string, bool) {
	toks := shellTokens(line)
	for i := 0; i < len(toks); i++ {
		raw := strings.Trim(toks[i], `"'`)
		if raw == "--model-reasoning-effort" {
			if i+1 < len(toks) {
				return strings.Trim(toks[i+1], `"'`), true
			}
			return "", true
		}
		if strings.HasPrefix(raw, "--model-reasoning-effort=") {
			return strings.TrimPrefix(raw, "--model-reasoning-effort="), true
		}
	}
	return "", false
}

func assertHasModelValue(t *testing.T, line, want string) {
	t.Helper()
	got, ok := modelFlagValue(line)
	if !ok {
		t.Fatalf("missing --model flag; line=%q", line)
	}
	if got != want {
		t.Fatalf("--model value=%q want %q; line=%q", got, want, line)
	}
}

func assertHasEffortValue(t *testing.T, line, want string) {
	t.Helper()
	got, ok := reasoningEffortFlagValue(line)
	if !ok {
		t.Fatalf("missing --model-reasoning-effort flag; line=%q", line)
	}
	if got != want {
		t.Fatalf("--model-reasoning-effort value=%q want %q; line=%q", got, want, line)
	}
}

func assertNoModelFlag(t *testing.T, line string) {
	t.Helper()
	if hasModelFlag(line) {
		t.Fatalf("must not emit --model; line=%q", line)
	}
}

func assertNoEffortFlag(t *testing.T, line string) {
	t.Helper()
	if hasReasoningEffortFlag(line) {
		t.Fatalf("must not emit --model-reasoning-effort; line=%q", line)
	}
}

```
