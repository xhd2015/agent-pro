# DSH Web browser runner

`agent-run run --agent-runner dsh-web` hands a task to an already-running DeepSeek Harness Web server through `dsh web open`. The `dsh` executable must be on `PATH` and support `--prompt`, `--prompt-file`, and `--no-submit`. The runner does not start a server.

```sh
# Open an editable draft without starting an agent turn.
agent-run run --agent-runner dsh-web --dir /path/to/project \
  --open --no-submit '/brainstorm Implement the task'

# Submit a long prompt from a file and open the session.
agent-run run --agent-runner dsh-web --dir /path/to/project \
  --open --prompt-file /path/to/task.txt

# Submit without opening a browser.
agent-run run --agent-runner dsh-web --dir /path/to/project 'Implement the task'
```

## Options and ownership

The browser runner accepts a positional prompt or `--prompt-file FILE`, plus `--dir`, `--open`, `--no-submit`, and `-v`/`--verbose`. Inline and file prompts are mutually exclusive. `--open` opens the browser; omitting it forwards `--no-open`. `--no-submit` saves a draft and is also supported without `--open`. The default workspace is the caller's current directory.

Prompt text is passed as one argument without shell evaluation or trimming. File paths are resolved against the caller's working directory and forwarded as `--prompt-file`; contents never expand into the child command line. DSH reads and validates the file, including UTF-8 and empty-input checks, before creating its session. Agent-run does not modify or delete the file.

DSH owns the session identity, transcript, draft, prompt admission, browser URL, and subsequent execution. Agent-run waits only for the launch command to finish and forwards its stdout/stderr. Success means the session handoff succeeded, not that the task completed. Launch failures return an error.

This is a CLI launch adapter, not a PTY provider or streaming `Ask` runner. Agent-run session IDs, resume/send/auto-send-or-resume, JSON event streaming, model overrides, custom runner binaries/config homes, terminal flags, color, and child environment flags are unsupported and rejected on `run`. Use DSH itself for session management and server configuration. It is not a runner for the agent-run Web API or the session-oriented `agentrunapi.AutoSendOrResume` API.

## wrk

With a compatible wrk installation:

```sh
wrk project -t 'Implement the task' --open-in-agent --agent-runner dsh-web
```

wrk supplies `/brainstorm ${task}` and the browser defaults `--open --no-submit`, without `--color` or `--session-id-from-prompt`. Long tasks remain file-backed through the complete handoff.

## Tests

```sh
go test ./pkgs/agentruncli -run TestDSHWeb -count=1
```

The CLI test uses a temporary `dsh` executable to capture exact arguments and simulate launch failure. It makes no browser, server, or model calls and checks that no agent-run session storage is created. Actual draft rendering and prompt admission are owned by DSH's integration tests.
