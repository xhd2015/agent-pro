## Expected

- Exit code 0 (open lifecycle completes with instant-attach hook).
- Stderr has a `grok-tty: <id>` session-id line (post-attach).
- PTY snapshot (or combined CLI output) must **not** contain `SUBMITTED:draft`
  — proves inject used `suffixCR=false` (no auto-submit Enter).
- Draft text may appear unsubmitted / in the input buffer; that is allowed.
- Must not fail solely as “unrecognized flag” for `--no-submit`.

## Side Effects

- Keep-alive TTY session may remain after open (registry/listen_addr); snapshot
  is taken while the session is still reachable when possible.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}

	combinedCLI := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if strings.Contains(combinedCLI, "unrecognized flag") || strings.Contains(combinedCLI, "unknown flag") {
		t.Fatalf("want implemented --no-submit path, got unrecognized/unknown flag:\n%s", resp.Stderr)
	}

	assertSuccess(t, resp)

	id, ok := parsePrefixedSessionID(resp.Stderr, "grok-tty")
	if !ok {
		t.Fatalf("missing post-attach grok-tty session id on stderr:\n%s", resp.Stderr)
	}
	if n := countPrefixedSessionIDLines(resp.Stderr, "grok-tty"); n != 1 {
		t.Fatalf("want exactly 1 grok-tty session id line, got %d (id=%q)\nstderr:\n%s", n, id, resp.Stderr)
	}

	snap := resp.Snapshot
	// Retry snapshot if Run did not capture it (registry race / Mode not wired).
	if strings.TrimSpace(snap) == "" {
		entry := resp.RegistryEntry
		if entry == nil {
			var rerr error
			entry, rerr = readRegistryEntryOptional(req.Home, "grok-tty", id)
			if rerr != nil {
				t.Fatalf("read registry for snapshot: %v (session=%s)\nstderr:\n%s", rerr, id, resp.Stderr)
			}
		}
		if entry.ListenAddr == "" {
			t.Fatalf("empty listen_addr for session %s; cannot prove inject-without-CR", id)
		}
		// Settle so a buggy CR inject would emit SUBMITTED before we look.
		time.Sleep(400 * time.Millisecond)
		text, serr := ttywatch.SnapshotText(entry.ListenAddr, id)
		if serr != nil {
			t.Fatalf("snapshot PTY for session %s at %s: %v", id, entry.ListenAddr, serr)
		}
		snap = text
	}

	submittedMarker := "SUBMITTED:" + req.Prompt
	if req.Prompt == "" {
		submittedMarker = "SUBMITTED:draft"
	}
	if strings.Contains(snap, submittedMarker) {
		t.Fatalf("--no-submit must not auto-submit; found %q in PTY snapshot:\n%s\nstderr:\n%s",
			submittedMarker, snap, resp.Stderr)
	}
	// Also reject SUBMITTED: with any draft-like line if prompt mismatched.
	if strings.Contains(snap, "SUBMITTED:") {
		t.Fatalf("--no-submit must not produce SUBMITTED: marker; snapshot:\n%s\nstderr:\n%s",
			snap, resp.Stderr)
	}

	// Soft check: banner should have appeared (session really started).
	if !strings.Contains(snap, grokTTYBannerMarker) && !strings.Contains(snap, "Grok") {
		// Do not hard-fail solely on banner absence if scrollback rendering strips it;
		// absence of SUBMITTED is the hard contract. Log-ish fail only when empty.
		if strings.TrimSpace(snap) == "" {
			t.Fatalf("empty PTY snapshot; cannot prove inject-without-CR (session=%s)", id)
		}
	}
}
```
