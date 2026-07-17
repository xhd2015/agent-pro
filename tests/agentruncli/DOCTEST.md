# agentruncli — extract agent-run CLI into importable library

Classic TDD doctests for plan phase **P1**: move essentially all of
`cmd/agent-run` (today `package main`) into importable package

```text
github.com/xhd2015/agent-pro/pkgs/agentruncli
```

with a **single public entrypoint** pinned by this tree:

```go
func Handle(args []string) error
```

`cmd/agent-run/main.go` becomes a thin wrapper that calls
`agentruncli.Handle(os.Args[1:])` and maps errors to stderr + exit 1.

**Zero intentional behavior change** for existing agent-run CLI users.
Existing `cmd/agent-run/tests/**` remain the behavior regression suite after
implementer; this tree only locks the **extract contract** (API + thin main +
help/unknown smoke).

**Out of scope:** spl, local-bot, agent-run-helper, tools/agent-run, full e2e
duplication of every subcommand.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — process entry (`cmd/agent-run` or any in-process importer) passes
  CLI argv **without** the program name (`os.Args[1:]`).
- **`pkgs/agentruncli`** — library package (not `package main`) that owns all
  agent-run CLI dispatch formerly in `cmd/agent-run`. Public surface for P1 is
  only **`Handle(args []string) error`**.
- **`Handle`** — parses top-level flags/subcommands and runs the same command
  tree as today: web, run, resume, attach, send, msg, snapshot, watch,
  sessions, status, tty, pty; top-level help; `--agent-runner`; internal
  serve/stub paths. Returns `nil` on success (including help). Returns a
  non-nil error for unknown commands and other failures (main prints
  `agent-run: …` to stderr and exits 1).
- **`cmd/agent-run`** — thin `package main`: import agentruncli, call
  `Handle(os.Args[1:])`, map error → stderr + `os.Exit(1)`. No business logic
  left in main beyond that glue.
- **Stdout / stderr** — help and command output continue to go to process
  stdout/stderr as today (`fmt.Print` help, etc.). Unit leaves capture by
  temporarily redirecting `os.Stdout` / `os.Stderr` around `Handle` (mutex so
  parallel leaves do not race).
- **Existing CLI trees** — `cmd/agent-run/tests/**` and `tests/agentrunapi`
  are regression, not redefined here.

**Behaviors**

```
# API surface
import pkgs/agentruncli
  -> package is library (not package main)
  -> Handle(args []string) error is public and callable

# CLI wire
cmd/agent-run/*.go (production, non-test)
  -> import github.com/xhd2015/agent-pro/pkgs/agentruncli
  -> call agentruncli.Handle(...)

# Handle smoke (in-process; no separate agent-run binary required)
Handle([]string{"--help"}) or Handle([]string{"-h"})
  -> err == nil
  -> stdout lists Usage and commands including web, run, sessions, status
     (and documents --agent-runner)

Handle([]string{"not-a-real-command"})
  -> err != nil (unknown command)

# Out of P1 leaf scope (covered by existing cmd/agent-run tests)
Handle(["run"|"web"|"sessions"|…]) full e2e / TTY / web
```

## Version

0.0.2

## Decision Tree

```
tests/agentruncli/
├── DOCTEST.md
├── SETUP.md
├── api-surface/                         # package + public Handle + not main
│   ├── SETUP.md
│   ├── handle-exists/                   # Handle is importable and callable
│   └── not-package-main/                # pkgs/agentruncli is not package main
├── source-wire/                         # thin cmd/agent-run cutover
│   ├── SETUP.md
│   └── cli-thin-main/                   # imports agentruncli and calls Handle
└── handle/                              # in-process Handle smoke
    ├── SETUP.md
    ├── help-lists-commands/             # --help → stdout usage tokens
    └── unknown-command/                 # unknown subcommand → error
```

Parameter ranking (most → least significant):

1. **Concern** — API surface | CLI source-wire | Handle runtime smoke
2. **API shape** (api-surface) — Handle callable vs package-not-main
3. **Handle outcome** (handle) — help success vs unknown-command error
4. **Smoke independence** — no real agent-run binary, TTY, web, or network

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `api-surface/handle-exists` | Package imports; `Handle` callable (e.g. empty/`--help` may succeed or return API error only after export exists) |
| 2 | `api-surface/not-package-main` | Sources under `pkgs/agentruncli` declare `package agentruncli` (not `main`) |
| 3 | `source-wire/cli-thin-main` | At least one `cmd/agent-run` production `.go` imports `pkgs/agentruncli` and references `Handle` |
| 4 | `handle/help-lists-commands` | `Handle([]string{"--help"})` → nil error; stdout contains Usage, web, run, sessions, status, `--agent-runner` |
| 5 | `handle/unknown-command` | `Handle([]string{"not-a-real-command"})` → non-nil error mentioning unknown command |

## How to Run

```sh
# from agent-pro module root (external/agent-pro-master-2026-07-16)
doctest vet ./tests/agentruncli
doctest test ./tests/agentruncli

doctest test -v ./tests/agentruncli/api-surface/handle-exists
doctest test -v ./tests/agentruncli/api-surface/not-package-main
doctest test -v ./tests/agentruncli/source-wire/cli-thin-main
doctest test -v ./tests/agentruncli/handle/help-lists-commands
doctest test -v ./tests/agentruncli/handle/unknown-command
```

Expect **RED** until implementer creates `pkgs/agentruncli` with `Handle` and
thins `cmd/agent-run`. After GREEN on this tree, smoke-check existing
`cmd/agent-run/tests/help` + `auto-send-or-resume` and `tests/agentrunapi`.

### Planned public API (RED until implementer)

```go
package agentruncli

// Handle is the full agent-run CLI entrypoint for args after the program name
// (os.Args[1:]). Same command surface as today's cmd/agent-run run():
// web, run, resume, attach, send, msg, snapshot, watch, sessions, status,
// tty, pty, top-level help, --agent-runner, and internal serve/stub paths.
// Success (including help) returns nil; failures return an error for the
// thin main to print and exit 1.
func Handle(args []string) error
```

Thin main (illustrative; implementer):

```go
package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
)

func main() {
	if err := agentruncli.Handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "agent-run: %v\n", err)
		os.Exit(1)
	}
}
```

```go
import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentruncli"
)

// Request drives one leaf via Mode. Leaves set Mode (and Args for handle/*).
type Request struct {
	// Mode: api_surface | not_package_main | source_wire | handle
	Mode string
	// Args passed to agentruncli.Handle when Mode == "handle"
	// (and optionally for api_surface Handle touch).
	Args []string
}

// Response is the harness observation after package call or source scan.
type Response struct {
	// Handle path
	Stdout    string
	Stderr    string
	ErrString string // Handle error text; empty if Handle returned nil

	// api_surface
	HandleCalled bool

	// not_package_main / source_wire
	ScannedFiles int
	PackageName  string // first non-test package clause under pkgs/agentruncli
	ImportFound  bool
	HandleRef    bool // agentruncli.Handle or .Handle call site in cmd sources
	PkgDir       string
	CmdDir       string
}

// stdoutMu serializes temporary os.Stdout/os.Stderr redirects across parallel leaves.
var stdoutMu sync.Mutex

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	switch req.Mode {
	case "api_surface":
		return runAPISurface(t, req, resp)
	case "not_package_main":
		return runNotPackageMain(t, req, resp)
	case "source_wire":
		return runSourceWire(t, req, resp)
	case "handle":
		return runHandle(t, req, resp)
	default:
		return nil, fmt.Errorf("unknown mode %q", req.Mode)
	}
}

func runAPISurface(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	// Touch public entry so the leaf fails to compile until Handle exists.
	args := req.Args
	if args == nil {
		args = []string{"--help"}
	}
	stdout, stderr, err := callHandleCaptured(args)
	resp.Stdout = stdout
	resp.Stderr = stderr
	resp.HandleCalled = true
	if err != nil {
		resp.ErrString = err.Error()
	}
	return resp, nil
}

func runNotPackageMain(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	// DOCTEST_ROOT is tests/agentruncli; module root is ../..
	pkgDir := filepath.Join(DOCTEST_ROOT, "..", "..", "pkgs", "agentruncli")
	resp.PkgDir = pkgDir
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		// Package dir missing is a RED outcome: report empty package, no harness error
		// so Assert can fail clearly (or treat as error — prefer error for missing path).
		return resp, fmt.Errorf("read pkgs/agentruncli: %w", err)
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(pkgDir, e.Name())
		resp.ScannedFiles++
		f, err := parser.ParseFile(fset, path, nil, parser.PackageClauseOnly)
		if err != nil {
			// Fallback: package line scan
			data, _ := os.ReadFile(path)
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "package ") {
					resp.PackageName = strings.TrimSpace(strings.TrimPrefix(line, "package "))
					resp.PackageName = strings.TrimSuffix(resp.PackageName, "\r")
					return resp, nil
				}
			}
			continue
		}
		if f.Name != nil {
			resp.PackageName = f.Name.Name
			return resp, nil
		}
	}
	return resp, nil
}

func runSourceWire(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	cmdDir := filepath.Join(DOCTEST_ROOT, "..", "..", "cmd", "agent-run")
	resp.CmdDir = cmdDir
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return nil, fmt.Errorf("read cmd/agent-run: %w", err)
	}
	importPath := "github.com/xhd2015/agent-pro/pkgs/agentruncli"
	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(cmdDir, e.Name())
		resp.ScannedFiles++
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		src := string(data)
		hasImport := false
		f, err := parser.ParseFile(fset, path, data, parser.ImportsOnly)
		if err == nil {
			for _, imp := range f.Imports {
				p := strings.Trim(imp.Path.Value, `"`)
				if p == importPath {
					hasImport = true
					break
				}
			}
		}
		if !hasImport {
			if strings.Contains(src, importPath) || strings.Contains(src, "pkgs/agentruncli") {
				hasImport = true
			}
		}
		hasHandle := strings.Contains(src, "agentruncli.Handle") ||
			// thin main may use dotted call after import alias; accept Handle( after package import
			(hasImport && strings.Contains(src, "Handle("))
		if hasImport {
			resp.ImportFound = true
		}
		if hasHandle {
			resp.HandleRef = true
		}
		if resp.ImportFound && resp.HandleRef {
			return resp, nil
		}
	}
	return resp, nil
}

func runHandle(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	args := req.Args
	if args == nil {
		args = []string{}
	}
	stdout, stderr, err := callHandleCaptured(args)
	resp.Stdout = stdout
	resp.Stderr = stderr
	resp.HandleCalled = true
	if err != nil {
		resp.ErrString = err.Error()
	}
	return resp, nil
}

func callHandleCaptured(args []string) (stdout, stderr string, handleErr error) {
	stdoutMu.Lock()
	defer stdoutMu.Unlock()

	rOut, wOut, err := os.Pipe()
	if err != nil {
		return "", "", err
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		_ = rOut.Close()
		_ = wOut.Close()
		return "", "", err
	}

	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = wOut, wErr

	// Drain pipes while Handle runs to avoid deadlock if output exceeds pipe buffer.
	var outBuf, errBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&outBuf, rOut)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&errBuf, rErr)
	}()

	handleErr = agentruncli.Handle(args)

	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	wg.Wait()
	_ = rOut.Close()
	_ = rErr.Close()

	return outBuf.String(), errBuf.String(), handleErr
}
```
