import { forwardRef, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import type { SessionSummary } from '../api/client'
import {
  formatSessionRecency,
  formatStatusLabel,
  isStaleRunningSession,
  sessionRowHasPrompt,
  sessionRowLabel,
  sessionWorkspaceLabel,
  statusPillClass,
} from '../sessionDisplay'
import './SessionList.css'

export const SessionList = forwardRef<
  HTMLElement,
  {
    sessions: SessionSummary[]
    onScroll?: () => void
    onWheel?: () => void
    onTouchStart?: () => void
    /** Called before navigating into a session (persist scroll, etc.). */
    onSessionNavigate?: () => void
    /** Rendered at the end of the scrollable list (e.g. Load more). */
    footer?: ReactNode
    /** User home for ~/ compact workspace labels (from status.home). */
    homeDir?: string
  }
>(function SessionList(
  { sessions, onScroll, onWheel, onTouchStart, onSessionNavigate, footer, homeDir },
  ref,
) {
  const navigate = useNavigate()

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
        const workspaceLabel = sessionWorkspaceLabel(s.workspace, homeDir)
        const href = `/sessions/${encodeURIComponent(s.session_id)}`
        return (
          <a
            key={`${s.runner}/${s.session_id}`}
            className={`session-item session-item--${s.status || 'unknown'}${staleRunning ? ' session-item--stale-running' : ''}`}
            data-testid="session-item"
            href={href}
            onClick={(e) => {
              // Client-side nav without focusing a link in a way that scroll-into-views
              // the row and clobbers the list scroll offset before leave.
              if (
                e.defaultPrevented ||
                e.button !== 0 ||
                e.metaKey ||
                e.ctrlKey ||
                e.shiftKey ||
                e.altKey
              ) {
                return
              }
              e.preventDefault()
              onSessionNavigate?.()
              navigate(href)
            }}
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
            {/* No mid-ellipsis session-id footer under prompt titles — it looked
                capped/junk (e.g. brainstorm…3-083040). Full id stays on the
                card title when the label is the id, and in the href. */}
          </a>
        )
      })}
      {footer}
    </nav>
  )
})
