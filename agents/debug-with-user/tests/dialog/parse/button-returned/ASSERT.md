## Expected

- Parse succeeds without error.
- `Button` equals `Yes — window opened`.
- `Text` is empty (preset button path).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != nil {
		t.Fatal(resp.Err)
	}
	if resp.Button != "Yes — window opened" {
		t.Fatalf("Button = %q, want %q", resp.Button, "Yes — window opened")
	}
	if resp.Text != "" {
		t.Fatalf("Text = %q, want empty for button-only output", resp.Text)
	}
}
```
