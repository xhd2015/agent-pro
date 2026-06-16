## Expected
- Output contains ASSISTANT marker.
- **After fix: only the Delta (" feature.") appears, NOT the full accumulated Content text.**
- The full accumulated text string "The user has given me..." must NOT appear.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "ASSISTANT")
	// After fix: formatter shows delta only, not accumulated text
	assertContains(t, resp.Output, " feature.")
	assertNotContains(t, resp.Output, "The user has given me a detailed requirement for creating a macOS menu bar app. I need to design a comprehensive doctest tree for this feature.")
}
```
