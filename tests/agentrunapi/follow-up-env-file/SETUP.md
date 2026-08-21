# Scenario

**Feature**: `BuildFollowUpCommand` auto-spills long session env to `--env-file`

```
FollowUpOpts{Env, EnvFile, EnvSpillDir, Open, …}
  -> BuildFollowUpCommand
  -> shell line with inline -e  OR  --env-file=… + spill file
```

## Preconditions

- `FollowUpOpts` has `EnvFile` / `EnvSpillDir`; threshold `EnvFileSpillMinRunes=64`.
- Parallel-safe: inject `EnvSpillDir` under `d.DOCTEST_CASE` only.

## Context

- Default: `SessionID=sess-fu-env-file`, `AgentRunner=grok-tty`, `Open=true`, `Prompt=hi`.

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
		req.SessionID = "sess-fu-env-file"
	}
	if req.AgentRunner == "" {
		req.AgentRunner = "grok-tty"
	}
	if req.Prompt == "" {
		req.Prompt = "hi"
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

func assertNoEnvFileFlag(t *testing.T, line string) {
	t.Helper()
	if hasEnvFileFlag(line) {
		t.Fatalf("FollowUp must not contain %s; got %q", flagEnvFile, line)
	}
}

func assertHasEnvFileFlag(t *testing.T, line string) {
	t.Helper()
	if !hasEnvFileFlag(line) {
		t.Fatalf("FollowUp must contain %s; got %q", flagEnvFile, line)
	}
}

func assertNoInlineE(t *testing.T, line string) {
	t.Helper()
	if hasInlineEnvFlag(line) {
		t.Fatalf("FollowUp must not contain inline -e/--env; got %q", line)
	}
}

func assertSpillDirEmpty(t *testing.T, resp *Response) {
	t.Helper()
	if len(resp.SpillDirEntries) != 0 {
		t.Fatalf("EnvSpillDir must have no auto-spill files; got %v", resp.SpillDirEntries)
	}
}

func assertAbsPath(t *testing.T, path string) {
	t.Helper()
	if path == "" {
		t.Fatal("env-file path is empty")
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("env-file path must be absolute; got %q", path)
	}
}

func assertOpenProfile(t *testing.T, line, sessionID string) {
	t.Helper()
	assertContains(t, line, sessionID)
	assertContains(t, line, "--open")
	assertContains(t, line, "--auto-send-or-resume")
	assertNotContains(t, line, "--new-terminal")
}

func assertSpillUnderDir(t *testing.T, path, spillDir string) {
	t.Helper()
	assertAbsPath(t, path)
	rel, relErr := filepath.Rel(spillDir, path)
	if relErr != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("spill path %q must be under EnvSpillDir %q", path, spillDir)
	}
}
```
