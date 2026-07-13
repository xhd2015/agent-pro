import { Link } from 'react-router-dom'
import { Shell } from '../components/Shell'
import './NotFoundPage.css'

export function NotFoundPage() {
  return (
    <Shell>
      <div className="main-panel not-found-page" data-testid="not-found">
        <h1>Page not found</h1>
        <p>This route does not exist in agent-run.</p>
        <Link className="not-found-home" data-testid="not-found-home" to="/">
          ← Back to home
        </Link>
      </div>
    </Shell>
  )
}
