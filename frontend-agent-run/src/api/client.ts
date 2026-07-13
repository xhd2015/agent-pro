const TOKEN_KEY = 'agent-run-token'
const RUNNER_KEY = 'agent-run-runner'
const DEFAULT_RUNNER = 'opencode'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

export function getRunner(): string {
  return localStorage.getItem(RUNNER_KEY) ?? DEFAULT_RUNNER
}

export function setRunner(runner: string): void {
  localStorage.setItem(RUNNER_KEY, runner)
}

async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const token = getToken()
  const headers = new Headers(init.headers)
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  return fetch(path, { ...init, headers })
}

export async function fetchHealth(): Promise<number> {
  const res = await apiFetch('/api/agent-run/health')
  return res.status
}

export type AgentRunStatus = {
  home: string
  workspace: string
  process_cwd?: string
  agent_run_home?: string
  recent_workspaces?: string[]
}

export async function fetchStatus(): Promise<AgentRunStatus | null> {
  const res = await apiFetch('/api/agent-run/status')
  if (!res.ok) {
    return null
  }
  return (await res.json()) as AgentRunStatus
}

export async function putWorkspace(path: string): Promise<AgentRunStatus | null> {
  const res = await apiFetch('/api/agent-run/workspace', {
    method: 'PUT',
    body: JSON.stringify({ path }),
  })
  if (!res.ok) {
    return null
  }
  return (await res.json()) as AgentRunStatus
}

export type FsListEntry = {
  name: string
  path: string
  type: 'dir' | 'file' | string
}

export type FsListResponse = {
  path: string
  parent?: string
  entries: FsListEntry[]
}

export async function fetchFsList(path: string): Promise<FsListResponse | null> {
  const q = new URLSearchParams({ path })
  const res = await apiFetch(`/api/agent-run/fs/list?${q.toString()}`)
  if (!res.ok) {
    return null
  }
  const data = (await res.json()) as FsListResponse
  return {
    path: data.path,
    parent: data.parent,
    entries: data.entries ?? [],
  }
}

export type SessionSummary = {
  runner: string
  session_id: string
  initial_prompt?: string
  status: string
  terminal_session_id?: string
  workspace?: string
  model?: string
  created_at?: string
  updated_at?: string
}

export type SessionStatusFilterParam = 'all' | 'running' | 'done'

export type SessionListCounts = {
  all: number
  running: number
  done: number
}

export type SessionsPage = {
  sessions: SessionSummary[]
  total: number
  limit: number
  offset: number
  has_more: boolean
  counts: SessionListCounts
}

export type FetchSessionsPageParams = {
  limit?: number
  offset?: number
  q?: string
  status?: SessionStatusFilterParam | string
}

/** Paginated/filtered sessions list (newest-first on server). */
export async function fetchSessionsPage(
  params: FetchSessionsPageParams = {},
): Promise<SessionsPage | null> {
  const sp = new URLSearchParams()
  if (params.limit != null && params.limit > 0) {
    sp.set('limit', String(params.limit))
  }
  if (params.offset != null && params.offset > 0) {
    sp.set('offset', String(params.offset))
  }
  const q = params.q?.trim()
  if (q) {
    sp.set('q', q)
  }
  const status = params.status?.trim()
  if (status && status !== 'all') {
    sp.set('status', status)
  }
  const qs = sp.toString()
  const path = qs ? `/api/agent-run/sessions?${qs}` : '/api/agent-run/sessions'
  const res = await apiFetch(path)
  if (res.status === 401) {
    throw new Error('unauthorized')
  }
  if (!res.ok) {
    return null
  }
  const data = (await res.json()) as {
    sessions?: SessionSummary[]
    total?: number
    limit?: number
    offset?: number
    has_more?: boolean
    counts?: Partial<SessionListCounts>
  }
  const sessions = data.sessions ?? []
  const total = typeof data.total === 'number' ? data.total : sessions.length
  const limit = typeof data.limit === 'number' ? data.limit : 0
  const offset = typeof data.offset === 'number' ? data.offset : 0
  const has_more = typeof data.has_more === 'boolean' ? data.has_more : false
  const counts: SessionListCounts = {
    all: typeof data.counts?.all === 'number' ? data.counts.all : total,
    running: typeof data.counts?.running === 'number' ? data.counts.running : 0,
    done: typeof data.counts?.done === 'number' ? data.counts.done : 0,
  }
  return { sessions, total, limit, offset, has_more, counts }
}

/** Full sessions list (no limit) — compat helper. */
export async function fetchSessions(): Promise<SessionSummary[] | null> {
  const page = await fetchSessionsPage({})
  return page?.sessions ?? null
}

export type RunnersResponse = {
  runners: string[]
  default: string
}

export async function fetchRunners(): Promise<RunnersResponse> {
  const res = await apiFetch('/api/agent-run/runners')
  if (!res.ok) {
    return { runners: [DEFAULT_RUNNER], default: DEFAULT_RUNNER }
  }
  return (await res.json()) as RunnersResponse
}

export type AgentEvent = {
  type: string
  role?: string
  text?: string
  phase?: string
  id?: string
  timestamp?: number
  tool?: string
  tool_input?: Record<string, unknown>
  tool_call_id?: string
  output?: string
  stderr?: string
  exit_code?: number
}

export type SessionDetail = {
  session: SessionSummary
  events: AgentEvent[]
  events_offset?: number
}

export function readSessionBootstrap(sessionId?: string): SessionDetail | null {
  const el = document.getElementById('agent-run-session-bootstrap')
  if (!el?.textContent) {
    return null
  }
  try {
    const data = JSON.parse(el.textContent) as SessionDetail
    if (sessionId && data.session?.session_id !== sessionId) {
      return null
    }
    return data
  } catch {
    return null
  }
}

export type TerminalStatus = {
  available: boolean
  runner: string
  session_id: string
  terminal_session_id?: string
}

export async function fetchSessionDetail(sessionId: string): Promise<SessionDetail | null> {
  const res = await apiFetch(
    `/api/agent-run/sessions/${encodeURIComponent(sessionId)}`,
  )
  if (!res.ok) {
    return null
  }
  return (await res.json()) as SessionDetail
}

export async function createSession(runner: string, prompt: string): Promise<SessionSummary | null> {
  const res = await apiFetch('/api/agent-run/sessions', {
    method: 'POST',
    body: JSON.stringify({ runner, prompt }),
  })
  if (!res.ok) {
    return null
  }
  const data = (await res.json()) as { session?: SessionSummary }
  return data.session ?? null
}

export async function sendSessionMessage(
  sessionId: string,
  text: string,
): Promise<boolean> {
  const res = await apiFetch(
    `/api/agent-run/sessions/${encodeURIComponent(sessionId)}/messages`,
    {
      method: 'POST',
      body: JSON.stringify({ text }),
    },
  )
  return res.ok
}

export async function fetchTerminalStatus(
  sessionId: string,
  runner = '',
): Promise<TerminalStatus> {
  const res = await apiFetch(
    `/api/agent-run/sessions/${encodeURIComponent(sessionId)}/terminal`,
  )
  if (!res.ok) {
    return { available: false, runner, session_id: sessionId }
  }
  return (await res.json()) as TerminalStatus
}

export function openTerminalWebSocket(sessionId: string): WebSocket {
  const token = getToken()
  const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const path = `/api/agent-run/sessions/${encodeURIComponent(sessionId)}/terminal/ws`
  const url = new URL(path, `${scheme}//${window.location.host}`)
  if (token) {
    url.searchParams.set('token', token)
  }
  return new WebSocket(url)
}

/** Tails session events via SSE (supports Bearer auth). Calls onEvent per AgentEvent JSON line. */
export function subscribeSessionEvents(
  sessionId: string,
  afterOffset: number,
  onEvent: (ev: AgentEvent) => void,
  signal?: AbortSignal,
  onClose?: () => void,
): void {
  const url = `/api/agent-run/sessions/${encodeURIComponent(sessionId)}/events/stream?after=${afterOffset}`
  void (async () => {
    try {
      const res = await apiFetch(url, {
        headers: { Accept: 'text/event-stream' },
        signal,
      })
      if (!res.ok || !res.body) {
        return
      }
      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      for (;;) {
        const { done, value } = await reader.read()
        if (done) {
          break
        }
        buffer += decoder.decode(value, { stream: true })
        for (;;) {
          const idx = buffer.indexOf('\n\n')
          if (idx < 0) {
            break
          }
          const block = buffer.slice(0, idx)
          buffer = buffer.slice(idx + 2)
          for (const line of block.split('\n')) {
            const trimmed = line.trim()
            if (!trimmed.startsWith('data:')) {
              continue
            }
            const payload = trimmed.slice(5).trim()
            if (!payload || payload === '[DONE]') {
              continue
            }
            try {
              const ev = JSON.parse(payload) as AgentEvent
              onEvent(ev)
            } catch {
              // ignore malformed chunks
            }
          }
        }
      }
      if (!signal?.aborted) {
        onClose?.()
      }
    } catch {
      // Aborted or failed — skip onClose so we don't refresh after intentional teardown.
    }
  })()
}
