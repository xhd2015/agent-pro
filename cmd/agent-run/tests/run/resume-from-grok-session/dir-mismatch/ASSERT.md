## Expected

- Exit code 1.
- Error indicates `--dir` / workspace mismatch with the Grok session cwd
  (hard error; no relocate).

## Exit Code

1

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)
	combined := combinedOut(resp)
	assertContainsAny(t, combined,
		"dir",
		"cwd",
		"workspace",
		"mismatch",
		"does not match",
		"differ",
		"different",
		"relocate",
	)
}
```
