## Expected

- Fixture contains legacy `Grok ›` (or `Grok` + `›`) prompt chrome.
- Writable: `ready=true`, `state=idle`.
- `banner_detected_legacy=true`.
- `open_ready=true`.
- Screen class may be `idle` (preferred) or another non-empty class; if Classify returns a
  class, it must not be `empty` or `modal`.

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
		text, readErr = os.ReadFile(filepath.Join(req.TestdataDir, fixtureLegacyAngleResponse))
		if readErr != nil {
			t.Fatal(readErr)
		}
	}
	s := string(text)
	if !strings.Contains(s, "Grok") || (!strings.Contains(s, "›") && !strings.Contains(s, "\u203a")) {
		t.Fatalf("fixture must contain legacy Grok › chrome, got %q", s)
	}

	assertWritable(t, "legacy angle", resp.Status, true, "idle", "")

	gotLegacy := agenttty.BannerDetected(text, "grok", grokLegacyBannerMarkers)
	gotOpen := agenttty.OpenReady(text)
	gotClass := agenttty.ClassifyGrokScreen(text)
	assertOpenReadyTriplet(t, "legacy angle", gotLegacy, true, gotOpen, true, gotClass, "")
	if gotClass == "empty" || gotClass == "modal" {
		t.Fatalf("legacy angle screen_class must not be empty/modal, got %q", gotClass)
	}
}
```
