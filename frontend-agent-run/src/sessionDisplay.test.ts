import { describe, expect, it } from 'vitest'
import type { SessionSummary } from './api/client'
import {
  countDoneSessions,
  countSessionsByStatus,
  filterSessionsByStatus,
  formatSessionRecency,
  formatStatusLabel,
  sessionRowLabel,
  sessionWorkspaceLabel,
  shortSessionId,
  shortWorkspaceLabel,
  sortSessionsOldestFirst,
  truncateSessionPreview,
} from './sessionDisplay'

function session(partial: Partial<SessionSummary> & Pick<SessionSummary, 'session_id'>): SessionSummary {
  return {
    runner: 'opencode',
    status: 'running',
    ...partial,
  }
}

describe('truncateSessionPreview', () => {
  it('collapses whitespace and trims', () => {
    expect(truncateSessionPreview('  hello   world  ')).toBe('hello world')
  })

  it('ellipsizes beyond max length', () => {
    const long = 'a'.repeat(120)
    expect(truncateSessionPreview(long, 40)).toBe(`${'a'.repeat(39)}…`)
  })
})

describe('shortSessionId', () => {
  it('keeps short ids intact', () => {
    expect(shortSessionId('web_abc123')).toBe('web_abc123')
  })

  it('truncates long opaque ids', () => {
    expect(shortSessionId('web_101c73bc7aeacc15ab')).toBe('web_101c73…eacc15ab')
  })
})

describe('shortWorkspaceLabel', () => {
  it('shows last two path segments for deep paths', () => {
    expect(shortWorkspaceLabel('/Users/me/projects/agent-pro/src')).toBe('…/agent-pro/src')
  })
})

describe('sessionWorkspaceLabel', () => {
  it('returns em dash when workspace is missing', () => {
    expect(sessionWorkspaceLabel(undefined)).toBe('—')
    expect(sessionWorkspaceLabel('')).toBe('—')
  })

  it('shortens present workspace paths', () => {
    expect(sessionWorkspaceLabel('/Users/me/projects/agent-pro/src')).toBe('…/agent-pro/src')
  })
})

describe('sessionRowLabel', () => {
  it('prefers initial_prompt over session id', () => {
    const label = sessionRowLabel(
      session({
        session_id: 'web_101c73bc7aeacc15',
        initial_prompt: 'How do I refactor the session list UX?',
      }),
    )
    expect(label).toBe('How do I refactor the session list UX?')
  })

  it('falls back to short session id when prompt missing', () => {
    expect(sessionRowLabel(session({ session_id: 'web_101c73bc7aeacc15ab' }))).toBe(
      'web_101c73…eacc15ab',
    )
  })
})

describe('formatSessionRecency', () => {
  const now = Date.parse('2026-07-08T12:00:00Z')

  it('formats minutes ago', () => {
    expect(formatSessionRecency('2026-07-08T11:58:00Z', undefined, now)).toBe('2m ago')
  })

  it('formats hours ago', () => {
    expect(formatSessionRecency('2026-07-08T09:00:00Z', undefined, now)).toBe('3h ago')
  })

  it('formats yesterday', () => {
    expect(formatSessionRecency('2026-07-07T12:00:00Z', undefined, now)).toBe('yesterday')
  })

  it('falls back to created_at', () => {
    expect(formatSessionRecency(undefined, '2026-07-08T11:59:30Z', now)).toBe('just now')
  })
})

describe('formatStatusLabel', () => {
  it('capitalizes running status', () => {
    expect(formatStatusLabel('running')).toBe('Running')
  })

  it('maps finished and idle to Done', () => {
    expect(formatStatusLabel('finished')).toBe('Done')
    expect(formatStatusLabel('idle')).toBe('Done')
  })
})

describe('countSessionsByStatus', () => {
  it('counts matching sessions', () => {
    const sessions = [
      session({ session_id: 'a', status: 'running' }),
      session({ session_id: 'b', status: 'finished' }),
      session({ session_id: 'c', status: 'running' }),
    ]
    expect(countSessionsByStatus(sessions, 'running')).toBe(2)
  })
})

describe('countDoneSessions', () => {
  it('counts only finished and idle, excluding running and error', () => {
    const sessions = [
      session({ session_id: 'a', status: 'running' }),
      session({ session_id: 'b', status: 'finished' }),
      session({ session_id: 'c', status: 'idle' }),
      session({ session_id: 'd', status: 'error' }),
    ]
    expect(countDoneSessions(sessions)).toBe(2)
    expect(countDoneSessions(sessions)).toBe(filterSessionsByStatus(sessions, 'done').length)
  })
})

describe('filterSessionsByStatus', () => {
  it('returns all sessions for all filter', () => {
    const sessions = [
      session({ session_id: 'a', status: 'running' }),
      session({ session_id: 'b', status: 'finished' }),
    ]
    expect(filterSessionsByStatus(sessions, 'all')).toHaveLength(2)
  })

  it('filters running sessions', () => {
    const sessions = [
      session({ session_id: 'a', status: 'running' }),
      session({ session_id: 'b', status: 'finished' }),
    ]
    expect(filterSessionsByStatus(sessions, 'running').map((s) => s.session_id)).toEqual(['a'])
  })

  it('filters done sessions', () => {
    const sessions = [
      session({ session_id: 'a', status: 'running' }),
      session({ session_id: 'b', status: 'finished' }),
      session({ session_id: 'c', status: 'idle' }),
    ]
    expect(filterSessionsByStatus(sessions, 'done').map((s) => s.session_id)).toEqual(['b', 'c'])
  })
})

describe('sortSessionsOldestFirst', () => {
  it('orders by updated_at ascending', () => {
    const sorted = sortSessionsOldestFirst([
      session({ session_id: 'b', updated_at: '2026-07-08T10:00:00Z' }),
      session({ session_id: 'a', updated_at: '2026-07-08T09:00:00Z' }),
      session({ session_id: 'c', updated_at: '2026-07-08T11:00:00Z' }),
    ])
    expect(sorted.map((s) => s.session_id)).toEqual(['a', 'b', 'c'])
  })

  it('uses session_id as tiebreaker', () => {
    const sorted = sortSessionsOldestFirst([
      session({ session_id: 'z', updated_at: '2026-07-08T10:00:00Z' }),
      session({ session_id: 'a', updated_at: '2026-07-08T10:00:00Z' }),
    ])
    expect(sorted.map((s) => s.session_id)).toEqual(['a', 'z'])
  })
})