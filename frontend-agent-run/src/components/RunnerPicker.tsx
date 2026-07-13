import './RunnerPicker.css'

export type RunnerPickerProps = {
  runners: string[]
  value: string
  onChange: (runner: string) => void
}

export function RunnerPicker({ runners, value, onChange }: RunnerPickerProps) {
  return (
    <label className="runner-picker" data-testid="runner-picker">
      <span>Runner</span>
      <select
        data-testid="runner-select"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      >
        {runners.map((r) => (
          <option key={r} value={r}>
            {r}
          </option>
        ))}
      </select>
    </label>
  )
}
