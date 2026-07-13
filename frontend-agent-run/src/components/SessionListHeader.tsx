import type { SessionListCounts, SessionSummary } from '../api/client'
import {
  countDoneSessions,
  countSessionsByStatus,
  sessionListCountLabel,
  type SessionStatusFilter,
} from '../sessionDisplay'
import './SessionListHeader.css'

export function SessionListHeader({
  sessions,
  visibleCount,
  filter,
  onFilterChange,
  refreshing,
  counts,
  searchQuery,
  onSearchChange,
  total,
}: {
  sessions: SessionSummary[]
  visibleCount: number
  filter: SessionStatusFilter
  onFilterChange: (filter: SessionStatusFilter) => void
  refreshing: boolean
  counts?: SessionListCounts | null
  searchQuery?: string
  onSearchChange?: (value: string) => void
  /** Full match total from server (after q+status); falls back to sessions.length. */
  total?: number
}) {
  const allCount = counts?.all ?? sessions.length
  const runningCount = counts?.running ?? countSessionsByStatus(sessions, 'running')
  const doneCount = counts?.done ?? countDoneSessions(sessions)
  const chips: { id: SessionStatusFilter; label: string; count: number }[] = [
    { id: 'all', label: 'All', count: allCount },
    { id: 'running', label: 'Running', count: runningCount },
    { id: 'done', label: 'Done', count: doneCount },
  ]
  const labelTotal = typeof total === 'number' ? total : sessions.length

  return (
    <div className="session-list-header" data-testid="session-list-header">
      {onSearchChange != null ? (
        <div className="session-search-row">
          <input
            type="search"
            className="session-search"
            data-testid="session-search"
            placeholder="Search sessions…"
            value={searchQuery ?? ''}
            onChange={(e) => onSearchChange(e.target.value)}
            aria-label="Search sessions"
            autoComplete="off"
            spellCheck={false}
          />
        </div>
      ) : null}
      <div className="session-list-header-row">
        <span className="session-list-count" data-testid="session-list-count">
          {sessionListCountLabel(labelTotal, visibleCount, filter)}
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
