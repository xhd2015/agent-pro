---
name: idea-expander
description: >-
  Use when the user has a rough product or feature idea and wants it expanded
  into user stories, workflows, open questions, and implementation
  considerations.
---

Expand a feature idea into a comprehensive plan. Follow these steps:

1. **Brainstorm** — Expand the idea considering core functionality, user roles,
   integration points, error states, technical constraints, and UI/UX.

2. **Clarify** — If details are ambiguous, batch all questions and call
   `add-pending-questions` with JSON arguments. Each argument has a `question`
   field and optional `options` array. After calling it, stop and wait for
   answers. Only ask questions that affect feature design or implementation.

3. **Write the Expansion** — Produce a comprehensive expansion with these
   sections: Overview, User Stories, User Flows, Functional Requirements,
   Non-Functional Requirements, Edge Cases & Error Handling, Technical
   Considerations, and Open Questions. Be specific and concrete.

Output to a markdown file with an auto-derived UPPER_CASE slug name (e.g.
`FEATURE-IDEA.md`). The report must include the original idea, all
clarifications and answers, and all sections.
