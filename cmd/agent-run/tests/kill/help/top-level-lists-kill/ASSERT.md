## Expected

- Exit code 0.
- Stdout lists `kill` as its own top-level command line (not only the word
  inside `pty` “kill orphan” description).
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
	// Command inventory lines look like: "  kill       stop a live …"
	// Do not match the word "kill" inside the pty "kill orphan" blurb.
	cmdLine := regexp.MustCompile(`(?m)^[ \t]+kill\b`)
	if !cmdLine.MatchString(resp.Stdout) {
		t.Fatalf("top-level --help must list kill as a command line, got:\n%s", resp.Stdout)
	}
	assertTrailingNewline(t, resp.Stdout, "top-level --help stdout")
}
```
