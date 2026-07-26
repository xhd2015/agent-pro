## Expected

- Fixture contains modern starting chrome markers: `Starting session`, `❯`, `Grok 4.5`, `Shift+Tab:mode`.
- Writable (option A): `ready=true`, `state=idle` (do **not** force `loading`).
- `banner_detected_legacy=false` (no `GROK_TTY_BANNER` / `Grok ›`).
- `open_ready=true`, `screen_class=starting`.

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
		text, readErr = os.ReadFile(filepath.Join(req.TestdataDir, fixtureModernStarting))
		if readErr != nil {
			t.Fatal(readErr)
		}
	}
	s := string(text)
	if !strings.Contains(s, "Starting session") {
		t.Fatalf("fixture must contain Starting session chrome")
	}
	if !strings.Contains(s, "❯") {
		t.Fatalf("fixture must contain ❯ prompt chrome")
	}
	if !strings.Contains(s, "Grok 4.5") {
		t.Fatalf("fixture must contain Grok 4.5 model chrome")
	}
	if !strings.Contains(s, "Shift+Tab:mode") {
		t.Fatalf("fixture must contain Shift+Tab:mode footer")
	}

	assertWritable(t, "modern starting", resp.Status, true, "idle", "")

	// Exported open-lifecycle APIs (RED until implementer exports them).
	gotLegacy := agenttty.BannerDetected(text, "grok", grokLegacyBannerMarkers)
	gotOpen := agenttty.OpenReady(text)
	gotClass := agenttty.ClassifyGrokScreen(text)
	assertOpenReadyTriplet(t, "modern starting", gotLegacy, false, gotOpen, true, gotClass, "starting")
}
```
