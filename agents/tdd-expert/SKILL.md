---
name: tdd-expert
description: >-
  Use when the user has a test-case-tree output and wants runnable Go tests
  plus stub implementation following Test-Driven Development.
---

Convert a test case tree into runnable Go tests following TDD (RED phase only).

**Input**: A path to a test case tree directory (output of
`test-case-tree-design-expert`) containing README.md, SETUP.md (with
inheritance), and ASSERT.md per leaf.

Follow these steps:

1. **Parse the Test Case Tree** — Read README.md for the overview and index,
   then walk the tree to collect every runnable leaf (directories containing
   ASSERT.md). Collect each leaf's SETUP.md and ASSERT.md plus all ancestor
   SETUP.md files.

2. **Design the Go API** — Derive package name from the feature slug. Design
   exported functions, types, and constants the tests will call.

3. **Generate Stub Implementation (RED)** — Create a Go source file where every
   exported function returns `"not implemented: <FuncName>"` error. The stub
   must compile, contain no real business logic, and return zero values for
   non-error returns.

4. **Generate Test File** — Create a Go test file. Translate SETUP.md into test
   setup code and ASSERT.md into Go assertions using standard `testing`
   patterns. Map Exit Code expectations to `exec.ExitError` checks.

5. **Initialize Module and Run** — `go mod init`, `go mod tidy`, then
   `go test ./... -v`. Verify all tests are in RED state (all fail).

6. **Report** — Report how many tests were generated, compilation status, and
   confirm RED state. If any test passes unexpectedly, adjust the stub.

**Rules**: Never write real implementation. Every stub returns an error. The
generated code must compile. All files go into a single Go module directory.
