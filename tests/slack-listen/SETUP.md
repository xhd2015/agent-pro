# Scenario

**Feature**: slack-listen doctest harness builds binary and runs quick-exit or daemon probes

```
# session cache: build slack-listen once; per-leaf temp workdir
doctest -> build slack-listen -> exec listen (quick or daemon) -> capture stdout/stderr/exit

# unit daemon branch
slacktest (apps.connections.open + WS) <- SLACK_API_URL
inject events_api envelope -> slack-listen -> mock agent-run (SLACK_LISTEN_AGENT_RUN) -> chat.postMessage

# integration branch
--config repo slack-config.json + live Socket Mode + real agent-run
```

## Preconditions

- `go` available in PATH.
- `script/debug/slack-listen/` exists (RED until P1 implementer lands).
- Root `testdata/`: `valid-config.json`, `empty-app-token-config.json`.
- Unit leaves require: `SLACK_API_URL` hook, `SLACK_LISTEN_AGENT_RUN` env for mock agent,
  startup `Using config from:` log line, singleton lock behavior.
- Session cache dir: `$TMPDIR/slack-listen-doctest-<DOCTEST_SESSION_ID>/` (shared binary).

## Steps

1. Resolve module root from `DOCTEST_ROOT/../..`.
2. Build `slack-listen` once per session into cache dir.
3. Leaf/grouping `Setup` sets `req.Args`, tokens, config, daemon flags, or injected events.
4. `Run` executes quick-exit or daemon probe and captures observations.

## Context

- No auto `slack-config.json` discovery; config only via explicit `--config`.
- Validation-error leaves clear `SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN`, `SLACK_CONFIG` from env.
- Mock agent script appends `INVOCATION ...` lines to `SLACK_LISTEN_AGENT_LOG` and prints fixed reply.

```go
import (
	"os/exec"
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
	bin, err := buildSlackListen(t)
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
	combined := resp.Stdout + resp.Stderr
	if !strings.Contains(combined, substr) {
		t.Fatalf("output missing %q\nstdout:\n%s\nstderr:\n%s", substr, resp.Stdout, resp.Stderr)
	}
}

func withConfigArg(t *testing.T, req *Request, fixtureOrInline string, isInline bool) error {
	t.Helper()
	if isInline {
		req.ConfigInline = fixtureOrInline
	} else {
		req.ConfigFixture = fixtureOrInline
	}
	return materializeConfig(t, req)
}

func prependListenTokens(req *Request) {
	if req.BotToken == "" {
		req.BotToken = slackTestBotToken
	}
	if req.AppToken == "" {
		req.AppToken = slackTestAppToken
	}
}
```