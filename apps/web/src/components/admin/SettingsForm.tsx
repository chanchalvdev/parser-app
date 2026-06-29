import { type ChangeEvent } from 'react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ParserToggleList } from './ParserToggleList'
import type { AdminSettings } from '@/types/admin'

type SettingsField = Omit<AdminSettings, 'tenant_id' | 'enabled_parsers'>

type SettingsFormProps = {
  values: AdminSettings
  availableParsers: string[]
  onValuesChange: (next: AdminSettings) => void
  onSubmit: () => void
  disabled?: boolean
  submitting?: boolean
}

const numberFieldChange =
  (onChange: (next: AdminSettings) => void, values: AdminSettings) =>
  (key: keyof SettingsField) =>
  (event: ChangeEvent<HTMLInputElement>) => {
    const value = Number(event.target.value)
    onChange({
      ...values,
      [key]: Number.isNaN(value) ? 0 : value,
    })
  }

export const SettingsForm = ({
  values,
  availableParsers,
  onValuesChange,
  onSubmit,
  disabled,
  submitting,
}: SettingsFormProps) => {
  const handleNumberChange = numberFieldChange(onValuesChange, values)

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault()
        onSubmit()
      }}
      className="grid gap-4"
    >
      <section className="grid gap-4 md:grid-cols-2">
        <label className="space-y-1 text-sm">
          <span>Max upload size MB</span>
          <Input type="number" step="0.1" min="0" value={values.max_upload_size_mb} onChange={handleNumberChange('max_upload_size_mb')} />
        </label>
        <label className="space-y-1 text-sm">
          <span>Max archive depth</span>
          <Input type="number" step="1" min="0" value={values.max_archive_depth} onChange={handleNumberChange('max_archive_depth')} />
        </label>
        <label className="space-y-1 text-sm">
          <span>Max extracted files</span>
          <Input type="number" step="1" min="0" value={values.max_extracted_files} onChange={handleNumberChange('max_extracted_files')} />
        </label>
        <label className="space-y-1 text-sm">
          <span>Max extracted size MB</span>
          <Input type="number" step="1" min="0" value={values.max_extracted_size_mb} onChange={handleNumberChange('max_extracted_size_mb')} />
        </label>
        <label className="space-y-1 text-sm">
          <span>Txt small file limit MB</span>
          <Input type="number" step="0.1" min="0" value={values.txt_small_file_limit_mb} onChange={handleNumberChange('txt_small_file_limit_mb')} />
        </label>
        <label className="space-y-1 text-sm">
          <span>Max expansion ratio</span>
          <Input type="number" step="0.1" min="0" value={values.max_expansion_ratio} onChange={handleNumberChange('max_expansion_ratio')} />
        </label>
        <label className="space-y-1 text-sm">
          <span>Parser batch size</span>
          <Input type="number" step="1" min="1" value={values.parser_batch_size} onChange={handleNumberChange('parser_batch_size')} />
        </label>
        <label className="space-y-1 text-sm">
          <span>Search index batch size</span>
          <Input type="number" step="1" min="1" value={values.search_index_batch_size} onChange={handleNumberChange('search_index_batch_size')} />
        </label>
      </section>

      <ParserToggleList
        availableParsers={availableParsers}
        selected={values.enabled_parsers}
        onChange={(next) => onValuesChange({ ...values, enabled_parsers: next })}
        disabled={disabled}
      />

      <div>
        <Button type="submit" disabled={disabled || submitting}>
          {submitting ? 'Saving settings…' : 'Save settings'}
        </Button>
      </div>
    </form>
  )
}

