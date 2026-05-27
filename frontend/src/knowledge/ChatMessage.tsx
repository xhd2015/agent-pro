import type { AgentTraceMessage } from '../api/agent_trace';
import { ToolCallItem } from './ToolCallItem';
import './ChatMessage.css';

export type LiveMessage = AgentTraceMessage;

function formatTimestamp(ms: number): string {
  const d = new Date(ms);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

export function ChatMessage({ msg }: { msg: LiveMessage }) {
  if (msg.role === 'tool_call' && msg.tool_call) {
    const isRunning = msg.started_at != null && msg.finished_at == null;
    return <ToolCallItem toolCall={msg.tool_call} startedAt={isRunning ? msg.started_at : undefined} />;
  }
  return (
    <div className={`kp-msg kp-msg-${msg.role}`}>
      <div className="kp-msg-content">{msg.content}</div>
      {msg.sources && msg.sources.length > 0 && (
        <div className="kp-msg-sources">
          <div className="kp-msg-sources-label">Sources:</div>
          <ul>
            {msg.sources.map((src, j) => (
              <li key={j}>{src}</li>
            ))}
          </ul>
        </div>
      )}
      {msg.started_at != null && (
        <div className="kp-msg-time">{formatTimestamp(msg.started_at)}</div>
      )}
    </div>
  );
}
