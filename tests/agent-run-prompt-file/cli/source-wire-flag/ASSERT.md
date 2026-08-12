## Expected

- ScannedFiles > 0.
- FlagFound true: production sources mention `--prompt-file`
  (less-flags String registration and/or help text).

## Side Effects

- None (read-only scan).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.ScannedFiles == 0 {
		t.Fatal("expected to scan pkgs/agentruncli .go files")
	}
	if !resp.FlagFound {
		t.Fatal("pkgs/agentruncli must register/document --prompt-file on agent-run run")
	}
}
```
