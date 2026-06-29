import { useMemo } from 'react'

type ParserToggleListProps = {
  availableParsers: string[]
  selected: string[]
  onChange: (next: string[]) => void
  disabled?: boolean
}

export const ParserToggleList = ({
  availableParsers,
  selected,
  onChange,
  disabled,
}: ParserToggleListProps) => {
  const parserOptions = useMemo(() => {
    return Array.from(new Set([...availableParsers, ...selected])).sort()
  }, [availableParsers, selected])

  const selectedSet = useMemo(() => new Set(selected), [selected])

  const handleToggle = (parser: string, enabled: boolean) => {
    const next = new Set(selected)

    if (enabled) {
      next.add(parser)
    } else {
      next.delete(parser)
    }

    onChange(Array.from(next).sort())
  }

  return (
    <fieldset className="space-y-2">
      <legend className="mb-2 text-sm font-semibold text-slate-200">Enabled parsers</legend>
      <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
        {parserOptions.map((parser) => {
          const checked = selectedSet.has(parser)
          return (
            <label key={parser} className="flex items-center gap-2 rounded border border-slate-700/60 px-3 py-2">
              <input
                type="checkbox"
                checked={checked}
                disabled={disabled}
                onChange={(event) => handleToggle(parser, event.target.checked)}
              />
              <span className="text-sm text-slate-100">{parser}</span>
            </label>
          )
        })}
      </div>
    </fieldset>
  )
}

