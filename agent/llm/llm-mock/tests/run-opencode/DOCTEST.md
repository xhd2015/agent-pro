# llm-mock run opencode Tests

Doc-style tests for `llm-mock run opencode` and the shortcut binary `llm-mock-run-opencode`.
The orchestrator starts an OpenAI-compatible mock HTTP server in the background,
configures an isolated opencode environment with inline `OPENCODE_CONFIG_CONTENT` pointing LLM
traffic at the mock (`@ai-sdk/openai-compatible` → `POST /v1/chat/completions`), runs opencode
in the foreground, and tears down the mock when opencode exits.

# DSN (Domain Specific Notion)

**Participants**

- **llm-mock orchestrator** — `run.RunOpencode()` shared by `llm-mock run opencode` and
  `llm-mock-run-opencode`; resolves config, spawns mock server, writes isolated opencode env
  (`HOME`, `OPENCODE_CONFIG_DIR`, `OPENCODE_CONFIG_CONTENT`), runs opencode, waits for exit,
  tears down mock promptly (no grok-style session mirror poll).
- **llm-mock HTTP server** — background process loading JSON config (`port`, `exchanges[]`)
  plus optional `LLM_MOCK_EVENTS_FILE` exchanges appended after config exchanges.
- **mockconfig loader** — shared config resolution (`--config` > `LLM_MOCK_CONFIG_FILE` >
  `LLM_MOCK_CONFIG` > default `{"exchanges": []}`) and events JSONL merge with duplicate-index detection.
- **Isolated opencode env** — fresh temp dirs by default (removed on exit) or explicit
  `LLM_MOCK_OPENCODE_HOME` / `LLM_MOCK_OPENCODE_CONFIG_DIR`; orchestrator sets
  `OPENCODE_CONFIG_CONTENT` with `@ai-sdk/openai-compatible` provider routing to mock.
- **opencode CLI** — real headless `opencode run` for integration leaves; production `opencode`
  on PATH for integration; `LLM_MOCK_RUN_OPENCODE_COMMAND` hook for plumbing tests.
- **Shortcut binary** — `llm-mock-run-opencode` thin wrapper calling `run.RunOpencode(os.Args[1:])`.
- **Run log-events** — optional `--log-events <path>.jsonl` on `llm-mock run` only; validated
  before opencode starts; orchestrator passes `--agent-events-file <path>` to the background mock server
  so each **served** mock response appends standard `agent/event/types` `AgentEvent` JSONL
  (`type`: `think`, `message`, `tool_call`, …).
- **Run log-http** — optional `--log-http <path>.jsonl` on `llm-mock run` only; validated before
  opencode starts; orchestrator passes `--log-http <path>` to the background mock server for full HTTP
  round-trip debug JSONL (request/response headers and bodies, or streaming `chunks[]`).
- **Run mock-events-preset** — optional `--mock-events-preset <name>` on `llm-mock run`; `list`
  prints catalog and exits 0 without opencode/mock; other names pass through to background mock
  `--mock-events-preset` to seed `genQueue` after config exchanges.

**Behaviors**

- Config optional: no `LLM_MOCK_CONFIG_FILE`, `LLM_MOCK_CONFIG`, or `--config` → inline default
  `{"exchanges": []}`; `llm-mock run opencode` with no config env starts opencode successfully.
- Orchestrator writes `OPENCODE_CONFIG_CONTENT` with `provider.llm-mock.options.baseURL` →
  `http://127.0.0.1:<port>/v1`, default model `llm-mock/mock-model`, permissions allow questions/plans.
- Orchestrator announces `OPENCODE_CONFIG_DIR=<path>` on stderr; sets isolation env vars
  (`OPENCODE_DISABLE_PROJECT_CONFIG=1`, `OPENCODE_PURE=1`, `OPENCODE_DISABLE_MODELS_FETCH=1`,
  `OPENCODE_AUTH_CONTENT={}`).
- Fake opencode hook (`LLM_MOCK_RUN_OPENCODE_COMMAND`) replaces opencode argv for deterministic
  plumbing tests; echoes `OPENCODE_CONFIG_DIR=$OPENCODE_CONFIG_DIR` and optionally curls mock via
  `baseURL` read from `$OPENCODE_CONFIG_CONTENT`.
- Opencode uses OpenAI Chat Completions API (`POST /v1/chat/completions`) against the mock provider.
- Integration leaf runs real `opencode run "What is the capital of France?" --model llm-mock/mock-model`
  with mocked Paris response in output; exit 0 not required (agent loop may continue).
- `--log-events` and `--log-http` use `lessflags` with `StopOnFirstArg()` so tokens after `opencode`
  are opencode argv unchanged; each path must end with `.jsonl` or CLI errors before mock/opencode start.
- `--mock-events-preset` uses `lessflags` `StopOnFirstArg()`; tokens after `opencode` are opencode argv
  unchanged; combinable with `--log-events` / `--log-http`.

## Version

0.0.2

## Decision Tree

```
run-opencode/
├── cli-entry/
│   ├── subcommand/                # llm-mock run opencode smoke (fake opencode)
│   └── shortcut-binary/           # llm-mock-run-opencode smoke (fake opencode)
├── opencode-home/
│   ├── default-temp-home/         # fake opencode prints OPENCODE_CONFIG_DIR in temp dir
│   └── explicit-home/             # LLM_MOCK_OPENCODE_CONFIG_DIR + LLM_MOCK_OPENCODE_HOME set
├── config-resolution/
│   ├── no-config-starts-opencode/ # no config env → fake opencode exit 0
│   └── config-file-env/           # LLM_MOCK_CONFIG_FILE + Paris exchange → fake exit 0
├── log-events/                    # --log-events run flag → mock --agent-events-file output
│   ├── appends-agent-events/      # preset simple + fake curl → message AgentEvent
│   ├── requires-jsonl-suffix/   # bad suffix → error before opencode; no log file
│   └── opencode-args-after-opencode/ # run argv passed through after opencode token
├── log-http/                      # --log-http run flag → mock --log-http output
│   ├── appends-round-trip/        # fake opencode 1 curl → log has /v1/chat/completions + status 200
│   ├── requires-jsonl-suffix/     # bad suffix → error before opencode; no log file
│   └── opencode-args-after-opencode/ # argv passthrough when --log-http set
├── mock-events-preset/            # --mock-events-preset run flag → mock genQueue seeding
│   ├── list-no-opencode/          # run --mock-events-preset=list exits 0 without opencode/mock
│   └── pass-through-simple/       # preset simple + fake curl → message in log-events
└── integration/                   # label: real-opencode, slow
    └── headless-mock-response/    # real opencode run → Paris in output; exit 0 not required
```

Parameter ranking (most → least significant):

1. **CLI entry** — `llm-mock run opencode` vs `llm-mock-run-opencode` shortcut
2. **Opencode home** — default temp vs explicit `LLM_MOCK_OPENCODE_HOME` / `LLM_MOCK_OPENCODE_CONFIG_DIR`
3. **Config source** — no config (default empty) vs `LLM_MOCK_CONFIG_FILE`
4. **Log events output** — `--log-events` set (valid `.jsonl` vs invalid suffix) vs argv passthrough
5. **Log HTTP output** — `--log-http` set (valid `.jsonl` vs invalid suffix) vs argv passthrough
6. **Mock events preset** — `list` (catalog only) vs named preset pass-through
7. **Opencode backend** — fake hook vs real opencode (`label: real-opencode, slow`)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `cli-entry/subcommand` | `llm-mock run opencode` smoke with fake opencode → exit 0, `OPENCODE_CONFIG_DIR=` |
| 2 | `cli-entry/shortcut-binary` | `llm-mock-run-opencode` smoke with fake opencode → exit 0 |
| 3 | `opencode-home/default-temp-home` | Fake opencode prints `OPENCODE_CONFIG_DIR`; path is fresh temp dir, not user `~/.config/opencode` |
| 4 | `opencode-home/explicit-home` | `LLM_MOCK_OPENCODE_CONFIG_DIR` + `LLM_MOCK_OPENCODE_HOME` set; `OPENCODE_CONFIG_CONTENT` has mock baseURL |
| 5 | `config-resolution/no-config-starts-opencode` | No config env vars → fake opencode exit 0, `OPENCODE_CONFIG_DIR=` printed |
| 6 | `config-resolution/config-file-env` | `LLM_MOCK_CONFIG_FILE` with Paris exchange, fake opencode → exit 0 |
| 7 | `log-events/appends-agent-events` | `--mock-events-preset=simple` + fake curl → log has `message` AgentEvent |
| 8 | `log-events/requires-jsonl-suffix` | `--log-events /tmp/events.log` → non-zero exit, `.jsonl` in error, opencode not started |
| 9 | `log-events/opencode-args-after-opencode` | `run --log-events f.jsonl opencode run -m llm-mock/mock-model hi`; fake opencode echoes argv |
| 10 | `log-http/appends-round-trip` | `--log-http <tmp>.jsonl`; fake opencode one curl → log has `request.path` `/v1/chat/completions` and `response.status` 200 |
| 11 | `log-http/requires-jsonl-suffix` | `--log-http bad.log opencode` → non-zero exit, `.jsonl` in error, opencode not started |
| 12 | `log-http/opencode-args-after-opencode` | `run --log-http f.jsonl opencode run -m llm-mock/mock-model hi`; fake opencode echoes argv |
| 13 | `mock-events-preset/list-no-opencode` | `llm-mock run --mock-events-preset=list` exits 0 without opencode/mock |
| 14 | `mock-events-preset/pass-through-simple` | `--mock-events-preset=simple` + fake curl → message in log-events |
| 15 | `integration/headless-mock-response` | Real `opencode run`; output contains Paris; mock HTTP log has chat/completions with mock model (`label: real-opencode, slow`; exit 0 not required) |

## Coverage

- **CLI entry**: subcommand and shortcut binary
- **Opencode home**: default temp isolation, explicit home + generated `OPENCODE_CONFIG_CONTENT`
- **Config resolution**: optional default empty exchanges, `LLM_MOCK_CONFIG_FILE`
- **Log events**: `--log-events` wires mock `--agent-events-file` (AgentEvent JSONL on serve), `.jsonl` suffix validation, opencode argv passthrough
- **Log HTTP**: `--log-http` wires mock `--log-http` (Chat Completions API `/v1/chat/completions` exchange JSONL), `.jsonl` suffix validation, opencode argv passthrough
- **Mock events preset**: `--mock-events-preset` catalog (`list` without opencode), orchestrator pass-through to mock `genQueue` (`simple`)
- **Integration**: real opencode headless with mocked LLM routing proof via stdout and HTTP log

## How to Run

```sh
# Discovery skips labeled e2e/heavy/slow leaves by default.
# Run e2e / full suite explicitly when needed:
doctest test ./agent/llm/llm-mock/tests/run-opencode                    # discovery (skips labeled e2e/heavy/slow)
doctest test --label e2e ./agent/llm/llm-mock/tests/run-opencode
doctest test --label-all ./agent/llm/llm-mock/tests/run-opencode

# Plumbing tests (fake opencode hook) — default CI
doctest vet ./agent/llm/llm-mock/tests/run-opencode
doctest test ./agent/llm/llm-mock/tests/run-opencode

# Single leaf
doctest test ./agent/llm/llm-mock/tests/run-opencode/config-resolution/config-file-env

# Real opencode integration (requires opencode on PATH)
doctest test --label real-opencode ./agent/llm/llm-mock/tests/run-opencode/integration/...
```

```go
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

type Request struct {
	RepoRoot            string
	BinaryPath          string // llm-mock
	ShortcutPath        string // llm-mock-run-opencode
	UseShortcut         bool
	ConfigJSON          string
	ConfigEnv           string // "file" | "" (neither — error test)
	ConfigPath          string // written by Run when ConfigJSON set
	OpencodeHome        string // LLM_MOCK_OPENCODE_HOME; empty = orchestrator picks temp
	OpencodeConfigDir   string // LLM_MOCK_OPENCODE_CONFIG_DIR; empty = orchestrator picks temp
	OpencodeArgs        []string
	WorkDir             string
	FakeOpencodeCmd     string // LLM_MOCK_RUN_OPENCODE_COMMAND
	OpencodePathPrepend string // prepend to PATH (fake opencode binary); skips hook when set
	LogEventsPath       string // --log-events output path; empty = flag omitted
	LogHTTPPath         string // --log-http output path; empty = flag omitted
	MockEventsPreset    string // --mock-events-preset value (preset name or "list")
	ListOnly            bool   // run --mock-events-preset only, omit opencode subcommand
	SkipRealOpencode    bool
	ExpectedExit        int
	ExecTimeout         time.Duration
	ExpectConfig        bool // post-run: require OPENCODE_CONFIG_CONTENT with mock provider
	ExpectParis         bool // integration: output must contain Paris
}

type Response struct {
	ExitCode              int
	Stdout                string
	Stderr                string
	OpencodeConfigDirUsed string
	OpencodeHomeUsed      string
	OpencodeConfigContent string
	LogEventsContent      string
	LogEventsLines        []string
	LogHTTPContent        string
	LogHTTPLines          []string
	Err                   error
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
		if req.ConfigEnv == "file" {
			env = envWithOverrides(env, map[string]string{"LLM_MOCK_CONFIG_FILE": configPath})
		}
	}

	if req.OpencodeHome != "" {
		if err := os.MkdirAll(req.OpencodeHome, 0755); err != nil {
			return nil, fmt.Errorf("mkdir opencode home: %w", err)
		}
		env = envWithOverrides(env, map[string]string{"LLM_MOCK_OPENCODE_HOME": req.OpencodeHome})
	}
	if req.OpencodeConfigDir != "" {
		if err := os.MkdirAll(req.OpencodeConfigDir, 0755); err != nil {
			return nil, fmt.Errorf("mkdir opencode config dir: %w", err)
		}
		env = envWithOverrides(env, map[string]string{"LLM_MOCK_OPENCODE_CONFIG_DIR": req.OpencodeConfigDir})
	}

	if req.OpencodePathPrepend != "" {
		env = envWithOverrides(env, map[string]string{
			"PATH": req.OpencodePathPrepend + string(os.PathListSeparator) + os.Getenv("PATH"),
		})
	} else if req.FakeOpencodeCmd != "" {
		env = envWithOverrides(env, map[string]string{"LLM_MOCK_RUN_OPENCODE_COMMAND": req.FakeOpencodeCmd})
	}

	timeout := req.ExecTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	if req.UseShortcut {
		if req.ShortcutPath == "" {
			return nil, fmt.Errorf("shortcut binary not built (llm-mock-run-opencode missing)")
		}
		args := append([]string{}, req.OpencodeArgs...)
		cmd = exec.CommandContext(ctx, req.ShortcutPath, args...)
	} else {
		args := []string{"run"}
		if req.LogEventsPath != "" {
			args = append(args, "--log-events", req.LogEventsPath)
		}
		if req.LogHTTPPath != "" {
			args = append(args, "--log-http", req.LogHTTPPath)
		}
		if req.MockEventsPreset != "" {
			args = append(args, "--mock-events-preset", req.MockEventsPreset)
		}
		if !req.ListOnly {
			args = append(args, "opencode")
			args = append(args, req.OpencodeArgs...)
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

	if resp.OpencodeConfigDirUsed == "" {
		resp.OpencodeConfigDirUsed = parseOpencodeConfigDirFromOutput(combined)
	}
	if resp.OpencodeConfigDirUsed == "" && req.OpencodeConfigDir != "" {
		resp.OpencodeConfigDirUsed = req.OpencodeConfigDir
	}
	if resp.OpencodeHomeUsed == "" {
		resp.OpencodeHomeUsed = parseOpencodeHomeFromOutput(combined)
	}
	if resp.OpencodeHomeUsed == "" && req.OpencodeHome != "" {
		resp.OpencodeHomeUsed = req.OpencodeHome
	}

	if req.ExpectConfig || req.OpencodeConfigDir != "" || strings.Contains(req.FakeOpencodeCmd, "OPENCODE_CONFIG_DIR") {
		resp.OpencodeConfigContent = parseOpencodeConfigContentFromEnv(env)
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