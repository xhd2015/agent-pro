# Scenario

**Feature**: per-message body rune cap (`MaxPerMessageRunes` / default 1000)

```
long Message.Text + Options{MaxPerMessageRunes}
  -> body shortened to max runes ending with "…" (U+2026)
```

## Preconditions

- Cap applies to **body** (`Text`) only, not id/sender prefixes.
- `MaxPerMessageRunes <= 0` uses `DefaultMaxPerMessageRunes` (1000).
- Truncated body rune count equals the effective max (prefix + one-rune `…`).
- Marker is `…` only (not ASCII `...`, not `[truncated]`).

## Steps

1. Branch Setup clears selection options so only body-cap varies.
2. Leaf builds a body of known rune length and optional MaxPerMessageRunes.
3. Assert body length, marker, and optionally `BodiesTruncated`.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/msgfmt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// body-cap branch: isolate per-message truncation (no count/budget drops).
	req.Opts = msgfmt.Options{
		MaxPerMessageRunes: 0, // 0 → DefaultMaxPerMessageRunes; leaves may override
		MaxMessages:        0,
		TotalBudgetRunes:   0,
	}
	return nil
}
```
