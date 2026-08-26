# Scenario

**Feature**: short matching fields use the full collapsed text with no snippet ellipsis

```
# short generated_title contains GREP_SNIP_SHORT_TOKEN (well under 1024 runes)
writeGrokSession(title with token) -> ListWithGrep

# hit snippet is the full title text; no leading/trailing "..." from windowing
[]SessionMatch -> FormatListTableWithHits(color=never)
```

## Preconditions

- Matching field is short (title well under 1024 runes after collapse).
- Full collapsed field must appear as the snippet with **no** windowing ellipsis.
- Existing short-fixture leaves stay valid under the same rule; this leaf pins it
  explicitly against false `...` insertion.
- Color is `never` so ellipsis checks are not confused with ANSI.

## Steps

1. Set `req.Grep = []string{"GREP_SNIP_SHORT_TOKEN"}`, `req.Limit = 10`, `req.Color = "never"`.
2. Write one session whose `generated_title` is a short phrase containing the token.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Grep = []string{"GREP_SNIP_SHORT_TOKEN"}
	req.Limit = 10
	req.Color = "never"

	writeGrokSession(t, req.GrokHome,
		"01900021-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		"2026-07-03T14:30:00.000Z",
		"/tmp/grep-snippet-short",
		"Ship GREP_SNIP_SHORT_TOKEN quickly")
	return nil
}
```
