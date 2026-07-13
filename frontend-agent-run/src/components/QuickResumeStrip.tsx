import { useMemo } from 'react'
import { Link } from 'react-router-dom'
import type { SessionSummary } from '../api/client'
import {
  formatSessionRecency,
  getQuickResumeSessions,
  sessionRowLabel,
} from '../sessionDisplay'
import './QuickResumeStrip.css'

export function QuickResumeStrip({ sessions }: { sessions: SessionSummary[] }) {
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
              to={`/sessions/${encodeURIComponent(s.session_id)}`}
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
