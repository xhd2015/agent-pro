---
name: test-case-tree-design-expert
description: >-
  Use when the user wants a decision-tree style test case design with inherited
  setup and per-leaf assertions.
---

Model test cases as a decision tree with inherited setup and per-leaf
assertions.

**Mode Selection**: Check the input. If terse or vague, follow all steps
(1-5). If already comprehensive with explicit decisions, skip to steps 3-5.

1. **Brainstorm** — Identify all modes, flags, subcommands, decision points,
   error states, and happy-path/sad-path outcomes.

2. **Clarify** — If details are ambiguous, call `add-pending-questions` with
   JSON arguments (each with a `question` and optional `options` array). Stop
   and wait after calling.

3. **Propose Outline** — Present a flat list of proposed test cases (name +
   one-line description). Call `add-pending-questions` to get user approval.

4. **Build the Decision Tree** — Structure as root (feature) → modes (top-level
   branches) → decision nodes → leaves (runnable test cases).

5. **Produce Output** — Create the directory structure:
   ```
   <feature-slug>-test-cases/
   ├── README.md       # mermaid graph + text tree + test case index
   ├── SETUP.md        # Global setup (inherited)
   ├── mode-xxx/
   │   ├── SETUP.md    # Inherits + mode-specific setup
   │   └── leaf-a/
   │       ├── SETUP.md
   │       └── ASSERT.md
   ...
   ```
   **SETUP.md**: Inherited along ancestor chain. Format: Preconditions, Steps,
   Context sections, plus optional executable Go code at the end.

   **ASSERT.md**: Per-leaf assertions. Format: Expected, Side Effects, Errors,
   Exit Code sections, plus mandatory executable Go Assert function at the end.

   **README.md**: DOT graph, ASCII text tree, test case index table.

   Every runnable leaf must have both SETUP.md and ASSERT.md. After writing,
   validate with `validate_test_case_tree <output-dir>`.
