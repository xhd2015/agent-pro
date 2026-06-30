## Expected

- Built prompt contains `first-topic` and `second-topic`.
- Index of `first-topic` is less than index of `second-topic` (chronological order).

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	iFirst := indexOf(t, resp.BuiltPrompt, "first-topic")
	iSecond := indexOf(t, resp.BuiltPrompt, "second-topic")
	if iFirst >= iSecond {
		t.Fatalf("expected first-topic before second-topic, indices %d vs %d in:\n%s", iFirst, iSecond, resp.BuiltPrompt)
	}
	assertContains(t, resp.BuiltPrompt, "summarize both topics")
}
```