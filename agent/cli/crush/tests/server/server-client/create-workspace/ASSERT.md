## Expected
- `createWorkspace` returns a non-empty workspace ID.
- The workspace is created on the server (can be probed).

```go
import (
	"encoding/json"
	"testing"
)

type workspaceResult struct {
	ID string `json:"id"`
}

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("create-workspace failed: %v", err)
	}
	if resp.Output == "" {
		t.Fatal("expected non-empty Output with workspace result")
	}
	var result workspaceResult
	if err := json.Unmarshal([]byte(resp.Output), &result); err != nil {
		t.Fatalf("failed to parse workspace result: %v\noutput: %s", err, resp.Output)
	}
	if result.ID == "" {
		t.Fatal("expected non-empty workspace ID")
	}
}
```
