## Expected

Exact text:

```text
Chat history (showing 2 of 3):
MID_UNIQUE
NEW_UNIQUE
```

- Does not contain `OLD_UNIQUE`
- `Shown=2`, `SourceCount=3`, `OldestDropped=1`
- `Text` rune count ≤ `opts.TotalBudgetRunes`

## Errors

- None from `Run`.

```go
import (
	"testing"
	"unicode/utf8"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoError(t, err)
	assertResp(t, resp)
	want := "" +
		"Chat history (showing 2 of 3):\n" +
		"MID_UNIQUE\n" +
		"NEW_UNIQUE\n"
	assertEqualString(t, "Text", resp.Text, want)
	assertNotContains(t, resp.Text, "OLD_UNIQUE")
	assertEqualInt(t, "Shown", resp.Detail.Shown, 2)
	assertEqualInt(t, "SourceCount", resp.Detail.SourceCount, 3)
	assertEqualInt(t, "OldestDropped", resp.Detail.OldestDropped, 1)
	if n := utf8.RuneCountInString(resp.Text); n > req.Opts.TotalBudgetRunes {
		t.Fatalf("Text runes=%d exceeds TotalBudgetRunes=%d", n, req.Opts.TotalBudgetRunes)
	}
	assertFormatEqualsDetail(t, resp)
}
```
