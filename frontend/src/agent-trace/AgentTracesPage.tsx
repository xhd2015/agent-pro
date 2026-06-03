import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  deleteAgentTrace,
  getAgentTrace,
  listAgentTraces,
  stopAgentTrace,
  subscribeAgentTrace,
  type AgentTraceDetail,
  type AgentTraceMetadata,
  type AgentTraceSummary,
  type AgentTraceUpdate,
} from "./api";
import { ChatMessage } from "../knowledge/ChatMessage";
import { AgentTracesPromptPanel } from "./AgentTracesPromptPanel";
import { commandTitle, formatDate, shortPath, statusClass, statusLabel, tagClass, tagLabel } from "./traceFormat";
import "./AgentTracesPage.css";

const AUTO_FOLLOW_BOTTOM_PX = 32;
const DEFAULT_ROUTE_BASE = "/knowledge/agent-traces";
const COMPLETED_TRACE_REFRESH_MS = 2500;

type RefreshOptions = {
  quiet?: boolean;
};

function isScrolledToBottom(el: HTMLElement): boolean {
  return el.scrollHeight - el.scrollTop - el.clientHeight <= AUTO_FOLLOW_BOTTOM_PX;
}

function traceRoute(routeBase: string, id: string): string {
  const encoded = encodeURIComponent(id);
  const base = routeBase.replace(/\/+$/, "");
  return base ? `${base}/${encoded}` : `/${encoded}`;
}

function nextTraceIdAfterDelete(sessions: AgentTraceSummary[], deletedId: string): string {
  const deletedIndex = sessions.findIndex((session) => session.id === deletedId);
  const nextSessions = sessions.filter((session) => session.id !== deletedId);
  if (nextSessions.length === 0) return "";
  if (deletedIndex < 0) return nextSessions[0].id;
  return nextSessions[deletedIndex]?.id ?? nextSessions[deletedIndex - 1]?.id ?? "";
}

function navigateToTraceOrEmpty(navigate: ReturnType<typeof useNavigate>, routeBase: string, id: string) {
  if (id) {
    navigate(traceRoute(routeBase, id), { replace: true });
  } else {
    navigate(routeBase || "/", { replace: true });
  }
}

function TraceTags({ tags }: { tags?: string[] }) {
  const visibleTags = (tags ?? []).filter(Boolean);
  if (visibleTags.length === 0) return null;
  return (
    <div className="agent-trace-tags">
      {visibleTags.map((tag) => (
        <span key={tag} className={`agent-trace-tag agent-trace-tag-${tagClass(tag)}`}>
          {tagLabel(tag)}
        </span>
      ))}
    </div>
  );
}

export function AgentTracesPage({ routeBase = DEFAULT_ROUTE_BASE }: { routeBase?: string }) {
  const { traceId } = useParams<{ traceId?: string }>();
  const navigate = useNavigate();
  const [sessions, setSessions] = useState<AgentTraceSummary[]>([]);
  const [detail, setDetail] = useState<AgentTraceDetail | null>(null);
  const [loadingSessions, setLoadingSessions] = useState(false);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [stoppingTrace, setStoppingTrace] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deletingTraceId, setDeletingTraceId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [showPrompt, setShowPrompt] = useState(false);
  const logRef = useRef<HTMLDivElement>(null);
  const autoFollowRef = useRef(true);
  const programmaticScrollRef = useRef(false);
  const selectedId = traceId ?? sessions[0]?.id ?? "";
  const selectedSession = sessions.find((session) => session.id === selectedId);
  const visibleDetail = detail?.metadata.id === selectedId ? detail : null;

  async function refreshSessions(options: RefreshOptions = {}) {
    if (!options.quiet) setLoadingSessions(true);
    try {
      const next = await listAgentTraces();
      setSessions(next);
      if (!traceId && next.length > 0) {
        navigate(traceRoute(routeBase, next[0].id), { replace: true });
      } else if (traceId && next.length > 0 && !next.some((session) => session.id === traceId)) {
        navigate(traceRoute(routeBase, next[0].id), { replace: true });
      } else if (traceId && next.length === 0) {
        navigate(routeBase || "/", { replace: true });
      }
    } finally {
      if (!options.quiet) setLoadingSessions(false);
    }
  }

  async function refreshDetail(id: string, options: RefreshOptions = {}) {
    if (!id) return;
    if (!options.quiet) {
      setLoadingDetail(true);
      setError(null);
    }
    try {
      const next = await getAgentTrace(id);
      if (!next) {
        if (!options.quiet) setError("Trace session not found.");
        return;
      }
      applyTraceDetail(next);
    } catch (e) {
      if (!options.quiet) setError(e instanceof Error ? e.message : String(e));
    } finally {
      if (!options.quiet) setLoadingDetail(false);
    }
  }

  useEffect(() => {
    refreshSessions();
  }, []);

  const selectedStatus = visibleDetail?.metadata.status ?? selectedSession?.status;
  const selectedIsRunning = selectedStatus === "running";

  useEffect(() => {
    if (!selectedId) return;
    setLoadingDetail(true);
    setError(null);
    const unsubscribe = subscribeAgentTrace(selectedId, {
      onDetail: (next) => {
        applyTraceDetail(next);
        setLoadingDetail(false);
      },
      onUpdate: (next) => {
        applyTraceUpdate(next);
        setLoadingDetail(false);
      },
      onDone: () => {
        setLoadingDetail(false);
      },
      onError: (e) => {
        setLoadingDetail(false);
        setError(e.message);
        if (!visibleDetail) {
          refreshDetail(selectedId);
        }
      },
    });
    return unsubscribe;
  }, [selectedId, selectedIsRunning]);

  useEffect(() => {
    if (!selectedId || selectedIsRunning) return;
    const timer = window.setInterval(() => {
      refreshDetail(selectedId, { quiet: true });
      refreshSessions({ quiet: true });
    }, COMPLETED_TRACE_REFRESH_MS);
    return () => window.clearInterval(timer);
  }, [selectedId, selectedIsRunning]);

  function applyTraceDetail(next: AgentTraceDetail) {
    setDetail(next);
    updateSessionFromTrace(next.metadata, next.raw_lines.length);
  }

  function applyTraceUpdate(next: AgentTraceUpdate) {
    setDetail((prev) => {
      if (!prev || prev.metadata.id !== next.metadata.id) return prev;
      return {
        ...prev,
        metadata: next.metadata,
        messages: next.messages,
        raw_lines: Array.from({ length: next.raw_line_count }),
      };
    });
    updateSessionFromTrace(next.metadata, next.raw_line_count);
  }

  function updateSessionFromTrace(metadata: AgentTraceMetadata, rawLineCount: number) {
    setSessions((prev) => prev.map((session) => (
      session.id === metadata.id
        ? { ...session, ...metadata, log_line_count: rawLineCount }
        : session
    )));
  }

  async function handleStopTrace() {
    if (!selectedId || stoppingTrace) return;
    setStoppingTrace(true);
    setError(null);
    try {
      const next = await stopAgentTrace(selectedId);
      if (!next) {
        setError("Unable to mark trace stopped.");
        return;
      }
      applyTraceDetail(next);
      refreshSessions({ quiet: true });
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setStoppingTrace(false);
    }
  }

  async function handleDeleteTrace(id: string) {
    if (!id || deletingTraceId) return;
    setDeletingTraceId(id);
    setError(null);
    try {
      const ok = await deleteAgentTrace(id);
      if (!ok) {
        setError("Unable to delete trace session.");
        return;
      }
      const nextSessions = sessions.filter((session) => session.id !== id);
      const nextSelectedId = id === selectedId ? nextTraceIdAfterDelete(sessions, id) : selectedId;
      setSessions(nextSessions);
      setShowDeleteConfirm(false);
      if (id === selectedId) {
        setDetail(null);
        navigateToTraceOrEmpty(navigate, routeBase, nextSelectedId);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setDeletingTraceId(null);
    }
  }

  useEffect(() => {
    autoFollowRef.current = true;
    setShowDeleteConfirm(false);
  }, [selectedId]);

  useEffect(() => {
    if (!autoFollowRef.current) return;
    const log = logRef.current;
    if (!log) return;
    programmaticScrollRef.current = true;
    log.scrollTo({ top: log.scrollHeight, behavior: "auto" });
    window.requestAnimationFrame(() => {
      programmaticScrollRef.current = false;
      autoFollowRef.current = isScrolledToBottom(log);
    });
  }, [
    selectedId,
    visibleDetail?.messages.length,
    visibleDetail?.raw_lines.length,
    visibleDetail?.metadata.status,
    visibleDetail?.metadata.updated_at,
  ]);

  function handleSelect(id: string) {
    navigate(traceRoute(routeBase, id));
  }

  function handleLogScroll() {
    if (programmaticScrollRef.current) return;
    const log = logRef.current;
    if (!log) return;
    autoFollowRef.current = isScrolledToBottom(log);
  }

  const hasSessions = sessions.length > 0;
  const messageCount = visibleDetail?.messages.length ?? 0;

  return (
    <div className="agent-traces-page">
      <aside className="agent-trace-sidebar">
        <div className="agent-trace-sidebar-header">
          <div>
            <h2>Agent Traces</h2>
            <p>{sessions.length} sessions</p>
          </div>
          <button type="button" className="agent-trace-refresh" onClick={() => refreshSessions()}>
            Refresh
          </button>
        </div>
        {loadingSessions && sessions.length === 0 && (
          <div className="agent-trace-empty">Loading sessions...</div>
        )}
        {!loadingSessions && !hasSessions && (
          <div className="agent-trace-empty">
            Run a headless knowledge-portal command to create the first trace.
          </div>
        )}
        <div className="agent-trace-session-list">
          {sessions.map((session) => (
            <div
              key={session.id}
              className={`agent-trace-session${session.id === selectedId ? " active" : ""}`}
            >
              <button
                type="button"
                className="agent-trace-session-select"
                onClick={() => handleSelect(session.id)}
              >
                <div className="agent-trace-session-top">
                  <span className="agent-trace-session-command">{commandTitle(session)}</span>
                  <span className={`agent-trace-status agent-trace-status-${statusClass(session.status)}`}>
                    {statusLabel(session.status)}
                  </span>
                </div>
                <div className="agent-trace-session-time">{formatDate(session.created_at)}</div>
                <div className="agent-trace-session-meta">
                  {(session.provider_id || "agent")} / {(session.model || "default")} / {session.log_line_count} lines
                </div>
                {session.children && session.children.length > 0 && (
                  <div className="agent-trace-session-links">
                    {session.children.length} delegated trace{session.children.length === 1 ? "" : "s"}
                  </div>
                )}
                <TraceTags tags={session.tags} />
                {session.topic_path && <div className="agent-trace-session-topic">{session.topic_path}</div>}
              </button>
            </div>
          ))}
        </div>
      </aside>

      <main className="agent-trace-main">
        {!visibleDetail && loadingDetail && (
          <div className="agent-trace-main-empty">Loading trace...</div>
        )}
        {!visibleDetail && !loadingDetail && !selectedId && (
          <div className="agent-trace-main-empty">Select a trace session.</div>
        )}
        {visibleDetail && (
          <>
            <header className="agent-trace-header">
              <div className="agent-trace-title-block">
                <div className="agent-trace-eyebrow">{visibleDetail.metadata.id}</div>
                <h1>{visibleDetail.metadata.command}</h1>
                <div className="agent-trace-command-line">{visibleDetail.metadata.command_line}</div>
              </div>
              <div className="agent-trace-actions">
                {loadingDetail && <span className="agent-trace-loading">Updating...</span>}
                {selectedIsRunning && (
                  <button
                    type="button"
                    className="agent-trace-action-button agent-trace-stop-button"
                    onClick={handleStopTrace}
                    disabled={stoppingTrace}
                  >
                    {stoppingTrace ? "Marking..." : "Mark stopped"}
                  </button>
                )}
                <button type="button" className="agent-trace-action-button" onClick={() => refreshDetail(selectedId)}>
                  Refresh
                </button>
                <button type="button" className="agent-trace-action-button" onClick={() => setShowPrompt(true)}>
                  Prompt
                </button>
                <div className="agent-trace-delete-action">
                  <button
                    type="button"
                    className="agent-trace-action-button agent-trace-delete-button"
                    onClick={() => setShowDeleteConfirm((open) => !open)}
                    disabled={deletingTraceId === selectedId}
                  >
                    Delete
                  </button>
                  {showDeleteConfirm && (
                    <div className="agent-trace-delete-popconfirm">
                      <span>Delete this session?</span>
                      <button type="button" onClick={() => handleDeleteTrace(selectedId)} disabled={deletingTraceId === selectedId}>
                        {deletingTraceId === selectedId ? "Deleting..." : "Delete"}
                      </button>
                      <button type="button" onClick={() => setShowDeleteConfirm(false)} disabled={deletingTraceId === selectedId}>
                        Cancel
                      </button>
                    </div>
                  )}
                </div>
              </div>
            </header>

            <section className="agent-trace-facts">
              <div>
                <span>Status</span>
                <div className="agent-trace-status-stack">
                  <strong className={`agent-trace-status agent-trace-status-${statusClass(visibleDetail.metadata.status)}`}>
                    {statusLabel(visibleDetail.metadata.status)}
                  </strong>
                  <TraceTags tags={visibleDetail.metadata.tags} />
                </div>
              </div>
              <div>
                <span>Agent</span>
                <strong>{visibleDetail.metadata.provider_id || "unknown"}</strong>
              </div>
              <div>
                <span>Model</span>
                <strong>{visibleDetail.metadata.model || "default"}</strong>
              </div>
              <div>
                <span>Started</span>
                <strong>{formatDate(visibleDetail.metadata.created_at)}</strong>
              </div>
              <div>
                <span>Workspace</span>
                <strong title={visibleDetail.metadata.workspace}>{shortPath(visibleDetail.metadata.workspace)}</strong>
              </div>
              <div>
                <span>Output</span>
                <strong title={visibleDetail.metadata.output_path}>{shortPath(visibleDetail.metadata.output_path)}</strong>
              </div>
            </section>

            {(visibleDetail.metadata.parent_trace_id || (visibleDetail.metadata.children?.length ?? 0) > 0) && (
              <section className="agent-trace-links-panel">
                {visibleDetail.metadata.parent_trace_id && (
                  <div className="agent-trace-link-group">
                    <span>Parent</span>
                    <button
                      type="button"
                      className="agent-trace-link-button"
                      onClick={() => handleSelect(visibleDetail.metadata.parent_trace_id ?? "")}
                    >
                      <strong>{visibleDetail.metadata.parent_session_id || visibleDetail.metadata.parent_trace_id}</strong>
                      {(visibleDetail.metadata.delegation_label || visibleDetail.metadata.delegation_id) && (
                        <em>{visibleDetail.metadata.delegation_label || visibleDetail.metadata.delegation_id}</em>
                      )}
                    </button>
                  </div>
                )}
                {(visibleDetail.metadata.children?.length ?? 0) > 0 && (
                  <div className="agent-trace-link-group">
                    <span>Delegated Traces</span>
                    <div className="agent-trace-child-list">
                      {visibleDetail.metadata.children?.map((child) => (
                        <button
                          key={child.id}
                          type="button"
                          className="agent-trace-link-button"
                          onClick={() => handleSelect(child.id)}
                        >
                          <strong>{child.command || child.id}</strong>
                          <em>{child.delegation_label || child.delegation_id || formatDate(child.created_at)}</em>
                          <small className={`agent-trace-status agent-trace-status-${statusClass(child.status)}`}>
                            {statusLabel(child.status)}
                          </small>
                        </button>
                      ))}
                    </div>
                  </div>
                )}
              </section>
            )}

            {visibleDetail.metadata.error && (
              <div className="agent-trace-error">{visibleDetail.metadata.error}</div>
            )}

            <div className="agent-trace-log" ref={logRef} onScroll={handleLogScroll}>
              {messageCount === 0 && (
                <div className="agent-trace-main-empty">
                  No normalized trace messages yet. Raw log lines: {visibleDetail.raw_lines.length}.
                </div>
              )}
              {visibleDetail.messages.map((msg, index) => (
                <ChatMessage key={`${visibleDetail.metadata.id}-${index}`} msg={msg} />
              ))}
              {!selectedIsRunning && visibleDetail.metadata.resume_command && (
                <div className="agent-trace-resume">
                  <span>Resume command</span>
                  <code>{visibleDetail.metadata.resume_command}</code>
                </div>
              )}
            </div>
          </>
        )}
      </main>
      {visibleDetail && showPrompt && <AgentTracesPromptPanel detail={visibleDetail} onClose={() => setShowPrompt(false)} />}
      {error && <div className="agent-trace-toast">{error}</div>}
    </div>
  );
}
