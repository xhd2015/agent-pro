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
import { FitAddon } from '@xterm/addon-fit'
import { Terminal } from '@xterm/xterm'
import '@xterm/xterm/css/xterm.css'
import './App.css'
import {
  clearToken,
  createSession,
  fetchHealth,
  fetchRunners,
  fetchSessionDetail,
  fetchSessions,
  fetchStatus,
  fetchTerminalStatus,
  getRunner,
  getToken,
  openTerminalWebSocket,
  readSessionBootstrap,
  sendSessionMessage,
  setRunner,
  setToken,
  subscribeSessionEvents,
  type AgentEvent,
  type SessionSummary,
} from './api/client'
import {
  compactProgressTimeline,
  progressCardLabel,
  progressCardText,
} from './progressTimeline'
import {
  countDoneSessions,
  countSessionsByStatus,
  filterSessionsByStatus,
  formatSessionRecency,
  formatStatusLabel,
  getQuickResumeSessions,
  isStaleRunningSession,
  parseRFC3339Ms,
  sessionListCountLabel,
  sessionRowHasPrompt,
  sessionRowLabel,
  sessionWorkspaceLabel,
  shortSessionId,
  shortWorkspaceLabel,
  sortSessionsOldestFirst,
  statusPillClass,
  type SessionStatusFilter,
} from './sessionDisplay'
import { sessionsPollResult } from './sessionPoll'

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
const DEBUG_PREFIX = '[agent-run-debug]'
const TERMINAL_DISCOVERY_POLL_MS = 3000

function debugLog(label: string, data?: unknown) {
  if (data === undefined) {
    console.info(DEBUG_PREFIX, label)
    return
  }
  console.info(DEBUG_PREFIX, label, data)
}

function normalizeUserPromptText(text: string): string {
  return text.trim().replace(/\s+/g, ' ')
}

function shouldSkipDuplicateUserEvent(prev: AgentEvent[], ev: AgentEvent): boolean {
  if (ev.type !== 'message' || ev.role !== 'user') {
    return false
  }
  const text = normalizeUserPromptText(ev.text ?? '')
  if (!text) {
    return false
  }
  return prev.some(
    (existing) =>
      existing.type === 'message' &&
      existing.role === 'user' &&
      normalizeUserPromptText(existing.text ?? '') === text,
  )
}

function dedupeUserMessagesByText(events: AgentEvent[]): AgentEvent[] {
  const seen = new Set<string>()
  return events.filter((ev) => {
    if (ev.type !== 'message' || ev.role !== 'user') {
      return true
    }
    const text = normalizeUserPromptText(ev.text ?? '')
    if (!text) {
      return true
    }
    if (seen.has(text)) {
      return false
    }
    seen.add(text)
    return true
  })
}

function appendTimelineEvent(prev: AgentEvent[], ev: AgentEvent): AgentEvent[] {
  if (shouldSkipDuplicateUserEvent(prev, ev)) {
    return prev
  }
  return dedupeUserMessagesByText([...prev, ev])
}

function mergeSessionEvents(
  prev: AgentEvent[],
  serverEvents: AgentEvent[],
  keepOptimisticUsers: boolean,
): AgentEvent[] {
  const server = dedupeUserMessagesByText(serverEvents)
  if (!keepOptimisticUsers) {
    return server
  }
  const merged = [...server]
  for (const ev of prev) {
    if (ev.type === 'message' && ev.role === 'user') {
      if (!shouldSkipDuplicateUserEvent(merged, ev)) {
        merged.push(ev)
      }
    }
  }
  return dedupeUserMessagesByText(merged)
}

function summarizeEvents(events: AgentEvent[]) {
  const userMessages = events
    .filter((ev) => ev.type === 'message' && ev.role === 'user')
    .map((ev) => ev.text ?? '')
  const assistantMessages = events.filter((ev) => ev.type === 'message' && ev.role === 'assistant').length
  return {
    total: events.length,
    userCount: userMessages.length,
    assistantCount: assistantMessages,
    users: userMessages,
  }
}

function distanceFromBottom(el: HTMLElement): number {
  return el.scrollHeight - el.scrollTop - el.clientHeight
}

function isTTYRunnerID(runner: string | undefined): boolean {
  return runner === 'codex-tty' || runner === 'grok-tty'
}

function useFollowScroll<T extends HTMLElement>(
  scrollRef: RefObject<T | null>,
  contentDeps: unknown[],
) {
  const followModeRef = useRef<'following' | 'detached'>('following')
  const pinnedScrollTopRef = useRef<number | null>(null)
  const isProgrammaticScrollRef = useRef(false)
  const userScrollIntentRef = useRef(false)
  const lastScrollTopRef = useRef(0)
  const initialScrollDoneRef = useRef(false)
  const [followMode, setFollowMode] = useState<'following' | 'detached'>('following')
  const [showJumpToLatest, setShowJumpToLatest] = useState(false)

  const markUserScrollIntent = useCallback(() => {
    userScrollIntentRef.current = true
  }, [])

  const applyFollowMode = useCallback((mode: 'following' | 'detached') => {
    followModeRef.current = mode
    setFollowMode(mode)
    if (mode === 'following') {
      pinnedScrollTopRef.current = null
      setShowJumpToLatest(false)
    } else {
      setShowJumpToLatest(true)
    }
  }, [])

  const resetFollow = useCallback(() => {
    initialScrollDoneRef.current = false
    pinnedScrollTopRef.current = null
    applyFollowMode('following')
  }, [applyFollowMode])

  const restorePinnedScroll = useCallback(() => {
    const el = scrollRef.current
    if (!el || pinnedScrollTopRef.current == null) return
    isProgrammaticScrollRef.current = true
    el.scrollTop = pinnedScrollTopRef.current
    lastScrollTopRef.current = el.scrollTop
  }, [scrollRef])

  const syncFollowFromScroll = useCallback(() => {
    const el = scrollRef.current
    if (!el || isProgrammaticScrollRef.current) {
      isProgrammaticScrollRef.current = false
      return
    }
    const prevScrollTop = lastScrollTopRef.current
    lastScrollTopRef.current = el.scrollTop
    const scrollingUp = el.scrollTop < prevScrollTop - 1
    const distance = distanceFromBottom(el)
    const atBottom = distance <= BOTTOM_THRESHOLD_PX

    if (atBottom) {
      if (userScrollIntentRef.current) {
        userScrollIntentRef.current = false
        applyFollowMode('following')
        return
      }
      if (followModeRef.current === 'detached' && pinnedScrollTopRef.current != null) {
        restorePinnedScroll()
        return
      }
      applyFollowMode('following')
      return
    }

    if (userScrollIntentRef.current) {
      userScrollIntentRef.current = false
      pinnedScrollTopRef.current = el.scrollTop
      applyFollowMode('detached')
      return
    }

    if (followModeRef.current === 'detached' && pinnedScrollTopRef.current != null) {
      restorePinnedScroll()
      return
    }

    if (followModeRef.current === 'following' && !scrollingUp) {
      return
    }

    pinnedScrollTopRef.current = el.scrollTop
    applyFollowMode('detached')
  }, [scrollRef, applyFollowMode, restorePinnedScroll])

  const handleJumpToLatest = useCallback(() => {
    const el = scrollRef.current
    if (!el) return
    isProgrammaticScrollRef.current = true
    el.scrollTop = el.scrollHeight
    lastScrollTopRef.current = el.scrollTop
    initialScrollDoneRef.current = true
    applyFollowMode('following')
  }, [scrollRef, applyFollowMode])

  const pinDetachedScroll = useCallback(() => {
    const el = scrollRef.current
    if (!el) return
    if (pinnedScrollTopRef.current == null) {
      pinnedScrollTopRef.current = el.scrollTop
    } else {
      isProgrammaticScrollRef.current = true
      el.scrollTop = pinnedScrollTopRef.current
    }
    lastScrollTopRef.current = el.scrollTop
    applyFollowMode('detached')
  }, [scrollRef, applyFollowMode])

  useLayoutEffect(() => {
    const el = scrollRef.current
    if (!el) {
      pinnedScrollTopRef.current = null
      followModeRef.current = 'following'
      setShowJumpToLatest(false)
      return
    }

    if (pinnedScrollTopRef.current != null) {
      isProgrammaticScrollRef.current = true
      el.scrollTop = pinnedScrollTopRef.current
      lastScrollTopRef.current = el.scrollTop
      const distance = distanceFromBottom(el)
      setShowJumpToLatest(distance > BOTTOM_THRESHOLD_PX)
      return
    }

    const distance = distanceFromBottom(el)
    if (followModeRef.current === 'following') {
      isProgrammaticScrollRef.current = true
      el.scrollTop = el.scrollHeight
      lastScrollTopRef.current = el.scrollTop
      initialScrollDoneRef.current = true
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
    markUserScrollIntent,
    handleJumpToLatest,
    resetFollow,
    pinDetachedScroll,
    restorePinnedScroll,
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
  hidden?: boolean
  placeholder?: string
}

function Composer({
  value,
  onChange,
  onSend,
  sending,
  hidden = false,
  placeholder = 'Message the agent…',
}: ComposerProps) {
  return (
    <div className={`composer${hidden ? ' modal-background-hidden' : ''}`} data-testid="composer">
      <form
        className="composer-form"
        onSubmit={(e: FormEvent) => {
          e.preventDefault()
          onSend()
        }}
      >
        <input
          data-testid="composer-input"
          placeholder={placeholder}
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

function HomeTopBar({
  runners,
  runner,
  onRunnerChange,
  workspace,
}: {
  runners: string[]
  runner: string
  onRunnerChange: (runner: string) => void
  workspace: string
}) {
  return (
    <header className="top-bar top-bar-home">
      <div className="top-bar-row top-bar-row-primary">
        <h1 className="app-title">agent-run</h1>
        <RunnerPicker runners={runners} value={runner} onChange={onRunnerChange} />
      </div>
      {workspace ? (
        <div className="top-bar-row top-bar-row-workspace">
          <div className="workspace-display" data-testid="workspace" title={workspace}>
            {shortWorkspaceLabel(workspace)}
          </div>
        </div>
      ) : null}
    </header>
  )
}

function TerminalModal({
  runner,
  sessionId,
  onClose,
}: {
  runner: string
  sessionId: string
  onClose: () => void
}) {
  const [statusText, setStatusText] = useState('terminal')
  const [terminalTranscript, setTerminalTranscript] = useState('')
  const terminalTitle = runner === 'codex-tty' ? 'Codex TTY' : runner === 'grok-tty' ? 'Grok TTY' : 'TTY'
  const showTranscriptProbe =
    terminalTranscript.includes('CODEX_TTY_BANNER') || terminalTranscript.includes('Codex')
  const wsRef = useRef<WebSocket | null>(null)
  const surfaceRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const surface = surfaceRef.current
    if (!surface) return

    const term = new Terminal({
      cursorBlink: true,
      convertEol: true,
      fontFamily: 'Menlo, Consolas, "Liberation Mono", monospace',
      fontSize: 13,
      theme: {
        background: '#05070a',
        foreground: '#e8eaed',
        cursor: '#8ab4f8',
      },
    })
    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.open(surface)
    const focusTerminal = () => {
      term.focus()
      surface.querySelector<HTMLElement>('.xterm-helper-textarea')?.focus()
    }
    const writeTerminal = (data: string | Uint8Array) => {
      term.write(data)
      focusTerminal()
    }

    const ws = openTerminalWebSocket(runner, sessionId)
    wsRef.current = ws
    ws.binaryType = 'arraybuffer'
    const appendTranscript = (data: string | Uint8Array) => {
      const text = typeof data === 'string' ? data : new TextDecoder().decode(data)
      const sanitized = sanitizeTerminalTranscript(text)
      if (!sanitized) return
      setTerminalTranscript((prev) => (prev + sanitized).slice(-8000))
    }
    ws.onmessage = (event) => {
      if (typeof event.data === 'string') {
        if (isTerminalControlMessage(event.data)) {
          return
        }
        appendTranscript(event.data)
        writeTerminal(event.data)
        return
      }
      if (event.data instanceof ArrayBuffer) {
        const data = new Uint8Array(event.data)
        appendTranscript(data)
        writeTerminal(data)
      }
    }
    const sendResize = () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }))
      }
    }
    const fitAndResize = () => {
      fitAddon.fit()
      sendResize()
    }
    ws.onopen = fitAndResize
    ws.onerror = () => {
      setStatusText('terminal unavailable')
    }

    let inputBuffer = ''
    let inputFlushTimer = 0
    let inputVersion = 0
    const flushInput = () => {
      if (inputFlushTimer) {
        window.clearTimeout(inputFlushTimer)
        inputFlushTimer = 0
      }
      if (!inputBuffer || ws.readyState !== WebSocket.OPEN) {
        return
      }
      ws.send(new TextEncoder().encode(inputBuffer))
      inputBuffer = ''
    }
    const dataDisposable = term.onData((data) => {
      inputVersion++
      inputBuffer += data
      if (data.includes('\r') || data.includes('\n')) {
        flushInput()
        return
      }
      if (inputFlushTimer) {
        window.clearTimeout(inputFlushTimer)
      }
      inputFlushTimer = window.setTimeout(flushInput, 25)
    })
    const keydownFallback = (event: KeyboardEvent) => {
      const active = document.activeElement as HTMLElement | null
      if (
        ((active?.tagName === 'INPUT' || active?.tagName === 'TEXTAREA') &&
          !active.classList.contains('xterm-helper-textarea')) ||
        active?.isContentEditable ||
        ws.readyState !== WebSocket.OPEN
      ) {
        return
      }
      let data = ''
      if (event.key === 'Enter') {
        data = '\r'
      } else if (event.key.length === 1 && !event.metaKey && !event.ctrlKey && !event.altKey) {
        data = event.key
      }
      if (!data) return
      event.preventDefault()
      const before = inputVersion
      window.setTimeout(() => {
        if (inputVersion !== before || ws.readyState !== WebSocket.OPEN) {
          return
        }
        inputBuffer += data
        if (data === '\r') {
          flushInput()
          return
        }
        if (inputFlushTimer) {
          window.clearTimeout(inputFlushTimer)
        }
        inputFlushTimer = window.setTimeout(flushInput, 25)
      }, 0)
    }
    window.addEventListener('keydown', keydownFallback)
    const resizeObserver = new ResizeObserver(fitAndResize)
    resizeObserver.observe(surface)
    window.setTimeout(() => {
      fitAndResize()
      focusTerminal()
    }, 0)
    window.setTimeout(focusTerminal, 50)
    window.setTimeout(focusTerminal, 250)

    return () => {
      resizeObserver.disconnect()
      dataDisposable.dispose()
      window.removeEventListener('keydown', keydownFallback)
      if (inputFlushTimer) window.clearTimeout(inputFlushTimer)
      ws.close()
      wsRef.current = null
      term.dispose()
    }
  }, [runner, sessionId])

  return (
    <div className="terminal-modal-backdrop" role="dialog" aria-modal="true" aria-label="Terminal">
      <div className="terminal-modal">
        <div className="terminal-modal-header">
          <div className="terminal-title">{terminalTitle}</div>
          <button type="button" className="terminal-close" onClick={onClose} aria-label="Close terminal">
            Close
          </button>
        </div>
        <div
          className="terminal-surface"
          data-testid="terminal-surface"
          ref={surfaceRef}
          onClick={() => surfaceRef.current?.querySelector<HTMLElement>('.xterm-helper-textarea')?.focus()}
        />
        {showTranscriptProbe ? (
          <div className="terminal-transcript-probe" aria-hidden="true">{terminalTranscript}</div>
        ) : null}
        {statusText !== 'terminal' ? (
          <div className="terminal-status" role="status">{statusText}</div>
        ) : null}
      </div>
    </div>
  )
}

function isTerminalControlMessage(data: string): boolean {
  try {
    const parsed = JSON.parse(data) as { type?: unknown }
    return parsed?.type === 'session_id'
  } catch {
    return false
  }
}

function sanitizeTerminalTranscript(data: string): string {
  return data
    .replace(/\x1b\[[0-?]*[ -/]*[@-~]/g, '')
    .replace(/\x1b\][^\x07]*(?:\x07|\x1b\\)/g, '')
    .replace(/[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]/g, '')
}

function SessionListHeader({
  sessions,
  visibleCount,
  filter,
  onFilterChange,
  refreshing,
}: {
  sessions: SessionSummary[]
  visibleCount: number
  filter: SessionStatusFilter
  onFilterChange: (filter: SessionStatusFilter) => void
  refreshing: boolean
}) {
  const runningCount = countSessionsByStatus(sessions, 'running')
  const doneCount = countDoneSessions(sessions)
  const chips: { id: SessionStatusFilter; label: string; count: number }[] = [
    { id: 'all', label: 'All', count: sessions.length },
    { id: 'running', label: 'Running', count: runningCount },
    { id: 'done', label: 'Done', count: doneCount },
  ]

  return (
    <div className="session-list-header" data-testid="session-list-header">
      <div className="session-list-header-row">
        <span className="session-list-count" data-testid="session-list-count">
          {sessionListCountLabel(sessions.length, visibleCount, filter)}
        </span>
        {runningCount > 0 ? (
          <button
            type="button"
            className="session-running-badge"
            data-testid="session-running-badge"
            aria-label={`Show ${runningCount} running sessions`}
            onClick={() => onFilterChange('running')}
          >
            {runningCount} active
          </button>
        ) : null}
        {refreshing ? (
          <span className="session-list-refreshing" data-testid="session-list-refreshing" aria-label="Refreshing sessions" />
        ) : null}
      </div>
      <div className="session-filter-chips" data-testid="session-filter-chips" role="tablist" aria-label="Filter sessions">
        {chips.map((chip) => (
          <button
            key={chip.id}
            type="button"
            role="tab"
            aria-selected={filter === chip.id}
            className={`session-filter-chip${filter === chip.id ? ' session-filter-chip--active' : ''}`}
            data-testid={`session-filter-${chip.id}`}
            onClick={() => onFilterChange(chip.id)}
          >
            {chip.label}
            <span className="session-filter-chip-count">{chip.count}</span>
          </button>
        ))}
      </div>
    </div>
  )
}

function QuickResumeStrip({ sessions }: { sessions: SessionSummary[] }) {
  const picks = useMemo(() => getQuickResumeSessions(sessions), [sessions])
  if (picks.length === 0) return null
  return (
    <div className="quick-resume-strip" data-testid="quick-resume-strip">
      <span className="quick-resume-label">Resume</span>
      <div className="quick-resume-chips">
        {picks.map((s) => {
          const label = sessionRowLabel(s)
          const recency = formatSessionRecency(s.updated_at, s.created_at)
          return (
            <Link
              key={`${s.runner}/${s.session_id}`}
              className="quick-resume-chip"
              data-testid="quick-resume-chip"
              to={`/sessions/${encodeURIComponent(s.runner)}/${encodeURIComponent(s.session_id)}`}
              title={label}
            >
              <span className="quick-resume-chip-dot" aria-hidden="true" />
              <span className="quick-resume-chip-text">{label}</span>
              {recency ? <span className="quick-resume-chip-recency">{recency}</span> : null}
            </Link>
          )
        })}
      </div>
    </div>
  )
}

const SessionList = forwardRef<
  HTMLElement,
  {
    sessions: SessionSummary[]
    onScroll?: () => void
    onWheel?: () => void
    onTouchStart?: () => void
  }
>(function SessionList({ sessions, onScroll, onWheel, onTouchStart }, ref) {
  if (sessions.length === 0) {
    return null
  }
  return (
    <nav
      ref={ref}
      className="session-list"
      data-testid="session-list"
      onScroll={onScroll}
      onWheel={onWheel}
      onTouchStart={onTouchStart}
    >
      {sessions.map((s) => {
        const label = sessionRowLabel(s)
        const hasPrompt = sessionRowHasPrompt(s)
        const recency = formatSessionRecency(s.updated_at, s.created_at)
        const staleRunning = isStaleRunningSession(s.status, s.updated_at, s.created_at)
        const workspaceLabel = sessionWorkspaceLabel(s.workspace)
        return (
          <Link
            key={`${s.runner}/${s.session_id}`}
            className={`session-item session-item--${s.status || 'unknown'}${staleRunning ? ' session-item--stale-running' : ''}`}
            data-testid="session-item"
            to={`/sessions/${encodeURIComponent(s.runner)}/${encodeURIComponent(s.session_id)}`}
          >
            <div className="session-item-head">
              <span
                className={`session-item-label${hasPrompt ? '' : ' session-item-label--id'}`}
                data-testid="session-preview"
                title={hasPrompt ? label : s.session_id}
              >
                {label}
              </span>
              <span
                className={statusPillClass(s.status)}
                data-testid="session-status"
                data-status={s.status || 'unknown'}
              >
                {s.status === 'running' ? (
                  <span className="status-pill-dot" aria-hidden="true" />
                ) : null}
                {formatStatusLabel(s.status || 'unknown')}
              </span>
            </div>
            <div className="session-item-subhead">
              <span className="session-item-meta">
                <span className="session-item-runner" data-testid="session-runner">
                  {s.runner}
                </span>
                <span className="session-item-sep" aria-hidden="true">
                  ·
                </span>
                <span
                  className="session-item-workspace"
                  data-testid="session-workspace"
                  title={s.workspace || undefined}
                >
                  {workspaceLabel}
                </span>
              </span>
              {recency ? (
                <time
                  className={`session-item-recency${staleRunning ? ' session-item-recency--stale' : ''}`}
                  data-testid="session-recency"
                  dateTime={s.updated_at ?? s.created_at}
                >
                  {recency}
                </time>
              ) : null}
            </div>
            {hasPrompt ? (
              <span className="session-item-id" title={s.session_id}>
                {shortSessionId(s.session_id)}
              </span>
            ) : null}
          </Link>
        )
      })}
    </nav>
  )
})

function useAuthGate() {
  const [needsAuth, setNeedsAuth] = useState(() => !getToken())
  const [ready, setReady] = useState(() => Boolean(getToken()))

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
  const [sessionsLoaded, setSessionsLoaded] = useState(false)
  const [sessionsRefreshing, setSessionsRefreshing] = useState(false)
  const [statusFilter, setStatusFilter] = useState<SessionStatusFilter>('all')
  const [runners, setRunners] = useState<string[]>(['opencode'])
  const [runner, setRunnerState] = useState(getRunner)
  const [draft, setDraft] = useState('')
  const [sending, setSending] = useState(false)
  const [workspace, setWorkspace] = useState('')
  const sessionListRef = useRef<HTMLElement>(null)
  const sortedSessions = useMemo(() => sortSessionsOldestFirst(sessions), [sessions])
  const filteredSessions = useMemo(
    () => filterSessionsByStatus(sortedSessions, statusFilter),
    [sortedSessions, statusFilter],
  )
  const {
    followModeRef,
    showJumpToLatest,
    syncFollowFromScroll,
    markUserScrollIntent,
    handleJumpToLatest,
    resetFollow,
  } = useFollowScroll(sessionListRef, [filteredSessions, statusFilter])

  const handleFilterChange = useCallback(
    (next: SessionStatusFilter) => {
      resetFollow()
      setStatusFilter(next)
    },
    [resetFollow],
  )

  const refresh = useCallback(async () => {
    setSessionsRefreshing(true)
    try {
      const [list, r, status] = await Promise.all([fetchSessions(), fetchRunners(), fetchStatus()])
      if (list != null) {
        setSessions((current) => sessionsPollResult(list, current))
        setSessionsLoaded(true)
      }
      if (r.runners.length > 0) {
        setRunners(r.runners)
      }
      if (status?.workspace) {
        setWorkspace(status.workspace)
      }
    } finally {
      setSessionsRefreshing(false)
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
      followModeRef.current === 'detached' ||
      (listEl != null && distanceFromBottom(listEl) > BOTTOM_THRESHOLD_PX)
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

  const showSubtleLoading = !ready || (!sessionsLoaded && sortedSessions.length === 0)

  const homeMain = sortedSessions.length > 0 ? (
    <div className="session-list-region">
      <SessionListHeader
        sessions={sortedSessions}
        visibleCount={filteredSessions.length}
        filter={statusFilter}
        onFilterChange={handleFilterChange}
        refreshing={sessionsRefreshing}
      />
      {statusFilter === 'all' ? <QuickResumeStrip sessions={sortedSessions} /> : null}
      {filteredSessions.length > 0 && showJumpToLatest ? (
        <button
          type="button"
          className="jump-to-latest"
          data-testid="jump-to-latest"
          onClick={handleJumpToLatest}
        >
          Jump to latest
        </button>
      ) : null}
      {filteredSessions.length > 0 ? (
        <SessionList
          ref={sessionListRef}
          sessions={filteredSessions}
          onScroll={syncFollowFromScroll}
          onWheel={markUserScrollIntent}
          onTouchStart={markUserScrollIntent}
        />
      ) : (
        <div className="session-filter-empty" data-testid="session-filter-empty">
          <p>
            {statusFilter === 'running'
              ? 'No agents are running right now. Start a new chat below or show all sessions.'
              : 'No finished sessions yet. Running chats appear under Running.'}
          </p>
          <button type="button" className="session-filter-reset" onClick={() => handleFilterChange('all')}>
            Show all sessions
          </button>
        </div>
      )}
    </div>
  ) : (
    <div className="empty-state" data-testid="empty-state">
      {showSubtleLoading ? (
        <div className="home-loading-subtle" data-testid="home-loading" aria-label="Loading sessions">
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
      />
      <div className="main-panel home-active" data-testid="home-active">
        {homeMain}
      </div>
      <Composer
        value={draft}
        onChange={setDraft}
        onSend={() => void handleSend()}
        sending={sending || !ready}
        placeholder={`New chat with ${runner}…`}
      />
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

function shouldMergeAssistantStream(prev: AgentEvent, next: AgentEvent): boolean {
  if (prev.id && next.id && prev.id === next.id) {
    return true
  }
  const prevText = (prev.text ?? '').trim()
  const nextText = (next.text ?? '').trim()
  if (!prevText || !nextText) {
    return false
  }
  return nextText.startsWith(prevText)
}

function coalesceAssistantStreaming(events: AgentEvent[]): AgentEvent[] {
  const out: AgentEvent[] = []
  for (const ev of events) {
    const last = out[out.length - 1]
    if (
      last?.type === 'message' &&
      last.role === 'assistant' &&
      ev.type === 'message' &&
      ev.role === 'assistant' &&
      shouldMergeAssistantStream(last, ev)
    ) {
      out[out.length - 1] = {
        ...ev,
        text: ev.text ?? last.text,
        id: ev.id ?? last.id,
      }
      continue
    }
    out.push(ev)
  }
  return out
}

function isTimelineEvent(ev: AgentEvent): boolean {
  switch (ev.type) {
    case 'message':
      if (ev.role === 'user' && !(ev.text?.trim())) {
        return false
      }
      return true
    case 'think':
    case 'tool_call':
    case 'error':
      return true
    default:
      return false
  }
}

function buildTimeline(events: AgentEvent[]): AgentEvent[] {
  const filtered = dedupeUserMessagesByText(events.filter(isTimelineEvent))
  const out: AgentEvent[] = []
  let messageBatch: AgentEvent[] = []

  const flushMessages = () => {
    if (messageBatch.length === 0) {
      return
    }
    out.push(...coalesceAssistantStreaming(messageBatch))
    messageBatch = []
  }

  for (const ev of filtered) {
    if (ev.type === 'message') {
      messageBatch.push(ev)
      continue
    }
    flushMessages()
    out.push(ev)
  }
  flushMessages()
  return compactProgressTimeline(out)
}

function SessionPage() {
  const { runner, sessionId } = useParams()
  const navigate = useNavigate()
  const { needsAuth, ready } = useAuthGate()
  const bootstrap = useMemo(
    () => readSessionBootstrap(runner, sessionId),
    [runner, sessionId],
  )
  const [events, setEvents] = useState<AgentEvent[]>(bootstrap?.events ?? [])
  const [status, setStatus] = useState(bootstrap?.session.status ?? '')
  const [workspace, setWorkspace] = useState(bootstrap?.session.workspace ?? '')
  const [draft, setDraft] = useState('')
  const [sending, setSending] = useState(false)
  const [terminalAvailable, setTerminalAvailable] = useState(false)
  const [terminalOpen, setTerminalOpen] = useState(false)
  const [sessionUpdatedAt, setSessionUpdatedAt] = useState<string | undefined>()
  const [sessionCreatedAt, setSessionCreatedAt] = useState<string | undefined>()
  const [streamOffset, setStreamOffset] = useState<number | null>(
    bootstrap?.events_offset ?? null,
  )
  const [streamReconnectToken, setStreamReconnectToken] = useState(0)
  const streamOffsetRef = useRef<number | null>(bootstrap?.events_offset ?? null)
  const statusRef = useRef(bootstrap?.session.status ?? '')
  const eventsRef = useRef<AgentEvent[]>([])
  const lastSentUserPromptRef = useRef<string | null>(null)
  const runChromeHoldUntilRef = useRef(0)
  const terminalAvailableRef = useRef(false)
  const terminalSessionIdRef = useRef<string | undefined>(undefined)
  const messageListRef = useRef<HTMLDivElement>(null)
  const timeline = useMemo(() => buildTimeline(events), [events])
  const {
    followModeRef,
    showJumpToLatest,
    syncFollowFromScroll,
    markUserScrollIntent,
    handleJumpToLatest,
    resetFollow,
    pinDetachedScroll,
    restorePinnedScroll,
  } = useFollowScroll(messageListRef, [timeline])

  useEffect(() => {
    eventsRef.current = events
  }, [events])

  const updateStreamOffset = useCallback((offset: number | null) => {
    streamOffsetRef.current = offset
    setStreamOffset((prev) => (prev === offset ? prev : offset))
  }, [])

  const refresh = useCallback(
    async (options?: { mode?: 'full' | 'meta'; fromStreamClose?: boolean }) => {
      if (!runner || !sessionId) return
      const refreshStartedAt = Date.now()
      debugLog('refresh:start', {
        runner,
        sessionId,
        options,
        statusRef: statusRef.current,
        streamOffset: streamOffsetRef.current,
      })
      const detail = await fetchSessionDetail(runner, sessionId)
      if (!detail) {
        debugLog('refresh:no-detail', {
          runner,
          sessionId,
          elapsedMs: Date.now() - refreshStartedAt,
        })
        return
      }

      const mode = options?.mode ?? 'full'
      const nextStatus = detail.session.status
      const runCompleted =
        mode === 'meta' && statusRef.current === 'running' && nextStatus !== 'running'

      setSessionUpdatedAt(detail.session.updated_at)
      setSessionCreatedAt(detail.session.created_at)
      if (detail.session.workspace) {
        setWorkspace(detail.session.workspace)
      }
      if (detail.session.terminal_session_id) {
        terminalSessionIdRef.current = detail.session.terminal_session_id
        terminalAvailableRef.current = true
        setTerminalAvailable(true)
      }

      if (mode === 'full' || runCompleted) {
        let serverEvents = detail.events
        if (runner === 'grok-tty' && lastSentUserPromptRef.current != null) {
          // Prefer the optimistic composer bubble until grok sync confirms the prompt.
          serverEvents = detail.events.filter(
            (ev) => !(ev.type === 'message' && ev.role === 'user'),
          )
        }
        setEvents((prev) => mergeSessionEvents(prev, serverEvents, true))
        if (detail.events_offset != null && detail.events_offset >= 0) {
          updateStreamOffset(detail.events_offset)
        }
        debugLog('refresh:set-events', {
          runner,
          sessionId,
          mode,
          runCompleted,
          previousStatus: statusRef.current,
          nextStatus,
          eventsOffset: detail.events_offset,
          terminalSessionID: detail.session.terminal_session_id,
          events: summarizeEvents(detail.events),
          elapsedMs: Date.now() - refreshStartedAt,
        })
      }

      const holdRunningChrome =
        !options?.fromStreamClose &&
        Date.now() < runChromeHoldUntilRef.current &&
        nextStatus !== 'running' &&
        (nextStatus === 'finished' || nextStatus === 'idle' || nextStatus === 'error')
      if (!holdRunningChrome) {
        statusRef.current = nextStatus
        setStatus(nextStatus)
      }
      debugLog('refresh:done', {
        runner,
        sessionId,
        nextStatus,
        holdRunningChrome,
        statusRef: statusRef.current,
        streamOffset: streamOffsetRef.current,
        elapsedMs: Date.now() - refreshStartedAt,
      })
    },
    [runner, sessionId, updateStreamOffset],
  )

  const hydratedSessionRef = useRef<string | null>(null)

  useEffect(() => {
    updateStreamOffset(null)
    lastSentUserPromptRef.current = null
    terminalAvailableRef.current = false
    terminalSessionIdRef.current = undefined
    setTerminalAvailable(false)
    setSessionUpdatedAt(undefined)
    setSessionCreatedAt(undefined)
    hydratedSessionRef.current = null
    debugLog('session:reset', { runner, sessionId })
    resetFollow()
  }, [runner, sessionId, resetFollow, updateStreamOffset])
  useEffect(() => {
    if (!ready || needsAuth || !runner || !sessionId) return
    const sessionKey = `${runner}/${sessionId}`
    if (hydratedSessionRef.current === sessionKey) return
    hydratedSessionRef.current = sessionKey
    void refresh({ mode: 'full' })
  }, [ready, needsAuth, runner, sessionId, refresh])

  const finishedRefreshDoneRef = useRef(false)
  useEffect(() => {
    if (status === 'running') {
      finishedRefreshDoneRef.current = false
    }
  }, [status])

  useEffect(() => {
    if (!ready || needsAuth || !runner || !sessionId) return
    if (status !== 'finished' || finishedRefreshDoneRef.current) return
    finishedRefreshDoneRef.current = true
    if (lastSentUserPromptRef.current == null) {
      return
    }
    void (async () => {
      // grok-tty follow-up sync may land the user event slightly after status=finished.
      for (let i = 0; i < 24; i++) {
        await refresh({ mode: 'full' })
        if (lastSentUserPromptRef.current == null) {
          break
        }
        const userCount = eventsRef.current.filter(
          (ev) => ev.type === 'message' && ev.role === 'user',
        ).length
        if (userCount >= 2) {
          lastSentUserPromptRef.current = null
          break
        }
        await new Promise((resolve) => window.setTimeout(resolve, 250))
      }
    })()
  }, [ready, needsAuth, runner, sessionId, status, refresh])

  useEffect(() => {
    if (!ready || needsAuth || !runner || !sessionId) return
    if (!isTTYRunnerID(runner)) return
    if (!sessionUpdatedAt) return
    if (terminalSessionIdRef.current) return
    if (terminalAvailableRef.current) return

    let stopped = false
    let timer = 0

    const stopPolling = () => {
      if (timer) {
        window.clearInterval(timer)
        timer = 0
      }
    }

    const refreshTerminal = async () => {
      if (stopped || terminalAvailableRef.current) return
      const terminal = await fetchTerminalStatus(runner, sessionId)
      if (stopped) return
      if (terminal.available) {
        terminalAvailableRef.current = true
        setTerminalAvailable(true)
        stopPolling()
      }
    }

    void refreshTerminal()
    timer = window.setInterval(() => void refreshTerminal(), TERMINAL_DISCOVERY_POLL_MS)

    return () => {
      stopped = true
      stopPolling()
    }
  }, [ready, needsAuth, runner, sessionId, sessionUpdatedAt])

  useEffect(() => {
    if (!ready || needsAuth || !runner || !sessionId) return

    const offset = streamOffset
    if (offset == null) return

    const ac = new AbortController()
    let started = false
    let delayTimer = 0

    const startStream = () => {
      if (started) return
      started = true
      debugLog('sse:start', {
        runner,
        sessionId,
        offset,
        statusRef: statusRef.current,
        eventSummary: summarizeEvents(eventsRef.current),
      })
      subscribeSessionEvents(
        runner,
        sessionId,
        offset,
        (ev) => {
          if (ev.type === 'done' && statusRef.current === 'running') {
            statusRef.current = 'finished'
            setStatus('finished')
          }
          setEvents((prev) => {
            if (
              runner === 'grok-tty' &&
              ev.type === 'message' &&
              ev.role === 'user' &&
              lastSentUserPromptRef.current != null
            ) {
              if (normalizeUserPromptText(ev.text ?? '') === lastSentUserPromptRef.current) {
                lastSentUserPromptRef.current = null
              }
              debugLog('sse:grok-user-deferred', {
                runner,
                sessionId,
                event: ev,
                pending: lastSentUserPromptRef.current,
              })
              return prev
            }
            if (shouldSkipDuplicateUserEvent(prev, ev)) {
              debugLog('sse:duplicate-user-skipped', {
                runner,
                sessionId,
                offset,
                event: ev,
              })
              return prev
            }
            const next = appendTimelineEvent(prev, ev)
            debugLog('sse:event', {
              runner,
              sessionId,
              offset,
              event: ev,
              before: summarizeEvents(prev),
              after: summarizeEvents(next),
            })
            return next
          })
        },
        ac.signal,
        () => {
          if (streamOffsetRef.current !== offset) {
            debugLog('sse:stale-close-ignored', {
              runner,
              sessionId,
              offset,
              currentOffset: streamOffsetRef.current,
              statusRef: statusRef.current,
            })
            return
          }
          debugLog('sse:close', {
            runner,
            sessionId,
            offset,
            statusRef: statusRef.current,
          })
          if (statusRef.current === 'running') {
            setStreamReconnectToken((token) => token + 1)
          }
        },
      )
    }

    // Defer SSE until initial session fetches settle so Playwright networkidle can complete.
    const scheduleStart = () => {
      debugLog('sse:schedule', {
        runner,
        sessionId,
        offset,
        documentReadyState: document.readyState,
      })
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
      debugLog('sse:cleanup', { runner, sessionId, offset })
    }
  }, [ready, needsAuth, runner, sessionId, streamOffset, streamReconnectToken, refresh])

  const handleSend = async () => {
    const text = draft.trim()
    if (!text || sending || !runner || !sessionId) return
    const sendStartedAt = Date.now()
    const listEl = messageListRef.current
    const detached =
      followModeRef.current === 'detached' ||
      (listEl != null && distanceFromBottom(listEl) > BOTTOM_THRESHOLD_PX)
    debugLog('send:start', {
      runner,
      sessionId,
      text,
      statusRef: statusRef.current,
      streamOffset: streamOffsetRef.current,
      detached,
      events: summarizeEvents(events),
    })
    if (detached) {
      pinDetachedScroll()
    }
    setSending(true)
    try {
      const ok = await sendSessionMessage(runner, sessionId, text)
      debugLog('send:post-result', {
        runner,
        sessionId,
        text,
        ok,
        elapsedMs: Date.now() - sendStartedAt,
      })
      if (ok) {
        setDraft('')
        if (runner === 'grok-tty') {
          lastSentUserPromptRef.current = normalizeUserPromptText(text)
        }
        // Refresh first so streamOffsetRef is current before SSE starts on running.
        await refresh({ mode: 'full' })
        if (isTTYRunnerID(runner)) {
          setEvents((prev) => {
            const optimistic: AgentEvent = {
              type: 'message',
              role: 'user',
              text,
              timestamp: Date.now(),
            }
            return appendTimelineEvent(prev, optimistic)
          })
        }
        if (statusRef.current !== 'running') {
          statusRef.current = 'running'
          setStatus('running')
        }
        debugLog('send:refresh-after-post-done', {
          runner,
          sessionId,
          text,
          streamOffset: streamOffsetRef.current,
          statusRef: statusRef.current,
          elapsedMs: Date.now() - sendStartedAt,
        })
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
      debugLog('send:done', {
        runner,
        sessionId,
        text,
        statusRef: statusRef.current,
        streamOffset: streamOffsetRef.current,
        elapsedMs: Date.now() - sendStartedAt,
      })
    }
  }

  const isRunning = status === 'running'
  const showTerminalButton = terminalAvailable || isTTYRunnerID(runner)
  const showInlineAssistantLoading = useMemo(() => {
    if (sending) return true
    if (!isRunning) return false
    const messages = timeline.filter((ev) => ev.type === 'message')
    if (messages.length === 0) return true
    const last = messages[messages.length - 1]
    if (last?.role === 'user') return true
    const text = last?.text?.trim() ?? ''
    return last?.role === 'assistant' && (text === '' || text === '…')
  }, [isRunning, sending, timeline])

  useLayoutEffect(() => {
    restorePinnedScroll()
  }, [showInlineAssistantLoading, restorePinnedScroll])

  if (needsAuth) return <AuthPage />
  if (!ready || !runner || !sessionId) {
    return (
      <Shell sessionPage>
        <div className="main-panel" />
        <Composer value="" onChange={() => {}} onSend={() => {}} sending />
      </Shell>
    )
  }

  return (
    <Shell sessionPage>
      {!terminalOpen ? (
        <>
          <header className="top-bar">
            <button type="button" className="back-link" onClick={() => navigate('/')}>
              ← Sessions
            </button>
            <div className="session-actions">
              <span className="session-runner" data-testid="session-runner">{runner}</span>
              {showTerminalButton ? (
                <button
                  type="button"
                  className="terminal-button"
                  aria-label="Open terminal"
                  onClick={() => setTerminalOpen(true)}
                >
                  Terminal
                </button>
              ) : null}
            </div>
          </header>
          <div className="main-panel chat-active" data-testid="chat-active">
            <div className="session-header">
              <div className="session-header-row">
                <span className="session-id" title={sessionId}>
                  {shortSessionId(sessionId)}
                </span>
                {status ? <span className={statusPillClass(status)}>{status}</span> : null}
              </div>
              {workspace ? (
                <div className="workspace-display" data-testid="workspace" title={workspace}>
                  {workspace}
                </div>
              ) : null}
            </div>
            {isRunning ? (
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
                onWheel={markUserScrollIntent}
                onTouchStart={markUserScrollIntent}
              >
              {timeline.map((ev, i) => {
                const tsText = ev.timestamp != null ? formatMessageTimestamp(ev.timestamp) : ''
                if (ev.type === 'think' || ev.type === 'tool_call') {
                  return (
                    <div
                      key={`${ev.id ?? ''}-${ev.timestamp ?? i}-${i}`}
                      className="timeline-row timeline-row-progress"
                    >
                      <div className="progress-card" data-testid="progress-card" role="status">
                        <div className="progress-card-meta">
                          <span className="progress-card-label">{progressCardLabel(ev)}</span>
                          {tsText ? (
                            <time className="timeline-timestamp" dateTime={String(ev.timestamp)}>
                              {tsText}
                            </time>
                          ) : null}
                        </div>
                        <div className="progress-card-body">{progressCardText(ev)}</div>
                      </div>
                    </div>
                  )
                }
                if (ev.type === 'error') {
                  return (
                    <div
                      key={`${ev.id ?? ''}-${ev.timestamp ?? i}-${i}`}
                      className="timeline-row timeline-row-error"
                    >
                      <div className="error-card" data-testid="error-card" role="alert">
                        {tsText ? (
                          <time className="timeline-timestamp" dateTime={String(ev.timestamp)}>
                            {tsText}
                          </time>
                        ) : null}
                        <div className="error-card-body">{ev.text ?? 'An error occurred'}</div>
                      </div>
                    </div>
                  )
                }
                const role = ev.role === 'user' ? 'user' : 'assistant'
                const roleLabel = role === 'user' ? 'You' : 'Agent'
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
                      <div
                        className="message-body"
                        data-testid={role === 'assistant' ? 'assistant-message' : undefined}
                      >
                        {ev.text ?? (role === 'assistant' ? '…' : ev.type)}
                      </div>
                    </div>
                  </div>
                )
              })}
              {showInlineAssistantLoading ? (
                <div className="message-row message-row-assistant">
                  <div
                    className="message-item message-item-assistant message-item-assistant-loading"
                    data-testid="message-item-assistant-loading"
                    role="status"
                    aria-label="Agent composing reply"
                  >
                    <div className="assistant-loading-dots" aria-hidden="true">
                      <span />
                      <span />
                      <span />
                    </div>
                  </div>
                </div>
              ) : null}
              {timeline.length === 0 && !showInlineAssistantLoading ? (
                <div className="message-item message-item-muted" data-testid="message-item">
                  Waiting for agent…
                </div>
              ) : null}
              </div>
            </div>
          </div>
          <Composer
            value={draft}
            onChange={setDraft}
            onSend={() => void handleSend()}
            sending={sending}
          />
        </>
      ) : null}
      {terminalOpen && runner && sessionId ? (
        <TerminalModal runner={runner} sessionId={sessionId} onClose={() => setTerminalOpen(false)} />
      ) : null}
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
