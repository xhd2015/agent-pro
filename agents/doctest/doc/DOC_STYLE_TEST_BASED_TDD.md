---
name: doc-style-test-based-tdd
description: adversarial two-agent TDD with doc-style tests
---

This document defines the workflow for an **adversarial two-agent TDD system**
built on doc-style tests(see `doctest skill doc-spec show`) with
[executable Go code](see `doctest skill code-spec show`).

The main agent writes and seals the tests. The sub-agent implements code to
make them pass. The two agents are adversarial: the main agent verifies that
the sub-agent never modifies sealed tests without justification.

## Role: Main Agent

The main agent is the **test writer and orchestrator**. It does not write
implementation code. Its responsibilities:

1. Elaborate requirements with the user
2. Design a comprehensive doctest tree
3. Run tests to confirm they fail (RED)
4. Seal the tests to prevent arbitrary modification
5. Hand off implementation to the sub-agent
6. Handle questions from the sub-agent during implementation
7. On completion: verify test integrity and confirm tests pass (GREEN)

## Prerequisites

The main agent must understand:

- **`doctest skill doc-spec show`** —
  how to structure test cases as markdown decision trees with `SETUP.md` and
  `ASSERT.md`
- **`doctest skill code-spec show`** —
  how to embed executable Go code in `SETUP.md` and `ASSERT.md`, including
  function signatures for `Setup`, `Run`, and `Assert`

## Workflow

### Phase 1: Requirements Elaboration

Discuss the feature with the user. Produce a design document that covers:

- What the feature does (purpose, inputs, outputs)
- Every flag, option, or parameter
- Every subcommand or operation mode
- Expected outcomes for common scenarios
- Error conditions and how they surface

Get explicit user approval before proceeding to test design.

### Phase 2: Test Design

Build a doctest tree following the doc-style test specification:

1. Propose a flat list of test cases (name + one-line description) and get
   approval before creating directories.
2. Create the directory structure with `SETUP.md` and `ASSERT.md` files.
3. Embed Go code blocks in each file per the code specification:
   - Root `SETUP.md`: define `Request` and `Response` types, and a default
     `func Run` (a stub returning `"error not implemented"` is acceptable at
     this stage, since the test must fail).
   - Child `SETUP.md`: `func Setup` populating `req` fields.
   - Leaf `ASSERT.md`: `func Assert` checking expected outcomes.

The tests must be **comprehensive**: cover happy paths, error paths, edge
cases, and input variants. Prefer more leaves over fewer.

### Phase 3: RED — Confirm Tests Fail

Run the tests to confirm every leaf is in a failing state:

```sh
doctest test -v ./tests/<test-for-this-feature>
```

Expected output: all tests fail with errors matching `"error not implemented"`
or an equivalent stub failure. If any test passes unexpectedly, re-examine
the test design — a passing test at this stage means the test is not testing
anything meaningful (no implementation exists yet).

### Phase 4: Seal Tests

Once all tests are confirmed failing, seal them to prevent the sub-agent from
arbitrarily modifying test cases:

```sh
git add ./tests/<test-for-this-feature>
```

This stages the test directory. The sub-agent may still read the tests, but
any modification to them will appear as an unstaged diff that the main agent
can detect in the verification phase.

### Phase 5: Handoff to Sub-Agent

Invoke the sub-agent with the design document and test overview:

```sh
doctest agent implement "<design doc + test summary>"
```

The prompt should include:

- A concise summary of the feature and its expected behaviour
- The test tree structure and what each leaf covers
- The fact that tests are sealed (staged) and must not be modified

### Phase 6: Handle Sub-Agent Questions

The sub-agent may encounter ambiguity during implementation, and it would return questions to stdout and wait for your followup.

When this happens:

1. Read the questions from the output.
2. Attempt to resolve based on the design document and test expectations.
3. If the design document and tests are sufficient to answer, provide the
   answer directly.
4. If the question requires domain knowledge or user preference, escalate to
   the user for confirmation.

Once answers are ready, feed them back by re-invoking the sub-agent:

```sh
# invoke with followup
doctest agent implement "<answers to questions>"
```

The CLI sends the message as a followup on the same thread. The sub-agent
resumes its context and continues implementation.

This re-invoke loop may repeat multiple times until the sub-agent reports
completion (all tests passing) with no further questions.

Do not guess about business logic or user intent. When in doubt, ask the user.

### Phase 7: Verify Completion

When the sub-agent reports completion:

**Step 1 — Check test integrity:**

```sh
git diff ./tests/<test-for-this-feature>
```

No unstaged changes should exist in the test directory. If modifications are
found, evaluate each one:

- **Unavoidable/necessary** — e.g. the sub-agent found a legitimate bug in a
  test assertion (the test expected wrong behaviour based on the spec). Accept
  only with explicit justification.
- **Unjustified** — reject the change and require the sub-agent to fix the
  implementation to match the original tests.

**Step 2 — Run tests:**

```sh
doctest test -v ./tests
```

All tests must pass (GREEN). If any test fails, feed the failure output back
to the sub-agent for correction. Repeat until all tests pass.

**Step 3 — Report:**

Summarize the results to the user: how many tests passed, any test
modifications accepted (with rationale).