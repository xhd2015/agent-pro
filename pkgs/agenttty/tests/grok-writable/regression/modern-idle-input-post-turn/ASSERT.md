## Expected

- Fixture contains modern idle chrome: `Turn completed`, `❯`, `Grok 4.5`, `Shift+Tab:mode`.
- Writable: `ready=true`, `state=idle`.
- `banner_detected_legacy=false`.
- `open_ready=true`, `screen_class=idle`.

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
		text, readErr = os.ReadFile(filepath.Join(req.TestdataDir, fixtureModernIdle))
		if readErr != nil {
			t.Fatal(readErr)
		}
	}
	s := string(text)
	if !strings.Contains(s, "Turn completed") {
		t.Fatalf("fixture must contain Turn completed marker")
	}
	if !strings.Contains(s, "❯") {
		t.Fatalf("fixture must contain idle ❯ prompt")
	}
	if !strings.Contains(s, "Grok 4.5") {
		t.Fatalf("fixture must contain Grok 4.5 model chrome")
	}
	if !strings.Contains(s, "Shift+Tab:mode") {
		t.Fatalf("fixture must contain Shift+Tab:mode footer")
	}

	assertWritable(t, "modern idle", resp.Status, true, "idle", "")

	gotLegacy := agenttty.BannerDetected(text, "grok", grokLegacyBannerMarkers)
	gotOpen := agenttty.OpenReady(text)
	gotClass := agenttty.ClassifyGrokScreen(text)
	assertOpenReadyTriplet(t, "modern idle", gotLegacy, false, gotOpen, true, gotClass, "idle")
}
```
