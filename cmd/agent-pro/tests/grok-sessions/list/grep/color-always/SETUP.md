# Scenario

**Feature**: Color=always wraps hit-line filename, line, part, and match span with ANSI

```
# single title hit for GREP_COLOR_TOKEN; force color on even though not a TTY
ListWithGrep + FormatListTableWithHits(colorMode=always)

# hit line contains ANSI SGR sequences (magenta file, green line/part, bold red match)
```

## Preconditions

- `req.Color = "always"` overrides TTY/NO_COLOR auto detection.
- Pattern appears once in title for a simple colored snippet.
- Styling contract (grep-like):
  - filename → magenta (`\x1b[35m`)
  - line number → green (`\x1b[32m`)
  - part → green (`\x1b[32m`)
  - matched substring in snippet → bold red (`\x1b[1m` + `\x1b[31m` or combined)

## Steps

1. Set `req.Grep = "GREP_COLOR_TOKEN"`, `req.Limit = 10`, `req.Color = "always"`.
2. Write one session whose title contains the token.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Grep = "GREP_COLOR_TOKEN"
	req.Limit = 10
	req.Color = "always"

	writeGrokSession(t, req.GrokHome,
		"01900018-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-color-always",
		"Enable GREP_COLOR_TOKEN highlighting")
	return nil
}
```
