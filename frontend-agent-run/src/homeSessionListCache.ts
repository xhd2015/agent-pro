import type { SessionListCounts, SessionSummary } from './api/client'
import type { SessionStatusFilter } from './sessionDisplay'

/**
 * App-level session list cache so HomePage unmount (enter session / workspace)
 * preserves loaded pages, filters, and scroll position.
 */
export type HomeSessionListCache = {
  sessions: SessionSummary[]
  sessionsLoaded: boolean
  statusFilter: SessionStatusFilter
  searchInput: string
  debouncedQ: string
  hasMore: boolean
  total: number
  counts: SessionListCounts | null
  /** Last known scrollTop of [data-testid=session-list]. */
  listScrollTop: number
}

const STORAGE_KEY = 'agent-run-home-list-scroll'

/** Module-level backup so scroll survives even if a state update races unmount. */
let rememberedListScrollTop = 0
/**
 * While frozen (session-row activate → home inactive), ignore overwrites from
 * focus scroll-into-view so restore still uses the pre-navigate offset.
 */
let frozenListScrollTop: number | null = null

function writeStorage(top: number): void {
  if (typeof window === 'undefined') return
  try {
    window.sessionStorage.setItem(STORAGE_KEY, String(Math.round(top)))
  } catch {
    // ignore quota / private mode
  }
}

export function rememberHomeListScrollTop(top: number, opts?: { force?: boolean }): void {
  if (!Number.isFinite(top) || top < 0) return
  if (frozenListScrollTop != null && !opts?.force) {
    return
  }
  rememberedListScrollTop = top
  writeStorage(top)
}

/**
 * Snapshot scroll at session-row activate; blocks later accidental overwrites.
 * Uses max(top, remembered) so focus scroll-into-view (which lowers scrollTop
 * before click handlers) cannot clobber a mid-list position already remembered.
 */
export function freezeHomeListScrollTop(top: number): void {
  if (!Number.isFinite(top) || top < 0) return
  if (frozenListScrollTop != null) return
  const next = Math.max(top, rememberedListScrollTop)
  frozenListScrollTop = next
  rememberedListScrollTop = next
  writeStorage(next)
}

export function unfreezeHomeListScrollTop(): void {
  frozenListScrollTop = null
}

export function readRememberedHomeListScrollTop(): number {
  if (frozenListScrollTop != null && frozenListScrollTop > 0) {
    return frozenListScrollTop
  }
  if (rememberedListScrollTop > 0) return rememberedListScrollTop
  if (typeof window === 'undefined') return 0
  try {
    const raw = window.sessionStorage.getItem(STORAGE_KEY)
    const n = raw != null ? Number(raw) : 0
    return Number.isFinite(n) && n > 0 ? n : 0
  } catch {
    return 0
  }
}

export const initialHomeSessionListCache = (): HomeSessionListCache => ({
  sessions: [],
  sessionsLoaded: false,
  statusFilter: 'all',
  searchInput: '',
  debouncedQ: '',
  hasMore: false,
  total: 0,
  counts: null,
  listScrollTop: readRememberedHomeListScrollTop(),
})
