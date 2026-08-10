## Expected

- Exit code 0.
- Stdout lists `takeover` as its own top-level command line (inventory style:
  leading whitespace + `takeover` word boundary — not a random substring).
- Stdout ends with trailing newline `\n`.

## Exit Code

0

```go
import (
	"regexp"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	// Command inventory lines look like: "  takeover   adopt a live …"
	cmdLine := regexp.MustCompile(`(?m)^[ \t]+takeover\b`)
	if !cmdLine.MatchString(resp.Stdout) {
		t.Fatalf("top-level --help must list takeover as a command line, got:\n%s", resp.Stdout)
	}
	assertTrailingNewline(t, resp.Stdout, "top-level --help stdout")
}
```
