# Scenario

**Feature**: `agent-run run` idle-exit flags — help + parse-only validation

```
# help
agent-run run -h -> RunHelpText documents --exit-on-idle and --idle-timeout (default 10m)

# parse (no TTY)
ParseRunIdle(exitOnIdle, timeoutRaw) -> enabled / duration / Error: on stderr
```

## Preconditions

- Nested DOCTEST root under `cmd/agent-run/tests/run-exit-on-idle/` (does not
  inherit parent `cmd/agent-run/tests` Setup/Run).
- Product APIs under design (RED until implementer):
  - `agentruncli.RunHelpText() string` — `run -h` body must document
    `--exit-on-idle` and `--idle-timeout` (default `10m`)
  - `agentruncli.ParseRunIdle(exitOnIdle bool, timeoutRaw string) (enabled bool, d time.Duration, err error)`
    — parse-only; no TTY
- Parallel-safe: no `os.Setenv` / `Chdir`; no process stdio reassignment.
  Help uses pure `RunHelpText()`. Parse never calls `Handle` (would start a
  run if flags parsed as strings).

## Steps

1. Root `Setup` validates Request.
2. Branch sets `Op` (`help` or `parse`).
3. Leaf sets `Args` (help documentary) or `ExitOnIdle` / `IdleTimeoutRaw`.
4. Root `Run` dispatches; leaf `Assert` checks stdout/stderr/exit.

## Context

- CLI presentation of parse errors: `Error: <msg>\n`, exit 1.
- Preferred messages locked by invalid leaves.
- Trailing newline required on CLI stdout (help).

```go
import (
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	if req == nil {
		return fmt.Errorf("nil Request")
	}
	return nil
}

func idleHelpExcerpt(help string) string {
	var out []string
	lines := strings.Split(strings.TrimSuffix(help, "\n"), "\n")
	capture := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.Contains(line, "--exit-on-idle") || strings.Contains(line, "--idle-timeout") {
			capture = true
			out = append(out, line)
			continue
		}
		if capture {
			if strings.HasPrefix(line, " ") && !strings.HasPrefix(trim, "--") {
				out = append(out, line)
				continue
			}
			break
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n") + "\n"
}
```
