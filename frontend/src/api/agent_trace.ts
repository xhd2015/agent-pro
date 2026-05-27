import { apiEventSource, apiFetch } from "./client";

const API_PREFIX = "/api/knowledge";

export interface AgentTraceMessage {
  role: "user" | "assistant" | "system" | "tool_call";
  content: string;
  sources?: string[];
  tool_call?: AgentTraceActivity;
  started_at?: number;
  finished_at?: number;
}

export interface AgentTraceActivity {
  subtype: string;
  call_id?: string;
  tool_name: string;
  summary: string;
  kind?: string;
  status?: string;
  file_changes?: AgentTraceFileChange[];
  replace_summary?: boolean;
}

export interface AgentTraceFileChange {
  path: string;
  kind?: string;
}

export interface AgentTraceMetadata {
  id: string;
  command: string;
  command_args?: string[];
  command_line?: string;
  topic_path?: string;
  workspace?: string;
  output_path?: string;
  resume_command?: string;
  provider_id?: string;
  model?: string;
  status: string;
  tags?: string[];
  error?: string;
  created_at: string;
  updated_at: string;
  prompt_path: string;
  log_path: string;
}

export interface AgentTraceSummary extends AgentTraceMetadata {
  log_line_count: number;
}

export interface AgentTraceDetail {
  metadata: AgentTraceMetadata;
  prompt: string;
  messages: AgentTraceMessage[];
  raw_lines: unknown[];
}

export interface AgentTraceUpdate {
  metadata: AgentTraceMetadata;
  messages: AgentTraceMessage[];
  raw_line_count: number;
}

export interface AgentTraceStreamCallbacks {
  onDetail: (detail: AgentTraceDetail) => void;
  onUpdate?: (update: AgentTraceUpdate) => void;
  onDone?: () => void;
  onError?: (error: Error) => void;
}

function normalizeAgentTraceMetadata<T extends AgentTraceMetadata>(metadata: T): T {
  return {
    ...metadata,
    tags: Array.isArray(metadata.tags) ? metadata.tags : [],
  };
}

function normalizeAgentTraceSummary(summary: AgentTraceSummary): AgentTraceSummary {
  return normalizeAgentTraceMetadata(summary);
}

function normalizeAgentTraceDetail(detail: AgentTraceDetail): AgentTraceDetail {
  return {
    ...detail,
    metadata: normalizeAgentTraceMetadata(detail.metadata),
    messages: Array.isArray(detail.messages) ? detail.messages : [],
    raw_lines: Array.isArray(detail.raw_lines) ? detail.raw_lines : [],
  };
}

function normalizeAgentTraceUpdate(update: AgentTraceUpdate): AgentTraceUpdate {
  return {
    ...update,
    metadata: normalizeAgentTraceMetadata(update.metadata),
    messages: Array.isArray(update.messages) ? update.messages : [],
    raw_line_count: update.raw_line_count ?? 0,
  };
}

export async function listAgentTraces(): Promise<AgentTraceSummary[]> {
  const res = await apiFetch(`${API_PREFIX}/agent-traces`);
  if (!res.ok) return [];
  const data = await res.json();
  return Array.isArray(data.sessions) ? data.sessions.map(normalizeAgentTraceSummary) : [];
}

export async function getAgentTrace(id: string): Promise<AgentTraceDetail | null> {
  const res = await apiFetch(`${API_PREFIX}/agent-traces/${encodeURIComponent(id)}`);
  if (!res.ok) return null;
  return normalizeAgentTraceDetail(await res.json());
}

export async function stopAgentTrace(id: string): Promise<AgentTraceDetail | null> {
  const res = await apiFetch(`${API_PREFIX}/agent-traces/${encodeURIComponent(id)}/stop`, { method: "POST" });
  if (!res.ok) return null;
  return normalizeAgentTraceDetail(await res.json());
}

export async function deleteAgentTrace(id: string): Promise<boolean> {
  const res = await apiFetch(`${API_PREFIX}/agent-traces/${encodeURIComponent(id)}`, { method: "DELETE" });
  return res.ok;
}

export function subscribeAgentTrace(id: string, callbacks: AgentTraceStreamCallbacks): () => void {
  const source = apiEventSource(`${API_PREFIX}/agent-traces/${encodeURIComponent(id)}/stream`);
  let closed = false;
  const close = () => {
    closed = true;
    source.close();
  };

  source.addEventListener("detail", (event) => {
    try {
      callbacks.onDetail(normalizeAgentTraceDetail(JSON.parse((event as MessageEvent).data) as AgentTraceDetail));
    } catch (e) {
      callbacks.onError?.(e instanceof Error ? e : new Error(String(e)));
    }
  });
  source.addEventListener("update", (event) => {
    try {
      callbacks.onUpdate?.(normalizeAgentTraceUpdate(JSON.parse((event as MessageEvent).data) as AgentTraceUpdate));
    } catch (e) {
      callbacks.onError?.(e instanceof Error ? e : new Error(String(e)));
    }
  });
  source.addEventListener("done", () => {
    callbacks.onDone?.();
    close();
  });
  source.addEventListener("trace_error", (event) => {
    if (closed) return;
    try {
      const data = JSON.parse((event as MessageEvent).data) as { error?: string };
      callbacks.onError?.(new Error(data.error ?? "Trace stream error"));
    } catch {
      callbacks.onError?.(new Error("Trace stream disconnected"));
    }
    close();
  });
  source.onerror = () => {
    if (closed) return;
    callbacks.onError?.(new Error("Trace stream disconnected"));
    close();
  };
  return close;
}
