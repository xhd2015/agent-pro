import { FormEvent } from 'react'
import './Composer.css'

export type ComposerProps = {
  value: string
  onChange: (value: string) => void
  onSend: () => void
  sending: boolean
  hidden?: boolean
  placeholder?: string
}

export function Composer({
  value,
  onChange,
  onSend,
  sending,
  hidden = false,
  placeholder = 'Message the agent…',
}: ComposerProps) {
  return (
    <div className={`composer${hidden ? ' modal-background-hidden' : ''}`} data-testid="composer">
      <form
        className="composer-form"
        onSubmit={(e: FormEvent) => {
          e.preventDefault()
          onSend()
        }}
      >
        <input
          data-testid="composer-input"
          placeholder={placeholder}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={sending}
          aria-label="composer"
        />
        <button type="submit" data-testid="send-button" disabled={sending || !value.trim()}>
          {sending ? '…' : 'Send'}
        </button>
      </form>
    </div>
  )
}
