import type { ComponentType } from 'react'
import type { LucideProps } from 'lucide-react'
import {
  FileText,
  FileArchive,
  FileJson,
  FileSpreadsheet,
  FileCode,
  Image,
} from './Icon'

const ARCHIVE_EXTS = new Set([
  'zip', 'tar', 'gz', 'tgz', 'bz2', '7z', 'rar', 'xz', 'zst',
])
const JSON_EXTS = new Set(['json', 'jsonl', 'ndjson'])
const SPREADSHEET_EXTS = new Set(['csv', 'tsv', 'xlsx', 'xls'])
const IMAGE_EXTS = new Set([
  'png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp', 'ico', 'avif', 'heic',
])
const CODE_EXTS = new Set([
  'ts', 'tsx', 'js', 'jsx', 'mjs', 'cjs',
  'py', 'go', 'rs', 'java', 'kt', 'swift',
  'cpp', 'c', 'cc', 'h', 'hpp',
  'rb', 'php', 'sh', 'bash', 'zsh',
  'html', 'htm', 'css', 'scss', 'sass', 'less',
  'yaml', 'yml', 'xml', 'toml', 'ini', 'conf',
  'sql', 'md', 'mdx',
])

const IMAGE_MIME_PREFIX = 'image/'

function getExtension(filename: string | undefined | null): string {
  if (!filename) return ''
  const idx = filename.lastIndexOf('.')
  if (idx < 0 || idx === filename.length - 1) return ''
  return filename.slice(idx + 1).toLowerCase().split('?')[0].split('#')[0]
}

export interface FileTypeIconProps {
  filename?: string | null
  mimeType?: string | null
  size?: number
  className?: string
}

export function FileTypeIcon({ filename, mimeType, size = 18, className }: FileTypeIconProps) {
  const ext = getExtension(filename)
  const Icon: ComponentType<LucideProps> = resolveIcon(ext, mimeType)
  return <Icon size={size} className={className} aria-hidden="true" />
}

function resolveIcon(ext: string, mimeType?: string | null): ComponentType<LucideProps> {
  if (mimeType && mimeType.toLowerCase().startsWith(IMAGE_MIME_PREFIX)) return Image
  if (mimeType === 'application/pdf') return FileText

  if (ARCHIVE_EXTS.has(ext)) return FileArchive
  if (JSON_EXTS.has(ext)) return FileJson
  if (SPREADSHEET_EXTS.has(ext)) return FileSpreadsheet
  if (IMAGE_EXTS.has(ext)) return Image
  if (CODE_EXTS.has(ext)) return FileCode
  if (ext === 'pdf' || ext === 'doc' || ext === 'docx' || ext === 'txt' || ext === 'rtf') {
    return FileText
  }

  return FileText
}
