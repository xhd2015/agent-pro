# Scenario

**Feature**: slack-msg CLI doctest harness builds binary and prepares isolated workdirs

```
# session cache: build slack-msg once; slacktest for unit send/history/channels/auth; listen daemon probes
doctest -> build ./cmd/slack-msg -> temp workdir -> exec with controlled env/args -> capture stdout/stderr/exit

# send / history / channels / auth / session unit branch
SLACK_API_URL -> slacktest (conversations.list + chat.postMessage + history/replies + auth.test + apps.connections.open)
session store under $HOME/.agent-pro/slack-local-bot (sessions.json + messages.jsonl)

# listen unit branch
slacktest Socket Mode + mock agent-run (SLACK_LISTEN_AGENT_RUN; handles tty status ready)
  -> agentrunbridge RunInteractiveOpen (thread) / Run+CaptureStdout (stateless)
  -> chat.postMessage only for stateless agent body
  -> default lock / banner / dedupe / operator logs / SYSTEM.md + sessions map + -e env inject

# integration branch
--config repo slack-config.json + real slack.com
```

## Preconditions

- `go` available in PATH.
- Implementer lands `cmd/slack-msg` (RED until then — build fails).
- Root `testdata/`: `valid-config.json`, `empty-token-config.json`, `empty-app-token-config.json`, `default-channel-name.json`.
- Session cache dir: `$TMPDIR/slack-msg-doctest-<DOCTEST_SESSION_ID>/` (shared binary + slacktest URLs).
- Unit leaves require: `less-flags` CLI, explicit `--config` only, `SLACK_API_URL` hook, chronological history, listen `--token` (not `--bot-token`).

## Steps

1. Resolve module root from `DOCTEST_ROOT` upward.
2. Build `slack-msg` once per session into cache dir from `./cmd/slack-msg`.
3. Leaf/grouping `Setup` sets `req.Args`, config fixture, `SlackAPIURL`, listen flags, or injected events.
4. `Run` materializes config, executes binary (quick or daemon), captures output.

## Context

- No auto `slack-config.json` discovery; config only via explicit `--config` / `SLACK_CONFIG`.
- Validation-error leaves clear Slack-related env vars.
- User-facing stdout ends with trailing newline after last content line.
- Lock message (listen): `another slack-msg is already running`.
- Default listen lock path: `$HOME/.agent-pro/slack-msg.listen.lock` (daemon harness
  auto-isolates with WorkDir lock unless `UseDefaultLock` / `NoLock`).
- Session store: `$HOME/.agent-pro/slack-local-bot/{sessions.json,sessions/<id>/…}`
  (isolate via `req.HomeDir` on CLI + daemon).
- Session reply posts capture: `req.CapturePosts` for simple-run PostMessage asserts.

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
	bin, err := buildSlackMsg(t)
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

func assertOutputContains(t *testing.T, resp *Response, substr string) {
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
	if err := materializeConfig(t, req); err != nil {
		return err
	}
	if !req.ListenMode {
		req.Args = insertConfigAfterSubcommand(req.Args, req.ConfigPath)
	}
	return nil
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
