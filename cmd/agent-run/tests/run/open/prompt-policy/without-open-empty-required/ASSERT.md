## Expected

- Exit code ≠ 0.
- Error contains `prompt is required` (existing wording).

## Errors

- Missing prompt without `--open`.

## Exit Code

non-zero

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit when prompt empty without --open; stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errText := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	if !strings.Contains(errText, "prompt is required") {
		t.Fatalf("want %q in error surface:\nstderr:\n%s\nstdout:\n%s",
			"prompt is required", resp.Stderr, resp.Stdout)
	}
}
```
