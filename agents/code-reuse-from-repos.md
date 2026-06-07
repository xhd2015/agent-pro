# Agent Code Reuse from Existing Repositories

> **Original Idea:** "I have a set of git repositories of shared code, when writing code, I want agent to search my existing projects instead of inventing new code every time, how would I do that?"

---

## Overview

Allow an AI coding agent (like opencode) to search a user-configured set of local git repositories before generating new code. Instead of inventing functions, components, utilities, or patterns from scratch, the agent first performs semantic or keyword search across existing repos, surfaces relevant matches, and either reuses, adapts, or references them in the generated response. This reduces redundant code, enforces internal consistency, and accelerates development by building on existing work.

---

## User Stories

1. **Developer, Team Lead** — "I maintain a monorepo of shared utilities and several microservice repos. I want opencode to find existing `ratelimiter`, `retry`, or `logger` implementations across my repos before writing a new one."

2. **Platform Engineer** — "I have a `common-lib` repo with standard middleware, error types, and config loaders. When I ask the agent to add a new API endpoint, it should check `common-lib` for existing patterns and use them instead of inventing a new style."

3. **Tech Lead onboarding junior devs** — "I want the agent to act as a code librarian — when a junior asks 'write a database migration', the agent should find last month's migration file and use the same pattern rather than generating an unfamiliar structure."

4. **Solo developer, multiple projects** — "I copy-paste helpers between my own projects. I want the agent to find my prior implementations automatically so I stop duplicating code."

---

## User Flows

### Flow 1: Configure Repositories

1. User specifies a list of local paths or git repo roots (e.g. in `opencode.json` or a `.opencode/code-repos.json` file).
2. Optionally, user provides include/exclude glob patterns (e.g. `**/*.go`, `!vendor/**`).
3. Optionally, user configures an indexing strategy (full clone, sparse checkout, git-archive, etc.).
4. System validates paths exist and are git repos, reports any invalid entries.

### Flow 2: Agent Searches Before Generating

1. User issues a task: "Add a retry utility with exponential backoff."
2. Agent parses the request, identifies that a reusable utility *might* already exist.
3. Agent performs search across configured repos — using:
   - Keyword search (grep for `retry`, `backoff`, `Retry`)
   - Semantic search (embedding-based similarity if configured)
   - File-name heuristics (files named `retry*`, `backoff*`)
4. Agent retrieves top-N matches (configurable, default 5).
5. Agent presents findings in its response:
   - "Found `pkg/retry/retry.go` in `common-lib` with `func RetryWithBackoff(...)`. Reusing it."
   - Or: "No existing retry implementation found. Generating a new one inline."
6. Agent writes code that either:
   - Calls the existing function directly (import it).
   - Copies and adapts the pattern (with attribution).
   - Extends the existing implementation.

### Flow 3: User Reviews References

1. Agent's response includes citations: repo path, file path, line range, and a snippet.
2. User can click paths to open in editor.
3. User can accept, reject, or modify the agent's reuse decision.
4. User can mark a match as "always reuse" or "never reuse" for future sessions.

### Flow 4: Refresh Index

1. Repos are re-indexed:
   - On a configurable schedule (e.g. daily).
   - On demand via a command (`opencode index`).
   - On git hooks (post-merge, post-checkout) if running as a daemon.

---

## Functional Requirements

| ID | Requirement |
|----|-------------|
| F1 | User can configure one or more local git repository paths. |
| F2 | System validates that each path is a valid git working tree. |
| F3 | Agent performs search before code generation when task contains keywords suggesting reuse. |
| F4 | Search supports case-insensitive keyword matching (grep). |
| F5 | Search supports semantic relevance matching (optional, embedding-based). |
| F6 | Search results include: repo name, file path, line range, and surrounding context (configurable lines, default 10). |
| F7 | Agent presents reuse candidates inline, citing source. |
| F8 | Agent supports three reuse modes: **import** (direct dependency), **adapt** (copy with changes), **reference** (link in comment). |
| F9 | User can exclude specific repos, directories, or file patterns from search. |
| F10 | Index updates are on-demand and/or periodic; no stale results older than the configured TTL. |
| F11 | A CLI command (`opencode code-index`) lists indexed repos and their index freshness. |
| F12 | Agent prefers code from repos ranked by user (priority order) or by recency (last commit date). |

---

## Non-Functional Requirements

| ID | Requirement |
|----|-------------|
| N1 | **Performance** — Search should complete in under 2 seconds for repos up to 500 MB / 50K files. |
| N2 | **Performance** — First-time indexing of a repo should not block user tasks; index asynchronously. |
| N3 | **Storage** — Index should not exceed 10% of repo size (for semantic embeddings, configurable limit). |
| N4 | **Security** — Agent must not exfiltrate indexed code via search; all processing is local. |
| N5 | **Security** — Respect `.gitignore` and any additional exclude rules; never leak private code. |
| N6 | **Usability** — Zero-config should work (auto-detect repos in parent/workspace directories). |
| N7 | **Reliability** — If index is stale or corrupted, fall back to live grep so search always works. |
| N8 | **Observability** — Agent should log when it reuses vs. generates new code, for audit/awareness. |

---

## Edge Cases & Error Handling

| Situation | Handling |
|-----------|----------|
| Repo path does not exist or is not a git repo | Log warning, remove from active list, notify user. |
| No results found | Agent generates from scratch and includes a note: "Searched N repos, no matches." |
| Multiple conflicting implementations found | Show top-3 with scores, ask user to pick, or use heuristics (most recent, most starred, in-priority-repo). |
| Index building fails (disk full, permissions) | Fall back to live grep on every request for that repo; warn user. |
| Git repo has uncommitted changes | Use working-tree content for search; flag as "uncommitted." |
| Very large repo (>1 GB) | Index only top-level directories matching include patterns; let user configure depth. |
| Repo deleted or moved after configuration | Detect on next index refresh, auto-remove, notify user. |
| Network path becomes unavailable (mounted drive) | Retry with backoff; after N failures, flag as unavailable. |
| Search query is too broad (e.g. single letter) | Require minimum query length (default 3 chars) or return empty set. |

---

## Technical Considerations

### Search Strategy Options

| Approach | Pros | Cons |
|----------|------|------|
| **Live grep** (`rg`, `git grep`) | Always fresh, no index, simple | Slow for large repos, no semantic understanding |
| **Keyword index** (trie + inverted index) | Fast, low resource, deterministic | No synonym/meaning matching |
| **Embedding index** (vector DB — SQLite + `sentence-transformers`) | Semantic matching, understands intent | Storage cost, model download, slower indexing |
| **Hybrid** (keyword pre-filter + embedding re-rank) | Best quality, reasonable speed | Most complex, two systems to maintain |

**Recommendation:** Start with live grep (simplest, always correct) and layer embedding search as an opt-in enhancement.

### Index Storage

- Use an embedded database (SQLite via `modernc.org/sqlite` in Go, or `better-sqlite3` in Node).
- Store: file path, repo ID, last commit hash, file hash, embedding vector (if applicable).
- Index version keyed by repo path + HEAD commit hash → skip re-index if unchanged.

### Integration with Agent Workflow

- The search step should happen **before** the agent begins code generation.
- Insert a system prompt directive: *"Before writing new code, search configured repositories for existing implementations. Cite what you find. If you reuse, explain your decision."*
- The search results should be injected into the LLM context as part of the conversation preamble (similar to tool results).

### Code Ranking Heuristics

- **Repo priority** — user-assigned weight.
- **Last commit time** — newer code is preferred.
- **Usage frequency** — files imported by many other files in the same repo score higher.
- **Test presence** — files with companion test files indicate production stability.

### Implementation Phases

1. **Phase 1** — Live grep only, manual repo list in config file, citation in agent output.
2. **Phase 2** — Inverted keyword index for faster search, auto-detect repos in workspace subtree.
3. **Phase 3** — Embedding-based semantic search, auto-attribution (import statements).
4. **Phase 4** — User feedback loop (upvote/downvote matches), adaptive reranking.

---

## Open Questions

1. Should the agent automatically decide when to search, or should the user explicitly invoke search (e.g. `/search` slash command)?
2. How should the agent handle code that needs modification before reuse? Should it diff and propose changes, or just reference and let the user adapt?
3. For **import-mode reuse**, should the agent attempt to add the repo as a Go module dependency / npm workspace package, or just inline the reference?
4. Should the search cache be shared across all users of a dev machine, or per-user?
5. How to handle binary files, generated code, and vendor directories — skip entirely or offer opt-in?
6. Should the agent search **all** branches or only the default branch / current working tree?
7. If two developers have different sets of repos, should the config be per-project (committed) or per-user (global)?
8. What license/attribution concerns exist when copying code from one internal project to another?
