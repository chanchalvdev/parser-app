# Plan: Password eye icon + Live job status + Restart button

## Context

User asked for three related UX improvements:

1. **Show/hide password eye icon** in the password input.
2. **Live processing status of the job** — what's happening right now, in real time.
3. **Restart the job** — a button that re-queues an ingestion job from the UI.

### Current state (after exploration)

| Area | Today |
| --- | --- |
| Archive-password input | `apps/web/src/components/upload/OptionalPasswordInput.tsx` — plain `<Input type="password" />`. |
| Archive-password modal | `apps/web/src/components/files/PasswordRequiredModal.tsx` — same plain `<Input type="password" />`. |
| Job detail page | `apps/web/src/pages/JobDetailPage.tsx` — uses `useJob(jobId)` and `useJobEvents(jobId, 1, 100)` from `apps/web/src/hooks/apiHooks.ts`. Neither hook sets `refetchInterval`, so the page only updates on user action (mount, retry click, manual invalidation). |
| Job retry button | `apps/web/src/components/jobs/RetryJobButton.tsx` calls `POST /api/v1/jobs/{job_id}/retry`. Visible **only when** `job.status.toLowerCase() === 'failed'` (see `JobDetailPage.tsx` line ~145). |
| Backend job endpoints | `GET /api/v1/jobs/{job_id}`, `GET /api/v1/jobs/{job_id}/events`, `POST /api/v1/jobs/{job_id}/retry` — all already exist (`apps/api/internal/http/handlers/job_handler.go`). No SSE / `text/event-stream` anywhere in `apps/api/`. |
| Restart button on Jobs list | `apps/web/src/pages/JobsPage.tsx` already shows a `RetryJobButton` per row but only on certain statuses. |

## Goal

Three improvements without backend changes:

1. Eye-icon toggle on **both** password fields (upload + modal).
2. Live job status: `JobDetailPage` refreshes itself while the job is non-terminal; show a clear live status banner + "live" indicator on the timeline.
3. Restart button is always visible on the Job Detail page (and on each row of the Jobs list), disabled while the job is currently running, labelled "Retry job" for `failed` and "Restart job" otherwise.

## Changes

### 1. New component: `apps/web/src/components/ui/PasswordInput.tsx`
A controlled password input with an eye/eye-off toggle button on the right. Purely presentational, no extra dependencies.

- Renders a `<div class="relative">` containing the existing `<Input>` (typed `password` or `text`) plus a small button positioned absolutely on the right (`top-1/2 right-2 -translate-y-1/2`) showing an inline SVG eye (open) / eye-off (slashed).
- Props: `value: string`, `onChange: (next: string) => void`, `placeholder?: string`, `disabled?: boolean`, `autoComplete?: string`. All other native input attributes forwarded.
- Internal state: `const [visible, setVisible] = useState(false)`. The button toggles `visible`.
- Accessibility: button has `aria-label={visible ? 'Hide password' : 'Show password'}` and `type="button"`.
- Use the project's existing slate/blue colour scheme so it matches the rest of the UI.

### 2. Patch `apps/web/src/components/upload/OptionalPasswordInput.tsx`
Replace the `<Input type="password" ... />` with `<PasswordInput value={password} onChange={onPasswordChange} placeholder="Optional archive password" disabled={disabled} />`. No other changes.

### 3. Patch `apps/web/src/components/files/PasswordRequiredModal.tsx`
Same swap — replace `<Input type="password" ... />` with `<PasswordInput value={password} onChange={setPassword} placeholder="Archive password" disabled={mutation.isPending} />`.

### 4. Patch `apps/web/src/hooks/apiHooks.ts::useJob` and `useJobEvents`
Make both hooks self-refresh while the job is in a non-terminal state. The cleanest way is to use React Query's `refetchInterval` option that stops when the status becomes terminal.

```ts
const TERMINAL_STATUSES = new Set(['completed', 'failed', 'cancelled', 'canceled', 'password_required', 'wrong_password'])

export const useJob = (jobId, options) => useQuery({
  queryKey: ['job', jobId],
  enabled: !!jobId,
  queryFn: () => getJob(jobId),
  refetchInterval: (q) => {
    const status = q.state.data?.status?.toLowerCase()
    return status && TERMINAL_STATUSES.has(status) ? false : 1500
  },
  refetchIntervalInBackground: false,
  ...options,
})
```

Apply the same pattern to `useJobEvents`. (Both hooks accept arbitrary options so the caller can override `refetchInterval` if needed.)

### 5. New component: `apps/web/src/components/jobs/LiveJobStatusBanner.tsx`
A small banner that appears above the timeline when the job is in a non-terminal state. Reads the same `useJob` data via props (no extra fetch). Shows:

- Current `status` (via existing `JobStatusBadge`).
- Current `current_stage` ("queued", "downloading", "extracting", "indexing", …).
- Progress bar (a thin div with `width: ${progress_percent}%`).
- A pulsing dot + "Live" label.

When status is terminal, the banner disappears entirely.

### 6. Patch `apps/web/src/pages/JobDetailPage.tsx`
- Insert `<LiveJobStatusBanner job={job} />` between the "Job summary" Card and the "Root file" Card.
- Remove the `job.status.toLowerCase() === 'failed'` guard around the Actions Card so the **Restart** button is **always** visible.
- The `RetryJobButton` already takes a `label` prop — pass `label={job.status.toLowerCase() === 'failed' ? 'Retry job' : 'Restart job'}`.
- The button is already disabled while the mutation is in flight; add `disabled={isLiveActive}` where `isLiveActive` is `!TERMINAL_STATUSES.has(job.status.toLowerCase())` so users can't spam-restart a running job.

### 7. Patch `apps/web/src/pages/JobsPage.tsx`
On each row, ensure a `RetryJobButton` is present (or add a small "Restart" action), disabled when the row's job is in a non-terminal state. Re-use the same `TERMINAL_STATUSES` set (export it from `hooks/apiHooks.ts` so the page can import it without duplicating logic).

## Files touched

| Path | Change |
| --- | --- |
| `apps/web/src/components/ui/PasswordInput.tsx` | new |
| `apps/web/src/components/upload/OptionalPasswordInput.tsx` | swap `<Input>` → `<PasswordInput>` |
| `apps/web/src/components/files/PasswordRequiredModal.tsx` | swap `<Input>` → `<PasswordInput>` |
| `apps/web/src/hooks/apiHooks.ts` | add `TERMINAL_STATUSES` export + `refetchInterval` on `useJob` and `useJobEvents` |
| `apps/web/src/components/jobs/LiveJobStatusBanner.tsx` | new |
| `apps/web/src/pages/JobDetailPage.tsx` | insert banner; always show Restart; disable while running |
| `apps/web/src/pages/JobsPage.tsx` | always show Retry/Restart per row; disable while running |

No backend changes, no DB changes, no migration. All changes are within `apps/web/src/`.

## Verification

1. Rebuild web container: `docker compose -f /Users/chanchal/mastering/chanchal-soc/pp-main/parser/docker-compose.yml up --build -d web`.
2. Open `http://localhost:5188/upload` → toggle the "Archive is password protected" checkbox → confirm a password field appears **with** an eye icon on the right. Click the icon → password reveals; click again → masks.
3. Open `http://localhost:5188/jobs/{id}` for a job in a non-terminal state (e.g. trigger `bash /Users/chanchal/mastering/chanchal-soc/pp-main/parser/scripts/e2e-upload.sh` which leaves a job behind; or trigger a fresh upload then immediately navigate to the job page).
4. Confirm:
   - Live banner with status, current_stage, animated progress bar, and a "Live" indicator.
   - The job's `status` and `current_stage` update in the UI without page reload (poll at ~1.5 s).
   - The "Status timeline" list grows as the worker emits more `job_events`.
   - The "Restart job" button is present and enabled; clicking it calls `POST /api/v1/jobs/{id}/retry` and the page reflects the new queued state.
   - The button is **disabled** while status is non-terminal (e.g. `processing`, `extracting`).
5. On `http://localhost:5188/files/{id}` for a file that needs a password, click "Enter archive password" → confirm the modal's input also has the eye icon.

## Risk / rollback

- All changes are UI-only; rollback is `git checkout -- apps/web/src/`.
- The polling `refetchInterval` adds ~1 req / 1.5 s while a job is running; negligible load. Polling stops as soon as the job hits a terminal status.
- `useJob`'s default behavior changes for all callers; if any caller relies on "fetch once, no refetch", they can pass `refetchInterval: false` in options. None of the current callers do.
