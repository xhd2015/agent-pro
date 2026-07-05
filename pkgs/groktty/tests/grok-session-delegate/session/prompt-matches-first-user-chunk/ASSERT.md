## Expected

- `DiscoverSession` returns the seeded session UUID.
- `updatesPath` points to an existing `updates.jsonl`.

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.SessionID != req.SessionUUID {
		t.Fatalf("session id: got %q want %q", resp.SessionID, req.SessionUUID)
	}
	if resp.UpdatesPath == "" {
		t.Fatal("updates path is empty")
	}
	if _, err := os.Stat(resp.UpdatesPath); err != nil {
		t.Fatalf("updates path missing: %v (%s)", err, resp.UpdatesPath)
	}
}
```