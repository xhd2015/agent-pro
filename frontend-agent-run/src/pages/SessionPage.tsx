import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  fetchSessionDetail,
  fetchTerminalStatus,
  readSessionBootstrap,
  sendSessionMessage,
  subscribeSessionEvents,
  type AgentEvent,
} from '../api/client'
import { AgentRunningCard } from '../components/AgentRunningCard'
import { Composer } from '../components/Composer'
import { MarkdownBody } from '../components/MarkdownBody'
import { Shell } from '../components/Shell'
import { TerminalModal } from '../components/TerminalModal'
import { WorkspacePath } from '../components/WorkspacePath'
import { useFollowScroll } from '../hooks/useFollowScroll'
import { debugLog, TERMINAL_DISCOVERY_POLL_MS } from '../lib/debug'
import { BOTTOM_THRESHOLD_PX, distanceFromBottom } from '../lib/followScroll'
import {
  appendTimelineEvent,
  buildTimeline,
  formatMessageTimestamp,
  isTTYRunnerID,
  mergeSessionEvents,
  normalizeUserPromptText,
  shouldSkipDuplicateUserEvent,
  summarizeEvents,
} from '../lib/timeline'
import {
  progressCardLabel,
  progressCardText,
  sanitizeProgressText,
} from '../progressTimeline'
import {
  shortSessionId,
  statusPillClass,
} from '../sessionDisplay'
import './SessionPage.css'

export function SessionPage() {
  const { sessionId } = useParams()
  const navigate = useNavigate()
  const bootstrap = useMemo(
    () => readSessionBootstrap(sessionId),
    [sessionId],
  )
  const [runner, setRunnerMeta] = useState(bootstrap?.session.runner ?? '')
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
      if (!sessionId) return
      const refreshStartedAt = Date.now()
      debugLog('refresh:start', {
        runner,
        sessionId,
        options,
        statusRef: statusRef.current,
        streamOffset: streamOffsetRef.current,
      })
      const detail = await fetchSessionDetail(sessionId)
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

      if (detail.session.runner) {
        setRunnerMeta(detail.session.runner)
      }
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
        const pendingUser = lastSentUserPromptRef.current
        // Prefer optimistic bubble for the in-flight prompt only — never strip
        // earlier user cards (that reordered follow-ups above prior history).
        if (runner === 'grok-tty' && pendingUser != null) {
          const beforeUsers = detail.events.filter(
            (ev) => ev.type === 'message' && ev.role === 'user',
          ).length
          serverEvents = detail.events.filter(
            (ev) =>
              !(
                ev.type === 'message' &&
                ev.role === 'user' &&
                normalizeUserPromptText(ev.text ?? '') === pendingUser
              ),
          )
          const afterUsers = serverEvents.filter(
            (ev) => ev.type === 'message' && ev.role === 'user',
          ).length
          debugLog('refresh:strip-pending-user-only', {
            runner,
            sessionId,
            pendingUser,
            serverUserCountBefore: beforeUsers,
            serverUserCountAfter: afterUsers,
            serverUsersKept: serverEvents
              .filter((ev) => ev.type === 'message' && ev.role === 'user')
              .map((ev) => ev.text ?? ''),
          })
        }
        setEvents((prev) => {
          const next = mergeSessionEvents(prev, serverEvents, true)
          debugLog('refresh:merge-result', {
            runner,
            sessionId,
            pendingUser,
            prev: summarizeEvents(prev),
            server: summarizeEvents(serverEvents),
            next: summarizeEvents(next),
            prevUsers: prev
              .filter((ev) => ev.type === 'message' && ev.role === 'user')
              .map((ev) => ev.text ?? ''),
            nextUsers: next
              .filter((ev) => ev.type === 'message' && ev.role === 'user')
              .map((ev) => ev.text ?? ''),
            nextRoles: next
              .filter((ev) => ev.type === 'message')
              .map((ev) => `${ev.role}:${(ev.text ?? '').slice(0, 40)}`),
          })
          return next
        })
        if (detail.events_offset != null && detail.events_offset >= 0) {
          updateStreamOffset(detail.events_offset)
        }
        debugLog('refresh:set-events', {
          runner,
          sessionId,
          mode,
          runCompleted,
          pendingUser,
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
    if (!sessionId) return
    const sessionKey = sessionId
    if (hydratedSessionRef.current === sessionKey) return
    hydratedSessionRef.current = sessionKey
    void refresh({ mode: 'full' })
  }, [sessionId, refresh])

  const finishedRefreshDoneRef = useRef(false)
  useEffect(() => {
    if (status === 'running') {
      finishedRefreshDoneRef.current = false
    }
  }, [status])

  useEffect(() => {
    if (!sessionId) return
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
  }, [sessionId, status, refresh])

  useEffect(() => {
    if (!sessionId) return
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
      const terminal = await fetchTerminalStatus(sessionId, runner)
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
  }, [runner, sessionId, sessionUpdatedAt])

  useEffect(() => {
    if (!sessionId) return

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
              const pending = lastSentUserPromptRef.current
              const text = normalizeUserPromptText(ev.text ?? '')
              // Only suppress the in-flight optimistic prompt (avoid double bubble).
              // Earlier history user events must still apply.
              if (text === pending) {
                lastSentUserPromptRef.current = null
                debugLog('sse:grok-pending-user-confirmed', {
                  runner,
                  sessionId,
                  pending,
                  event: ev,
                })
                return prev
              }
              debugLog('sse:grok-other-user-while-pending', {
                runner,
                sessionId,
                pending,
                event: ev,
              })
              // fall through — apply non-pending user events
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
  }, [runner, sessionId, streamOffset, streamReconnectToken, refresh])

  const handleSend = async () => {
    const text = draft.trim()
    if (!text || sending || !sessionId) return
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
      const ok = await sendSessionMessage(sessionId, text)
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

  if (!sessionId) {
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
                <WorkspacePath path={workspace} className="workspace-display" />
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
                  const isThink = ev.type === 'think'
                  // Think: full sanitized text as markdown (truncation breaks markers).
                  // Tool: keep plain truncated preview (paths/logs, not prose).
                  const thinkText =
                    sanitizeProgressText(ev.text?.trim() || 'Thinking…') || 'Thinking…'
                  const body = isThink ? (
                    <MarkdownBody text={thinkText} compact className="progress-card-markdown" />
                  ) : (
                    progressCardText(ev)
                  )
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
                        <div className="progress-card-body">{body}</div>
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
                        {role === 'assistant' ? (
                          <MarkdownBody text={ev.text ?? '…'} />
                        ) : (
                          ev.text ?? ev.type
                        )}
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
