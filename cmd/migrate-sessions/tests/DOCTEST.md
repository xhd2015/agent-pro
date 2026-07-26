# migrate-sessions Tests

Doc-style tests for the one-shot nested→flat session layout migrator intended
to be built as `/tmp/migrate-sessions` (repo path `cmd/migrate-sessions`).

# DSN (Domain Specific Notion)

`migrate-sessions` is a standalone CLI that rewrites agent-run session catalog
storage from nested `sessions/<runner>/<session_id>/` to flat
`sessions/<session_id>/` while preserving `meta.runner` as metadata.

```
migrate-sessions [--home PATH] [--dry-run] [--backup-dir PATH]

Participants:
- Operator: runs migrate-sessions against an agent-run home
- Migrator: backs up sessions/, plans moves, renames collisions, writes .layout
- Nested layout: sessions/<runner>/<id>/{meta.json,events.jsonl,…}
- Flat layout: sessions/<id>/… + sessions/.layout {"version":2}
- Collision: same id under two runners → keep newer updated_at at bare id;
  rename loser to {id}__{runner}; log renames
```

Pipeline:

```
# resolve home, backup sessions/, detect already-flat or nested
migrate-sessions --home H -> backup H/backups/sessions-<ts>/ -> plan moves

# nested unique ids
sessions/r1/a + sessions/r2/b -> sessions/a + sessions/b + .layout v2

# collision
sessions/codex/s + sessions/grok/s (newer) -> sessions/s (grok) + sessions/s__codex

# dry-run
--dry-run -> print plan; no moves; no .layout write required
```

Does **not** migrate `*-registry` or `send-queue` trees.

## Version

0.0.2

## Decision Tree

```
cmd/migrate-sessions/tests/
├── DOCTEST.md
├── SETUP.md
├── nested-unique-ids/
│   └── migrate-moves-to-flat/     unique ids across runners → flat + backup + .layout
├── collision/
│   └── keep-newer-rename-loser/   same id two runners → keep newer; rename loser
├── already-flat/
│   └── no-op-success/             .layout v2 / flat dirs → exit 0 no destructive change
└── dry-run/
    └── nested-no-moves/           --dry-run prints plan; source nested tree intact
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `nested-unique-ids/migrate-moves-to-flat` | Nested two runners unique ids → flat dirs, events preserved, backup exists, `.layout` v2 |
| 2 | `collision/keep-newer-rename-loser` | Same id under two runners → bare id is newer; loser `{id}__{runner}`; both meta intact |
| 3 | `already-flat/no-op-success` | Already flat (`.layout` v2) → exit 0; sessions unchanged |
| 4 | `dry-run/nested-no-moves` | Nested + `--dry-run` → exit 0; no moves; nested tree remains |

## How to Run

```sh
doctest vet ./cmd/migrate-sessions/tests
doctest test -v ./cmd/migrate-sessions/tests
doctest test -v ./cmd/migrate-sessions/tests/nested-unique-ids/migrate-moves-to-flat
```

```go
import (
	"bytes"
	"context"
	"encoding/json"
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
	RepoRoot   string
	TempDir    string
	Home       string
	Bin        string
	Args       []string
	Env        []string
	DryRun     bool
	BackupDir  string
	SeedMode   string // nested_unique | collision | already_flat
	ExecTimeout time.Duration
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
	Home     string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	if req.ExecTimeout <= 0 {
		req.ExecTimeout = 30 * time.Second
	}
	args := append([]string{}, req.Args...)
	if len(args) == 0 {
		args = []string{"--home", req.Home}
		if req.DryRun {
			args = append(args, "--dry-run")
		}
		if req.BackupDir != "" {
			args = append(args, "--backup-dir", req.BackupDir)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), req.ExecTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, args...)
	cmd.Env = append(os.Environ(), req.Env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Home:   req.Home,
		Err:    err,
	}
	if err == nil {
		return resp, nil
	}
	if ctx.Err() != nil {
		return resp, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		resp.ExitCode = exitErr.ExitCode()
		return resp, nil
	}
	return resp, err
}

// compile-time reference so imports stay used in harness fragments
var (
	_ = json.Marshal
	_ = filepath.Join
	_ = strings.Contains
	_ = fmt.Sprintf
	_ = time.Second
)
```
