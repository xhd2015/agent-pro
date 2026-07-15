# gen-commit-msg Tests

Doc-style tests for `github.com/xhd2015/agent-pro/agent/commit_msg`, covering
AI commit message generation with `fake-opencode`, the `--commit` git commit path,
and post-parse sanitize of commit-message anti-patterns.

# DSN (Domain Specific Notion)

`gen-commit-msg` stages a diff, calls an opencode-compatible agent runner to
produce a JSON commit message, optionally runs `git commit`. Tests use
`fake-opencode` via `--agent-runner-binary` and `FAKE_OPENCODE_MOCK_CONFIG` to
emit deterministic `step_start` / `text` / `step_finish` events. Race tests
spawn a background `git status` loop during the agent phase to reproduce
`index.lock` contention at commit time. The `--no-verify` flag forwards to
`git commit --no-verify` when used with `--commit`; alone it must fail before
the agent runs.

After the agent text is parsed into a commit message, a **post-parse sanitize**
step strips real-world anti-patterns (outer backticks, markdown title meta,
leaked `git commit -m` wrappers, tool noise such as todowrite completions) before
stdout and before `--commit`. Unusable garbage hard-fails with non-zero status
and must not create a commit. Shared fixtures live under
`agent/commit_msg/testdata/anti_patterns/` (`.in` / `.want` / `.want_err`).

## Version

0.0.2

## Decision Tree

```
agent/commit_msg/tests/
├── DOCTEST.md
├── SETUP.md
├── auto-unstage/
│   └── subdir-repo-root-paths/ nested --dir + repo-root staged paths → auto unstage succeeds
├── commit-with-fake-opencode/
│   └── succeeds/              fake-opencode returns JSON → message printed, no --commit
├── commit-race/
│   ├── background-git-loop/   agent spawns background git loop → --commit must succeed
│   └── worktree-concurrent-git/
│                               git worktree + background git loop → --commit must succeed
├── no-verify/
│   ├── requires-commit/       --no-verify without --commit → early error, no agent
│   └── passes-to-git/         failing pre-commit hook + --commit --no-verify → commit succeeds
└── sanitize/                  post-parse anti-pattern cleanup (classic TDD RED)
    ├── accepted/              dirty/clean agent text → usable sanitized message
    │   ├── dirty-json-title-backticks/  JSON title outer `...` → clean; --commit subject clean
    │   ├── md-title-meta/               **Title (N chars):** `...` → clean title
    │   ├── git-commit-m-wrapper/        `git commit -m "…"` → inner title
    │   ├── git-commit-m-double/         multi -m → title + body paragraphs
    │   ├── raw-json-subject/            bare JSON object → formatted message
    │   ├── clean-passthrough/           clean JSON unchanged
    │   └── inner-backticks-preserved/   legitimate inner ` code spans kept
    ├── rejected/
    │   └── todowrite-garbage/           tool noise → hard fail; HEAD unchanged with --commit
    └── fixtures-corpus/                 walk testdata/anti_patterns full table
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `auto-unstage/subdir-repo-root-paths` | Nested `--dir` with repo-root staged paths must not fail auto unstage |
| 2 | `commit-with-fake-opencode/succeeds` | fake-opencode mock returns commit JSON; generation succeeds |
| 3 | `commit-race/background-git-loop` | Background git status loop during agent run must not break `--commit` |
| 4 | `commit-race/worktree-concurrent-git` | Same race reproduced from a linked git worktree |
| 5 | `no-verify/requires-commit` | `--no-verify` without `--commit` errors before agent |
| 6 | `no-verify/passes-to-git` | `--commit --no-verify` skips failing pre-commit hook |
| 7 | `sanitize/accepted/dirty-json-title-backticks` | Outer-backticked JSON title cleaned; git subject clean with `--commit` |
| 8 | `sanitize/accepted/md-title-meta` | Markdown `**Title (N chars):**` meta stripped |
| 9 | `sanitize/accepted/git-commit-m-wrapper` | Single `git commit -m` wrapper → inner title |
| 10 | `sanitize/accepted/git-commit-m-double` | Multi `-m` → title + `\n\n`-joined body |
| 11 | `sanitize/accepted/raw-json-subject` | Raw JSON object → formatted title/description |
| 12 | `sanitize/accepted/clean-passthrough` | Clean JSON passes sanitize unchanged |
| 13 | `sanitize/accepted/inner-backticks-preserved` | Inner code-span backticks preserved |
| 14 | `sanitize/rejected/todowrite-garbage` | Tool noise hard-fails; no commit |
| 15 | `sanitize/fixtures-corpus` | Full `testdata/anti_patterns` table (all `.in` / `.want` / `.want_err`) |

## How to Run

```sh
doctest vet ./agent/commit_msg/tests
doctest test -v ./agent/commit_msg/tests
doctest test -v ./agent/commit_msg/tests/sanitize
doctest test -v ./agent/commit_msg/tests/sanitize/accepted/dirty-json-title-backticks
doctest test -v ./agent/commit_msg/tests/commit-race/background-git-loop
doctest test -v ./agent/commit_msg/tests/no-verify/requires-commit
doctest test -v ./agent/commit_msg/tests/no-verify/passes-to-git
```

```go
import "testing"

type Request struct {
	RepoRoot          string
	TempDir           string
	GitDir            string
	WorktreeDir       string
	FakeOpencode      string
	MockConfigPath    string
	Model             string
	Commit            bool
	NoVerify          bool
	Operation         string
	AgentRunner       string
	AgentRunnerBinary string
}

type Response struct {
	Stdout   string
	Stderr   string
	Message  string
	Err      error
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	return captureRunGenCommitMsg(t, req)
}
```
