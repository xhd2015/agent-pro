import type { SessionSummary } from './api/client'

/** Keep the current list when a poll fetch fails (null). */
export function sessionsPollResult(
  fetched: SessionSummary[] | null,
  current: SessionSummary[],
): SessionSummary[] {
  if (fetched != null) return fetched
  return current
}