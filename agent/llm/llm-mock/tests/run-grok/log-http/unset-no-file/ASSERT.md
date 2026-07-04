## Expected

- Exit code 0.
- No `*.jsonl` files created under the test workdir (log-http output not written when flag omitted).

## Exit Code

0

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)

	matches, globErr := filepath.Glob(filepath.Join(req.WorkDir, "*.jsonl"))
	if globErr != nil {
		t.Fatalf("glob workdir jsonl: %v", globErr)
	}
	if len(matches) > 0 {
		t.Fatalf("unexpected jsonl files in workdir without --log-http: %v", matches)
	}
}
```