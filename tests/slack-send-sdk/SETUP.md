# Scenario

**Feature**: slack-send CLI doctest harness builds binary and prepares isolated workdirs

```
# session cache: build slack-send once; optional slacktest server for unit sends
doctest -> build slack-send -> temp workdir (go.mod + slack-config.json) -> exec -> capture stdout/stderr/exit
send-success branch -> SLACK_API_URL -> slacktest fake API -> OK ts=... line
integration branch -> repo slack-config.json + real slack.com
```

## Preconditions

- `go` available in PATH.
- `script/debug/slack-send/main.go` exists.
- Root `testdata/valid-config.json` and `testdata/empty-token-config.json` fixtures.
- `send-success` / `send-errors` leaves require implementer to add `SLACK_API_URL` support with `github.com/slack-go/slack` (tests expected RED until then).
- Session cache dir: `$TMPDIR/slack-send-sdk-doctest-<DOCTEST_SESSION_ID>/` (shared binary + slacktest URL).

## Steps

1. Resolve module root from `DOCTEST_ROOT/../..`.
2. Build `slack-send` once per session into cache dir.
3. Leaf/grouping `Setup` sets `req.Args`, config fixture, `SlackAPIURL`, or `UseRepoConfig`.
4. `Run` writes workdir, executes binary, captures output.

## Context

- Isolated tests use a minimal `go.mod` in `WorkDir` so `findConfigPath` stops at that directory.
- Channel-resolve leaves use fake `botToken` and assert the `Sending to channel=...` line; send may fail afterward.
- Success stdout must end with trailing newline after the `OK` line.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("skipping: go not found in PATH")
	}
	repoRoot, err := findModuleRoot()
	if err != nil {
		return err
	}
	req.RepoRoot = repoRoot
	bin, err := buildSlackSend(t)
	if err != nil {
		return err
	}
	req.Bin = bin
	if req.WriteGoMod == false && !req.UseRepoConfig && req.WorkDir == "" {
		// help leaves: still need a cwd; temp dir without go.mod is fine
		req.WorkDir = t.TempDir()
	}
	return nil
}

func assertStderrContains(t *testing.T, resp *Response, substr string) {
	t.Helper()
	if !strings.Contains(resp.Stderr, substr) {
		t.Fatalf("stderr missing %q\ngot:\n%s", substr, resp.Stderr)
	}
}

func assertStdoutContains(t *testing.T, resp *Response, substr string) {
	t.Helper()
	if !strings.Contains(resp.Stdout, substr) {
		t.Fatalf("stdout missing %q\ngot:\n%s", substr, resp.Stdout)
	}
}
```