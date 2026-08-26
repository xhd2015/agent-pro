# Scenario

**Feature**: Color=never prints plain hit lines without ANSI escapes

```
# same title hit as color-always; force color off
ListWithGrep + FormatListTableWithHits(colorMode=never)

# hit line is plain "  summary.json:1:title: ..." with no CSI sequences
```

## Preconditions

- `req.Color = "never"` even if environment would otherwise enable color.
- Pattern appears in title for one clear hit line.

## Steps

1. Set `req.Grep = []string{"GREP_COLOR_TOKEN"}`, `req.Limit = 10`, `req.Color = "never"`.
2. Write one session whose title contains the token (same content shape as color-always).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Grep = []string{"GREP_COLOR_TOKEN"}
	req.Limit = 10
	req.Color = "never"

	writeGrokSession(t, req.GrokHome,
		"01900019-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-color-never",
		"Enable GREP_COLOR_TOKEN highlighting")
	return nil
}
```
