import { useEffect, useState } from 'react';
import type { AgentTraceActivity, AgentTraceFileChange } from '../api/agent_trace';
import './ToolCallItem.css';

const SUMMARY_COLLAPSE_THRESHOLD = 3;
const FILE_CHANGE_COLLAPSE_THRESHOLD = 4;
const CODEX_HOOKS_DEPRECATED_PREFIX = '`[features].codex_hooks` is deprecated.';

const FILE_CHANGE_KIND_META: Record<string, { label: string; symbol: string; className: string }> = {
  add: { label: 'Added', symbol: '+', className: 'add' },
  modify: { label: 'Modified', symbol: '~', className: 'modify' },
  delete: { label: 'Deleted', symbol: '-', className: 'delete' },
  rename: { label: 'Renamed', symbol: '>', className: 'rename' },
};

const KIND_LABELS: Record<string, string> = {
  command_execution: 'Command',
  reasoning: 'Reasoning',
  plan_update: 'Plan',
  todo_list: 'Plan',
  todo_write: 'Plan',
  web_search: 'Web Search',
  file_search: 'Search',
  mcp_call: 'MCP',
  mcp_tool_call: 'MCP',
  auth_error: 'Auth',
  runtime_error: 'Error',
  warning: 'Warning',
};

function useElapsedSeconds(startedAt?: number): number | null {
  const [elapsed, setElapsed] = useState<number | null>(startedAt ? Math.floor((Date.now() - startedAt) / 1000) : null);
  useEffect(() => {
    if (startedAt == null) {
      setElapsed(null);
      return;
    }
    setElapsed(Math.floor((Date.now() - startedAt) / 1000));
    const id = window.setInterval(() => {
      setElapsed(Math.floor((Date.now() - startedAt) / 1000));
    }, 1000);
    return () => window.clearInterval(id);
  }, [startedAt]);
  return elapsed;
}

function normalizeFileChangeKind(kind?: string): string {
  switch ((kind ?? '').trim().toLowerCase()) {
    case 'add':
    case 'added':
    case 'create':
    case 'created':
    case 'new':
      return 'add';
    case 'delete':
    case 'deleted':
    case 'remove':
    case 'removed':
      return 'delete';
    case 'rename':
    case 'renamed':
    case 'move':
    case 'moved':
      return 'rename';
    case 'modify':
    case 'modified':
    case 'update':
    case 'updated':
    case 'edit':
    case 'edited':
    case 'write':
    case 'wrote':
      return 'modify';
    default:
      return (kind ?? '').trim().toLowerCase();
  }
}

function parseFileChangeSummary(summary: string): AgentTraceFileChange[] {
  return summary
    .split('\n')
    .slice(1)
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const match = /^([+~>*-])\s+(.+)$/.exec(line);
      if (!match) return { path: line };
      const [, symbol, path] = match;
      const kind = symbol === '+' ? 'add' : symbol === '-' ? 'delete' : symbol === '>' ? 'rename' : 'modify';
      return { path: path.trim(), kind };
    })
    .filter((change) => Boolean(change.path));
}

function buildFileChangeList(toolCall: AgentTraceActivity): AgentTraceFileChange[] {
  if (toolCall.file_changes && toolCall.file_changes.length > 0) {
    return toolCall.file_changes;
  }
  if (toolCall.kind === 'file_change' && toolCall.summary) {
    return parseFileChangeSummary(toolCall.summary);
  }
  return [];
}

function fileName(filePath: string): string {
  const trimmed = filePath.replace(/\/+$/, '');
  const lastSlash = trimmed.lastIndexOf('/');
  return lastSlash >= 0 ? trimmed.slice(lastSlash + 1) : trimmed;
}

function fileParentDir(filePath: string): string {
  const trimmed = filePath.replace(/\/+$/, '');
  const lastSlash = trimmed.lastIndexOf('/');
  return lastSlash >= 0 ? trimmed.slice(0, lastSlash + 1) : '';
}

function isCodexHooksDeprecation(summary: string): boolean {
  return summary.trim().startsWith(CODEX_HOOKS_DEPRECATED_PREFIX);
}

function FileChangeCard({ changes, summary }: { changes: AgentTraceFileChange[]; summary: string }) {
  const [expanded, setExpanded] = useState(false);
  const visibleChanges = expanded ? changes : changes.slice(0, FILE_CHANGE_COLLAPSE_THRESHOLD);
  const hasExtraChanges = changes.length > FILE_CHANGE_COLLAPSE_THRESHOLD;
  const extraSummaryLines = summary
    .split('\n')
    .slice(changes.length > 0 ? 1 : 0)
    .map((line) => line.trim())
    .filter(Boolean)
    .filter((line) => !/^([+~>*-])\s+/.test(line));

  return (
    <div className="kp-tool-call-file-change-card">
      <div className="kp-tool-call-file-change-meta">
        <span className="kp-tool-call-file-change-count">{changes.length} file{changes.length === 1 ? '' : 's'} changed</span>
      </div>
      <div className="kp-tool-call-file-change-list">
        {visibleChanges.map((change, index) => {
          const normalizedKind = normalizeFileChangeKind(change.kind);
          const meta = FILE_CHANGE_KIND_META[normalizedKind] ?? { label: normalizedKind || 'Changed', symbol: '*', className: 'other' };
          const name = fileName(change.path);
          const parentDir = fileParentDir(change.path);
          return (
            <div key={`${change.path}-${index}`} className="kp-tool-call-file-change-entry">
              <span className={`kp-tool-call-file-change-kind kp-tool-call-file-change-kind-${meta.className}`}>
                <span className="kp-tool-call-file-change-symbol">{meta.symbol}</span>
                {meta.label}
              </span>
              <span className="kp-tool-call-file-change-text">
                <span className="kp-tool-call-file-change-name">{name || change.path}</span>
                {parentDir && <span className="kp-tool-call-file-change-path">{parentDir}</span>}
              </span>
            </div>
          );
        })}
      </div>
      {hasExtraChanges && (
        <button type="button" className="kp-tool-call-toggle kp-tool-call-file-change-toggle" onClick={() => setExpanded((value) => !value)}>
          {expanded ? 'Show less' : `Show ${changes.length - FILE_CHANGE_COLLAPSE_THRESHOLD} more`}
        </button>
      )}
      {extraSummaryLines.length > 0 && (
        <div className="kp-tool-call-file-change-notes">{extraSummaryLines.join('\n')}</div>
      )}
    </div>
  );
}

export function ToolCallItem({ toolCall, startedAt }: { toolCall: AgentTraceActivity; startedAt?: number }) {
  const [expanded, setExpanded] = useState(false);
  const elapsed = useElapsedSeconds(startedAt);
  const running = startedAt != null;
  const summary = toolCall.summary;
  const deprecationWarning = isCodexHooksDeprecation(summary);
  const isWarning = toolCall.status === 'warning' || deprecationWarning;
  const name = isWarning && toolCall.tool_name.toLowerCase() === 'error' ? 'Config Warning' : toolCall.tool_name;
  const kind = isWarning && toolCall.kind === 'error' ? 'warning' : toolCall.kind;
  const status = isWarning ? 'warning' : toolCall.status;
  const kindLabel = kind ? (KIND_LABELS[kind] ?? kind.replace(/[_-]+/g, ' ')) : '';
  const statusLabel = status === 'failed' ? 'failed' : status === 'warning' ? 'warning' : running ? 'running' : (status === 'completed' ? 'done' : '');
  const statusClass = status ?? 'completed';
  const warningClass = status === 'warning' ? ' kp-tool-call-warning-item' : '';
  const isFileChange = toolCall.kind === 'file_change';
  const fileChanges = buildFileChangeList(toolCall);
  const runningBadge = running && (
    <span className="kp-tool-call-running">
      <span className="kp-tool-call-running-dot" />
      running {elapsed != null ? `${elapsed}s` : ''}
    </span>
  );

  const lines = summary.split('\n');
  const alwaysExpand = name === 'Update Todos' || toolCall.kind === 'todo_list' || toolCall.kind === 'todo_write' || toolCall.kind === 'plan_update';
  const collapsible = !alwaysExpand && lines.length > SUMMARY_COLLAPSE_THRESHOLD;
  const displayText = !collapsible || expanded ? summary : `${lines.slice(0, SUMMARY_COLLAPSE_THRESHOLD).join('\n')}\n...`;

  return (
    <div className={`kp-tool-call-item${isFileChange ? ' kp-tool-call-file-change-item' : ''}${running ? ' kp-tool-call-running-item' : ''}${warningClass}`}>
      <div className="kp-tool-call-header">
        <div className="kp-tool-call-title">
          <span className="kp-tool-call-name">{name}</span>
          {kindLabel && kindLabel.toLowerCase() !== name.toLowerCase() && <span className="kp-tool-call-kind">{kindLabel}</span>}
          {statusLabel && !running && <span className={`kp-tool-call-status kp-tool-call-status-${statusClass}`}>{statusLabel}</span>}
        </div>
        {runningBadge}
      </div>
      {isFileChange && fileChanges.length > 0 ? (
        <FileChangeCard changes={fileChanges} summary={summary} />
      ) : summary ? (
        <span className="kp-tool-call-summary">
          {displayText}
          {collapsible && (
            <button type="button" className="kp-tool-call-toggle" onClick={() => setExpanded((v) => !v)}>
              {expanded ? 'Collapse' : 'Expand'}
            </button>
          )}
        </span>
      ) : (
        <span className="kp-tool-call-file-change-empty">Waiting for details...</span>
      )}
    </div>
  );
}
