## Expected

- **Paris** appears in first-turn snapshot and/or `events.jsonl` (mock assistant).
- After `send /exit`, status reports `runner.exited=true` (human and/or JSON).
- `resume --open "hello"` succeeds (exit 0 preferred) and does **not** print
  `already in use`.
- Post-resume snapshot and/or events show proper followup text:
  `HELLO_RESUME_MARKER` and/or `hello` / interactive UI markers — not empty /
  not only a dead "Terminal exited" footer.

## Side Effects

- Session under `AGENT_RUN_HOME/sessions/grok-tty/<session-id>/`.
- Grok session tree under leaf `.grok/sessions/.../updates.jsonl`.
- Registry serve PIDs cleaned on test teardown.

## Exit Code

- Open: 0 preferred (bind may still surface session lines on soft failure paths).
- Resume: 0 preferred after reclaim; non-zero with `already in use` is a hard fail.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.Err != nil {
		t.Fatalf("flow error: %v\nopen stderr:\n%s\nresume stderr:\n%s",
			resp.Err, resp.Open.Stderr, resp.Resume.Stderr)
	}

	// 1) Paris visible.
	if !resp.HasParis {
		t.Fatalf("Paris %q not found in snapshot/events after open\nparis_snapshot:\n%s\nevents:\n%s\nopen stderr:\n%s",
			req.WantParis, resp.ParisSnapshot, resp.EventsBlob, resp.Open.Stderr)
	}

	// 2) Exited after /exit.
	if !resp.ExitedTrue {
		t.Fatalf("expected runner.exited=true after /exit\nstatus:\n%s\njson:\n%s\nsend_exit stderr:\n%s\nparis_snapshot:\n%s",
			resp.StatusAfterExit.Stdout, resp.StatusJSONAfterExit, resp.SendExit.Stderr, resp.ParisSnapshot)
	}

	// 3) Resume must not report already-in-use.
	if resp.AlreadyInUse {
		t.Fatalf("resume --open reported already in use\nresume stderr:\n%s\nstdout:\n%s",
			resp.Resume.Stderr, resp.Resume.Stdout)
	}
	resumeCombined := strings.ToLower(resp.Resume.Stderr + "\n" + resp.Resume.Stdout)
	if strings.Contains(resumeCombined, "already in use") {
		t.Fatalf("resume --open already in use:\n%s", resumeCombined)
	}

	// 4) Resume should complete the open path (exit 0 preferred).
	if resp.Resume.ExitCode != 0 {
		// Gate denials are fixture bugs for this leaf.
		if strings.Contains(resumeCombined, "cannot resume") ||
			strings.Contains(resumeCombined, "not exited") ||
			strings.Contains(resumeCombined, "still active") {
			t.Fatalf("resume blocked by gate (expected exited+bound after mock /exit): exit=%d\n%s\nstatus:\n%s",
				resp.Resume.ExitCode, resumeCombined, resp.StatusAfterExit.Stdout)
		}
		t.Fatalf("resume --open exit=%d (want 0)\nstderr:\n%s\nstdout:\n%s",
			resp.Resume.ExitCode, resp.Resume.Stderr, resp.Resume.Stdout)
	}

	// 5) Post-resume snapshot/events are proper text.
	snap := strings.TrimSpace(resp.ResumeSnapshot)
	events := resp.EventsBlob
	combined := snap + "\n" + events
	if snap == "" && !strings.Contains(events, req.HelloMarker) && !strings.Contains(strings.ToLower(events), "hello") {
		t.Fatalf("resume snapshot empty and events lack followup; events:\n%s", events)
	}
	lower := strings.ToLower(combined)
	onlyExited := strings.Contains(lower, "terminal exited") &&
		!strings.Contains(combined, req.FollowupPrompt) &&
		!strings.Contains(lower, "hello") &&
		!strings.Contains(combined, req.HelloMarker) &&
		!strings.Contains(combined, "❯") &&
		!strings.Contains(lower, "grok")
	if onlyExited {
		t.Fatalf("resume snapshot looks like dead/exited terminal only:\n%s", snap)
	}
	hasHello := resp.HasHello ||
		strings.Contains(combined, req.HelloMarker) ||
		strings.Contains(combined, req.FollowupPrompt) ||
		strings.Contains(lower, "hello")
	hasUI := strings.Contains(snap, "❯") ||
		strings.Contains(lower, "enter:send") ||
		strings.Contains(lower, "resume") ||
		strings.Contains(snap, "Grok") ||
		strings.Contains(snap, "GROK_TTY_BANNER") ||
		strings.Contains(snap, "Grok ›")
	if !hasHello && !hasUI {
		t.Fatalf("resume snapshot/events missing followup %q / marker %q and UI markers\nsnapshot:\n%s\nevents:\n%s",
			req.FollowupPrompt, req.HelloMarker, snap, events)
	}
}
```
