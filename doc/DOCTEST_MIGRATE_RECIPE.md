# Doctest migrate recipe (`d *session.Doctest`)

Per-tree checklist for hardening doctest roots so the driver no longer auto-injects free package vars. Follow **live `doctest vet` errors**, not outdated docs.

## Required signatures

```go
import "github.com/xhd2015/doctest/session"

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error)

// Prefer the same `d` parameter on Setup/Assert (matches migrated trees; unused OK).
func Setup(t *testing.T, d *session.Doctest, req *Request) error
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error)
```

- `d` is **required** on `Run`. Free package vars are **not** auto-injected.
- Setup/Assert: update when they touch paths/session fields, or when vet requires it. Pilot recipe updates **all** Setup/Assert in the tree for consistency.

## Inject field mapping

| Old (forbidden free / env) | New |
|----------------------------|-----|
| `DOCTEST_ROOT` | `d.DOCTEST_ROOT` |
| `DOCTEST_CASE` | `d.DOCTEST_CASE` |
| `DOCTEST_SESSION_ID` | `d.DOCTEST_SESSION_ID` |
| `os.Getenv("DOCTEST_ROOT")` etc. | `d` fields above |

Fixtures / case-local files:

```go
path := filepath.Join(d.DOCTEST_CASE, "fixture.json")
// or relative to tree root:
path := filepath.Join(d.DOCTEST_ROOT, "testdata", "…")
```

Prose in SETUP/ASSERT may say `d.DOCTEST_ROOT` so later readers do not reintroduce free names.

## Forbidden isolation patterns

Do **not** use for leaf isolation:

- `os.Chdir(...)`
- `os.Setenv("DOCTEST_*", ...)` (or any Setenv of inject fields)

Use `t.TempDir()`, `cmd.Dir`, and `d.DOCTEST_*` paths instead. Subprocess env for **production** knobs (e.g. `AGENT_PRO_*`) is fine; that is not inject isolation.

## Per-tree steps

1. **Locate roots** — each directory with its own `DOCTEST.md` is a separate root (including nested, e.g. `tests/agentrunapi/wait-driver/`).
2. **Update `Run`** — add `d *session.Doctest`; import `"github.com/xhd2015/doctest/session"`.
3. **Replace free injects** — `DOCTEST_*` / `os.Getenv("DOCTEST_*")` → `d.DOCTEST_*`.
4. **Update Setup/Assert** — signatures to include `d`; path helpers use `d` fields.
5. **No Chdir/Setenv** for leaf isolation.
6. **Verify one tree fully green before the next**:

```sh
doctest vet ./path/to/tree
doctest test ./path/to/tree
# nested roots separately:
doctest vet ./path/to/tree/nested-root
doctest test ./path/to/tree/nested-root
```

7. **Sanity after migrate** (no free inject left in pilot tree):

```sh
rg -n '\bDOCTEST_(ROOT|CASE|SESSION_ID)\b' ./path/to/tree --glob '*.md' | rg -v 'd\.DOCTEST_'
# expect no matches in code (prose should already say d.DOCTEST_*)
```

## Count remaining free `Run` signatures (repo baseline)

Free (not yet migrated):

```sh
rg -n 'func Run\(t \*testing\.T, req \*Request\)' --glob '**/DOCTEST.md' | wc -l
```

Already migrated:

```sh
rg -n 'func Run\(t \*testing\.T, d \*session\.Doctest, req \*Request\)' --glob '**/DOCTEST.md' | wc -l
```

List free roots:

```sh
rg -l 'func Run\(t \*testing\.T, req \*Request\)' --glob '**/DOCTEST.md'
```

## Signature notes for later phases

- **`Run`**: always `func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error)`.
- **`Setup` / `Assert`**: pilot pattern always threads `d`. Unused params are fine in Go; optional `_ = d` for clarity.
- **Session import**: required in `DOCTEST.md` (and any file that names `session.Doctest` in a compiled block). Leaf SETUP/ASSERT often omit the import; the driver resolves `session.Doctest` when the type is used in signatures (same as migrated `tests/procresolve`).
- **Nested DOCTEST roots** need their own migrate + own `doctest vet`/`test` (parent `./…` may not compile/run the nested root’s suite).
- **Source-wire path depth**: `d.DOCTEST_ROOT` is the **nested** root when the tree is nested (e.g. `wait-driver` → three `..` to module root, not two).

## Gotchas discovered in P2

- **Name collision with intermediate `session/` dirs**: generated leaves import both `"github.com/xhd2015/doctest/session"` and the intermediate package under a grouping named `session/`. Rename the grouping (e.g. `session-ops/`) before or during migrate. See `pkgs/agentstorage/tests/session-ops/`.
- **Local variable `d` shadows the inject param**: if `Run` already used `d` for duration/dir, rename the local (e.g. `dur`) after adding `d *session.Doctest`.
- **Helpers that read fixtures must take `d`**: `fixturesDir(req)` that uses free `DOCTEST_ROOT` must become `fixturesDir(d, req)` and every caller (including leaf Setup) must pass `d`.
- **Case-relative `os.ReadFile("fixture")`**: use `filepath.Join(d.DOCTEST_CASE, "fixture")`. Do not rely on process cwd.
- **`t.Setenv` / `t.Chdir` panic under doctest**: the driver runs leaves with `t.Parallel()`. Prefer `os.Setenv` + restore/`t.Cleanup`, or `cmd.Env` / `cmd.Dir` for subprocesses. Do **not** use `t.Setenv` for isolation.
- **Dual leaf+grouping**: a directory with both `ASSERT.md` and nested leaves generates two packages in one folder and fails build. Promote the intermediate assert into its own leaf (e.g. `global/`).
- **Nested Go modules** (e.g. `cmd/go.mod` → `cmd-doctest-harness`): doctest’s generated suite only replaces `agent-pro`; packages that live only in the nested module may not resolve. Prefer helpers importable from the main module, or document as product/module red after inject is correct.
- **Pre-existing `doctest vet` DSN gaps**: some older trees lack a `# DSN` section; `doctest test` can still pass. Fix DSN separately if vet is required green.

## Reference trees (already green)

- `tests/procresolve/` (+ nested `p2/`)
- `tests/proc-resolve-cli/`
- P1 pilots: `tests/agentruncli/`, `tests/agentrunapi/` (+ nested `wait-driver/`), `tests/slack-msg-agent-wire/`
- P2 L2 inject complete for listed roots (see phase report); pure conversion trees fully green.