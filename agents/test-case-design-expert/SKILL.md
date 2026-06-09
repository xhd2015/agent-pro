---
name: test-case-design-expert
description: >-
  Use when the user wants comprehensive user-facing end-to-end test scenarios
  for a feature description.
---

Design user-facing end-to-end test cases for a feature. Follow these steps:

1. **Brainstorm** — Expand the feature considering core functionality, user
   roles, integration points, error states, and edge conditions.

2. **Clarify** — If details are ambiguous, batch all questions and call
   `add-pending-questions` with JSON arguments. Each argument has a `question`
   field and optional `options` array. After calling it, stop and wait for
   answers. Only ask questions that affect test design.

3. **Design Test Cases** — For each test case, include:
   - **Scenario**: what the user does (the user-facing story)
   - **Steps**: concrete, reproducible steps
   - **Expected**: what should happen

   Cover happy path, common edge cases, error states, and boundary conditions.
   Be specific and concrete.

Output to a markdown file with an auto-derived UPPER_CASE slug name (e.g.
`FEATURE-TEST-DESIGN.md`). The report must include the original feature
request, all clarifications and answers, the expanded feature plan, and all
test cases.
