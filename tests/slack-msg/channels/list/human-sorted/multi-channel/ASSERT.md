---
label: unit
explanation: human list lines sorted by name; member column; archived excluded
---

## Expected Output

```
C0AGENTDBG1  #agent-pro-debug  private  -
C0ALE44K5J6  #general  public  member
C0OTHERCHAN  #random  public  -
```

## Expected

- Exit code 0.
- Three human lines sorted by channel name ascending.
- Format: `id  #name  kind  member|-` with two spaces between columns.
  - kind: `public` / `private`
  - member: `member` if is_member else `-`
- Archived `#old-stuff` not listed.
- Trailing newline after last line; stderr empty.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "old-stuff") || strings.Contains(resp.Stdout, "C0ARCHIVED1") {
		t.Fatalf("archived channel must be excluded by default:\n%s", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
C0AGENTDBG1  #agent-pro-debug  private  -
C0ALE44K5J6  #general  public  member
C0OTHERCHAN  #random  public  -
`)
}
```
