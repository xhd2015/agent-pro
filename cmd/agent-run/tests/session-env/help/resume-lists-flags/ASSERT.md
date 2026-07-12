## Expected

- Exit code 0.
- Stdout documents `--prepend-path`.
- Stdout documents `-e` and/or `--env`.
- Stdout ends with trailing newline `\n`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, "--prepend-path")
	out := resp.Stdout
	if !strings.Contains(out, "-e") && !strings.Contains(out, "--env") {
		t.Fatalf("resume --help must document -e and/or --env; stdout:\n%s", out)
	}
	assertTrailingNewline(t, resp.Stdout, "resume help stdout")
}
```
