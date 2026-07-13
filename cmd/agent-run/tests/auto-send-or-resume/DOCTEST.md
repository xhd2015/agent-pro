# agent-run run --auto-send-or-resume + resume workspace + --new-terminal

Doc-style tests for `agent-run run --auto-send-or-resume --session-id <id>`
branching (run / send / resume), resume workspace resolution
(`meta.workspace` vs `--dir` vs process cwd), Grok resume `--dir` vs session
`info.cwd` relocate gate (`--allow-relocate-resume-session-dir`), and
`--new-terminal` (iTerm2 ModeForceNew launcher that strips the flag and
re-invokes `run` in a new window). Prefer stub fixtures (seeded meta, registry,
fake ptywrap with inject APIs, argv/cwd-recording fake runner, fake Grok home
under `--agent-runner-config-home`, `KOOL_ITERM2_*` script capture) over real
grok / real iTerm2 UI.

# DSN (Domain Specific Notion)

**Participants**

- **agent-run CLI (`run`)** — with `--auto-send-or-resume` classifies a stable
  `--session-id` into MODE=run | send | resume and dispatches existing
  run/send/resume semantics. Without the flag, `run` is unchanged.
  Optional `--new-terminal` is valid only with `--auto-send-or-resume`.
- **agent-run CLI (`resume`)** — original subcommand; workspace must prefer
  `meta.workspace` when `--dir` is unset (same helper as auto→resume).
- **Session storage** — `AGENT_RUN_HOME/sessions/<session_id>/meta.json` (flat;
  runner is a field on `SessionMeta`: session_id, runner, runner_session_id,
  terminal_session_id, status, **workspace**, model, …). Missing session is a
  valid auto path (→ run).
- **TTY registry** — `AGENT_RUN_HOME/grok-tty-registry/<terminal_session_id>.json`
  (`pid`, `listen_addr`). Live send and status layers depend on it.
- **ptywrap / terminal** — TCP + WebSocket scrollback + HTTP inject
  (`prepare-inject`, `input`). Tests use an in-process fake.
- **Provider runner** — bound when `meta.runner_session_id` is set; **exited**
  true when terminal dead/missing or zombie keep-alive (exit scrollback) while
  bound. Resume injects `provider --resume <runner_session_id>`.
- **Fake runner** — `--agent-runner-binary` argv/cwd recorder; must not set
  `AGENT_RUN_GROK_TTY_COMMAND` on argv-sensitive leaves (hook hides `--resume`).
- **iTerm2 launcher** — `github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2`
  `OpenConfig` with `ModeForceNew`. Tests set `KOOL_ITERM2_INSTALLED=1` and
  `KOOL_ITERM2_SCRIPT_OUT=<file>` so AppleScript is written without osascript UI.

**Behaviors**

```
# validation
run --auto-send-or-resume (no --session-id)
  -> exit 1; error requires --session-id
run --auto-send-or-resume --session-id X --session-id-from-prompt …
  -> exit 1; mutually exclusive
run --new-terminal … (without --auto-send-or-resume)
  -> exit 1; error requires --auto-send-or-resume
run -h / run --help
  -> documents --auto-send-or-resume and --new-terminal; stdout ends with \n

# decision (after resolve session by id; missing OK)
probeSessionStatus(meta):
  Resume.Ready (bound && exited) -> MODE=resume
  live (exited==false)           -> MODE=send
  else (missing / unbound / …)   -> MODE=run

# MODE=run
run --auto-send-or-resume --session-id NEW "prompt" (+ argv recorder)
  -> agentui.Run creates session; argv has prompt; NO --resume
  -> meta.workspace set from --dir or create-time cwd

# MODE=send (live)
+ non-empty prompt -> send to meta.terminal_session_id (fallback session_id);
                     stdout msg_N; honor --no-submit; NO provider --resume spawn
+ empty prompt     -> stderr warning (live / no message); exit 0; no enqueue
+ --open           -> accepted; open ignored while live (still send); exit 0

# MODE=resume (bound+exited, including zombie keep-alive)
+ followup         -> resume path; argv has --resume <runner_session_id>
+ empty prompt     -> keep-tty reopen OK (exit 0 with fake runner)
+ --open/--no-submit -> accepted on resume path (not live-open error)

# --new-terminal (requires --auto-send-or-resume)
run --auto-send-or-resume --new-terminal --session-id ID [flags…] [prompt]
  MODE=send  -> ignore --new-terminal (optional stderr note); existing send path;
                no iTerm script; enqueue/deliver still works
  MODE=run | resume ->
    1. record parsed flags (less-flags CollectParsedFlags)
    2. strip --new-terminal; reconstruct run-level argv
    3. argv = ["run"] + reconstructed + (prompt != "" ? ["--", prompt] : [])
    4. iterm2.OpenConfig(workspaceDir, ModeForceNew, shell-quoted full command)
    5. launcher exits 0 WITHOUT in-process provider spawn
  workspaceDir: --dir if set, else meta.workspace (resume), else process cwd
  ForceNew script: always "create window" (no session scan / create tab)

# workspace (resume + auto→resume)
1) --dir if set (exists, is dir) wins
2) else non-empty meta.workspace:
   - if path exists and is a directory → child cwd = meta.workspace
   - if path missing / not a dir → exit 1; clear session-workspace error naming
     the path + hint to pass --dir <existing-directory>
     (do NOT silently fall back to process cwd; do NOT present as missing
     agent-run binary only / fork/exec without workspace context)
3) else process cwd + stderr warning
Child provider cwd must be that workspace (not CLI process cwd when different).

# resume --dir vs Grok info.cwd (grok-tty only; resume + auto→resume)
When --dir is explicitly set and runner is grok-tty:
  resolve Grok home = --agent-runner-config-home | meta.agent_runner_config_home | default
  read sessions/**/<meta.runner_session_id>/summary.json → info.cwd
  path equality: Abs + EvalSymlinks + treat /private prefix variants as equal
  --dir == grok cwd     → OK; no RelocateCWD
  --dir ≠ grok cwd, no --allow-relocate-resume-session-dir
                        → exit ≠0; stderr mismatch + flag name; session NOT moved
  --dir ≠ grok cwd + --allow-relocate-resume-session-dir
                        → stderr warning; sessions.RelocateCWD(runnerSessionID, dir,
                          &RelocateCWDOptions{GrokHome}); update meta.workspace to --dir;
                          continue resume
  --dir unset / MODE=run / non-grok-tty → no Grok cwd check
Flag exact name: --allow-relocate-resume-session-dir (on run + resume help)
Grok layout fixture: $GROK_HOME/sessions/<url.PathEscape(abs_cwd)>/<id>/summary.json
  (inactive: no active_sessions entry for id)
```


## Version

0.0.2

## Decision Tree

```
cmd/agent-run/tests/auto-send-or-resume/
├── DOCTEST.md
├── SETUP.md
├── validation/                              # flag / arg gates (pre-branch)
│   ├── missing-session-id/                  # A1 no --session-id → exit 1
│   ├── session-id-from-prompt-mutex/        # A2 mutex with --session-id-from-prompt
│   └── run-help-lists-flag/                 # A3 run -h documents --auto-send-or-resume
├── run-mode/                                # MODE=run (missing / unbound / else)
│   └── missing-session-creates/             # B1 new id + prompt → create, no --resume
├── send-mode/                               # MODE=send (live exited=false)
│   ├── with-prompt/                         # C1 enqueue/deliver msg_N
│   ├── empty-prompt-warns/                  # C2 exit 0 + stderr warning
│   └── open-rejected/                       # C3 --open while live → exit 1
├── resume-mode/                             # MODE=resume (bound+exited)
│   ├── followup-argv/                       # D1 argv has --resume <id>
│   └── empty-prompt-reopen/                 # D2 empty followup reopen exit 0
├── workspace/                               # resume + auto→resume cwd
│   ├── resume-meta-workspace/               # E1 resume uses meta.workspace
│   ├── auto-resume-meta-workspace/          # E2 auto→resume uses meta.workspace
│   ├── dir-override/                        # E3 --dir wins over meta.workspace
│   ├── missing-meta-workspace/              # E4 resume: missing meta.workspace → exit 1 + --dir hint
│   ├── auto-missing-meta-workspace/         # E5 auto→resume: missing meta.workspace → exit 1 + --dir hint
│   └── resume-dir-relocate/                 # E6+ --dir vs Grok info.cwd (classic RED)
│       ├── dir-matches-grok-cwd/            # E6 --dir == grok cwd → OK; no move
│       ├── dir-mismatch-errors/             # E7 resume --dir ≠ grok cwd → exit 1; no move
│       ├── dir-mismatch-allow-relocates/    # E8 allow flag → RelocateCWD + meta.workspace
│       ├── help-lists-flag/                 # E9 run -h + resume -h list allow flag
│       └── auto-dir-mismatch-errors/        # E10 auto→resume same mismatch error as E7
└── new-terminal/                            # --new-terminal + iTerm2 ModeForceNew
    ├── requires-auto-flag/                  # NT-V1 without --auto-send-or-resume → exit 1
    ├── run-help-lists-flag/                 # NT-V2 run -h documents --new-terminal
    ├── mode-run-opens-iterm/                # NT-D1 MODE=run → ForceNew script; no in-process spawn
    ├── mode-resume-opens-iterm/             # NT-D2 MODE=resume → ForceNew script; no parent resume spawn
    ├── mode-send-ignores/                   # NT-D3 MODE=send → no script; still msg_N
    └── prompt-dash-separator/               # NT-D4 prompt starts with "-" → `--` before prompt
```

Parameter ranking (most → least significant):

1. **Invocation gate** — validation (missing session-id / mutex / help) vs runtime auto
2. **Session lifecycle mode** — run (missing) | send (live) | resume (exited)
3. **`--new-terminal` dispatch** — gate (requires auto) | ForceNew launcher (run/resume) | ignore (send)
4. **Prompt / open flags within mode** — empty vs non-empty; dash-leading prompt; `--open` on live vs resume
5. **Workspace source** — meta.workspace | auto path | `--dir` override | missing meta.workspace error
6. **Grok cwd relocate gate** — `--dir` vs summary `info.cwd`; allow flag; RelocateCWD
7. **Fixtures** — argv probe, cwd probe, fake Grok home, fake ptywrap inject, KOOL_ITERM2 script out

## Test Index

| # | Leaf | Req | Description |
|---|------|-----|-------------|
| 1 | `validation/missing-session-id` | A1 | auto without session-id → exit 1; requires session-id |
| 2 | `validation/session-id-from-prompt-mutex` | A2 | auto + `--session-id-from-prompt` → exit 1 mutual exclusion |
| 3 | `validation/run-help-lists-flag` | A3 | `run -h` documents `--auto-send-or-resume`; trailing `\n` |
| 4 | `run-mode/missing-session-creates` | B1 | new session-id + prompt → run create; no `--resume`; meta.workspace set |
| 5 | `send-mode/with-prompt` | C1 | live + prompt → exit 0; stdout `msg_N`; inject/enqueue; no `--resume` |
| 6 | `send-mode/empty-prompt-warns` | C2 | live + empty prompt → exit 0; stderr warning; no msg_N |
| 7 | `send-mode/open-rejected` | C3 | live + `--open` → exit 1; open not allowed while live |
| 8 | `resume-mode/followup-argv` | D1 | exited + followup → argv `--resume <runner_session_id>` |
| 9 | `resume-mode/empty-prompt-reopen` | D2 | exited + empty prompt → reopen exit 0; `--resume` in argv |
| 10 | `workspace/resume-meta-workspace` | E1 | `resume` child cwd = meta.workspace (CLI cwd differs) |
| 11 | `workspace/auto-resume-meta-workspace` | E2 | auto→resume child cwd = meta.workspace |
| 12 | `workspace/dir-override` | E3 | resume `--dir` override wins; child cwd = override |
| 13 | `workspace/missing-meta-workspace` | E4 | `resume` without `--dir` when meta.workspace missing → exit 1; path + workspace wording + `--dir` hint |
| 14 | `workspace/auto-missing-meta-workspace` | E5 | auto→resume same missing-workspace contract as E4 |
| 15 | `workspace/resume-dir-relocate/dir-matches-grok-cwd` | E6 | resume `--dir` == Grok `info.cwd` → exit 0; session not moved |
| 16 | `workspace/resume-dir-relocate/dir-mismatch-errors` | E7 | resume `--dir` ≠ Grok cwd, no allow → exit 1; flag hint; session unmoved |
| 17 | `workspace/resume-dir-relocate/dir-mismatch-allow-relocates` | E8 | allow flag → warning; RelocateCWD; meta.workspace=`--dir`; resume continues |
| 18 | `workspace/resume-dir-relocate/help-lists-flag` | E9 | `run -h` and `resume -h` list `--allow-relocate-resume-session-dir` |
| 19 | `workspace/resume-dir-relocate/auto-dir-mismatch-errors` | E10 | auto→resume same mismatch error as E7 |
| 20 | `new-terminal/requires-auto-flag` | NT-V1 | `--new-terminal` without auto → exit 1; requires `--auto-send-or-resume` |
| 21 | `new-terminal/run-help-lists-flag` | NT-V2 | `run -h` documents `--new-terminal`; trailing `\n` |
| 22 | `new-terminal/mode-run-opens-iterm` | NT-D1 | auto+new-terminal missing session → exit 0; ForceNew script; no `--new-terminal` in follow-up; no in-process provider |
| 23 | `new-terminal/mode-resume-opens-iterm` | NT-D2 | auto+new-terminal bound+exited → exit 0; script; no parent argv probe spawn |
| 24 | `new-terminal/mode-send-ignores` | NT-D3 | auto+new-terminal live → exit 0; no iTerm script; still `msg_N` |
| 25 | `new-terminal/prompt-dash-separator` | NT-D4 | prompt like `-v explain` → follow-up has `--` then that prompt |

## How to Run

```sh
doctest vet ./cmd/agent-run/tests/auto-send-or-resume
doctest test ./cmd/agent-run/tests/auto-send-or-resume

# validation
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/validation/missing-session-id
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/validation/session-id-from-prompt-mutex
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/validation/run-help-lists-flag

# run / send / resume modes
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/run-mode/missing-session-creates
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/send-mode/with-prompt
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/send-mode/empty-prompt-warns
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/send-mode/open-rejected
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/resume-mode/followup-argv
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/resume-mode/empty-prompt-reopen

# workspace
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/workspace/resume-meta-workspace
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/workspace/auto-resume-meta-workspace
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/workspace/dir-override
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/workspace/missing-meta-workspace
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/workspace/auto-missing-meta-workspace

# resume --dir vs Grok info.cwd (relocate gate; classic RED until implementer)
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/workspace/resume-dir-relocate
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/workspace/resume-dir-relocate/dir-matches-grok-cwd
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/workspace/resume-dir-relocate/dir-mismatch-errors
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/workspace/resume-dir-relocate/dir-mismatch-allow-relocates
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/workspace/resume-dir-relocate/help-lists-flag
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/workspace/resume-dir-relocate/auto-dir-mismatch-errors

# --new-terminal (iTerm2 ModeForceNew)
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/new-terminal
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/new-terminal/requires-auto-flag
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/new-terminal/run-help-lists-flag
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/new-terminal/mode-run-opens-iterm
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/new-terminal/mode-resume-opens-iterm
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/new-terminal/mode-send-ignores
doctest test -v ./cmd/agent-run/tests/auto-send-or-resume/new-terminal/prompt-dash-separator
```

```go
import (
	"testing"
	"time"
)

type Request struct {
	RepoRoot string
	TempDir  string
	// WorkDir is the CLI process cwd (cmd.Dir). May differ from meta.workspace.
	WorkDir  string
	Home     string
	AgentRun string
	Args     []string
	Env      []string

	// Mode selects post-exec enrichment in Run:
	//   ""            — plain CLI exec
	//   "read-meta"   — load sessions/<session>/meta.json into MetaAfter
	//   "read-probes" — read ArgvProbePath / CwdProbePath into response fields
	Mode        string
	ExecTimeout time.Duration

	// Session seed (meta.json under flat sessions/<SessionID>/)
	SeedMeta          bool
	Runner            string // default grok-tty
	SessionID         string
	MetaStatus        string // running | finished | error
	RunnerSessionID   string // provider (grok) session id — resume key
	TerminalSessionID string
	Workspace         string // meta.workspace seed / expected create workspace
	Model             string
	InitialPrompt     string

	// Registry / process / terminal layers
	WriteRegistry         bool
	RegistryPID           int  // 0 = os.Getpid() (alive); >0 use as-is; -1 = dead sentinel
	RegistryClosedPort    bool // listen on closed port (unreachable)
	StartFakePTYWrap      bool
	FakePTYWrapPort       int
	FakePTYWrapScrollback string
	// FakePTYInjectLog records HTTP inject bodies (for send-mode asserts).
	FakePTYInjectLog *[]string

	// Open / hooks
	OpenInstantAttach bool
	GrokTTYCommand    string

	// Argv / cwd probes (no AGENT_RUN_GROK_TTY_COMMAND — use binary path)
	ArgvProbePath     string
	CwdProbePath      string
	RunnerScriptPath  string
	AgentRunnerBinary string
	FollowupPrompt    string
	DirOverride       string // --dir value when set in leaf Args

	// Grok home / resume-dir-relocate fixtures (workspace/resume-dir-relocate/*).
	// GrokHome is passed as --agent-runner-config-home. Seed sessions under
	// sessions/<url.PathEscape(abs_cwd)>/<RunnerSessionID>/summary.json.
	GrokHome       string
	GrokSessionCwd string // seeded summary info.cwd (Abs)
	GrokSessionDir string // pre-call session directory path

	// iTerm2 test hooks (--new-terminal leaves). When set, Run/Setup env gets
	// KOOL_ITERM2_INSTALLED=1 and KOOL_ITERM2_SCRIPT_OUT=<path>.
	ItermScriptOut string
}

type RegistryEntryData struct {
	SessionID  string `json:"session_id"`
	ListenAddr string `json:"listen_addr"`
	PID        int    `json:"pid"`
	CreatedAt  string `json:"created_at"`
}

type Response struct {
	Stdout    string
	Stderr    string
	ExitCode  int
	Err       error
	MetaAfter map[string]any
	ArgvProbe string
	CwdProbe  string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return runAgentRun(t, req, req.Args...)
}
```
