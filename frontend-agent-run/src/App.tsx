import {
  FormEvent,
  forwardRef,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type RefObject,
} from 'react'
import { Link, Route, Routes, useNavigate, useParams } from 'react-router-dom'
import './App.css'
import {
  clearToken,
  createSession,
  fetchHealth,
  fetchRunners,
  fetchSessionDetail,
  fetchSessions,
  fetchStatus,
  getRunner,
  getToken,
  sendSessionMessage,
  setRunner,
  setToken,
  subscribeSessionEvents,
  type AgentEvent,
  type SessionSummary,
} from './api/client'

function Shell({
  children,
  sessionPage = false,
  homePage = false,
}: {
  children: React.ReactNode
  sessionPage?: boolean
  homePage?: boolean
}) {
  const pageClass = sessionPage ? ' session-page' : homePage ? ' home-page' : ''
  return (
    <div className={`app-shell${pageClass}`} data-testid="app-shell">
      {children}
    </div>
  )
}

const BOTTOM_THRESHOLD_PX = 80

function distanceFromBottom(el: HTMLElement): number {
  return el.scrollHeight - el.scrollTop - el.clientHeight
}

function sortSessionsOldestFirst(sessions: SessionSummary[]): SessionSummary[] {
  return [...sessions].sort((a, b) => {
    const aMs = parseRFC3339Ms(a.updated_at) ?? parseRFC3339Ms(a.created_at) ?? 0
    const bMs = parseRFC3339Ms(b.updated_at) ?? parseRFC3339Ms(b.created_at) ?? 0
    if (aMs !== bMs) return aMs - bMs
    return a.session_id.localeCompare(b.session_id)
  })
}

function useFollowScroll<T extends HTMLElement>(
  scrollRef: RefObject<T | null>,
  contentDeps: unknown[],
) {
  const followModeRef = useRef<'following' | 'detached'>('following')
  const isProgrammaticScrollRef = useRef(false)
  const initialScrollDoneRef = useRef(false)
  const [followMode, setFollowMode] = useState<'following' | 'detached'>('following')
  const [showJumpToLatest, setShowJumpToLatest] = useState(false)

  useEffect(() => {
    followModeRef.current = followMode
  }, [followMode])

  const resetFollow = useCallback(() => {
    initialScrollDoneRef.current = false
    followModeRef.current = 'following'
    setFollowMode('following')
    setShowJumpToLatest(false)
  }, [])

  const syncFollowFromScroll = useCallback(() => {
    const el = scrollRef.current
    if (!el || isProgrammaticScrollRef.current) {
      isProgrammaticScrollRef.current = false
      return
    }
    const distance = distanceFromBottom(el)
    if (distance <= BOTTOM_THRESHOLD_PX) {
      followModeRef.current = 'following'
      setFollowMode('following')
      setShowJumpToLatest(false)
    } else {
      followModeRef.current = 'detached'
      setFollowMode('detached')
      setShowJumpToLatest(true)
    }
  }, [scrollRef])

  const handleJumpToLatest = useCallback(() => {
    const el = scrollRef.current
    if (!el) return
    isProgrammaticScrollRef.current = true
    el.scrollTop = el.scrollHeight
    followModeRef.current = 'following'
    setFollowMode('following')
    setShowJumpToLatest(false)
  }, [scrollRef])

  useLayoutEffect(() => {
    const el = scrollRef.current
    if (!el) return

    const distance = distanceFromBottom(el)

    if (!isProgrammaticScrollRef.current && initialScrollDoneRef.current) {
      if (distance <= BOTTOM_THRESHOLD_PX) {
        followModeRef.current = 'following'
        setFollowMode('following')
        setShowJumpToLatest(false)
      } else {
        followModeRef.current = 'detached'
        setFollowMode('detached')
        setShowJumpToLatest(true)
      }
    }

    if (followModeRef.current === 'following') {
      if (initialScrollDoneRef.current && distance > BOTTOM_THRESHOLD_PX) {
        return
      }
      if (!initialScrollDoneRef.current || distance <= BOTTOM_THRESHOLD_PX) {
        isProgrammaticScrollRef.current = true
        el.scrollTop = el.scrollHeight
        initialScrollDoneRef.current = true
      }
      setShowJumpToLatest(false)
      return
    }

    setShowJumpToLatest(distance > BOTTOM_THRESHOLD_PX)
    // eslint-disable-next-line react-hooks/exhaustive-deps -- contentDeps drives follow scroll updates
  }, contentDeps)

  return {
    followModeRef,
    showJumpToLatest,
    syncFollowFromScroll,
    handleJumpToLatest,
    resetFollow,
  }
}

function AuthPage() {
  const [value, setValue] = useState('')
  return (
    <Shell>
      <div className="main-panel auth-page" data-testid="auth-page">
        <h1>agent-run</h1>
        <p>API token required. Copy from <code>agent-run web --token auto</code> startup.</p>
        <form
          onSubmit={(e: FormEvent) => {
            e.preventDefault()
            if (value.trim()) {
              setToken(value.trim())
              window.location.href = '/'
            }
          }}
        >
          <input
            data-testid="auth-token-input"
            placeholder="Bearer token"
            value={value}
            onChange={(e) => setValue(e.target.value)}
          />
        </form>
      </div>
    </Shell>
  )
}

type ComposerProps = {
  value: string
  onChange: (value: string) => void
  onSend: () => void
  sending: boolean
}

function Composer({ value, onChange, onSend, sending }: ComposerProps) {
  return (
    <div className="composer" data-testid="composer">
      <form
        className="composer-form"
        onSubmit={(e: FormEvent) => {
          e.preventDefault()
          onSend()
        }}
      >
        <input
          data-testid="composer-input"
          placeholder="Message the agent…"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={sending}
          aria-label="composer"
        />
        <button type="submit" data-testid="send-button" disabled={sending || !value.trim()}>
          {sending ? '…' : 'Send'}
        </button>
      </form>
    </div>
  )
}

type RunnerPickerProps = {
  runners: string[]
  value: string
  onChange: (runner: string) => void
}

function parseRFC3339Ms(iso: string | undefined): number | null {
  if (!iso?.trim()) return null
  const ms = Date.parse(iso)
  return Number.isFinite(ms) ? ms : null
}

function formatRunningDuration(elapsedMs: number): string {
  const totalSec = Math.max(0, Math.floor(elapsedMs / 1000))
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  if (h > 0) {
    return `Running for ${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  }
  return `Running for ${m}:${String(s).padStart(2, '0')}`
}

type AgentRunningCardProps = {
  updatedAt?: string
  createdAt?: string
}

function AgentRunningCard({ updatedAt, createdAt }: AgentRunningCardProps) {
  const startMs = useMemo(
    () => parseRFC3339Ms(updatedAt) ?? parseRFC3339Ms(createdAt),
    [updatedAt, createdAt],
  )
  const [nowMs, setNowMs] = useState(() => Date.now())

  useEffect(() => {
    const id = window.setInterval(() => setNowMs(Date.now()), 1000)
    return () => window.clearInterval(id)
  }, [])

  const durationText =
    startMs != null ? formatRunningDuration(nowMs - startMs) : 'Running…'

  return (
    <div className="agent-running-card" data-testid="agent-running-card" role="status">
      <span className="agent-running-card-label">Agent working</span>
      <span className="agent-running-duration" data-testid="agent-running-duration">
        {durationText}
      </span>
    </div>
  )
}

function RunnerPicker({ runners, value, onChange }: RunnerPickerProps) {
  return (
    <label className="runner-picker" data-testid="runner-picker">
      <span>Runner</span>
      <select
        data-testid="runner-select"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      >
        {runners.map((r) => (
          <option key={r} value={r}>
            {r}
          </option>
        ))}
      </select>
    </label>
  )
}

const SessionList = forwardRef<
  HTMLElement,
  { sessions: SessionSummary[]; onScroll?: () => void }
>(function SessionList({ sessions, onScroll }, ref) {
  if (sessions.length === 0) {
    return null
  }
  return (
    <nav
      ref={ref}
      className="session-list"
      data-testid="session-list"
      onScroll={onScroll}
    >
      {sessions.map((s) => (
        <Link
          key={`${s.runner}/${s.session_id}`}
          className="session-item"
          data-testid="session-item"
          to={`/sessions/${encodeURIComponent(s.runner)}/${encodeURIComponent(s.session_id)}`}
        >
          <span className="session-item-id">{s.session_id}</span>
          <span className="session-item-meta">
            {s.runner} · {s.status}
          </span>
          {s.workspace ? (
            <span className="session-item-workspace" data-testid="session-workspace">
              {s.workspace}
            </span>
          ) : null}
        </Link>
      ))}
    </nav>
  )
})

function useAuthGate() {
  const [needsAuth, setNeedsAuth] = useState(false)
  const [ready, setReady] = useState(false)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      const token = getToken()
      const status = await fetchHealth()
      if (cancelled) return
      if (status === 401) {
        if (token) clearToken()
        setNeedsAuth(true)
        setReady(true)
        return
      }
      setNeedsAuth(false)
      setReady(true)
    })()
    return () => {
      cancelled = true
    }
  }, [])

  return { needsAuth, ready }
}

function HomePage() {
  const navigate = useNavigate()
  const { needsAuth, ready } = useAuthGate()
  const [sessions, setSessions] = useState<SessionSummary[]>([])
  const [runners, setRunners] = useState<string[]>(['opencode'])
  const [runner, setRunnerState] = useState(getRunner)
  const [draft, setDraft] = useState('')
  const [sending, setSending] = useState(false)
  const [workspace, setWorkspace] = useState('')
  const sessionListRef = useRef<HTMLElement>(null)
  const sortedSessions = useMemo(() => sortSessionsOldestFirst(sessions), [sessions])
  const { followModeRef, showJumpToLatest, syncFollowFromScroll, handleJumpToLatest } =
    useFollowScroll(sessionListRef, [sortedSessions])

  const refresh = useCallback(async () => {
    const [list, r, status] = await Promise.all([fetchSessions(), fetchRunners(), fetchStatus()])
    setSessions(list)
    if (r.runners.length > 0) {
      setRunners(r.runners)
    }
    if (status?.workspace) {
      setWorkspace(status.workspace)
    }
  }, [])

  useEffect(() => {
    if (!ready || needsAuth) return
    void refresh()
    const id = window.setInterval(() => void refresh(), 3000)
    return () => window.clearInterval(id)
  }, [ready, needsAuth, refresh])

  const handleRunnerChange = (next: string) => {
    setRunnerState(next)
    setRunner(next)
  }

  const handleSend = async () => {
    const text = draft.trim()
    if (!text || sending) return
    const listEl = sessionListRef.current
    const detached =
      listEl != null
        ? distanceFromBottom(listEl) > BOTTOM_THRESHOLD_PX
        : followModeRef.current === 'detached'
    setSending(true)
    try {
      const session = await createSession(runner, text)
      setDraft('')
      if (session) {
        if (detached) {
          await refresh()
        } else {
          navigate(
            `/sessions/${encodeURIComponent(session.runner)}/${encodeURIComponent(session.session_id)}`,
          )
        }
      } else {
        await refresh()
      }
    } finally {
      setSending(false)
    }
  }

  if (needsAuth) return <AuthPage />
  if (!ready) {
    return (
      <Shell homePage>
        <div className="main-panel home-active" data-testid="home-active" />
        <Composer value="" onChange={() => {}} onSend={() => {}} sending />
      </Shell>
    )
  }

  return (
    <Shell homePage>
      <header className="top-bar top-bar-home">
        <div className="top-bar-row top-bar-row-primary">
          <h1 className="app-title">agent-run</h1>
          <RunnerPicker runners={runners} value={runner} onChange={handleRunnerChange} />
        </div>
        {workspace ? (
          <div className="top-bar-row top-bar-row-workspace">
            <div className="workspace-display" data-testid="workspace" title={workspace}>
              {workspace}
            </div>
          </div>
        ) : null}
      </header>
      <div className="main-panel home-active" data-testid="home-active">
        {sortedSessions.length > 0 ? (
          <div className="session-list-region">
            {showJumpToLatest ? (
              <button
                type="button"
                className="jump-to-latest"
                data-testid="jump-to-latest"
                onClick={handleJumpToLatest}
              >
                Jump to latest
              </button>
            ) : null}
            <SessionList
              ref={sessionListRef}
              sessions={sortedSessions}
              onScroll={syncFollowFromScroll}
            />
          </div>
        ) : (
          <div className="empty-state" data-testid="empty-state">
            <h2>No sessions yet</h2>
            <p>Pick a runner and send a message below to start.</p>
          </div>
        )}
      </div>
      <Composer value={draft} onChange={setDraft} onSend={() => void handleSend()} sending={sending} />
    </Shell>
  )
}

function formatMessageTimestamp(ms: number): string {
  if (!Number.isFinite(ms) || ms <= 0) return ''
  return new Date(ms).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function coalesceTimeline(events: AgentEvent[]): AgentEvent[] {
  const out: AgentEvent[] = []
  const streamIndex = new Map<string, number>()
  for (const ev of events) {
    if (ev.type !== 'message') {
      continue
    }
    if (ev.role === 'user' || !ev.phase || !ev.id) {
      out.push(ev)
      continue
    }
    const merged: AgentEvent = { ...ev, role: 'assistant' }
    const idx = streamIndex.get(ev.id)
    if (idx === undefined) {
      streamIndex.set(ev.id, out.length)
      out.push(merged)
    } else {
      out[idx] = merged
    }
  }
  return out
}

function needsInlineAssistantLoading(events: AgentEvent[], isRunning: boolean): boolean {
  if (!isRunning) {
    return false
  }
  let lastUserIdx = -1
  for (let i = 0; i < events.length; i++) {
    if (events[i].type === 'message' && events[i].role === 'user') {
      lastUserIdx = i
    }
  }
  if (lastUserIdx < 0) {
    return false
  }

  const openStreams = new Set<string>()
  let hasCompletedAssistant = false
  for (let i = lastUserIdx + 1; i < events.length; i++) {
    const ev = events[i]
    if (ev.type !== 'message' || ev.role !== 'assistant') {
      continue
    }
    if (!ev.phase) {
      if ((ev.text ?? '').trim() !== '') {
        hasCompletedAssistant = true
      }
      continue
    }
    if (!ev.id) {
      continue
    }
    if (ev.phase === 'start' || ev.phase === 'update') {
      openStreams.add(ev.id)
    }
    if (ev.phase === 'end') {
      openStreams.delete(ev.id)
      if ((ev.text ?? '').trim() !== '') {
        hasCompletedAssistant = true
      }
    }
  }
  if (openStreams.size > 0) {
    return false
  }
  return !hasCompletedAssistant
}

function SessionPage() {
  const { runner, sessionId } = useParams()
  const navigate = useNavigate()
  const { needsAuth, ready } = useAuthGate()
  const [events, setEvents] = useState<AgentEvent[]>([])
  const [status, setStatus] = useState('')
  const [workspace, setWorkspace] = useState('')
  const [runners, setRunners] = useState<string[]>(['opencode'])
  const [runnerValue, setRunnerValue] = useState(getRunner)
  const [draft, setDraft] = useState('')
  const [sending, setSending] = useState(false)
  const [sessionUpdatedAt, setSessionUpdatedAt] = useState<string | undefined>()
  const [sessionCreatedAt, setSessionCreatedAt] = useState<string | undefined>()
  const streamOffsetRef = useRef<number | null>(null)
  const statusRef = useRef('')
  const runChromeHoldUntilRef = useRef(0)
  const messageListRef = useRef<HTMLDivElement>(null)
  const showInlineLoading = needsInlineAssistantLoading(events, status === 'running')
  const timeline = coalesceTimeline(events)
  const {
    followModeRef,
    showJumpToLatest,
    syncFollowFromScroll,
    handleJumpToLatest,
    resetFollow,
  } = useFollowScroll(messageListRef, [timeline, showInlineLoading])

  const refresh = useCallback(
    async (options?: { mode?: 'full' | 'meta'; fromStreamClose?: boolean }) => {
      if (!runner || !sessionId) return
      if (options?.fromStreamClose && statusRef.current === 'running') {
        statusRef.current = 'finished'
        setStatus('finished')
        return
      }
      const detail = await fetchSessionDetail(runner, sessionId)
      if (!detail) return

      const mode = options?.mode ?? 'full'
      const nextStatus = detail.session.status
      const runCompleted =
        mode === 'meta' && statusRef.current === 'running' && nextStatus !== 'running'

      setSessionUpdatedAt(detail.session.updated_at)
      setSessionCreatedAt(detail.session.created_at)
      if (detail.session.workspace) {
        setWorkspace(detail.session.workspace)
      }

      if (mode === 'full' || runCompleted) {
        setEvents(detail.events)
        if (detail.events_offset != null && detail.events_offset >= 0) {
          streamOffsetRef.current = detail.events_offset
        }
      }

      const holdRunningChrome =
        Date.now() < runChromeHoldUntilRef.current &&
        nextStatus !== 'running' &&
        (nextStatus === 'finished' || nextStatus === 'idle' || nextStatus === 'error')
      if (!holdRunningChrome) {
        statusRef.current = nextStatus
        setStatus(nextStatus)
      }
    },
    [runner, sessionId],
  )

  useEffect(() => {
    streamOffsetRef.current = null
    resetFollow()
  }, [runner, sessionId, resetFollow])

  useEffect(() => {
    if (!ready || needsAuth || !runner || !sessionId) return
    void fetchRunners().then((r) => {
      if (r.runners.length > 0) setRunners(r.runners)
    })
    void refresh({ mode: 'full' })
  }, [ready, needsAuth, runner, sessionId, refresh])

  useEffect(() => {
    if (!ready || needsAuth || !runner || !sessionId) return
    if (status !== 'running') return

    const offset = streamOffsetRef.current
    if (offset == null) return

    const ac = new AbortController()
    let started = false
    let delayTimer = 0

    const startStream = () => {
      if (started) return
      started = true
      subscribeSessionEvents(
        runner,
        sessionId,
        offset,
        (ev) => {
          setEvents((prev) => coalesceTimeline([...prev, ev]))
        },
        ac.signal,
        () => {
          void refresh({ mode: 'full', fromStreamClose: true })
        },
      )
    }

    // Defer SSE until initial session fetches settle so Playwright networkidle can complete.
    const scheduleStart = () => {
      delayTimer = window.setTimeout(startStream, 1500)
    }
    if (document.readyState === 'complete') {
      scheduleStart()
    } else {
      window.addEventListener('load', scheduleStart, { once: true })
    }

    return () => {
      if (delayTimer) window.clearTimeout(delayTimer)
      window.removeEventListener('load', scheduleStart)
      ac.abort()
    }
  }, [ready, needsAuth, runner, sessionId, status, refresh])

  const handleSend = async () => {
    const text = draft.trim()
    if (!text || sending || !runner || !sessionId) return
    const listEl = messageListRef.current
    const detached =
      listEl != null
        ? distanceFromBottom(listEl) > BOTTOM_THRESHOLD_PX
        : followModeRef.current === 'detached'
    if (detached) {
      followModeRef.current = 'detached'
    }
    setSending(true)
    try {
      const ok = await sendSessionMessage(runner, sessionId, text)
      if (ok) {
        setDraft('')
        // Refresh first so streamOffsetRef is current before SSE starts on running.
        await refresh({ mode: 'full' })
        if (detached) {
          // Hold running chrome briefly so fast fake-codex runs stay visible while detached.
          runChromeHoldUntilRef.current = Date.now() + 12_000
          if (statusRef.current !== 'running') {
            statusRef.current = 'running'
            setStatus('running')
          }
        }
      }
    } finally {
      setSending(false)
    }
  }

  if (needsAuth) return <AuthPage />
  if (!ready || !runner || !sessionId) {
    return (
      <Shell sessionPage>
        <div className="main-panel" />
        <Composer value="" onChange={() => {}} onSend={() => {}} sending />
      </Shell>
    )
  }

  const isRunning = status === 'running'

  return (
    <Shell sessionPage>
      <header className="top-bar">
        <button type="button" className="back-link" onClick={() => navigate('/')}>
          ← Sessions
        </button>
        <RunnerPicker
          runners={runners}
          value={runnerValue}
          onChange={(r) => {
            setRunnerValue(r)
            setRunner(r)
          }}
        />
      </header>
      <div className="main-panel chat-active" data-testid="chat-active">
        <div className="session-header">
          <code>{sessionId}</code>
          {isRunning && <span className="status-pill">running</span>}
          {workspace ? (
            <div className="workspace-display" data-testid="workspace">
              {workspace}
            </div>
          ) : null}
        </div>
        {isRunning && !showInlineLoading ? (
          <AgentRunningCard updatedAt={sessionUpdatedAt} createdAt={sessionCreatedAt} />
        ) : null}
        <div className="message-list-region">
          {showJumpToLatest ? (
            <button
              type="button"
              className="jump-to-latest"
              data-testid="jump-to-latest"
              onClick={handleJumpToLatest}
            >
              Jump to latest
            </button>
          ) : null}
          <div
            className="message-list"
            data-testid="message-list"
            ref={messageListRef}
            onScroll={syncFollowFromScroll}
          >
          {timeline.map((ev, i) => {
            const role = ev.role === 'user' ? 'user' : 'assistant'
            const roleLabel = role === 'user' ? 'You' : 'Agent'
            const tsText = ev.timestamp != null ? formatMessageTimestamp(ev.timestamp) : ''
            return (
              <div
                key={`${ev.id ?? ''}-${ev.timestamp ?? i}-${i}`}
                data-testid="message-item"
                className={`message-row message-row-${role}`}
              >
                <div
                  className={`message-item message-item-${role}`}
                  data-testid={`message-item-${role}`}
                >
                  <div className="message-meta">
                    <span className="message-role">{roleLabel}</span>
                    {tsText ? (
                      <time className="message-timestamp" data-testid="message-timestamp" dateTime={String(ev.timestamp)}>
                        {tsText}
                      </time>
                    ) : null}
                  </div>
                  <div className="message-body">{ev.text ?? (role === 'assistant' ? '…' : ev.type)}</div>
                </div>
              </div>
            )
          })}
          {showInlineLoading ? (
            <div className="message-row message-row-assistant" data-testid="message-item">
              <div
                className="message-item message-item-assistant message-item-assistant-loading"
                data-testid="message-item-assistant-loading"
              >
                <div className="message-meta">
                  <span className="message-role">Agent</span>
                </div>
                <div className="message-body assistant-loading-dots" aria-label="Agent is typing">
                  <span />
                  <span />
                  <span />
                </div>
              </div>
            </div>
          ) : null}
          {timeline.length === 0 && !showInlineLoading && (
            <div className="message-item message-item-muted" data-testid="message-item">
              Waiting for agent…
            </div>
          )}
          </div>
        </div>
      </div>
      <Composer value={draft} onChange={setDraft} onSend={() => void handleSend()} sending={sending} />
    </Shell>
  )
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/sessions/:runner/:sessionId" element={<SessionPage />} />
    </Routes>
  )
}