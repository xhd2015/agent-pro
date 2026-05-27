import type { AgentTraceDetail } from "./api";
import "./AgentTracesPrompt.css";

export function AgentTracesPromptPanel({ detail, onClose }: { detail: AgentTraceDetail; onClose: () => void }) {
  return (
    <div className="agent-trace-prompt-overlay" onClick={onClose}>
      <div className="agent-trace-prompt-modal" onClick={(e) => e.stopPropagation()}>
        <div className="agent-trace-prompt-header">
          <div>
            <div className="agent-trace-prompt-title">Prompt</div>
            <div className="agent-trace-prompt-path">{detail.metadata.prompt_path}</div>
          </div>
          <button type="button" className="agent-trace-icon-button" onClick={onClose}>x</button>
        </div>
        <pre className="agent-trace-prompt-body">{detail.prompt}</pre>
      </div>
    </div>
  );
}
