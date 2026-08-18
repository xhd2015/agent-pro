# Scenario

**Feature**: `FollowUpOpts` idle-exit fields control `--exit-on-idle` / `--idle-timeout` emit

```
FollowUpOpts{ExitOnIdle, IdleTimeout}
  -> NormalizeIdle -> enabled / duration / err
  -> BuildFollowUpCommand -> own tokens before -- / prompt, or omit, or API error
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agentrunapi` exports
  `BuildFollowUpCommand`, `FollowUpOpts`, and `Opts` with **`ExitOnIdle bool`**
  and **`IdleTimeout time.Duration`**, plus **`DefaultIdleTimeout`** and
  **`NormalizeIdle`** (RED until implementer adds fields + helper + emission).
- Nested root: `d.DOCTEST_ROOT` is `tests/agentrunapi/follow-up-idle`.
- No real agent-run binary / iTerm / TTY.
- Parallel-safe: no `os.Setenv` / `t.Setenv` / `Chdir` / process stdio hijack.

## Steps

1. Root seeds default session / prompt / open profile.
2. Grouping sets `ExitOnIdle` true or false.
3. Leaf sets `IdleTimeout`.
4. `Run` calls `NormalizeIdle` then `BuildFollowUpCommand`; assert tokens or error.

## Context

- Default: `SessionID=sess-fu-idle`, `Prompt=follow-up idle`, `AgentRunner=grok-tty`, `Open=true`.
- Compact emit tokens: `--idle-timeout=10m`, `--idle-timeout=2m`, `--idle-timeout=2s`.
- Token checks: `strings.Fields` + trim quotes (never substring).

```go
import (
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	if req.SessionID == "" {
		req.SessionID = "sess-fu-idle"
	}
	if req.Prompt == "" {
		req.Prompt = "follow-up idle"
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
		t.Fatalf("unexpected emit error: %s", resp.ErrString)
	}
	if resp != nil && resp.NormalizeErr != "" {
		t.Fatalf("unexpected NormalizeIdle error: %s", resp.NormalizeErr)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in %q", want, got)
	}
}

func assertOmitsIdleFlags(t *testing.T, line string) {
	t.Helper()
	if hasExactToken(line, "--exit-on-idle") {
		t.Fatalf("must not emit --exit-on-idle when disabled; got %q", line)
	}
	if hasIdleTimeoutPrefix(line) {
		t.Fatalf("must not emit --idle-timeout when disabled; got %q", line)
	}
}

func assertEmitsIdle(t *testing.T, line, timeoutTok string) {
	t.Helper()
	prefix := tokensBeforeDashDash(line)
	if !sliceHasToken(prefix, "--exit-on-idle") {
		t.Fatalf("missing --exit-on-idle before -- in %q (prefix %q)", line, strings.Join(prefix, " "))
	}
	if !sliceHasToken(prefix, timeoutTok) {
		t.Fatalf("missing %q before -- in %q (prefix %q)", timeoutTok, line, strings.Join(prefix, " "))
	}
	if strings.Contains(line, "--new-terminal") {
		t.Fatalf("must not emit --new-terminal; got %q", line)
	}
}

func assertOpenProfile(t *testing.T, line, sessionID string) {
	t.Helper()
	assertContains(t, line, sessionID)
	if !hasExactToken(line, "--open") {
		t.Fatalf("open profile must include --open; got %q", line)
	}
}

func assertNormalized(t *testing.T, resp *Response, enabled bool, want time.Duration) {
	t.Helper()
	if resp.Enabled != enabled {
		t.Fatalf("NormalizeIdle enabled: got %v, want %v", resp.Enabled, enabled)
	}
	if resp.Normalized != want {
		t.Fatalf("NormalizeIdle duration: got %s, want %s", resp.Normalized, want)
	}
}
```
