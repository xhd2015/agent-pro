import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react'
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

function Shell({ children }: { children: React.ReactNode }) {
  return (
    <div className="app-shell" data-testid="app-shell">
      {children}
    </div>
  )
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

function SessionList({ sessions }: { sessions: SessionSummary[] }) {
  if (sessions.length === 0) {
    return null
  }
  return (
    <nav className="session-list" data-testid="session-list">
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
}

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
    setSending(true)
    try {
      const session = await createSession(runner, text)
      setDraft('')
      if (session) {
        navigate(`/sessions/${encodeURIComponent(session.runner)}/${encodeURIComponent(session.session_id)}`)
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
      <Shell>
        <div className="main-panel" />
        <Composer value="" onChange={() => {}} onSend={() => {}} sending />
      </Shell>
    )
  }

  return (
    <Shell>
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
      <div className="main-panel">
        <SessionList sessions={sessions} />
        {sessions.length === 0 && (
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

function hasOpenAssistantStream(events: AgentEvent[]): boolean {
  const ended = new Set<string>()
  let open = false
  for (const ev of events) {
    if (ev.type !== 'message' || ev.role !== 'assistant' || !ev.phase || !ev.id) {
      continue
    }
    if (ev.phase === 'end') {
      ended.add(ev.id)
    }
    if ((ev.phase === 'start' || ev.phase === 'update') && !ended.has(ev.id)) {
      open = true
    }
  }
  return open
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
  const [eventsOffset, setEventsOffset] = useState<number | null>(null)
  const [aggressivePoll, setAggressivePoll] = useState(false)

  const refresh = useCallback(async () => {
    if (!runner || !sessionId) return
    const detail = await fetchSessionDetail(runner, sessionId)
    if (detail) {
      setEvents(detail.events)
      setStatus(detail.session.status)
      setSessionUpdatedAt(detail.session.updated_at)
      setSessionCreatedAt(detail.session.created_at)
      if (detail.events_offset != null && detail.events_offset >= 0) {
        setEventsOffset(detail.events_offset)
      }
      if (detail.session.workspace) {
        setWorkspace(detail.session.workspace)
      }
      const liveTail =
        detail.session.status === 'running' &&
        (needsInlineAssistantLoading(detail.events, true) || hasOpenAssistantStream(detail.events))
      setAggressivePoll(liveTail)
    }
  }, [runner, sessionId])

  useEffect(() => {
    if (!ready || needsAuth || !runner || !sessionId) return
    void fetchRunners().then((r) => {
      if (r.runners.length > 0) setRunners(r.runners)
    })
    void refresh()
    const pollMs = aggressivePoll ? 250 : 2000
    const id = window.setInterval(() => void refresh(), pollMs)
    return () => window.clearInterval(id)
  }, [ready, needsAuth, runner, sessionId, refresh, aggressivePoll])

  const showInlineLoading = needsInlineAssistantLoading(events, status === 'running')

  useEffect(() => {
    if (!ready || needsAuth || !runner || !sessionId || eventsOffset == null) return
    const skipSSE =
      status === 'running' &&
      !needsInlineAssistantLoading(events, true) &&
      !hasOpenAssistantStream(events)
    if (skipSSE) {
      return
    }
    const ac = new AbortController()
    const tailAfter = status === 'running' ? 0 : eventsOffset
    subscribeSessionEvents(runner, sessionId, tailAfter, (ev) => {
      setEvents((prev) => [...prev, ev])
    }, ac.signal)
    const abortTimer = window.setTimeout(() => ac.abort(), 1500)
    return () => {
      window.clearTimeout(abortTimer)
      ac.abort()
    }
  }, [ready, needsAuth, runner, sessionId, eventsOffset, status, events])

  const handleSend = async () => {
    const text = draft.trim()
    if (!text || sending || !runner || !sessionId) return
    setSending(true)
    try {
      const ok = await sendSessionMessage(runner, sessionId, text)
      if (ok) {
        setDraft('')
        await refresh()
      }
    } finally {
      setSending(false)
    }
  }

  if (needsAuth) return <AuthPage />
  if (!ready || !runner || !sessionId) {
    return (
      <Shell>
        <div className="main-panel" />
        <Composer value="" onChange={() => {}} onSend={() => {}} sending />
      </Shell>
    )
  }

  const isRunning = status === 'running'
  const timeline = coalesceTimeline(events)

  return (
    <Shell>
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
        <div className="message-list" data-testid="message-list">
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