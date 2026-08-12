## Expected

- Help source (`run_cmd.go`) was read successfully.
- Source text documents `--prompt-file` (in `runHelp` and/or flag registration).

## Side Effects

- None (read-only file).

## Errors

- None (missing flag is an assert failure, not a harness error).

## Exit Code

N/A

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp == nil || !resp.HelpSourceRead {
		t.Fatal("expected run_cmd.go help source to be read")
	}
	if !strings.Contains(resp.Stdout, flagPromptFile) {
		t.Fatalf("run help surface (run_cmd.go) must document %s; got %d bytes",
			flagPromptFile, len(resp.Stdout))
	}
	// Locked name is --prompt-file, not --prompt-from-file as the only form.
	if strings.Contains(resp.Stdout, "--prompt-from-file") && !strings.Contains(resp.Stdout, flagPromptFile) {
		t.Fatal("must use locked flag name --prompt-file")
	}
}
```
