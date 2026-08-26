# knowledgesink ship result — `git_commit_files` object contract

L2 in-process doctests for `ReadValidateShipResult` /
`ShipCommitFiles` in `github.com/xhd2015/agent-pro/pkgs/knowledgesink`.

`--create-mr` agents must emit:

```json
"git_commit_files": { "add": [], "update": [], "delete": [] }
```

Legacy flat `string[]` is rejected (no compat).

# DSN (Domain Specific Notion)

| Term | Meaning |
|------|---------|
| **add** | New hub-relative file; must exist on disk after agent |
| **update** | Modified hub-relative file; must exist on disk |
| **delete** | Removed hub-relative file; absent on disk AND still tracked in git |
| **AllPaths** | `add` then `update` then `delete` (ship `git add` pathspecs) |

**Behaviors (locked)**

1. Object shape required; JSON array → error containing `must be object`.
2. At least one path across add/update/delete.
3. Empty buckets may be omitted (`omitempty`).
4. add/update: missing path → `file missing after agent`.
5. delete still on disk → `still exists`.
6. delete not in git → `not tracked`.
7. Same path in two buckets → `both`.
8. Hub-relative only; `..` escapes rejected.

**Out of scope:** full `ShipToMR` remote push (covered by package Go tests), Marcus CLI, agent-run launch.

## Version

0.0.1

## Decision Tree

```
tests/knowledgesink-ship-result/
├── DOCTEST.md
├── SETUP.md
├── shape/                         # wire format (most user-visible break)
│   ├── SETUP.md
│   ├── object-ok/
│   └── legacy-array-rejected/
├── buckets/                       # per-op disk/git rules
│   ├── SETUP.md
│   ├── add-exists/
│   ├── update-exists/
│   ├── delete-tracked-absent/
│   ├── delete-still-present-rejected/
│   ├── delete-untracked-rejected/
│   ├── add-missing-rejected/
│   └── empty-all-rejected/
└── conflicts/
    ├── SETUP.md
    └── cross-bucket-duplicate/
```

Parameter ranking (most → least significant):

1. **Wire shape** — object vs legacy array
2. **Bucket presence rules** — add/update exist; delete absent+tracked
3. **Cross-bucket identity** — one path one op

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `shape/object-ok` | Object with update path validates |
| 2 | `shape/legacy-array-rejected` | Flat `string[]` → must be object |
| 3 | `buckets/add-exists` | New file under add OK |
| 4 | `buckets/update-exists` | Existing file under update OK |
| 5 | `buckets/delete-tracked-absent` | Tracked deletion OK |
| 6 | `buckets/delete-still-present-rejected` | delete but file still there |
| 7 | `buckets/delete-untracked-rejected` | delete invents path |
| 8 | `buckets/add-missing-rejected` | add path not on disk |
| 9 | `buckets/empty-all-rejected` | all buckets empty |
| 10 | `conflicts/cross-bucket-duplicate` | same path in add+update |

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/knowledgesink"
	"github.com/xhd2015/doctest/session"
)

// Request drives one ReadValidateShipResult call.
type Request struct {
	// HubMode: "plain" (temp dir + files) or "git" (init+commit then mutations).
	HubMode string
	// SeedFiles written before optional deletes (hub-relative → content).
	SeedFiles map[string]string
	// DeleteAfterSeed removes these paths after seed (for delete scenarios).
	DeleteAfterSeed []string
	// ResultJSON is the raw result.json body.
	ResultJSON []byte
	// ExpectOK when true, Run returns ship; else ExpectErrSubstr must match.
	ExpectOK bool
	// ExpectErrSubstr matched when ExpectOK is false.
	ExpectErrSubstr string
}

// Response observes validated ship result.
type Response struct {
	Ship *knowledgesink.ShipResult
	Err  string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	hub := t.TempDir()
	switch req.HubMode {
	case "", "plain":
		writeSeed(t, hub, req.SeedFiles)
		removePaths(t, hub, req.DeleteAfterSeed)
	case "git":
		initGitHub(t, hub, req.SeedFiles)
		removePaths(t, hub, req.DeleteAfterSeed)
	default:
		return nil, fmt.Errorf("unknown HubMode %q", req.HubMode)
	}
	resultPath := filepath.Join(t.TempDir(), "result.json")
	if err := os.WriteFile(resultPath, req.ResultJSON, 0o644); err != nil {
		return nil, err
	}
	ship, err := knowledgesink.ReadValidateShipResult(resultPath, hub)
	resp := &Response{}
	if err != nil {
		resp.Err = err.Error()
	} else {
		resp.Ship = ship
	}
	return resp, nil
}

func writeSeed(t *testing.T, hub string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		full := filepath.Join(hub, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func removePaths(t *testing.T, hub string, rels []string) {
	t.Helper()
	for _, rel := range rels {
		_ = os.Remove(filepath.Join(hub, filepath.FromSlash(rel)))
	}
}

func initGitHub(t *testing.T, hub string, files map[string]string) {
	t.Helper()
	writeSeed(t, hub, files)
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", hub}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %s (%v)", args, out, err)
		}
	}
	run("init", "-b", "master")
	run("config", "user.email", "tester@example.com")
	run("config", "user.name", "Tester")
	run("add", ".")
	run("commit", "-m", "init")
}
```
