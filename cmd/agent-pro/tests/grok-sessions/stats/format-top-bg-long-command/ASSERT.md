## Expected

- Output includes section `Top background tasks`.
- Table header includes `EXIT` and `COMMAND` (not a LABEL-only layout).
- Exact header fragment present: `#  DURATION  EXIT  COMMAND` (column order
  as locked; surrounding whitespace may vary only if implementer keeps the
  two-space indent — assert the contiguous header token string).
- The **120-rune** command appears **in full** in the top-bg section (no
  mid-string `…` truncation from the old 60-display / 80-store caps).
- Marker `MID120_OK` is present.
- The full **220-rune** command does **not** appear; marker `LONG220_TAIL`
  does **not** appear.
- A Unicode ellipsis `…` appears in the top-bg section (truncation of the
  220-rune command at display max **200 runes**).
- Background tasks aggregate Count still 2 (optional smoke).

## Errors

- None.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	out := resp.Output
	assertContains(t, out, "Top background tasks")

	// Locked header: #  DURATION  EXIT  COMMAND
	assertContains(t, out, "#  DURATION  EXIT  COMMAND")
	// Must not remain on LABEL-only top-bg layout.
	bg := sectionAfter(out, "Top background tasks")
	if bg == "" {
		t.Fatalf("missing Top background tasks section:\n%s", out)
	}
	if strings.Contains(bg, "LABEL") && !strings.Contains(bg, "COMMAND") {
		t.Fatalf("top-bg still LABEL-only (want COMMAND):\n%s", bg)
	}

	// 120-rune command fully visible (defeats old 60-display / 80-store).
	if !strings.Contains(bg, midBgCommand) {
		t.Fatalf("120-rune command not fully present in top-bg (truncated too early?):\nsection:\n%s\nwant full len=%d", bg, len([]rune(midBgCommand)))
	}
	assertContains(t, bg, "MID120_OK")

	// 220-rune command must be display-truncated at 200 runes + …
	if strings.Contains(bg, longBgCommand) {
		t.Fatalf("full 220-rune command must not appear (want display truncate 200+…):\n%s", bg)
	}
	if strings.Contains(bg, "LONG220_TAIL") {
		t.Fatalf("LONG220_TAIL tail of 220-rune command must not appear after truncate:\n%s", bg)
	}
	if !strings.Contains(bg, "…") {
		t.Fatalf("top-bg missing Unicode ellipsis … for >200-rune command:\n%s", bg)
	}
	// Prefix of long command should still be visible (at least first 100 runes).
	prefix := string([]rune(longBgCommand)[:100])
	if !strings.Contains(bg, prefix) {
		t.Fatalf("top-bg missing long-command prefix %q:\n%s", prefix, bg)
	}

	if resp.Stats.BackgroundTasks == nil || resp.Stats.BackgroundTasks.Count != 2 {
		t.Fatalf("BackgroundTasks.Count = %v, want 2", resp.Stats.BackgroundTasks)
	}
}
```
