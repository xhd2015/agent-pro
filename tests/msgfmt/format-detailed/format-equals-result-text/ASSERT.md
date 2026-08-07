## Expected

- `resp.Text == resp.Detail.Text` (primary contract).
- Non-empty output with multi header `showing 3 of 4`.
- At least one truncated body (`BodiesTruncated >= 1`).

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	assertFormatEqualsDetail(t, resp)
	if resp.Text == "" {
		t.Fatal("expected non-empty formatted text")
	}
	assertContains(t, resp.Text, "Chat history (showing 3 of 4):")
	if resp.Detail.BodiesTruncated < 1 {
		t.Fatalf("BodiesTruncated=%d, want >= 1", resp.Detail.BodiesTruncated)
	}
	assertEqualInt(t, "Shown", resp.Detail.Shown, 3)
	assertEqualInt(t, "SourceCount", resp.Detail.SourceCount, 4)
}
```
