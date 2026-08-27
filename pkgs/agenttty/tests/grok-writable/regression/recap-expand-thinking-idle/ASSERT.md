## Expected

- Fixture contains post-turn Recap chrome: `Worked for`, `Recap`, `Build anything`,
  `expand thinking`, `❯`, `Grok 4.5`.
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
		text, readErr = os.ReadFile(filepath.Join(req.TestdataDir, fixtureRecapExpandThinkingIdle))
		if readErr != nil {
			t.Fatal(readErr)
		}
	}
	s := string(text)
	for _, want := range []string{"Worked for", "Recap", "Build anything", "expand thinking", "❯", "Grok 4.5"} {
		if !strings.Contains(s, want) {
			t.Fatalf("fixture must contain %q", want)
		}
	}

	assertWritable(t, "recap expand-thinking idle", resp.Status, true, "idle", "")

	gotLegacy := agenttty.BannerDetected(text, "grok", grokLegacyBannerMarkers)
	gotOpen := agenttty.OpenReady(text)
	gotClass := agenttty.ClassifyGrokScreen(text)
	assertOpenReadyTriplet(t, "recap expand-thinking idle", gotLegacy, false, gotOpen, true, gotClass, "idle")
}
```
