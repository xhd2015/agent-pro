import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
} from 'react'
import { useNavigate } from 'react-router-dom'
import {
  createSession,
  fetchRunners,
  fetchSessionsPage,
  fetchStatus,
  getRunner,
  setRunner,
  type SessionListCounts,
  type SessionSummary,
} from '../api/client'
import { Composer } from '../components/Composer'
import { HomeTopBar } from '../components/HomeTopBar'
import { QuickResumeStrip } from '../components/QuickResumeStrip'
import { SessionList } from '../components/SessionList'
import { SessionListHeader } from '../components/SessionListHeader'
import { Shell } from '../components/Shell'
import {
  freezeHomeListScrollTop,
  rememberHomeListScrollTop,
  readRememberedHomeListScrollTop,
  unfreezeHomeListScrollTop,
  type HomeSessionListCache,
} from '../homeSessionListCache'
import { useFollowScroll } from '../hooks/useFollowScroll'
import { TOP_THRESHOLD_PX, distanceFromTop } from '../lib/followScroll'
import {
  sortSessionsNewestFirst,
  type SessionStatusFilter,
} from '../sessionDisplay'
import './HomePage.css'

const PAGE_SIZE = 30
const POLL_WINDOW_CAP = 150
const SEARCH_DEBOUNCE_MS = 300
/** Sessions list poll — only while home is visible. Faster when something is running. */
const SESSIONS_POLL_MS_RUNNING = 5000
const SESSIONS_POLL_MS_IDLE = 15000

export type HomePageProps = {
  /** App-level composer draft (survives /workspace round-trip). */
  draft: string
  onDraftChange: (value: string) => void
  /** App-level list cache (survives session detail round-trip). */
  listCache: HomeSessionListCache
  onListCacheChange: Dispatch<SetStateAction<HomeSessionListCache>>
  /** False while keep-alive home is covered by session/workspace routes. */
  isActive?: boolean
}

export function HomePage({
  draft,
  onDraftChange,
  listCache,
  onListCacheChange,
  isActive = true,
}: HomePageProps) {
  const navigate = useNavigate()
  const {
    sessions,
    sessionsLoaded,
    statusFilter,
    searchInput,
    debouncedQ,
    hasMore,
    total,
    counts,
    listScrollTop,
  } = listCache

  const patchCache = useCallback(
    (patch: Partial<HomeSessionListCache>) => {
      onListCacheChange((prev) => ({ ...prev, ...patch }))
    },
    [onListCacheChange],
  )

  const setSessions = useCallback(
    (next: SessionSummary[] | ((prev: SessionSummary[]) => SessionSummary[])) => {
      onListCacheChange((prev) => ({
        ...prev,
        sessions: typeof next === 'function' ? next(prev.sessions) : next,
      }))
    },
    [onListCacheChange],
  )

  const [sessionsRefreshing, setSessionsRefreshing] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [runners, setRunners] = useState<string[]>(['opencode'])
  const [runner, setRunnerState] = useState(getRunner)
  const [sending, setSending] = useState(false)
  const [workspace, setWorkspace] = useState('')
  /** User home dir from status — used for ~/ compact labels on session cards. */
  const [userHome, setUserHome] = useState('')
  const [ready, setReady] = useState(sessionsLoaded)
  /** Near list end: hide floating jump so it does not cover Load more. */
  const [nearListEnd, setNearListEnd] = useState(false)
  const sessionListRef = useRef<HTMLElement>(null)
  // Live scroll offset — do NOT mirror App listScrollTop every render (that lags
  // behind the user and caused poll-driven pin restore to jump back).
  const listScrollTopRef = useRef(
    Math.max(listScrollTop, readRememberedHomeListScrollTop()),
  )
  const sessionsRef = useRef(sessions)
  sessionsRef.current = sessions
  const statusFilterRef = useRef(statusFilter)
  statusFilterRef.current = statusFilter
  const debouncedQRef = useRef(debouncedQ)
  debouncedQRef.current = debouncedQ
  // Capture restore target once per mount (App cache + module/sessionStorage backup).
  const restoreScrollTopRef = useRef(
    Math.max(listScrollTop, readRememberedHomeListScrollTop()),
  )
  const isActiveRef = useRef(isActive)
  isActiveRef.current = isActive
  const wasActiveRef = useRef(isActive)
  /** After session-row mousedown, ignore further scroll persists (focus scroll-into-view). */
  const freezeScrollPersistRef = useRef(false)
  const nearListEndTimerRef = useRef<number | null>(null)

  // Trust server newest-first order; re-sort only as a stable client guard.
  const sortedSessions = useMemo(() => sortSessionsNewestFirst(sessions), [sessions])

  const {
    followModeRef,
    showJumpToLatest,
    syncFollowFromScroll,
    markUserScrollIntent,
    handleJumpToLatest,
    resetFollow,
    pinScrollAt,
  } = useFollowScroll(sessionListRef, [sortedSessions, statusFilter, debouncedQ], {
    anchor: 'top',
    restoreScrollTop: restoreScrollTopRef.current,
  })

  // Re-apply cached scroll ONLY when keep-alive home becomes active again
  // (session/workspace → home). Never on poll/sortedSessions updates — that was
  // snapping the list back to a stale pin after the user scrolled further.
  useLayoutEffect(() => {
    const becameActive = isActive && !wasActiveRef.current
    wasActiveRef.current = isActive
    if (!becameActive) {
      if (!isActive) return
      return
    }
    const target = Math.max(
      restoreScrollTopRef.current,
      listScrollTopRef.current,
      readRememberedHomeListScrollTop(),
    )
    if (target <= TOP_THRESHOLD_PX) {
      freezeScrollPersistRef.current = false
      unfreezeHomeListScrollTop()
      return
    }
    freezeScrollPersistRef.current = true
    const apply = () => {
      const el = sessionListRef.current
      if (!el) return
      if (el.scrollHeight <= el.clientHeight + 2) return
      pinScrollAt(target)
      listScrollTopRef.current = el.scrollTop
    }
    apply()
    const raf = window.requestAnimationFrame(apply)
    const t1 = window.setTimeout(apply, 50)
    const t2 = window.setTimeout(() => {
      apply()
      freezeScrollPersistRef.current = false
      unfreezeHomeListScrollTop()
      rememberHomeListScrollTop(target, { force: true })
    }, 250)
    return () => {
      window.cancelAnimationFrame(raf)
      window.clearTimeout(t1)
      window.clearTimeout(t2)
    }
  }, [isActive, pinScrollAt])

  useEffect(() => {
    const id = window.setTimeout(() => {
      const next = searchInput.trim()
      onListCacheChange((prev) =>
        prev.debouncedQ === next ? prev : { ...prev, debouncedQ: next },
      )
    }, SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(id)
  }, [searchInput, onListCacheChange])

  const persistListScroll = useCallback(() => {
    const el = sessionListRef.current
    const top = el?.scrollTop ?? listScrollTopRef.current
    listScrollTopRef.current = top
    rememberHomeListScrollTop(top)
    onListCacheChange((prev) =>
      Math.abs(prev.listScrollTop - top) < 1 ? prev : { ...prev, listScrollTop: top },
    )
  }, [onListCacheChange])

  // Native scroll listener while home is active (ignore reflow noise while hidden).
  // Capture pointerdown on session rows so we snapshot scroll before focus scroll-into-view.
  useEffect(() => {
    const el = sessionListRef.current
    if (!el) return
    const onScrollNative = () => {
      if (!isActiveRef.current || freezeScrollPersistRef.current) return
      listScrollTopRef.current = el.scrollTop
      // Module/sessionStorage only — avoid setState on every scroll (was janking).
      rememberHomeListScrollTop(el.scrollTop)
      const distBottom = el.scrollHeight - el.scrollTop - el.clientHeight
      const nextNear = distBottom < 140
      // Debounce React state for jump-chip hide near list end.
      if (nearListEndTimerRef.current != null) {
        window.clearTimeout(nearListEndTimerRef.current)
      }
      nearListEndTimerRef.current = window.setTimeout(() => {
        setNearListEnd((prev) => (prev === nextNear ? prev : nextNear))
      }, 50)
    }
    const onRowPointerDownCapture = (ev: Event) => {
      if (!isActiveRef.current) return
      const target = ev.target
      if (!(target instanceof Element)) return
      if (!target.closest('[data-testid="session-item"]')) return
      const top = el.scrollTop
      listScrollTopRef.current = top
      restoreScrollTopRef.current = top
      freezeHomeListScrollTop(top)
      onListCacheChange((prev) =>
        Math.abs(prev.listScrollTop - top) < 1 ? prev : { ...prev, listScrollTop: top },
      )
      freezeScrollPersistRef.current = true
    }
    el.addEventListener('scroll', onScrollNative, { passive: true })
    el.addEventListener('pointerdown', onRowPointerDownCapture, true)
    el.addEventListener('mousedown', onRowPointerDownCapture, true)
    return () => {
      if (nearListEndTimerRef.current != null) {
        window.clearTimeout(nearListEndTimerRef.current)
      }
      if (isActiveRef.current && !freezeScrollPersistRef.current) {
        // Flush App cache once on teardown, not on every scroll tick.
        const top = el.scrollTop
        listScrollTopRef.current = top
        rememberHomeListScrollTop(top)
        onListCacheChange((prev) =>
          Math.abs(prev.listScrollTop - top) < 1 ? prev : { ...prev, listScrollTop: top },
        )
      }
      el.removeEventListener('scroll', onScrollNative)
      el.removeEventListener('pointerdown', onRowPointerDownCapture, true)
      el.removeEventListener('mousedown', onRowPointerDownCapture, true)
    }
  }, [onListCacheChange, sortedSessions.length])

  // Also flush on page hide / unmount of Home.
  useEffect(() => {
    return () => {
      persistListScroll()
    }
  }, [persistListScroll])

  const applyPageMeta = useCallback(
    (page: {
      total: number
      has_more: boolean
      counts: SessionListCounts
    }) => {
      patchCache({
        total: page.total,
        hasMore: page.has_more,
        counts: page.counts,
      })
    },
    [patchCache],
  )

  const refresh = useCallback(
    async (mode: 'reset' | 'poll' = 'reset') => {
      // Poll is sessions-only. Runners + workspace status almost never change on a
      // 3s cadence; fetch them on bootstrap/reset (and once when returning home).
      if (mode === 'poll') {
        try {
          const loaded = sessionsRef.current.length
          const limit =
            loaded > 0 ? Math.min(loaded, POLL_WINDOW_CAP) : PAGE_SIZE
          const page = await fetchSessionsPage({
            limit,
            offset: 0,
            q: debouncedQRef.current,
            status: statusFilterRef.current,
          })
          if (page != null) {
            setSessions(page.sessions)
            applyPageMeta(page)
            patchCache({ sessionsLoaded: true })
          }
        } finally {
          setReady(true)
        }
        return
      }

      setSessionsRefreshing(true)
      try {
        const [page, r, status] = await Promise.all([
          fetchSessionsPage({
            limit: PAGE_SIZE,
            offset: 0,
            q: debouncedQRef.current,
            status: statusFilterRef.current,
          }),
          fetchRunners(),
          fetchStatus(),
        ])
        if (page != null) {
          // Replace list without wiping first (keep current until new data arrives).
          setSessions(page.sessions)
          applyPageMeta(page)
          patchCache({ sessionsLoaded: true })
        }
        if (r.runners.length > 0) {
          setRunners(r.runners)
        }
        if (status?.workspace) {
          setWorkspace(status.workspace)
        }
        if (status?.home) {
          setUserHome(status.home)
        }
      } finally {
        setSessionsRefreshing(false)
        setReady(true)
      }
    },
    [applyPageMeta, patchCache, setSessions],
  )

  const loadMore = useCallback(async () => {
    if (loadingMore || !hasMore) return
    setLoadingMore(true)
    try {
      const offset = sessionsRef.current.length
      const page = await fetchSessionsPage({
        limit: PAGE_SIZE,
        offset,
        q: debouncedQRef.current,
        status: statusFilterRef.current,
      })
      if (page == null) return
      setSessions((prev) => {
        const seen = new Set(prev.map((s) => s.session_id))
        const appended = page.sessions.filter((s) => !seen.has(s.session_id))
        return [...prev, ...appended]
      })
      patchCache({
        total: page.total,
        hasMore: page.has_more,
        counts: page.counts,
      })
    } finally {
      setLoadingMore(false)
    }
  }, [hasMore, loadingMore, patchCache, setSessions])

  const handleFilterChange = useCallback(
    (next: SessionStatusFilter) => {
      resetFollow()
      restoreScrollTopRef.current = 0
      rememberHomeListScrollTop(0)
      onListCacheChange((prev) => ({
        ...prev,
        statusFilter: next,
        listScrollTop: 0,
      }))
    },
    [onListCacheChange, resetFollow],
  )

  // Reset when filter/search changes; on remount with cached rows, poll window only.
  // Snapshot cache at mount so sessionsLoaded flipping true does not re-fire as a "filter" reset.
  const sessionsLoadedOnMountRef = useRef(sessionsLoaded)
  const sessionsLenOnMountRef = useRef(sessions.length)
  const fetchGenRef = useRef(0)
  useEffect(() => {
    const gen = ++fetchGenRef.current
    const isFirstForThisInstance = gen === 1
    if (
      isFirstForThisInstance &&
      sessionsLoadedOnMountRef.current &&
      sessionsLenOnMountRef.current > 0
    ) {
      void refresh('poll')
    } else {
      if (!isFirstForThisInstance) {
        // Filter/search change — start from top of first page.
        restoreScrollTopRef.current = 0
        rememberHomeListScrollTop(0)
        patchCache({ listScrollTop: 0 })
      }
      void refresh('reset')
    }
  }, [refresh, statusFilter, debouncedQ, patchCache])

  // Workspace path: one fetch when returning to home (e.g. after /workspace).
  // Initial bootstrap already loads status; do not re-fetch on first paint.
  const prevActiveForStatusRef = useRef(isActive)
  useEffect(() => {
    const becameActive = isActive && !prevActiveForStatusRef.current
    prevActiveForStatusRef.current = isActive
    if (!becameActive) return
    let cancelled = false
    void (async () => {
      const status = await fetchStatus()
      if (cancelled) return
      if (status?.workspace) setWorkspace(status.workspace)
      if (status?.home) setUserHome(status.home)
    })()
    return () => {
      cancelled = true
    }
  }, [isActive])

  // Sessions-only poll while home is active and the tab is visible.
  // Idle (no running agents): slow. Running: faster so status pills stay fresh.
  // Runners are NOT polled — they change only with server config / process restart.
  useEffect(() => {
    if (!isActive) return
    const hasRunning =
      (counts?.running ?? 0) > 0 ||
      sessionsRef.current.some((s) => s.status === 'running')
    const intervalMs = hasRunning ? SESSIONS_POLL_MS_RUNNING : SESSIONS_POLL_MS_IDLE
    const tick = () => {
      if (document.visibilityState !== 'visible') return
      void refresh('poll')
    }
    const id = window.setInterval(tick, intervalMs)
    const onVis = () => {
      if (document.visibilityState === 'visible') tick()
    }
    document.addEventListener('visibilitychange', onVis)
    return () => {
      window.clearInterval(id)
      document.removeEventListener('visibilitychange', onVis)
    }
  }, [refresh, isActive, counts?.running])

  const handleListScroll = useCallback(() => {
    // Explicit load only via Load more button — no infinite scroll auto-fetch.
    if (freezeScrollPersistRef.current) return
    syncFollowFromScroll()
    const el = sessionListRef.current
    if (!el) return
    listScrollTopRef.current = el.scrollTop
    const distBottom = el.scrollHeight - el.scrollTop - el.clientHeight
    setNearListEnd(distBottom < 140)
  }, [syncFollowFromScroll])

  const handleSessionNavigate = useCallback(() => {
    // Snapshot scroll before the browser scrolls the focused link into view.
    const el = sessionListRef.current
    const top = el?.scrollTop ?? listScrollTopRef.current
    listScrollTopRef.current = top
    restoreScrollTopRef.current = top
    freezeHomeListScrollTop(top)
    onListCacheChange((prev) =>
      Math.abs(prev.listScrollTop - top) < 1 ? prev : { ...prev, listScrollTop: top },
    )
    freezeScrollPersistRef.current = true
  }, [onListCacheChange])

  const handleJumpToLatestAndPersist = useCallback(() => {
    handleJumpToLatest()
    listScrollTopRef.current = 0
    rememberHomeListScrollTop(0)
    patchCache({ listScrollTop: 0 })
  }, [handleJumpToLatest, patchCache])

  const handleRunnerChange = (next: string) => {
    setRunnerState(next)
    setRunner(next)
  }

  const handleSend = async () => {
    const text = draft.trim()
    if (!text || sending) return
    const listEl = sessionListRef.current
    const detached =
      followModeRef.current === 'detached' ||
      (listEl != null && distanceFromTop(listEl) > TOP_THRESHOLD_PX)
    setSending(true)
    try {
      const session = await createSession(runner, text)
      onDraftChange('')
      if (session) {
        if (detached) {
          await refresh('reset')
        } else {
          navigate(`/sessions/${encodeURIComponent(session.session_id)}`)
        }
      } else {
        await refresh('reset')
      }
    } finally {
      setSending(false)
    }
  }

  const openWorkspaceSelector = useCallback(() => {
    navigate('/workspace')
  }, [navigate])

  // Do not hide existing content during refresh; only subtle loading when empty.
  const showSubtleLoading = !ready || (!sessionsLoaded && sortedSessions.length === 0)
  const hasAnyInStore = (counts?.all ?? 0) > 0 || sortedSessions.length > 0
  const showListChrome =
    hasAnyInStore ||
    debouncedQ !== '' ||
    searchInput !== '' ||
    statusFilter !== 'all'

  const loadMoreFooter =
    hasMore ? (
      <button
        type="button"
        className="session-load-more"
        data-testid="session-load-more"
        onClick={() => void loadMore()}
        disabled={loadingMore}
      >
        {loadingMore ? 'Loading…' : 'Load more'}
      </button>
    ) : null

  const homeMain = showListChrome ? (
      <div className="session-list-region">
        <SessionListHeader
          sessions={sortedSessions}
          visibleCount={sortedSessions.length}
          filter={statusFilter}
          onFilterChange={handleFilterChange}
          refreshing={sessionsRefreshing}
          counts={counts}
          total={total}
          searchQuery={searchInput}
          onSearchChange={(value) => patchCache({ searchInput: value })}
        />
        {statusFilter === 'all' && !debouncedQ ? (
          <QuickResumeStrip sessions={sortedSessions} />
        ) : null}
        {sortedSessions.length > 0 && showJumpToLatest && !nearListEnd ? (
          <button
            type="button"
            className="jump-to-latest"
            data-testid="jump-to-latest"
            onClick={handleJumpToLatestAndPersist}
          >
            Jump to latest
          </button>
        ) : null}
        {sortedSessions.length > 0 ? (
          <SessionList
            ref={sessionListRef}
            sessions={sortedSessions}
            onScroll={handleListScroll}
            onWheel={markUserScrollIntent}
            onTouchStart={markUserScrollIntent}
            onSessionNavigate={handleSessionNavigate}
            footer={loadMoreFooter}
            homeDir={userHome}
          />
        ) : (
          <div className="session-filter-empty" data-testid="session-filter-empty">
            <p>
              {debouncedQ
                ? 'No sessions match your search.'
                : statusFilter === 'running'
                  ? 'No agents are running right now. Start a new chat below or show all sessions.'
                  : 'No finished sessions yet. Running chats appear under Running.'}
            </p>
            {debouncedQ ? (
              <button
                type="button"
                className="session-filter-reset"
                onClick={() => patchCache({ searchInput: '' })}
              >
                Clear search
              </button>
            ) : (
              <button
                type="button"
                className="session-filter-reset"
                onClick={() => handleFilterChange('all')}
              >
                Show all sessions
              </button>
            )}
          </div>
        )}
      </div>
    ) : (
      <div className="empty-state" data-testid="empty-state">
        {showSubtleLoading ? (
          <div
            className="home-loading-subtle"
            data-testid="home-loading"
            aria-label="Loading sessions"
          >
            <span className="home-loading-indicator" aria-hidden="true" />
          </div>
        ) : null}
        <div className="empty-state-icon" aria-hidden="true">
          ◇
        </div>
        <h2>Start a session</h2>
        <p>Choose a runner above, then send a message to kick off your agent.</p>
      </div>
    )

  return (
    <Shell homePage>
      <HomeTopBar
        runners={runners}
        runner={runner}
        onRunnerChange={handleRunnerChange}
        workspace={workspace}
        onOpenWorkspaceSelector={openWorkspaceSelector}
      />
      <div className="main-panel home-active" data-testid="home-active">
        {homeMain}
      </div>
      <Composer
        value={draft}
        onChange={onDraftChange}
        onSend={() => void handleSend()}
        sending={sending || !ready}
        placeholder={`New chat with ${runner}…`}
      />
    </Shell>
  )
}
