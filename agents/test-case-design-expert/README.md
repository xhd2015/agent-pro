# test-case-design-expert

Takes a feature description and designs user-facing e2e test cases. The LLM brainstorms, asks clarifying questions interactively via `ask_user`, and produces a complete test plan.

## Run

```bash
# Fast tests (no LLM)
go run ./test/integrations/ -short

# Full tests (includes LLM calls)
go run ./test/integrations/

# Run the expert directly
go run . "a user dashboard that shows analytics"
```

## Architecture

```
main.go (orchestrator)
 ├─ Copies itself as ask_user → temp dir
 ├─ Prepend temp dir to PATH
 ├─ Answer goroutine: polls question file, reads stdin, writes answer
 └─ Calls LLM (opencode via run.Run)
      ├─ LLM brainstorms → expands plan
      ├─ LLM calls ask_user "…?" when info missing
      │    ├─ ask_user writes question → question.txt
      │    ├─ goroutine prints to user, reads stdin
      │    ├─ goroutine writes answer → answer.txt
      │    └─ ask_user reads answer → stdout → LLM continues
      └─ LLM produces test cases → stdout
```

## File Structure

```
├── main.go              orchestrator + ask_user dispatch
├── PROMPT.md            LLM system prompt (__FEATURE__ template)
├── ask_user/
│   ├── main.go          thin wrapper: calls ask_user.Run()
│   └── lib/lib.go       shared Run() logic (os.Args[0] dispatch)
└── test/integrations/   e2e test suite (go test output format)
```

## Design Principles

### Unified Binary (os.Args[0] dispatch)
A single binary serves both roles. The orchestrator copies itself as `ask_user` to a temp dir. When invoked as `ask_user`, the binary runs the question/answer protocol instead of the orchestrator. No `go build` at runtime.

### File-Based Communication
`QUESTION_FILE` and `ANSWER_FILE` env vars point to temp files. The ask_user process writes a question and polls for an answer. The orchestrator goroutine detects new questions, prompts the user, and writes answers. This avoids IPC complexity.

### Embed Prompt
`PROMPT.md` is embedded at compile time via `//go:embed`. The `__FEATURE__` placeholder is replaced with the user's input at runtime.
