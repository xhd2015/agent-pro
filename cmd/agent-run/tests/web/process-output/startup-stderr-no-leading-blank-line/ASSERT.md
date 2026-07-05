## Expected Output

Stderr should start immediately with the no-token warning (no leading blank line):

```text
no API token configured
--token <secret> or --token auto to require Bearer authentication
agent-run web listening at http://127.0.0.1:<port>
```

## Expected

- Stderr does not begin with `\n`.
- First bytes of stderr are `no API token`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	stderr := webProcessStderrText(req)
	if strings.HasPrefix(stderr, "\n") {
		t.Fatalf("stderr must not start with blank line, got:\n%q", stderr)
	}
	if !strings.HasPrefix(stderr, "no API token") {
		t.Fatalf("stderr must start with %q, got:\n%q", "no API token", stderr)
	}
}
```