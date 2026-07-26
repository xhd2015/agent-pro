# llm-mock run grok Tests

Doc-style tests for `llm-mock run grok` and the shortcut binary `llm-mock-run-grok`.
The orchestrator starts an OpenAI-compatible mock HTTP server in the background,
configures an isolated `GROK_HOME` with `config.toml` pointing LLM traffic at the mock,
runs grok in the foreground (interactive TUI by default), and tears down the mock when
grok exits.

# DSN (Domain Specific Notion)

**Participants**

- **llm-mock orchestrator** — `run.RunGrok()` shared by `llm-mock run grok` and
  `llm-mock-run-grok`; resolves config, merges events input, spawns mock server,
  writes isolated grok `config.toml`, sets `GROK_HOME` / `XAI_API_KEY` /
  `GROK_MODELS_BASE_URL`, runs grok, waits for exit, tears down mock.
- **llm-mock HTTP server** — background process loading JSON config (`port`, `exchanges[]`)
  plus optional `LLM_MOCK_EVENTS_FILE` exchanges appended after config exchanges.
- **mockconfig loader** — shared config resolution (`--config` > `LLM_MOCK_CONFIG_FILE` >
  `LLM_MOCK_CONFIG` > default `{"exchanges": []}`) and events JSONL merge with duplicate-index detection.
- **Isolated GROK_HOME** — fresh temp dir by default (removed on exit) or explicit
  `LLM_MOCK_GROK_HOME`; contains generated `config.toml` and grok session tree under
  `sessions/<url-encoded-cwd>/<uuid>/`.
- **grok CLI** — real interactive TUI or headless `-p` mode; production `grok` on PATH
  for integration leaves; `LLM_MOCK_RUN_GROK_COMMAND` hook for plumbing tests.
- **Shortcut binary** — `llm-mock-run-grok` thin wrapper calling `run.RunGrok(os.Args[1:])`.
- **Run log-events** — optional `--log-events <path>.jsonl` on `llm-mock run` only; validated
  before grok starts; orchestrator passes `--agent-events-file <path>` to the background mock server
  so each **served** mock response appends standard `agent/event/types` `AgentEvent` JSONL
  (`type`: `think`, `message`, `tool_call`, …). Separate from mock `--events-file` RecordedRequest log.
- **Run log-http** — optional `--log-http <path>.jsonl` on `llm-mock run` only; validated before
  grok starts; orchestrator passes `--log-http <path>` to the background mock server for full HTTP
  round-trip debug JSONL (request/response headers and bodies, or streaming `chunks[]`).
- **Run mock-events-preset** — optional `--mock-events-preset <name>` on `llm-mock run`; `list`
  prints catalog and exits 0 without grok/mock; other names pass through to background mock
  `--mock-events-preset` to seed `genQueue` after config exchanges.
- **LLM_MOCK_RUN_FLAGS** — space-separated run flags (GOFLAGS-style) prepended before argv
  parsing via simple whitespace split (`strings.Fields`); applies to `llm-mock run grok` and
  `llm-mock-run-grok`; unset or empty is a no-op; explicit CLI flags on `llm-mock run` come
  after env tokens and win on duplicate flags (last-wins).

**Behaviors**

- Config optional: no `LLM_MOCK_CONFIG_FILE`, `LLM_MOCK_CONFIG`, or `--config` → inline default
  `{"exchanges": []}`; `llm-mock run grok` with no config env starts grok successfully.
- `LLM_MOCK_EVENTS_FILE` is **input** JSONL of additional exchanges appended after config
  `exchanges[]`; unset/empty is OK.
- Orchestrator writes `config.toml` with `models_base_url` → `http://127.0.0.1:<port>/v1`,
  default model `mock-model`, and sets matching env vars.
- Fake grok hook (`LLM_MOCK_RUN_GROK_COMMAND`) replaces grok argv for deterministic
  plumbing tests (config resolution, grok home, CLI entry).
- Integration leaf runs headless real grok (`-p`, `--always-approve`, `-m mock-model`),
  asserts mocked LLM response in output and `turn_started` with `model_id: mock-model`
  in newest session `events.jsonl`.
- `--log-events` and `--log-http` use `lessflags` with `StopOnFirstArg()` so tokens after `grok`
  are grok argv unchanged; each path must end with `.jsonl` or CLI errors before mock/grok start.
- `--log-events` and `--log-http` are combinable on the same `llm-mock run` invocation.
- `--mock-events-preset` uses `lessflags` `StopOnFirstArg()`; tokens after `grok` are grok argv
  unchanged; combinable with `--log-events` / `--log-http`.
- `LLM_MOCK_RUN_FLAGS` prepends default run flags before `ParseRunFlags`; shortcut binary has no
  run-flag argv — env is the only source for `llm-mock-run-grok`; duplicate flag on `llm-mock run`
  CLI overrides the env value.

## Version

0.0.2

## Decision Tree

```
run-grok/
├── config-resolution/
│   ├── no-config-starts-grok/     # no config env → fake grok exit 0
│   ├── config-file-env/           # LLM_MOCK_CONFIG_FILE + fake grok → exit 0
│   └── legacy-config-env/         # LLM_MOCK_CONFIG only + fake grok → exit 0
├── events-input/
│   ├── appends-exchanges/         # config 1 exchange + events file 1 more → 2nd response
│   └── optional-unset/            # no LLM_MOCK_EVENTS_FILE → config-only works
├── grok-home/
│   ├── default-temp-home/         # fake grok prints GROK_HOME in temp dir
│   └── explicit-home/             # LLM_MOCK_GROK_HOME set, config.toml written
├── cli-entry/
│   ├── subcommand/                # llm-mock run grok smoke (fake grok)
│   └── shortcut-binary/           # llm-mock-run-grok smoke (fake grok)
├── random-fallback/
│   └── two-turn-no-config/        # no config; fake grok 3 curls; turn 2 must not no_match
├── log-events/                    # --log-events run flag → mock --agent-events-file output
│   ├── appends-requests/          # fake grok 2 curls → ≥2 message AgentEvent JSONL lines
│   ├── random-fallback-think-message/ # no config, 1 curl → think then message AgentEvents
│   ├── requires-jsonl-suffix/     # bad suffix → error before grok; no log file
│   ├── grok-args-after-grok/      # -p hello passed through after grok token
│   └── unset-no-file/             # omit flag → no log-events JSONL in workdir
├── log-http/                      # --log-http run flag → mock --log-http output
│   ├── appends-round-trip/        # fake grok 1 curl → log has request path + response status 200
│   ├── requires-jsonl-suffix/     # bad suffix → error before grok; no log file
│   ├── grok-args-after-grok/      # -p hello passed through when --log-http set
│   └── unset-no-file/             # omit flag → no log-http JSONL in workdir
├── teardown/                      # orchestrator must exit promptly after grok
│   ├── exits-promptly-empty-session/ # session dir w/o events.jsonl → exit <5s (not 60s poll)
│   ├── exits-promptly-no-session/    # no sessions/ tree → exit <5s
│   └── no-spurious-mirror-warning/   # no events.jsonl → no "mirror sessions: not ready" stderr
├── mock-events-preset/            # --mock-events-preset run flag → mock genQueue seeding
│   ├── list-no-grok/              # run --mock-events-preset=list exits 0 without grok/mock
│   ├── pass-through-think-message/ # preset think-message + fake grok 2 curls → think then message
│   └── grok-args-after-preset/    # --mock-events-preset=simple grok -p hello passes argv through
├── run-flags-from-env/            # LLM_MOCK_RUN_FLAGS prepends run flags before argv parse
│   ├── subcommand-log-events/     # env --log-events; llm-mock run grok → AgentEvent JSONL
│   ├── shortcut-log-events/       # env --log-events; llm-mock-run-grok → AgentEvent JSONL
│   ├── cli-overrides-env/         # env a.jsonl + CLI b.jsonl → only b.jsonl written
│   ├── preset-list-subcommand/    # env --mock-events-preset=list; run grok → catalog, no grok
│   └── preset-list-shortcut/      # env --mock-events-preset=list; shortcut → catalog, no grok
├── shortcut-env-help/             # LLM_MOCK_RUN_FLAGS=--help → help text, no GROK_HOME=
├── shortcut-forwards-grok-flags/  # llm-mock-run-grok --always-approve → fake grok argv
└── integration/                   # label: real-grok, slow
    ├── headless-mock-response/    # real grok -p → Paris + grok home events proof
    └── headless-no-config-hello/  # no config, random fallback → assistant reply within 30s
```

Parameter ranking (most → least significant):

1. **Config source** — no config (default empty) vs `LLM_MOCK_CONFIG_FILE` vs `LLM_MOCK_CONFIG`
2. **Events input** — unset vs `LLM_MOCK_EVENTS_FILE` append
3. **GROK_HOME** — default temp vs explicit `LLM_MOCK_GROK_HOME`
4. **Log events output** — `--log-events` unset vs set (valid `.jsonl` vs invalid suffix)
5. **Log HTTP output** — `--log-http` unset vs set (valid `.jsonl` vs invalid suffix)
6. **Mock events preset** — unset vs `list` (catalog only) vs named preset pass-through
7. **Run flags source** — CLI argv vs `LLM_MOCK_RUN_FLAGS` env (CLI wins on duplicate)
8. **CLI entry** — `llm-mock run grok` vs `llm-mock-run-grok` shortcut
9. **Grok backend** — fake hook vs real grok (`label: real-grok, slow`)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `config-resolution/no-config-starts-grok` | No config env vars or file → fake grok exit 0, `GROK_HOME=` printed |
| 2 | `config-resolution/config-file-env` | `LLM_MOCK_CONFIG_FILE` set, fake grok → exit 0 |
| 3 | `config-resolution/legacy-config-env` | `LLM_MOCK_CONFIG` only (legacy), fake grok → exit 0 |
| 4 | `events-input/appends-exchanges` | Config 1 exchange + events file 1 more; fake grok curls twice → 2nd response from events |
| 5 | `events-input/optional-unset` | No `LLM_MOCK_EVENTS_FILE`; fake grok single curl → config exchange only |
| 6 | `grok-home/default-temp-home` | Fake grok prints `GROK_HOME`; path is fresh temp dir with `config.toml` |
| 7 | `grok-home/explicit-home` | `LLM_MOCK_GROK_HOME` set; `config.toml` written with `models_base_url` to mock |
| 8 | `cli-entry/subcommand` | `llm-mock run grok` smoke with fake grok → exit 0 |
| 9 | `cli-entry/shortcut-binary` | `llm-mock-run-grok` smoke with fake grok → exit 0 |
| 10 | `random-fallback/two-turn-no-config` | No mock config; fake grok curls think+message then second user turn; must not `no_match` |
| 11 | `log-events/appends-requests` | `--log-events <tmp>.jsonl`; fake grok curls twice → ≥2 `type:message` AgentEvents with `from-config`/`from-events` text |
| 12 | `log-events/random-fallback-think-message` | No config; fake grok one curl → log has `think` then `message` AgentEvents (≥2 lines) |
| 13 | `log-events/requires-jsonl-suffix` | `--log-events /tmp/events.log` → non-zero exit, `.jsonl` in error, grok not started |
| 14 | `log-events/grok-args-after-grok` | `run --log-events f.jsonl grok -p hello`; fake grok on PATH echoes `-p hello` |
| 15 | `log-events/unset-no-file` | Omit `--log-events`; fake grok curl → no log-events JSONL written under workdir |
| 16 | `log-http/appends-round-trip` | `--log-http <tmp>.jsonl`; fake grok one curl → log has `request.path` `/v1/chat/completions` and `response.status` 200 |
| 17 | `log-http/requires-jsonl-suffix` | `--log-http bad.log grok` → non-zero exit, `.jsonl` in error, grok not started |
| 18 | `log-http/grok-args-after-grok` | `run --log-http f.jsonl grok -p hello`; fake grok echoes `-p hello` |
| 19 | `log-http/unset-no-file` | Omit `--log-http`; fake grok curl → no log-http JSONL written under workdir |
| 20 | `teardown/exits-promptly-empty-session` | Fake grok creates session without `events.jsonl`; `--log-events`; orchestrator exits within 5s (not 60s mirror poll) |
| 21 | `teardown/exits-promptly-no-session` | Fake grok immediate exit; `--log-events`; orchestrator exits within 5s |
| 22 | `teardown/no-spurious-mirror-warning` | Session without `events.jsonl`; stderr must not print `mirror sessions: not ready` |
| 23 | `integration/headless-mock-response` | Real grok `-p` headless; stdout contains Paris; `events.jsonl` has `turn_started` `model_id: mock-model` (`label: real-grok, slow`) |
| 24 | `integration/headless-no-config-hello` | No mock config; real grok `-p hello`; stdout has assistant text and `events.jsonl` has `first_token` within 30s (`label: real-grok, slow`) |
| 25 | `mock-events-preset/list-no-grok` | `llm-mock run --mock-events-preset=list` exits 0 without grok/mock |
| 26 | `mock-events-preset/pass-through-think-message` | `--mock-events-preset=think-message` + fake grok 2 curls → think then message in log-events |
| 27 | `mock-events-preset/grok-args-after-preset` | `--mock-events-preset=simple grok -p hello` passes `-p hello` to grok |
| 28 | `run-flags-from-env/subcommand-log-events` | `LLM_MOCK_RUN_FLAGS=--log-events <tmp>.jsonl`; `llm-mock run grok` (no CLI run flags) → ≥2 `type:message` AgentEvents |
| 29 | `run-flags-from-env/shortcut-log-events` | Same env; `llm-mock-run-grok` → log file created with ≥2 message AgentEvents |
| 30 | `run-flags-from-env/cli-overrides-env` | Env `a.jsonl` + CLI `--log-events b.jsonl` → only `b.jsonl` written (CLI wins) |
| 31 | `run-flags-from-env/preset-list-subcommand` | `LLM_MOCK_RUN_FLAGS=--mock-events-preset=list`; `llm-mock run grok` → catalog stdout, grok not started |
| 32 | `run-flags-from-env/preset-list-shortcut` | Same env; `llm-mock-run-grok` → catalog stdout, grok not started |
| 33 | `shortcut-env-help` | `LLM_MOCK_RUN_FLAGS=--help llm-mock-run-grok` → exit 0, help text, no `GROK_HOME=` |
| 34 | `shortcut-forwards-grok-flags` | `llm-mock-run-grok --always-approve` → fake grok `GROK_ARGV=--always-approve` |

## Coverage

- **Config resolution**: optional default empty exchanges, env priority (`LLM_MOCK_CONFIG_FILE`, legacy `LLM_MOCK_CONFIG`)
- **Events input**: append after config exchanges, optional unset
- **GROK_HOME**: default temp isolation, explicit home + generated `config.toml`
- **CLI entry**: subcommand and shortcut binary
- **Random fallback**: multi-turn no-config session must not 400 `no_match` after stream exhaustion
- **Log events**: `--log-events` wires mock `--agent-events-file` (AgentEvent JSONL on serve), prefix/random-fallback logging, `.jsonl` suffix validation, grok argv passthrough
- **Log HTTP**: `--log-http` wires mock `--log-http` (full HTTP exchange JSONL), `.jsonl` suffix validation, grok argv passthrough, unset produces no file
- **Mock events preset**: `--mock-events-preset` catalog (`list` without grok), orchestrator pass-through to mock `genQueue`, grok argv passthrough after preset flag
- **Run flags from env**: `LLM_MOCK_RUN_FLAGS` prepends `--log-events` / `--mock-events-preset` for subcommand and shortcut; CLI duplicate overrides env; `list` via env without grok
- **Integration**: real grok headless with mocked LLM routing proof via session events

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./agent/llm/llm-mock/tests/run-grok                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./agent/llm/llm-mock/tests/run-grok
doctest test --label-all ./agent/llm/llm-mock/tests/run-grok

# Plumbing tests (fake grok hook) — default CI
doctest vet ./agent/llm/llm-mock/tests/run-grok
doctest test ./agent/llm/llm-mock/tests/run-grok

# Single leaf
doctest test ./agent/llm/llm-mock/tests/run-grok/config-resolution/config-file-env

# LLM_MOCK_RUN_FLAGS grouping only
doctest test ./agent/llm/llm-mock/tests/run-grok/run-flags-from-env/...

# Shortcut argv / env help leaves
doctest test ./agent/llm/llm-mock/tests/run-grok/shortcut-env-help
doctest test ./agent/llm/llm-mock/tests/run-grok/shortcut-forwards-grok-flags

# Real grok integration (requires grok on PATH)
doctest test --label real-grok ./agent/llm/llm-mock/tests/run-grok/integration/...
```

```go
import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot       string
	BinaryPath     string // llm-mock
	ShortcutPath   string // llm-mock-run-grok
	UseShortcut    bool
	ConfigJSON     string
	ConfigEnv      string // "file" | "legacy" | "" (neither — error test)
	ConfigPath     string // written by Run when ConfigJSON set
	EventsJSONL    string
	EventsPath     string // written by Run when EventsJSONL set
	GrokHome       string // LLM_MOCK_GROK_HOME; empty = orchestrator picks temp
	GrokArgs       []string
	WorkDir        string
	FakeGrokCmd    string // LLM_MOCK_RUN_GROK_COMMAND
	GrokPathPrepend string // prepend to PATH (fake grok binary); skips hook when set
	LogEventsPath  string // --log-events output path; empty = flag omitted
	LogHTTPPath    string // --log-http output path; empty = flag omitted
	MockEventsPreset string // --mock-events-preset value (preset name or "list")
	ListOnly       bool   // run --mock-events-preset only, omit grok subcommand
	RunFlagsEnv    string // LLM_MOCK_RUN_FLAGS value (space-separated run flags)
	OmitCLIRunFlags bool  // when true, do not pass --log-events/--log-http/--mock-events-preset on CLI
	EnvLogEventsOverridePath string // cli-overrides-env: env-only path that must not be written
	SkipRealGrok   bool
	ExpectedExit   int
	ExecTimeout    time.Duration
	ExpectConfig   bool // post-run: require config.toml under grok home
	ExpectParis    bool // integration: stdout must contain Paris
	ExpectMockModel bool // integration: events.jsonl turn_started model_id mock-model
	ExpectAssistantReply bool // integration: stdout has assistant text; events.jsonl has first_token
}

type Response struct {
	ExitCode        int
	Stdout          string
	Stderr          string
	GrokHomeUsed    string
	GrokSessionDir  string
	GrokEventsPath  string
	GrokEventsLines []string
	ConfigToml        string
	LogEventsContent  string
	LogEventsLines    []string
	LogHTTPContent    string
	LogHTTPLines      []string
	Err               error
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	resp := &Response{}

	workDir := req.WorkDir
	if workDir == "" {
		workDir = t.TempDir()
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir workdir: %w", err)
	}

	env := envWithOverrides(os.Environ(), nil)

	if req.ConfigJSON != "" {
		configPath := filepath.Join(t.TempDir(), "llm-mock-config.json")
		if err := os.WriteFile(configPath, []byte(req.ConfigJSON), 0644); err != nil {
			return nil, fmt.Errorf("write config: %w", err)
		}
		req.ConfigPath = configPath
		switch req.ConfigEnv {
		case "file":
			env = envWithOverrides(env, map[string]string{"LLM_MOCK_CONFIG_FILE": configPath})
		case "legacy":
			env = envWithOverrides(env, map[string]string{"LLM_MOCK_CONFIG": configPath})
		}
	}

	if req.EventsJSONL != "" {
		eventsPath := filepath.Join(t.TempDir(), "llm-mock-events-input.jsonl")
		if err := os.WriteFile(eventsPath, []byte(req.EventsJSONL), 0644); err != nil {
			return nil, fmt.Errorf("write events: %w", err)
		}
		req.EventsPath = eventsPath
		env = envWithOverrides(env, map[string]string{"LLM_MOCK_EVENTS_FILE": eventsPath})
	}

	if req.GrokHome != "" {
		if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
			return nil, fmt.Errorf("mkdir grok home: %w", err)
		}
		env = envWithOverrides(env, map[string]string{"LLM_MOCK_GROK_HOME": req.GrokHome})
	}

	if req.RunFlagsEnv != "" {
		env = envWithOverrides(env, map[string]string{"LLM_MOCK_RUN_FLAGS": req.RunFlagsEnv})
	}

	if req.GrokPathPrepend != "" {
		env = envWithOverrides(env, map[string]string{
			"PATH": req.GrokPathPrepend + string(os.PathListSeparator) + os.Getenv("PATH"),
		})
	} else if req.FakeGrokCmd != "" {
		env = envWithOverrides(env, map[string]string{"LLM_MOCK_RUN_GROK_COMMAND": req.FakeGrokCmd})
	}

	timeout := req.ExecTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	if req.UseShortcut {
		args := append([]string{}, req.GrokArgs...)
		cmd = exec.CommandContext(ctx, req.ShortcutPath, args...)
	} else {
		args := []string{"run"}
		if !req.OmitCLIRunFlags {
			if req.LogEventsPath != "" {
				args = append(args, "--log-events", req.LogEventsPath)
			}
			if req.LogHTTPPath != "" {
				args = append(args, "--log-http", req.LogHTTPPath)
			}
			if req.MockEventsPreset != "" {
				args = append(args, "--mock-events-preset", req.MockEventsPreset)
			}
		}
		if !req.ListOnly {
			args = append(args, "grok")
			args = append(args, req.GrokArgs...)
		}
		cmd = exec.CommandContext(ctx, req.BinaryPath, args...)
	}
	cmd.Dir = workDir
	cmd.Env = env

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	runErr := cmd.Run()
	resp.Stdout = stdoutBuf.String()
	resp.Stderr = stderrBuf.String()
	combined := resp.Stdout + resp.Stderr

	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			resp.ExitCode = exitErr.ExitCode()
		} else if ctx.Err() != nil {
			resp.ExitCode = -1
			resp.Err = fmt.Errorf("timeout after %s: %w", timeout, runErr)
		} else {
			resp.ExitCode = -1
			resp.Err = runErr
		}
	}

	if resp.GrokHomeUsed == "" {
		resp.GrokHomeUsed = parseGrokHomeFromOutput(combined)
	}
	if resp.GrokHomeUsed == "" && req.GrokHome != "" {
		resp.GrokHomeUsed = req.GrokHome
	}

	if req.ExpectConfig || req.GrokHome != "" || strings.Contains(req.FakeGrokCmd, "GROK_HOME") {
		home := resp.GrokHomeUsed
		if home != "" {
			data, err := os.ReadFile(filepath.Join(home, "config.toml"))
			if err == nil {
				resp.ConfigToml = string(data)
			}
		}
	}

	if req.ExpectMockModel || req.ExpectParis || req.ExpectAssistantReply {
		home := resp.GrokHomeUsed
		if home == "" {
			home = req.GrokHome
		}
		if home != "" {
			sessionDir, eventsPath, lines, err := FindNewestGrokSessionEvents(home, workDir)
			if err == nil {
				resp.GrokSessionDir = sessionDir
				resp.GrokEventsPath = eventsPath
				resp.GrokEventsLines = lines
			}
		}
	}

	if req.LogEventsPath != "" {
		data, readErr := os.ReadFile(req.LogEventsPath)
		if readErr == nil {
			resp.LogEventsContent = string(data)
			resp.LogEventsLines, _ = readJSONLLinesFromContent(string(data))
		}
	}
	if req.LogHTTPPath != "" {
		data, readErr := os.ReadFile(req.LogHTTPPath)
		if readErr == nil {
			resp.LogHTTPContent = string(data)
			resp.LogHTTPLines, _ = readJSONLLinesFromContent(string(data))
		}
	}

	return resp, nil
}
```