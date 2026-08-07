## Expected

- `TRIGGER_LAST` appears in `Text` on a locked id-only line.
- `FIRST_DROP_ME` does not appear.
- `Shown=1`, `SourceCount=2`, `OldestDropped=1`
- `LastMessageID="m2"`
- Exact text (last-only block; budget may be exceeded):

```text
Chat history (showing 1 of 2):
message_id=m2 : TRIGGER_LAST
```

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	want := "" +
		"Chat history (showing 1 of 2):\n" +
		"message_id=m2 : TRIGGER_LAST\n"
	assertEqualString(t, "Text", resp.Text, want)
	assertContains(t, resp.Text, "TRIGGER_LAST")
	assertNotContains(t, resp.Text, "FIRST_DROP_ME")
	assertEqualInt(t, "Shown", resp.Detail.Shown, 1)
	assertEqualInt(t, "SourceCount", resp.Detail.SourceCount, 2)
	assertEqualInt(t, "OldestDropped", resp.Detail.OldestDropped, 1)
	assertEqualString(t, "LastMessageID", resp.Detail.LastMessageID, "m2")
	assertFormatEqualsDetail(t, resp)
}
```
