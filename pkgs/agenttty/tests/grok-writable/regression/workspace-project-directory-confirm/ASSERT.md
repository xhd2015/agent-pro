## Expected

- Fixture contains distinctive workspace-confirm markers:
  `Run Grok Build in a project directory?`, `Type your answer here`, `Enter:submit`, `(○)`.
- `ready=false` — must **not** be sendable.
- `state` must **not** be `idle`.
- Current classification may be `unknown` (no classic prompt marker) or a clearer
  modal/loading/busy state after implementer hardens detection; either is OK as long
  as ready stays false and state is not idle.
- `reason` is non-empty.
- **Legacy banner characterization:** `banner_detected_legacy=true` (false-positive on
  `"grok build"` substring in the modal title).
- **Open readiness:** `open_ready=false` — modal must override legacy match so open wait
  does not treat the picker as ready to inject.
- `screen_class=modal`.

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	text := resp.Scrollback
	if len(text) == 0 {
		var readErr error
		text, readErr = os.ReadFile(filepath.Join(req.TestdataDir, fixtureWorkspaceProjectDirectoryConfirm))
		if readErr != nil {
			t.Fatal(readErr)
		}
	}
	s := string(text)
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "run grok build in a project directory") {
		t.Fatalf("fixture must contain project-directory confirm title")
	}
	if !strings.Contains(lower, "type your answer here") {
		t.Fatalf("fixture must contain Type your answer here")
	}
	if !strings.Contains(s, "Enter:submit") {
		t.Fatalf("fixture must contain Enter:submit (not idle Enter:send)")
	}
	if !strings.Contains(s, "(○)") && !strings.Contains(s, "(o)") {
		t.Fatalf("fixture must contain radio option chrome (○)")
	}

	if resp.Status.Ready {
		t.Fatalf("workspace project-directory confirm must not be ready=true (got state=%q reason=%q) — prompt would be injected into picker",
			resp.Status.State, resp.Status.Reason)
	}
	if resp.Status.State == "idle" {
		t.Fatalf("workspace project-directory confirm must not be state=idle (got ready=%v reason=%q)",
			resp.Status.Ready, resp.Status.Reason)
	}
	if strings.TrimSpace(resp.Status.Reason) == "" {
		t.Fatalf("expected non-empty reason for non-ready workspace confirm (state=%q)", resp.Status.State)
	}

	// Open-lifecycle characterization (RED until OpenReady rejects modal despite legacy FP).
	gotLegacy := agenttty.BannerDetected(text, "grok", grokLegacyBannerMarkers)
	gotOpen := agenttty.OpenReady(text)
	gotClass := agenttty.ClassifyGrokScreen(text)
	assertOpenReadyTriplet(t, "workspace modal", gotLegacy, true, gotOpen, false, gotClass, "modal")
}
```
