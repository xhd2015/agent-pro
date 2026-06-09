# Validation Rules

Every SETUP.md and ASSERT.md must have exactly one Go code block at the end of the file. No more than one code block is allowed; no trailing content after the final block.

## Rules

| # | Rule | Category |
|---|---|---|
| 1 | SETUP.md must have a Go code block | SETUP block |
| 2 | Go block must be the final content in the file | Go block |
| 3 | Root SETUP.md must define `type Request` and `type Response` | Root SETUP |
| 4 | Every SETUP.md must have `func Setup` or `func Run` | SETUP |
| 5 | `func Setup` must be `func Setup(t *testing.T, req *Request) error` | Signature |
| 6 | `func Run` must be `func Run(t *testing.T, req *Request) (*Response, error)` | Signature |
| 7 | ASSERT.md must have a Go code block with `func Assert` | ASSERT |
| 8 | `func Assert` must be `func Assert(t *testing.T, req *Request, resp *Response, err error)` | Signature |
| 9 | Child SETUP.md cannot redefine `type Request` or `type Response` | Inheritance |
| 10 | At least one SETUP.md in the chain must define `func Run` | Chain |
| 11 | `func Setup` body must not be a stub (`return nil` alone) | Body |

## Testdata Fixtures

### `valid-tree/`
All rules satisfied. Root has Request, Response, Setup, Run. Leaf has Setup and Assert.
- **Expected**: compile passes.

### `missing-go-block/` (rules 1, 4)
Root SETUP.md has no Go code block (prose-only).
- **Expected**: compile fails. Error: `must have a Go code block`

### `missing-setup-and-run/` (rule 4)
Root SETUP.md has Go block with types only — no Setup, no Run.
- **Expected**: compile fails. Error: `must have func Setup or func Run`

### `missing-root-setup/` **(now valid)**
Root has Request, Response, Run — satisfies "Setup or Run" rule.
- **Expected**: compile passes.

### `missing-leaf-setup/` **(now valid)**
Leaf SETUP.md has Run — satisfies "Setup or Run" rule.
- **Expected**: compile passes.

### `missing-assert/` (rule 7)
Leaf ASSERT.md has `func Check` instead of `func Assert`.
- **Expected**: compile fails. Error: `missing func Assert`

### `missing-assert-go-block/` (rule 7)
Leaf ASSERT.md has no Go code block.
- **Expected**: compile fails. Error: `missing go block`

### `missing-run/` (rule 10)
Entire setup chain has no Run.
- **Expected**: compile fails. Error: `no Run in setup chain`

### `multiple-violations/` (rules 4, 7, 10)
Root missing Setup or Run. Leaf1 ASSERT missing Assert. Leaf2 chain missing Run.
- **Expected**: compile fails. 3 errors collected in one pass.

### `missing-request-type/` (rule 3)
Root defines Response and Run but no Request type.
- **Expected**: compile fails. Error: `must define type Request`

### `non-final-go-block/` (rule 2)
Root SETUP.md has a Go code block followed by markdown text — block is not final.
- **Expected**: compile fails. Error: `go block must be final content`

### `wrong-setup-signature/` (rule 5)
Leaf SETUP.md has Setup with wrong parameter: `func Setup(t *testing.T, req string) error`.
- **Expected**: compile fails. Error: `Setup must be func Setup(t *testing.T, req *Request) error`

### `wrong-run-signature/` (rule 6)
Root SETUP.md has Run with wrong result type: `func Run(...) error` (missing *Response).
- **Expected**: compile fails. Error: `Run must be func Run(t *testing.T, req *Request) (*Response, error)`

### `wrong-assert-signature/` (rule 8)
Leaf ASSERT.md has Assert with wrong params: `func Assert(t *testing.T)`.
- **Expected**: compile fails. Error: `Assert must be func Assert(t *testing.T, req *Request, resp *Response, err error)`

### `child-redefines-request/` (rule 9)
Leaf SETUP.md redefines `type Request` — not allowed for child SETUP.md.
- **Expected**: compile fails. Error: `cannot redefine Request`

### `setup-stub-body/` (rule 11)
Leaf SETUP.md has Setup with body `{ return nil }` — no actual logic.
- **Expected**: compile fails. Error: `func Setup body must not be a stub`
