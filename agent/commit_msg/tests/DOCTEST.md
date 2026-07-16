# gen-commit-msg Tests

Doc-style tests for `github.com/xhd2015/agent-pro/agent/commit_msg`, covering
AI commit message generation with `fake-opencode` and `llm-mock-run-commandcode`,
the `--commit` git commit path, post-parse sanitize of commit-message anti-patterns,
pure-plan `--dry-run`, and optional `--add-all` staging (`git add -A` before generate).

# DSN (Domain Specific Notion)

`gen-commit-msg` inspects the staged set, optionally auto-unstages binaries and
submodules, calls a supported agent runner to produce a JSON commit message
(`{"title","description"}`), optionally runs `git commit`. Supported runners:

- **`opencode`** (default): tests use `fake-opencode` via `--agent-runner-binary`
  and `FAKE_OPENCODE_MOCK_CONFIG` for deterministic NDJSON events.
- **`commandcode`**: tests use `llm-mock-run-commandcode` as
  `--agent-runner-binary` plus `LLM_MOCK_RUN_COMMANDCODE_COMMAND` hook so the
  mock prints fixed JSON on stdout without a live Command Code API. Default
  LookPath binary name is `cmd` (override via `--agent-runner-binary`).

Race tests spawn a background `git status` loop during the agent phase to
reproduce `index.lock` contention at commit time. The `--no-verify` flag
forwards to `git commit --no-verify` when used with `--commit`; alone it must
fail before the agent runs.

**`--dry-run` pure plan**: inspect staged files only; do **not** call the agent
or mutate the index/HEAD. Stdout is the fixed mock message B
(`dry-run: would generate commit message for N staged file(s)`). When binaries
or submodules would normally auto-unstage, print `would: unstage …` on stderr
instead of unstaging. With `--commit`, print `would: git commit -m '…'` on
stderr (append ` --no-verify` when that flag is set) and do not commit.
`--agent-runner` is still validated (unknown runners error even under dry-run;
supported list is `opencode, commandcode`). `--model` is accepted but unused
under dry-run.

**`--add-all`**: before the generate pipeline, stage like `git add -A` via
`gitwrite.AddAll`. Does **not** require `--commit`.

- **Real (non-dry-run)**: log `$ git add -A` on stderr (non-silent), then
  `gitwrite.AddAll(dir)`, then the existing pipeline (auto-unstage binaries →
  generate → optional `--commit`).
- **Dry-run**: print `would: git add -A` on stderr; **do not** mutate the index;
  staged-file count and empty-index checks use the **current** index only (so an
  empty index still yields `no staged changes` after the would-line).

After the agent text is parsed into a commit message, a **post-parse sanitize**
step strips real-world anti-patterns (outer backticks, markdown title meta,
leaked `git commit -m` wrappers, tool noise such as todowrite completions) before
stdout and before `--commit`. Unusable garbage hard-fails with non-zero status
and must not create a commit. Shared fixtures live under
`agent/commit_msg/testdata/anti_patterns/` (`.in` / `.want` / `.want_err`).

## Version

0.0.3

## Decision Tree

```
agent/commit_msg/tests/
├── DOCTEST.md
├── SETUP.md
├── not-a-git-repo/            non-git --dir → early error: not a git repository
├── auto-unstage/
│   └── subdir-repo-root-paths/ nested --dir + repo-root staged paths → auto unstage succeeds
├── commit-with-fake-opencode/
│   └── succeeds/              fake-opencode returns JSON → message printed, no --commit
├── commandcode/               --agent-runner=commandcode (classic TDD RED until implemented)
│   ├── help/
│   │   └── mentions-commandcode/  -h help lists commandcode as supported runner
│   ├── dry-run/
│   │   └── succeeds/          --dry-run --agent-runner commandcode → mock B; no agent
│   └── generate/
│       ├── no-commit/         mock binary + hook → title/description; HEAD unchanged
│       └── with-commit/       mock + --commit → new commit with sanitized subject
├── commit-race/
│   ├── background-git-loop/   agent spawns background git loop → --commit must succeed
│   └── worktree-concurrent-git/
│                               git worktree + background git loop → --commit must succeed
│   # pure classifier + CommitWithRetry recovery (unable to write new index file,
│   # stale lock, hook fail-fast) lives in agent/git_runner/tests/
├── no-verify/
│   ├── requires-commit/       --no-verify without --commit → early error, no agent
│   └── passes-to-git/         failing pre-commit hook + --commit --no-verify → commit succeeds
├── sanitize/                  post-parse anti-pattern cleanup
│   ├── accepted/              dirty/clean agent text → usable sanitized message
│   │   ├── dirty-json-title-backticks/  JSON title outer `...` → clean; --commit subject clean
│   │   ├── md-title-meta/               **Title (N chars):** `...` → clean title
│   │   ├── git-commit-m-wrapper/        `git commit -m "…"` → inner title
│   │   ├── git-commit-m-double/         multi -m → title + body paragraphs
│   │   ├── raw-json-subject/            bare JSON object → formatted message
│   │   ├── clean-passthrough/           clean JSON unchanged
│   │   └── inner-backticks-preserved/   legitimate inner ` code spans kept
│   ├── rejected/
│   │   └── todowrite-garbage/           tool noise → hard fail; HEAD unchanged with --commit
│   └── fixtures-corpus/                 walk testdata/anti_patterns full table
├── dry-run/                   pure-plan --dry-run
│   ├── mock-message-count/    staged N files → mock B on stdout; no agent
│   ├── no-unstage-binary/     binary+text staged → would unstage; index unchanged; N before unstage
│   ├── with-commit-no-mutate/ --dry-run --commit → would: git commit; HEAD unchanged
│   ├── with-commit-no-verify-plan/ --dry-run --commit --no-verify → would-line has --no-verify
│   ├── rejects-unknown-agent-runner/ --agent-runner codex → unsupported; supported includes commandcode
│   ├── accepts-model-no-agent/ --model set → success mock; no agent call
│   └── no-staged-errors/      empty index → no staged changes error
└── add-all/                   --add-all stages like git add -A (classic TDD RED until implemented)
    ├── dry-run-would-line/    --add-all --dry-run → would: git add -A; index unchanged
    ├── stages-untracked/      --add-all real → $ git add -A; untracked becomes staged; no --commit
    ├── with-commit/           --add-all --commit → stages then commits; HEAD advances
    └── help-mentions-add-all/ -h help text documents --add-all
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `not-a-git-repo` | Non-git `--dir` → early error containing `not a git repository` |
| 2 | `auto-unstage/subdir-repo-root-paths` | Nested `--dir` with repo-root staged paths must not fail auto unstage |
| 3 | `commit-with-fake-opencode/succeeds` | fake-opencode mock returns commit JSON; generation succeeds |
| 4 | `commandcode/help/mentions-commandcode` | `-h` help text mentions `commandcode` as a supported agent runner |
| 5 | `commandcode/dry-run/succeeds` | `--dry-run --agent-runner commandcode` succeeds with mock B; no agent |
| 6 | `commandcode/generate/no-commit` | commandcode mock binary returns JSON; message printed; HEAD unchanged |
| 7 | `commandcode/generate/with-commit` | commandcode mock + `--commit` creates commit with sanitized subject |
| 8 | `commit-race/background-git-loop` | Background git status loop during agent run must not break `--commit` |
| 9 | `commit-race/worktree-concurrent-git` | Same race reproduced from a linked git worktree |
| 10 | `no-verify/requires-commit` | `--no-verify` without `--commit` errors before agent |
| 11 | `no-verify/passes-to-git` | `--commit --no-verify` skips failing pre-commit hook |
| 12 | `sanitize/accepted/dirty-json-title-backticks` | Outer-backticked JSON title cleaned; git subject clean with `--commit` |
| 13 | `sanitize/accepted/md-title-meta` | Markdown `**Title (N chars):**` meta stripped |
| 14 | `sanitize/accepted/git-commit-m-wrapper` | Single `git commit -m` wrapper → inner title |
| 15 | `sanitize/accepted/git-commit-m-double` | Multi `-m` → title + `\n\n`-joined body |
| 16 | `sanitize/accepted/raw-json-subject` | Raw JSON object → formatted title/description |
| 17 | `sanitize/accepted/clean-passthrough` | Clean JSON passes sanitize unchanged |
| 18 | `sanitize/accepted/inner-backticks-preserved` | Inner code-span backticks preserved |
| 19 | `sanitize/rejected/todowrite-garbage` | Tool noise hard-fails; no commit |
| 20 | `sanitize/fixtures-corpus` | Full `testdata/anti_patterns` table (all `.in` / `.want` / `.want_err`) |
| 21 | `dry-run/mock-message-count` | `--dry-run` prints exact mock B with correct staged file count; no agent |
| 22 | `dry-run/no-unstage-binary` | Dry-run plans binary unstage on stderr; binary stays staged; N includes it |
| 23 | `dry-run/with-commit-no-mutate` | `--dry-run --commit` would-line only; HEAD subject unchanged |
| 24 | `dry-run/with-commit-no-verify-plan` | Would commit line includes `--no-verify`; HEAD unchanged |
| 25 | `dry-run/rejects-unknown-agent-runner` | `--dry-run --agent-runner codex` → unsupported; supported list includes `commandcode` |
| 26 | `dry-run/accepts-model-no-agent` | `--dry-run --model` accepted; mock success without agent |
| 27 | `dry-run/no-staged-errors` | Dry-run with empty index → no staged changes error |
| 28 | `add-all/dry-run-would-line` | `--add-all --dry-run` → `would: git add -A`; index unchanged (honest empty-index error OK) |
| 29 | `add-all/stages-untracked` | Real `--add-all` logs `$ git add -A`, stages untracked; HEAD unchanged without `--commit` |
| 30 | `add-all/with-commit` | `--add-all --commit` stages untracked then creates commit with mock subject |
| 31 | `add-all/help-mentions-add-all` | `-h` help text mentions `--add-all` |

## How to Run

```sh
doctest vet ./agent/commit_msg/tests
doctest test -v ./agent/commit_msg/tests
doctest test -v ./agent/commit_msg/tests/not-a-git-repo
doctest test -v ./agent/commit_msg/tests/commandcode
doctest test -v ./agent/commit_msg/tests/dry-run
doctest test -v ./agent/commit_msg/tests/add-all
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
	DryRun            bool
	// AddAll requests --add-all (stage like git add -A before generate).
	AddAll            bool
	Help              bool
	Operation         string
	AgentRunner       string
	AgentRunnerBinary string
	// MockCommandCode is the path to the llm-mock-run-commandcode binary.
	MockCommandCode string
	// CommandCodeHook is the shell snippet for LLM_MOCK_RUN_COMMANDCODE_COMMAND
	// (deterministic stdout from the mock without a live Command Code API).
	CommandCodeHook string
	// GenCommitMsgBin is the path to cmd/gen-commit-msg (help subprocess).
	GenCommitMsgBin string
	// HEADSubjectBefore records git subject before generate (HEAD-unchanged asserts).
	HEADSubjectBefore string
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
