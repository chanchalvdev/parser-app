import * as Lucide from 'lucide-react'
import { forwardRef } from 'react'
import type { LucideIcon, LucideProps } from 'lucide-react'

const ICON_DEFAULTS = { size: 20, strokeWidth: 1.75 } as const

function wrap(Component: LucideIcon): LucideIcon {
  const Wrapped = forwardRef<SVGSVGElement, LucideProps>(function WrappedIcon(props, ref) {
    return <Component ref={ref} {...ICON_DEFAULTS} {...props} />
  })
  Wrapped.displayName = Component.displayName ?? 'Icon'
  return Wrapped as LucideIcon
}

// Navigation
export const LayoutDashboard = wrap(Lucide.LayoutDashboard)
export const UploadCloud = wrap(Lucide.UploadCloud)
export const ListChecks = wrap(Lucide.ListChecks)
export const Search = wrap(Lucide.Search)
export const FolderTree = wrap(Lucide.FolderTree)
export const Settings = wrap(Lucide.Settings)

// Controls
export const Menu = wrap(Lucide.Menu)
export const X = wrap(Lucide.X)
export const ChevronDown = wrap(Lucide.ChevronDown)
export const ChevronRight = wrap(Lucide.ChevronRight)
export const Plus = wrap(Lucide.Plus)
export const Filter = wrap(Lucide.Filter)
export const Download = wrap(Lucide.Download)
export const Sun = wrap(Lucide.Sun)
export const Moon = wrap(Lucide.Moon)
export const RefreshCw = wrap(Lucide.RefreshCw)
export const Trash2 = wrap(Lucide.Trash2)
export const Copy = wrap(Lucide.Copy)
export const ExternalLink = wrap(Lucide.ExternalLink)
export const Eye = wrap(Lucide.Eye)
export const PlayCircle = wrap(Lucide.PlayCircle)

// File types
export const FileText = wrap(Lucide.FileText)
export const FileArchive = wrap(Lucide.FileArchive)
export const FileJson = wrap(Lucide.FileJson)
export const FileSpreadsheet = wrap(Lucide.FileSpreadsheet)
export const FileCode = wrap(Lucide.FileCode)
export const Image = wrap(Lucide.Image)
export const File = wrap(Lucide.File)
export const Folder = wrap(Lucide.Folder)
export const FolderOpen = wrap(Lucide.FolderOpen)

// Status / feedback
export const AlertCircle = wrap(Lucide.AlertCircle)
export const AlertTriangle = wrap(Lucide.AlertTriangle)
export const CheckCircle2 = wrap(Lucide.CheckCircle2)
export const Clock = wrap(Lucide.Clock)
export const XCircle = wrap(Lucide.XCircle)
export const Loader2 = wrap(Lucide.Loader2)
export const Info = wrap(Lucide.Info)
export const HelpCircle = wrap(Lucide.HelpCircle)
export const Sparkles = wrap(Lucide.Sparkles)

// Data / meta
export const Hash = wrap(Lucide.Hash)
export const Calendar = wrap(Lucide.Calendar)
export const HardDrive = wrap(Lucide.HardDrive)
export const Activity = wrap(Lucide.Activity)
export const TrendingUp = wrap(Lucide.TrendingUp)
export const Inbox = wrap(Lucide.Inbox)
