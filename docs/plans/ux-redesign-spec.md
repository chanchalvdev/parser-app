# UX/UI Redesign Spec — Enterprise File Ingestion Platform

## Goal
Make the React frontend approachable for new and general users. Today it looks like a developer demo: raw UUIDs, plain text statuses, no icons, no empty/loading states, broken mobile. We will turn it into a polished app with clear visual hierarchy, consistent feedback, and a friendly tone.

Working directory: `parser/frontend/`

## Design system (new)

`src/styles.css` — full rewrite. Tokens:

```
--bg-1, --bg-2          page gradient
--surface               card background
--surface-2             nested surface
--surface-hover         row hover
--border                subtle border
--border-strong         focus ring
--text                  primary text
--text-muted            secondary text
--text-dim              tertiary text
--accent                primary accent (cyan)
--accent-hover
--accent-soft           translucent accent bg
--success / --warning / --danger / --info  status colors with -soft, -bg variants
--radius-sm/md/lg/xl
--shadow-sm/md/lg
--space-1..8            spacing scale
--font-sm/md/lg/xl/2xl/3xl
```

Add:
- Modern light theme (`[data-theme="light"]` overrides) with toggle in top bar.
- Smooth transitions (150ms) on interactive elements.
- Focus ring utility (visible 2px outline).
- Skeleton shimmer keyframe.
- Spinner keyframe.
- Toast slide-in keyframe.

## Component library (new)

Create these in `src/components/`:

1. **Button.tsx** — variants: `primary | secondary | ghost | danger`; sizes: `sm | md | lg`; props: `loading`, `disabled`, `iconLeft`, `iconRight`. Hover/active/focus states. Disabled reduces opacity.
2. **Badge.tsx** — variants: `neutral | info | success | warning | danger | accent`; size: `sm | md`. Used for status pills.
3. **Card.tsx** — props: `title`, `subtitle`, `actions` (right-aligned), `padding`. Used everywhere instead of raw `<section className="card">`.
4. **Spinner.tsx** — sizes `sm | md | lg`, optional `label`.
5. **EmptyState.tsx** — props: `icon`, `title`, `description`, `action` (Button). Friendly tone.
6. **Modal.tsx** — props: `open`, `onClose`, `title`, `children`, `footer`, `size`. Backdrop click + Escape close. Focus trap (use a simple effect, not a library). Body scroll lock.
7. **Toast.tsx + ToastProvider.tsx** — `useToast()` hook returning `{ success, error, info, warning }`. Auto-dismiss after 4s, stack top-right, max 5. ARIA live region.
8. **Skeleton.tsx** — props: `width`, `height`, `circle`. Use shimmer animation.
9. **Icon.tsx** — thin wrapper around `lucide-react` to control `size` and `strokeWidth` defaults. Re-export named icons we use: `LayoutDashboard, UploadCloud, ListChecks, Search, FolderTree, Settings, Menu, X, ChevronDown, ChevronRight, FileText, FileArchive, FileJson, FileSpreadsheet, FileCode, Image, AlertCircle, CheckCircle2, Clock, XCircle, Loader2, RefreshCw, Trash2, Copy, ExternalLink, Plus, Filter, Download, Sun, Moon, Info, HelpCircle, PlayCircle, Eye, Sparkles`.
10. **FileTypeIcon.tsx** — maps `mime_type` / `filename` extension to one of the file icons from `Icon.tsx`. Defaults to `FileText`.
11. **StatusBadge.tsx** — maps job status (`queued|running|completed|failed|retrying|cancelled`) and upload status (`uploading|completed|failed|processing`) to a Badge variant + label.
12. **Breadcrumbs.tsx** — array of `{label, to?}` items, last item is current page (not a link).
13. **TopBar.tsx** — see Layout.
14. **Sidebar.tsx** — see Layout.
15. **CopyButton.tsx** — copies text to clipboard, shows check icon briefly.

## Dependencies

Add `lucide-react@^0.460.0` to `package.json` dependencies. (No other new deps; we write Modal/Toast from scratch to keep bundle small.)

## Layout.tsx — full redesign

- Two-column grid becomes:
  - Mobile: single column, sidebar becomes a drawer (slide-in from left), backdrop overlay.
  - Desktop: 240px sidebar + content area, with a 56px sticky top bar.
- **TopBar** contents:
  - Left: hamburger (mobile only) + page title (derived from route) + breadcrumb trail.
  - Right: global search shortcut hint, theme toggle (Sun/Moon), "?" help link.
- **Sidebar** contents:
  - Brand block at top: small logo mark + "File Ingestion" wordmark + tiny tagline "Local enterprise ingestion lab" (was cryptic, now muted small).
  - Nav items with icon + label. Active state: filled accent background with text accent and small left indicator bar.
  - Footer block: "v0.1.0 · local" environment chip.
- Mobile drawer: tap hamburger, sidebar slides in, backdrop closes it on tap. Lock body scroll.

## Page redesigns

Each page should follow the pattern: `Card → header with title/subtitle → content with clear loading/empty/error/data states → friendly copy`.

### Dashboard
- 4 stat tiles in a grid: **Total Uploads, Total Files, Parsed Records, Jobs in Progress**. Each tile has icon, value, delta hint (placeholder "since yesterday" subtle muted text).
- **Job status breakdown** card: horizontal stacked bar showing proportions of statuses with legend chips. Below: small status counts table.
- **Recent jobs** card: 5 most recent jobs as clickable rows with: filename, status badge, stage, relative time ("2m ago"), attempts count. Click opens JobDetailModal.
- **Quick actions** row: 3 prominent buttons → "Upload a file", "Search content", "View all jobs".
- Loading: skeleton tiles + shimmer. Error: ErrorState with retry button. Empty: friendly "No activity yet — upload your first file to get started" with CTA.

### UploadPage
- Big drag-and-drop zone at top (dashed border, large icon, "Drop a file here or click to browse"). Accepts single or multiple files. Shows file chips below with filename, size, type icon, remove button.
- Optional archive password field (only enabled when file is `.zip`/`.rar`/`.7z`).
- "Upload N files" primary button. Disabled when nothing selected.
- Progress: while uploading, each chip shows a progress bar (we don't have real upload progress in api.ts, so simulate with optimistic state and complete on response — call out the actual limitation in code comment).
- After success: toast success, list of created (upload_id, job_id) with copy buttons + link to job detail. Also show a "Recent uploads" card below with last 5 from `api.listUploads`.
- Helpful hints: supported formats list, max size (read from settings).

### JobsPage
- Header: title "Jobs", subtitle "Track parsing and indexing jobs in real time".
- Filter row: status multi-select chips (All + each status), free-text search by id/filename, refresh button (manual). Auto-refresh every 5s with a "Live" indicator (pulsing dot) — pause when tab hidden.
- Table columns: **Job** (full id with copy button), **Upload filename** (joined from upload), **Status** (StatusBadge), **Stage**, **Attempts**, **Created** (relative), **Updated** (relative), **Error** (truncated with tooltip or "Details" link).
- Row click opens JobDetailModal with full info: job fields, error message, recent events timeline (from `/api/v1/jobs/{id}/events`), retry-with-password inline form (password input + button; only enabled for failed jobs).
- Empty state: "No jobs yet" with CTA to upload.
- Pagination via limit param + "Load more".

### SearchPage
- Hero search input (large) with icon, Enter to search, placeholder "Search parsed content… e.g. invoice number, customer name".
- Below input: filter chips (file type, date range placeholder).
- Results as cards (not a list): filename + path, highlighted snippet (we'll do simple substring highlight), score badge, click → file detail.
- Suggestions: when input has 2+ chars and no search yet, show recent/suggested queries from `api.searchSuggestions`.
- Empty state when no query yet: large icon + "Try searching for invoice, contract, log error…" + example chips.
- Empty state after search with 0 hits: "No matches for 'X'. Try a different keyword."

### FilesPage
- Header: title "Files", subtitle "Browse the parsed file tree from any upload".
- Left pane: list of uploads (with thumbnail/file-type icon, filename, status badge, size, date). Selected one highlighted.
- Right pane: collapsible tree of files for the selected upload. Tree node click selects a node and shows details panel below the tree: filename, kind, mime, size, sha256 (with copy), path breadcrumb, "Open in search" link (pre-fills search query).
- Tree expand/collapse with chevron buttons, not the current auto-render everything.
- Search-within-tree input that filters visible nodes.
- Empty state when no uploads: "No uploads yet — your parsed files will appear here."

### SettingsPage
- Group into sections with section headers: **Upload Limits**, **Processing**, **Audit** (placeholder cards).
- Each row: label + helper text + control.
- Save button per section with success toast. Inline validation messages.
- Show API call errors as inline alerts, not raw throws.
- Add a "Danger zone" section with disabled "Reset all data" button (no-op for now) and explanatory text.

## Accessibility
- All interactive elements focusable; visible focus ring.
- Icon-only buttons have `aria-label`.
- Modals trap focus, close on Escape, restore focus on close.
- Toasts in `role="status"` / `role="alert"`.
- Color contrast AA on text.
- Skip-to-content link.

## Responsive breakpoints
- Mobile (<768px): single column, drawer sidebar, stacked cards, scrollable tables.
- Tablet (768–1024px): 2-col where useful.
- Desktop (>1024px): full layout.

## Files touched
- `parser/frontend/package.json` (add lucide-react)
- `parser/frontend/index.html` (title, theme-color meta, font preload)
- `parser/frontend/src/main.tsx` (wrap with ToastProvider, set initial theme)
- `parser/frontend/src/styles.css` (full rewrite)
- `parser/frontend/src/App.tsx` (minor: add /help route if needed)
- `parser/frontend/src/components/Layout.tsx` (rewrite)
- `parser/frontend/src/components/TopBar.tsx` (new)
- `parser/frontend/src/components/Sidebar.tsx` (new)
- `parser/frontend/src/components/Button.tsx` (new)
- `parser/frontend/src/components/Badge.tsx` (new)
- `parser/frontend/src/components/StatusBadge.tsx` (new)
- `parser/frontend/src/components/Card.tsx` (new)
- `parser/frontend/src/components/Spinner.tsx` (new)
- `parser/frontend/src/components/Skeleton.tsx` (new)
- `parser/frontend/src/components/EmptyState.tsx` (new)
- `parser/frontend/src/components/Modal.tsx` (new)
- `parser/frontend/src/components/Toast.tsx` + `ToastProvider.tsx` (new)
- `parser/frontend/src/components/Icon.tsx` (new)
- `parser/frontend/src/components/FileTypeIcon.tsx` (new)
- `parser/frontend/src/components/Breadcrumbs.tsx` (new)
- `parser/frontend/src/components/CopyButton.tsx` (new)
- `parser/frontend/src/components/JobDetailModal.tsx` (new)
- `parser/frontend/src/pages/Dashboard.tsx` (rewrite)
- `parser/frontend/src/pages/UploadPage.tsx` (rewrite)
- `parser/frontend/src/pages/JobsPage.tsx` (rewrite)
- `parser/frontend/src/pages/SearchPage.tsx` (rewrite)
- `parser/frontend/src/pages/FilesPage.tsx` (rewrite)
- `parser/frontend/src/pages/SettingsPage.tsx` (rewrite)

## Verification
- `cd parser/frontend && npm install && npm run build` must succeed.
- No TypeScript errors.
- Dev server starts on `npm run dev`.
- Manual smoke: navigate Dashboard → Upload → Jobs → Search → Files → Settings, no console errors, all states render.

## Out of scope (for this redesign)
- Real upload progress (would need backend changes).
- Real pagination on Jobs/Search (use existing limit param; add Load more).
- Authentication UI (no auth in current backend).
- Multi-tenant theming.
