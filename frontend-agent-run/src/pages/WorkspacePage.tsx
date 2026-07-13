import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  fetchFsList,
  fetchStatus,
  putWorkspace,
  type AgentRunStatus,
  type FsListEntry,
} from '../api/client'
import { Shell } from '../components/Shell'
import { shortWorkspaceLabel } from '../sessionDisplay'
import './WorkspacePage.css'

export function WorkspacePage() {
  const navigate = useNavigate()
  const [status, setStatus] = useState<AgentRunStatus | null>(null)
  const [browsePath, setBrowsePath] = useState('')
  const [parentPath, setParentPath] = useState<string | undefined>(undefined)
  const [entries, setEntries] = useState<FsListEntry[]>([])
  const [listing, setListing] = useState(false)
  const [using, setUsing] = useState(false)
  const [error, setError] = useState('')
  // Files hidden by default; re-hide on every path change (enter/parent/quick/recent).
  const [showFiles, setShowFiles] = useState(false)

  const loadList = useCallback(async (path: string) => {
    if (!path) return
    setListing(true)
    setError('')
    setShowFiles(false)
    try {
      const list = await fetchFsList(path)
      if (!list) {
        setError('Could not list directory')
        return
      }
      setBrowsePath(list.path)
      setParentPath(list.parent || undefined)
      setEntries(list.entries)
    } finally {
      setListing(false)
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    void (async () => {
      const s = await fetchStatus()
      if (cancelled) return
      if (s) {
        setStatus(s)
        const start = s.workspace || s.process_cwd || s.home || ''
        if (start) {
          setBrowsePath(start)
          void loadList(start)
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [loadList])

  const handleQuickHome = () => {
    const home = status?.home
    if (home) void loadList(home)
  }

  const handleQuickCwd = () => {
    const cwd = status?.process_cwd
    if (cwd) void loadList(cwd)
  }

  const handleRecent = (path: string) => {
    if (path) void loadList(path)
  }

  const handleParent = () => {
    if (parentPath) void loadList(parentPath)
  }

  const handleEnterDir = (entry: FsListEntry) => {
    if (entry.type !== 'dir' && entry.type !== 'directory') return
    void loadList(entry.path)
  }

  const handleUseFolder = async () => {
    if (!browsePath || using) return
    setUsing(true)
    setError('')
    try {
      const result = await putWorkspace(browsePath)
      if (!result) {
        setError('Could not set workspace')
        return
      }
      navigate('/')
    } finally {
      setUsing(false)
    }
  }

  const handleCancel = () => {
    navigate('/')
  }

  const recent = status?.recent_workspaces ?? []

  const isDirEntry = (entry: FsListEntry) =>
    entry.type === 'dir' || entry.type === 'directory'
  const dirEntries = entries.filter((e) => isDirEntry(e))
  const fileEntries = entries.filter((e) => !isDirEntry(e))
  const fileCount = fileEntries.length

  return (
    <Shell>
      <div className="workspace-selector" data-testid="workspace-selector">
        <header className="workspace-selector-header">
          <button
            type="button"
            className="workspace-cancel"
            data-testid="workspace-cancel"
            onClick={handleCancel}
          >
            ← Cancel
          </button>
          <h1 className="workspace-selector-title">Workspace</h1>
        </header>

        <section className="workspace-quick" aria-label="Quick paths">
          <button
            type="button"
            className="workspace-chip"
            data-testid="workspace-quick-home"
            onClick={handleQuickHome}
            disabled={!status?.home}
          >
            Home
          </button>
          <button
            type="button"
            className="workspace-chip"
            data-testid="workspace-quick-cwd"
            onClick={handleQuickCwd}
            disabled={!status?.process_cwd}
          >
            Server cwd
          </button>
        </section>

        {recent.length > 0 ? (
          <section className="workspace-recent" aria-label="Recent workspaces">
            <h2 className="workspace-section-label">Recent</h2>
            <ul className="workspace-recent-list">
              {recent.map((path) => (
                <li key={path}>
                  <button
                    type="button"
                    className="workspace-recent-item"
                    data-testid="workspace-recent-item"
                    data-path={path}
                    onClick={() => handleRecent(path)}
                    title={path}
                  >
                    {shortWorkspaceLabel(path)}
                  </button>
                </li>
              ))}
            </ul>
          </section>
        ) : null}

        <section className="workspace-browser" aria-label="Browse folders">
          <div
            className="workspace-browser-path"
            data-testid="workspace-browser-path"
            title={browsePath}
          >
            {browsePath || '…'}
          </div>
          <button
            type="button"
            className="workspace-browser-parent"
            data-testid="workspace-browser-parent"
            onClick={handleParent}
            disabled={!parentPath}
          >
            ↑ Parent
          </button>
          {listing && entries.length === 0 ? (
            <div className="workspace-browser-loading">Loading…</div>
          ) : null}
          <ul className="workspace-browser-list">
            {dirEntries.map((entry) => (
              <li key={entry.path}>
                <button
                  type="button"
                  className="workspace-browser-entry workspace-browser-entry-dir"
                  data-testid="workspace-browser-entry"
                  data-path={entry.path}
                  data-entry-type="dir"
                  onClick={() => handleEnterDir(entry)}
                >
                  <span className="workspace-entry-icon" aria-hidden="true">
                    📁
                  </span>
                  <span className="workspace-entry-name">{entry.name}</span>
                </button>
              </li>
            ))}
            {fileCount > 0 ? (
              <li className="workspace-show-files-row">
                <button
                  type="button"
                  className="workspace-show-files"
                  data-testid="workspace-show-files"
                  aria-expanded={showFiles}
                  onClick={() => setShowFiles((v) => !v)}
                >
                  {showFiles ? 'Hide files' : `Show files (${fileCount})`}
                </button>
              </li>
            ) : null}
            {showFiles
              ? fileEntries.map((entry) => (
                  <li key={entry.path}>
                    <button
                      type="button"
                      className="workspace-browser-entry workspace-browser-entry-file disabled"
                      data-testid="workspace-browser-entry"
                      data-path={entry.path}
                      data-entry-type="file"
                      aria-disabled="true"
                      disabled
                      onClick={() => handleEnterDir(entry)}
                    >
                      <span className="workspace-entry-icon" aria-hidden="true">
                        📄
                      </span>
                      <span className="workspace-entry-name">{entry.name}</span>
                    </button>
                  </li>
                ))
              : null}
          </ul>
        </section>

        {error ? <p className="workspace-error">{error}</p> : null}

        <footer className="workspace-selector-footer">
          <button
            type="button"
            className="workspace-use-folder"
            data-testid="workspace-use-folder"
            onClick={() => void handleUseFolder()}
            disabled={!browsePath || using}
          >
            {using ? 'Saving…' : 'Use this folder'}
          </button>
        </footer>
      </div>
    </Shell>
  )
}
