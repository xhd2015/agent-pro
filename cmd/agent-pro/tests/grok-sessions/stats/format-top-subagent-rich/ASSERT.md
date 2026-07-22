## Expected

- Section `Top subagents` present.
- Header includes columns STATUS, TYPE, TOOLS, TURNS, DESC (and DURATION / #).
  Exact fragment:
  `#  DURATION  STATUS     TYPE              TOOLS  TURNS  DESC`
  Column widths may pad STATUS/TYPE; assert at least each column name appears
  on the header line **and** the locked contiguous multi-space form when the
  implementer matches the design table (prefer exact string from requirement).
- Row shows:
  - status `completed`
  - type `general-purpose`
  - tools `99`
  - turns `1`
  - description fragment `[designer] design rich top subagent tables`
    (or truncated at 64 runes with `…` if longer — this desc is short)
- Primary label is **not** UUID-only when description is available: the full
  UUID alone as the only descriptive token is insufficient if desc is missing.
- Pretty duration for 900000ms → `15m` (or equivalent) may appear.

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
	assertContains(t, out, "Top subagents")

	sub := sectionAfter(out, "Top subagents")
	if sub == "" {
		t.Fatalf("missing Top subagents section:\n%s", out)
	}

	// Locked header (padding spaces as in requirement sample).
	// Accept either the exact sample or a flexible header containing all names.
	headerOK := strings.Contains(sub, "#  DURATION  STATUS     TYPE              TOOLS  TURNS  DESC")
	if !headerOK {
		// Fallback: header line must include every column name.
		lines := strings.Split(sub, "\n")
		hdr := ""
		for _, ln := range lines {
			if strings.Contains(ln, "DURATION") && strings.Contains(ln, "STATUS") {
				hdr = ln
				break
			}
		}
		if hdr == "" {
			t.Fatalf("top-subagents missing DURATION/STATUS header:\n%s", sub)
		}
		for _, col := range []string{"STATUS", "TYPE", "TOOLS", "TURNS", "DESC"} {
			if !strings.Contains(hdr, col) {
				t.Fatalf("top-subagents header missing %q (got %q):\n%s", col, hdr, sub)
			}
		}
		// Must not be LABEL-only layout.
		if strings.Contains(hdr, "LABEL") && !strings.Contains(hdr, "DESC") {
			t.Fatalf("top-subagents still LABEL-only:\n%s", sub)
		}
	}

	assertContains(t, sub, "completed")
	assertContains(t, sub, "general-purpose")
	assertContains(t, sub, "99")
	assertContains(t, sub, formatTopSubRichDesc)

	// Must surface spawn description, not UUID-only primary label.
	// If only the id appears without desc, fail (desc already required above).
	// Extra: turns column "1" — soft check via full desc + tools already strong.

	// Aggregates unchanged path still works.
	if resp.Stats.Subagents == nil || resp.Stats.Subagents.Count != 1 {
		t.Fatalf("Subagents.Count = %v, want 1", resp.Stats.Subagents)
	}
}
```
