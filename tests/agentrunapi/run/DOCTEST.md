# agentrunapi — Run + RunJSON

Classic TDD doctests for **Run** (start + wait until turn-done) and
**RunJSON** (schema example + temp result file). Nested under
`tests/agentrunapi/` so P1 classify/auto leaves stay compile-isolated.

**No real agent-run binary, iTerm, or grok.** Leaves inject `Launch` /
`Wait` / `OpenFn` / `SoftExit`.

# DSN

**Participants**

- **`Run`** — prompt + dir + options → launch (detach or open-terminal) →
  wait until runner exit / resume-ready / idle-after-work / `ResultFile`
  valid JSON.
- **`RunJSON`** — wraps `Run`; generates an absolute temp result path
  (not under WorkspaceDir), appends schema example, reads the file after
  wait.
- **Caller** supplies `Store` / `StoreHome`. Empty both → store package
  default. No invented `{out}/agent-run-home`.
- **Lifetime** — detach: `KeepAliveDetached` (default false → `/exit`).
  open: `ExitOnFinishTerminal` (default false → keep iTerm).

**Behaviors**

```
Run(OpenTerminal=false) -> detach launch; default /exit after wait
Run(OpenTerminal=true)  -> follow-up --open; AGENT_RUN_HOME= only if Store/StoreHome set
Timeout 0               -> 30m
RunJSON                 -> prompt contains schema + abs temp path; return file JSON
ResultFile valid JSON   -> done even if TTY still busy
```

## Version

0.0.1

## Decision Tree

```
tests/agentrunapi/run/
├── DOCTEST.md
├── SETUP.md
├── run/
│   ├── SETUP.md
│   ├── detach-default/
│   ├── open-terminal/
│   ├── timeout-default/
│   ├── exit-on-finish-detach/
│   ├── keep-alive-detached/
│   ├── exit-on-finish-terminal/
│   └── missing-prompt/
└── run-json/
    ├── SETUP.md
    ├── appends-schema/
    ├── reads-file/
    ├── missing-file/
    ├── invalid-json/
    └── file-is-done/
```

## How to Run

```sh
doctest vet ./tests/agentrunapi/run
doctest test ./tests/agentrunapi/run
```

```go
import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

type Request struct {
	// Mode: run | run_json
	Mode string

	Prompt                 string
	WorkspaceDir           string
	SchemaExample          string
	OpenTerminal           bool
	KeepAliveDetached      bool
	ExitOnFinishTerminal   bool
	Timeout                time.Duration
	PollInterval           time.Duration
	StoreHome              string
	SessionID              string
	UseZeroTimeout         bool // leave Timeout at 0 so product default applies

	InstallLaunch      bool
	InstallWait        bool
	LaunchWritesResult bool
	WaitWriteJSON      string
	WaitWriteRaw       string
	SkipWaitWrite      bool
}

type Response struct {
	ErrString string
	JSON      string
	SessionID string

	LaunchCalls        int
	LaunchOpenTerminal bool
	LaunchTimeout      time.Duration
	LaunchPrompt       string
	LaunchResultFile   string
	LaunchWorkspace    string

	OpenCalls    int
	OpenDir      string
	OpenFollowUp string

	SoftExitCalls int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}
	if req.WorkspaceDir == "" {
		req.WorkspaceDir = t.TempDir()
	}

	opts := agentrunapi.RunOpts{
		Prompt:               req.Prompt,
		WorkspaceDir:         req.WorkspaceDir,
		OpenTerminal:         req.OpenTerminal,
		KeepAliveDetached:    req.KeepAliveDetached,
		ExitOnFinishTerminal: req.ExitOnFinishTerminal,
		Timeout:              req.Timeout,
		PollInterval:         req.PollInterval,
		StoreHome:            req.StoreHome,
		SessionID:            req.SessionID,
	}
	if req.UseZeroTimeout {
		opts.Timeout = 0
	}
	if req.Prompt == "" && req.Mode != "run" {
		// run_json leaves always send a prompt; missing-prompt is run-only.
	}

	if req.InstallLaunch || req.LaunchWritesResult {
		opts.Launch = func(ctx context.Context, o agentrunapi.RunOpts) error {
			resp.LaunchCalls++
			resp.LaunchOpenTerminal = o.OpenTerminal
			resp.LaunchTimeout = o.Timeout
			resp.LaunchPrompt = o.Prompt
			resp.LaunchResultFile = o.ResultFile
			resp.LaunchWorkspace = o.WorkspaceDir
			if req.LaunchWritesResult && o.ResultFile != "" && req.WaitWriteJSON != "" {
				return os.WriteFile(o.ResultFile, []byte(req.WaitWriteJSON), 0o644)
			}
			return nil
		}
	}
	if req.InstallWait {
		opts.Wait = func(ctx context.Context, h agentrunapi.RunHandle) error {
			path := h.ResultFile
			if path == "" {
				path = resp.LaunchResultFile
			}
			if req.SkipWaitWrite {
				return nil
			}
			if req.WaitWriteRaw != "" && path != "" {
				return os.WriteFile(path, []byte(req.WaitWriteRaw), 0o644)
			}
			if req.WaitWriteJSON != "" && path != "" {
				return os.WriteFile(path, []byte(req.WaitWriteJSON), 0o644)
			}
			return nil
		}
	}
	opts.OpenFn = func(dir, followUp string) error {
		resp.OpenCalls++
		resp.OpenDir = dir
		resp.OpenFollowUp = followUp
		return nil
	}
	opts.SoftExit = func(store agentstorage.Store, meta agentstorage.SessionMeta, runner string) {
		resp.SoftExitCalls++
	}

	switch req.Mode {
	case "run":
		res, err := agentrunapi.Run(context.Background(), opts)
		if err != nil {
			resp.ErrString = err.Error()
			return resp, nil
		}
		if res != nil {
			resp.SessionID = res.SessionID
		}
		return resp, nil
	case "run_json":
		out, err := agentrunapi.RunJSON(context.Background(), agentrunapi.RunJSONOpts{
			RunOpts:       opts,
			SchemaExample: req.SchemaExample,
		})
		if err != nil {
			resp.ErrString = err.Error()
			return resp, nil
		}
		resp.JSON = out
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected harness error: %v", err)
	}
}

func assertAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil || resp.ErrString == "" {
		t.Fatal("expected API error, got nil/empty")
	}
}

func assertNoAPIError(t *testing.T, resp *Response) {
	t.Helper()
	if resp != nil && resp.ErrString != "" {
		t.Fatalf("unexpected API error: %s", resp.ErrString)
	}
}

func assertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %#v, want %#v", field, got, want)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in %q", want, got)
	}
}

func assertNotContains(t *testing.T, got, forbidden string) {
	t.Helper()
	if strings.Contains(got, forbidden) {
		t.Fatalf("unexpected %q in %q", forbidden, got)
	}
}
```
