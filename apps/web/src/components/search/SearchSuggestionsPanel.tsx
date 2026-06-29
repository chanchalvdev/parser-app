import { Badge } from '@/components/ui/Badge'
import type { SearchSuggestions } from '@/types/domain'

export const SearchSuggestionsPanel = ({ suggestions }: { suggestions: SearchSuggestions }) => {
  return (
    <section className="panel p-3">
      <h3 className="panel-title mb-2">Suggestions</h3>
      <div className="flex flex-wrap gap-2 text-xs">
        {suggestions.suggestions.record_type.map((item) => (
          <Badge tone="blue" key={`rt-${item.value}`}>
            {item.value}
          </Badge>
        ))}
        {suggestions.suggestions.detected_file_type.map((item) => (
          <Badge tone="gray" key={`dt-${item.value}`}>
            {item.value}
          </Badge>
        ))}
      </div>
    </section>
  )
}
