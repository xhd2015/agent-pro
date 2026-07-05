# llm-mock run codex Tests

Doc-style tests for `llm-mock run codex` and the shortcut binary `llm-mock-run-codex`.
The orchestrator starts an OpenAI-compatible mock HTTP server in the background,
configures an isolated `CODEX_HOME` with `config.toml` pointing LLM traffic at the mock
(`wire_api = "responses"`), runs codex in the foreground, and tears down the mock when
codex exits.

# DSN (Domain Specific Notion)

**Participants**

- **llm-mock orchestrator** — `run.RunCodex()` shared by `llm-mock run codex` and
  `llm-mock-run-codex`; resolves config, spawns mock server, writes isolated codex
  `config.toml`, sets `CODEX_HOME` / `OPENAI_API_KEY`, runs codex, waits for exit,
  tears down mock promptly (no grok-style session mirror poll).
- **llm-mock HTTP server** — background process loading JSON config (`port`, `exchanges[]`)
  plus optional `LLM_MOCK_EVENTS_FILE` exchanges appended after config exchanges.
- **mockconfig loader** — shared config resolution (`--config` > `LLM_MOCK_CONFIG_FILE` >
  `LLM_MOCK_CONFIG` > default `{"exchanges": []}`) and events JSONL merge with duplicate-index detection.
- **Isolated CODEX_HOME** — fresh temp dir by default (removed on exit) or explicit
  `LLM_MOCK_CODEX_HOME`; contains generated `config.toml` and codex session transcripts
  under `sessions/<date>/<workspace-hash>/<uuid>/`.
- **codex CLI** — real headless `codex exec` for integration leaves; production `codex`
  on PATH for integration; `LLM_MOCK_RUN_CODEX_COMMAND` hook for plumbing tests.
- **Shortcut binary** — `llm-mock-run-codex` thin wrapper calling `run.RunCodex(os.Args[1:])`.
- **Run log-events** — optional `--log-events <path>.jsonl` on `llm-mock run` only; validated
  before codex starts; orchestrator passes `--agent-events-file <path>` to the background mock server
  so each **served** mock response appends standard `agent/event/types` `AgentEvent` JSONL
  (`type`: `think`, `message`, `tool_call`, …).
- **Run log-http** — optional `--log-http <path>.jsonl` on `llm-mock run` only; validated before
  codex starts; orchestrator passes `--log-http <path>` to the background mock server for full HTTP
  round-trip debug JSONL (request/response headers and bodies, or streaming `chunks[]`).
- **Run mock-events-preset** — optional `--mock-events-preset <name>` on `llm-mock run`; `list`
  prints catalog and exits 0 without codex/mock; other names pass through to background mock
  `--mock-events-preset` to seed `genQueue` after config exchanges.

**Behaviors**

- Config optional: no `LLM_MOCK_CONFIG_FILE`, `LLM_MOCK_CONFIG`, or `--config` → inline default
  `{"exchanges": []}`; `llm-mock run codex` with no config env starts codex successfully.
- Orchestrator writes `config.toml` with `model_providers.llm-mock.base_url` →
  `http://127.0.0.1:<port>/v1`, `wire_api = "responses"`, `approval_policy = "never"`,
  `[features] shell_tool = false`, default model `mock-model`.
- Orchestrator announces `CODEX_HOME=<path>` on stderr; sets `OPENAI_API_KEY=sk-mock`.
- Fake codex hook (`LLM_MOCK_RUN_CODEX_COMMAND`) replaces codex argv for deterministic
  plumbing tests; echoes `CODEX_HOME=$CODEX_HOME` and optionally curls mock via base URL
  read from `$CODEX_HOME/config.toml`.
- Codex uses OpenAI Responses API (`POST /v1/responses`) against the mock provider.
- Integration leaf runs real `codex exec --skip-git-repo-check -m mock-model` with mocked
  Paris response in stdout within 60s.
- `--log-events` and `--log-http` use `lessflags` with `StopOnFirstArg()` so tokens after `codex`
  are codex argv unchanged; each path must end with `.jsonl` or CLI errors before mock/codex start.
- `--mock-events-preset` uses `lessflags` `StopOnFirstArg()`; tokens after `codex` are codex argv
  unchanged; combinable with `--log-events` / `--log-http`.

## Version

0.0.2

## Decision Tree

```
run-codex/
├── cli-entry/
│   ├── subcommand/                # llm-mock run codex smoke (fake codex)
│   └── shortcut-binary/           # llm-mock-run-codex smoke (fake codex)
├── codex-home/
│   ├── default-temp-home/         # fake codex prints CODEX_HOME in temp dir
│   ├── explicit-home/             # LLM_MOCK_CODEX_HOME set, config.toml written
│   └── models-cache-for-mock-model/ # orchestrator writes models_cache.json for mock-model
├── config-resolution/
│   ├── no-config-starts-codex/    # no config env → fake codex exit 0
│   └── config-file-env/           # LLM_MOCK_CONFIG_FILE + fake codex → exit 0
├── log-events/                    # --log-events run flag → mock --agent-events-file output
│   ├── appends-agent-events/      # preset think-tool-message + fake curl → think + message
│   ├── requires-jsonl-suffix/     # bad suffix → error before codex; no log file
│   └── codex-args-after-codex/    # exec -m mock-model "hi" passed through after codex token
├── log-http/                      # --log-http run flag → mock --log-http output
│   ├── appends-round-trip/        # fake codex 1 curl → log has /v1/responses + status 200
│   ├── requires-jsonl-suffix/     # bad suffix → error before codex; no log file
│   └── codex-args-after-codex/    # argv passthrough when --log-http set
├── mock-events-preset/            # --mock-events-preset run flag → mock genQueue seeding
│   ├── list-no-codex/             # run --mock-events-preset=list exits 0 without codex/mock
│   └── pass-through-think-tool-message/ # preset + fake curl → think, tool_call, message in log-events
└── integration/                   # label: real-codex, slow
    ├── headless-mock-response/    # real codex exec → Paris in stdout
    ├── headless-no-metadata-warning/ # think-tool-message preset; no mock-model metadata warning
    └── headless-tool-bash-first-turn/ # tool-bash preset; first turn shows preset-bash bash output
```

Parameter ranking (most → least significant):

1. **CLI entry** — `llm-mock run codex` vs `llm-mock-run-codex` shortcut
2. **CODEX_HOME** — default temp vs explicit `LLM_MOCK_CODEX_HOME`
3. **Config source** — no config (default empty) vs `LLM_MOCK_CONFIG_FILE`
4. **Log events output** — `--log-events` set (valid `.jsonl` vs invalid suffix) vs argv passthrough
5. **Log HTTP output** — `--log-http` set (valid `.jsonl` vs invalid suffix) vs argv passthrough
6. **Mock events preset** — `list` (catalog only) vs named preset pass-through
7. **Codex backend** — fake hook vs real codex (`label: real-codex, slow`)

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `cli-entry/subcommand` | `llm-mock run codex` smoke with fake codex → exit 0, `CODEX_HOME=` |
| 2 | `cli-entry/shortcut-binary` | `llm-mock-run-codex` smoke with fake codex → exit 0 |
| 3 | `codex-home/default-temp-home` | Fake codex prints `CODEX_HOME`; path is fresh temp dir with codex `config.toml` |
| 4 | `codex-home/explicit-home` | `LLM_MOCK_CODEX_HOME` set; `config.toml` written with `base_url` to mock |
| 4b | `codex-home/models-cache-for-mock-model` | Orchestrator writes `models_cache.json` with slug `mock-model` before codex starts |
| 5 | `config-resolution/no-config-starts-codex` | No config env vars → fake codex exit 0, `CODEX_HOME=` printed |
| 6 | `config-resolution/config-file-env` | `LLM_MOCK_CONFIG_FILE` set, fake codex → exit 0 |
| 7 | `log-events/appends-agent-events` | `--mock-events-preset=think-tool-message` + fake curl → log has `think` and `message` AgentEvents |
| 8 | `log-events/requires-jsonl-suffix` | `--log-events /tmp/events.log` → non-zero exit, `.jsonl` in error, codex not started |
| 9 | `log-events/codex-args-after-codex` | `run --log-events f.jsonl codex exec -m mock-model "hi"`; fake codex echoes argv |
| 10 | `log-http/appends-round-trip` | `--log-http <tmp>.jsonl`; fake codex one curl → log has `request.path` `/v1/responses` and `response.status` 200 |
| 11 | `log-http/requires-jsonl-suffix` | `--log-http bad.log codex` → non-zero exit, `.jsonl` in error, codex not started |
| 12 | `log-http/codex-args-after-codex` | `run --log-http f.jsonl codex exec -m mock-model "hi"`; fake codex echoes argv |
| 13 | `mock-events-preset/list-no-codex` | `llm-mock run --mock-events-preset=list` exits 0 without codex/mock |
| 14 | `mock-events-preset/pass-through-think-tool-message` | `--mock-events-preset=think-tool-message` + fake curl → think, tool_call, message in log-events |
| 15 | `integration/headless-mock-response` | Real `codex exec` headless; stdout contains Paris within 60s (`label: real-codex, slow`) |
| 16 | `integration/headless-no-metadata-warning` | Real `codex exec` with think-tool-message preset; no `Model metadata for mock-model not found` warning (`label: real-codex, slow`) |
| 17 | `integration/headless-tool-bash-first-turn` | Real `codex exec` with tool-bash preset; first turn stdout contains `preset-bash` (`label: real-codex, slow`) |

## Coverage

- **CLI entry**: subcommand and shortcut binary
- **CODEX_HOME**: default temp isolation, explicit home + generated codex `config.toml`
- **Config resolution**: optional default empty exchanges, `LLM_MOCK_CONFIG_FILE`
- **Log events**: `--log-events` wires mock `--agent-events-file` (AgentEvent JSONL on serve), `.jsonl` suffix validation, codex argv passthrough
- **Log HTTP**: `--log-http` wires mock `--log-http` (Responses API `/v1/responses` exchange JSONL), `.jsonl` suffix validation, codex argv passthrough
- **Mock events preset**: `--mock-events-preset` catalog (`list` without codex), orchestrator pass-through to mock `genQueue` (`think-tool-message`)
- **Integration**: real codex headless with mocked LLM routing proof via stdout

## How to Run

```sh
# Plumbing tests (fake codex hook) — default CI
doctest vet ./agent/llm/llm-mock/tests/run-codex
doctest test ./agent/llm/llm-mock/tests/run-codex

# Single leaf
doctest test ./agent/llm/llm-mock/tests/run-codex/config-resolution/config-file-env

# Real codex integration (requires codex on PATH)
doctest test --label real-codex ./agent/llm/llm-mock/tests/run-codex/integration/...
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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

type Request struct {
	RepoRoot         string
	BinaryPath       string // llm-mock
	ShortcutPath     string // llm-mock-run-codex
	UseShortcut      bool
	ConfigJSON       string
	ConfigEnv        string // "file" | "" (neither — error test)
	ConfigPath       string // written by Run when ConfigJSON set
	CodexHome        string // LLM_MOCK_CODEX_HOME; empty = orchestrator picks temp
	CodexArgs        []string
	WorkDir          string
	FakeCodexCmd     string // LLM_MOCK_RUN_CODEX_COMMAND
	CodexPathPrepend string // prepend to PATH (fake codex binary); skips hook when set
	LogEventsPath    string // --log-events output path; empty = flag omitted
	LogHTTPPath      string // --log-http output path; empty = flag omitted
	MockEventsPreset string // --mock-events-preset value (preset name or "list")
	ListOnly         bool   // run --mock-events-preset only, omit codex subcommand
	SkipRealCodex    bool
	ExpectedExit     int
	ExecTimeout      time.Duration
	ExpectConfig     bool // post-run: require config.toml under codex home
	ExpectParis      bool // integration: stdout must contain Paris
}

type Response struct {
	ExitCode         int
	Stdout           string
	Stderr           string
	CodexHomeUsed    string
	ConfigToml       string
	LogEventsContent string
	LogEventsLines   []string
	LogHTTPContent   string
	LogHTTPLines     []string
	Err              error
}

func Run(t *testing.T, req *Request) (*Response, error) {
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

	if req.CodexHome != "" {
		if err := os.MkdirAll(req.CodexHome, 0755); err != nil {
			return nil, fmt.Errorf("mkdir codex home: %w", err)
		}
		env = envWithOverrides(env, map[string]string{"LLM_MOCK_CODEX_HOME": req.CodexHome})
	}

	if req.CodexPathPrepend != "" {
		env = envWithOverrides(env, map[string]string{
			"PATH": req.CodexPathPrepend + string(os.PathListSeparator) + os.Getenv("PATH"),
		})
	} else if req.FakeCodexCmd != "" {
		env = envWithOverrides(env, map[string]string{"LLM_MOCK_RUN_CODEX_COMMAND": req.FakeCodexCmd})
	}

	timeout := req.ExecTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	if req.UseShortcut {
		args := append([]string{}, req.CodexArgs...)
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
			args = append(args, "codex")
			args = append(args, req.CodexArgs...)
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

	if resp.CodexHomeUsed == "" {
		resp.CodexHomeUsed = parseCodexHomeFromOutput(combined)
	}
	if resp.CodexHomeUsed == "" && req.CodexHome != "" {
		resp.CodexHomeUsed = req.CodexHome
	}

	if req.ExpectConfig || req.CodexHome != "" || strings.Contains(req.FakeCodexCmd, "CODEX_HOME") {
		home := resp.CodexHomeUsed
		if home != "" {
			data, err := os.ReadFile(filepath.Join(home, "config.toml"))
			if err == nil {
				resp.ConfigToml = string(data)
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