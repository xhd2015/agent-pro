import { useEffect, useMemo, useState } from 'react'
import { parseRFC3339Ms } from '../sessionDisplay'
import './AgentRunningCard.css'

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

export type AgentRunningCardProps = {
  updatedAt?: string
  createdAt?: string
}

export function AgentRunningCard({ updatedAt, createdAt }: AgentRunningCardProps) {
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
