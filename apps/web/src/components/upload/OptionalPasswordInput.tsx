import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'

type OptionalPasswordInputProps = {
  enabled: boolean
  password: string
  onEnabledChange: (next: boolean) => void
  onPasswordChange: (next: string) => void
  disabled?: boolean
}

export const OptionalPasswordInput = ({
  enabled,
  password,
  onEnabledChange,
  onPasswordChange,
  disabled,
}: OptionalPasswordInputProps) => {
  return (
    <Card title="Archive password" subtitle="Enable only if your file is password-protected">
      <div className="space-y-2">
        <label className="flex items-center gap-2 text-sm text-slate-200">
          <input
            type="checkbox"
            checked={enabled}
            onChange={(event) => onEnabledChange(event.target.checked)}
            disabled={disabled}
          />
          <span>Archive is password protected</span>
        </label>
        {enabled ? (
          <Input
            type="password"
            value={password}
            placeholder="Optional archive password"
            onChange={(event) => onPasswordChange(event.target.value)}
            disabled={disabled}
          />
        ) : null}
      </div>
    </Card>
  )
}
