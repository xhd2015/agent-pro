## Expected

- Parse succeeds without error.
- `Text` equals the typed user report.
- `Button` is empty (free-text path).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != nil {
		t.Fatal(resp.Err)
	}
	if resp.Text != "VS Code opened but wrong workspace" {
		t.Fatalf("Text = %q, want %q", resp.Text, "VS Code opened but wrong workspace")
	}
	if resp.Button != "" {
		t.Fatalf("Button = %q, want empty for text-only output", resp.Button)
	}
}
```
