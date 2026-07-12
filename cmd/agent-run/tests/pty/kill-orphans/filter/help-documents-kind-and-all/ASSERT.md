## Expected

- Exit code 0.
- Stdout documents `--kind` and `--all`.
- Stdout ends with trailing newline `\n`.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assertOutput(t, resp, "stdout", "--kind", "--all")
	assertTrailingNewline(t, resp.Stdout, "kill-orphans --help stdout")
}
```
