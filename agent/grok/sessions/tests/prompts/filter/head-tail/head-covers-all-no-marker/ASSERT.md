## Expected

- No error.
- 3 prompts a,b,c; OmittedAfter=0; OmittedBefore=0.
- Output contains a,b,c.
- Output does **not** contain `omitted`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertPromptCount(t, resp.Single, 3)
	if resp.Single.OmittedAfter != 0 || resp.Single.OmittedBefore != 0 {
		t.Fatalf("omitted counts want 0,0 got before=%d after=%d",
			resp.Single.OmittedBefore, resp.Single.OmittedAfter)
	}
	out := resp.Output
	assertContains(t, out, "a")
	assertContains(t, out, "b")
	assertContains(t, out, "c")
	assertNotContains(t, out, "omitted")
	assertTrailingNewline(t, out)
}
```
