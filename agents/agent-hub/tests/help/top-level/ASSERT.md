## Expected
- Exit code 0.
- Stdout contains "Usage:", lists all top-level commands with descriptions: daemon, notify, hook, fetch, replay, status, consumers, sessions, partitions, session, integration.
- Mentions "Run agent-hub <command> --help for command-specific options."

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertExitCode(t, resp, 0)
    assertContains(t, resp.Stdout, "Usage:")
    assertContains(t, resp.Stdout, "daemon")
    assertContains(t, resp.Stdout, "notify")
    assertContains(t, resp.Stdout, "hook")
    assertContains(t, resp.Stdout, "fetch")
    assertContains(t, resp.Stdout, "replay")
    assertContains(t, resp.Stdout, "consumers")
    assertContains(t, resp.Stdout, "sessions")
    assertContains(t, resp.Stdout, "partitions")
    assertContains(t, resp.Stdout, "session")
    assertContains(t, resp.Stdout, "integration")
    assertContains(t, resp.Stdout, "--help")
}
```
