# opencode config add-provider Doctests

End-to-end doctests for the `agent-pro opencode config add-provider` leaf
subcommand. This command writes a **v1-format** opencode provider entry under
the top-level `provider` (singular) key of the opencode config file, mirroring
what opencode's `connect` custom-provider flow writes to disk.

```
agent-pro opencode config add-provider \
  --id <id> \
  --base-url <url> \
  --api-shape anthropic|openai \
  --model <m1> [--model <m2> ...] \
  [--name <display-name>] \
  [--dir <project-dir>]
```

The written provider entry looks like:

```jsonc
{
  "provider": {
    "<id>": {
      "npm":    "<@ai-sdk/anthropic | @ai-sdk/openai-compatible>",
      "name":   "<name or id>",
      "options": { "baseURL": "<base-url>" },
      "models":  { "<model-id>": { "name": "<model-id>" }, ... }
    }
  }
}
```

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — the user/shell invoking `agent-pro opencode config add-provider`
  with flags `--id`, `--base-url`, `--api-shape`, `--model` (repeatable),
  `--name` (optional), `--dir` (optional).
- **agent-pro CLI** — hand-rolled dispatch in `cmd/agent-pro/main.go`
  (`handleOpenCode` → `handleOpenCodeConfig` → new `add-provider` leaf). Parses
  flags with `github.com/xhd2015/less-gen/flags`.
- **opencode config layer** — `agent/opencode/config/config.go`: `ReadDir`
  reads `opencode.jsonc` then `opencode.json` (empty `Data` if none);
  `(*Config).Write()` does atomic temp+rename with 2-space JSON and creates
  the dir if missing.
- **Config file** — either global `$HOME/.config/opencode/opencode.json` (the
  default target) or project-local `<--dir>/.opencode/opencode.json`.
- **Test harness** — builds the `agent-pro` binary once, runs it with `HOME`
  pointed at an isolated `t.TempDir()` so the global config path lands in temp.

**Behaviors**

- Flag parsing: `--id`, `--base-url`, `--api-shape` are mandatory strings;
  `--model` is a repeatable string slice with at least one value required;
  `--name` is optional (defaults to `--id`); `--dir` is optional (default:
  global target). `-h,--help` prints usage.
- api-shape → npm mapping: `anthropic` → `@ai-sdk/anthropic`,
  `openai` → `@ai-sdk/openai-compatible`; any other value is an error.
- Target resolution: with `--dir <d>`, `opencodeDir = <d>/.opencode`;
  otherwise `opencodeDir = $HOME/.config/opencode`.
- On success the command writes the v1 provider entry under
  `Data["provider"][id]`, preserving all other existing top-level keys, then
  prints a human-readable confirmation line mentioning the provider id; exit
  code 0.
- Duplicate provider id (`provider[id]` already present) → error mentioning
  the id, file left byte-for-byte unchanged.
- Missing/invalid mandatory flags → error mentioning the offending flag (or
  the valid api-shape values); non-zero exit; no file written.

## Version

0.0.2

## Decision Tree

```
[opencode config add-provider]
 |
 +-- success/
 |    |
 |    +-- global-default/              (LEAF) no --dir: writes $HOME/.config/opencode;
 |    |                                       anthropic shape, name==id, single model (RED)
 |    +-- with-dir-local/              (LEAF) --dir <proj>: writes <dir>/.opencode (RED)
 |    +-- openai-shape/                (LEAF) --api-shape openai -> @ai-sdk/openai-compatible (RED)
 |    +-- name-flag/                   (LEAF) --name given -> provider.name == given (RED)
 |    +-- multiple-models/             (LEAF) two --model -> models map has both (RED)
 |
 +-- preserves-existing-config/        (LEAF) other provider + permission key kept (RED)
 |
 +-- errors/
      |
      +-- missing-id/                  (LEAF) no --id -> error mentions --id (RED)
      +-- missing-base-url/            (LEAF) no --base-url -> error mentions --base-url (RED)
      +-- missing-api-shape/           (LEAF) no --api-shape -> error mentions --api-shape (RED)
      +-- invalid-api-shape/           (LEAF) --api-shape gemini -> error mentions valid values (RED)
      +-- missing-model/               (LEAF) no --model -> error mentions model (RED)
      +-- duplicate-provider-id/       (LEAF) pre-existing id -> error, file unchanged (RED)
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `success/global-default` | no `--dir`: global path, anthropic npm, name==id, one model (RED) |
| 2 | `success/with-dir-local` | `--dir <proj>`: writes `<dir>/.opencode/opencode.json` (RED) |
| 3 | `success/openai-shape` | `--api-shape openai` → `npm` == `@ai-sdk/openai-compatible` (RED) |
| 4 | `success/name-flag` | `--name <display>`: provider `name` == display, not id (RED) |
| 5 | `success/multiple-models` | two `--model` flags: `models` map has both, each `{name: id}` (RED) |
| 6 | `preserves-existing-config` | pre-existing provider + `permission` key survive the write (RED) |
| 7 | `errors/missing-id` | missing `--id` → non-zero exit, stderr mentions `--id` (RED) |
| 8 | `errors/missing-base-url` | missing `--base-url` → non-zero exit, stderr mentions `--base-url` (RED) |
| 9 | `errors/missing-api-shape` | missing `--api-shape` → non-zero exit, stderr mentions `--api-shape` (RED) |
| 10 | `errors/invalid-api-shape` | `--api-shape gemini` → non-zero exit, stderr mentions valid values (RED) |
| 11 | `errors/missing-model` | no `--model` → non-zero exit, stderr mentions `model` (RED) |
| 12 | `errors/duplicate-provider-id` | pre-existing id → non-zero exit, file unchanged (RED) |

## How to Run

```sh
doctest vet ./tests/opencode-config-add-provider
doctest test ./tests/opencode-config-add-provider/...
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
	"sync"
	"testing"
	"time"

	opencodecfg "github.com/xhd2015/agent-pro/agent/opencode/config"
)

// Request describes one add-provider invocation. Leaves populate Args (the
// full `opencode config add-provider ...` arg list) and may set PreConfig
// (file content to write before running) plus WorkDir (cwd for the command,
// used when --dir is relative).
type Request struct {
	Bin       string   // path to the built agent-pro binary (set by Setup)
	Args      []string // full args after the binary, e.g. ["opencode","config","add-provider",...]
	Home      string   // isolated HOME dir (set by Setup); global config lands here
	WorkDir   string   // cwd for the command; "" means a neutral temp dir
	ProjectDir string  // optional project dir for --dir leaves ( informational )
	PreConfig string   // if non-empty, written to the resolved config path before running
	PreConfigPath string // if set, PreConfig is written here; else derived from --dir/global
	Snapshot   []byte   // optional bytes captured by a leaf Setup (e.g. before-file content) for later comparison

	// resolved after Run for leaf assertions
	ConfigPath string // the config file path the command should write to
}

// Response captures the command's observable outcome plus the resolved
// config file path so leaves can read the file back.
type Response struct {
	ExitCode   int
	Stdout     string
	Stderr     string
	Err        error
	ConfigPath string // resolved config file path (global or project-local)
	Home       string // the isolated HOME used for this run
}

var (
	builtBinOnce sync.Once
	builtBinPath string
	builtBinErr  error
)

// buildAgentPro builds ./cmd/agent-pro from the repo root once per process
// and caches the resulting binary path. The repo root is located by walking
// up from DOCTEST_ROOT (falling back to the working directory) for a dir
// containing both go.mod and cmd/agent-pro — robust to doctest's cache-mapped
// execution where DOCTEST_ROOT points at a per-test directory.
//
// cmd/agent-pro transitively imports the `frontend` package, whose embed.go
// does `//go:embed dist`. That dist/ is a frontend build artifact that is not
// checked in (it is gitignored) and may be absent in a fresh checkout /
// cache-mapped copy. The config subcommand under test never serves the
// frontend, so a stub frontend/dist/index.html is sufficient to satisfy the
// compiler. buildAgentPro creates one if dist/ has no embeddable file. The
// stub is intentionally left in place (it is gitignored) so that parallel
// test processes do not race on create/remove during their builds.
func buildAgentPro(t *testing.T) (string, error) {
	t.Helper()
	builtBinOnce.Do(func() {
		repoRoot, err := findModuleRoot()
		if err != nil {
			builtBinErr = err
			return
		}
		distDir := filepath.Join(repoRoot, "frontend", "dist")
		if err := ensureStubDist(distDir); err != nil {
			builtBinErr = fmt.Errorf("ensure frontend/dist stub: %w", err)
			return
		}
		tmp, err := os.MkdirTemp("", "agent-pro-doctest-*")
		if err != nil {
			builtBinErr = err
			return
		}
		binPath := filepath.Join(tmp, "agent-pro")
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-buildvcs=false", "-o", binPath, "./cmd/agent-pro")
		cmd.Dir = repoRoot
		var be bytes.Buffer
		cmd.Stderr = &be
		if err := cmd.Run(); err != nil {
			builtBinErr = fmt.Errorf("go build ./cmd/agent-pro: %w\n%s", err, be.String())
			return
		}
		builtBinPath = binPath
	})
	return builtBinPath, builtBinErr
}

// ensureStubDist makes sure distDir contains at least one embeddable (non-hidden)
// file so `//go:embed dist` compiles. It is idempotent: if distDir already has
// an embeddable file, it does nothing. Otherwise it creates distDir (if needed)
// and writes a stub index.html. Concurrent calls from parallel test processes
// are safe — all writers produce identical content.
func ensureStubDist(distDir string) error {
	entries, statErr := os.ReadDir(distDir)
	if statErr == nil {
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			return nil // an embeddable file already exists
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(distDir, "index.html"), []byte("stub\n"), 0o644)
}

// findModuleRoot walks up from DOCTEST_ROOT (then the working directory) until
// it finds a directory containing both go.mod and cmd/agent-pro. This works
// whether doctest runs tests in-place or in a cache-mapped copy of the repo.
func findModuleRoot() (string, error) {
	start := os.Getenv("DOCTEST_ROOT")
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = wd
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "cmd", "agent-pro")); err == nil {
				return dir, nil
			}
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("could not find module root (go.mod + cmd/agent-pro) above %s", start)
		}
	}
}

// resolveConfigPath derives the config file path the command will target,
// mirroring the command's own resolution: --dir <d> -> <d>/.opencode/opencode.json,
// otherwise <home>/.config/opencode/opencode.json.
func resolveConfigPath(home string, args []string) string {
	dir := ""
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--dir" {
			dir = args[i+1]
		}
	}
	var opencodeDir string
	if dir != "" {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(home, dir)
		}
		opencodeDir = filepath.Join(dir, ".opencode")
	} else {
		opencodeDir = filepath.Join(home, ".config", "opencode")
	}
	return filepath.Join(opencodeDir, "opencode.json")
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.Bin == "" {
		t.Fatal("req.Bin not set; root Setup must build agent-pro")
	}

	// Resolve which config path the command targets, for both pre-writing
	// fixtures and for the response (so leaves can read it back).
	configPath := req.PreConfigPath
	if configPath == "" {
		configPath = resolveConfigPath(req.Home, req.Args)
	}
	if req.PreConfig != "" {
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			return nil, fmt.Errorf("mkdir for pre-config: %w", err)
		}
		if err := os.WriteFile(configPath, []byte(req.PreConfig), 0o644); err != nil {
			return nil, fmt.Errorf("write pre-config: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Bin, req.Args...)
	// Neutral cwd by default so a relative --dir resolves under the isolated
	// HOME, not the real repo. Leaves may override with req.WorkDir.
	cmd.Dir = req.WorkDir
	if cmd.Dir == "" {
		cmd.Dir = req.Home
	}
	cmd.Env = append(os.Environ(), "HOME="+req.Home)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	resp := &Response{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		Err:        err,
		ConfigPath: configPath,
		Home:       req.Home,
	}
	if err != nil {
		if ctx.Err() != nil {
			return resp, ctx.Err()
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			resp.ExitCode = exitErr.ExitCode()
		}
	}
	return resp, nil
}

// readProviderEntry reads the config file at path (tolerating JSONC via
// opencodecfg.ReadDir) and returns the provider[id] entry plus the full Data.
// Returns an error if the file is absent or unparseable.
func readProviderEntry(path, id string) (map[string]interface{}, map[string]interface{}, error) {
	opencodeDir := filepath.Dir(path)
	cfg, err := opencodecfg.ReadDir(opencodeDir)
	if err != nil {
		return nil, nil, err
	}
	provAny, ok := cfg.Data["provider"]
	if !ok {
		return nil, cfg.Data, fmt.Errorf("config has no top-level provider key")
	}
	prov, ok := provAny.(map[string]interface{})
	if !ok {
		return nil, cfg.Data, fmt.Errorf("provider is not a map")
	}
	entry, ok := prov[id]
	if !ok {
		return nil, cfg.Data, fmt.Errorf("provider[%s] missing", id)
	}
	em, ok := entry.(map[string]interface{})
	if !ok {
		return nil, cfg.Data, fmt.Errorf("provider[%s] is not a map", id)
	}
	return em, cfg.Data, nil
}

// mustExist reads the file at path and fails the test if it is missing.
func mustExist(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected config file at %s: %v", path, err)
	}
	return b
}

// stderrContains is a small helper for error leaves that assert a substring
// appears in stderr (case-insensitive) without using the output DSL.
func stderrContains(resp *Response, want string) bool {
	return strings.Contains(strings.ToLower(resp.Stderr), strings.ToLower(want))
}
```
