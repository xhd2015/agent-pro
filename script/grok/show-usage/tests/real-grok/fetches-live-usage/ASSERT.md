---
label: real-grok, slow
explanation: Requires real grok CLI on PATH; live usage reset date changes daily.
---

## Expected Output

```
<contains>
<regex>^Weekly limit: \d+%$</regex>
<start-with>Next reset: </start-with>
</contains>
```

## Expected

- Exit code 0.
- Stdout has a `Weekly limit: <digits>%` line.
- Stdout has a `Next reset: <non-empty>` line.
- Stderr is empty.
- Exact reset date is **not** asserted (live value changes).

## Side Effects

- None (ephemeral PTY session).

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
)

var weeklyLimitRE = regexp.MustCompile(`(?m)^Weekly limit: \d+%$`)
var nextResetRE = regexp.MustCompile(`(?m)^Next reset: .+$`)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccessExit(t, resp)

	stdout := strings.TrimSpace(resp.Stdout)
	if !weeklyLimitRE.MatchString(stdout) {
		t.Fatalf("stdout missing Weekly limit: <digits>%% line:\n%s", resp.Stdout)
	}
	if !nextResetRE.MatchString(stdout) {
		t.Fatalf("stdout missing Next reset: <non-empty> line:\n%s", resp.Stdout)
	}

	assert.Output(t, stdout, `<contains>
<regex>^Weekly limit: \d+%$</regex>
<start-with>Next reset: </start-with>
</contains>`)
}
```