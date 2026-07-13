import { RunnerPicker } from './RunnerPicker'
import { WorkspacePath } from './WorkspacePath'
import './HomeTopBar.css'

export function HomeTopBar({
  runners,
  runner,
  onRunnerChange,
  workspace,
  onOpenWorkspaceSelector,
}: {
  runners: string[]
  runner: string
  onRunnerChange: (runner: string) => void
  workspace: string
  /** Primary action: open SPA path selector at /workspace. */
  onOpenWorkspaceSelector?: () => void
}) {
  return (
    <header className="top-bar top-bar-home">
      <div className="top-bar-row top-bar-row-primary">
        <h1 className="app-title">agent-run</h1>
        <RunnerPicker runners={runners} value={runner} onChange={onRunnerChange} />
      </div>
      {workspace ? (
        <div className="top-bar-row top-bar-row-workspace">
          <WorkspacePath
            path={workspace}
            className="workspace-display"
            onOpenSelector={onOpenWorkspaceSelector}
            showCopy={onOpenWorkspaceSelector == null}
          />
        </div>
      ) : null}
    </header>
  )
}
