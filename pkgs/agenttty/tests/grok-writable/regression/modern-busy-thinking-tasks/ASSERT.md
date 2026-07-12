## Expected

- Fixture contains modern busy chrome: `Thinking…` or `Thinking`, Tasks chrome, `❯`, `Grok 4.5`.
- Writable: `ready=false`, `state=busy`, `reason=agent still responding`.
- `banner_detected_legacy=false`.
- `open_ready=true`, `screen_class=busy`.

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	text := resp.Scrollback
	if len(text) == 0 {
		var readErr error
		text, readErr = os.ReadFile(filepath.Join(req.TestdataDir, fixtureModernBusy))
		if readErr != nil {
			t.Fatal(readErr)
		}
	}
	s := string(text)
	if !strings.Contains(s, "Thinking") {
		t.Fatalf("fixture must contain Thinking busy marker")
	}
	if !strings.Contains(s, "❯") {
		t.Fatalf("fixture must contain ❯ chrome while busy")
	}
	if !strings.Contains(s, "Grok 4.5") {
		t.Fatalf("fixture must contain Grok 4.5 model chrome")
	}

	assertWritable(t, "modern busy", resp.Status, false, "busy", "agent still responding")

	gotLegacy := agenttty.BannerDetected(text, "grok", grokLegacyBannerMarkers)
	gotOpen := agenttty.OpenReady(text)
	gotClass := agenttty.ClassifyGrokScreen(text)
	assertOpenReadyTriplet(t, "modern busy", gotLegacy, false, gotOpen, true, gotClass, "busy")
}
```
