## Expected

- Writable: `ready=false`, `state=unknown`, `reason=no terminal output`.
- Open-lifecycle: `banner_detected_legacy=false`, `open_ready=false`, `screen_class=empty`.

## Exit Code

N/A (direct package call)

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertWritable(t, "boot empty", resp.Status, false, "unknown", "no terminal output")

	text := resp.Scrollback
	gotLegacy := agenttty.BannerDetected(text, "grok", grokLegacyBannerMarkers)
	gotOpen := agenttty.OpenReady(text)
	gotClass := agenttty.ClassifyGrokScreen(text)
	assertOpenReadyTriplet(t, "boot empty", gotLegacy, false, gotOpen, false, gotClass, "empty")
}
```
