## Expected

- Section `Top background tasks` present.
- Header string includes `EXIT` and `COMMAND`:
  exact fragment `#  DURATION  EXIT  COMMAND`.
- The failed-command row is visible with command `false-cmd-exit-one`.
- EXIT column value **1** appears on the same table row as `false-cmd-exit-one`
  (not only as a rank number elsewhere).
- Command `ok-cmd-exit-zero` appears; EXIT **0** for that row is acceptable
  (integer `0`, not `-`).
- When exit_code is nil, EXIT would be `-` — not exercised here.

## Errors

- None.

```go
import (
	"regexp"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Stats == nil {
		t.Fatal("stats is nil")
	}
	out := resp.Output
	assertContains(t, out, "Top background tasks")
	assertContains(t, out, "#  DURATION  EXIT  COMMAND")

	bg := sectionAfter(out, "Top background tasks")
	if bg == "" {
		t.Fatalf("missing top-bg section:\n%s", out)
	}
	assertContains(t, bg, "false-cmd-exit-one")
	assertContains(t, bg, "ok-cmd-exit-zero")

	// Row with failing command must show exit code 1 as a column (word boundary).
	// Example shape: "  1     5s     1  false-cmd-exit-one"
	failLine := ""
	for _, line := range strings.Split(bg, "\n") {
		if strings.Contains(line, "false-cmd-exit-one") {
			failLine = line
			break
		}
	}
	if failLine == "" {
		t.Fatalf("no row for false-cmd-exit-one:\n%s", bg)
	}
	// EXIT column is an integer field before COMMAND; require a standalone 1
	// that is not only the leading rank when rank could also be 1.
	// Accept any whitespace-separated token "1" before the command text.
	re := regexp.MustCompile(`\b1\b`)
	// Find all "1" tokens; need at least one that is the exit column.
	// Heuristic: strip leading rank, then require \b1\b before the command.
	cmdIdx := strings.Index(failLine, "false-cmd-exit-one")
	beforeCmd := failLine[:cmdIdx]
	// After optional rank and duration, exit code 1 must appear.
	if !re.MatchString(beforeCmd) {
		t.Fatalf("EXIT column missing 1 before command on fail row:\n%q\nsection:\n%s", failLine, bg)
	}
	// Stronger: last integer token before command is 1.
	fields := strings.Fields(beforeCmd)
	if len(fields) < 2 {
		t.Fatalf("fail row too short for EXIT column: %q", failLine)
	}
	// fields typically: rank, duration, exit
	exitTok := fields[len(fields)-1]
	if exitTok != "1" {
		t.Fatalf("EXIT token before command = %q, want 1 (row %q)", exitTok, failLine)
	}
}
```
