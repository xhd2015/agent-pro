## Expected

Exact text:

```text
Chat history (showing 3 of 3):
alpha
beta
gamma
```

- `OldestDropped=0`, `Shown=3`, `SourceCount=3`

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
		"Chat history (showing 3 of 3):\n" +
		"alpha\n" +
		"beta\n" +
		"gamma\n"
	assertEqualString(t, "Text", resp.Text, want)
	assertEqualInt(t, "OldestDropped", resp.Detail.OldestDropped, 0)
	assertEqualInt(t, "Shown", resp.Detail.Shown, 3)
	assertEqualInt(t, "SourceCount", resp.Detail.SourceCount, 3)
	assertFormatEqualsDetail(t, resp)
}
```
