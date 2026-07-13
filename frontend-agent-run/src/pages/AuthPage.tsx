import { FormEvent, useState } from 'react'
import { setToken } from '../api/client'
import { Shell } from '../components/Shell'
import './AuthPage.css'

export function AuthPage({ onAuthenticated }: { onAuthenticated: () => void }) {
  const [value, setValue] = useState('')
  return (
    <Shell>
      <div className="main-panel auth-page" data-testid="auth-page">
        <h1>agent-run</h1>
        <p>API token required. Copy from <code>agent-run web --token auto</code> startup.</p>
        <form
          onSubmit={(e: FormEvent) => {
            e.preventDefault()
            if (value.trim()) {
              setToken(value.trim())
              // Soft auth: flip gate without full document reload so SPA markers survive.
              onAuthenticated()
            }
          }}
        >
          <input
            data-testid="auth-token-input"
            placeholder="Bearer token"
            value={value}
            onChange={(e) => setValue(e.target.value)}
          />
        </form>
      </div>
    </Shell>
  )
}
