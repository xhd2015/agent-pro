## Expected

- Pane text written to -o path; stdout empty.

```go
import (
	"os"
	"path/filepath"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	if resp.Stdout != "" {
		t.Fatalf("stdout must be empty when -o set; got %q", resp.Stdout)
	}
	out := filepath.Join(req.TempDir, "pane.txt")
	body, readErr := os.ReadFile(out)
	if readErr != nil {
		t.Fatalf("read output file: %v", readErr)
	}
	if string(body) != "file pane\n" {
		t.Fatalf("file body = %q, want file pane\\n", body)
	}
}
```
