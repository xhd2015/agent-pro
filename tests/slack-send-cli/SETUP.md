# Scenario

**Feature**: slack-send CLI doctest harness builds binary and prepares isolated workdirs

```
# session cache: build slack-send once; slacktest with conversations.list for unit sends
doctest -> build slack-send -> temp workdir -> exec with controlled env/args -> capture stdout/stderr/exit
unit branch -> SLACK_API_URL -> slacktest (conversations.list + chat.postMessage) -> OK ts=... line
integration branch -> --config repo slack-config.json + real slack.com
```

## Preconditions

- `go` available in PATH.
- `script/debug/slack-send/main.go` exists.
- Root `testdata/` fixtures: `valid-config.json`, `empty-token-config.json`, `default-channel-name.json`.
- Unit leaves require implementer: `less-flags` CLI, explicit `--config` only, required MESSAGE, `conversations.list` resolution, `SLACK_API_URL` hook.
- Session cache dir: `$TMPDIR/slack-send-cli-doctest-<DOCTEST_SESSION_ID>/` (shared binary + slacktest URLs).

## Steps

1. Resolve module root from `DOCTEST_ROOT/../..`.
2. Build `slack-send` once per session into cache dir.
3. Leaf/grouping `Setup` sets `req.Args`, config fixture, `SlackAPIURL`, or env overrides.
4. `Run` materializes config file when requested, executes binary, captures output.

## Context

- No auto `slack-config.json` discovery; config only via explicit `--config` in args.
- Validation-error leaves clear `SLACK_BOT_TOKEN`, `SLACK_CHANNEL`, `SLACK_CONFIG` from env.
- Success stdout must end with trailing newline after the `OK` line.

```go
import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
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

func withConfigArg(t *testing.T, req *Request, fixtureOrInline string, isInline bool) error {
	t.Helper()
	if isInline {
		req.ConfigInline = fixtureOrInline
	} else {
		req.ConfigFixture = fixtureOrInline
	}
	if err := materializeConfig(t, req); err != nil {
		return err
	}
	req.Args = prependConfigArgs(req.Args, req.ConfigPath)
	return nil
}

func prependConfigArgs(args []string, configPath string) []string {
	prefix := []string{"--config", configPath}
	if len(args) == 0 {
		return prefix
	}
	return append(prefix, args...)
}
```