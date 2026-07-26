---
label: e2e, real-codex, slow
explanation: Requires real codex CLI on PATH; live usage values change over time.
---

## Expected Output

```
<contains>
<regex>^Monthly usage: \d+%$</regex>
<regex>^Credits used: \d+ of \d+$</regex>
<start-with>Next reset: </start-with>
</contains>
```

## Expected

- Exit code 0.
- Stdout has a `Monthly usage: <digits>%` line.
- Stdout has a `Credits used: <digits> of <digits>` line.
- Stdout has a `Next reset: <non-empty>` line.
- Stderr is empty.
- Exact credit counts and reset date are **not** asserted (live values change).

## Side Effects

- Ephemeral tty-watch session killed after fetch.

## Errors

- None on success.

## Exit Code

0

```go
import (
	"regexp"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"

	"github.com/xhd2015/doctest/session"
)

var monthlyUsageRE = regexp.MustCompile(`(?m)^Monthly usage: \d+%$`)
var creditsUsedRE = regexp.MustCompile(`(?m)^Credits used: \d+ of \d+$`)
var nextResetRE = regexp.MustCompile(`(?m)^Next reset: .+$`)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccessExit(t, resp)

	stdout := strings.TrimSpace(resp.Stdout)
	if !monthlyUsageRE.MatchString(stdout) {
		t.Fatalf("stdout missing Monthly usage: <digits>%% line:\n%s", resp.Stdout)
	}
	if !creditsUsedRE.MatchString(stdout) {
		t.Fatalf("stdout missing Credits used: <digits> of <digits> line:\n%s", resp.Stdout)
	}
	if !nextResetRE.MatchString(stdout) {
		t.Fatalf("stdout missing Next reset: <non-empty> line:\n%s", resp.Stdout)
	}

	assert.Output(t, stdout, `<contains>
<regex>^Monthly usage: \d+%$</regex>
<regex>^Credits used: \d+ of \d+$</regex>
<start-with>Next reset: </start-with>
</contains>`)
}
```